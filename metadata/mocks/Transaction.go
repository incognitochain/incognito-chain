// Code generated by mockery v2.0.0. DO NOT EDIT.

package mocks

import (
	common "github.com/incognitochain/incognito-chain/common"
	metadata "github.com/incognitochain/incognito-chain/metadata"

	mock "github.com/stretchr/testify/mock"

	statedb "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

	zkp "github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
)

// Transaction is an autogenerated mock type for the Transaction type
type Transaction struct {
	mock.Mock
}

// CalculateBurningTxValue provides a mock function with given fields: bcr, retriever, viewRetriever, beaconHeight
func (_m *Transaction) CalculateBurningTxValue(bcr metadata.ChainRetriever, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) (bool, uint64) {
	ret := _m.Called(bcr, retriever, viewRetriever, beaconHeight)

	var r0 bool
	if rf, ok := ret.Get(0).(func(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever, uint64) bool); ok {
		r0 = rf(bcr, retriever, viewRetriever, beaconHeight)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 uint64
	if rf, ok := ret.Get(1).(func(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever, uint64) uint64); ok {
		r1 = rf(bcr, retriever, viewRetriever, beaconHeight)
	} else {
		r1 = ret.Get(1).(uint64)
	}

	return r0, r1
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

// GetFullTxValues provides a mock function with given fields:
func (_m *Transaction) GetFullTxValues() (uint64, uint64) {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	var r1 uint64
	if rf, ok := ret.Get(1).(func() uint64); ok {
		r1 = rf()
	} else {
		r1 = ret.Get(1).(uint64)
	}

	return r0, r1
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
func (_m *Transaction) GetMetadata() metadata.Metadata {
	ret := _m.Called()

	var r0 metadata.Metadata
	if rf, ok := ret.Get(0).(func() metadata.Metadata); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(metadata.Metadata)
		}
	}

	return r0
}

// GetMetadataFromVinsTx provides a mock function with given fields: _a0, _a1, _a2
func (_m *Transaction) GetMetadataFromVinsTx(_a0 metadata.ChainRetriever, _a1 metadata.ShardViewRetriever, _a2 metadata.BeaconViewRetriever) (metadata.Metadata, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 metadata.Metadata
	if rf, ok := ret.Get(0).(func(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever) metadata.Metadata); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(metadata.Metadata)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
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
func (_m *Transaction) GetProof() *zkp.PaymentProof {
	ret := _m.Called()

	var r0 *zkp.PaymentProof
	if rf, ok := ret.Get(0).(func() *zkp.PaymentProof); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*zkp.PaymentProof)
		}
	}

	return r0
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

// GetSender provides a mock function with given fields:
func (_m *Transaction) GetSender() []byte {
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
func (_m *Transaction) GetTokenID() *common.Hash {
	ret := _m.Called()

	var r0 *common.Hash
	if rf, ok := ret.Get(0).(func() *common.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*common.Hash)
		}
	}

	return r0
}

// GetTokenReceivers provides a mock function with given fields:
func (_m *Transaction) GetTokenReceivers() ([][]byte, []uint64) {
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

// GetTokenUniqueReceiver provides a mock function with given fields:
func (_m *Transaction) GetTokenUniqueReceiver() (bool, []byte, uint64) {
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

	return r0, r1, r2
}

// GetTransferData provides a mock function with given fields:
func (_m *Transaction) GetTransferData() (bool, []byte, uint64, *common.Hash) {
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

	var r3 *common.Hash
	if rf, ok := ret.Get(3).(func() *common.Hash); ok {
		r3 = rf()
	} else {
		if ret.Get(3) != nil {
			r3 = ret.Get(3).(*common.Hash)
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

// GetUniqueReceiver provides a mock function with given fields:
func (_m *Transaction) GetUniqueReceiver() (bool, []byte, uint64) {
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

	return r0, r1, r2
}

// Hash provides a mock function with given fields:
func (_m *Transaction) Hash() *common.Hash {
	ret := _m.Called()

	var r0 *common.Hash
	if rf, ok := ret.Get(0).(func() *common.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*common.Hash)
		}
	}

	return r0
}

// IsCoinsBurning provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *Transaction) IsCoinsBurning(_a0 metadata.ChainRetriever, _a1 metadata.ShardViewRetriever, _a2 metadata.BeaconViewRetriever, _a3 uint64) bool {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	var r0 bool
	if rf, ok := ret.Get(0).(func(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever, uint64) bool); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsFullBurning provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *Transaction) IsFullBurning(_a0 metadata.ChainRetriever, _a1 metadata.ShardViewRetriever, _a2 metadata.BeaconViewRetriever, _a3 uint64) bool {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	var r0 bool
	if rf, ok := ret.Get(0).(func(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever, uint64) bool); ok {
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

// ListSerialNumbersHashH provides a mock function with given fields:
func (_m *Transaction) ListSerialNumbersHashH() []common.Hash {
	ret := _m.Called()

	var r0 []common.Hash
	if rf, ok := ret.Get(0).(func() []common.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]common.Hash)
		}
	}

	return r0
}

