package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
)

type BeaconCommitteeState interface {
	GetAllStaker() (map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey)
	GetBeaconCommittee() []incognitokey.CommitteePublicKey
	GetBeaconSubstitute() []incognitokey.CommitteePublicKey
	GetBeaconLocking() []incognitokey.CommitteePublicKey
	GetBeaconWaiting() []incognitokey.CommitteePublicKey
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
	AssignRuleVersion() int
	Clone(db *statedb.StateDB) BeaconCommitteeState
	UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
		*BeaconCommitteeStateHash,
		*CommitteeChange,
		[][]string,
		error)
	Upgrade(*BeaconCommitteeStateEnvironment) BeaconCommitteeState
	Hash(*CommitteeChange) (*BeaconCommitteeStateHash, error)
}

type AssignInstructionsGenerator interface {
	GenerateAssignInstructions(env *BeaconCommitteeStateEnvironment) []*instruction.AssignInstruction
}

type SwapShardInstructionsGenerator interface {
	GenerateSwapShardInstructions(env *BeaconCommitteeStateEnvironment) ([]*instruction.SwapShardInstruction, error)
}

type RandomInstructionsGenerator interface {
	GenerateRandomInstructions(env *BeaconCommitteeStateEnvironment) (*instruction.RandomInstruction, int64)
}

type SplitRewardRuleProcessor interface {
	SplitReward(environment *SplitRewardEnvironment) (map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, error)
}

type SplitRewardEnvironment struct {
	ShardID                   byte
	SubsetID                  byte
	BeaconHeight              uint64
	TotalReward               map[common.Hash]uint64
	IsSplitRewardForCustodian bool
	PercentCustodianReward    uint64
	DAOPercent                int
	ActiveShards              int
	MaxSubsetCommittees       byte
	BeaconCommittee           []incognitokey.CommitteePublicKey
	ShardCommittee            map[byte][]incognitokey.CommitteePublicKey
}

func NewSplitRewardEnvironmentMultiset(
	shardID, subsetID, maxSubsetsCommittee byte, beaconHeight uint64,
	totalReward map[common.Hash]uint64,
	isSplitRewardForCustodian bool,
	percentCustodianReward uint64,
	DAOPercent int,
	beaconCommittee []incognitokey.CommitteePublicKey,
	shardCommittee map[byte][]incognitokey.CommitteePublicKey,
) *SplitRewardEnvironment {
	return &SplitRewardEnvironment{
		ShardID:                   shardID,
		SubsetID:                  subsetID,
		BeaconHeight:              beaconHeight,
		TotalReward:               totalReward,
		IsSplitRewardForCustodian: isSplitRewardForCustodian,
		PercentCustodianReward:    percentCustodianReward,
		DAOPercent:                DAOPercent,
		MaxSubsetCommittees:       maxSubsetsCommittee,
		ShardCommittee:            shardCommittee,
		BeaconCommittee:           beaconCommittee,
	}
}
func NewSplitRewardEnvironmentV1(
	shardID byte,
	beaconHeight uint64,
	totalReward map[common.Hash]uint64,
	isSplitRewardForCustodian bool,
	percentCustodianReward uint64,
	DAOPercent int,
	activeShards int,
	beaconCommittee []incognitokey.CommitteePublicKey,
	shardCommittee map[byte][]incognitokey.CommitteePublicKey,
) *SplitRewardEnvironment {
	return &SplitRewardEnvironment{
		ShardID:                   shardID,
		BeaconHeight:              beaconHeight,
		TotalReward:               totalReward,
		IsSplitRewardForCustodian: isSplitRewardForCustodian,
		PercentCustodianReward:    percentCustodianReward,
		DAOPercent:                DAOPercent,
		ActiveShards:              activeShards,
		MaxSubsetCommittees:       1,
		ShardCommittee:            shardCommittee,
		BeaconCommittee:           beaconCommittee,
	}
}
