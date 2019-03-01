package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type BuyGOVTokenRequest struct {
	BuyerAddress privacy.PaymentAddress
	TokenID      common.Hash
	Amount       uint64
	BuyPrice     uint64 // in Constant unit
	MetadataBase
}

func NewBuyGOVTokenRequest(
	buyerAddress privacy.PaymentAddress,
	tokenID common.Hash,
	amount uint64,
	buyPrice uint64,
	metaType int,
) *BuyGOVTokenRequest {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &BuyGOVTokenRequest{
		BuyerAddress: buyerAddress,
		Amount:       amount,
		BuyPrice:     buyPrice,
		MetadataBase: metadataBase,
		TokenID:      tokenID,
	}
}

func (bgtr *BuyGOVTokenRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	govParams := bcr.GetGOVParams()
	sellingGOVTokensParams := govParams.SellingGOVTokens
	if sellingGOVTokensParams == nil {
		return false, errors.New("SellingGOVTokensParams params are not existed.")
	}

	if !bytes.Equal(common.GOVTokenID[:], bgtr.TokenID[:]) {
		return false, errors.New("Requested GOV tokenID has not been selling yet.")
	}

	// check if buy price againsts SellingGOVTokens params' GOVTokenPrice is correct or not
	if bgtr.BuyPrice < sellingGOVTokensParams.GOVTokenPrice {
		return false, errors.New("Requested buy price is under SellingGOVTokens params' buy price.")
	}
	return true, nil
}

func (bgtr *BuyGOVTokenRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(bgtr.BuyerAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if len(bgtr.BuyerAddress.Tk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	if bgtr.BuyPrice == 0 {
		return false, false, errors.New("Wrong request info's buy price")
	}
	if bgtr.Amount == 0 {
		return false, false, errors.New("Wrong request info's amount")
	}
	if len(bgtr.TokenID) != common.HashSize {
		return false, false, errors.New("Wrong request info's asset type")
	}
	return true, true, nil
}

func (bgtr *BuyGOVTokenRequest) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (bgtr *BuyGOVTokenRequest) Hash() *common.Hash {
	record := bgtr.BuyerAddress.String()
	record += bgtr.TokenID.String()
	record += string(bgtr.Amount)
	record += string(bgtr.BuyPrice)
	record += bgtr.MetadataBase.Hash().String()

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (bgtr *BuyGOVTokenRequest) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := map[string]interface{}{
		"txReqId": *(tx.Hash()),
		"meta":    *bgtr,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(BuyGOVTokenRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (bgtr *BuyGOVTokenRequest) CalculateSize() uint64 {
	return calculateSize(bgtr)
}
