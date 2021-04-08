package committeestate

import (
	"github.com/incognitochain/incognito-chain/instruction"
	"reflect"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

func Test_beaconCommitteeStateSlashingBase_clone(t *testing.T) {

	initTestParams()
	initLog()

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	hash, _ := common.Hash{}.NewHashFromStr("123")
	hash6, _ := common.Hash{}.NewHashFromStr("456")

	mutex := &sync.RWMutex{}

	type fields struct {
		beaconCommitteeStateBase   beaconCommitteeStateBase
		shardCommonPool            []string
		numberOfAssignedCandidates int
		swapRule                   SwapRuleProcessor
	}
	tests := []struct {
		name   string
		fields fields
		want   *beaconCommitteeStateSlashingBase
	}{
		{
			name: "[valid input] full data",
			fields: fields{
				beaconCommitteeStateBase: beaconCommitteeStateBase{
					beaconCommittee: []string{
						key6, key7, key8,
					},
					shardCommittee: map[byte][]string{
						0: []string{
							key3, key4, key5,
						},
					},
					shardSubstitute: map[byte][]string{
						0: []string{
							key,
						},
					},
					autoStake: map[string]bool{
						key:  true,
						key8: false,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey.GetIncKeyBase58(): paymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key:  *hash,
						key6: *hash6,
					},
					mu: mutex,
				},
				numberOfAssignedCandidates: 1,
				shardCommonPool: []string{
					key2,
				},
			},
			want: &beaconCommitteeStateSlashingBase{
				beaconCommitteeStateBase: beaconCommitteeStateBase{
					beaconCommittee: []string{
						key6, key7, key8,
					},
					shardCommittee: map[byte][]string{
						0: []string{
							key3, key4, key5,
						},
					},
					shardSubstitute: map[byte][]string{
						0: []string{
							key,
						},
					},
					autoStake: map[string]bool{
						key:  true,
						key8: false,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey.GetIncKeyBase58(): paymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key:  *hash,
						key6: *hash6,
					},
					mu: mutex,
				},
				numberOfAssignedCandidates: 1,
				shardCommonPool: []string{
					key2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := beaconCommitteeStateSlashingBase{
				beaconCommitteeStateBase:   tt.fields.beaconCommitteeStateBase,
				shardCommonPool:            tt.fields.shardCommonPool,
				numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				swapRule:                   tt.fields.swapRule,
			}
			if got := b.clone(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("beaconCommitteeStateV2.clone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_beaconCommitteeStateSlashingBase_processSwap(t *testing.T) {
	type fields struct {
		beaconCommitteeStateBase   beaconCommitteeStateBase
		shardCommonPool            []string
		numberOfAssignedCandidates int
		swapRule                   SwapRuleProcessor
	}
	type args struct {
		swapShardInstruction *instruction.SwapShardInstruction
		env                  *BeaconCommitteeStateEnvironment
		committeeChange      *CommitteeChange
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CommitteeChange
		want1   []string
		want2   []string
		want3   []string
		wantErr bool
	}{
		//TODO: @hung add testcase
		// Testcase 1: no swap out, only swap in
		// Testcase 2: only swap out, no swap in
		// Testcase 3: only slash, no swap in
		// Testcase 4: only slash, swap in
		// Testcase 5: only normal swap out, no swap in
		// Testcase 6: only normal swap out, swap in
		// Testcase 7: both slash and normal swap out, swap in
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &beaconCommitteeStateSlashingBase{
				beaconCommitteeStateBase:   tt.fields.beaconCommitteeStateBase,
				shardCommonPool:            tt.fields.shardCommonPool,
				numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				swapRule:                   tt.fields.swapRule,
			}
			got, got1, got2, got3, err := b.processSwap(tt.args.swapShardInstruction, tt.args.env, tt.args.committeeChange)
			if (err != nil) != tt.wantErr {
				t.Errorf("processSwap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processSwap() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("processSwap() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("processSwap() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("processSwap() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}

func Test_beaconCommitteeStateSlashingBase_processSlashing(t *testing.T) {
	type fields struct {
		beaconCommitteeStateBase   beaconCommitteeStateBase
		shardCommonPool            []string
		numberOfAssignedCandidates int
		swapRule                   SwapRuleProcessor
	}
	type args struct {
		env                      *BeaconCommitteeStateEnvironment
		slashingPublicKeys       []string
		returnStakingInstruction *instruction.ReturnStakeInstruction
		committeeChange          *CommitteeChange
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *instruction.ReturnStakeInstruction
		want1   *CommitteeChange
		wantErr bool
	}{
		//TODO: @hung add testcase
		// Testcase 1: force unstake successfully 1 node
		// Testcase 2: force unstake successfully 2 nodes
		// Testcase 3: fail to get node(s) from database
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &beaconCommitteeStateSlashingBase{
				beaconCommitteeStateBase:   tt.fields.beaconCommitteeStateBase,
				shardCommonPool:            tt.fields.shardCommonPool,
				numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				swapRule:                   tt.fields.swapRule,
			}
			got, got1, err := b.processSlashing(tt.args.env, tt.args.slashingPublicKeys, tt.args.returnStakingInstruction, tt.args.committeeChange)
			if (err != nil) != tt.wantErr {
				t.Errorf("processSlashing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processSlashing() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("processSlashing() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
