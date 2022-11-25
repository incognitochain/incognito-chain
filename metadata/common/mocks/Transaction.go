// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import coin "github.com/incognitochain/incognito-chain/privacy/coin"
import common "github.com/incognitochain/incognito-chain/metadata/common"
import incognito_chaincommon "github.com/incognitochain/incognito-chain/common"
import mock "github.com/stretchr/testify/mock"
import proof "github.com/incognitochain/incognito-chain/privacy/proof"
import statedb "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

// Transaction is an autogenerated mock type for the Transaction type
type Transaction struct {
	mock.Mock
}

// CalculateTxValue provides a mock function with given fields:
func (_m *Transaction) CalculateTxValue() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// CheckData provides a mock function with given fields: db
func (_m *Transaction) CheckData(db *statedb.StateDB) error {
	ret := _m.Called(db)

	var r0 error
	if rf, ok := ret.Get(0).(func(*statedb.StateDB) error); ok {
		r0 = rf(db)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CheckTxVersion provides a mock function with given fields: _a0
func (_m *Transaction) CheckTxVersion(_a0 int8) bool {
	ret := _m.Called(_a0)

	var r0 bool
	if rf, ok := ret.Get(0).(func(int8) bool); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// GetInfo provides a mock function with given fields:
func (_m *Transaction) GetInfo() []byte {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	return r0
}

// GetLockTime provides a mock function with given fields:
func (_m *Transaction) GetLockTime() int64 {
	ret := _m.Called()

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// GetMetadata provides a mock function with given fields:
func (_m *Transaction) GetMetadata() common.Metadata {
	ret := _m.Called()

	var r0 common.Metadata
	if rf, ok := ret.Get(0).(func() common.Metadata); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.Metadata)
		}
	}

	return r0
}

// GetMetadataType provides a mock function with given fields:
func (_m *Transaction) GetMetadataType() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// GetProof provides a mock function with given fields:
func (_m *Transaction) GetProof() proof.Proof {
	ret := _m.Called()

	var r0 proof.Proof
	if rf, ok := ret.Get(0).(func() proof.Proof); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(proof.Proof)
		}
	}

	return r0
}

// GetReceiverData provides a mock function with given fields:
func (_m *Transaction) GetReceiverData() ([]coin.Coin, error) {
	ret := _m.Called()

	var r0 []coin.Coin
	if rf, ok := ret.Get(0).(func() []coin.Coin); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]coin.Coin)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetReceivers provides a mock function with given fields:
func (_m *Transaction) GetReceivers() ([][]byte, []uint64) {
	ret := _m.Called()

	var r0 [][]byte
	if rf, ok := ret.Get(0).(func() [][]byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]byte)
		}
	}

	var r1 []uint64
	if rf, ok := ret.Get(1).(func() []uint64); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]uint64)
		}
	}

	return r0, r1
}

// GetSenderAddrLastByte provides a mock function with given fields:
func (_m *Transaction) GetSenderAddrLastByte() byte {
	ret := _m.Called()

	var r0 byte
	if rf, ok := ret.Get(0).(func() byte); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(byte)
	}

	return r0
}

// GetSig provides a mock function with given fields:
func (_m *Transaction) GetSig() []byte {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	return r0
}

// GetSigPubKey provides a mock function with given fields:
func (_m *Transaction) GetSigPubKey() []byte {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	return r0
}

// GetTokenID provides a mock function with given fields:
func (_m *Transaction) GetTokenID() *incognito_chaincommon.Hash {
	ret := _m.Called()

	var r0 *incognito_chaincommon.Hash
	if rf, ok := ret.Get(0).(func() *incognito_chaincommon.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*incognito_chaincommon.Hash)
		}
	}

	return r0
}

