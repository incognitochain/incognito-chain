package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/wallet"
	"github.com/pkg/errors"
)

type BuySellRequest struct {
	PaymentAddress privacy.PaymentAddress
	TokenID        common.Hash
	Amount         uint64
	BuyPrice       uint64 // in Constant unit

	TradeID []byte // To trade bond with DCB
	MetadataBase
}

func NewBuySellRequest(
	paymentAddress privacy.PaymentAddress,
	tokenID common.Hash,
	amount uint64,
	buyPrice uint64,
	metaType int,
) *BuySellRequest {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	result := &BuySellRequest{
		PaymentAddress: paymentAddress,
		Amount:         amount,
		BuyPrice:       buyPrice,
		MetadataBase:   metadataBase,
		TokenID:        tokenID,
	}
	return result
}

func (bsReq *BuySellRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	if len(bsReq.TradeID) > 0 {
		// Validation for trading bonds
		_, buy, _, amount, _ := bcr.GetLatestTradeActivation(bsReq.TradeID)
		if amount < bsReq.Amount {
			return false, errors.Errorf("trade bond requested amount too high, %d > %d\n", bsReq.Amount, amount)
		}
		if !buy {
			return false, errors.New("trade is for selling bonds, not buying")
		}
	}

	// no need to do other validations here since it'll be checked on beacon chain
	return true, nil
}

func (bsReq *BuySellRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(bsReq.PaymentAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if len(bsReq.TradeID) == 0 && len(bsReq.PaymentAddress.Tk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if bsReq.BuyPrice == 0 {
		return false, false, errors.New("Wrong request info's buy price")
	}
	if bsReq.Amount == 0 {
		return false, false, errors.New("Wrong request info's amount")
	}
	if len(bsReq.TokenID) != common.HashSize {
		return false, false, errors.New("Wrong request info's asset type")
	}
	if txr.CalculateTxValue() < bsReq.BuyPrice*bsReq.Amount {
		return false, false, errors.New("Sending constant amount is not enough for buying bonds.")
	}
	if !txr.IsCoinsBurning() {
		return false, false, errors.New("Must send coin to burning address")
	}

	// For DCB trading bonds with GOV
	if len(bsReq.TradeID) > 0 {
		keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
		dcbAddress := keyWalletDCBAccount.KeySet.PaymentAddress
		if !bytes.Equal(dcbAddress.Pk, bsReq.PaymentAddress.Pk) {
			return false, false, errors.New("buy bond request with TradeID must send assets to DCB's address")
		}
	}
	return true, true, nil
}

func (bsReq *BuySellRequest) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (bsReq *BuySellRequest) Hash() *common.Hash {
	record := bsReq.PaymentAddress.String()
	record += bsReq.TokenID.String()
	record += string(bsReq.Amount)
	record += string(bsReq.BuyPrice)
	record += bsReq.MetadataBase.Hash().String()
	record += string(bsReq.TradeID)

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (bsReq *BuySellRequest) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := map[string]interface{}{
		"txReqId": *(tx.Hash()),
		"meta":    *bsReq,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(BuyFromGOVRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (bsReq *BuySellRequest) CalculateSize() uint64 {
	return calculateSize(bsReq)
}

func (bsReq *BuySellRequest) VerifyMinerCreatedTxBeforeGettingInBlock(
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
	accumulatedData *component.UsedInstData,
) (bool, error) {
	meta := bsReq
	if len(meta.TradeID) == 0 {
		return true, nil
	}

	fmt.Printf("[db] verifying buy from GOV Request tx\n")
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
			BondID:    &meta.TokenID,
			Buy:       true,
			Activated: false,
			Amount:    td.Amount, // no need to check
			ReqAmount: meta.Amount,
		}

		buyPrice := bcr.GetSellBondPrice(txData.BondID)
		if td.Compare(txData) && meta.BuyPrice == buyPrice {
			idx = i
			break
		}
	}

	if idx == -1 {
		return false, errors.Errorf("no instruction found for BuySellRequest tx %s", tx.Hash().String())
	}

	instUsed[idx] += 1
	accumulatedData.TradeActivated[string(meta.TradeID)] = true
	fmt.Printf("[db] inst %d matched\n", idx)
	return true, nil
}

func (bsReq *BuySellRequest) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	if len(bsReq.TradeID) > 0 {
		// no need to have fee for this tx
		return true
	}
	txFee := tr.GetTxFee()
	fullFee := minFee * tr.GetTxActualSize()
	return !(txFee < fullFee)
}
