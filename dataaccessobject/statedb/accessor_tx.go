package statedb

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

func StoreSerialNumbers(statedb StateDB, tokenID common.Hash, serialNumbers [][]byte, shardID byte) error {
	for _, serialNumber := range serialNumbers {
		key := GenerateSerialNumberObjectKey(tokenID, shardID, serialNumber)
		value := NewSerialNumberStateWithValue(tokenID, shardID, serialNumber)
		err := statedb.SetStateObject(SerialNumberObjectType, key, value)
		if err != nil {
			return NewStatedbError(StoreSerialNumberError, err)
		}
	}
	return nil
}

func HasSerialNumber(statedb StateDB, tokenID common.Hash, serialNumber []byte, shardID byte) (bool, error) {
	key := GenerateSerialNumberObjectKey(tokenID, shardID, serialNumber)
	s, has, err := statedb.GetSerialNumberState(key)
	if err != nil {
		return false, NewStatedbError(GetSerialNumberError, err)
	}
	if bytes.Compare(s.SerialNumber(), serialNumber) != 0 {
		panic("same key wrong value")
		return false, nil
	}
	return has, nil
}
func ListSerialNumber(statedb StateDB, tokenID common.Hash, shardID byte) (map[string]struct{}, error) {
	tempSerialNumbers := statedb.GetAllSerialNumberByPrefix(tokenID, shardID)
	m := make(map[string]struct{})
	for _, tempSerialNumber := range tempSerialNumbers {
		serialNumber := base58.Base58Check{}.Encode(tempSerialNumber, common.Base58Version)
		m[serialNumber] = struct{}{}
	}
	return m, nil
}

func StoreCommitments(statedb StateDB, tokenID common.Hash, pubkey []byte, commitments [][]byte, shardID byte) error {
	commitmentLengthKey := GenerateCommitmentLengthObjectKey(tokenID, shardID)
	commitmentLength, has, err := statedb.GetCommitmentLengthState(commitmentLengthKey)
	if err != nil {
		return NewStatedbError(GetCommitmentLengthError, err)
	}
	if !has {
		commitmentLength.SetUint64(0)
	}
	for _, commitment := range commitments {
		// store commitment
		keyCommitment := GenerateCommitmentObjectKey(tokenID, shardID, commitment)
		valueCommitment := NewCommitmentStateWithValue(tokenID, shardID, commitment, commitmentLength)
		err := statedb.SetStateObject(CommitmentObjectType, keyCommitment, valueCommitment)
		if err != nil {
			return NewStatedbError(StoreCommitmentError, err)
		}
		// store commitment index
		keyCommitmentIndex := GenerateCommitmentIndexObjectKey(tokenID, shardID, commitmentLength)
		valueCommitmentIndex := keyCommitment
		err = statedb.SetStateObject(CommitmentIndexObjectType, keyCommitmentIndex, valueCommitmentIndex)
		if err != nil {
			return NewStatedbError(StoreCommitmentIndexError, err)
		}
		// store commitment length
		keyCommitmentLength := GenerateCommitmentLengthObjectKey(tokenID, shardID)
		valueCommitmentLength := commitmentLength
		err = statedb.SetStateObject(CommitmentLengthObjectType, keyCommitmentLength, valueCommitmentLength)
		if err != nil {
			return NewStatedbError(StoreCommitmentLengthError, err)
		}
		temp := commitmentLength.Uint64() + 1
		commitmentLength.SetUint64(temp)
	}
	return nil
}

func HasCommitment(statedb StateDB, tokenID common.Hash, commitment []byte, shardID byte) (bool, error) {
	key := GenerateCommitmentObjectKey(tokenID, shardID, commitment)
	c, has, err := statedb.GetCommitmentState(key)
	if err != nil {
		return false, NewStatedbError(GetCommitmentError, err)
	}
	if bytes.Compare(c.Commitment(), commitment) != 0 {
		panic("same key wrong value")
		return false, nil
	}
	return has, nil
}

func HasCommitmentIndex(statedb StateDB, tokenID common.Hash, commitmentIndex uint64, shardID byte) (bool, error) {
	commitmentIndexTemp := new(big.Int).SetUint64(commitmentIndex)
	key := GenerateCommitmentIndexObjectKey(tokenID, shardID, commitmentIndexTemp)
	c, has, err := statedb.GetCommitmentIndexState(key)
	if err != nil {
		return false, NewStatedbError(GetCommitmentIndexError, err)
	}
	if c.Index().Uint64() != commitmentIndex {
		panic("same key wrong value")
		return false, nil
	}
	return has, nil
}

