package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/key"
)

type BeaconCommitteeState interface {
	GetAllStaker() (map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey)
	GetBeaconCommittee() []incognitokey.CommitteePublicKey
	GetBeaconSubstitute() []incognitokey.CommitteePublicKey
	GetBeaconLocking() []incognitokey.CommitteePublicKey
	GetNonSlashingRewardReceiver(staker []incognitokey.CommitteePublicKey) ([]key.PaymentAddress, error)
	GetNonSlashingRewardReceiverByCPK(staker []incognitokey.CommitteePublicKey) (map[string]key.PaymentAddress, error)
	GetBeaconWaiting() []incognitokey.CommitteePublicKey
	GetUnsyncBeaconValidator() []incognitokey.CommitteePublicKey
	IsFinishSync(string) bool
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
	GetBeaconCandidateUID(candidatePK string) (string, error)
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

	GetBeaconStakerInfo(cpk string) *StakerInfo
	GetAllShardCandidateSubstituteCommittee() []string
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
	Version() int
}

type SplitRewardEnvironment struct {
	ShardID                   byte
	SubsetID                  byte
	BeaconHeight              uint64
	TotalReward               map[common.Hash]uint64
	IsSplitRewardForCustodian bool
	PercentCustodianReward    uint64
	DAOPercent                int
	CommitteePercent          int
	ActiveShards              int
	MaxSubsetCommittees       byte
	BeaconCommittee           []incognitokey.CommitteePublicKey
	ShardCommittee            map[byte][]incognitokey.CommitteePublicKey
	TotalCreditSize           uint64
	BeaconCreditSize          uint64
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
		CommitteePercent:          0,
		ActiveShards:              config.Param().ActiveShards,
		MaxSubsetCommittees:       maxSubsetsCommittee,
		BeaconCommittee:           beaconCommittee,
		ShardCommittee:            shardCommittee,
		TotalCreditSize:           0,
		BeaconCreditSize:          0,
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
		SubsetID:                  0,
		BeaconHeight:              beaconHeight,
		TotalReward:               totalReward,
		IsSplitRewardForCustodian: isSplitRewardForCustodian,
		PercentCustodianReward:    percentCustodianReward,
		DAOPercent:                DAOPercent,
		CommitteePercent:          0,
		ActiveShards:              activeShards,
		MaxSubsetCommittees:       1,
		BeaconCommittee:           beaconCommittee,
		ShardCommittee:            shardCommittee,
		TotalCreditSize:           0,
		BeaconCreditSize:          0,
	}
}

func NewSplitRewardEnvironmentForDelegation(
	beaconHeight uint64,
	totalReward map[common.Hash]uint64,
	isSplitRewardForCustodian bool,
	percentCustodianReward uint64,
	DAOPercent int,
	committeePercent int,
	activeShards int,
	totalCredit uint64,
	beaconCredit uint64,
) *SplitRewardEnvironment {
	return &SplitRewardEnvironment{
		ShardID:                   0,
		SubsetID:                  0,
		BeaconHeight:              beaconHeight,
		TotalReward:               totalReward,
		IsSplitRewardForCustodian: isSplitRewardForCustodian,
		PercentCustodianReward:    percentCustodianReward,
		DAOPercent:                DAOPercent,
		CommitteePercent:          committeePercent,
		ActiveShards:              activeShards,
		MaxSubsetCommittees:       1,
		BeaconCommittee:           []incognitokey.CommitteePublicKey{},
		ShardCommittee:            map[byte][]incognitokey.CommitteePublicKey{},
		TotalCreditSize:           totalCredit,
		BeaconCreditSize:          beaconCredit,
	}
}
