package committeestate

import (
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type BeaconCommitteeStateEnvironment struct {
	BeaconHeight                      uint64
	Epoch                             uint64
	BeaconHash                        common.Hash
	ParamEpoch                        uint64
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
	NumberOfFixedBeaconBlockValidator uint64
	NumberOfFixedShardBlockValidator  int
	MissingSignaturePenalty           map[string]signaturecounter.Penalty
	allCandidateSubstituteCommittee   []string
	unassignedCommonPool              []string
	allSubstituteCommittees           []string
	LatestShardsState                 map[byte][]types.ShardState
	SwapSubType                       uint
	ShardID                           byte
	TotalReward                       map[common.Hash]uint64
	IsSplitRewardForCustodian         bool
	PercentCustodianReward            uint64
	DAOPercent                        uint64
}

type BeaconCommitteeStateHash struct {
	BeaconCommitteeAndValidatorHash common.Hash
	BeaconCandidateHash             common.Hash
	ShardCandidateHash              common.Hash
	ShardCommitteeAndValidatorHash  common.Hash
	AutoStakeHash                   common.Hash
}
