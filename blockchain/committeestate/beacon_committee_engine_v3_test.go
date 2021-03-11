package committeestate

import (
	"reflect"
	"strconv"
	"strings"
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

func TestBeaconCommitteeEngineV3_UpdateCommitteeState(t *testing.T) {
	initLog()
	initTestParams()

	finalMutex := &sync.RWMutex{}
	uncommittedMutex := &sync.RWMutex{}

	type fields struct {
		beaconCommitteeEngineSlashingBase beaconCommitteeEngineSlashingBase
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{
			*incKey6, *incKey7, *incKey17,
		},
		map[string]privacy.PaymentAddress{
			incKey6.GetIncKeyBase58():  paymentAddress,
			incKey7.GetIncKeyBase58():  paymentAddress,
			incKey17.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key6:  true,
			key7:  false,
			key17: true,
		},
		map[string]common.Hash{
			key6:  *hash,
			key7:  *hash,
			key17: *hash,
		},
	)

	randomInstructionCommitteeChange := NewCommitteeChange()
	randomInstructionCommitteeChange.NextEpochShardCandidateRemoved =
		append(randomInstructionCommitteeChange.NextEpochShardCandidateRemoved, []incognitokey.CommitteePublicKey{*incKey0, *incKey}...)
	randomInstructionCommitteeChange.SyncingPoolAdded[0] = []incognitokey.CommitteePublicKey{*incKey}
	randomInstructionCommitteeChange.SyncingPoolAdded[1] = []incognitokey.CommitteePublicKey{*incKey0}

	unstakeInstructionCommitteeChange := NewCommitteeChange()
	unstakeInstructionCommitteeChange.StopAutoStake = []string{key}

	finishSyncInstructionCommitteeChange := NewCommitteeChange()
	finishSyncInstructionCommitteeChange.FinishedSyncValidators[1] = []string{key15}
	finishSyncInstructionCommitteeChange.SyncingPoolRemoved[1] = []incognitokey.CommitteePublicKey{*incKey15}
	finishSyncInstructionCommitteeChange.ShardSubstituteAdded[1] = []incognitokey.CommitteePublicKey{*incKey15}

	swapShardInstructionCommitteeChange := NewCommitteeChange()
	swapShardInstructionCommitteeChange.RemovedStaker = []string{key7, key6}
	swapShardInstructionCommitteeChange.SlashingCommittee[1] = []string{key6}
	swapShardInstructionCommitteeChange.ShardSubstituteAdded[1] = []incognitokey.CommitteePublicKey{*incKey17}
	swapShardInstructionCommitteeChange.ShardCommitteeAdded[1] = []incognitokey.CommitteePublicKey{*incKey10, *incKey11, *incKey18}
	swapShardInstructionCommitteeChange.ShardCommitteeRemoved[1] = []incognitokey.CommitteePublicKey{*incKey6, *incKey7, *incKey17}
	swapShardInstructionCommitteeChange.ShardSubstituteRemoved[1] = []incognitokey.CommitteePublicKey{*incKey10, *incKey11, *incKey18}

	swapRule := &mocks.SwapRule{}
	swapRule.On("GenInstructions", uint8(1),
		[]string{key6, key7, key17},
		[]string{key10, key11, key18},
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int"),
		map[string]signaturecounter.Penalty{}).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey6, *incKey7, *incKey17,
			},
			OutPublicKeys: []string{key6, key7, key17},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey10, *incKey11, *incKey18,
			},
			InPublicKeys: []string{key10, key11, key18},
			ChainID:      1,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key10, key11, key18}, []string{}, []string{key6}, []string{key7, key17})
	swapRule.On("Version").Return(swapRuleTestVersion)

	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *BeaconCommitteeStateHash
		want1              *CommitteeChange
		want2              [][]string
		wantErr            bool
	}{
		{
			name: "Process Random Instruction",
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						finalState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: finalMutex,
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey4, *incKey5},
										1: []incognitokey.CommitteePublicKey{*incKey6, *incKey7},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey8, *incKey9},
										1: []incognitokey.CommitteePublicKey{*incKey10, *incKey11},
									},
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
								numberOfAssignedCandidates: 2,
							},
							syncPool: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
								1: []incognitokey.CommitteePublicKey{},
							},
						},
						uncommittedState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: uncommittedMutex,
								},
							},
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						finalState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: finalMutex,
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu:              uncommittedMutex,
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey4, *incKey5},
										1: []incognitokey.CommitteePublicKey{*incKey6, *incKey7},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey8, *incKey9},
										1: []incognitokey.CommitteePublicKey{*incKey10, *incKey11},
									},
									autoStake:      map[string]bool{},
									rewardReceiver: map[string]privacy.PaymentAddress{},
									stakingTx:      map[string]common.Hash{},
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{
									*incKey2, *incKey3,
								},
								numberOfAssignedCandidates: 0,
							},
							syncPool: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{*incKey},
								1: []incognitokey.CommitteePublicKey{*incKey0},
							},
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ActiveShards: 2,
					BeaconInstructions: [][]string{
						[]string{
							instruction.RANDOM_ACTION,
							"800000",
							"120000",
							"350000",
							"190000",
						},
					},
				},
			},
			want1:   randomInstructionCommitteeChange,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Process Unstake Instruction",
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						finalState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: finalMutex,
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey4, *incKey5},
										1: []incognitokey.CommitteePublicKey{*incKey6, *incKey7},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey8, *incKey9},
										1: []incognitokey.CommitteePublicKey{*incKey10, *incKey11},
									},
									autoStake: map[string]bool{
										key:  true,
										key7: false,
									},
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
								numberOfAssignedCandidates: 0,
							},
							syncPool: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{*incKey12, *incKey13},
								1: []incognitokey.CommitteePublicKey{*incKey14, *incKey15},
							},
						},
						uncommittedState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: uncommittedMutex,
								},
							},
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						finalState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: finalMutex,
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu:              uncommittedMutex,
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey4, *incKey5},
										1: []incognitokey.CommitteePublicKey{*incKey6, *incKey7},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey8, *incKey9},
										1: []incognitokey.CommitteePublicKey{*incKey10, *incKey11},
									},
									autoStake: map[string]bool{
										key:  false,
										key7: false,
									},
									rewardReceiver: map[string]privacy.PaymentAddress{},
									stakingTx:      map[string]common.Hash{},
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
								numberOfAssignedCandidates: 0,
							},
							syncPool: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{*incKey12, *incKey13},
								1: []incognitokey.CommitteePublicKey{*incKey14, *incKey15},
							},
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ActiveShards: 2,
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							strings.Join([]string{key, key7}, instruction.SPLITTER),
						},
					},
					newValidators: []string{
						key0, key, key2, key3, key4, key5, key6, key7, key8,
						key9, key10, key11, key12, key13, key14, key15,
					},
				},
			},
			want1:   unstakeInstructionCommitteeChange,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Process Finish Sync Instruction",
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						finalState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: finalMutex,
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey4, *incKey5},
										1: []incognitokey.CommitteePublicKey{*incKey6, *incKey7},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey8, *incKey9},
										1: []incognitokey.CommitteePublicKey{*incKey10, *incKey11},
									},
									autoStake: map[string]bool{
										key:  true,
										key7: false,
									},
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
								numberOfAssignedCandidates: 0,
							},
							syncPool: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{*incKey12, *incKey13},
								1: []incognitokey.CommitteePublicKey{*incKey14, *incKey15},
							},
						},
						uncommittedState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: uncommittedMutex,
								},
							},
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						finalState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: finalMutex,
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu:              uncommittedMutex,
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey4, *incKey5},
										1: []incognitokey.CommitteePublicKey{*incKey6, *incKey7},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey8, *incKey9},
										1: []incognitokey.CommitteePublicKey{*incKey10, *incKey15, *incKey11},
									},
									autoStake: map[string]bool{
										key:  true,
										key7: false,
									},
									rewardReceiver: map[string]privacy.PaymentAddress{},
									stakingTx:      map[string]common.Hash{},
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
								numberOfAssignedCandidates: 0,
							},
							syncPool: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{*incKey12, *incKey13},
								1: []incognitokey.CommitteePublicKey{*incKey14},
							},
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ActiveShards: 2,
					BeaconInstructions: [][]string{
						[]string{
							instruction.FINISH_SYNC_ACTION,
							"1",
							strings.Join([]string{key15}, instruction.SPLITTER),
						},
					},
				},
			},
			want1:   finishSyncInstructionCommitteeChange,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Process Swap Shard Instruction",
			fields: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						finalState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: finalMutex,
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey4, *incKey5, *incKey16},
										1: []incognitokey.CommitteePublicKey{*incKey6, *incKey7, *incKey17},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey8, *incKey9},
										1: []incognitokey.CommitteePublicKey{*incKey10, *incKey11, *incKey18},
									},
									autoStake: map[string]bool{
										key6:  true,  //slashing
										key7:  false, //swap out
										key17: true,  // return to substitute
									},
									rewardReceiver: map[string]privacy.PaymentAddress{},
									stakingTx:      map[string]common.Hash{},
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
								numberOfAssignedCandidates: 0,
								swapRule:                   swapRule,
							},
							syncPool: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{*incKey12, *incKey13},
								1: []incognitokey.CommitteePublicKey{*incKey14, *incKey15},
							},
						},
						uncommittedState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: uncommittedMutex,
								},
							},
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						finalState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu: finalMutex,
								},
							},
						},
						uncommittedState: &BeaconCommitteeStateV3{
							beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
								beaconCommitteeStateBase: beaconCommitteeStateBase{
									mu:              uncommittedMutex,
									beaconCommittee: []incognitokey.CommitteePublicKey{},
									shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey4, *incKey5},
										1: []incognitokey.CommitteePublicKey{*incKey10, *incKey11, *incKey18},
									},
									shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
										0: []incognitokey.CommitteePublicKey{*incKey8, *incKey9},
										1: []incognitokey.CommitteePublicKey{*incKey17},
									},
									autoStake: map[string]bool{
										key17: true,
									},
									rewardReceiver: map[string]privacy.PaymentAddress{},
									stakingTx:      map[string]common.Hash{},
								},
								shardCommonPool: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3,
								},
								numberOfAssignedCandidates: 0,
								swapRule:                   swapRule,
							},
							syncPool: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{*incKey12, *incKey13},
								1: []incognitokey.CommitteePublicKey{*incKey14, *incKey15},
							},
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
					ActiveShards:     2,
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							strings.Join([]string{key10, key11, key18}, instruction.SPLITTER),
							strings.Join([]string{key6, key7, key17}, instruction.SPLITTER),
							"1",
							strconv.Itoa(instruction.SWAP_BY_END_EPOCH),
						},
					},
					MinShardCommitteeSize:            4,
					MaxShardCommitteeSize:            4,
					NumberOfFixedShardBlockValidator: 0,
					MissingSignaturePenalty:          map[string]signaturecounter.Penalty{},
				},
			},
			want1: swapShardInstructionCommitteeChange,
			want2: [][]string{
				[]string{
					instruction.RETURN_ACTION,
					strings.Join([]string{key7, key6}, instruction.SPLITTER),
					strings.Join([]string{hash.String(), hash.String()}, instruction.SPLITTER),
					strings.Join([]string{"100", "100"}, instruction.SPLITTER),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &BeaconCommitteeEngineV3{
				beaconCommitteeEngineSlashingBase: tt.fields.beaconCommitteeEngineSlashingBase,
			}
			_, got1, got2, err := engine.UpdateCommitteeState(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got1.ShardSubstituteAdded, tt.want1.ShardSubstituteAdded) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got2 = %v, want %v", got2, tt.want2)
			}
			gotUncommittedState := engine.uncommittedState.(*BeaconCommitteeStateV3)
			wantUncommittedState := tt.fieldsAfterProcess.beaconCommitteeEngineSlashingBase.uncommittedState.(*BeaconCommitteeStateV3)
			if !reflect.DeepEqual(gotUncommittedState.syncPool, wantUncommittedState.syncPool) {
				Logger.log.Info()
				t.Errorf("uncommittedState got = %v, want %v", gotUncommittedState, wantUncommittedState)
			}
		})
	}
}

func TestBeaconCommitteeEngineV3_UpdateCommitteeState_MultipleInstructions(t *testing.T) {
	type fields struct {
		beaconCommitteeEngineSlashingBase beaconCommitteeEngineSlashingBase
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *BeaconCommitteeStateHash
		want1   *CommitteeChange
		want2   [][]string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &BeaconCommitteeEngineV3{
				beaconCommitteeEngineSlashingBase: tt.fields.beaconCommitteeEngineSlashingBase,
			}
			got, got1, got2, err := engine.UpdateCommitteeState(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
