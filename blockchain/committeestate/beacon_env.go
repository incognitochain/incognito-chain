package committeestate

import (
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type BeaconCommitteeStateEnvironment struct {
	BeaconHeight                      uint64
	Epoch                             uint64
	BeaconHash                        common.Hash
	BeaconInstructions                [][]string
	EpochBreakPointSwapNewKey         []uint64
	RandomNumber                      int64
	IsFoundRandomNumber               bool
	IsBeaconRandomTime                bool
	AssignOffset                      int
	DefaultOffset                     int
	SwapOffset                        int
	ActiveShards                      int
	MinShardCommitteeSize             int
	MinBeaconCommitteeSize            int
	MaxBeaconCommitteeSize            int
	MaxShardCommitteeSize             int
	ConsensusStateDB                  *statedb.StateDB
	IsReplace                         bool
	newValidators                     []string
	newUnassignedCommonPool           []string
	newAllSubstituteCommittees        []string
	LatestShardsState                 map[byte][]types.ShardState
	SwapSubType                       uint
	ShardID                           byte
	TotalReward                       map[common.Hash]uint64
	IsSplitRewardForCustodian         bool
	PercentCustodianReward            uint64
	DAOPercent                        int
	NumberOfFixedBeaconBlockValidator uint64
	NumberOfFixedShardBlockValidator  int
	MissingSignaturePenalty           map[string]signaturecounter.Penalty
	FinishedSyncValidators            map[byte][]incognitokey.CommitteePublicKey
	DcsMinShardCommitteeSize          int
	DcsMaxShardCommitteeSize          int
	StakingV3Height                   uint64
}

type BeaconCommitteeStateHash struct {
	BeaconCommitteeAndValidatorHash common.Hash
	BeaconCandidateHash             common.Hash
	ShardCandidateHash              common.Hash
	ShardCommitteeAndValidatorHash  common.Hash
	AutoStakeHash                   common.Hash
	ShardSyncValidatorsHash         common.Hash
}

func NewBeaconCommitteeStateEnvironmentForUpdateDB(
	statedb *statedb.StateDB,
) *BeaconCommitteeStateEnvironment {
	return &BeaconCommitteeStateEnvironment{
		ConsensusStateDB: statedb,
	}
}

func NewBeaconCommitteeStateEnvironmentForSwapRule(currentHeight, stakingV3Height uint64) *BeaconCommitteeStateEnvironment {
	return &BeaconCommitteeStateEnvironment{
		BeaconHeight:    currentHeight,
		StakingV3Height: stakingV3Height,
	}
}

func NewBeaconCommitteeStateEnvironmentForAssigningToPendingList(randomNumber int64, assignOffset int, beaconHeight uint64) *BeaconCommitteeStateEnvironment {
	return &BeaconCommitteeStateEnvironment{
		RandomNumber: randomNumber,
		AssignOffset: assignOffset,
		BeaconHeight: beaconHeight,
	}
}

func NewBeaconCommitteeStateEnvironmentForUpgrading(currentHeight, stakingV3Height uint64, beaconBlockHash common.Hash) *BeaconCommitteeStateEnvironment {
	return &BeaconCommitteeStateEnvironment{
		BeaconHeight:    currentHeight,
		StakingV3Height: stakingV3Height,
		BeaconHash:      beaconBlockHash,
	}
}
