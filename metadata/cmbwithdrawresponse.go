package metadata

import (
	"bytes"
	"encoding/hex"

	"github.com/big0t/constant-chain/common"
	"github.com/big0t/constant-chain/database"
	"github.com/pkg/errors"
)

type CMBWithdrawResponse struct {
	RequestTxID common.Hash

	MetadataBase
}

func NewCMBWithdrawResponse(data map[string]interface{}) *CMBWithdrawResponse {
	request, err := hex.DecodeString(data["RequestTxID"].(string))
	if err != nil {
		return nil
	}
	requestHash, _ := (&common.Hash{}).NewHash(request)
	result := CMBWithdrawResponse{
		RequestTxID: *requestHash,
	}
	result.Type = CMBWithdrawResponseMeta
	return &result
}

func (cwres *CMBWithdrawResponse) Hash() *common.Hash {
	record := cwres.RequestTxID.String()

	// final hash
	record += cwres.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (cwres *CMBWithdrawResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// Check if request existed
	_, _, _, txRequest, err := bcr.GetTransactionByHash(&cwres.RequestTxID)
	if err != nil {
		return false, errors.Errorf("Error retrieving request for withdraw response")
	}

	// Get contract of the deposit
	metaReq := txRequest.GetMetadata().(*CMBWithdrawRequest)
	_, _, _, txContract, err := bcr.GetTransactionByHash(&metaReq.ContractID)
	metaContract := txContract.GetMetadata().(*CMBDepositContract)
	blockHeight, err := bcr.GetTxChainHeight(txr)
	if err != nil {
		return false, errors.Errorf("Error retrieving block height of tx chain")
	}

	// Check if amount is enough
	_, receiver, amount := txr.GetUniqueReceiver()
	if !bytes.Equal(receiver, metaContract.Receiver.Pk[:]) {
		return false, errors.Errorf("Withdraw response receiver incorrect")
	}
	if blockHeight < metaContract.MaturityAt {
		// Early withdrawal
		elapsed := uint64(blockHeight) - metaContract.ValidUntil
		depositTerm := uint64(metaContract.MaturityAt - metaContract.ValidUntil)
		expectedAmount := metaContract.TotalInterest*elapsed/depositTerm + metaContract.DepositValue
		if amount < expectedAmount {
			return false, errors.Errorf("Value of withdraw response is %s instead of %s", amount, expectedAmount)
		}
	} else {
		// Normal withdrawal
		expectedAmount := metaContract.TotalInterest + metaContract.DepositValue
		if amount < expectedAmount {
			return false, errors.Errorf("Value of withdraw response is %s instead of %s", amount, expectedAmount)
		}
	}
	return true, nil
}

func (cwres *CMBWithdrawResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	// TODO(@0xbunyip)
	return true, true, nil
}

func (cwres *CMBWithdrawResponse) ValidateMetadataByItself() bool {
	// TODO(@0xbunyip)
	return true
}

func (cwres *CMBWithdrawResponse) CalculateSize() uint64 {
	return calculateSize(cwres)
}