// GetTransferData provides a mock function with given fields:
func (_m *Transaction) GetTransferData() (bool, []byte, uint64, *incognito_chaincommon.Hash) {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 []byte
	if rf, ok := ret.Get(1).(func() []byte); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]byte)
		}
	}

	var r2 uint64
	if rf, ok := ret.Get(2).(func() uint64); ok {
		r2 = rf()
	} else {
		r2 = ret.Get(2).(uint64)
	}

	var r3 *incognito_chaincommon.Hash
	if rf, ok := ret.Get(3).(func() *incognito_chaincommon.Hash); ok {
		r3 = rf()
	} else {
		if ret.Get(3) != nil {
			r3 = ret.Get(3).(*incognito_chaincommon.Hash)
		}
	}

	return r0, r1, r2, r3
}

// GetTxActualSize provides a mock function with given fields:
func (_m *Transaction) GetTxActualSize() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetTxBurnData provides a mock function with given fields:
func (_m *Transaction) GetTxBurnData() (bool, coin.Coin, *incognito_chaincommon.Hash, error) {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 coin.Coin
	if rf, ok := ret.Get(1).(func() coin.Coin); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(coin.Coin)
		}
	}

	var r2 *incognito_chaincommon.Hash
	if rf, ok := ret.Get(2).(func() *incognito_chaincommon.Hash); ok {
		r2 = rf()
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(*incognito_chaincommon.Hash)
		}
	}

	var r3 error
	if rf, ok := ret.Get(3).(func() error); ok {
		r3 = rf()
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// GetTxFee provides a mock function with given fields:
func (_m *Transaction) GetTxFee() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetTxFeeToken provides a mock function with given fields:
func (_m *Transaction) GetTxFeeToken() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetTxFullBurnData provides a mock function with given fields:
func (_m *Transaction) GetTxFullBurnData() (bool, coin.Coin, coin.Coin, *incognito_chaincommon.Hash, error) {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 coin.Coin
	if rf, ok := ret.Get(1).(func() coin.Coin); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(coin.Coin)
		}
	}

	var r2 coin.Coin
	if rf, ok := ret.Get(2).(func() coin.Coin); ok {
		r2 = rf()
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(coin.Coin)
		}
	}

	var r3 *incognito_chaincommon.Hash
	if rf, ok := ret.Get(3).(func() *incognito_chaincommon.Hash); ok {
		r3 = rf()
	} else {
		if ret.Get(3) != nil {
			r3 = ret.Get(3).(*incognito_chaincommon.Hash)
		}
	}

	var r4 error
	if rf, ok := ret.Get(4).(func() error); ok {
		r4 = rf()
	} else {
		r4 = ret.Error(4)
	}

	return r0, r1, r2, r3, r4
}

// GetTxMintData provides a mock function with given fields:
func (_m *Transaction) GetTxMintData() (bool, coin.Coin, *incognito_chaincommon.Hash, error) {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 coin.Coin
	if rf, ok := ret.Get(1).(func() coin.Coin); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(coin.Coin)
		}
	}

	var r2 *incognito_chaincommon.Hash
	if rf, ok := ret.Get(2).(func() *incognito_chaincommon.Hash); ok {
		r2 = rf()
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(*incognito_chaincommon.Hash)
		}
	}

	var r3 error
	if rf, ok := ret.Get(3).(func() error); ok {
		r3 = rf()
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// GetType provides a mock function with given fields:
func (_m *Transaction) GetType() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetValidationEnv provides a mock function with given fields:
func (_m *Transaction) GetValidationEnv() common.ValidationEnviroment {
	ret := _m.Called()

	var r0 common.ValidationEnviroment
	if rf, ok := ret.Get(0).(func() common.ValidationEnviroment); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.ValidationEnviroment)
		}
	}

	return r0
}

// GetVersion provides a mock function with given fields:
func (_m *Transaction) GetVersion() int8 {
	ret := _m.Called()

	var r0 int8
	if rf, ok := ret.Get(0).(func() int8); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int8)
	}

	return r0
}

