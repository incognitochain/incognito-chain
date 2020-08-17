package blockchain

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"reflect"
	"testing"
)

// TODO: @lam
// TESTCASE
// 1. RETURN SHARD-CANDIDATES,NO-BEACON-CANDIDATES,NO-ERROR
// 	INPUT: mock 1 stake shard instruction with 3 committee
// 2. RETURN NO-SHARD-CANDIDATES,BEACON-CANDIDATES,NO-ERROR
// 	INPUT: mock 1 stake beacon instruction with 3 committee
// 3. RETURN SHARD-CANDIDATES,BEACON-CANDIDATES,NO-ERROR
// 	INPUT: mock 1 stake beacon instruction with 3 committee & 1 stake shard instruction with 3 committee
// 4. RETURN NO-SHARD-CANDIDATES,NO-BEACON-CANDIDATES,ERROR NOT PASS CONDITION check
// 	len(beaconCandidatesStructs) != len(beaconRewardReceivers) && len(beaconRewardReceivers) != len(beaconAutoReStaking)
// 	len(shardCandidates) != len(shardRewardReceivers) && len(shardRewardReceivers) != len(shardAutoReStaking)
func TestBeaconBestState_processStakeInstruction(t *testing.T) {
	type fields struct {
		AutoStaking    map[string]bool
		RewardReceiver map[string]string
	}
	type args struct {
		instruction     []string
		blockchain      *BlockChain
		committeeChange *committeeChange
		autoStaking     map[string]bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   error
		want1  bool
		want2  []incognitokey.CommitteePublicKey
		want3  []incognitokey.CommitteePublicKey
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beaconBestState := &BeaconBestState{
				AutoStaking:    tt.fields.AutoStaking,
				RewardReceiver: tt.fields.RewardReceiver,
			}
			got, got1, got2, got3 := beaconBestState.processInstruction(tt.args.instruction, tt.args.blockchain, tt.args.committeeChange, tt.args.autoStaking)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processInstruction() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("processInstruction() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("processInstruction() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("processInstruction() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}
