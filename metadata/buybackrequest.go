package metadata

import (
	"errors"
	"strconv"

	"github.com/ninjadotorg/constant/common"
)

type BuyBackRequest struct {
	BuyBackFromTxID common.Hash
	VoutIndex       int
	MetadataBase
}

func NewBuyBackRequest(bbReqData map[string]interface{}) *BuyBackRequest {
	metadataBase := MetadataBase{
		Type: int(bbReqData["type"].(float64)),
	}
	return &BuyBackRequest{
		BuyBackFromTxID: bbReqData["buyBackFromTxId"].(common.Hash),
		VoutIndex:       bbReqData["voutIndex"].(int),
		MetadataBase:    metadataBase,
	}
}

func (bbReq *BuyBackRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	chainID byte,
) (bool, error) {
	// check double spending on fee tx
	err := txr.ValidateConstDoubleSpendWithBlockchain(bcr, chainID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (bbReq *BuyBackRequest) ValidateSanityData(
	bcr BlockchainRetriever,
	txr Transaction,
) (bool, bool, error) {
	ok, err := txr.ValidateSanityData(bcr)
	if err != nil || !ok {
		return false, ok, err
	}
	if bbReq.VoutIndex < 0 {
		return false, false, errors.New("Wrong request info's vout index")
	}
	if len(bbReq.BuyBackFromTxID) == 0 {
		return false, false, errors.New("Wrong request info's BuyBackFromTxID")
	}
	return false, true, nil
}

func (bbReq *BuyBackRequest) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (bbReq *BuyBackRequest) Hash() *common.Hash {
	record := bbReq.BuyBackFromTxID.String()
	record += strconv.Itoa(bbReq.VoutIndex)
	record += string(bbReq.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
