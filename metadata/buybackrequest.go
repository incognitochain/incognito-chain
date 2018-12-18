package metadata

import (
	"errors"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type BuyBackRequest struct {
	BuyBackFromTxID common.Hash
	VoutIndex       int
	MetadataBase
}

func NewBuyBackRequest(
	buyBackFromTxID common.Hash,
	voutIndex int,
	metaType int,
) *BuyBackRequest {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &BuyBackRequest{
		BuyBackFromTxID: buyBackFromTxID,
		VoutIndex:       voutIndex,
		MetadataBase:    metadataBase,
	}
}

func (bbReq *BuyBackRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) (bool, error) {
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
