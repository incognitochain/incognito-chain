package committeestate

import (
	"fmt"
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

func SampleCandidateList(len int) []string {
	res := []string{}
	for i := 0; i < len; i++ {
		res = append(res, fmt.Sprintf("committeepubkey%v", i))
	}
	return res
}

func GetMinMaxRange(sizeMap map[byte]int) int {
	min := -1
	max := -1
	for _, v := range sizeMap {
		if min == -1 {
			min = v
		}
		if min > v {
			min = v
		}
		if max < v {
			max = v
		}
	}
	return max - min
}

func TestBeaconCommitteeStateV2_processStakeInstruction(t *testing.T) {
	initTestParams()

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	txHash, err := common.Hash{}.NewHashFromStr("123")

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
	}
	type args struct {
		stakeInstruction *instruction.StakeInstruction
		committeeChange  *CommitteeChange
		env              *BeaconCommitteeStateEnvironment
		oldState         *BeaconCommitteeStateV2
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		want           *CommitteeChange
		wantSideEffect *fields
		wantErr        bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee:  map[byte][]incognitokey.CommitteePublicKey{},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{},
						autoStake:       map[string]bool{},
						rewardReceiver:  map[string]privacy.PaymentAddress{},
						stakingTx:       map[string]common.Hash{},
					},
				},
			},
			args: args{
				stakeInstruction: &instruction.StakeInstruction{
					PublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey,
					},
					PublicKeys: []string{key},
					RewardReceiverStructs: []privacy.PaymentAddress{
						paymentAddress,
					},
					AutoStakingFlag: []bool{true},
					TxStakeHashes: []common.Hash{
						*txHash,
					},
					TxStakes: []string{"123"},
				},
				committeeChange: &CommitteeChange{},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
				},
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []incognitokey.CommitteePublicKey{},
							shardCommittee:  map[byte][]incognitokey.CommitteePublicKey{},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{},
							autoStake:       map[string]bool{},
							rewardReceiver:  map[string]privacy.PaymentAddress{},
							stakingTx:       map[string]common.Hash{},
							mu:              &sync.RWMutex{},
						},
						shardCommonPool:            []incognitokey.CommitteePublicKey{},
						numberOfAssignedCandidates: 0,
					},
				},
			},
			want: &CommitteeChange{
				NextEpochShardCandidateAdded: []incognitokey.CommitteePublicKey{
					*incKey,
				},
			},
			wantSideEffect: &fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee:  map[byte][]incognitokey.CommitteePublicKey{},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{},
						autoStake: map[string]bool{
							key: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key: *txHash,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{*incKey},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV2{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: tt.fields.beaconCommitteeStateSlashingBase.beaconCommitteeStateBase.beaconCommittee,
						shardCommittee:  tt.fields.beaconCommitteeStateSlashingBase.beaconCommitteeStateBase.shardCommittee,
						shardSubstitute: tt.fields.beaconCommitteeStateSlashingBase.beaconCommitteeStateBase.shardSubstitute,
						autoStake:       tt.fields.beaconCommitteeStateSlashingBase.beaconCommitteeStateBase.autoStake,
						rewardReceiver:  tt.fields.beaconCommitteeStateSlashingBase.beaconCommitteeStateBase.rewardReceiver,
						stakingTx:       tt.fields.beaconCommitteeStateSlashingBase.beaconCommitteeStateBase.stakingTx,
					},
					shardCommonPool:            tt.fields.beaconCommitteeStateSlashingBase.shardCommonPool,
					numberOfAssignedCandidates: tt.fields.beaconCommitteeStateSlashingBase.numberOfAssignedCandidates,
				},
			}
			got, err := b.processStakeInstruction(
				tt.args.stakeInstruction,
				tt.args.committeeChange,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.shardCommonPool, tt.wantSideEffect.beaconCommitteeStateSlashingBase.shardCommonPool) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), shardCommonPool = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.shardCommittee, tt.wantSideEffect.beaconCommitteeStateSlashingBase.beaconCommitteeStateBase.shardCommittee) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), shardCommittee = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.shardSubstitute, tt.wantSideEffect.beaconCommitteeStateSlashingBase.beaconCommitteeStateBase.shardSubstitute) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), shardSubstitute = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.rewardReceiver, tt.wantSideEffect.beaconCommitteeStateSlashingBase.beaconCommitteeStateBase.rewardReceiver) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), rewardReceiver = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.autoStake, tt.wantSideEffect.beaconCommitteeStateSlashingBase.beaconCommitteeStateBase.autoStake) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), autoStake = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.stakingTx, tt.wantSideEffect.beaconCommitteeStateSlashingBase.beaconCommitteeStateBase.stakingTx) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), stakingTx = %v, want %v", got, tt.want)
			}
			_, has, _ := statedb.GetStakerInfo(tt.args.env.ConsensusStateDB, key)
			if has {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), StoreStakerInfo found, %+v", key)
			}
		})
	}
}

func TestBeaconCommitteeStateV2_processAssignWithRandomInstruction(t *testing.T) {

	initLog()
	initTestParams()

	committeeChangeValidInput := NewCommitteeChange()
	committeeChangeValidInput.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey2,
	}
	committeeChangeValidInput.ShardSubstituteAdded[1] = []incognitokey.CommitteePublicKey{
		*incKey2,
	}

	type fields struct {
		beaconCommitteeStateSlashingBase
	}
	type args struct {
		rand            int64
		activeShards    int
		committeeChange *CommitteeChange
		oldState        *BeaconCommitteeStateV2
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantSideEffect *fields
		want           *CommitteeChange
	}{
		{
			name: "Valid Input",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey,
								*incKey5,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey6,
							},
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey2,
					},
					numberOfAssignedCandidates: 1,
				},
			},
			args: args{
				rand:            10000,
				activeShards:    2,
				committeeChange: NewCommitteeChange(),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey,
									*incKey5,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey6,
								},
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{
							*incKey2,
						},
						numberOfAssignedCandidates: 1,
					},
				},
			},
			wantSideEffect: &fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey,
								*incKey5,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey6,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey2,
							},
						},
					},
					shardCommonPool:            []incognitokey.CommitteePublicKey{},
					numberOfAssignedCandidates: 0,
				},
			},
			want: committeeChangeValidInput,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV2{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: tt.fields.beaconCommittee,
						shardCommittee:  tt.fields.shardCommittee,
						shardSubstitute: tt.fields.shardSubstitute,
						autoStake:       tt.fields.autoStake,
						rewardReceiver:  tt.fields.rewardReceiver,
						stakingTx:       tt.fields.stakingTx,
					},
					shardCommonPool:            tt.fields.shardCommonPool,
					numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				},
			}
			if got := b.processAssignWithRandomInstruction(
				tt.args.rand,
				tt.args.activeShards,
				tt.args.committeeChange,
				tt.args.oldState,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV2.processAssignWithRandomInstruction() = %v, want %v", got, tt.want)
			}
			if b.numberOfAssignedCandidates != tt.wantSideEffect.numberOfAssignedCandidates {
				t.Errorf("BeaconCommitteeStateV2.processAssignWithRandomInstruction(), numberOfAssignedCandidates = %v, want %v", b.numberOfAssignedCandidates, tt.wantSideEffect.numberOfAssignedCandidates)
			}
			for shardID, gotV := range b.shardSubstitute {
				wantV := tt.wantSideEffect.beaconCommitteeStateBase.shardSubstitute[shardID]
				if !reflect.DeepEqual(gotV, wantV) {
					t.Errorf("BeaconCommitteeStateV2.processAssignWithRandomInstruction(), shardSubstitute = %v, want %v", gotV, wantV)
				}
			}
			for shardID, gotV := range b.shardCommittee {
				wantV := tt.wantSideEffect.beaconCommitteeStateBase.shardCommittee[shardID]
				if !reflect.DeepEqual(gotV, wantV) {
					t.Errorf("BeaconCommitteeStateV2.processAssignWithRandomInstruction(), shardSubstitute = %v, want %v", gotV, wantV)
				}
			}
		})
	}
}

