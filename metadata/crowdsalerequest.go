package metadata

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/wallet"
	"github.com/pkg/errors"
)

// CrowdsaleRequest represents a buying request created by user to send to DCB
type CrowdsaleRequest struct {
	SaleID []byte

	ValidUntil     uint64 // in originated shard, not beacon or payment shard, 0 for not limit
	PaymentAddress privacy.PaymentAddress
	MetadataBase
}

func NewCrowdsaleRequest(csReqData map[string]interface{}) (Metadata, error) {
	saleIDStr, okID := csReqData["SaleID"].(string)
	saleID, errSale := hex.DecodeString(saleIDStr)
	validUntil, okValid := csReqData["ValidUntil"].(float64)
	paymentAddressStr, okAddr := csReqData["PaymentAddress"].(string)
	keyWallet, errPayment := wallet.Base58CheckDeserialize(paymentAddressStr)

	if !okID || !okValid || !okAddr {
		return nil, errors.Errorf("Error parsing crowdsale request data")
	}
	if err := common.CheckError(errSale, errPayment); err != nil {
		return nil, err
	}

	result := &CrowdsaleRequest{
		PaymentAddress: keyWallet.KeySet.PaymentAddress,
		SaleID:         saleID,
		ValidUntil:     uint64(validUntil),
	}
	result.Type = CrowdsaleRequestMeta
	return result, nil
}

func (csReq *CrowdsaleRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// Check if sale exists and ongoing
	sale, err := bcr.GetSaleData(csReq.SaleID) // okay to use unsync data since we only use immutable fields
	if err != nil {
		return false, err
	}

	beaconHeight := bcr.GetBeaconHeight()
	if beaconHeight >= sale.EndBlock {
		return false, errors.Errorf("crowdsale ended")
	}

	// Check if request is still valid
	shardHeight := bcr.GetChainHeight(shardID)
	if csReq.ValidUntil > 0 && shardHeight >= csReq.ValidUntil {
		return false, errors.Errorf("crowdsale request is not valid anymore")
	}

	// Check if asset is sent to correct address
	if !sale.Buy {
		keyWalletBurnAccount, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
		unique, pubkey, _ := txr.GetUniqueReceiver()
		if !unique || !bytes.Equal(pubkey, keyWalletBurnAccount.KeySet.PaymentAddress.Pk[:]) {
			return false, errors.Errorf("crowdsale request must send CST to Burning address")
		}
	} else {
		keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
		unique, pubkey, _ := txr.GetTokenUniqueReceiver()
		if !unique || !bytes.Equal(pubkey, keyWalletDCBAccount.KeySet.PaymentAddress.Pk[:]) {
			return false, errors.Errorf("crowdsale request must send tokens to DCB address")
		}
	}

	return true, nil
}

func (csReq *CrowdsaleRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(csReq.PaymentAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	return false, true, nil
}

func (csReq *CrowdsaleRequest) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (csReq *CrowdsaleRequest) Hash() *common.Hash {
	record := csReq.PaymentAddress.String()
	record += string(csReq.SaleID)
	record += string(csReq.ValidUntil)

	// final hash
	record += csReq.MetadataBase.Hash().String()
	hash := common.HashH([]byte(record))
	return &hash
}

type CrowdsaleRequestAction struct {
	SaleID         []byte
	PaymentAddress privacy.PaymentAddress
	SentAmount     uint64
}

func (csReq *CrowdsaleRequest) BuildReqActions(txr Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	value, err := getCrowdsaleRequestActionValue(csReq, txr, bcr)
	if err != nil {
		return nil, err
	}
	action := []string{strconv.Itoa(CrowdsaleRequestMeta), value}
	return [][]string{action}, nil
}

func getCrowdsaleRequestActionValue(csReq *CrowdsaleRequest, txr Transaction, bcr BlockchainRetriever) (string, error) {
	// Calculate value of asset sent in request tx
	sale, err := bcr.GetSaleData(csReq.SaleID)
	if err != nil {
		return "", err
	}
	sentAmount := uint64(0)
	if !sale.Buy {
		_, _, sentAmount = txr.GetUniqueReceiver()
	} else {
		_, _, sentAmount = txr.GetTokenUniqueReceiver()
	}

	action := &CrowdsaleRequestAction{
		SaleID:         csReq.SaleID,
		PaymentAddress: csReq.PaymentAddress,
		SentAmount:     sentAmount,
	}
	value, err := json.Marshal(action)
	return string(value), err
}

func ParseCrowdsaleRequestActionValue(value string) ([]byte, privacy.PaymentAddress, uint64, error) {
	action := &CrowdsaleRequestAction{}
	err := json.Unmarshal([]byte(value), action)
	if err != nil {
		return nil, privacy.PaymentAddress{}, 0, err
	}
	return action.SaleID, action.PaymentAddress, action.SentAmount, nil
}

func (csReq *CrowdsaleRequest) CalculateSize() uint64 {
	return calculateSize(csReq)
}
