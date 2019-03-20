package metadata

import (
	"bytes"
	"encoding/hex"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	WithdrawRequested = uint8(iota)
	WithdrawFulfilled
)

type CMBWithdrawRequest struct {
	ContractID common.Hash

	MetadataBase
}

func NewCMBWithdrawRequest(data map[string]interface{}) *CMBWithdrawRequest {
	contract, err := hex.DecodeString(data["ContractID"].(string))
	if err != nil {
		return nil
	}
	contractHash, _ := (&common.Hash{}).NewHash(contract)
	result := CMBWithdrawRequest{
		ContractID: *contractHash,
	}

	result.Type = CMBWithdrawRequestMeta
	return &result
}

func (cwr *CMBWithdrawRequest) Hash() *common.Hash {
	record := string(cwr.ContractID[:])

	// final hash
	record += cwr.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (cwr *CMBWithdrawRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// Check if request is made by receiver of the contract
	sender := txr.GetSigPubKey()
	_, _, _, txContract, err := bcr.GetTransactionByHash(&cwr.ContractID)
	if err != nil {
		return false, errors.Errorf("Error retrieving contract for withdrawal")
	}
	contractMeta := txContract.GetMetadata().(*CMBDepositContract)
	if !bytes.Equal(sender, contractMeta.Receiver.Pk[:]) {
		return false, errors.Errorf("Only contract receiver can initiate withdrawal")
	}

	// Check if no withdrawal request for the same contract
	_, _, err = bcr.GetWithdrawRequest(cwr.ContractID[:])
	if err != leveldb.ErrNotFound {
		if err != nil {
			return false, err
		}
		return false, errors.Errorf("Contract already had withdraw request")
	}

	// TODO(@0xbunyip): validate that no 2 withdrawal requests of a contract in the same block
	return true, nil
}

func (cwr *CMBWithdrawRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	// TODO(@0xbunyip)
	return true, true, nil // continue to check for fee
}

func (cwr *CMBWithdrawRequest) ValidateMetadataByItself() bool {
	// TODO(@0xbunyip)
	return true
}

func (cwr *CMBWithdrawRequest) CalculateSize() uint64 {
	return calculateSize(cwr)
}
