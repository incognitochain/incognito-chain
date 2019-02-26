package metadata

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
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
	// Check if loan hasn't been withdrawed
	_, _, _, err := bcr.GetLoanPayment(lw.LoanID)
	if err == nil {
		fmt.Printf("validate LoanWithdraw: withdrawed")
		return false, errors.Errorf("Loan has been withdrawed")
	}

	// Check that only shard of loan requester can process this type of tx
	reqMeta, err := bcr.GetLoanRequestMeta(lw.LoanID)
	if err != nil {
		return false, err
	}
	lastByte := reqMeta.ReceiveAddress.Pk[len(reqMeta.ReceiveAddress.Pk)-1]
	if shardID != common.GetShardIDFromLastByte(lastByte) {
		fmt.Printf("[db] validate LoanWithdraw fail: %d %d", shardID, lastByte)
		return false, errors.Errorf("Shard %d cannot process LoanWithdraw tx with lastByte", lastByte)
	}

	// TODO(@0xbunyip): validate that for a loan, there's only one withdraw in a single block
	// TODO(@0xbunyip): make sure withdraw is not too close to escrowDeadline in SimpleLoan smart contract

	// Check number of accepted responses
	_, responses, err := bcr.GetLoanResps(lw.LoanID)
	if err != nil {
		return false, err
	}
	foundResponse := 0
	for _, resp := range responses {
		if resp == Accept {
			foundResponse += 1
		}
	}
	minResponse := bcr.GetDCBParams().MinLoanResponseRequire
	if foundResponse < int(minResponse) {
		return false, errors.New("Not enough loan accepted response")
	}

	// Check if key is correct
	reqHash, err := bcr.GetLoanReq(lw.LoanID)
	if err != nil {
		return false, err
	}
	_, _, _, txReq, err := bcr.GetTransactionByHash(reqHash)
	if txReq == nil || err != nil {
		return false, errors.New("Error finding corresponding loan request on current shard")
	}
	requestMeta, ok := txReq.GetMetadata().(*LoanRequest)
	if !ok {
		return false, errors.New("Error parsing loan request of tx loan withdraw")
	}
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(lw.Key)
	digest := hasher.Sum(nil)
	if !bytes.Equal(digest, requestMeta.KeyDigest) {
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
