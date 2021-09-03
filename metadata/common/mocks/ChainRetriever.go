// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	common "github.com/incognitochain/incognito-chain/metadata/common"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"

	incognito_chaincommon "github.com/incognitochain/incognito-chain/common"

	mock "github.com/stretchr/testify/mock"

	time "time"
)

// ChainRetriever is an autogenerated mock type for the ChainRetriever type
type ChainRetriever struct {
	mock.Mock
}

// CheckBlockTimeIsReached provides a mock function with given fields: recentBeaconHeight, beaconHeight, recentShardHeight, shardHeight, duration
func (_m *ChainRetriever) CheckBlockTimeIsReached(recentBeaconHeight uint64, beaconHeight uint64, recentShardHeight uint64, shardHeight uint64, duration time.Duration) bool {
	ret := _m.Called(recentBeaconHeight, beaconHeight, recentShardHeight, shardHeight, duration)

	var r0 bool
	if rf, ok := ret.Get(0).(func(uint64, uint64, uint64, uint64, time.Duration) bool); ok {
		r0 = rf(recentBeaconHeight, beaconHeight, recentShardHeight, shardHeight, duration)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// GetBNBChainID provides a mock function with given fields:
func (_m *ChainRetriever) GetBNBChainID() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetBNBDataHash provides a mock function with given fields: blockHeight
func (_m *ChainRetriever) GetBNBDataHash(blockHeight int64) ([]byte, error) {
	ret := _m.Called(blockHeight)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(int64) []byte); ok {
		r0 = rf(blockHeight)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(blockHeight)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBTCChainID provides a mock function with given fields:
func (_m *ChainRetriever) GetBTCChainID() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetBTCHeaderChain provides a mock function with given fields:
func (_m *ChainRetriever) GetBTCHeaderChain() *btcrelaying.BlockChain {
	ret := _m.Called()

	var r0 *btcrelaying.BlockChain
	if rf, ok := ret.Get(0).(func() *btcrelaying.BlockChain); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*btcrelaying.BlockChain)
		}
	}

	return r0
}