func TestSnapshotShardCommonPoolV2(t *testing.T) {

	initTestParams()
	initLog()

	swapRule1 := &mocks.SwapRule{}
	swapRule1.On("AssignOffset", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(1)
	swapRule2 := &mocks.SwapRule{}
	swapRule2.On("AssignOffset", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(10)

	type args struct {
		shardCommonPool        []incognitokey.CommitteePublicKey
		shardCommittee         map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute        map[byte][]incognitokey.CommitteePublicKey
		numberOfFixedValidator int
		minCommitteeSize       int
		swapRule               SwapRule
	}
	tests := []struct {
		name                           string
		args                           args
		wantNumberOfAssignedCandidates int
	}{
		{
			name: "number of assigned candidates < number of committee in shard pool",
			args: args{
				shardCommonPool: []incognitokey.CommitteePublicKey{
					*incKey8, *incKey9, *incKey10,
				},
				shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey4, *incKey5, *incKey6, *incKey7,
					},
				},
				shardSubstitute:        map[byte][]incognitokey.CommitteePublicKey{},
				numberOfFixedValidator: 1,
				minCommitteeSize:       3,
				swapRule:               &swapRuleV2{},
			},
			wantNumberOfAssignedCandidates: 2,
		},
		{
			name: "number of assigned candidates > number of committee in shard pool",
			args: args{
				shardCommonPool: []incognitokey.CommitteePublicKey{
					*incKey8, *incKey9, *incKey10,
				},
				shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3, *incKey11, *incKey12,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey4, *incKey5, *incKey6, *incKey7, *incKey13, *incKey14,
					},
				},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey15, *incKey16,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey17, *incKey18,
					},
				},
				numberOfFixedValidator: 4,
				minCommitteeSize:       6,
				swapRule:               swapRule2,
			},
			wantNumberOfAssignedCandidates: 3,
		},
		{
			name: "First time assign candidates",
			args: args{
				shardCommonPool: []incognitokey.CommitteePublicKey{
					*incKey4,
				},
				shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3,
					},
				},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
					1: []incognitokey.CommitteePublicKey{},
				},
				numberOfFixedValidator: 0,
				minCommitteeSize:       4,
				swapRule:               swapRule1,
			},
			wantNumberOfAssignedCandidates: 1,
		},
		{
			name: "assign 0 candidates",
			args: args{
				shardCommonPool: []incognitokey.CommitteePublicKey{},
				shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey, *incKey0, *incKey2, *incKey3,
					},
				},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
					1: []incognitokey.CommitteePublicKey{},
				},
				numberOfFixedValidator: 0,
				minCommitteeSize:       4,
				swapRule:               &swapRuleV2{},
			},
			wantNumberOfAssignedCandidates: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotNumberOfAssignedCandidates := SnapshotShardCommonPoolV2(tt.args.shardCommonPool, tt.args.shardCommittee, tt.args.shardSubstitute, tt.args.numberOfFixedValidator, tt.args.minCommitteeSize, tt.args.swapRule); gotNumberOfAssignedCandidates != tt.wantNumberOfAssignedCandidates {
				t.Errorf("SnapshotShardCommonPoolV2() = %v, want %v", gotNumberOfAssignedCandidates, tt.wantNumberOfAssignedCandidates)
			}
		})
	}
}

