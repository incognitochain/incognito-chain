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
	EpochLengthV1                      uint64
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
	PreviousBlockHashes                *BeaconCommitteeStateHash
}

type BeaconCommitteeStateHash struct {
	BeaconCommitteeAndValidatorHash common.Hash
	BeaconCandidateHash             common.Hash
	ShardCandidateHash              common.Hash
	ShardCommitteeAndValidatorHash  common.Hash
	AutoStakeHash                   common.Hash
}

func NewBeaconCommitteeStateHash() *BeaconCommitteeStateHash {
	return &BeaconCommitteeStateHash{
		BeaconCommitteeAndValidatorHash: common.Hash{},
		BeaconCandidateHash:             common.Hash{},
		ShardCandidateHash:              common.Hash{},
		ShardCommitteeAndValidatorHash:  common.Hash{},
		AutoStakeHash:                   common.Hash{},
	}
}

func isNilOrBeaconCommitteeAndValidatorHash(h *BeaconCommitteeStateHash) bool {
	if h == nil {
		return true
	}
	if h.BeaconCommitteeAndValidatorHash.IsEqual(&common.Hash{}) {
		return true
	}
	return false
}

func isNilOrBeaconCandidateHash(h *BeaconCommitteeStateHash) bool {
	if h == nil {
		return true
	}
	if h.BeaconCandidateHash.IsEqual(&common.Hash{}) {
		return true
	}
	return false
}

func isNilOrShardCandidateHash(h *BeaconCommitteeStateHash) bool {
	if h == nil {
		return true
	}
	if h.ShardCandidateHash.IsEqual(&common.Hash{}) {
		return true
	}
	return false
}

func isNilOrShardCommitteeAndValidatorHash(h *BeaconCommitteeStateHash) bool {
	if h == nil {
		return true
	}
	if h.ShardCommitteeAndValidatorHash.IsEqual(&common.Hash{}) {
		return true
	}
	return false
}

func isNilOrAutoStakeHash(h *BeaconCommitteeStateHash) bool {
	if h == nil {
		return true
	}
	if h.AutoStakeHash.IsEqual(&common.Hash{}) {
		return true
	}
	return false
}

func NewBeaconCommitteeStateEnvironmentForUpdateDB(
	statedb *statedb.StateDB,
) *BeaconCommitteeStateEnvironment {
	return &BeaconCommitteeStateEnvironment{
		ConsensusStateDB: statedb,
	}
}
