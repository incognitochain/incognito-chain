package metadata

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"golang.org/x/crypto/sha3"
)

type LoanWithdraw struct {
	LoanID []byte
	Key    []byte

	MetadataBase
}

func NewLoanWithdraw(data map[string]interface{}) (Metadata, error) {
	result := LoanWithdraw{}
	s, _ := hex.DecodeString(data["LoanID"].(string))
	result.LoanID = s
	s, _ = hex.DecodeString(data["Key"].(string))
	result.Key = s

	result.Type = LoanWithdrawMeta
	return &result, nil
}

func (lw *LoanWithdraw) Hash() *common.Hash {
	record := string(lw.LoanID)
	record += string(lw.Key)

	// final hash
	record += lw.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (lw *LoanWithdraw) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	fmt.Println("Validating LoanWithdraw with blockchain!!!")
	// Check if a loan response with the same id exists on any chain
	txHashes, err := bcr.GetLoanTxs(lw.LoanID)
	if err != nil {
		return false, err
	}

	// Check if loan hasn't been withdrawed
	_, _, _, err = bcr.GetLoanPayment(lw.LoanID)
	if err != leveldb.ErrNotFound {
		if err == nil {
			return false, errors.Errorf("Loan has been withdrawed")
		}
		return false, err
	}

	// TODO(@0xbunyip): validate that for a loan, there's only one withdraw in a single block

	foundResponse := 0
	keyCorrect := false
	// TODO(@0xbunyip): make sure withdraw is not too close to escrowDeadline in SimpleLoan smart contract
	for _, txHash := range txHashes {
		hash := &common.Hash{}
		copy(hash[:], txHash)
		_, _, _, txOld, err := bcr.GetTransactionByHash(hash)
		if txOld == nil || err != nil {
			return false, errors.New("Error finding corresponding loan request")
		}
		switch txOld.GetMetadataType() {
		case LoanRequestMeta:
			{
				// Check if key is correct
				meta := txOld.GetMetadata()
				if meta == nil {
					return false, errors.New("Loan request metadata of tx loan withdraw is nil")
				}
				requestMeta, ok := meta.(*LoanRequest)
				if !ok {
					return false, errors.New("Error parsing loan request of tx loan withdraw")
				}
				hasher := sha3.NewLegacyKeccak256()
				hasher.Write(lw.Key)
				digest := hasher.Sum(nil)
				fmt.Printf("Found committed digest, checking key and digest: %x\n%x\n%x\n", lw.Key, digest, requestMeta.KeyDigest)
				if bytes.Equal(digest, requestMeta.KeyDigest) {
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
				fmt.Printf("Found an accept response\n")
				if responseMeta.Response == Accept {
					foundResponse += 1
				}
			}
		}
	}

	minResponse := bcr.GetDCBParams().MinLoanResponseRequire
	if foundResponse < int(minResponse) {
		return false, errors.New("Not enough loan accepted response")
	}
	if !keyCorrect {
		return false, errors.New("Provided key is incorrect")
	}
	return true, nil
}

func (lw *LoanWithdraw) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return true, true, nil // continue checking for fee
}

func (lw *LoanWithdraw) ValidateMetadataByItself() bool {
	return true
}
