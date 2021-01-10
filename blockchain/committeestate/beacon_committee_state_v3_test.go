package committeestate

import (
	"reflect"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
)

func TestBeaconCommitteeStateV3_processSwapShardInstruction(t *testing.T) {
	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]incognitokey.CommitteePublicKey
		terms                            map[string]uint64
	}
	type args struct {
		swapShardInstruction     *instruction.SwapShardInstruction
		env                      *BeaconCommitteeStateEnvironment
		committeeChange          *CommitteeChange
		returnStakingInstruction *instruction.ReturnStakeInstruction
		oldState                 BeaconCommitteeState
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CommitteeChange
		want1   *instruction.ReturnStakeInstruction
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
				terms:                            tt.fields.terms,
			}
			got, got1, err := b.processSwapShardInstruction(tt.args.swapShardInstruction, tt.args.env, tt.args.committeeChange, tt.args.returnStakingInstruction, tt.args.oldState)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV3.processSwapShardInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.processSwapShardInstruction() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeStateV3.processSwapShardInstruction() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_processAssignWithRandomInstruction(t *testing.T) {

	initLog()
	initTestParams()

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]incognitokey.CommitteePublicKey
		terms                            map[string]uint64
	}
	type args struct {
		rand            int64
		activeShards    int
		committeeChange *CommitteeChange
		oldState        BeaconCommitteeState
		beaconHeight    uint64
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
	}{
		{
			name: "valid input",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3,
					},
					numberOfAssignedCandidates: 2,
				},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{},
				terms:    map[string]uint64{},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey2, *incKey3,
					},
					numberOfAssignedCandidates: 0,
				},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey0,
					},
				},
				terms: map[string]uint64{
					key:  1000,
					key0: 1000,
				},
			},
			args: args{
				rand:         1000,
				activeShards: 2,
				committeeChange: &CommitteeChange{
					SyncingPoolAdded: map[byte][]incognitokey.CommitteePublicKey{},
				},
				oldState: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3,
						},
						numberOfAssignedCandidates: 2,
					},
					syncPool: map[byte][]incognitokey.CommitteePublicKey{},
					terms:    map[string]uint64{},
				},
				beaconHeight: 1000,
			},
			want: &CommitteeChange{
				NextEpochShardCandidateRemoved: []incognitokey.CommitteePublicKey{
					*incKey0, *incKey,
				},
				SyncingPoolAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{*incKey},
					1: []incognitokey.CommitteePublicKey{*incKey0},
				},
				TermsAdded: []string{
					key0, key,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
				terms:                            tt.fields.terms,
			}
			if got := b.processAssignWithRandomInstruction(tt.args.rand, tt.args.activeShards, tt.args.committeeChange, tt.args.oldState, tt.args.beaconHeight); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.processAssignWithRandomInstruction() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase) {
				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() b = %v, tt.fieldsAfterProcess %v", b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase)
			}
			if !reflect.DeepEqual(b.syncPool, tt.fieldsAfterProcess.syncPool) {
				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() b = %v, tt.fieldsAfterProcess %v", b.syncPool, tt.fieldsAfterProcess.syncPool)
			}
			if !reflect.DeepEqual(b.terms, tt.fieldsAfterProcess.terms) {
				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() b = %v, tt.fieldsAfterProcess %v", b.terms, tt.fieldsAfterProcess.terms)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_processAfterNormalSwap(t *testing.T) {
	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]incognitokey.CommitteePublicKey
		terms                            map[string]uint64
	}
	type args struct {
		env                      *BeaconCommitteeStateEnvironment
		outPublicKeys            []string
		committeeChange          *CommitteeChange
		returnStakingInstruction *instruction.ReturnStakeInstruction
		oldState                 BeaconCommitteeState
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CommitteeChange
		want1   *instruction.ReturnStakeInstruction
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
				terms:                            tt.fields.terms,
			}
			got, got1, err := b.processAfterNormalSwap(tt.args.env, tt.args.outPublicKeys, tt.args.committeeChange, tt.args.returnStakingInstruction, tt.args.oldState)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_processAssignInstruction(t *testing.T) {
	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]incognitokey.CommitteePublicKey
		terms                            map[string]uint64
	}
	type args struct {
		assignInstruction        *instruction.AssignInstruction
		env                      *BeaconCommitteeStateEnvironment
		committeeChange          *CommitteeChange
		returnStakingInstruction *instruction.ReturnStakeInstruction
		oldState                 BeaconCommitteeState
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CommitteeChange
		want1   *instruction.ReturnStakeInstruction
		wantErr bool
	}{
		/*{*/
		//name:    "Valid input",
		//fields:  fields{
		//beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{},
		//syncPool: map[byte][]incognitokey.CommitteePublicKey{},
		//terms: map[string]uint64{},
		//},
		//args:    args{
		//assignInstruction: &instruction.AssignInstruction{},
		//env: &BeaconCommitteeStateEnvironment{},
		//committeeChange: &CommitteeChange{},
		//returnStakingInstruction: &instruction.ReturnStakeInstruction{},
		//oldState: &BeaconCommitteeStateV3{},
		//},
		//want:    &CommitteeChange{},
		//want1:   &instruction.ReturnStakeInstruction{},
		//wantErr: false,
		/*},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
				terms:                            tt.fields.terms,
			}
			got, got1, err := b.processAssignInstruction(tt.args.assignInstruction, tt.args.env, tt.args.committeeChange, tt.args.returnStakingInstruction, tt.args.oldState)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV3.processAssignInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.processAssignInstruction() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeStateV3.processAssignInstruction() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_assignToPending(t *testing.T) {

	initTestParams()
	initLog()

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]incognitokey.CommitteePublicKey
		terms                            map[string]uint64
	}
	type args struct {
		candidates      []string
		rand            int64
		shardID         byte
		committeeChange *CommitteeChange
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
	}{
		{
			name: "Valid input",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{*incKey0},
							1: []incognitokey.CommitteePublicKey{*incKey},
						},
					},
				},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{},
				terms:    map[string]uint64{},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{*incKey0, *incKey2},
							1: []incognitokey.CommitteePublicKey{*incKey},
						},
					},
				},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{},
				terms:    map[string]uint64{},
			},
			args: args{
				candidates: []string{key2},
				rand:       1000,
				shardID:    0,
				committeeChange: &CommitteeChange{
					ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
				},
			},
			want: &CommitteeChange{
				ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{*incKey2},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
				terms:                            tt.fields.terms,
			}
			if got := b.assignToPending(tt.args.candidates, tt.args.rand, tt.args.shardID, tt.args.committeeChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.assignToPending() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.shardSubstitute, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase.shardSubstitute) {
				t.Errorf("BeaconCommitteeStateV3.assignToPending() b = %v, tt.fieldsAfterProcess %v", b.shardSubstitute, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase.shardSubstitute)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_assignAfterNormalSwapOut(t *testing.T) {
	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]incognitokey.CommitteePublicKey
		terms                            map[string]uint64
	}
	type args struct {
		candidates      []string
		rand            int64
		activeShards    int
		committeeChange *CommitteeChange
		oldState        BeaconCommitteeState
		oldShardID      byte
		beaconHeight    uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *CommitteeChange
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
				terms:                            tt.fields.terms,
			}
			if got := b.assignAfterNormalSwapOut(tt.args.candidates, tt.args.rand, tt.args.activeShards, tt.args.committeeChange, tt.args.oldState, tt.args.oldShardID, tt.args.beaconHeight); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.assignAfterNormalSwapOut() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_assignToSync(t *testing.T) {
	initTestParams()
	initLog()

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]incognitokey.CommitteePublicKey
		terms                            map[string]uint64
	}
	type args struct {
		shardID         byte
		candidates      []string
		committeeChange *CommitteeChange
		beaconHeight    uint64
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
	}{
		{
			name: "not empty list candidates",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey4, *incKey5, *incKey6, *incKey7,
					},
				},
				terms: map[string]uint64{
					key0: 100,
					key4: 30,
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey4, *incKey5, *incKey6, *incKey7, *incKey8, *incKey9, *incKey10,
					},
				},
				terms: map[string]uint64{
					key0:  100,
					key4:  30,
					key8:  1000,
					key9:  1000,
					key10: 1000,
				},
			},
			args: args{
				shardID:    1,
				candidates: []string{key8, key9, key10},
				committeeChange: &CommitteeChange{
					SyncingPoolAdded: map[byte][]incognitokey.CommitteePublicKey{
						1: []incognitokey.CommitteePublicKey{},
					},
				},
				beaconHeight: 1000,
			},
			want: &CommitteeChange{
				SyncingPoolAdded: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{
						*incKey8, *incKey9, *incKey10,
					},
				},
				TermsAdded: []string{key8, key9, key10},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
				terms:                            tt.fields.terms,
			}
			if got := b.assignToSync(tt.args.shardID, tt.args.candidates, tt.args.committeeChange, tt.args.beaconHeight); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.assignToSync() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.syncPool, tt.fieldsAfterProcess.syncPool) {
				t.Errorf("BeaconCommitteeStateV3.assignToSync() b = %v, tt.fieldsAfterProcess %v", b, tt.fieldsAfterProcess)
			}
			if !reflect.DeepEqual(b.terms, tt.fieldsAfterProcess.terms) {
				t.Errorf("BeaconCommitteeStateV3.assignToSync() b = %v, tt.fieldsAfterProcess %v", b, tt.fieldsAfterProcess)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_cloneFrom(t *testing.T) {
	initTestParams()
	initLog()

	mutex := &sync.RWMutex{}
	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	txHash, _ := common.Hash{}.NewHashFromStr("123")

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]incognitokey.CommitteePublicKey
		terms                            map[string]uint64
	}
	type args struct {
		fromB BeaconCommitteeStateV3
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "[valid input]",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{},
				syncPool:                         map[byte][]incognitokey.CommitteePublicKey{},
				terms:                            map[string]uint64{},
			},
			args: args{
				fromB: BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
							},
							autoStake: map[string]bool{
								key:  true,
								key0: false,
								key2: true,
								key3: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey.GetIncKeyBase58():  paymentAddress,
								incKey0.GetIncKeyBase58(): paymentAddress,
								incKey2.GetIncKeyBase58(): paymentAddress,
								incKey3.GetIncKeyBase58(): paymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key:  *txHash,
								key0: *txHash,
								key2: *txHash,
								key3: *txHash,
							},
							mu: mutex,
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey, *incKey0, *incKey2, *incKey3,
						},
						numberOfAssignedCandidates: 1,
						swapRule:                   NewSwapRuleV3(),
					},
					syncPool: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey0, *incKey2, *incKey3,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey, *incKey0, *incKey2, *incKey3,
						},
					},
					terms: map[string]uint64{
						key:  20,
						key0: 10,
						key2: 3,
						key3: 30,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
				terms:                            tt.fields.terms,
			}
			b.cloneFrom(tt.args.fromB)
		})
	}
}

func TestBeaconCommitteeStateV3_clone(t *testing.T) {
	initTestParams()
	initLog()

	mutex := &sync.RWMutex{}
	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	txHash, _ := common.Hash{}.NewHashFromStr("123")

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]incognitokey.CommitteePublicKey
		terms                            map[string]uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   *BeaconCommitteeStateV3
	}{
		{
			name: "[Valid input]",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3,
						},
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
						},
						autoStake: map[string]bool{
							key:  true,
							key0: false,
							key2: true,
							key3: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58():  paymentAddress,
							incKey0.GetIncKeyBase58(): paymentAddress,
							incKey2.GetIncKeyBase58(): paymentAddress,
							incKey3.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key:  *txHash,
							key0: *txHash,
							key2: *txHash,
							key3: *txHash,
						},
						mu: mutex,
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3,
					},
					numberOfAssignedCandidates: 1,
					swapRule:                   NewSwapRuleV3(),
				},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3,
					},
				},
				terms: map[string]uint64{
					key:  20,
					key0: 10,
					key2: 3,
					key3: 30,
				},
			},
			want: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3,
						},
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
						},
						autoStake: map[string]bool{
							key:  true,
							key0: false,
							key2: true,
							key3: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58():  paymentAddress,
							incKey0.GetIncKeyBase58(): paymentAddress,
							incKey2.GetIncKeyBase58(): paymentAddress,
							incKey3.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key:  *txHash,
							key0: *txHash,
							key2: *txHash,
							key3: *txHash,
						},
						mu: mutex,
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3,
					},
					numberOfAssignedCandidates: 1,
					swapRule:                   NewSwapRuleV3(),
				},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3,
					},
				},
				terms: map[string]uint64{
					key:  20,
					key0: 10,
					key2: 3,
					key3: 30,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
				terms:                            tt.fields.terms,
			}
			if got := b.clone(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.clone() = %v, want %v", got, tt.want)
			}
		})
	}
}
