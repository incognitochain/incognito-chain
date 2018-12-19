package metadata

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type CrowdsaleResponse struct {
	RequestedTxID *common.Hash

	MetadataBase
}

func NewCrowdsaleResponse(data map[string]interface{}) *CrowdsaleResponse {
	s, err := hex.DecodeString(data["RequestedTxID"].(string))
	if err != nil {
		return nil
	}
	result := &CrowdsaleResponse{
		RequestedTxID: &common.Hash{},
	}
	result.Type = CrowdsaleResponseMeta
	copy(result.RequestedTxID[:], s)
	return result
}

func (cr *CrowdsaleResponse) Hash() *common.Hash {
	record := string(cr.RequestedTxID[:])

	// final hash
	record += string(cr.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (cr *CrowdsaleResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	// Check if only board members created this tx
	isBoard := false
	for _, gov := range bcr.GetDCBBoardPubKeys() {
		// TODO(@0xbunyip): change gov to []byte or use Base58Decode for entire payment address of governors
		if bytes.Equal([]byte(gov), txr.GetJSPubKey()) {
			isBoard = true
		}
	}
	if !isBoard {
		return false, fmt.Errorf("Tx must be created by DCB Governor")
	}

	// Check if crowdsale request exists
	txHashes, err := bcr.GetCrowdsaleTxs(cr.RequestedTxID[:])
	if err != nil {
		return false, err
	}
	if len(txHashes) == 0 {
		return false, fmt.Errorf("Found no request for current crowdsale response")
	}
	for _, txHash := range txHashes {
		hash, _ := (&common.Hash{}).NewHash(txHash)
		_, _, _, txOld, err := bcr.GetTransactionByHash(hash)
		if txOld == nil || err != nil {
			return false, fmt.Errorf("Error finding corresponding loan request")
		}
		switch txOld.GetMetadataType() {
		case CrowdsaleResponseMeta:
			{
				// Check if the same user responses twice
				if bytes.Equal(txOld.GetJSPubKey(), txr.GetJSPubKey()) {
					return false, fmt.Errorf("Current board member already responded to crowdsale request")
				}
			}
		}
	}

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
	if !bytes.Equal(saleData.SellingAsset, common.ConstantID[:]) && !bytes.Equal(saleData.SellingAsset[:8], common.BondTokenID[:8]) && !bytes.Equal(saleData.SellingAsset, common.DCBTokenID[:]) {
		return false, fmt.Errorf("Selling asset of the crowdsale cannot have response tx")
	}
	return true, nil
}

func (cr *CrowdsaleResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return false, true, nil // No need to check for fee
}

func (cr *CrowdsaleResponse) ValidateMetadataByItself() bool {
	return true
}

// CheckTransactionFee returns true since loan response tx doesn't have fee
func (cr *CrowdsaleResponse) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	return true
}
