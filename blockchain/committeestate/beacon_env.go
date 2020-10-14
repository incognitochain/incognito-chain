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
	ParamEpoch                         uint64
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
	NumberOfFixedBeaconBlockValidators uint64
	NumberOfFixedShardBlockValidators  uint64
	allCandidateSubstituteCommittee    []string
	unassignedCommonPool               []string
	allSubstituteCommittees            []string
	LatestShardsState                  map[byte][]types.ShardState
	SwapSubType                        uint
	ShardID                            byte
	TotalRewardForBeacon               map[common.Hash]uint64
	TotalRewardForShard                map[common.Hash]uint64
	TotalRewardForIncDAO               map[common.Hash]uint64
	TotalRewardForCustodian            map[common.Hash]uint64
	IsSplitRewardForCustodian          bool
	PercentCustodianReward             uint64
	DAOPercent                         int
	NumberOfFixedBeaconBlockValidator  uint64
	NumberOfFixedShardBlockValidator   int
	MissingSignaturePenalty            map[string]signaturecounter.Penalty
}

type BeaconCommitteeStateHash struct {
	BeaconCommitteeAndValidatorHash common.Hash
	BeaconCandidateHash             common.Hash
	ShardCandidateHash              common.Hash
	ShardCommitteeAndValidatorHash  common.Hash
	AutoStakeHash                   common.Hash
}
