package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type BeaconCommitteeStateV1 struct {
	beaconCommittee                        []incognitokey.CommitteePublicKey
	beaconPendingValidator                 []incognitokey.CommitteePublicKey
	candidateShardWaitingForCurrentRandom  []incognitokey.CommitteePublicKey
	candidateBeaconWaitingForCurrentRandom []incognitokey.CommitteePublicKey
	candidateShardWaitingForNextRandom     []incognitokey.CommitteePublicKey
	candidateBeaconWaitingForNextRandom    []incognitokey.CommitteePublicKey
	shardCommittee                         map[byte][]incognitokey.CommitteePublicKey
	shardPendingValidator                  map[byte][]incognitokey.CommitteePublicKey
	autoStaking                            map[string]bool
	rewardReceiver                         map[string]string
}

func NewBeaconCommitteeStateV1() *BeaconCommitteeStateV1 {
	return &BeaconCommitteeStateV1{}
}

func NewBeaconCommitteeStateV1WithValue(beaconCommittee []incognitokey.CommitteePublicKey, beaconPendingValidator []incognitokey.CommitteePublicKey, candidateShardWaitingForCurrentRandom []incognitokey.CommitteePublicKey, candidateBeaconWaitingForCurrentRandom []incognitokey.CommitteePublicKey, candidateShardWaitingForNextRandom []incognitokey.CommitteePublicKey, candidateBeaconWaitingForNextRandom []incognitokey.CommitteePublicKey, shardCommittee map[byte][]incognitokey.CommitteePublicKey, shardPendingValidator map[byte][]incognitokey.CommitteePublicKey, autoStaking map[string]bool, rewardReceiver map[string]string) *BeaconCommitteeStateV1 {
	return &BeaconCommitteeStateV1{beaconCommittee: beaconCommittee, beaconPendingValidator: beaconPendingValidator, candidateShardWaitingForCurrentRandom: candidateShardWaitingForCurrentRandom, candidateBeaconWaitingForCurrentRandom: candidateBeaconWaitingForCurrentRandom, candidateShardWaitingForNextRandom: candidateShardWaitingForNextRandom, candidateBeaconWaitingForNextRandom: candidateBeaconWaitingForNextRandom, shardCommittee: shardCommittee, shardPendingValidator: shardPendingValidator, autoStaking: autoStaking, rewardReceiver: rewardReceiver}
}

func (b BeaconCommitteeStateV1) GenerateCommitteeRootHashes(shardInstruction []string) ([]common.Hash, error) {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) UpdateCommitteeState(beaconInstruction []string) (interface{}, error) {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) ValidateCommitteeRootHashes(rootHashes []common.Hash) (bool, error) {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) GetBeaconCommittee() []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) GetBeaconSubstitute() []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) GetCandidateBeaconWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) GetCandidateBeaconWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) GetShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) GetShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) GetAutoStaking() map[string]bool {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) GetRewardReceiver() map[string]string {
	panic("implement me")
}