// Hash provides a mock function with given fields:
func (_m *Transaction) Hash() *incognito_chaincommon.Hash {
	ret := _m.Called()

	var r0 *incognito_chaincommon.Hash
	if rf, ok := ret.Get(0).(func() *incognito_chaincommon.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*incognito_chaincommon.Hash)
		}
	}

	return r0
}

// HashWithoutMetadataSig provides a mock function with given fields:
func (_m *Transaction) HashWithoutMetadataSig() *incognito_chaincommon.Hash {
	ret := _m.Called()

	var r0 *incognito_chaincommon.Hash
	if rf, ok := ret.Get(0).(func() *incognito_chaincommon.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*incognito_chaincommon.Hash)
		}
	}

	return r0
}

// Init provides a mock function with given fields: _a0
func (_m *Transaction) Init(_a0 interface{}) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IsCoinsBurning provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *Transaction) IsCoinsBurning(_a0 common.ChainRetriever, _a1 common.ShardViewRetriever, _a2 common.BeaconViewRetriever, _a3 uint64) bool {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	var r0 bool
	if rf, ok := ret.Get(0).(func(common.ChainRetriever, common.ShardViewRetriever, common.BeaconViewRetriever, uint64) bool); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsPrivacy provides a mock function with given fields:
func (_m *Transaction) IsPrivacy() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsSalaryTx provides a mock function with given fields:
func (_m *Transaction) IsSalaryTx() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ListOTAHashH provides a mock function with given fields:
func (_m *Transaction) ListOTAHashH() []incognito_chaincommon.Hash {
	ret := _m.Called()

	var r0 []incognito_chaincommon.Hash
	if rf, ok := ret.Get(0).(func() []incognito_chaincommon.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]incognito_chaincommon.Hash)
		}
	}

	return r0
}

// ListSerialNumbersHashH provides a mock function with given fields:
func (_m *Transaction) ListSerialNumbersHashH() []incognito_chaincommon.Hash {
	ret := _m.Called()

	var r0 []incognito_chaincommon.Hash
	if rf, ok := ret.Get(0).(func() []incognito_chaincommon.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]incognito_chaincommon.Hash)
		}
	}

	return r0
}

