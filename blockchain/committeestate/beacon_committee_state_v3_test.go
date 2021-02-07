package committeestate

import (
	"reflect"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate/mocks"
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBeaconCommitteeStateV3_processSwapShardInstruction(t *testing.T) {
	initTestParams()
	initLog()

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey, *incKey2, *incKey3, *incKey6,
			*incKey8, *incKey9, *incKey10, *incKey11},
		map[string]privacy.PaymentAddress{
			incKey0.GetIncKeyBase58():  paymentAddress,
			incKey.GetIncKeyBase58():   paymentAddress,
			incKey2.GetIncKeyBase58():  paymentAddress,
			incKey3.GetIncKeyBase58():  paymentAddress,
			incKey6.GetIncKeyBase58():  paymentAddress,
			incKey8.GetIncKeyBase58():  paymentAddress,
			incKey9.GetIncKeyBase58():  paymentAddress,
			incKey10.GetIncKeyBase58(): paymentAddress,
			incKey11.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key0:  true,
			key:   true,
			key2:  true,
			key3:  true,
			key6:  true,
			key8:  true,
			key9:  true,
			key10: false,
			key11: true,
		},
		map[string]common.Hash{
			key0:  *hash,
			key:   *hash,
			key2:  *hash,
			key3:  *hash,
			key6:  *hash,
			key8:  *hash,
			key9:  *hash,
			key10: *hash,
			key11: *hash,
		},
	)

	swapRule := &mocks.SwapRule{}
	swapRule.On("GenInstructions", uint8(1),
		[]string{key4, key5, key6, key7},
		[]string{key12, key13, key14, key15},
		4,
		4,
		mock.AnythingOfType("int"),
		2,
		map[string]signaturecounter.Penalty{}).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey6,
			},
			OutPublicKeys: []string{key6},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey12,
			},
			InPublicKeys: []string{key12},
			ChainID:      1,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key4, key5, key7, key12}, []string{}, []string{}, []string{key6})
	swapRule.On("Version").Return(swapRuleTestVersion)

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]incognitokey.CommitteePublicKey
	}
	type args struct {
		swapShardInstruction     *instruction.SwapShardInstruction
		env                      *BeaconCommitteeStateEnvironment
		committeeChange          *CommitteeChange
		returnStakingInstruction *instruction.ReturnStakeInstruction
		oldState                 BeaconCommitteeState
		newState                 BeaconCommitteeState
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
		want1              *instruction.ReturnStakeInstruction
		wantErr            bool
	}{
		{
			name: "[valid input]",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey4, *incKey5, *incKey6, *incKey7,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey8, *incKey9, *incKey10, *incKey11,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey12, *incKey13, *incKey14, *incKey15,
							},
						},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  false,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: true,
							key11: false,
						},
					},
					swapRule: swapRule,
				},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{*incKey16, *incKey17},
					1: []incognitokey.CommitteePublicKey{*incKey18, *incKey19},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey4, *incKey5, *incKey7, *incKey12,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey8, *incKey9, *incKey10, *incKey11,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey13, *incKey6, *incKey14, *incKey15,
							},
						},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  false,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: true,
							key11: false,
						},
					},
					swapRule: swapRule,
				},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{*incKey16, *incKey17},
					1: []incognitokey.CommitteePublicKey{*incKey18, *incKey19},
				},
			},
			args: args{
				returnStakingInstruction: &instruction.ReturnStakeInstruction{},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					RandomNumber:                     1,
					MinShardCommitteeSize:            4,
					MaxShardCommitteeSize:            4,
					NumberOfFixedShardBlockValidator: 2,
					MissingSignaturePenalty:          map[string]signaturecounter.Penalty{},
				},
				swapShardInstruction: &instruction.SwapShardInstruction{
					InPublicKeys:        []string{key12},
					InPublicKeyStructs:  []incognitokey.CommitteePublicKey{*incKey12},
					OutPublicKeys:       []string{key6},
					OutPublicKeyStructs: []incognitokey.CommitteePublicKey{*incKey6},
					ChainID:             1,
				},
				committeeChange: &CommitteeChange{
					ShardSubstituteAdded:   map[byte][]incognitokey.CommitteePublicKey{},
					ShardSubstituteRemoved: map[byte][]incognitokey.CommitteePublicKey{},
					ShardCommitteeRemoved:  map[byte][]incognitokey.CommitteePublicKey{},
					ShardCommitteeAdded:    map[byte][]incognitokey.CommitteePublicKey{},
					SlashingCommittee: map[byte][]string{
						1: []string{},
					},
				},
				oldState: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey4, *incKey5, *incKey6, *incKey7,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey8, *incKey9, *incKey10, *incKey11,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey12, *incKey13, *incKey14, *incKey15,
								},
							},
							autoStake: map[string]bool{
								key0:  true,
								key:   true,
								key2:  true,
								key3:  false,
								key6:  true,
								key8:  true,
								key9:  true,
								key10: true,
								key11: false,
							},
						},
						swapRule: swapRule,
					},
					syncPool: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{*incKey16, *incKey17},
						1: []incognitokey.CommitteePublicKey{*incKey18, *incKey19},
					},
				},
			},
			want: &CommitteeChange{
				ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{*incKey6},
				},
				ShardCommitteeRemoved: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{*incKey6},
				},
				ShardCommitteeAdded: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{*incKey12},
				},
				ShardSubstituteRemoved: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{*incKey12},
				},
				SlashingCommittee: map[byte][]string{
					1: []string{},
				},
			},
			want1:   &instruction.ReturnStakeInstruction{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
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
			if !reflect.DeepEqual(b.syncPool, tt.fieldsAfterProcess.syncPool) {
				t.Errorf("BeaconCommitteeStateV3.processSwapShardInstruction() b.syncPool = %v, tt.fieldsAfterProcess.syncPool %v", b.syncPool, tt.fieldsAfterProcess.syncPool)
			}
			if !reflect.DeepEqual(b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase) {
				t.Errorf("BeaconCommitteeStateV3.processSwapShardInstruction() b.beaconCommitteeStateSlashingBase = %v, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase %v", b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase)
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
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
		})
	}
}

