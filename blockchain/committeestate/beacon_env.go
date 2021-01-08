package committeestate

import (
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type BeaconCommitteeStateEnvironment struct {
	BeaconHeight                       uint64
	Epoch                              uint64
	BeaconHash                         common.Hash
	BeaconInstructions                 [][]string
	EpochBreakPointSwapNewKey          []uint64
	RandomNumber                       int64
	IsFoundRandomNumber                bool
	IsBeaconRandomTime                 bool
	AssignOffset                       int
	DefaultOffset                      int
	SwapOffset                         int
	ActiveShards                       int
	MinShardCommitteeSize              int
	MinBeaconCommitteeSize             int
	MaxBeaconCommitteeSize             int
	MaxShardCommitteeSize              int
	ConsensusStateDB                   *statedb.StateDB
	IsReplace                          bool
	newAllCandidateSubstituteCommittee []string
	newUnassignedCommonPool            []string
	newAllSubstituteCommittees         []string
	LatestShardsState                  map[byte][]types.ShardState
	SwapSubType                        uint
	ShardID                            byte
	TotalReward                        map[common.Hash]uint64
	IsSplitRewardForCustodian          bool
	PercentCustodianReward             uint64
	DAOPercent                         int
	NumberOfFixedBeaconBlockValidator  uint64
	NumberOfFixedShardBlockValidator   int
	MissingSignaturePenalty            map[string]signaturecounter.Penalty
	DcsMinShardCommitteeSize           int
	DcsMaxShardCommitteeSize           int
	BeaconStateV3Height                uint64
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

func NewBeaconCommitteeStateEnvironmentForUnstakeRule(currentHeight, beaconStateV3Height uint64) *BeaconCommitteeStateEnvironment {
	return &BeaconCommitteeStateEnvironment{
		BeaconHeight:        currentHeight,
		BeaconStateV3Height: beaconStateV3Height,
	}
}

func NewBeaconCommitteeStateEnvironmentForSwapRule(currentHeight, beaconStateV3Height uint64) *BeaconCommitteeStateEnvironment {
	return &BeaconCommitteeStateEnvironment{
		BeaconHeight:        currentHeight,
		BeaconStateV3Height: beaconStateV3Height,
	}
}

func NewBeaconCommitteeStateEnvironmentForAssigningToPendingList(randomNumber int64, assignOffset int, beaconHeight uint64) *BeaconCommitteeStateEnvironment {
	return &BeaconCommitteeStateEnvironment{
		RandomNumber: randomNumber,
		AssignOffset: assignOffset,
		BeaconHeight: beaconHeight,
	}
}
