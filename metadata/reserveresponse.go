package metadata

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type ReserveResponse struct {
	RequestedTxID *common.Hash

	MetadataBase
}

func NewReserveResponse(data map[string]interface{}) *ReserveResponse {
	s, err := hex.DecodeString(data["RequestedTxID"].(string))
	if err != nil {
		return nil
	}
	result := &ReserveResponse{
		RequestedTxID: &common.Hash{},
	}
	result.Type = ReserveResponseMeta
	copy(result.RequestedTxID[:], s)
	return result
}

func (cr *ReserveResponse) Hash() *common.Hash {
	record := string(cr.RequestedTxID[:])

	// final hash
	record += cr.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (cr *ReserveResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// Check if only board members created this tx
	isBoard := false
	for _, gov := range bcr.GetBoardPubKeys(common.DCBBoard) {
		// TODO(@0xbunyip): change gov to []byte or use Base58Decode for entire payment address of governors
		if bytes.Equal([]byte(gov), txr.GetSigPubKey()) {
			isBoard = true
		}
	}
	if !isBoard {
		return false, errors.New("Tx must be created by DCB Governor")
	}

	//	// Check if crowdsale request exists
	//	txHashes, err := bcr.GetCrowdsaleTxs(cr.RequestedTxID[:])
	//	if err != nil {
	//		return false, err
	//	}
	//	if len(txHashes) == 0 {
	//		return false, errors.New("Found no request for current crowdsale response")
	//	}
	//	for _, txHash := range txHashes {
	//		hash, _ := (&common.Hash{}).NewHash(txHash)
	//		_, _, _, txOld, err := bcr.GetTransactionByHash(hash)
	//		if txOld == nil || err != nil {
	//			return false, errors.New("Error finding corresponding loan request")
	//		}
	//		switch txOld.GetMetadataType() {
	//		case ReserveResponseMeta:
	//			{
	//				// Check if the same user responses twice
	//				if bytes.Equal(txOld.GetSigPubKey(), txr.GetSigPubKey()) {
	//					return false, errors.New("Current board member already responded to crowdsale request")
	//				}
	//			}
	//		}
	//	}

	// Check if selling asset is of the right type
	_, _, _, txRequest, err := bcr.GetTransactionByHash(cr.RequestedTxID)
	if err != nil {
		return false, err
	}
	requestMeta := txRequest.GetMetadata().(*CrowdsaleRequest)
	saleData, err := bcr.GetCrowdsaleData(requestMeta.SaleID)
	if err != nil {
		return false, err
	}
	if !saleData.SellingAsset.IsEqual(&common.ConstantID) && !bytes.Equal(saleData.SellingAsset[:8], common.BondTokenID[:8]) && !saleData.SellingAsset.IsEqual(&common.DCBTokenID) {
		return false, errors.New("Selling asset of the crowdsale cannot have response tx")
	}
	return true, nil
}

func (cr *ReserveResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return false, true, nil // No need to check for fee
}

func (cr *ReserveResponse) ValidateMetadataByItself() bool {
	return true
}

// CheckTransactionFee returns true since loan response tx doesn't have fee
func (cr *ReserveResponse) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	return true
}

func (cr *ReserveResponse) CalculateSize() uint64 {
	return calculateSize(cr)
}
