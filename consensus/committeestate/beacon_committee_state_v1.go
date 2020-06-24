package committeestate

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type BeaconCommitteeStateEnvironment struct {
	shardInstructions         [][]string
	newBeaconHeight           uint64
	epochLength               uint64
	epochBreakPointSwapNewKey []uint64
	randomNumber              uint64
}

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