func TestBeaconCommitteeEngineV2_GenerateAllSwapShardInstructions(t *testing.T) {

	initTestParams()
	initLog()

	emptySwapShardInstruction := &mocks.SwapRule{}
	emptySwapShardInstruction.On("GenInstructions", mock.AnythingOfType("uint8"), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(&instruction.SwapShardInstruction{}, []string{}, []string{}, []string{}, []string{})

	validInputSwapShardInstruction := &mocks.SwapRule{}
	validInputSwapShardInstruction.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			InPublicKeys: []string{
				key5,
			},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey5,
			},
			OutPublicKeys: []string{
				key,
			},
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey,
			},
			ChainID: 0,
			Type:    instruction.SWAP_BY_END_EPOCH,
		},
		[]string{}, []string{}, []string{}, []string{})

	swapRuleV3SingleShard := &mocks.SwapRule{}
	swapRuleV3SingleShard.On("GenInstructions", mock.AnythingOfType("uint8"), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			InPublicKeys: []string{
				key13, key14,
			},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey13, *incKey14,
			},
			OutPublicKeys: []string{
				key11,
			},
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey11,
			},
			ChainID: 0,
			Type:    instruction.SWAP_BY_END_EPOCH,
		},
		[]string{}, []string{}, []string{}, []string{})

	validInputSwapShardInstruction.On("GenInstructions", uint8(1), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			InPublicKeys: []string{
				key10,
			},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey10,
			},
			OutPublicKeys: []string{
				key6,
			},
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey6,
			},
			ChainID: 1,
			Type:    instruction.SWAP_BY_END_EPOCH,
		},
		[]string{}, []string{}, []string{}, []string{})

	type fields struct {
		beaconHeight                      uint64
		beaconHash                        common.Hash
		finalBeaconCommitteeStateV2       *BeaconCommitteeStateV2
		uncommittedBeaconCommitteeStateV2 *BeaconCommitteeStateV2
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*instruction.SwapShardInstruction
		wantErr bool
	}{
		{
			name: "len(subtitutes) == len(committeess) == 0",
			fields: fields{
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
							},
						},
						swapRule: emptySwapShardInstruction,
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					NumberOfFixedShardBlockValidator: 0,
					ActiveShards:                     2,
				},
			},
			want:    []*instruction.SwapShardInstruction{},
			wantErr: false,
		},
		{
			name: "Valid Input",
			fields: fields{
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey6, *incKey7, *incKey8, *incKey9,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey5,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey10,
								},
							},
						},
						swapRule: validInputSwapShardInstruction,
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					NumberOfFixedShardBlockValidator: 0,
					ActiveShards:                     2,
					MaxShardCommitteeSize:            4,
				},
			},
			want: []*instruction.SwapShardInstruction{
				&instruction.SwapShardInstruction{
					InPublicKeys: []string{
						key5,
					},
					InPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey5,
					},
					OutPublicKeys: []string{
						key,
					},
					OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey,
					},
					ChainID: 0,
					Type:    instruction.SWAP_BY_END_EPOCH,
				},
				&instruction.SwapShardInstruction{
					InPublicKeys: []string{
						key10,
					},
					InPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey10,
					},
					OutPublicKeys: []string{
						key6,
					},
					OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey6,
					},
					ChainID: 1,
					Type:    instruction.SWAP_BY_END_EPOCH,
				},
			},
			wantErr: false,
		},
		{
			name: "[Valid input] SwapruleV3 - slashing + normal + swap in + 1 shard",
			fields: fields{
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5, *incKey6, *incKey7,
									*incKey8, *incKey9, *incKey10, *incKey11, *incKey12,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey13, *incKey14, *incKey15, *incKey16,
									*incKey17, *incKey18, *incKey19,
								},
							},
						},
						swapRule: swapRuleV3SingleShard,
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key0:  signaturecounter.Penalty{},
						key11: signaturecounter.Penalty{},
					},
					MinShardCommitteeSize:            4,
					NumberOfFixedShardBlockValidator: 8,
					ActiveShards:                     1,
					MaxShardCommitteeSize:            8,
				},
			},
			want: []*instruction.SwapShardInstruction{
				&instruction.SwapShardInstruction{
					InPublicKeys: []string{
						key13, key14,
					},
					InPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey13, *incKey14,
					},
					OutPublicKeys: []string{
						key11,
					},
					OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey11,
					},
					ChainID: 0,
					Type:    instruction.SWAP_BY_END_EPOCH,
				},
			},
			wantErr: false,
		},
		{
			name: "[Valid input] SwapruleV3 - slashing + normal + swap in + 2 shards",
			fields: fields{
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5, *incKey6, *incKey7,
									*incKey8, *incKey9, *incKey10, *incKey11, *incKey12,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5, *incKey6, *incKey7,
									*incKey8, *incKey9, *incKey10, *incKey11, *incKey12,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey13, *incKey14, *incKey15, *incKey16,
									*incKey17, *incKey18, *incKey19,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey13, *incKey14, *incKey15, *incKey16,
									*incKey17, *incKey18, *incKey19,
								},
							},
						},
						swapRule: swapRuleV3SingleShard,
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key0:  signaturecounter.Penalty{},
						key11: signaturecounter.Penalty{},
					},
					MinShardCommitteeSize:            4,
					NumberOfFixedShardBlockValidator: 8,
					ActiveShards:                     2,
					MaxShardCommitteeSize:            8,
				},
			},
			want: []*instruction.SwapShardInstruction{
				&instruction.SwapShardInstruction{
					InPublicKeys: []string{
						key13, key14,
					},
					InPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey13, *incKey14,
					},
					OutPublicKeys: []string{
						key11,
					},
					OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey11,
					},
					ChainID: 0,
					Type:    instruction.SWAP_BY_END_EPOCH,
				},
				&instruction.SwapShardInstruction{
					InPublicKeys: []string{
						key13, key14,
					},
					InPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey13, *incKey14,
					},
					OutPublicKeys: []string{
						key11,
					},
					OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey11,
					},
					ChainID: 0,
					Type:    instruction.SWAP_BY_END_EPOCH,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &BeaconCommitteeEngineV2{
				beaconCommitteeEngineSlashingBase: beaconCommitteeEngineSlashingBase{
					beaconCommitteeEngineBase: beaconCommitteeEngineBase{
						beaconHeight:     tt.fields.beaconHeight,
						beaconHash:       tt.fields.beaconHash,
						finalState:       tt.fields.finalBeaconCommitteeStateV2,
						uncommittedState: tt.fields.uncommittedBeaconCommitteeStateV2,
					},
				},
			}
			got, err := engine.GenerateAllSwapShardInstructions(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV2.GenerateAllRequestShardSwapInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, v := range got {
				if !reflect.DeepEqual(*v, *tt.want[i]) {
					t.Errorf("*v = %v, want %v", *v, *tt.want[i])
					return
				}
			}
		})
	}
}

