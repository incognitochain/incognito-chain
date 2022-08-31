// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	common "github.com/incognitochain/incognito-chain/common"
	mock "github.com/stretchr/testify/mock"
)

// BlockInterface is an autogenerated mock type for the BlockInterface type
type BlockInterface struct {
	mock.Mock
}

func (_m *BlockInterface) GetBeaconHeight() uint64 {
	//TODO implement me
	panic("implement me")
}

func (_m *BlockInterface) ProposeHash() *common.Hash {
	//TODO implement me
	panic("implement me")
}

func (_m *BlockInterface) FullHashString() string {
	//TODO implement me
	panic("implement me")
}

// BodyHash provides a mock function with given fields:
func (_m *BlockInterface) BodyHash() common.Hash {
	ret := _m.Called()

	var r0 common.Hash
	if rf, ok := ret.Get(0).(func() common.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.Hash)
		}
	}

	return r0
}

// CommitteeFromBlock provides a mock function with given fields:
func (_m *BlockInterface) CommitteeFromBlock() common.Hash {
	ret := _m.Called()

	var r0 common.Hash
	if rf, ok := ret.Get(0).(func() common.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.Hash)
		}
	}

	return r0
}

// GetAggregateRootHash provides a mock function with given fields:
func (_m *BlockInterface) GetAggregateRootHash() common.Hash {
	ret := _m.Called()

	var r0 common.Hash
	if rf, ok := ret.Get(0).(func() common.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.Hash)
		}
	}

	return r0
}

// GetConsensusType provides a mock function with given fields:
func (_m *BlockInterface) GetConsensusType() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetCurrentEpoch provides a mock function with given fields:
func (_m *BlockInterface) GetCurrentEpoch() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetFinalityHeight provides a mock function with given fields:
func (_m *BlockInterface) GetFinalityHeight() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetHeight provides a mock function with given fields:
func (_m *BlockInterface) GetHeight() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetInstructions provides a mock function with given fields:
func (_m *BlockInterface) GetInstructions() [][]string {
	ret := _m.Called()

	var r0 [][]string
	if rf, ok := ret.Get(0).(func() [][]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]string)
		}
	}

	return r0
}

// GetPrevHash provides a mock function with given fields:
func (_m *BlockInterface) GetPrevHash() common.Hash {
	ret := _m.Called()

	var r0 common.Hash
	if rf, ok := ret.Get(0).(func() common.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.Hash)
		}
	}

	return r0
}

// GetProduceTime provides a mock function with given fields:
func (_m *BlockInterface) GetProduceTime() int64 {
	ret := _m.Called()

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// GetProducer provides a mock function with given fields:
func (_m *BlockInterface) GetProducer() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetProposeTime provides a mock function with given fields:
func (_m *BlockInterface) GetProposeTime() int64 {
	ret := _m.Called()

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// GetProposer provides a mock function with given fields:
func (_m *BlockInterface) GetProposer() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetRound provides a mock function with given fields:
func (_m *BlockInterface) GetRound() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// GetRoundKey provides a mock function with given fields:
func (_m *BlockInterface) GetRoundKey() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetShardID provides a mock function with given fields:
func (_m *BlockInterface) GetShardID() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// GetValidationField provides a mock function with given fields:
func (_m *BlockInterface) GetValidationField() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetVersion provides a mock function with given fields:
func (_m *BlockInterface) GetVersion() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// Hash provides a mock function with given fields:
func (_m *BlockInterface) Hash() *common.Hash {
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

// ToBytes provides a mock function with given fields:
func (_m *BlockInterface) ToBytes() ([]byte, error) {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
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

// Type provides a mock function with given fields:
func (_m *BlockInterface) Type() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
