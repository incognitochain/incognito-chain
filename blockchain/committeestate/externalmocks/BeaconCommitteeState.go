// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package externalmocks

import (
	committeestate "github.com/incognitochain/incognito-chain/blockchain/committeestate"
	common "github.com/incognitochain/incognito-chain/common"

	incognitokey "github.com/incognitochain/incognito-chain/incognitokey"

	mock "github.com/stretchr/testify/mock"

	privacy "github.com/incognitochain/incognito-chain/privacy"
)

// BeaconCommitteeState is an autogenerated mock type for the BeaconCommitteeState type
type BeaconCommitteeState struct {
	mock.Mock
}

// Clone provides a mock function with given fields:
func (_m *BeaconCommitteeState) Clone() committeestate.BeaconCommitteeState {
	ret := _m.Called()

	var r0 committeestate.BeaconCommitteeState
	if rf, ok := ret.Get(0).(func() committeestate.BeaconCommitteeState); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(committeestate.BeaconCommitteeState)
		}
	}

	return r0
}

// GetAllCandidateSubstituteCommittee provides a mock function with given fields:
func (_m *BeaconCommitteeState) GetAllCandidateSubstituteCommittee() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// GetAutoStaking provides a mock function with given fields:
func (_m *BeaconCommitteeState) GetAutoStaking() map[string]bool {
	ret := _m.Called()

	var r0 map[string]bool
	if rf, ok := ret.Get(0).(func() map[string]bool); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]bool)
		}
	}

	return r0
}

// GetBeaconCommittee provides a mock function with given fields:
func (_m *BeaconCommitteeState) GetBeaconCommittee() []incognitokey.CommitteePublicKey {
	ret := _m.Called()

	var r0 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(0).(func() []incognitokey.CommitteePublicKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]incognitokey.CommitteePublicKey)
		}
	}

	return r0
}

// GetBeaconSubstitute provides a mock function with given fields:
func (_m *BeaconCommitteeState) GetBeaconSubstitute() []incognitokey.CommitteePublicKey {
	ret := _m.Called()

	var r0 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(0).(func() []incognitokey.CommitteePublicKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]incognitokey.CommitteePublicKey)
		}
	}

	return r0
}

// GetCandidateBeaconWaitingForCurrentRandom provides a mock function with given fields:
func (_m *BeaconCommitteeState) GetCandidateBeaconWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	ret := _m.Called()

	var r0 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(0).(func() []incognitokey.CommitteePublicKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]incognitokey.CommitteePublicKey)
		}
	}

	return r0
}

// GetCandidateBeaconWaitingForNextRandom provides a mock function with given fields:
func (_m *BeaconCommitteeState) GetCandidateBeaconWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	ret := _m.Called()

	var r0 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(0).(func() []incognitokey.CommitteePublicKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]incognitokey.CommitteePublicKey)
		}
	}

	return r0
}

// GetCandidateShardWaitingForCurrentRandom provides a mock function with given fields:
func (_m *BeaconCommitteeState) GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	ret := _m.Called()

	var r0 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(0).(func() []incognitokey.CommitteePublicKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]incognitokey.CommitteePublicKey)
		}
	}

	return r0
}

// GetCandidateShardWaitingForNextRandom provides a mock function with given fields:
func (_m *BeaconCommitteeState) GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	ret := _m.Called()

	var r0 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(0).(func() []incognitokey.CommitteePublicKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]incognitokey.CommitteePublicKey)
		}
	}

	return r0
}

// GetNumberOfActiveShards provides a mock function with given fields:
func (_m *BeaconCommitteeState) GetNumberOfActiveShards() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// GetOneShardCommittee provides a mock function with given fields: shardID
func (_m *BeaconCommitteeState) GetOneShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	ret := _m.Called(shardID)

	var r0 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(0).(func(byte) []incognitokey.CommitteePublicKey); ok {
		r0 = rf(shardID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]incognitokey.CommitteePublicKey)
		}
	}

	return r0
}

// GetOneShardSubstitute provides a mock function with given fields: shardID
func (_m *BeaconCommitteeState) GetOneShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey {
	ret := _m.Called(shardID)

	var r0 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(0).(func(byte) []incognitokey.CommitteePublicKey); ok {
		r0 = rf(shardID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]incognitokey.CommitteePublicKey)
		}
	}

	return r0
}

