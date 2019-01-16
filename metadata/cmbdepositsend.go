package metadata

import (
	"bytes"
	"encoding/hex"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

type CMBDepositSend struct {
	ContractID common.Hash

	MetadataBase
}

func NewCMBDepositSend(data map[string]interface{}) *CMBDepositSend {
	contract, err := hex.DecodeString(data["ContractID"].(string))
	if err != nil {
		return nil
	}
	contractHash, _ := (&common.Hash{}).NewHash(contract)
	result := CMBDepositSend{
		ContractID: *contractHash,
	}

	result.Type = CMBDepositSendMeta
	return &result
}

func (ds *CMBDepositSend) Hash() *common.Hash {
	record := string(ds.ContractID[:])

	// final hash
	record += string(ds.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (ds *CMBDepositSend) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	// Check if contract is still valid
	_, _, _, txContract, err := bcr.GetTransactionByHash(&ds.ContractID)
	if err != nil {
		return common.FalseValue, errors.Errorf("Error retrieving contract for sending deposit")
	}
	contractMeta := txContract.GetMetadata().(*CMBDepositContract)
	height, err := bcr.GetTxChainHeight(txr)
	if err != nil || contractMeta.ValidUntil >= height {
		return common.FalseValue, errors.Errorf("Deposit contract is not valid anymore")
	}

	// Check if the contract is not accepted
	_, err = bcr.GetDepositSend(ds.ContractID[:])
	if err != leveldb.ErrNotFound {
		if err != nil {
			return common.FalseValue, err
		}
		return common.FalseValue, errors.Errorf("Deposit contract already had response")
	}

	// Check if contract is addressed to current user
	sender := txr.GetSigPubKey()
	if !bytes.Equal(sender, contractMeta.Receiver.Pk[:]) {
		return common.FalseValue, errors.Errorf("Invalid sender for deposit contract")
	}

	// Check if deposit amount is correct
	cmbPubKey := txContract.GetSigPubKey()
	unique, pubkey, amount := txContract.GetUniqueReceiver()
	if !unique || !bytes.Equal(pubkey, cmbPubKey) {
		return common.FalseValue, errors.Errorf("Deposit can only be send to CMB")
	}
	if amount < contractMeta.DepositValue {
		return common.FalseValue, errors.Errorf("Deposit not enough")
	}

	return common.TrueValue, nil
}

func (ds *CMBDepositSend) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return common.TrueValue, common.TrueValue, nil // continue to check for fee
}

func (ds *CMBDepositSend) ValidateMetadataByItself() bool {
	return common.TrueValue
}
