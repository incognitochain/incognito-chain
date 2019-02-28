package metadata

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
)

type ErrorSaver struct {
	err error
}

func (s *ErrorSaver) Save(errs ...error) error {
	if s.err != nil {
		return s.err
	}
	for _, err := range errs {
		if err != nil {
			s.err = err
			return s.err
		}
	}
	return nil
}

func (s *ErrorSaver) Get() error {
	return s.err
}

type ReserveRequest struct {
	PaymentAddress privacy.PaymentAddress
	SaleID         []byte
	Info           []byte // offchain payment info (e.g. ETH/BTC txhash)

	Amount     *big.Int // amount of offchain asset (ignored if buying asset is not offchain)
	AssetPrice uint64   // ignored if buying asset is not offchain; otherwise, represents the price of buying asset; set by miner at mining time

	MetadataBase
}

func NewReserveRequest(rreqData map[string]interface{}) *ReserveRequest {
	// TODO(@0xbunyip) use error saver
	saleID, err := hex.DecodeString(rreqData["SaleId"].(string))
	if err != nil {
		return nil
	}
	info, err := hex.DecodeString(rreqData["Info"].(string))
	if err != nil {
		return nil
	}
	n := big.NewInt(0)
	n, ok := n.SetString(rreqData["Amount"].(string), 10)
	if !ok {
		n = big.NewInt(0)
	}
	result := &ReserveRequest{
		PaymentAddress: rreqData["PaymentAddress"].(privacy.PaymentAddress),
		SaleID:         saleID,
		Info:           info,
		Amount:         n,
		AssetPrice:     0,
	}
	result.Type = ReserveRequestMeta
	return result
}

func (rreq *ReserveRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// Check if sale exists and ongoing
	saleData, err := bcr.GetCrowdsaleData(rreq.SaleID)
	if err != nil {
		return false, err
	}
	// TODO(@0xbunyip): get height of beacon chain on new consensus
	height, err := bcr.GetTxChainHeight(txr)
	if err != nil || saleData.EndBlock >= height {
		return false, errors.Errorf("Crowdsale ended")
	}

	// Check if Payment address is DCB's
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	if !bytes.Equal(rreq.PaymentAddress.Pk[:], keyWalletDCBAccount.KeySet.PaymentAddress.Pk[:]) || !bytes.Equal(rreq.PaymentAddress.Tk[:], keyWalletDCBAccount.KeySet.PaymentAddress.Tk[:]) {
		return false, errors.Errorf("Crowdsale request must send CST to DCBAddress")
	}
	return true, nil
}

func (rreq *ReserveRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(rreq.PaymentAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	return false, true, nil
}

func (rreq *ReserveRequest) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (rreq *ReserveRequest) Hash() *common.Hash {
	record := rreq.PaymentAddress.String()
	record += string(rreq.SaleID)
	record += string(rreq.Info)
	record += string(rreq.Amount.String())
	record += string(rreq.AssetPrice)

	// final hash
	record += rreq.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (rreq *ReserveRequest) CalculateSize() uint64 {
	return calculateSize(rreq)
}