func TestBeaconCommitteeStateV2_processSwapShardInstruction(t *testing.T) {

	initTestParams()
	initLog()

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})

	hash, _ := common.Hash{}.NewHashFromStr("123")
	hash6, _ := common.Hash{}.NewHashFromStr("456")
	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{*incKey, *incKey6, *incKey11},
		map[string]privacy.PaymentAddress{
			incKey.GetIncKeyBase58():   paymentAddress,
			incKey6.GetIncKeyBase58():  paymentAddress,
			incKey11.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key:   true,
			key6:  false,
			key11: true,
		},
		map[string]common.Hash{
			key:   *hash,
			key6:  *hash6,
			key11: *hash,
		},
	)

	rootHash, _ := sDB.Commit(true)
	sDB.Database().TrieDB().Commit(rootHash, false)

	committeeChangeValidInputSwapOut := NewCommitteeChange()
	committeeChangeValidInputSwapOut.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeValidInputSwapOut.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeValidInputSwapOut.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey6,
	}
	committeeChangeValidInputSwapOut.RemovedStaker = []string{key6}
	committeeChangeValidInputSwapOut.TermsRemoved = []string{}

	committeeChangeValidInputSwapOut2 := NewCommitteeChange()
	committeeChangeValidInputSwapOut2.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}
	committeeChangeValidInputSwapOut2.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}
	committeeChangeValidInputSwapOut2.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey6,
	}
	committeeChangeValidInputSwapOut2.RemovedStaker = []string{key6}
	committeeChangeValidInputSwapOut2.TermsRemoved = []string{}

	committeeChangeValidInputBackToSub := NewCommitteeChange()
	committeeChangeValidInputBackToSub.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}
	committeeChangeValidInputBackToSub.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeValidInputBackToSub.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeValidInputBackToSub.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}

	committeeChangeValidInputBackToSub2 := NewCommitteeChange()
	committeeChangeValidInputBackToSub2.ShardSubstituteAdded[1] = []incognitokey.CommitteePublicKey{
		*incKey,
	}
	committeeChangeValidInputBackToSub2.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeValidInputBackToSub2.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeValidInputBackToSub2.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}

	committeeChangeSlashingForceSwapOut := NewCommitteeChange()
	committeeChangeSlashingForceSwapOut.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeSlashingForceSwapOut.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeSlashingForceSwapOut.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}
	committeeChangeSlashingForceSwapOut.RemovedStaker = []string{key}
	committeeChangeSlashingForceSwapOut.SlashingCommittee[0] = []string{key}
	committeeChangeSlashingForceSwapOut.TermsRemoved = []string{}

	committeeChangeSwapRuleV3 := NewCommitteeChange()
	committeeChangeSwapRuleV3.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey13, *incKey14,
	}
	committeeChangeSwapRuleV3.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey13, *incKey14,
	}
	committeeChangeSwapRuleV3.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey11,
	}
	committeeChangeSwapRuleV3.RemovedStaker = []string{key11}
	committeeChangeSwapRuleV3.SlashingCommittee[0] = []string{key11}
	committeeChangeSwapRuleV3.TermsRemoved = []string{}

	//Define swap rule mock here
	swapRuleSingleInstructionOut := &mocks.SwapRule{}
	swapRuleSingleInstructionOut.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey6,
			},
			OutPublicKeys: []string{key6},
			ChainID:       0,
			Type:          instruction.SWAP_BY_END_EPOCH,
		},
		[]string{}, []string{}, []string{}, []string{})

	swapRuleSingleInstructionIn := &mocks.SwapRule{}
	swapRuleSingleInstructionIn.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey6,
			},
			InPublicKeys: []string{key6},
			ChainID:      0,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{}, []string{}, []string{}, []string{})

	swapRule1 := &mocks.SwapRule{}
	swapRule1.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		instruction.NewSwapShardInstructionWithValue(
			[]string{key5},
			[]string{key7},
			0,
			instruction.SWAP_BY_END_EPOCH,
		),
		[]string{}, []string{}, []string{}, []string{})

	swapRule2 := &mocks.SwapRule{}
	swapRule2.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		instruction.NewSwapShardInstructionWithValue(
			[]string{key5},
			[]string{key},
			0,
			instruction.SWAP_BY_END_EPOCH,
		),
		[]string{key2, key3, key4, key5}, []string{}, []string{}, []string{key})

	swapRule3 := &mocks.SwapRule{}
	swapRule3.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		instruction.NewSwapShardInstructionWithValue(
			[]string{key5},
			[]string{key6},
			0,
			instruction.SWAP_BY_END_EPOCH,
		),
		[]string{key2, key3, key4, key5}, []string{}, []string{}, []string{key6})

	swapRule4 := &mocks.SwapRule{}
	swapRule4.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		instruction.NewSwapShardInstructionWithValue(
			[]string{key},
			[]string{key6},
			0,
			instruction.SWAP_BY_END_EPOCH,
		),
		[]string{key2, key3, key4, key}, []string{}, []string{}, []string{key6})

	swapRule4.On("GenInstructions", uint8(1), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		instruction.NewSwapShardInstructionWithValue(
			[]string{key},
			[]string{key6},
			0,
			instruction.SWAP_BY_END_EPOCH,
		),
		[]string{key7, key8, key9, key10}, []string{}, []string{}, []string{})

	swapRule5 := &mocks.SwapRule{}
	swapRule5.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		instruction.NewSwapShardInstructionWithValue(
			[]string{key5},
			[]string{key},
			0,
			instruction.SWAP_BY_END_EPOCH,
		),
		[]string{key6, key2, key3, key5}, []string{}, []string{key}, []string{})

	swapRuleV3 := &mocks.SwapRule{}
	swapRuleV3.On("GenInstructions", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		instruction.NewSwapShardInstructionWithValue(
			[]string{key13, key14},
			[]string{key11},
			0,
			instruction.SWAP_BY_END_EPOCH,
		),
		[]string{
			key0, key, key2, key3, key4, key5, key6, key7,
			key8, key9, key10, key12, key13, key14,
		}, []string{}, []string{key11}, []string{})

	unstakeRule1 := &mocks.UnstakeRule{}
	unstakeRule1.On("RemoveFromState",
		*incKey6,
		map[string]bool{},
		map[string]privacy.PaymentAddress{},
		map[string]common.Hash{},
		map[string]uint64{},
		[]string(nil),
		[]string(nil)).
		Return(
			map[string]bool{},
			map[string]privacy.PaymentAddress{},
			map[string]common.Hash{},
			[]string{
				key6,
			},
			[]string{},
			nil,
		)

	unstakeRule2 := &mocks.UnstakeRule{}
	unstakeRule2.On("RemoveFromState",
		*incKey,
		map[string]bool{},
		map[string]privacy.PaymentAddress{},
		map[string]common.Hash{},
		map[string]uint64{},
		[]string(nil),
		[]string(nil)).
		Return(
			map[string]bool{},
			map[string]privacy.PaymentAddress{},
			map[string]common.Hash{},
			[]string{
				key,
			},
			[]string{},
			nil,
		)

	unstakeRule3 := &mocks.UnstakeRule{}
	unstakeRule3.On("RemoveFromState",
		*incKey11,
		map[string]bool{},
		map[string]privacy.PaymentAddress{},
		map[string]common.Hash{},
		map[string]uint64{},
		[]string(nil),
		[]string(nil)).
		Return(
			map[string]bool{},
			map[string]privacy.PaymentAddress{},
			map[string]common.Hash{},
			[]string{
				key11,
			},
			[]string{},
			nil,
		)

	type fields struct {
		beaconCommitteeStateSlashingBase
	}

	type committeeAfterProcess struct {
		shardCommittee  map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute map[byte][]incognitokey.CommitteePublicKey
	}

	type args struct {
		swapShardInstruction      *instruction.SwapShardInstruction
		returnStakingInstructions *instruction.ReturnStakeInstruction
		env                       *BeaconCommitteeStateEnvironment
		committeeChange           *CommitteeChange
		oldState                  *BeaconCommitteeStateV2
	}
	tests := []struct {
		name                  string
		committeeAfterProcess committeeAfterProcess
		fields                fields
		args                  args
		want1                 *CommitteeChange
		want2                 *instruction.ReturnStakeInstruction
		wantErr               bool
	}{
		{
			name: "Swap Out Not Valid In List Committees Public Key",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey, *incKey2, *incKey3, *incKey4,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey5,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					swapRule: swapRuleSingleInstructionOut,
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey, *incKey2, *incKey3, *incKey4,
					},
				},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey5,
					},
				},
			},
			args: args{
				swapShardInstruction: &instruction.SwapShardInstruction{
					OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey6,
					},
				},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					NumberOfFixedShardBlockValidator: 0,
					MaxShardCommitteeSize:            4,
					ActiveShards:                     1,
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey5,
								},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						swapRule: swapRuleSingleInstructionOut,
					},
				},
			},
			want1:   nil,
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: true,
		},
		{
			name: "Swap In Not Valid In List Substitutes Public Key",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey, *incKey2, *incKey3, *incKey4,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey5,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					swapRule: swapRuleSingleInstructionIn,
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey, *incKey2, *incKey3, *incKey4,
					},
				},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey5,
					},
				},
			},
			args: args{
				swapShardInstruction: &instruction.SwapShardInstruction{
					InPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey6,
					},
				},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					NumberOfFixedShardBlockValidator: 0,
					MaxShardCommitteeSize:            4,
					ActiveShards:                     1,
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey5,
								},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						swapRule: swapRuleSingleInstructionIn,
					},
				},
			},
			want1:   nil,
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: true,
		},
		{
			name: "Valid Input [Back To Common Pool And Re-assign]",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey, *incKey2, *incKey3, *incKey4,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey5,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					swapRule: swapRule2,
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey2, *incKey3, *incKey4, *incKey5,
					},
				},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey,
					},
				},
			},
			args: args{
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key5},
					[]string{key},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				env: &BeaconCommitteeStateEnvironment{
					NumberOfFixedShardBlockValidator: 0,
					ConsensusStateDB:                 sDB,
					RandomNumber:                     5000,
					ActiveShards:                     1,
					MaxShardCommitteeSize:            4,
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey5,
								},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						swapRule: swapRule2,
					},
				},
			},
			want1:   committeeChangeValidInputBackToSub,
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: false,
		},
		{
			name: "Valid Input [Back To Common Pool And Re-assign], Data in 2 Shard",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey, *incKey2, *incKey3, *incKey4,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey7, *incKey8, *incKey9, *incKey10,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey5,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey11, *incKey12,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					swapRule: swapRule2,
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey2, *incKey3, *incKey4, *incKey5,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey7, *incKey8, *incKey9, *incKey10,
					},
				},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
					1: []incognitokey.CommitteePublicKey{
						*incKey11, *incKey12, *incKey,
					},
				},
			},
			args: args{
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key5},
					[]string{key},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				env: &BeaconCommitteeStateEnvironment{
					NumberOfFixedShardBlockValidator: 0,
					ConsensusStateDB:                 sDB,
					RandomNumber:                     5000,
					ActiveShards:                     2,
					MaxShardCommitteeSize:            4,
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey7, *incKey8, *incKey9, *incKey10,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey5,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey11, *incKey12,
								},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						swapRule: swapRule2,
					},
				},
			},
			want1:   committeeChangeValidInputBackToSub2,
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: false,
		},
		{
			name: "Valid Input [Swap Out]",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey6, *incKey2, *incKey3, *incKey4,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey5,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					unstakeRule: unstakeRule1,
					swapRule:    swapRule3,
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey2, *incKey3, *incKey4, *incKey5,
					},
				},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
				},
			},
			args: args{
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key5},
					[]string{key6},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				env: &BeaconCommitteeStateEnvironment{
					NumberOfFixedShardBlockValidator: 0,
					ConsensusStateDB:                 sDB,
					RandomNumber:                     5000,
					ActiveShards:                     1,
					MaxShardCommitteeSize:            4,
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey6, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey5,
								},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						unstakeRule: unstakeRule1,
						swapRule:    swapRule3,
					},
				},
			},
			want1: committeeChangeValidInputSwapOut,
			want2: instruction.NewReturnStakeInsWithValue(
				[]string{key6},
				[]string{hash6.String()},
			),
			wantErr: false,
		},
		{
			name: "Valid Input [Swap Out], Data in 2 Shards",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey6, *incKey2, *incKey3, *incKey4,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey7, *incKey8, *incKey9, *incKey10,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey, *incKey13,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey11, *incKey12,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					unstakeRule: unstakeRule1,
					swapRule:    swapRule4,
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey2, *incKey3, *incKey4, *incKey,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey7, *incKey8, *incKey9, *incKey10,
					},
				},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey13,
					},
					1: []incognitokey.CommitteePublicKey{
						*incKey11, *incKey12,
					},
				},
			},
			args: args{
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key},
					[]string{key6},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				env: &BeaconCommitteeStateEnvironment{
					NumberOfFixedShardBlockValidator: 0,
					ConsensusStateDB:                 sDB,
					RandomNumber:                     5000,
					ActiveShards:                     2,
					MaxShardCommitteeSize:            4,
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey6, *incKey2, *incKey3, *incKey4,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey7, *incKey8, *incKey9, *incKey10,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey13,
								},
								1: []incognitokey.CommitteePublicKey{
									*incKey11, *incKey12,
								},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						unstakeRule: unstakeRule1,
						swapRule:    swapRule4,
					},
				},
			},
			want1: committeeChangeValidInputSwapOut2,
			want2: instruction.NewReturnStakeInsWithValue(
				[]string{key6},
				[]string{hash6.String()},
			),
			wantErr: false,
		},
		{
			name: "Valid Input [Slashing Force Swap Out]",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey6, *incKey2, *incKey3, *incKey,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey5,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					unstakeRule: unstakeRule2,
					swapRule:    swapRule5,
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey6, *incKey2, *incKey3, *incKey5,
					},
				},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
				},
			},
			args: args{
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key5},
					[]string{key},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				env: &BeaconCommitteeStateEnvironment{
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key:  samplePenalty,
						key7: samplePenalty,
					},
					NumberOfFixedShardBlockValidator: 0,
					ConsensusStateDB:                 sDB,
					RandomNumber:                     5000,
					ActiveShards:                     1,
					MaxShardCommitteeSize:            4,
					MinShardCommitteeSize:            0,
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: &instruction.ReturnStakeInstruction{},
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey6, *incKey2, *incKey3, *incKey,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey5,
								},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						unstakeRule: unstakeRule2,
						swapRule:    swapRule5,
					},
				},
			},
			want1: committeeChangeSlashingForceSwapOut,
			want2: instruction.NewReturnStakeInsWithValue(
				[]string{key},
				[]string{hash.String()},
			),
			wantErr: false,
		},
		{
			name: "[Valid input] swapruleV3 - slash + normal + swap in",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5, *incKey6, *incKey7,
								*incKey8, *incKey9, *incKey10, *incKey11, *incKey12,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey13, *incKey14, *incKey15, *incKey16,
								*incKey17, *incKey18, *incKey19,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					unstakeRule: unstakeRule3,
					swapRule:    swapRuleV3,
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5,
						*incKey6, *incKey7, *incKey8, *incKey9, *incKey10, *incKey12, *incKey13, *incKey14,
					},
				},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey15, *incKey16, *incKey17, *incKey18, *incKey19,
					},
				},
			},
			args: args{
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key13, key14},
					[]string{key11},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				env: &BeaconCommitteeStateEnvironment{
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key11: samplePenalty,
						key0:  samplePenalty,
					},
					NumberOfFixedShardBlockValidator: 8,
					ConsensusStateDB:                 sDB,
					RandomNumber:                     5000,
					ActiveShards:                     1,
					MaxShardCommitteeSize:            8,
					MinShardCommitteeSize:            4,
					DcsMaxShardCommitteeSize:         51,
					DcsMinShardCommitteeSize:         15,
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: &instruction.ReturnStakeInstruction{},
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey0, *incKey, *incKey2, *incKey3, *incKey4, *incKey5, *incKey6, *incKey7,
									*incKey8, *incKey9, *incKey10, *incKey11, *incKey12,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey13, *incKey14, *incKey15, *incKey16,
									*incKey17, *incKey18, *incKey19,
								},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						unstakeRule: unstakeRule3,
						swapRule:    swapRuleV3,
					},
				},
			},
			want1: committeeChangeSwapRuleV3,
			want2: instruction.NewReturnStakeInsWithValue(
				[]string{key11},
				[]string{hash.String()},
			),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV2{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
			}
			got1, got2, err := b.processSwapShardInstruction(
				tt.args.swapShardInstruction,
				tt.args.env,
				tt.args.committeeChange,
				tt.args.returnStakingInstructions,
				tt.args.oldState,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV2.processSwapShardInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeStateV2.processSwapShardInstruction() = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("BeaconCommitteeStateV2.processSwapShardInstruction() = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(b.shardCommittee, tt.committeeAfterProcess.shardCommittee) {
				t.Errorf("BeaconCommitteeStateV2.shardCommittee After Process() = %v, want %v", b.shardCommittee, tt.committeeAfterProcess.shardCommittee)
			}
			if !reflect.DeepEqual(b.shardSubstitute, tt.committeeAfterProcess.shardSubstitute) {
				t.Errorf("BeaconCommitteeStateV2.shardSubstitute After Process() = %v, want %v", b.shardSubstitute, tt.committeeAfterProcess.shardSubstitute)
			}
		})
	}
}

