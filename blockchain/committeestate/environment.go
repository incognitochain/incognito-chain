package committeestate

import (
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

type BeaconCommitteeStateEnvironment struct {
	EpochLengthV1                    uint64
	BeaconHeight                     uint64
	Epoch                            uint64
	BeaconHeader                     types.BeaconHeader
	BeaconInstructions               [][]string
	BeaconHash                       common.Hash
	BestShardHash                    map[byte]common.Hash
	EpochBreakPointSwapNewKey        []uint64
	RandomNumber                     int64
	IsFoundRandomNumber              bool
	IsBeaconRandomTime               bool
	AssignOffset                     int
	DefaultOffset                    int
	SwapOffset                       int
	ActiveShards                     int
	MinShardCommitteeSize            int
	MinBeaconCommitteeSize           int
	MaxBeaconCommitteeSize           int
	MaxShardCommitteeSize            int
	ConsensusStateDB                 *statedb.StateDB
	IsReplace                        bool
	newAllRoles                      []string
	newUnassignedCommonPool          []string
	newAllSubstituteCommittees       []string
	LatestShardsState                map[byte][]types.ShardState
	SwapSubType                      uint
	ShardID                          byte
	TotalReward                      map[common.Hash]uint64
	IsSplitRewardForCustodian        bool
	PercentCustodianReward           uint64
	DAOPercent                       int
	NumberOfFixedShardBlockValidator int
	MissingSignaturePenalty          map[string]signaturecounter.Penalty
	MissingSignature                 map[string]signaturecounter.MissingSignature
	StakingV2Height                  uint64
	AssignRuleV3Height               uint64
	StakingV3Height                  uint64
	shardCommittee                   map[byte][]string
	shardSubstitute                  map[byte][]string
	numberOfValidator                []int
	PreviousBlockHashes              *BeaconCommitteeStateHash
	TriggeredFeature                 map[string]uint64
}

type BeaconCommitteeStateHash struct {
	BeaconCommitteeAndValidatorHash common.Hash
	BeaconCandidateHash             common.Hash
	ShardSyncValidatorsHash         common.Hash
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
		ShardSyncValidatorsHash:         common.Hash{},
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

func NewBeaconCommitteeStateEnvironmentForAssigningToPendingList(randomNumber int64, assignOffset int, beaconHeight uint64) *BeaconCommitteeStateEnvironment {
	return &BeaconCommitteeStateEnvironment{
		RandomNumber: randomNumber,
		AssignOffset: assignOffset,
		BeaconHeight: beaconHeight,
	}
}

//ShardCommitteeStateEnvironment :
type ShardCommitteeStateEnvironment struct {
	ShardHeight                  uint64
	ShardBlockHash               common.Hash
	GenesisBeaconHash            common.Hash
	ShardInstructions            [][]string
	BeaconInstructions           [][]string
	Txs                          []metadata.Transaction
	BeaconHeight                 uint64
	Epoch                        uint64
	EpochBreakPointSwapNewKey    []uint64
	ShardID                      byte
	MaxShardCommitteeSize        int
	MinShardCommitteeSize        int
	Offset                       int
	SwapOffset                   int
	StakingTx                    map[string]string
	NumberOfFixedBlockValidators int
	CommitteesFromBlock          common.Hash
	CommitteesFromBeaconView     []string
}

func NewShardCommitteeStateEnvironmentForAssignInstruction(
	beaconInstructions [][]string,
	shardID byte,
	numberOfFixedBlockValidators int,
	shardHeight uint64) *ShardCommitteeStateEnvironment {
	return &ShardCommitteeStateEnvironment{
		BeaconInstructions:           beaconInstructions,
		ShardID:                      shardID,
		NumberOfFixedBlockValidators: numberOfFixedBlockValidators,
		ShardHeight:                  shardHeight,
	}
}

func NewShardCommitteeStateEnvironmentForSwapInstruction(
	shardHeight uint64,
	shardID byte,
	maxShardCommitteeSize int,
	minShardCommitteeSize int,
	offset int, swapOffset int,
	numberOfFixedBlockValidators int) *ShardCommitteeStateEnvironment {
	return &ShardCommitteeStateEnvironment{
		ShardHeight:                  shardHeight,
		ShardID:                      shardID,
		MaxShardCommitteeSize:        maxShardCommitteeSize,
		MinShardCommitteeSize:        minShardCommitteeSize,
		Offset:                       offset,
		SwapOffset:                   swapOffset,
		NumberOfFixedBlockValidators: numberOfFixedBlockValidators,
	}
}
