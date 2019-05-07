package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/wallet"
	"github.com/pkg/errors"
)

type BuyBackRequest struct {
	PaymentAddress privacy.PaymentAddress
	Amount         uint64

	TradeID []byte // To trade bond with DCB
	MetadataBase
}

func NewBuyBackRequest(
	paymentAddress privacy.PaymentAddress,
	amount uint64,
	metaType int,
) *BuyBackRequest {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &BuyBackRequest{
		PaymentAddress: paymentAddress,
		Amount:         amount,
		MetadataBase:   metadataBase,
	}
}

func (bbReq *BuyBackRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	if len(bbReq.TradeID) > 0 {
		// Validation for trading bonds
		_, buy, _, amount, _ := bcr.GetLatestTradeActivation(bbReq.TradeID)
		if amount < bbReq.Amount {
			return false, errors.Errorf("trade bond requested amount too high, %d > %d\n", bbReq.Amount, amount)
		}
		if buy {
			return false, errors.New("trade is for buying bonds, not selling")
		}
	}

	return true, nil
}

func (bbReq *BuyBackRequest) ValidateSanityData(
	bcr BlockchainRetriever,
	txr Transaction,
) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	if reflect.TypeOf(txr).String() == "*transaction.Tx" {
		return true, true, nil
	}

	if len(bbReq.PaymentAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if len(bbReq.TradeID) == 0 && len(bbReq.PaymentAddress.Tk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if bbReq.Amount == 0 {
		return false, false, errors.New("Wrong request info's amount")
	}
	if !txr.IsCoinsBurning() {
		return false, false, errors.New("Must send bonds to burning address")
	}
	if txr.CalculateTxValue() < bbReq.Amount {
		return false, false, errors.New("Burning bond amount in Vouts should be equal to metadata's amount")
	}

	// For DCB trading bonds with GOV
	if len(bbReq.TradeID) > 0 {
		keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
		dcbAddress := keyWalletDCBAccount.KeySet.PaymentAddress
		if !bytes.Equal(dcbAddress.Pk, bbReq.PaymentAddress.Pk) {
			return false, false, errors.New("buy back request with TradeID must send assets to DCB's address")
		}
	}
	return true, true, nil
}

func (bbReq *BuyBackRequest) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (bbReq *BuyBackRequest) Hash() *common.Hash {
	record := bbReq.PaymentAddress.String()
	record += string(bbReq.Amount)
	record += bbReq.MetadataBase.Hash().String()
	record += string(bbReq.TradeID)
	hash := common.HashH([]byte(record))
	return &hash
}

func (bbReq *BuyBackRequest) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	bondID := tx.GetTokenID()
	actionContent := map[string]interface{}{
		"txReqId":        *(tx.Hash()),
		"buyBackReqMeta": bbReq,
		"bondId":         *bondID,
	}

	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(BuyBackRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (bbReq *BuyBackRequest) CalculateSize() uint64 {
	return calculateSize(bbReq)
}

func (bbReq *BuyBackRequest) VerifyMinerCreatedTxBeforeGettingInBlock(
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
	accumulatedData *component.UsedInstData,
) (bool, error) {
	meta := bbReq
	if len(meta.TradeID) == 0 {
		return true, nil
	}
	fmt.Printf("[db] verifying buy back GOV Request tx\n")

	bondID := tx.GetTokenID()
	idx := -1
	for i, inst := range insts {
		if instUsed[i] > 0 || inst[0] != strconv.Itoa(TradeActivationMeta) || inst[1] != strconv.Itoa(int(shardID)) {
			continue
		}
		td, err := bcr.CalcTradeData(inst[2])
		if err != nil || !bytes.Equal(meta.TradeID, td.TradeID) {
			continue
		}

		// PaymentAddress is validated in metadata's ValidateWithBlockChain
		txData := &component.TradeData{
			TradeID:   meta.TradeID,
			BondID:    td.BondID, // not available for BuyBackRequest meta
			Buy:       false,
			Activated: false,
			Amount:    td.Amount, // no need to check
			ReqAmount: meta.Amount,
		}

		if td.Compare(txData) && bondID.IsEqual(td.BondID) {
			idx = i
			break
		}
	}

	if idx == -1 {
		return false, errors.Errorf("no instruction found for BuyBackRequest tx %s", tx.Hash().String())
	}

	instUsed[idx] += 1
	fmt.Printf("[db] inst %d matched\n", idx)
	return true, nil
}

func (bbReq *BuyBackRequest) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	if len(bbReq.TradeID) > 0 {
		// no need to have fee for this tx
		return true
	}
	txFee := tr.GetTxFee()
	fullFee := minFee * tr.GetTxActualSize()
	return !(txFee < fullFee)
}