func TestBeaconCommitteeStateV2_processAfterNormalSwap(t *testing.T) {

	initTestParams()
	initLog()

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	sDB2, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	sDB3, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})

	hash, err := common.Hash{}.NewHashFromStr("123")
	hash6, err := common.Hash{}.NewHashFromStr("456")
	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{*incKey, *incKey6},
		map[string]privacy.PaymentAddress{
			incKey.GetIncKeyBase58():  paymentAddress,
			incKey6.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key:  true,
			key6: false,
		},
		map[string]common.Hash{
			key:  *hash,
			key6: *hash6,
		},
	)
	statedb.StoreStakerInfo(
		sDB2,
		[]incognitokey.CommitteePublicKey{*incKey, *incKey6},
		map[string]privacy.PaymentAddress{
			incKey.GetIncKeyBase58():  paymentAddress,
			incKey6.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key:  false,
			key6: false,
		},
		map[string]common.Hash{
			key:  *hash,
			key6: *hash6,
		},
	)
	statedb.StoreStakerInfo(
		sDB3,
		[]incognitokey.CommitteePublicKey{*incKey, *incKey6},
		map[string]privacy.PaymentAddress{
			incKey.GetIncKeyBase58():  paymentAddress,
			incKey6.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key:  false,
			key6: false,
		},
		map[string]common.Hash{
			key:  *hash,
			key6: *hash6,
		},
	)

	unstakeRule := &mocks.UnstakeRule{}
	unstakeRule.On("RemoveFromState",
		*incKey,
		map[string]bool{
			key:  true,
			key5: true,
			key8: true,
		},
		map[string]privacy.PaymentAddress{
			incKey.GetIncKeyBase58(): paymentAddress,
		},
		map[string]common.Hash{
			key:  *hash,
			key5: *hash6,
			key6: *hash6,
		},
		map[string]uint64{},
		[]string(nil),
		[]string(nil)).
		Return(
			map[string]bool{
				key5: true,
				key8: true,
			},
			map[string]privacy.PaymentAddress{},
			map[string]common.Hash{
				key5: *hash6,
				key6: *hash6,
			},
			[]string{
				key,
			},
			[]string{},
			nil,
		)

	rootHash, _ := sDB.Commit(true)
	sDB.Database().TrieDB().Commit(rootHash, false)

	rootHash2, _ := sDB2.Commit(true)
	sDB2.Database().TrieDB().Commit(rootHash2, false)

	rootHash3, _ := sDB3.Commit(true)
	sDB3.Database().TrieDB().Commit(rootHash3, false)

	type fields struct {
		beaconCommitteeStateSlashingBase
	}
	type args struct {
		env                     *BeaconCommitteeStateEnvironment
		outPublicKeys           []string
		committeeChange         *CommitteeChange
		returnStakeInstructions *instruction.ReturnStakeInstruction
		oldState                *BeaconCommitteeStateV2
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want1              *CommitteeChange
		want2              *instruction.ReturnStakeInstruction
		wantErr            bool
	}{
		{
			name: "[Back To Substitute] Not Found Staker Info",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey, *incKey2, *incKey3, *incKey4,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{},
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
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey, *incKey2, *incKey3, *incKey4,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{},
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
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
					ActiveShards:     1,
					RandomNumber:     10000,
				},
				outPublicKeys:           []string{key5},
				committeeChange:         &CommitteeChange{},
				returnStakeInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
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
						},
					},
				},
			},
			want1:   &CommitteeChange{},
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: true,
		},
		{
			name: "[Swap Out] Return Staking Amount",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey2, *incKey3, *incKey4, *incKey5,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{},
						},
						autoStake: map[string]bool{
							key:  true,
							key5: true,
							key8: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key:  *hash,
							key5: *hash6,
							key6: *hash6,
						},
					},
					unstakeRule: unstakeRule,
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey2, *incKey3, *incKey4, *incKey5,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{},
						},
						autoStake: map[string]bool{
							key5: true,
							key8: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx: map[string]common.Hash{
							key5: *hash6,
							key6: *hash6,
						},
					},
					unstakeRule: unstakeRule,
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB2,
					ActiveShards:     1,
					RandomNumber:     10000,
				},
				outPublicKeys:           []string{key},
				committeeChange:         &CommitteeChange{},
				returnStakeInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey2, *incKey3, *incKey4, *incKey5,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
							},
							autoStake: map[string]bool{
								key:  true,
								key5: true,
								key8: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey.GetIncKeyBase58(): paymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key:  *hash,
								key5: *hash6,
								key6: *hash6,
							},
						},
						unstakeRule: unstakeRule,
					},
				},
			},
			want1: &CommitteeChange{
				RemovedStaker: []string{key},
				TermsRemoved:  []string{},
			},
			want2: instruction.NewReturnStakeInsWithValue(
				[]string{key},
				[]string{hash.String()},
			),
			wantErr: false,
		},
		{
			name: "[Swap Out] Not Found Staker Info",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey, *incKey2, *incKey3, *incKey4,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{},
						},
						autoStake: map[string]bool{
							key:  true,
							key8: false,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key: *hash,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey, *incKey2, *incKey3, *incKey4,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{},
						},
						autoStake: map[string]bool{
							key:  true,
							key8: false,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key: *hash,
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
					ActiveShards:     1,
					RandomNumber:     10000,
				},
				outPublicKeys:           []string{key5},
				committeeChange:         &CommitteeChange{},
				returnStakeInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey, *incKey2, *incKey3, *incKey4,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
							},
							autoStake: map[string]bool{
								key:  true,
								key8: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey.GetIncKeyBase58(): paymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key: *hash,
							},
						},
					},
				},
			},
			want1:   &CommitteeChange{},
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: true,
		},
		{
			name: "[Back To Substitute] Valid Input",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey2, *incKey3, *incKey4, *incKey5,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{},
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
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey2, *incKey3, *incKey4, *incKey5,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey,
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
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
					ActiveShards:     1,
					RandomNumber:     10000,
				},
				outPublicKeys: []string{key},
				committeeChange: &CommitteeChange{
					ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
				},
				returnStakeInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey2, *incKey3, *incKey4, *incKey5,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
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
						},
					},
				},
			},
			want1: &CommitteeChange{
				ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey,
					},
				},
			},
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: false,
		},
		{
			name: "[Back To Substitute] Valid Input 2",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey2, *incKey3, *incKey4, *incKey5,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{*incKey6, *incKey7},
							1: []incognitokey.CommitteePublicKey{*incKey8, *incKey9},
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
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey2, *incKey3, *incKey4, *incKey5,
							},
						},
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{
								*incKey6, *incKey7,
							},
							1: []incognitokey.CommitteePublicKey{
								*incKey8, *incKey9, *incKey,
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
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB,
					ActiveShards:     2,
					RandomNumber:     10000,
				},
				outPublicKeys: []string{key},
				committeeChange: &CommitteeChange{
					ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
				},
				returnStakeInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{
									*incKey2, *incKey3, *incKey4, *incKey5,
								},
							},
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
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
						},
					},
				},
			},
			want1: &CommitteeChange{
				ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
					1: []incognitokey.CommitteePublicKey{
						*incKey,
					},
				},
			},
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV2{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: tt.fields.beaconCommittee,
						shardCommittee:  tt.fields.shardCommittee,
						shardSubstitute: tt.fields.shardSubstitute,
						autoStake:       tt.fields.autoStake,
						rewardReceiver:  tt.fields.rewardReceiver,
						stakingTx:       tt.fields.stakingTx,
						mu:              tt.fields.mu,
					},
					shardCommonPool:            tt.fields.shardCommonPool,
					unstakeRule:                tt.fields.unstakeRule,
					numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				},
			}
			got1, got2, err := b.processAfterNormalSwap(
				tt.args.env,
				tt.args.outPublicKeys,
				tt.args.committeeChange,
				tt.args.returnStakeInstructions,
				tt.args.oldState,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("processAfterNormalSwap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase) {
				t.Errorf("processAfterSwap() tt.fields = %v, tt.fieldsAfterProcess %v", tt.fields, tt.fieldsAfterProcess)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("processAfterSwap() got1 = %v, want1 %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("processAfterSwap() got2 = %v, want2 %v", got2, tt.want2)
			}
		})
	}
}

