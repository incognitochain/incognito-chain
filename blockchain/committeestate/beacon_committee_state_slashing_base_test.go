package committeestate

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

func Test_beaconCommitteeStateSlashingBase_clone(t *testing.T) {

	initTestParams()

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
					hashes: NewBeaconCommitteeStateHash(),
					mu:     mutex,
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
					hashes: NewBeaconCommitteeStateHash(),
					mu:     mutex,
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
		// this function is covered in processSwapShard unit-test
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

	initTestParams()

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey2, *incKey4, *incKey6, *incKey8, *incKey10, *incKey12,
		},
		map[string]privacy.PaymentAddress{
			incKey0.GetIncKeyBase58():  paymentAddress,
			incKey2.GetIncKeyBase58():  paymentAddress,
			incKey4.GetIncKeyBase58():  paymentAddress,
			incKey6.GetIncKeyBase58():  paymentAddress,
			incKey8.GetIncKeyBase58():  paymentAddress,
			incKey10.GetIncKeyBase58(): paymentAddress,
			incKey12.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key0:  true,
			key2:  true,
			key4:  true,
			key6:  false,
			key8:  false,
			key10: false,
			key12: false,
		},
		map[string]common.Hash{
			key0:  *hash,
			key2:  *hash,
			key4:  *hash,
			key6:  *hash,
			key8:  *hash,
			key10: *hash,
			key12: *hash,
		},
		1,
	)

	type fields struct {
		beaconCommitteeStateBase   beaconCommitteeStateBase
		shardCommonPool            []string
		numberOfAssignedCandidates int
		swapRule                   SwapRuleProcessor
	}
	type args struct {
		shardID                  byte
		env                      *BeaconCommitteeStateEnvironment
		slashingPublicKeys       []string
		returnStakingInstruction *instruction.ReturnStakeInstruction
		committeeChange          *CommitteeChange
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess *beaconCommitteeStateSlashingBase
		args               args
		want               *instruction.ReturnStakeInstruction
		want1              *CommitteeChange
		wantErr            bool
	}{
		{
			name: "force unstake successfully 1 node",
			fields: fields{
				beaconCommitteeStateBase: beaconCommitteeStateBase{
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key2: *hash,
						key4: *hash,
						key6: *hash,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress,
						incKey2.GetIncKeyBase58(): paymentAddress,
						incKey4.GetIncKeyBase58(): paymentAddress,
						incKey6.GetIncKeyBase58(): paymentAddress,
					},
					autoStake: map[string]bool{
						key0: true,
						key2: true,
						key4: true,
						key6: false,
					},
				},
			},
			fieldsAfterProcess: &beaconCommitteeStateSlashingBase{
				beaconCommitteeStateBase: beaconCommitteeStateBase{
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key2: *hash,
						key4: *hash,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress,
						incKey2.GetIncKeyBase58(): paymentAddress,
						incKey4.GetIncKeyBase58(): paymentAddress,
					},
					autoStake: map[string]bool{
						key0: true,
						key2: true,
						key4: true,
					},
				},
			},
			args: args{
				shardID: 0,
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
				},
				slashingPublicKeys: []string{
					key6,
				},
				committeeChange:          NewCommitteeChange(),
				returnStakingInstruction: instruction.NewReturnStakeIns(),
			},
			want: instruction.NewReturnStakeInsWithValue([]string{key6}, []string{hash.String()}),
			want1: NewCommitteeChange().
				AddRemovedStaker(key6).
				AddSlashingCommittees(0, []string{key6}),
			wantErr: false,
		},
		{
			name: "force unstake successfully 2 node",
			fields: fields{
				beaconCommitteeStateBase: beaconCommitteeStateBase{
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key2: *hash,
						key4: *hash,
						key6: *hash,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress,
						incKey2.GetIncKeyBase58(): paymentAddress,
						incKey4.GetIncKeyBase58(): paymentAddress,
						incKey6.GetIncKeyBase58(): paymentAddress,
					},
					autoStake: map[string]bool{
						key0: true,
						key2: true,
						key4: true,
						key6: false,
					},
				},
			},
			fieldsAfterProcess: &beaconCommitteeStateSlashingBase{
				beaconCommitteeStateBase: beaconCommitteeStateBase{
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key2: *hash,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress,
						incKey2.GetIncKeyBase58(): paymentAddress,
					},
					autoStake: map[string]bool{
						key0: true,
						key2: true,
					},
				},
			},
			args: args{
				shardID: 0,
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
				},
				slashingPublicKeys: []string{
					key6, key4,
				},
				committeeChange:          NewCommitteeChange(),
				returnStakingInstruction: instruction.NewReturnStakeIns(),
			},
			want: instruction.NewReturnStakeInsWithValue([]string{key6, key4}, []string{hash.String(), hash.String()}),
			want1: NewCommitteeChange().
				AddRemovedStaker(key6).AddRemovedStaker(key4).
				AddSlashingCommittees(0, []string{key6, key4}),
			wantErr: false,
		},
		{
			name: "fail to get node from database",
			fields: fields{
				beaconCommitteeStateBase: beaconCommitteeStateBase{
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key2: *hash,
						key4: *hash,
						key6: *hash,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress,
						incKey2.GetIncKeyBase58(): paymentAddress,
						incKey4.GetIncKeyBase58(): paymentAddress,
						incKey6.GetIncKeyBase58(): paymentAddress,
					},
					autoStake: map[string]bool{
						key0: true,
						key2: true,
						key4: true,
						key6: false,
					},
				},
			},
			fieldsAfterProcess: &beaconCommitteeStateSlashingBase{
				beaconCommitteeStateBase: beaconCommitteeStateBase{
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key2: *hash,
						key4: *hash,
						key6: *hash,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress,
						incKey2.GetIncKeyBase58(): paymentAddress,
						incKey4.GetIncKeyBase58(): paymentAddress,
						incKey6.GetIncKeyBase58(): paymentAddress,
					},
					autoStake: map[string]bool{
						key0: true,
						key2: true,
						key4: true,
						key6: false,
					},
				},
			},
			args: args{
				shardID: 0,
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
				},
				slashingPublicKeys: []string{
					key7,
				},
				committeeChange:          NewCommitteeChange(),
				returnStakingInstruction: instruction.NewReturnStakeIns(),
			},
			want:    instruction.NewReturnStakeIns(),
			want1:   NewCommitteeChange(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &beaconCommitteeStateSlashingBase{
				beaconCommitteeStateBase:   tt.fields.beaconCommitteeStateBase,
				shardCommonPool:            tt.fields.shardCommonPool,
				numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				swapRule:                   tt.fields.swapRule,
			}
			got, got1, err := b.processSlashing(tt.args.shardID, tt.args.env, tt.args.slashingPublicKeys, tt.args.returnStakingInstruction, tt.args.committeeChange)
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
