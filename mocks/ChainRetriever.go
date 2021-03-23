// Code generated by mockery v2.6.0. DO NOT EDIT.

package mocks

import (
	chaincfg "github.com/btcsuite/btcd/chaincfg"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"

	common "github.com/incognitochain/incognito-chain/common"

	metadata "github.com/incognitochain/incognito-chain/metadata"

	mock "github.com/stretchr/testify/mock"

	privacy "github.com/incognitochain/incognito-chain/privacy"

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

// CheckBlockTimeIsReachedByBeaconHeight provides a mock function with given fields: recentBeaconHeight, beaconHeight, duration
func (_m *ChainRetriever) CheckBlockTimeIsReachedByBeaconHeight(recentBeaconHeight uint64, beaconHeight uint64, duration time.Duration) bool {
	ret := _m.Called(recentBeaconHeight, beaconHeight, duration)

	var r0 bool
	if rf, ok := ret.Get(0).(func(uint64, uint64, time.Duration) bool); ok {
		r0 = rf(recentBeaconHeight, beaconHeight, duration)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// GetBCHeightBreakPointPortalV3 provides a mock function with given fields:
func (_m *ChainRetriever) GetBCHeightBreakPointPortalV3() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
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

// GetBTCChainParams provides a mock function with given fields:
func (_m *ChainRetriever) GetBTCChainParams() *chaincfg.Params {
	ret := _m.Called()

	var r0 *chaincfg.Params
	if rf, ok := ret.Get(0).(func() *chaincfg.Params); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*chaincfg.Params)
		}
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

// GetBeaconHeightBreakPointBurnAddr provides a mock function with given fields:
func (_m *ChainRetriever) GetBeaconHeightBreakPointBurnAddr() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
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

// GetETHRemoveBridgeSigEpoch provides a mock function with given fields:
func (_m *ChainRetriever) GetETHRemoveBridgeSigEpoch() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetFixedRandomForShardIDCommitment provides a mock function with given fields: beaconHeight
func (_m *ChainRetriever) GetFixedRandomForShardIDCommitment(beaconHeight uint64) *privacy.Scalar {
	ret := _m.Called(beaconHeight)

	var r0 *privacy.Scalar
	if rf, ok := ret.Get(0).(func(uint64) *privacy.Scalar); ok {
		r0 = rf(beaconHeight)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*privacy.Scalar)
		}
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

// GetMinAmountPortalToken provides a mock function with given fields: tokenIDStr, beaconHeight, version
func (_m *ChainRetriever) GetMinAmountPortalToken(tokenIDStr string, beaconHeight uint64, version uint) (uint64, error) {
	ret := _m.Called(tokenIDStr, beaconHeight, version)

	var r0 uint64
	if rf, ok := ret.Get(0).(func(string, uint64, uint) uint64); ok {
		r0 = rf(tokenIDStr, beaconHeight, version)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, uint64, uint) error); ok {
		r1 = rf(tokenIDStr, beaconHeight, version)
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

// GetPortalReplacementAddress provides a mock function with given fields: beaconHeight
func (_m *ChainRetriever) GetPortalReplacementAddress(beaconHeight uint64) string {
	ret := _m.Called(beaconHeight)

	var r0 string
	if rf, ok := ret.Get(0).(func(uint64) string); ok {
		r0 = rf(beaconHeight)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetPortalV4MinUnshieldAmount provides a mock function with given fields: tokenIDStr, beaconHeight
func (_m *ChainRetriever) GetPortalV4MinUnshieldAmount(tokenIDStr string, beaconHeight uint64) uint64 {
	ret := _m.Called(tokenIDStr, beaconHeight)

	var r0 uint64
	if rf, ok := ret.Get(0).(func(string, uint64) uint64); ok {
		r0 = rf(tokenIDStr, beaconHeight)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetPortalV4MultiSigAddress provides a mock function with given fields: tokenIDStr, beaconHeight
func (_m *ChainRetriever) GetPortalV4MultiSigAddress(tokenIDStr string, beaconHeight uint64) string {
	ret := _m.Called(tokenIDStr, beaconHeight)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, uint64) string); ok {
		r0 = rf(tokenIDStr, beaconHeight)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetStakingAmountShard provides a mock function with given fields:
func (_m *ChainRetriever) GetStakingAmountShard() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetTransactionByHash provides a mock function with given fields: _a0
func (_m *ChainRetriever) GetTransactionByHash(_a0 common.Hash) (byte, common.Hash, uint64, int, metadata.Transaction, error) {
	ret := _m.Called(_a0)

	var r0 byte
	if rf, ok := ret.Get(0).(func(common.Hash) byte); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(byte)
	}

	var r1 common.Hash
	if rf, ok := ret.Get(1).(func(common.Hash) common.Hash); ok {
		r1 = rf(_a0)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(common.Hash)
		}
	}

	var r2 uint64
	if rf, ok := ret.Get(2).(func(common.Hash) uint64); ok {
		r2 = rf(_a0)
	} else {
		r2 = ret.Get(2).(uint64)
	}

	var r3 int
	if rf, ok := ret.Get(3).(func(common.Hash) int); ok {
		r3 = rf(_a0)
	} else {
		r3 = ret.Get(3).(int)
	}

	var r4 metadata.Transaction
	if rf, ok := ret.Get(4).(func(common.Hash) metadata.Transaction); ok {
		r4 = rf(_a0)
	} else {
		if ret.Get(4) != nil {
			r4 = ret.Get(4).(metadata.Transaction)
		}
	}

	var r5 error
	if rf, ok := ret.Get(5).(func(common.Hash) error); ok {
		r5 = rf(_a0)
	} else {
		r5 = ret.Error(5)
	}

	return r0, r1, r2, r3, r4, r5
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

// IsPortalToken provides a mock function with given fields: beaconHeight, tokenIDStr, version
func (_m *ChainRetriever) IsPortalToken(beaconHeight uint64, tokenIDStr string, version uint) (bool, error) {
	ret := _m.Called(beaconHeight, tokenIDStr, version)

	var r0 bool
	if rf, ok := ret.Get(0).(func(uint64, string, uint) bool); ok {
		r0 = rf(beaconHeight, tokenIDStr, version)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint64, string, uint) error); ok {
		r1 = rf(beaconHeight, tokenIDStr, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
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

// IsValidPortalRemoteAddress provides a mock function with given fields: tokenIDStr, remoteAddr, beaconHeight, version
func (_m *ChainRetriever) IsValidPortalRemoteAddress(tokenIDStr string, remoteAddr string, beaconHeight uint64, version uint) (bool, error) {
	ret := _m.Called(tokenIDStr, remoteAddr, beaconHeight, version)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string, uint64, uint) bool); ok {
		r0 = rf(tokenIDStr, remoteAddr, beaconHeight, version)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, uint64, uint) error); ok {
		r1 = rf(tokenIDStr, remoteAddr, beaconHeight, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListPrivacyTokenAndBridgeTokenAndPRVByShardID provides a mock function with given fields: _a0
func (_m *ChainRetriever) ListPrivacyTokenAndBridgeTokenAndPRVByShardID(_a0 byte) ([]common.Hash, error) {
	ret := _m.Called(_a0)

	var r0 []common.Hash
	if rf, ok := ret.Get(0).(func(byte) []common.Hash); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]common.Hash)
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

// ValidatePortalRemoteAddresses provides a mock function with given fields: remoteAddresses, beaconHeight, version
func (_m *ChainRetriever) ValidatePortalRemoteAddresses(remoteAddresses map[string]string, beaconHeight uint64, version uint) (bool, error) {
	ret := _m.Called(remoteAddresses, beaconHeight, version)

	var r0 bool
	if rf, ok := ret.Get(0).(func(map[string]string, uint64, uint) bool); ok {
		r0 = rf(remoteAddresses, beaconHeight, version)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(map[string]string, uint64, uint) error); ok {
		r1 = rf(remoteAddresses, beaconHeight, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