func TestBeaconCommitteeStateV2_processUnstakeInstruction(t *testing.T) {

	// Init data for testcases
	initTestParams()
	initLog()

	rewardReceiverkey := incKey.GetIncKeyBase58()
	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})

	validSDB1, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreStakerInfo(
		validSDB1,
		[]incognitokey.CommitteePublicKey{*incKey},
		map[string]privacy.PaymentAddress{
			rewardReceiverkey: paymentAddress,
		},
		map[string]bool{
			key: true,
		},
		map[string]common.Hash{
			key: *hash,
		},
	)

	validSDB2, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	statedb.StoreStakerInfo(
		validSDB2,
		[]incognitokey.CommitteePublicKey{*incKey, *incKey2, *incKey5, *incKey6},
		map[string]privacy.PaymentAddress{
			rewardReceiverkey:         paymentAddress,
			incKey2.GetIncKeyBase58(): paymentAddress,
			incKey3.GetIncKeyBase58(): paymentAddress,
			incKey4.GetIncKeyBase58(): paymentAddress,
			incKey5.GetIncKeyBase58(): paymentAddress,
			incKey6.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key:  false,
			key2: false,
			key3: true,
			key4: true,
			key5: false,
			key6: false,
		},
		map[string]common.Hash{
			key:  *hash,
			key2: *hash,
			key3: *hash,
			key4: *hash,
			key5: *hash,
			key6: *hash,
		},
	)

	committeePublicKeyWrongFormat := incognitokey.CommitteePublicKey{}
	committeePublicKeyWrongFormat.MiningPubKey = nil

	unstakeRule1 := &mocks.UnstakeRule{}
	unstakeRule1.On("RemoveFromState",
		*incKey,
		map[string]bool{
			key: true,
		},
		map[string]privacy.PaymentAddress{
			rewardReceiverkey: paymentAddress,
		},
		map[string]common.Hash{
			key: *hash,
		},
		map[string]uint64{},
		[]string(nil),
		[]string(nil)).
		Return(
			map[string]bool{},
			map[string]privacy.PaymentAddress{},
			map[string]common.Hash{},
			[]string{
				key,
			},
			[]string{},
			nil,
		)

	unstakeRule2 := &mocks.UnstakeRule{}
	unstakeRule2.On("RemoveFromState",
		*incKey,
		map[string]bool{
			key:  false,
			key2: false,
			key3: true,
			key4: true,
			key5: false,
			key6: false,
		},
		map[string]privacy.PaymentAddress{
			rewardReceiverkey: paymentAddress,
		},
		map[string]common.Hash{
			key: *hash,
		},
		map[string]uint64{},
		[]string(nil),
		[]string(nil)).
		Return(
			map[string]bool{
				key2: false,
				key3: true,
				key4: true,
				key5: false,
				key6: false,
			},
			map[string]privacy.PaymentAddress{},
			map[string]common.Hash{},
			[]string{
				key,
			},
			[]string{},
			nil,
		)

	unstakeRule2.On("RemoveFromState",
		*incKey2,
		map[string]bool{
			key2: false,
			key3: true,
			key4: true,
			key5: false,
			key6: false,
		},
		map[string]privacy.PaymentAddress{},
		map[string]common.Hash{},
		map[string]uint64{},
		[]string{key},
		[]string{}).
		Return(
			map[string]bool{
				key3: true,
				key4: true,
				key5: false,
				key6: false,
			},
			map[string]privacy.PaymentAddress{},
			map[string]common.Hash{},
			[]string{
				key, key2,
			},
			[]string{},
			nil,
		)

	unstakeRule2.On("RemoveFromState",
		*incKey5,
		map[string]bool{
			key3: true,
			key4: true,
			key5: false,
			key6: false,
		},
		map[string]privacy.PaymentAddress{},
		map[string]common.Hash{},
		map[string]uint64{},
		[]string{key, key2},
		[]string{}).
		Return(
			map[string]bool{
				key3: true,
				key4: true,
				key6: false,
			},
			map[string]privacy.PaymentAddress{},
			map[string]common.Hash{},
			[]string{
				key, key2, key5,
			},
			[]string{},
			nil,
		)

	unstakeRule2.On("RemoveFromState",
		*incKey6,
		map[string]bool{
			key3: true,
			key4: true,
			key6: false,
		},
		map[string]privacy.PaymentAddress{},
		map[string]common.Hash{},
		map[string]uint64{},
		[]string{key, key2, key5},
		[]string{}).
		Return(
			map[string]bool{
				key3: true,
				key4: true,
			},
			map[string]privacy.PaymentAddress{},
			map[string]common.Hash{},
			[]string{
				key, key2, key5, key6,
			},
			[]string{},
			nil,
		)

	type fields struct {
		beaconCommitteeStateSlashingBase
	}
	type args struct {
		unstakeInstruction        *instruction.UnstakeInstruction
		returnStakingInstructions *instruction.ReturnStakeInstruction
		env                       *BeaconCommitteeStateEnvironment
		committeeChange           *CommitteeChange
		oldState                  *BeaconCommitteeStateV2
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CommitteeChange
		want1   *instruction.ReturnStakeInstruction
		wantErr bool
	}{
		{
			name: "[Subtitute List] Invalid Format Of Committee Public Key In Unstake Instruction",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase:   beaconCommitteeStateBase{},
					shardCommonPool:            []incognitokey.CommitteePublicKey{*incKey},
					numberOfAssignedCandidates: 0,
				},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys: []string{"123"},
				},
				env: &BeaconCommitteeStateEnvironment{
					newUnassignedCommonPool: []string{"123"},
				},
				committeeChange:           &CommitteeChange{},
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase:   beaconCommitteeStateBase{},
						shardCommonPool:            []incognitokey.CommitteePublicKey{*incKey},
						numberOfAssignedCandidates: 0,
					},
				},
			},
			want:    &CommitteeChange{},
			want1:   &instruction.ReturnStakeInstruction{},
			wantErr: true,
		},
		{
			name: "[Subtitute List] Can't find staker info in database",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						autoStake: map[string]bool{
							key: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							rewardReceiverkey: paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key: *hash,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{*incKey},
				},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys:       []string{key2},
					CommitteePublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey2},
				},
				env: &BeaconCommitteeStateEnvironment{
					newUnassignedCommonPool: []string{key2},
					ConsensusStateDB:        validSDB1,
				},
				committeeChange:           &CommitteeChange{},
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							autoStake: map[string]bool{
								key: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								rewardReceiverkey: paymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key: *hash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{*incKey},
					},
				},
			},
			want:    &CommitteeChange{},
			want1:   new(instruction.ReturnStakeInstruction),
			wantErr: true,
		},
		{
			name: "Valid Input Key In Candidates List",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						autoStake: map[string]bool{
							key: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							rewardReceiverkey: paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key: *hash,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{*incKey},
					unstakeRule:     unstakeRule1,
				},
			},
			args: args{
				committeeChange: &CommitteeChange{},
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys:       []string{key},
					CommitteePublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey},
				},
				env: &BeaconCommitteeStateEnvironment{
					newUnassignedCommonPool: []string{key},
					ConsensusStateDB:        validSDB1,
				},
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							autoStake: map[string]bool{
								key: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								rewardReceiverkey: paymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key: *hash,
							},
						},
						shardCommonPool: []incognitokey.CommitteePublicKey{*incKey},
						unstakeRule:     unstakeRule1,
					},
				},
			},
			want: &CommitteeChange{
				NextEpochShardCandidateRemoved: []incognitokey.CommitteePublicKey{*incKey},
				RemovedStaker:                  []string{key},
				TermsRemoved:                   []string{},
			},
			want1: instruction.NewReturnStakeInsWithValue(
				[]string{key},
				[]string{hash.String()},
			),
			wantErr: false,
		},
		{
			name: "Valid Input Key In Validators List",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{*incKey},
						},
						autoStake: map[string]bool{
							key: true,
						},
					},
				},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys:       []string{key},
					CommitteePublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey},
				},
				env: &BeaconCommitteeStateEnvironment{
					newAllSubstituteCommittees: []string{key},
				},
				committeeChange:           &CommitteeChange{},
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{*incKey},
							},
							autoStake: map[string]bool{
								key: true,
							},
						},
					},
				},
			},
			want: &CommitteeChange{
				StopAutoStake: []string{key},
			},
			want1:   new(instruction.ReturnStakeInstruction),
			wantErr: false,
		},
		{
			name: "Remove 4 keys in shard common pool",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
							0: []incognitokey.CommitteePublicKey{},
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							rewardReceiverkey: paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key: *hash,
						},
						autoStake: map[string]bool{
							key:  false,
							key2: false,
							key3: true,
							key4: true,
							key5: false,
							key6: false,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey, *incKey2, *incKey3, *incKey4, *incKey5, *incKey6,
					},
					unstakeRule: unstakeRule2,
				},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys: []string{key, key2, key5, key6},
					CommitteePublicKeysStruct: []incognitokey.CommitteePublicKey{
						*incKey, *incKey2, *incKey5, *incKey6,
					},
				},
				env: &BeaconCommitteeStateEnvironment{
					newAllSubstituteCommittees: []string{key, key2, key3, key4, key5, key6},
					ConsensusStateDB:           validSDB2,
					newUnassignedCommonPool:    []string{key, key2, key3, key4, key5, key6},
				},
				committeeChange:           &CommitteeChange{},
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
								0: []incognitokey.CommitteePublicKey{},
							},
							autoStake: map[string]bool{
								key:  false,
								key2: false,
								key3: true,
								key4: true,
								key5: false,
								key6: false,
							},
						},
						unstakeRule: unstakeRule2,
					},
				},
			},
			want: &CommitteeChange{
				NextEpochShardCandidateRemoved: []incognitokey.CommitteePublicKey{
					*incKey, *incKey2, *incKey5, *incKey6,
				},
				RemovedStaker: []string{key, key2, key5, key6},
				TermsRemoved:  []string{},
			},
			want1: &instruction.ReturnStakeInstruction{
				PublicKeys: []string{key, key2, key5, key6},
				PublicKeysStruct: []incognitokey.CommitteePublicKey{
					*incKey, *incKey2, *incKey5, *incKey6,
				},
				StakingTXIDs: []string{
					hash.String(),
					hash.String(),
					hash.String(),
					hash.String(),
				},
				StakingTxHashes: []common.Hash{
					*hash,
					*hash,
					*hash,
					*hash,
				},
				PercentReturns: []uint{100, 100, 100, 100},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV2{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: tt.fields.beaconCommittee,
						shardCommittee:  tt.fields.shardCommittee,
						shardSubstitute: tt.fields.shardSubstitute,
						autoStake:       tt.fields.autoStake,
						rewardReceiver:  tt.fields.rewardReceiver,
						stakingTx:       tt.fields.stakingTx,
						mu:              tt.fields.mu,
					},
					shardCommonPool:            tt.fields.shardCommonPool,
					numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
					unstakeRule:                tt.fields.unstakeRule,
				},
			}
			got, got1, err := b.processUnstakeInstruction(
				tt.args.unstakeInstruction,
				tt.args.env,
				tt.args.committeeChange,
				tt.args.returnStakingInstructions,
				tt.args.oldState,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("processUnstakeInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processUnstakeInstruction() got = %v, want %v", got.TermsRemoved, tt.want.TermsRemoved)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("processUnstakeInstruction() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestBeaconCommitteeStateV2_processStopAutoStakeInstruction(t *testing.T) {

	initTestParams()

	type fields struct {
		beaconCommitteeStateSlashingBase
	}
	type args struct {
		stopAutoStakeInstruction *instruction.StopAutoStakeInstruction
		env                      *BeaconCommitteeStateEnvironment
		committeeChange          *CommitteeChange
		oldState                 *BeaconCommitteeStateV2
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
	}{
		{
			name:               "Not Found In List Subtitutes",
			fields:             fields{},
			fieldsAfterProcess: fields{},
			args: args{
				stopAutoStakeInstruction: &instruction.StopAutoStakeInstruction{
					CommitteePublicKeys: []string{key},
				},
				env: &BeaconCommitteeStateEnvironment{
					newAllCandidateSubstituteCommittee: []string{key2},
				},
				committeeChange: &CommitteeChange{},
				oldState:        &BeaconCommitteeStateV2{},
			},
			want: &CommitteeChange{},
		},
		{
			name: "Found In List Subtitutes",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						autoStake: map[string]bool{
							key: true,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						autoStake: map[string]bool{
							key: false,
						},
					},
				},
			},
			args: args{
				stopAutoStakeInstruction: &instruction.StopAutoStakeInstruction{
					CommitteePublicKeys: []string{key},
				},
				env: &BeaconCommitteeStateEnvironment{
					newAllCandidateSubstituteCommittee: []string{key},
				},
				committeeChange: &CommitteeChange{},
				oldState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							autoStake: map[string]bool{
								key: true,
							},
						},
					},
				},
			},
			want: &CommitteeChange{
				StopAutoStake: []string{key},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV2{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: tt.fields.beaconCommittee,
						shardCommittee:  tt.fields.shardCommittee,
						shardSubstitute: tt.fields.shardSubstitute,
						autoStake:       tt.fields.autoStake,
						rewardReceiver:  tt.fields.rewardReceiver,
						stakingTx:       tt.fields.stakingTx,
						mu:              tt.fields.mu,
					},
					shardCommonPool:            tt.fields.shardCommonPool,
					numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				},
			}
			if got := b.processStopAutoStakeInstruction(
				tt.args.stopAutoStakeInstruction,
				tt.args.env,
				tt.args.committeeChange,
				tt.args.oldState,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processStopAutoStakeInstruction() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(tt.fields, tt.fieldsAfterProcess) {
				t.Errorf("processAfterSwap() tt.fields = %v, tt.fieldsAfterProcess %v", tt.fields, tt.fieldsAfterProcess)
			}
		})
	}
}
