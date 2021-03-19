package committeestate

import (
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
)

type BeaconCommitteeState interface {
	GetBeaconCommittee() []incognitokey.CommitteePublicKey
	GetBeaconSubstitute() []incognitokey.CommitteePublicKey
	GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey
	GetCandidateBeaconWaitingForCurrentRandom() []incognitokey.CommitteePublicKey
	GetCandidateBeaconWaitingForNextRandom() []incognitokey.CommitteePublicKey
	GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey
	GetOneShardCommittee(shardID byte) []incognitokey.CommitteePublicKey
	GetShardCommittee() map[byte][]incognitokey.CommitteePublicKey
	GetOneShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey
	GetShardSubstitute() map[byte][]incognitokey.CommitteePublicKey
	GetAutoStaking() map[string]bool
	GetStakingTx() map[string]common.Hash
	GetRewardReceiver() map[string]privacy.PaymentAddress
	GetAllCandidateSubstituteCommittee() []string
	GetNumberOfActiveShards() int
	GetShardCommonPool() []incognitokey.CommitteePublicKey
	GetSyncingValidators() map[byte][]incognitokey.CommitteePublicKey

	Version() int
	Clone() BeaconCommitteeState
	UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
		*BeaconCommitteeStateHash,
		*CommitteeChange,
		[][]string,
		error)
	Upgrade(*BeaconCommitteeStateEnvironment) BeaconCommitteeState
	Hash() (*BeaconCommitteeStateHash, error)
}

type AssignInstructionsGenerator interface {
	GenerateInstructions(env *BeaconCommitteeStateEnvironment) []*instruction.AssignInstruction
}

type SwapShardInstructionsGenerator interface {
	GenerateInstructions(env *BeaconCommitteeStateEnvironment) ([]*instruction.SwapShardInstruction, error)
}

type RandomInstructionsGenerator interface {
	GenerateIntrucstions(env *BeaconCommitteeStateEnvironment) (*instruction.RandomInstruction, int64)
}

type SplitRewardRuleProcessor interface {
	SplitReward(environment *SplitRewardEnvironment) (map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, error)
}

type BeaconCommitteeStateEnvironment struct {
	BeaconHeight                     uint64
	Epoch                            uint64
	BeaconHash                       common.Hash
	BestShardHash                    map[byte]common.Hash
	BeaconInstructions               [][]string
	EpochBreakPointSwapNewKey        []uint64
	RandomNumber                     int64
	IsFoundRandomNumber              bool
	IsBeaconRandomTime               bool
	AssignOffset                     int
	ActiveShards                     int
	MinShardCommitteeSize            int
	MaxShardCommitteeSize            int
	ConsensusStateDB                 *statedb.StateDB
	newValidators                    []string
	newUnassignedCommonPool          []string
	newAllSubstituteCommittees       []string
	shardCommittee                   map[byte][]string
	shardSubstitute                  map[byte][]string
	ShardID                          byte
	NumberOfFixedShardBlockValidator int
	MissingSignaturePenalty          map[string]signaturecounter.Penalty
	StakingV3Height                  uint64
	LatestShardsState                map[byte][]types.ShardState
}

func NewBeaconCommitteeStateEnvironment() *BeaconCommitteeStateEnvironment {
	return &BeaconCommitteeStateEnvironment{}
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

type SplitRewardEnvironment struct {
	ShardID                   byte
	BeaconHeight              uint64
	TotalReward               map[common.Hash]uint64
	IsSplitRewardForCustodian bool
	PercentCustodianReward    uint64
	DAOPercent                int
	ActiveShards              int
}

func NewSplitRewardEnvironment(shardID byte, beaconHeight uint64, totalReward map[common.Hash]uint64, isSplitRewardForCustodian bool, percentCustodianReward uint64, DAOPercent int, activeShards int) *SplitRewardEnvironment {
	return &SplitRewardEnvironment{ShardID: shardID, BeaconHeight: beaconHeight, TotalReward: totalReward, IsSplitRewardForCustodian: isSplitRewardForCustodian, PercentCustodianReward: percentCustodianReward, DAOPercent: DAOPercent, ActiveShards: activeShards}
}

type BeaconCommitteeStateHash struct {
	BeaconCommitteeAndValidatorHash common.Hash
	BeaconCandidateHash             common.Hash
	ShardCandidateHash              common.Hash
	ShardCommitteeAndValidatorHash  common.Hash
	AutoStakeHash                   common.Hash
	ShardSyncValidatorsHash         common.Hash
}