func TestBeaconCommitteeStateV3_processAfterNormalSwap(t *testing.T) {
	initTestParams()
	initLog()

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{*incKey10, *incKey11, *incKey12, *incKey13},
		map[string]privacy.PaymentAddress{
			incKey10.GetIncKeyBase58(): paymentAddress,
			incKey11.GetIncKeyBase58(): paymentAddress,
			incKey12.GetIncKeyBase58(): paymentAddress,
			incKey13.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key10: true,
			key11: true,
			key12: false,
			key13: true,
		},
		map[string]common.Hash{
			key10: *hash,
			key11: *hash,
			key12: *hash,
			key13: *hash,
		},
	)

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]incognitokey.CommitteePublicKey
	}
	type args struct {
		env                      *BeaconCommitteeStateEnvironment
		outPublicKeys            []string
		committeeChange          *CommitteeChange
		returnStakingInstruction *instruction.ReturnStakeInstruction
		oldState                 BeaconCommitteeState
		newState                 BeaconCommitteeState
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
		want1              *instruction.ReturnStakeInstruction
		wantErr            bool
	}{
		{
			name: "[valid input]",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey4, *incKey5, *incKey6, *incKey7,
							},
						},
						autoStake: map[string]bool{
							key10: true,
							key11: true,
							key12: false,
							key13: true,
						},
					},
				},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{*incKey8},
					1: []incognitokey.CommitteePublicKey{*incKey9},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey11, *incKey, *incKey2, *incKey13, *incKey10, *incKey3,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey4, *incKey5, *incKey6, *incKey7,
							},
						},
						autoStake: map[string]bool{
							key10: true,
							key11: true,
							key13: true,
						},
					},
				},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{*incKey8},
					1: []incognitokey.CommitteePublicKey{*incKey9},
				},
			},
			args: args{
				returnStakingInstruction: &instruction.ReturnStakeInstruction{},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
					BeaconHeight:     500000,
					ShardID:          0,
					ActiveShards:     2,
					RandomNumber:     1,
				},
				outPublicKeys: []string{key10, key11, key12, key13},
				committeeChange: &CommitteeChange{
					ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{},
				},
				oldState: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey4, *incKey5, *incKey6, *incKey7,
								},
							},
							autoStake: map[string]bool{
								key10: true,
								key11: true,
								key12: false,
								key13: true,
							},
						},
					},
					syncPool: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{*incKey8},
						1: []incognitokey.CommitteePublicKey{*incKey9},
					},
				},
			},
			want: &CommitteeChange{
				ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{*incKey10, *incKey11, *incKey13},
				},
				RemovedStaker: []string{key12},
			},
			want1: &instruction.ReturnStakeInstruction{
				PublicKeys:       []string{key12},
				PublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey12},
				StakingTXIDs:     []string{hash.String()},
				StakingTxHashes:  []common.Hash{*hash},
				PercentReturns:   []uint{100},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
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
			if !reflect.DeepEqual(b.syncPool, tt.fieldsAfterProcess.syncPool) {
				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() b.syncPool = %v, tt.fieldsAfterProcess.syncPool %v", b.syncPool, tt.fieldsAfterProcess.syncPool)
			}
			if !reflect.DeepEqual(b.beaconCommitteeStateSlashingBase.shardSubstitute[0], tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase.shardSubstitute[0]) {
				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() b.beaconCommitteeStateSlashingBase = %v, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase %v", b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase)
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

func TestBeaconCommitteeStateV3_assignToSync(t *testing.T) {
	initTestParams()
	initLog()

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]incognitokey.CommitteePublicKey
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			if got := b.assignToSync(tt.args.shardID, tt.args.candidates, tt.args.committeeChange, tt.args.beaconHeight); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.assignToSync() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.syncPool, tt.fieldsAfterProcess.syncPool) {
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
		finishedSyncValidators           map[byte][]incognitokey.CommitteePublicKey
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
				finishedSyncValidators:           map[byte][]incognitokey.CommitteePublicKey{},
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
					finishedSyncValidators: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey0, *incKey2, *incKey3,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey, *incKey0, *incKey2, *incKey3,
						},
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
		finishedSyncValidators           map[byte][]incognitokey.CommitteePublicKey
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
				finishedSyncValidators: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3,
					},
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
				finishedSyncValidators: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3,
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
				finishedSyncValidators:           tt.fields.finishedSyncValidators,
			}
			if got := b.clone(); !reflect.DeepEqual(got.finishedSyncValidators, tt.want.finishedSyncValidators) {
				t.Errorf("BeaconCommitteeStateV3.clone() = %v, want %v", got.finishedSyncValidators, tt.want.finishedSyncValidators)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_AddFinishedSyncValidators(t *testing.T) {

	initLog()
	initTestParams()

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]incognitokey.CommitteePublicKey
		finishedSyncValidators           map[byte][]incognitokey.CommitteePublicKey
	}
	type args struct {
		syncingValidators []incognitokey.CommitteePublicKey
		shardID           byte
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
	}{
		{
			name: "Valid Input",
			fields: fields{
				syncPool: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{*incKey0, *incKey},
					1: []incognitokey.CommitteePublicKey{*incKey2, *incKey3, *incKey5},
				},
				finishedSyncValidators: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{*incKey0, *incKey},
					1: []incognitokey.CommitteePublicKey{*incKey2},
				},
			},
			fieldsAfterProcess: fields{
				finishedSyncValidators: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{*incKey0, *incKey},
					1: []incognitokey.CommitteePublicKey{*incKey2, *incKey3},
				},
			},
			args: args{
				syncingValidators: []incognitokey.CommitteePublicKey{
					*incKey2, *incKey3, *incKey4,
				},
				shardID: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
				finishedSyncValidators:           tt.fields.finishedSyncValidators,
			}
			b.AddFinishedSyncValidators(tt.args.syncingValidators, tt.args.shardID)
			if !reflect.DeepEqual(b.finishedSyncValidators, tt.fieldsAfterProcess.finishedSyncValidators) {
				t.Errorf("BeaconCommitteeStateV3.AddFinishedSyncValidators() b.finishedSyncValidators = %v, tt.fieldsAfterProcess.finishedSyncValidators %v", b.finishedSyncValidators, tt.fieldsAfterProcess.finishedSyncValidators)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_processUnstakeInstruction(t *testing.T) {

	initLog()
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
			key6:  true,
			key8:  true,
			key10: true,
			key12: true,
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
	)

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]incognitokey.CommitteePublicKey
		finishedSyncValidators           map[byte][]incognitokey.CommitteePublicKey
	}
	type args struct {
		unstakeInstruction *instruction.UnstakeInstruction
		env                *BeaconCommitteeStateEnvironment
		committeeChange    *CommitteeChange
		oldState           BeaconCommitteeState
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
	}{
		{
			name: "Valid Input",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					shardCommonPool: []incognitokey.CommitteePublicKey{*incKey0, *incKey},
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []incognitokey.CommitteePublicKey{},
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{*incKey2, *incKey3},
							1: []incognitokey.CommitteePublicKey{*incKey4, *incKey5},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{*incKey6, *incKey7},
							1: []incognitokey.CommitteePublicKey{*incKey8, *incKey9},
						},
						autoStake: map[string]bool{
							key0:  true,
							key2:  true,
							key4:  true,
							key6:  true,
							key8:  true,
							key10: true,
							key12: true,
						},
					},
				},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{*incKey10, *incKey11},
					1: []incognitokey.CommitteePublicKey{*incKey12, *incKey13},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					shardCommonPool: []incognitokey.CommitteePublicKey{*incKey0, *incKey},
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []incognitokey.CommitteePublicKey{},
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{*incKey2, *incKey3},
							1: []incognitokey.CommitteePublicKey{*incKey4, *incKey5},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{*incKey6, *incKey7},
							1: []incognitokey.CommitteePublicKey{*incKey8, *incKey9},
						},
						autoStake: map[string]bool{
							key0:  true,
							key2:  true,
							key4:  true,
							key6:  true,
							key8:  true,
							key10: true,
							key12: true,
						},
					},
				},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{*incKey10, *incKey11},
					1: []incognitokey.CommitteePublicKey{*incKey12, *incKey13},
				},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys: []string{key0, key2, key4, key6, key8, key10, key12, key14},
				},
				env: &BeaconCommitteeStateEnvironment{
					newValidators: []string{
						key0, key, key2, key3, key4, key5, key6,
						key7, key8, key9, key10, key11, key12, key13,
					},
				},
				committeeChange: &CommitteeChange{},
				oldState: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						shardCommonPool: []incognitokey.CommitteePublicKey{*incKey0, *incKey},
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{*incKey2, *incKey3},
								1: []incognitokey.CommitteePublicKey{*incKey4, *incKey5},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{*incKey6, *incKey7},
								1: []incognitokey.CommitteePublicKey{*incKey8, *incKey9},
							},
							autoStake: map[string]bool{
								key0:  true,
								key2:  true,
								key4:  true,
								key6:  true,
								key8:  true,
								key10: true,
								key12: true,
							},
						},
					},
					syncPool: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{*incKey10, *incKey11},
						1: []incognitokey.CommitteePublicKey{*incKey12, *incKey13},
					},
				},
			},
			want: &CommitteeChange{
				StopAutoStake: []string{key0, key2, key4, key6, key8, key10, key12},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
				finishedSyncValidators:           tt.fields.finishedSyncValidators,
			}
			got := b.processUnstakeInstruction(tt.args.unstakeInstruction, tt.args.env, tt.args.committeeChange, tt.args.oldState)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.processUnstakeInstruction() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_processFinishSyncInstruction(t *testing.T) {

	initLog()
	initTestParams()

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]incognitokey.CommitteePublicKey
		finishedSyncValidators           map[byte][]incognitokey.CommitteePublicKey
	}
	type args struct {
		finishSyncInstruction *instruction.FinishSyncInstruction
		env                   *BeaconCommitteeStateEnvironment
		committeeChange       *CommitteeChange
		oldState              BeaconCommitteeState
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
		wantErr            bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{},
							1: []incognitokey.CommitteePublicKey{
								*incKey4, *incKey5, *incKey6, *incKey7,
							},
						},
					},
				},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
					1: []incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey2, *incKey3},
				},
				finishedSyncValidators: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
					1: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey2,
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{},
							1: []incognitokey.CommitteePublicKey{
								*incKey4, *incKey0, *incKey2, *incKey5, *incKey6, *incKey7,
							},
						},
					},
				},
				syncPool: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
					1: []incognitokey.CommitteePublicKey{*incKey, *incKey3},
				},
				finishedSyncValidators: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
					1: []incognitokey.CommitteePublicKey{},
				},
			},
			args: args{
				finishSyncInstruction: &instruction.FinishSyncInstruction{
					ChainID: 1,
					PublicKeys: []string{
						key0, key2,
					},
					PublicKeysStruct: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey2,
					},
				},
				env: &BeaconCommitteeStateEnvironment{},
				committeeChange: &CommitteeChange{
					SyncingPoolRemoved:     map[byte][]incognitokey.CommitteePublicKey{},
					ShardSubstituteAdded:   map[byte][]incognitokey.CommitteePublicKey{},
					FinishedSyncValidators: map[byte][]string{},
				},
				oldState: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
								1: []incognitokey.CommitteePublicKey{
									*incKey4, *incKey5, *incKey6, *incKey7,
								},
							},
						},
					},
					syncPool: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
						1: []incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey2, *incKey3},
					},
					finishedSyncValidators: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
						1: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey2,
						},
					},
				},
			},
			want: &CommitteeChange{
				SyncingPoolRemoved: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey2,
					},
				},
				ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey2,
					},
				},
				FinishedSyncValidators: map[byte][]string{
					1: []string{key0, key2},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
				finishedSyncValidators:           tt.fields.finishedSyncValidators,
			}
			got, err := b.processFinishSyncInstruction(tt.args.finishSyncInstruction, tt.args.env, tt.args.committeeChange, tt.args.oldState)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV3.processFinishSyncInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.processFinishSyncInstruction() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase) {
				t.Errorf("BeaconCommitteeStateV3.beaconCommitteeStateSlashingBase = %v, want %v", b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase)
			}
			if !reflect.DeepEqual(b.syncPool, tt.fieldsAfterProcess.syncPool) {
				t.Errorf("BeaconCommitteeStateV3.syncPool = %v, want %v", b.syncPool, tt.fieldsAfterProcess.syncPool)
			}
			if !reflect.DeepEqual(b.finishedSyncValidators, tt.fieldsAfterProcess.finishedSyncValidators) {
				t.Errorf("BeaconCommitteeStateV3.finishedSyncValidators = %v, want %v", b.finishedSyncValidators, tt.fieldsAfterProcess.finishedSyncValidators)
			}
		})
	}
}
