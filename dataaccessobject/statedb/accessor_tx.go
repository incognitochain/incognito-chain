package statedb

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

func StoreSerialNumbers(stateDB *StateDB, tokenID common.Hash, serialNumbers [][]byte, shardID byte) error {
	for _, serialNumber := range serialNumbers {
		key := GenerateSerialNumberObjectKey(tokenID, shardID, serialNumber)
		value := NewSerialNumberStateWithValue(tokenID, shardID, serialNumber)
		err := stateDB.SetStateObject(SerialNumberObjectType, key, value)
		if err != nil {
			return NewStatedbError(StoreSerialNumberError, err)
		}
	}
	return nil
}

func HasSerialNumber(stateDB *StateDB, tokenID common.Hash, serialNumber []byte, shardID byte) (bool, error) {
	key := GenerateSerialNumberObjectKey(tokenID, shardID, serialNumber)
	s, has, err := stateDB.getSerialNumberState(key)
	if err != nil {
		return false, NewStatedbError(GetSerialNumberError, err)
	}
	if has && bytes.Compare(s.SerialNumber(), serialNumber) != 0 {
		panic("same key wrong value")
		return false, nil
	}
	return has, nil
}
func ListSerialNumber(stateDB *StateDB, tokenID common.Hash, shardID byte) (map[string]struct{}, error) {
	tempSerialNumbers := stateDB.getAllSerialNumberByPrefix(tokenID, shardID)
	m := make(map[string]struct{})
	for _, tempSerialNumber := range tempSerialNumbers {
		serialNumber := base58.Base58Check{}.Encode(tempSerialNumber, common.Base58Version)
		m[serialNumber] = struct{}{}
	}
	return m, nil
}

func StoreCommitments(stateDB *StateDB, tokenID common.Hash, pubkey []byte, commitments [][]byte, shardID byte) error {
	commitmentLengthKey := GenerateCommitmentLengthObjectKey(tokenID, shardID)
	commitmentLength, has, err := stateDB.getCommitmentLengthState(commitmentLengthKey)
	if err != nil {
		return NewStatedbError(GetCommitmentLengthError, err)
	}
	if !has {
		commitmentLength.SetBytes([]byte{0})
	} else {
		temp := commitmentLength.Uint64() + 1
		commitmentLength = new(big.Int).SetUint64(temp)
	}
	for _, commitment := range commitments {
		// store commitment
		keyCommitment := GenerateCommitmentObjectKey(tokenID, shardID, commitment)
		valueCommitment := NewCommitmentStateWithValue(tokenID, shardID, commitment, commitmentLength)
		err = stateDB.SetStateObject(CommitmentObjectType, keyCommitment, valueCommitment)
		if err != nil {
			return NewStatedbError(StoreCommitmentError, err)
		}
		// store commitment index
		keyCommitmentIndex := GenerateCommitmentIndexObjectKey(tokenID, shardID, commitmentLength)
		valueCommitmentIndex := keyCommitment
		err = stateDB.SetStateObject(CommitmentIndexObjectType, keyCommitmentIndex, valueCommitmentIndex)
		if err != nil {
			return NewStatedbError(StoreCommitmentIndexError, err)
		}
		// store commitment length
		keyCommitmentLength := GenerateCommitmentLengthObjectKey(tokenID, shardID)
		valueCommitmentLength := commitmentLength
		err = stateDB.SetStateObject(CommitmentLengthObjectType, keyCommitmentLength, valueCommitmentLength)
		if err != nil {
			return NewStatedbError(StoreCommitmentLengthError, err)
		}
		temp2 := commitmentLength.Uint64() + 1
		commitmentLength = new(big.Int).SetUint64(temp2)
	}
	return nil
}

func HasCommitment(stateDB *StateDB, tokenID common.Hash, commitment []byte, shardID byte) (bool, error) {
	key := GenerateCommitmentObjectKey(tokenID, shardID, commitment)
	c, has, err := stateDB.getCommitmentState(key)
	if err != nil {
		return false, NewStatedbError(GetCommitmentError, err)
	}
	if has && bytes.Compare(c.Commitment(), commitment) != 0 {
		panic("same key wrong value")
		return false, nil
	}
	return has, nil
}

func HasCommitmentIndex(stateDB *StateDB, tokenID common.Hash, commitmentIndex uint64, shardID byte) (bool, error) {
	commitmentIndexTemp := new(big.Int).SetUint64(commitmentIndex)
	key := GenerateCommitmentIndexObjectKey(tokenID, shardID, commitmentIndexTemp)
	c, has, err := stateDB.getCommitmentIndexState(key)
	if err != nil {
		return false, NewStatedbError(GetCommitmentIndexError, err)
	}
	if has && c.Index().Uint64() != commitmentIndex {
		panic("same key wrong value")
		return false, nil
	}
	return has, nil
}

