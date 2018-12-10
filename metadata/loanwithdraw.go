package metadata

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"golang.org/x/crypto/sha3"
)

type LoanWithdraw struct {
	LoanID []byte
	Key    []byte

	MetadataBase
}

func NewLoanWithdraw(data map[string]interface{}) *LoanWithdraw {
	result := LoanWithdraw{}
	s, _ := hex.DecodeString(data["LoanID"].(string))
	result.LoanID = s
	s, _ = hex.DecodeString(data["Key"].(string))
	result.Key = s

	result.Type = LoanWithdrawMeta
	return &result
}

func (lw *LoanWithdraw) Hash() *common.Hash {
	record := string(lw.LoanID)
	record += string(lw.Key)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (lw *LoanWithdraw) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte) (bool, error) {
	// Check if a loan response with the same id exists on any chain
	txHashes, err := bcr.GetLoanTxs(lw.LoanID)
	if err != nil {
		return false, err
	}
	foundResponse := 0
	keyCorrect := false
	validUntil := int32(0)
	for _, txHash := range txHashes {
		hash := &common.Hash{}
		copy(hash[:], txHash)
		_, _, _, txOld, err := bcr.GetTransactionByHash(hash)
		if txOld == nil || err != nil {
			return false, fmt.Errorf("Error finding corresponding loan request")
		}
		switch txOld.GetMetadataType() {
		case LoanRequestMeta:
			{
				// Check if key is correct
				meta := txOld.GetMetadata()
				if meta == nil {
					return false, fmt.Errorf("Loan request metadata of tx loan withdraw is nil")
				}
				requestMeta, ok := meta.(*LoanRequest)
				if !ok {
					return false, fmt.Errorf("Error parsing loan request of tx loan withdraw")
				}
				h := make([]byte, 32)
				sha3.ShakeSum256(h, lw.Key)
				if bytes.Equal(h, requestMeta.KeyDigest) {
					keyCorrect = true
				}
			}
		case LoanResponseMeta:
			{
				// Check if loan is accepted
				meta := txOld.GetMetadata()
				if meta == nil {
					continue
				}
				responseMeta, ok := meta.(*LoanResponse)
				if !ok {
					continue
				}
				if responseMeta.Response == Accept {
					foundResponse += 1
					validUntil = responseMeta.ValidUntil
				}
			}
		}
	}

	minResponse := bcr.GetDCBParams().MinLoanResponseRequire
	if foundResponse < int(minResponse) {
		return false, fmt.Errorf("Not enough loan accepted response")
	}
	if bcr.GetHeight() >= validUntil {
		return false, fmt.Errorf("Loan is not valid anymore, cannot claim Constant")
	}
	if !keyCorrect {
		return false, fmt.Errorf("Provided key is incorrect")
	}
	return true, nil
}

func (lw *LoanWithdraw) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(lw.Key) != LoanKeyLength {
		return false, false, nil
	}
	return true, true, nil // continue checking for fee
}

func (lw *LoanWithdraw) ValidateMetadataByItself() bool {
	return true
}