func GetCommitmentByIndex(statedb StateDB, tokenID common.Hash, commitmentIndex uint64, shardID byte) ([]byte, error) {
	commitmentIndexTemp := new(big.Int).SetUint64(commitmentIndex)
	key := GenerateCommitmentIndexObjectKey(tokenID, shardID, commitmentIndexTemp)
	c, has, err := statedb.GetCommitmentIndexState(key)
	if err != nil {
		return []byte{}, NewStatedbError(GetCommitmentIndexError, err)
	}
	if !has {
		return []byte{}, NewStatedbError(GetCommitmentIndexError, errors.New("no value exist"))
	}
	if c.Index().Uint64() != commitmentIndex {
		panic("same key wrong value")
		return []byte{}, nil
	}

	return c.commitment, nil
}

// GetCommitmentIndex - return index of commitment in db list
func GetCommitmentIndex(statedb StateDB, tokenID common.Hash, commitment []byte, shardID byte) (*big.Int, error) {
	key := GenerateCommitmentObjectKey(tokenID, shardID, commitment)
	c, has, err := statedb.GetCommitmentState(key)
	if err != nil {
		return nil, NewStatedbError(GetCommitmentError, err)
	}
	if !has {
		return nil, NewStatedbError(GetCommitmentError, errors.New("no value exist"))
	}
	if bytes.Compare(c.Commitment(), commitment) != 0 {
		panic("same key wrong value")
		return nil, nil
	}
	return c.Index(), nil
}

// GetCommitmentIndex - return index of commitment in db list
func GetCommitmentLength(statedb StateDB, tokenID common.Hash, shardID byte) (*big.Int, error) {
	key := GenerateCommitmentLengthObjectKey(tokenID, shardID)
	length, has, err := statedb.GetCommitmentLengthState(key)
	if err != nil {
		return nil, NewStatedbError(GetCommitmentLengthError, err)
	}
	if !has {
		return new(big.Int).SetUint64(0), NewStatedbError(GetCommitmentLengthError, errors.New("no value exist"))
	}
	return length, nil
}
func ListCommitment(statedb StateDB, tokenID common.Hash, shardID byte) (map[string]uint64, error) {
	m := statedb.GetAllCommitmentState(tokenID, shardID)
	return m, nil
}

// ListCommitmentIndices -  return all commitment index and its value
func ListCommitmentIndices(statedb StateDB, tokenID common.Hash, shardID byte) (map[uint64]string, error) {
	m := statedb.GetAllCommitmentState(tokenID, shardID)
	reverseM := make(map[uint64]string)
	for k, v := range m {
		reverseM[v] = k
	}
	return reverseM, nil
}

func StoreOutputCoins(statedb StateDB, tokenID common.Hash, publicKey []byte, outputCoins [][]byte, shardID byte) error {
	key := GenerateOutputCoinObjectKey(tokenID, shardID, publicKey)
	currentValue, has, err := statedb.GetOutputCoinState(key)
	if err != nil {
		return NewStatedbError(StoreOutputCoinError, err)
	}
	if has {
		outputCoins = append(currentValue.OutputCoins(), outputCoins...)
	}
	value := NewOutputCoinStateWithValue(tokenID, shardID, publicKey, outputCoins)
	err = statedb.SetStateObject(OutputCoinObjectType, key, value)
	if err != nil {
		return NewStatedbError(StoreOutputCoinError, err)
	}
	return nil
}
func GetOutcoinsByPubkey(statedb StateDB, tokenID common.Hash, publicKey []byte, shardID byte) ([][]byte, error) {
	key := GenerateOutputCoinObjectKey(tokenID, shardID, publicKey)
	o, has, err := statedb.GetOutputCoinState(key)
	if err != nil {
		return [][]byte{}, NewStatedbError(GetOutputCoinError, err)
	}
	if !has {
		return [][]byte{}, NewStatedbError(GetOutputCoinError, errors.New("no value exist"))
	}
	return o.OutputCoins(), nil
}