// GetRewardReceiver provides a mock function with given fields:
func (_m *BeaconCommitteeState) GetRewardReceiver() map[string]privacy.PaymentAddress {
	ret := _m.Called()

	var r0 map[string]privacy.PaymentAddress
	if rf, ok := ret.Get(0).(func() map[string]privacy.PaymentAddress); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]privacy.PaymentAddress)
		}
	}

	return r0
}

// GetShardCommittee provides a mock function with given fields:
func (_m *BeaconCommitteeState) GetShardCommittee() map[byte][]incognitokey.CommitteePublicKey {
	ret := _m.Called()

	var r0 map[byte][]incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(0).(func() map[byte][]incognitokey.CommitteePublicKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[byte][]incognitokey.CommitteePublicKey)
		}
	}

	return r0
}

// GetShardCommonPool provides a mock function with given fields:
func (_m *BeaconCommitteeState) GetShardCommonPool() []incognitokey.CommitteePublicKey {
	ret := _m.Called()

	var r0 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(0).(func() []incognitokey.CommitteePublicKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]incognitokey.CommitteePublicKey)
		}
	}

	return r0
}

// GetShardSubstitute provides a mock function with given fields:
func (_m *BeaconCommitteeState) GetShardSubstitute() map[byte][]incognitokey.CommitteePublicKey {
	ret := _m.Called()

	var r0 map[byte][]incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(0).(func() map[byte][]incognitokey.CommitteePublicKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[byte][]incognitokey.CommitteePublicKey)
		}
	}

	return r0
}

// GetStakingTx provides a mock function with given fields:
func (_m *BeaconCommitteeState) GetStakingTx() map[string]common.Hash {
	ret := _m.Called()

	var r0 map[string]common.Hash
	if rf, ok := ret.Get(0).(func() map[string]common.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]common.Hash)
		}
	}

	return r0
}

// GetSyncingValidators provides a mock function with given fields:
func (_m *BeaconCommitteeState) GetSyncingValidators() map[byte][]incognitokey.CommitteePublicKey {
	ret := _m.Called()

	var r0 map[byte][]incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(0).(func() map[byte][]incognitokey.CommitteePublicKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[byte][]incognitokey.CommitteePublicKey)
		}
	}

	return r0
}

// Hash provides a mock function with given fields: _a0
func (_m *BeaconCommitteeState) Hash(_a0 *committeestate.CommitteeChange) (*committeestate.BeaconCommitteeStateHash, error) {
	ret := _m.Called(_a0)

	var r0 *committeestate.BeaconCommitteeStateHash
	if rf, ok := ret.Get(0).(func(*committeestate.CommitteeChange) *committeestate.BeaconCommitteeStateHash); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*committeestate.BeaconCommitteeStateHash)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*committeestate.CommitteeChange) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateCommitteeState provides a mock function with given fields: env
func (_m *BeaconCommitteeState) UpdateCommitteeState(env *committeestate.BeaconCommitteeStateEnvironment) (*committeestate.BeaconCommitteeStateHash, *committeestate.CommitteeChange, [][]string, error) {
	ret := _m.Called(env)

	var r0 *committeestate.BeaconCommitteeStateHash
	if rf, ok := ret.Get(0).(func(*committeestate.BeaconCommitteeStateEnvironment) *committeestate.BeaconCommitteeStateHash); ok {
		r0 = rf(env)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*committeestate.BeaconCommitteeStateHash)
		}
	}

	var r1 *committeestate.CommitteeChange
	if rf, ok := ret.Get(1).(func(*committeestate.BeaconCommitteeStateEnvironment) *committeestate.CommitteeChange); ok {
		r1 = rf(env)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*committeestate.CommitteeChange)
		}
	}

	var r2 [][]string
	if rf, ok := ret.Get(2).(func(*committeestate.BeaconCommitteeStateEnvironment) [][]string); ok {
		r2 = rf(env)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).([][]string)
		}
	}

	var r3 error
	if rf, ok := ret.Get(3).(func(*committeestate.BeaconCommitteeStateEnvironment) error); ok {
		r3 = rf(env)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// Upgrade provides a mock function with given fields: _a0
func (_m *BeaconCommitteeState) Upgrade(_a0 *committeestate.BeaconCommitteeStateEnvironment) committeestate.BeaconCommitteeState {
	ret := _m.Called(_a0)

	var r0 committeestate.BeaconCommitteeState
	if rf, ok := ret.Get(0).(func(*committeestate.BeaconCommitteeStateEnvironment) committeestate.BeaconCommitteeState); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(committeestate.BeaconCommitteeState)
		}
	}

	return r0
}

// Version provides a mock function with given fields:
func (_m *BeaconCommitteeState) Version() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}