// GetBurningAddress provides a mock function with given fields: blockHeight
func (_m *ChainRetriever) GetBurningAddress(blockHeight uint64) string {
	ret := _m.Called(blockHeight)

	var r0 string
	if rf, ok := ret.Get(0).(func(uint64) string); ok {
		r0 = rf(blockHeight)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetCentralizedWebsitePaymentAddress provides a mock function with given fields: _a0
func (_m *ChainRetriever) GetCentralizedWebsitePaymentAddress(_a0 uint64) string {
	ret := _m.Called(_a0)

	var r0 string
	if rf, ok := ret.Get(0).(func(uint64) string); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetLatestBNBBlkHeight provides a mock function with given fields:
func (_m *ChainRetriever) GetLatestBNBBlkHeight() (int64, error) {
	ret := _m.Called()

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMinAmountPortalToken provides a mock function with given fields: tokenIDStr, beaconHeight
func (_m *ChainRetriever) GetMinAmountPortalToken(tokenIDStr string, beaconHeight uint64) (uint64, error) {
	ret := _m.Called(tokenIDStr, beaconHeight)

	var r0 uint64
	if rf, ok := ret.Get(0).(func(string, uint64) uint64); ok {
		r0 = rf(tokenIDStr, beaconHeight)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, uint64) error); ok {
		r1 = rf(tokenIDStr, beaconHeight)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPortalETHContractAddrStr provides a mock function with given fields: beaconHeight
func (_m *ChainRetriever) GetPortalETHContractAddrStr(beaconHeight uint64) string {
	ret := _m.Called(beaconHeight)

	var r0 string
	if rf, ok := ret.Get(0).(func(uint64) string); ok {
		r0 = rf(beaconHeight)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetPortalFeederAddress provides a mock function with given fields: beaconHeight
func (_m *ChainRetriever) GetPortalFeederAddress(beaconHeight uint64) string {
	ret := _m.Called(beaconHeight)

	var r0 string
	if rf, ok := ret.Get(0).(func(uint64) string); ok {
		r0 = rf(beaconHeight)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetShardStakingTx provides a mock function with given fields: shardID, beaconHeight
func (_m *ChainRetriever) GetShardStakingTx(shardID byte, beaconHeight uint64) (map[string]string, error) {
	ret := _m.Called(shardID, beaconHeight)

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func(byte, uint64) map[string]string); ok {
		r0 = rf(shardID, beaconHeight)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(byte, uint64) error); ok {
		r1 = rf(shardID, beaconHeight)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTransactionByHash provides a mock function with given fields: _a0
func (_m *ChainRetriever) GetTransactionByHash(_a0 incognito_chaincommon.Hash) (byte, incognito_chaincommon.Hash, uint64, int, common.Transaction, error) {
	ret := _m.Called(_a0)

	var r0 byte
	if rf, ok := ret.Get(0).(func(incognito_chaincommon.Hash) byte); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(byte)
	}

	var r1 incognito_chaincommon.Hash
	if rf, ok := ret.Get(1).(func(incognito_chaincommon.Hash) incognito_chaincommon.Hash); ok {
		r1 = rf(_a0)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(incognito_chaincommon.Hash)
		}
	}

	var r2 uint64
	if rf, ok := ret.Get(2).(func(incognito_chaincommon.Hash) uint64); ok {
		r2 = rf(_a0)
	} else {
		r2 = ret.Get(2).(uint64)
	}

	var r3 int
	if rf, ok := ret.Get(3).(func(incognito_chaincommon.Hash) int); ok {
		r3 = rf(_a0)
	} else {
		r3 = ret.Get(3).(int)
	}

	var r4 common.Transaction
	if rf, ok := ret.Get(4).(func(incognito_chaincommon.Hash) common.Transaction); ok {
		r4 = rf(_a0)
	} else {
		if ret.Get(4) != nil {
			r4 = ret.Get(4).(common.Transaction)
		}
	}

	var r5 error
	if rf, ok := ret.Get(5).(func(incognito_chaincommon.Hash) error); ok {
		r5 = rf(_a0)
	} else {
		r5 = ret.Error(5)
	}

	return r0, r1, r2, r3, r4, r5
}

// IsAfterNewZKPCheckPoint provides a mock function with given fields: beaconHeight
func (_m *ChainRetriever) IsAfterNewZKPCheckPoint(beaconHeight uint64) bool {
	ret := _m.Called(beaconHeight)

	var r0 bool
	if rf, ok := ret.Get(0).(func(uint64) bool); ok {
		r0 = rf(beaconHeight)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsAfterPdexv3CheckPoint provides a mock function with given fields: beaconHeight
func (_m *ChainRetriever) IsAfterPdexv3CheckPoint(beaconHeight uint64) bool {
	ret := _m.Called(beaconHeight)

	var r0 bool
	if rf, ok := ret.Get(0).(func(uint64) bool); ok {
		r0 = rf(beaconHeight)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsAfterPrivacyV2CheckPoint provides a mock function with given fields: beaconHeight
func (_m *ChainRetriever) IsAfterPrivacyV2CheckPoint(beaconHeight uint64) bool {
	ret := _m.Called(beaconHeight)

	var r0 bool
	if rf, ok := ret.Get(0).(func(uint64) bool); ok {
		r0 = rf(beaconHeight)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsEnableFeature provides a mock function with given fields: featureFlag, epoch
func (_m *ChainRetriever) IsEnableFeature(featureFlag int, epoch uint64) bool {
	ret := _m.Called(featureFlag, epoch)

	var r0 bool
	if rf, ok := ret.Get(0).(func(int, uint64) bool); ok {
		r0 = rf(featureFlag, epoch)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsPortalExchangeRateToken provides a mock function with given fields: beaconHeight, tokenIDStr
func (_m *ChainRetriever) IsPortalExchangeRateToken(beaconHeight uint64, tokenIDStr string) bool {
	ret := _m.Called(beaconHeight, tokenIDStr)

	var r0 bool
	if rf, ok := ret.Get(0).(func(uint64, string) bool); ok {
		r0 = rf(beaconHeight, tokenIDStr)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsPortalToken provides a mock function with given fields: beaconHeight, tokenIDStr
func (_m *ChainRetriever) IsPortalToken(beaconHeight uint64, tokenIDStr string) bool {
	ret := _m.Called(beaconHeight, tokenIDStr)

	var r0 bool
	if rf, ok := ret.Get(0).(func(uint64, string) bool); ok {
		r0 = rf(beaconHeight, tokenIDStr)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsSupportedTokenCollateralV3 provides a mock function with given fields: beaconHeight, externalTokenID
func (_m *ChainRetriever) IsSupportedTokenCollateralV3(beaconHeight uint64, externalTokenID string) bool {
	ret := _m.Called(beaconHeight, externalTokenID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(uint64, string) bool); ok {
		r0 = rf(beaconHeight, externalTokenID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsValidPortalRemoteAddress provides a mock function with given fields: tokenIDStr, remoteAddr, beaconHeight
func (_m *ChainRetriever) IsValidPortalRemoteAddress(tokenIDStr string, remoteAddr string, beaconHeight uint64) (bool, error) {
	ret := _m.Called(tokenIDStr, remoteAddr, beaconHeight)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string, uint64) bool); ok {
		r0 = rf(tokenIDStr, remoteAddr, beaconHeight)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, uint64) error); ok {
		r1 = rf(tokenIDStr, remoteAddr, beaconHeight)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListPrivacyTokenAndBridgeTokenAndPRVByShardID provides a mock function with given fields: _a0
func (_m *ChainRetriever) ListPrivacyTokenAndBridgeTokenAndPRVByShardID(_a0 byte) ([]incognito_chaincommon.Hash, error) {
	ret := _m.Called(_a0)

	var r0 []incognito_chaincommon.Hash
	if rf, ok := ret.Get(0).(func(byte) []incognito_chaincommon.Hash); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]incognito_chaincommon.Hash)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(byte) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidatePortalRemoteAddresses provides a mock function with given fields: remoteAddresses, beaconHeight
func (_m *ChainRetriever) ValidatePortalRemoteAddresses(remoteAddresses map[string]string, beaconHeight uint64) (bool, error) {
	ret := _m.Called(remoteAddresses, beaconHeight)

	var r0 bool
	if rf, ok := ret.Get(0).(func(map[string]string, uint64) bool); ok {
		r0 = rf(remoteAddresses, beaconHeight)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(map[string]string, uint64) error); ok {
		r1 = rf(remoteAddresses, beaconHeight)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