func GetCommitmentByIndex(stateDB *StateDB, tokenID common.Hash, commitmentIndex uint64, shardID byte) ([]byte, error) {
	commitmentIndexTemp := new(big.Int).SetUint64(commitmentIndex)
	key := GenerateCommitmentIndexObjectKey(tokenID, shardID, commitmentIndexTemp)
	c, has, err := stateDB.getCommitmentIndexState(key)
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

func GetCommitmentStateByIndex(stateDB *StateDB, tokenID common.Hash, commitmentIndex uint64, shardID byte) (*CommitmentState, error) {
	commitmentIndexTemp := new(big.Int).SetUint64(commitmentIndex)
	key := GenerateCommitmentIndexObjectKey(tokenID, shardID, commitmentIndexTemp)
	c, has, err := stateDB.getCommitmentIndexState(key)
	if err != nil  {
		return nil, NewStatedbError(GetCommitmentIndexError, err)
	}
	if !has {
		return nil, NewStatedbError(GetCommitmentIndexError, errors.New("no value exist"))
	}
	return c, nil
}

// GetCommitmentIndex - return index of commitment in db list
func GetCommitmentIndex(stateDB *StateDB, tokenID common.Hash, commitment []byte, shardID byte) (*big.Int, error) {
	key := GenerateCommitmentObjectKey(tokenID, shardID, commitment)
	c, has, err := stateDB.getCommitmentState(key)
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
func GetCommitmentLength(stateDB *StateDB, tokenID common.Hash, shardID byte) (*big.Int, error) {
	key := GenerateCommitmentLengthObjectKey(tokenID, shardID)
	length, has, err := stateDB.getCommitmentLengthState(key)
	if err != nil {
		return nil, NewStatedbError(GetCommitmentLengthError, err)
	}
	if !has {
		return new(big.Int).SetUint64(0), NewStatedbError(GetCommitmentLengthError, errors.New("no value exist"))
	}
	return new(big.Int).SetUint64(length.Uint64() + 1), nil
}
func ListCommitment(stateDB *StateDB, tokenID common.Hash, shardID byte) (map[string]uint64, error) {
	m := stateDB.getAllCommitmentStateByPrefix(tokenID, shardID)
	return m, nil
}

// ListCommitmentIndices -  return all commitment index and its value
func ListCommitmentIndices(stateDB *StateDB, tokenID common.Hash, shardID byte) (map[uint64]string, error) {
	m := stateDB.getAllCommitmentStateByPrefix(tokenID, shardID)
	reverseM := make(map[uint64]string)
	for k, v := range m {
		reverseM[v] = k
	}
	return reverseM, nil
}

func StoreOutputCoins(stateDB *StateDB, tokenID common.Hash, publicKey []byte, outputCoins [][]byte, shardID byte) error {
	for _, outputCoin := range outputCoins {
		key := GenerateOutputCoinObjectKey(tokenID, shardID, publicKey, outputCoin)
		value := NewOutputCoinStateWithValue(tokenID, shardID, publicKey, outputCoin)
		err := stateDB.SetStateObject(OutputCoinObjectType, key, value)
		if err != nil {
			return NewStatedbError(StoreOutputCoinError, err)
		}
	}
	return nil
}
func GetOutcoinsByPubkey(stateDB *StateDB, tokenID common.Hash, publicKey []byte, shardID byte) ([][]byte, error) {
	outputCoinStates := stateDB.getAllOutputCoinState(tokenID, shardID, publicKey)
	o := [][]byte{}
	for _, outputCoinState := range outputCoinStates {
		o = append(o, outputCoinState.OutputCoin())
	}
	return o, nil
}

// StoreSNDerivators - store list serialNumbers by shardID
func StoreSNDerivators(stateDB *StateDB, tokenID common.Hash, snds [][]byte) error {
	for _, snd := range snds {
		key := GenerateSNDerivatorObjectKey(tokenID, snd)
		value := NewSNDerivatorStateWithValue(tokenID, snd)
		err := stateDB.SetStateObject(SNDerivatorObjectType, key, value)
		if err != nil {
			return NewStatedbError(StoreSNDerivatorError, err)
		}
	}
	return nil
}

// HasSNDerivator - Check SnDerivator in list SnDerivators by shardID
func HasSNDerivator(stateDB *StateDB, tokenID common.Hash, snd []byte) (bool, error) {
	key := GenerateSNDerivatorObjectKey(tokenID, snd)
	sndState, has, err := stateDB.getSNDerivatorState(key)
	if err != nil {
		return false, NewStatedbError(GetSNDerivatorError, err)
	}
	if has && bytes.Compare(sndState.Snd(), snd) != 0 {
		panic("same key wrong value")
	}
	return has, nil
}

func ListSNDerivator(stateDB *StateDB, tokenID common.Hash) ([][]byte, error) {
	return stateDB.getAllSNDerivatorStateByPrefix(tokenID), nil
}
