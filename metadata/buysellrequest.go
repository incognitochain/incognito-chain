package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
)

type BuySellRequest struct {
	PaymentAddress privacy.PaymentAddress
	TokenID        common.Hash
	Amount         uint64
	BuyPrice       uint64 // in Constant unit
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

	// TODO: support and validate for either bonds or govs buy requests

	govParams := bcr.GetGOVParams()
	sellingBondsParams := govParams.SellingBonds
	if sellingBondsParams == nil {
		return false, errors.New("SellingBonds component are not existed.")
	}

	bondID := sellingBondsParams.GetID()
	if !bytes.Equal(bondID[:], bsReq.TokenID[:]) {
		return false, errors.New("Requested tokenID has not been selling yet.")
	}

	// check if buy price againsts SellingBonds component' BondPrice is correct or not
	if bsReq.BuyPrice < sellingBondsParams.BondPrice {
		return false, errors.New("Requested buy price is under SellingBonds component' buy price.")
	}
	return true, nil
}

func (bsReq *BuySellRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(bsReq.PaymentAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if len(bsReq.PaymentAddress.Tk) == 0 {
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

	// final hash
	hash := common.DoubleHashH([]byte(record))
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