// LoadData provides a mock function with given fields: db
func (_m *Transaction) LoadData(db *statedb.StateDB) error {
	ret := _m.Called(db)

	var r0 error
	if rf, ok := ret.Get(0).(func(*statedb.StateDB) error); ok {
		r0 = rf(db)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetGetSenderAddrLastByte provides a mock function with given fields: _a0
func (_m *Transaction) SetGetSenderAddrLastByte(_a0 byte) {
	_m.Called(_a0)
}

// SetInfo provides a mock function with given fields: _a0
func (_m *Transaction) SetInfo(_a0 []byte) {
	_m.Called(_a0)
}

// SetLockTime provides a mock function with given fields: _a0
func (_m *Transaction) SetLockTime(_a0 int64) {
	_m.Called(_a0)
}

// SetMetadata provides a mock function with given fields: _a0
func (_m *Transaction) SetMetadata(_a0 common.Metadata) {
	_m.Called(_a0)
}

// SetProof provides a mock function with given fields: _a0
func (_m *Transaction) SetProof(_a0 proof.Proof) {
	_m.Called(_a0)
}

// SetSig provides a mock function with given fields: _a0
func (_m *Transaction) SetSig(_a0 []byte) {
	_m.Called(_a0)
}

// SetSigPubKey provides a mock function with given fields: _a0
func (_m *Transaction) SetSigPubKey(_a0 []byte) {
	_m.Called(_a0)
}

// SetTxFee provides a mock function with given fields: _a0
func (_m *Transaction) SetTxFee(_a0 uint64) {
	_m.Called(_a0)
}

// SetType provides a mock function with given fields: _a0
func (_m *Transaction) SetType(_a0 string) {
	_m.Called(_a0)
}

// SetValidationEnv provides a mock function with given fields: _a0
func (_m *Transaction) SetValidationEnv(_a0 common.ValidationEnviroment) {
	_m.Called(_a0)
}

// SetVersion provides a mock function with given fields: _a0
func (_m *Transaction) SetVersion(_a0 int8) {
	_m.Called(_a0)
}

// String provides a mock function with given fields:
func (_m *Transaction) String() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// UnmarshalJSON provides a mock function with given fields: data
func (_m *Transaction) UnmarshalJSON(data []byte) error {
	ret := _m.Called(data)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte) error); ok {
		r0 = rf(data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ValidateDoubleSpendWithBlockchain provides a mock function with given fields: _a0, _a1, _a2
func (_m *Transaction) ValidateDoubleSpendWithBlockchain(_a0 byte, _a1 *statedb.StateDB, _a2 *incognito_chaincommon.Hash) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(byte, *statedb.StateDB, *incognito_chaincommon.Hash) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ValidateSanityData provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *Transaction) ValidateSanityData(_a0 common.ChainRetriever, _a1 common.ShardViewRetriever, _a2 common.BeaconViewRetriever, _a3 uint64) (bool, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	var r0 bool
	if rf, ok := ret.Get(0).(func(common.ChainRetriever, common.ShardViewRetriever, common.BeaconViewRetriever, uint64) bool); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(common.ChainRetriever, common.ShardViewRetriever, common.BeaconViewRetriever, uint64) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateSanityDataByItSelf provides a mock function with given fields:
func (_m *Transaction) ValidateSanityDataByItSelf() (bool, error) {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateSanityDataWithBlockchain provides a mock function with given fields: chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight
func (_m *Transaction) ValidateSanityDataWithBlockchain(chainRetriever common.ChainRetriever, shardViewRetriever common.ShardViewRetriever, beaconViewRetriever common.BeaconViewRetriever, beaconHeight uint64) (bool, error) {
	ret := _m.Called(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)

	var r0 bool
	if rf, ok := ret.Get(0).(func(common.ChainRetriever, common.ShardViewRetriever, common.BeaconViewRetriever, uint64) bool); ok {
		r0 = rf(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(common.ChainRetriever, common.ShardViewRetriever, common.BeaconViewRetriever, uint64) error); ok {
		r1 = rf(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateTransaction provides a mock function with given fields: _a0, _a1, _a2, _a3, _a4
func (_m *Transaction) ValidateTransaction(_a0 map[string]bool, _a1 *statedb.StateDB, _a2 *statedb.StateDB, _a3 byte, _a4 *incognito_chaincommon.Hash) (bool, []proof.Proof, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3, _a4)

	var r0 bool
	if rf, ok := ret.Get(0).(func(map[string]bool, *statedb.StateDB, *statedb.StateDB, byte, *incognito_chaincommon.Hash) bool); ok {
		r0 = rf(_a0, _a1, _a2, _a3, _a4)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 []proof.Proof
	if rf, ok := ret.Get(1).(func(map[string]bool, *statedb.StateDB, *statedb.StateDB, byte, *incognito_chaincommon.Hash) []proof.Proof); ok {
		r1 = rf(_a0, _a1, _a2, _a3, _a4)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]proof.Proof)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(map[string]bool, *statedb.StateDB, *statedb.StateDB, byte, *incognito_chaincommon.Hash) error); ok {
		r2 = rf(_a0, _a1, _a2, _a3, _a4)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ValidateTxByItself provides a mock function with given fields: _a0, _a1, _a2, _a3, _a4, _a5, _a6
func (_m *Transaction) ValidateTxByItself(_a0 map[string]bool, _a1 *statedb.StateDB, _a2 *statedb.StateDB, _a3 common.ChainRetriever, _a4 byte, _a5 common.ShardViewRetriever, _a6 common.BeaconViewRetriever) (bool, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3, _a4, _a5, _a6)

	var r0 bool
	if rf, ok := ret.Get(0).(func(map[string]bool, *statedb.StateDB, *statedb.StateDB, common.ChainRetriever, byte, common.ShardViewRetriever, common.BeaconViewRetriever) bool); ok {
		r0 = rf(_a0, _a1, _a2, _a3, _a4, _a5, _a6)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(map[string]bool, *statedb.StateDB, *statedb.StateDB, common.ChainRetriever, byte, common.ShardViewRetriever, common.BeaconViewRetriever) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3, _a4, _a5, _a6)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateTxCorrectness provides a mock function with given fields: db
func (_m *Transaction) ValidateTxCorrectness(db *statedb.StateDB) (bool, error) {
	ret := _m.Called(db)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*statedb.StateDB) bool); ok {
		r0 = rf(db)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*statedb.StateDB) error); ok {
		r1 = rf(db)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateTxSalary provides a mock function with given fields: _a0
func (_m *Transaction) ValidateTxSalary(_a0 *statedb.StateDB) (bool, error) {
	ret := _m.Called(_a0)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*statedb.StateDB) bool); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*statedb.StateDB) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateTxWithBlockChain provides a mock function with given fields: chainRetriever, shardViewRetriever, beaconViewRetriever, shardID, stateDB
func (_m *Transaction) ValidateTxWithBlockChain(chainRetriever common.ChainRetriever, shardViewRetriever common.ShardViewRetriever, beaconViewRetriever common.BeaconViewRetriever, shardID byte, stateDB *statedb.StateDB) error {
	ret := _m.Called(chainRetriever, shardViewRetriever, beaconViewRetriever, shardID, stateDB)

	var r0 error
	if rf, ok := ret.Get(0).(func(common.ChainRetriever, common.ShardViewRetriever, common.BeaconViewRetriever, byte, *statedb.StateDB) error); ok {
		r0 = rf(chainRetriever, shardViewRetriever, beaconViewRetriever, shardID, stateDB)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ValidateTxWithCurrentMempool provides a mock function with given fields: _a0
func (_m *Transaction) ValidateTxWithCurrentMempool(_a0 common.MempoolRetriever) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(common.MempoolRetriever) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ValidateType provides a mock function with given fields:
func (_m *Transaction) ValidateType() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Verify provides a mock function with given fields: _a0, _a1, _a2, _a3, _a4
func (_m *Transaction) Verify(_a0 map[string]bool, _a1 *statedb.StateDB, _a2 *statedb.StateDB, _a3 byte, _a4 *incognito_chaincommon.Hash) (bool, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3, _a4)

	var r0 bool
	if rf, ok := ret.Get(0).(func(map[string]bool, *statedb.StateDB, *statedb.StateDB, byte, *incognito_chaincommon.Hash) bool); ok {
		r0 = rf(_a0, _a1, _a2, _a3, _a4)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(map[string]bool, *statedb.StateDB, *statedb.StateDB, byte, *incognito_chaincommon.Hash) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3, _a4)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// VerifyMinerCreatedTxBeforeGettingInBlock provides a mock function with given fields: _a0, _a1, _a2, _a3, _a4, _a5
func (_m *Transaction) VerifyMinerCreatedTxBeforeGettingInBlock(_a0 *common.MintData, _a1 byte, _a2 common.ChainRetriever, _a3 *common.AccumulatedValues, _a4 common.ShardViewRetriever, _a5 common.BeaconViewRetriever) (bool, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3, _a4, _a5)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*common.MintData, byte, common.ChainRetriever, *common.AccumulatedValues, common.ShardViewRetriever, common.BeaconViewRetriever) bool); ok {
		r0 = rf(_a0, _a1, _a2, _a3, _a4, _a5)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*common.MintData, byte, common.ChainRetriever, *common.AccumulatedValues, common.ShardViewRetriever, common.BeaconViewRetriever) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3, _a4, _a5)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
