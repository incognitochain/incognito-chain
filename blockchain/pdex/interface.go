package pdex

import (
	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type State interface {
	Version() uint
	Clone() State
	Process(StateEnvironment) error
	StoreToDB(StateEnvironment, *StateChange) error
	BuildInstructions(StateEnvironment) ([][]string, error)
	TransformKeyWithNewBeaconHeight(beaconHeight uint64)
	ClearCache()
	GetDiff(State, *StateChange) (State, *StateChange, error)
	Reader() StateReader
	Validator() StateValidator
}

type StateReader interface {
	Params() *Params
	WaitingContributions() []byte
	PoolPairs() []byte
	Shares() map[string]uint64
	TradingFees() map[string]uint64
	NftIDs() map[string]uint64
	StakingPools() map[string]*StakingPoolState
	NFTAssetTags() (map[string]*common.Hash, error)
}

type StateValidator interface {
	IsValidNftID(nftID string) (bool, error)
	IsValidPoolPairID(poolPairID string) (bool, error)
	IsValidMintNftRequireAmount(amount uint64) (bool, error)
	IsValidStakingPool(stakingPoolID string) (bool, error)
	IsValidUnstakingAmount(tokenID, stakerID string, unstakingAmount uint64) (bool, error)
	IsValidShareAmount(poolPairID, lpID string, shareAmount uint64) (bool, error)
	IsValidStaker(stakingPoolID, stakerID string) (bool, error)
	IsValidAccessOTA(metadataCommon.Pdexv3ExtendedAccessID) (bool, error)
	IsValidLP(poolPairID, lpID string) (bool, error)
}
