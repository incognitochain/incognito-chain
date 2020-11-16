// Code generated by mockery v2.0.0. DO NOT EDIT.

package mocks

import (
	common "github.com/incognitochain/incognito-chain/common"
	incognitokey "github.com/incognitochain/incognito-chain/incognitokey"

	mock "github.com/stretchr/testify/mock"

	statedb "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

// BeaconViewRetriever is an autogenerated mock type for the BeaconViewRetriever type
type BeaconViewRetriever struct {
	mock.Mock
}

// CandidateWaitingForNextRandom provides a mock function with given fields:
func (_m *BeaconViewRetriever) CandidateWaitingForNextRandom() []incognitokey.CommitteePublicKey {
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

// GetAllBridgeTokens provides a mock function with given fields:
func (_m *BeaconViewRetriever) GetAllBridgeTokens() ([]common.Hash, error) {
	ret := _m.Called()

	var r0 []common.Hash
	if rf, ok := ret.Get(0).(func() []common.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]common.Hash)
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

// GetAllCommitteeValidatorCandidate provides a mock function with given fields:
func (_m *BeaconViewRetriever) GetAllCommitteeValidatorCandidate() (map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, error) {
	ret := _m.Called()

	var r0 map[byte][]incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(0).(func() map[byte][]incognitokey.CommitteePublicKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[byte][]incognitokey.CommitteePublicKey)
		}
	}

	var r1 map[byte][]incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(1).(func() map[byte][]incognitokey.CommitteePublicKey); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(map[byte][]incognitokey.CommitteePublicKey)
		}
	}

	var r2 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(2).(func() []incognitokey.CommitteePublicKey); ok {
		r2 = rf()
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).([]incognitokey.CommitteePublicKey)
		}
	}

	var r3 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(3).(func() []incognitokey.CommitteePublicKey); ok {
		r3 = rf()
	} else {
		if ret.Get(3) != nil {
			r3 = ret.Get(3).([]incognitokey.CommitteePublicKey)
		}
	}

	var r4 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(4).(func() []incognitokey.CommitteePublicKey); ok {
		r4 = rf()
	} else {
		if ret.Get(4) != nil {
			r4 = ret.Get(4).([]incognitokey.CommitteePublicKey)
		}
	}

	var r5 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(5).(func() []incognitokey.CommitteePublicKey); ok {
		r5 = rf()
	} else {
		if ret.Get(5) != nil {
			r5 = ret.Get(5).([]incognitokey.CommitteePublicKey)
		}
	}

	var r6 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(6).(func() []incognitokey.CommitteePublicKey); ok {
		r6 = rf()
	} else {
		if ret.Get(6) != nil {
			r6 = ret.Get(6).([]incognitokey.CommitteePublicKey)
		}
	}

	var r7 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(7).(func() []incognitokey.CommitteePublicKey); ok {
		r7 = rf()
	} else {
		if ret.Get(7) != nil {
			r7 = ret.Get(7).([]incognitokey.CommitteePublicKey)
		}
	}

	var r8 error
	if rf, ok := ret.Get(8).(func() error); ok {
		r8 = rf()
	} else {
		r8 = ret.Error(8)
	}

	return r0, r1, r2, r3, r4, r5, r6, r7, r8
}

// GetAllCommitteeValidatorCandidateFlattenListFromDatabase provides a mock function with given fields:
func (_m *BeaconViewRetriever) GetAllCommitteeValidatorCandidateFlattenListFromDatabase() ([]string, error) {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
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

// GetAutoStakingList provides a mock function with given fields:
func (_m *BeaconViewRetriever) GetAutoStakingList() map[string]bool {
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

// GetBeaconFeatureStateDB provides a mock function with given fields:
func (_m *BeaconViewRetriever) GetBeaconFeatureStateDB() *statedb.StateDB {
	ret := _m.Called()

	var r0 *statedb.StateDB
	if rf, ok := ret.Get(0).(func() *statedb.StateDB); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*statedb.StateDB)
		}
	}

	return r0
}

// GetBeaconRewardStateDB provides a mock function with given fields:
func (_m *BeaconViewRetriever) GetBeaconRewardStateDB() *statedb.StateDB {
	ret := _m.Called()

	var r0 *statedb.StateDB
	if rf, ok := ret.Get(0).(func() *statedb.StateDB); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*statedb.StateDB)
		}
	}

	return r0
}

// GetBeaconSlashStateDB provides a mock function with given fields:
func (_m *BeaconViewRetriever) GetBeaconSlashStateDB() *statedb.StateDB {
	ret := _m.Called()

	var r0 *statedb.StateDB
	if rf, ok := ret.Get(0).(func() *statedb.StateDB); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*statedb.StateDB)
		}
	}

	return r0
}