// SetMetadata provides a mock function with given fields: _a0
func (_m *Transaction) SetMetadata(_a0 metadata.Metadata) {
	_m.Called(_a0)
}

// ValidateDoubleSpendWithBlockchain provides a mock function with given fields: _a0, _a1, _a2
func (_m *Transaction) ValidateDoubleSpendWithBlockchain(_a0 byte, _a1 *statedb.StateDB, _a2 *common.Hash) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(byte, *statedb.StateDB, *common.Hash) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ValidateSanityData provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *Transaction) ValidateSanityData(_a0 metadata.ChainRetriever, _a1 metadata.ShardViewRetriever, _a2 metadata.BeaconViewRetriever, _a3 uint64) (bool, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	var r0 bool
	if rf, ok := ret.Get(0).(func(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever, uint64) bool); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever, uint64) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateTransaction provides a mock function with given fields: _a0, _a1, _a2, _a3, _a4, _a5, _a6
func (_m *Transaction) ValidateTransaction(_a0 bool, _a1 *statedb.StateDB, _a2 *statedb.StateDB, _a3 byte, _a4 *common.Hash, _a5 bool, _a6 bool) (bool, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3, _a4, _a5, _a6)

	var r0 bool
	if rf, ok := ret.Get(0).(func(bool, *statedb.StateDB, *statedb.StateDB, byte, *common.Hash, bool, bool) bool); ok {
		r0 = rf(_a0, _a1, _a2, _a3, _a4, _a5, _a6)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(bool, *statedb.StateDB, *statedb.StateDB, byte, *common.Hash, bool, bool) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3, _a4, _a5, _a6)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateTxByItself provides a mock function with given fields: _a0, _a1, _a2, _a3, _a4, _a5, _a6, _a7
func (_m *Transaction) ValidateTxByItself(_a0 bool, _a1 *statedb.StateDB, _a2 *statedb.StateDB, _a3 metadata.ChainRetriever, _a4 byte, _a5 bool, _a6 metadata.ShardViewRetriever, _a7 metadata.BeaconViewRetriever) (bool, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3, _a4, _a5, _a6, _a7)

	var r0 bool
	if rf, ok := ret.Get(0).(func(bool, *statedb.StateDB, *statedb.StateDB, metadata.ChainRetriever, byte, bool, metadata.ShardViewRetriever, metadata.BeaconViewRetriever) bool); ok {
		r0 = rf(_a0, _a1, _a2, _a3, _a4, _a5, _a6, _a7)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(bool, *statedb.StateDB, *statedb.StateDB, metadata.ChainRetriever, byte, bool, metadata.ShardViewRetriever, metadata.BeaconViewRetriever) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3, _a4, _a5, _a6, _a7)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateTxWithBlockChain provides a mock function with given fields: chainRetriever, shardViewRetriever, beaconViewRetriever, shardID, stateDB
func (_m *Transaction) ValidateTxWithBlockChain(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, shardID byte, stateDB *statedb.StateDB) error {
	ret := _m.Called(chainRetriever, shardViewRetriever, beaconViewRetriever, shardID, stateDB)

	var r0 error
	if rf, ok := ret.Get(0).(func(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever, byte, *statedb.StateDB) error); ok {
		r0 = rf(chainRetriever, shardViewRetriever, beaconViewRetriever, shardID, stateDB)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ValidateTxWithCurrentMempool provides a mock function with given fields: _a0
func (_m *Transaction) ValidateTxWithCurrentMempool(_a0 metadata.MempoolRetriever) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(metadata.MempoolRetriever) error); ok {
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

// VerifyMinerCreatedTxBeforeGettingInBlock provides a mock function with given fields: _a0, _a1, _a2, _a3, _a4, _a5, _a6, _a7, _a8
func (_m *Transaction) VerifyMinerCreatedTxBeforeGettingInBlock(_a0 []metadata.Transaction, _a1 []int, _a2 [][]string, _a3 []int, _a4 byte, _a5 metadata.ChainRetriever, _a6 *metadata.AccumulatedValues, _a7 metadata.ShardViewRetriever, _a8 metadata.BeaconViewRetriever) (bool, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3, _a4, _a5, _a6, _a7, _a8)

	var r0 bool
	if rf, ok := ret.Get(0).(func([]metadata.Transaction, []int, [][]string, []int, byte, metadata.ChainRetriever, *metadata.AccumulatedValues, metadata.ShardViewRetriever, metadata.BeaconViewRetriever) bool); ok {
		r0 = rf(_a0, _a1, _a2, _a3, _a4, _a5, _a6, _a7, _a8)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]metadata.Transaction, []int, [][]string, []int, byte, metadata.ChainRetriever, *metadata.AccumulatedValues, metadata.ShardViewRetriever, metadata.BeaconViewRetriever) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3, _a4, _a5, _a6, _a7, _a8)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
