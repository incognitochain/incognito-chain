package committeestate

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/wallet"

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
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee:  map[byte][]string{},
						shardSubstitute: map[byte][]string{},
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
					PublicKeys: []string{key1},
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
			},
			want: &CommitteeChange{
				NextEpochShardCandidateAdded: []incognitokey.CommitteePublicKey{
					*incKey,
				},
			},
			wantSideEffect: &fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee:  map[byte][]string{},
						shardSubstitute: map[byte][]string{},
						autoStake: map[string]bool{
							key1: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key1: *txHash,
						},
					},
					shardCommonPool: []string{key1},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV2{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
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
			_, has, _ := statedb.GetShardStakerInfo(tt.args.env.ConsensusStateDB, key1)
			if has {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), StoreStakerInfo found, %+v", key1)
			}
		})
	}
}

func TestBeaconCommitteeStateV2_processAssignWithRandomInstruction(t *testing.T) {

	initTestParams()

	committeeChangeValidInput := NewCommitteeChange()
	committeeChangeValidInput.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey2,
	}
	committeeChangeValidInput.ShardSubstituteAdded[1] = []incognitokey.CommitteePublicKey{
		*incKey2,
	}
	finalMu := new(sync.RWMutex)
	type fields struct {
		beaconCommitteeStateSlashingBase
	}
	type args struct {
		rand              int64
		numberOfValidator []int
		committeeChange   *CommitteeChange
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
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key1,
								key5,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key6,
							},
						},
					},
					shardCommonPool: []string{
						key2,
					},
					numberOfAssignedCandidates: 1,
				},
			},
			args: args{
				rand:              10000,
				numberOfValidator: []int{3, 0},
				committeeChange:   NewCommitteeChange(),
			},
			wantSideEffect: &fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key1,
								key5,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key6,
							},
							1: []string{
								key2,
							},
						},
						mu: finalMu,
					},
					shardCommonPool:            []string{},
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
					assignRule: NewAssignRuleV2(),
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
				tt.args.numberOfValidator,
				tt.args.committeeChange,
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

	swapRule1 := &mocks.SwapRule{}
	swapRule1.On("CalculateAssignOffset", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(1)
	swapRule2 := &mocks.SwapRule{}
	swapRule2.On("CalculateAssignOffset", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(10)

	type args struct {
		shardCommonPool        []string
		shardCommittee         map[byte][]string
		shardSubstitute        map[byte][]string
		numberOfFixedValidator int
		minCommitteeSize       int
		swapRule               SwapRuleProcessor
	}
	tests := []struct {
		name                           string
		args                           args
		wantNumberOfAssignedCandidates int
	}{
		{
			name: "number of assigned candidates < number of committee in shard pool",
			args: args{
				shardCommonPool: []string{
					key8, key9, key10,
				},
				shardCommittee: map[byte][]string{
					0: []string{
						key1, key0, key2, key3,
					},
					1: []string{
						key4, key5, key6, key7,
					},
				},
				shardSubstitute:        map[byte][]string{},
				numberOfFixedValidator: 1,
				minCommitteeSize:       3,
				swapRule:               &swapRuleV2{},
			},
			wantNumberOfAssignedCandidates: 2,
		},
		{
			name: "number of assigned candidates > number of committee in shard pool",
			args: args{
				shardCommonPool: []string{
					key8, key9, key10,
				},
				shardCommittee: map[byte][]string{
					0: []string{
						key1, key0, key2, key3, key11, key12,
					},
					1: []string{
						key4, key5, key6, key7, key13, key14,
					},
				},
				shardSubstitute: map[byte][]string{
					0: []string{
						key15, key16,
					},
					1: []string{
						key17, key18,
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
				shardCommonPool: []string{
					key4,
				},
				shardCommittee: map[byte][]string{
					0: []string{
						key1, key0, key2, key3,
					},
					1: []string{
						key1, key0, key2, key3,
					},
				},
				shardSubstitute: map[byte][]string{
					0: []string{},
					1: []string{},
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
				shardCommonPool: []string{},
				shardCommittee: map[byte][]string{
					0: []string{
						key1, key0, key2, key3,
					},
					1: []string{
						key1, key0, key2, key3,
					},
				},
				shardSubstitute: map[byte][]string{
					0: []string{},
					1: []string{},
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

	emptySwapShardInstruction := &mocks.SwapRule{}
	emptySwapShardInstruction.On("Process", mock.AnythingOfType("uint8"), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(&instruction.SwapShardInstruction{}, []string{}, []string{}, []string{}, []string{})

	validInputSwapShardInstruction := &mocks.SwapRule{}
	validInputSwapShardInstruction.On("Process", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			InPublicKeys: []string{
				key5,
			},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey5,
			},
			OutPublicKeys: []string{
				key1,
			},
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey,
			},
			ChainID: 0,
			Type:    instruction.SWAP_BY_END_EPOCH,
		},
		[]string{}, []string{}, []string{}, []string{})

	swapRuleV3SingleShard := &mocks.SwapRule{}
	swapRuleV3SingleShard.On("Process", mock.AnythingOfType("uint8"), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
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

	validInputSwapShardInstruction.On("Process", uint8(1), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
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
		BeaconCommitteeState *BeaconCommitteeStateV2
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
				BeaconCommitteeState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]string{
								0: []string{},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
						},
						swapRule: NewSwapRuleV2(),
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
				BeaconCommitteeState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
								1: []string{
									key6, key7, key8, key9,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key5,
								},
								1: []string{
									key10,
								},
							},
						},
						swapRule: NewSwapRuleV2(),
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					NumberOfFixedShardBlockValidator: 0,
					ActiveShards:                     1,
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
						key1,
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
			name: "[Valid input] SwapruleV2 - slashing + normal + swap in + 1 shard",
			fields: fields{
				BeaconCommitteeState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key1, key2, key3, key4, key5, key6, key7,
									key8, key9, key10, key11, key12,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key13, key14, key15, key16,
									key17, key18, key19,
								},
							},
						},
						swapRule: NewSwapRuleV2(),
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
					MaxShardCommitteeSize:            13,
				},
			},
			want: []*instruction.SwapShardInstruction{
				&instruction.SwapShardInstruction{
					InPublicKeys: []string{
						key13, key14, key15, key16,
					},
					InPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey13, *incKey14, *incKey15, *incKey16,
					},
					OutPublicKeys: []string{
						key11,
						key8, key9, key10,
					},
					OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey11,
						*incKey8, *incKey9, *incKey10,
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
				BeaconCommitteeState: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key1, key2, key3, key4, key5, key6, key7,
									key8, key9, key10, key11, key12,
								},
								1: []string{
									key0, key1, key2, key3, key4, key5, key6, key7,
									key8, key9, key10, key11, key12,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key13, key14, key15, key16,
									key17, key18, key19,
								},
								1: []string{
									key13, key14, key15, key16,
									key17, key18, key19,
								},
							},
						},
						swapRule: NewSwapRuleV2(),
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
					MaxShardCommitteeSize:            13,
				},
			},
			want: []*instruction.SwapShardInstruction{
				&instruction.SwapShardInstruction{
					InPublicKeys: []string{
						key13, key14, key15, key16,
					},
					InPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey13, *incKey14, *incKey15, *incKey16,
					},
					OutPublicKeys: []string{
						key11, key8, key9, key10,
					},
					OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey11, *incKey8, *incKey9, *incKey10,
					},
					ChainID: 0,
					Type:    instruction.SWAP_BY_END_EPOCH,
				},
				&instruction.SwapShardInstruction{
					InPublicKeys: []string{
						key13, key14, key15, key16,
					},
					InPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey13, *incKey14, *incKey15, *incKey16,
					},
					OutPublicKeys: []string{
						key11, key8, key9, key10,
					},
					OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey11, *incKey8, *incKey9, *incKey10,
					},
					ChainID: 1,
					Type:    instruction.SWAP_BY_END_EPOCH,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.fields.BeaconCommitteeState
			got, err := b.GenerateSwapShardInstructions(tt.args.env)
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

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})

	hash, _ := common.Hash{}.NewHashFromStr("123")
	hash6, _ := common.Hash{}.NewHashFromStr("456")
	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{
			*incKey,
			*incKey6,
			*incKey11,
			*incKey8,
			*incKey9,
			*incKey10,
		},
		map[string]privacy.PaymentAddress{
			incKey.GetIncKeyBase58():   paymentAddress,
			incKey6.GetIncKeyBase58():  paymentAddress,
			incKey11.GetIncKeyBase58(): paymentAddress,
			incKey8.GetIncKeyBase58():  paymentAddress,
			incKey9.GetIncKeyBase58():  paymentAddress,
			incKey10.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key1:  true,
			key6:  false,
			key11: true,
			key8:  true,
			key9:  true,
			key10: false,
		},
		map[string]common.Hash{
			key1:  *hash,
			key6:  *hash6,
			key11: *hash,
			key8:  *hash,
			key9:  *hash6,
			key10: *hash,
		},
		1,
	)

	rootHash, _ := sDB.Commit(true)
	sDB.Database().TrieDB().Commit(rootHash, false)

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
	committeeChangeSlashingForceSwapOut.RemovedStaker = []string{key1}
	committeeChangeSlashingForceSwapOut.SlashingCommittee[0] = []string{key1}

	committeeChangeSwapRuleV3 := NewCommitteeChange()
	committeeChangeSwapRuleV3.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey13, *incKey14, *incKey15, *incKey16,
	}
	committeeChangeSwapRuleV3.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey13, *incKey14, *incKey15, *incKey16,
	}
	committeeChangeSwapRuleV3.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey11, *incKey8, *incKey9, *incKey10,
	}
	committeeChangeSwapRuleV3.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey8, *incKey9,
	}
	committeeChangeSwapRuleV3.RemovedStaker = []string{key10, key11}
	committeeChangeSwapRuleV3.SlashingCommittee[0] = []string{key11}

	type fields struct {
		beaconCommitteeStateSlashingBase
	}

	type committeeAfterProcess struct {
		shardCommittee  map[byte][]string
		shardSubstitute map[byte][]string
	}

	type args struct {
		swapShardInstruction      *instruction.SwapShardInstruction
		numberOfValidator         []int
		returnStakingInstructions *instruction.ReturnStakeInstruction
		env                       *BeaconCommitteeStateEnvironment
		committeeChange           *CommitteeChange
		newState                  *BeaconCommitteeStateV2
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
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key1, key2, key3, key4,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key5,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					swapRule: NewSwapRuleV2(),
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]string{
					0: []string{
						key1, key2, key3, key4,
					},
				},
				shardSubstitute: map[byte][]string{
					0: []string{
						key5,
					},
				},
			},
			args: args{
				swapShardInstruction: &instruction.SwapShardInstruction{
					OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey6,
					},
				},
				numberOfValidator: []int{5},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					NumberOfFixedShardBlockValidator: 0,
					MaxShardCommitteeSize:            4,
					ActiveShards:                     1,
					shardCommittee: map[byte][]string{
						0: []string{
							key1, key2, key3, key4,
						},
					},
					shardSubstitute: map[byte][]string{
						0: []string{
							key5,
						},
					},
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
			},
			want1:   nil,
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: true,
		},
		{
			name: "Swap In Not Valid In List Substitutes Public Key",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key1, key2, key3, key4,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key5,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					swapRule: NewSwapRuleV2(),
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]string{
					0: []string{
						key1, key2, key3, key4,
					},
				},
				shardSubstitute: map[byte][]string{
					0: []string{
						key5,
					},
				},
			},
			args: args{
				swapShardInstruction: &instruction.SwapShardInstruction{
					InPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey6,
					},
				},
				numberOfValidator: []int{5},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					NumberOfFixedShardBlockValidator: 0,
					MaxShardCommitteeSize:            4,
					ActiveShards:                     1,
					shardCommittee: map[byte][]string{
						0: []string{
							key1, key2, key3, key4,
						},
					},
					shardSubstitute: map[byte][]string{
						0: []string{
							key5,
						},
					},
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
			},
			want1:   nil,
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: true,
		},
		{
			name: "Valid Input [Back To Common Pool And Re-assign]",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key1, key2, key3, key4,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key5,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					swapRule: NewSwapRuleV2(),
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]string{
					0: []string{
						key2, key3, key4, key5,
					},
				},
				shardSubstitute: map[byte][]string{
					0: []string{
						key1,
					},
				},
			},
			args: args{
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key5},
					[]string{key1},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				numberOfValidator: []int{5},
				env: &BeaconCommitteeStateEnvironment{
					NumberOfFixedShardBlockValidator: 0,
					MinShardCommitteeSize:            0,
					ConsensusStateDB:                 sDB,
					RandomNumber:                     5000,
					ActiveShards:                     1,
					MaxShardCommitteeSize:            4,
					shardCommittee: map[byte][]string{
						0: []string{
							key1, key2, key3, key4,
						},
					},
					shardSubstitute: map[byte][]string{
						0: []string{
							key5,
						},
					},
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
			},
			want1: NewCommitteeChange().
				AddShardSubstituteAdded(0, []string{key1}).
				AddShardSubstituteRemoved(0, []string{key5}).
				AddShardCommitteeAdded(0, []string{key5}).
				AddShardCommitteeRemoved(0, []string{key1}),
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: false,
		},
		{
			name: "Valid Input [Back To Common Pool And Re-assign], Data in 2 Shard",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key1, key2, key3, key4,
							},
							1: []string{
								key7, key8, key9, key10,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key5,
							},
							1: []string{
								key11, key12,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					swapRule: NewSwapRuleV2(),
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]string{
					0: []string{
						key2, key3, key4, key5,
					},
					1: []string{
						key7, key8, key9, key10,
					},
				},
				shardSubstitute: map[byte][]string{
					0: []string{},
					1: []string{
						key11, key12, key1,
					},
				},
			},
			args: args{
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key5},
					[]string{key1},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				numberOfValidator: []int{5, 6},
				env: &BeaconCommitteeStateEnvironment{
					NumberOfFixedShardBlockValidator: 0,
					ConsensusStateDB:                 sDB,
					RandomNumber:                     5000,
					ActiveShards:                     2,
					ShardID:                          0,
					MaxShardCommitteeSize:            4,
					shardCommittee: map[byte][]string{
						0: []string{
							key1, key2, key3, key4,
						},
						1: []string{
							key7, key8, key9, key10,
						},
					},
					shardSubstitute: map[byte][]string{
						0: []string{
							key5,
						},
						1: []string{
							key11, key12,
						},
					},
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
			},
			want1: NewCommitteeChange().
				AddShardSubstituteAdded(1, []string{key1}).
				AddShardSubstituteRemoved(0, []string{key5}).
				AddShardCommitteeAdded(0, []string{key5}).
				AddShardCommitteeRemoved(0, []string{key1}),
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: false,
		},
		{
			name: "Valid Input [Swap Out]",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key6, key2, key3, key4,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key5,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					swapRule: NewSwapRuleV2(),
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]string{
					0: []string{
						key2, key3, key4, key5,
					},
				},
				shardSubstitute: map[byte][]string{
					0: []string{},
				},
			},
			args: args{
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key5},
					[]string{key6},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				numberOfValidator: []int{5},
				env: &BeaconCommitteeStateEnvironment{
					NumberOfFixedShardBlockValidator: 0,
					ConsensusStateDB:                 sDB,
					RandomNumber:                     5000,
					ActiveShards:                     1,
					MaxShardCommitteeSize:            4,
					shardCommittee: map[byte][]string{
						0: []string{
							key6, key2, key3, key4,
						},
					},
					shardSubstitute: map[byte][]string{
						0: []string{
							key5,
						},
					},
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
			},
			want1: NewCommitteeChange().
				AddShardCommitteeAdded(0, []string{key5}).
				AddShardSubstituteRemoved(0, []string{key5}).
				AddShardCommitteeRemoved(0, []string{key6}).
				AddRemovedStaker(key6),
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
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key6, key2, key3, key4,
							},
							1: []string{
								key7, key8, key9, key10,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key1, key13,
							},
							1: []string{
								key11, key12,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					swapRule: NewSwapRuleV2(),
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]string{
					0: []string{
						key2, key3, key4, key1,
					},
					1: []string{
						key7, key8, key9, key10,
					},
				},
				shardSubstitute: map[byte][]string{
					0: []string{
						key13,
					},
					1: []string{
						key11, key12,
					},
				},
			},
			args: args{
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key1},
					[]string{key6},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				numberOfValidator: []int{6, 6},
				env: &BeaconCommitteeStateEnvironment{
					NumberOfFixedShardBlockValidator: 0,
					ConsensusStateDB:                 sDB,
					RandomNumber:                     5000,
					ActiveShards:                     2,
					MaxShardCommitteeSize:            4,
					shardCommittee: map[byte][]string{
						0: []string{
							key6, key2, key3, key4,
						},
						1: []string{
							key7, key8, key9, key10,
						},
					},
					shardSubstitute: map[byte][]string{
						0: []string{
							key1, key13,
						},
						1: []string{
							key11, key12,
						},
					},
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
			},
			want1: NewCommitteeChange().
				AddShardCommitteeAdded(0, []string{key1}).
				AddShardSubstituteRemoved(0, []string{key1}).
				AddShardCommitteeRemoved(0, []string{key6}).
				AddRemovedStaker(key6),
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
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key6, key2, key3, key1,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key5,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					swapRule: NewSwapRuleV2(),
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]string{
					0: []string{
						key6, key2, key3, key5,
					},
				},
				shardSubstitute: map[byte][]string{
					0: []string{},
				},
			},
			args: args{
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key5},
					[]string{key1},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				numberOfValidator: []int{5},
				env: &BeaconCommitteeStateEnvironment{
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key1: samplePenalty,
						key7: samplePenalty,
					},
					NumberOfFixedShardBlockValidator: 0,
					ConsensusStateDB:                 sDB,
					RandomNumber:                     5000,
					ActiveShards:                     1,
					MaxShardCommitteeSize:            4,
					MinShardCommitteeSize:            0,
					shardCommittee: map[byte][]string{
						0: []string{
							key6, key2, key3, key1,
						},
					},
					shardSubstitute: map[byte][]string{
						0: []string{
							key5,
						},
					},
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: &instruction.ReturnStakeInstruction{},
			},
			want1: committeeChangeSlashingForceSwapOut,
			want2: instruction.NewReturnStakeInsWithValue(
				[]string{key1},
				[]string{hash.String()},
			),
			wantErr: false,
		},
		{
			name: "[Valid input] swapruleV3 - slash + normal + swap in",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key1, key2, key3, key4, key5, key6, key7,
								key8, key9, key10, key11, key12,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key13, key14, key15, key16,
								key17, key18, key19,
							},
						},
						autoStake:      map[string]bool{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						stakingTx:      map[string]common.Hash{},
					},
					swapRule: NewSwapRuleV2(),
				},
			},
			committeeAfterProcess: committeeAfterProcess{
				shardCommittee: map[byte][]string{
					0: []string{
						key0, key1, key2, key3, key4, key5, key6, key7,
						key12, key13, key14, key15, key16,
					},
				},
				shardSubstitute: map[byte][]string{
					0: []string{
						key17, key18, key19, key8, key9,
					},
				},
			},
			args: args{
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key13, key14, key15, key16},
					[]string{key11, key8, key9, key10},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				numberOfValidator: []int{20},
				env: &BeaconCommitteeStateEnvironment{
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key11: samplePenalty,
						key0:  samplePenalty,
					},
					NumberOfFixedShardBlockValidator: 8,
					ConsensusStateDB:                 sDB,
					RandomNumber:                     5000,
					ActiveShards:                     1,
					MaxShardCommitteeSize:            13,
					MinShardCommitteeSize:            4,
					shardCommittee: map[byte][]string{
						0: []string{
							key0, key1, key2, key3, key4, key5, key6, key7,
							key8, key9, key10, key11, key12,
						},
					},
					shardSubstitute: map[byte][]string{
						0: []string{
							key13, key14, key15, key16,
							key17, key18, key19,
						},
					},
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: &instruction.ReturnStakeInstruction{},
			},
			want1: committeeChangeSwapRuleV3,
			want2: instruction.NewReturnStakeInsWithValue(
				[]string{key10, key11},
				[]string{hash.String(), hash.String()},
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
				tt.args.numberOfValidator,
				tt.args.env,
				tt.args.committeeChange,
				tt.args.returnStakingInstructions,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV2.processSwapShardInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeStateV2.processSwapShardInstruction() got1 = %v, want %v", got1, tt.want1)
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
			key1: true,
			key6: false,
		},
		map[string]common.Hash{
			key1: *hash,
			key6: *hash6,
		},
		1,
	)
	statedb.StoreStakerInfo(
		sDB2,
		[]incognitokey.CommitteePublicKey{*incKey, *incKey6},
		map[string]privacy.PaymentAddress{
			incKey.GetIncKeyBase58():  paymentAddress,
			incKey6.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key1: false,
			key6: false,
		},
		map[string]common.Hash{
			key1: *hash,
			key6: *hash6,
		},
		1,
	)
	statedb.StoreStakerInfo(
		sDB3,
		[]incognitokey.CommitteePublicKey{*incKey, *incKey6},
		map[string]privacy.PaymentAddress{
			incKey.GetIncKeyBase58():  paymentAddress,
			incKey6.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key1: false,
			key6: false,
		},
		map[string]common.Hash{
			key1: *hash,
			key6: *hash6,
		},
		1,
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
		numberOfValidator       []int
		committeeChange         *CommitteeChange
		returnStakeInstructions *instruction.ReturnStakeInstruction
		newState                *BeaconCommitteeStateV2
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
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key1, key2, key3, key4,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{},
						},
						autoStake: map[string]bool{
							key1: true,
							key8: false,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key1: *hash,
							key6: *hash6,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key1, key2, key3, key4,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{},
						},
						autoStake: map[string]bool{
							key1: true,
							key8: false,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key1: *hash,
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
				numberOfValidator:       []int{5},
				committeeChange:         &CommitteeChange{},
				returnStakeInstructions: new(instruction.ReturnStakeInstruction),
			},
			want1:   &CommitteeChange{},
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: true,
		},
		{
			name: "[Swap Out] Return Staking Amount",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key2, key3, key4, key5,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{},
						},
						autoStake: map[string]bool{
							key1: true,
							key5: true,
							key8: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key1: *hash,
							key5: *hash6,
							key6: *hash6,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key2, key3, key4, key5,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{},
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
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB: sDB2,
					ActiveShards:     1,
					RandomNumber:     10000,
				},
				outPublicKeys:           []string{key1},
				numberOfValidator:       []int{4},
				committeeChange:         &CommitteeChange{},
				returnStakeInstructions: new(instruction.ReturnStakeInstruction),
			},
			want1: &CommitteeChange{
				RemovedStaker: []string{key1},
			},
			want2: instruction.NewReturnStakeInsWithValue(
				[]string{key1},
				[]string{hash.String()},
			),
			wantErr: false,
		},
		{
			name: "[Swap Out] Not Found Staker Info",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key1, key2, key3, key4,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{},
						},
						autoStake: map[string]bool{
							key1: true,
							key8: false,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key1: *hash,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key1, key2, key3, key4,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{},
						},
						autoStake: map[string]bool{
							key1: true,
							key8: false,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key1: *hash,
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
				numberOfValidator:       []int{4},
				committeeChange:         &CommitteeChange{},
				returnStakeInstructions: new(instruction.ReturnStakeInstruction),
			},
			want1:   &CommitteeChange{},
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: true,
		},
		{
			name: "[Back To Substitute] Valid Input",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key2, key3, key4, key5,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{},
						},
						autoStake: map[string]bool{
							key1: true,
							key8: false,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key1: *hash,
							key6: *hash6,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key2, key3, key4, key5,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key1,
							},
						},
						autoStake: map[string]bool{
							key1: true,
							key8: false,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key1: *hash,
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
				outPublicKeys:     []string{key1},
				numberOfValidator: []int{4},
				committeeChange: &CommitteeChange{
					ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
				},
				returnStakeInstructions: new(instruction.ReturnStakeInstruction),
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
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key2, key3, key4, key5,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{key6, key7},
							1: []string{key8, key9},
						},
						autoStake: map[string]bool{
							key1: true,
							key8: false,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key1: *hash,
							key6: *hash6,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key2, key3, key4, key5,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key6, key7,
							},
							1: []string{
								key8, key9, key1,
							},
						},
						autoStake: map[string]bool{
							key1: true,
							key8: false,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key1: *hash,
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
				outPublicKeys:     []string{key1},
				numberOfValidator: []int{6, 2},
				committeeChange: &CommitteeChange{
					ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
				},
				returnStakeInstructions: new(instruction.ReturnStakeInstruction),
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
					assignRule: NewAssignRuleV2(),
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
			got1, got2, err := b.processAfterNormalSwap(
				tt.args.env,
				tt.args.outPublicKeys,
				tt.args.numberOfValidator,
				tt.args.committeeChange,
				tt.args.returnStakeInstructions,
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
			key1: true,
		},
		map[string]common.Hash{
			key1: *hash,
		},
		1,
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
			key1: false,
			key2: false,
			key3: true,
			key4: true,
			key5: false,
			key6: false,
		},
		map[string]common.Hash{
			key1: *hash,
			key2: *hash,
			key3: *hash,
			key4: *hash,
			key5: *hash,
			key6: *hash,
		},
		1,
	)

	committeePublicKeyWrongFormat := incognitokey.CommitteePublicKey{}
	committeePublicKeyWrongFormat.MiningPubKey = nil

	type fields struct {
		beaconCommitteeStateSlashingBase
	}
	type args struct {
		unstakeInstruction        *instruction.UnstakeInstruction
		returnStakingInstructions *instruction.ReturnStakeInstruction
		env                       *BeaconCommitteeStateEnvironment
		committeeChange           *CommitteeChange
		newState                  *BeaconCommitteeStateV2
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
			name: "Valid Input Key In Candidates List",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						autoStake: map[string]bool{
							key1: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							rewardReceiverkey: paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key1: *hash,
						},
					},
					shardCommonPool: []string{key1},
				},
			},
			args: args{
				committeeChange: &CommitteeChange{},
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys:       []string{key1},
					CommitteePublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey},
				},
				env: &BeaconCommitteeStateEnvironment{
					newUnassignedCommonPool: []string{key1},
					ConsensusStateDB:        validSDB1,
				},
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
			},
			want: &CommitteeChange{
				NextEpochShardCandidateRemoved: []incognitokey.CommitteePublicKey{*incKey},
				RemovedStaker:                  []string{key1},
			},
			want1: instruction.NewReturnStakeInsWithValue(
				[]string{key1},
				[]string{hash.String()},
			),
			wantErr: false,
		},
		{
			name: "Valid Input Key In Validators List",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key1},
						},
						autoStake: map[string]bool{
							key1: true,
						},
					},
				},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys:       []string{key1},
					CommitteePublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey},
				},
				env: &BeaconCommitteeStateEnvironment{
					newAllSubstituteCommittees: []string{key1},
				},
				committeeChange:           &CommitteeChange{},
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
			},
			want: &CommitteeChange{
				StopAutoStake: []string{key1},
			},
			want1:   new(instruction.ReturnStakeInstruction),
			wantErr: false,
		},
		{
			name: "Remove 4 keys in shard common pool",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{},
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							rewardReceiverkey: paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key1: *hash,
						},
						autoStake: map[string]bool{
							key1: false,
							key2: false,
							key3: true,
							key4: true,
							key5: false,
							key6: false,
						},
					},
					shardCommonPool: []string{
						key1, key2, key3, key4, key5, key6,
					},
				},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys: []string{key1, key2, key5, key6},
					CommitteePublicKeysStruct: []incognitokey.CommitteePublicKey{
						*incKey, *incKey2, *incKey5, *incKey6,
					},
				},
				env: &BeaconCommitteeStateEnvironment{
					newAllSubstituteCommittees: []string{key1, key2, key3, key4, key5, key6},
					ConsensusStateDB:           validSDB2,
					newUnassignedCommonPool:    []string{key1, key2, key3, key4, key5, key6},
				},
				committeeChange:           &CommitteeChange{},
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
			},
			want: &CommitteeChange{
				NextEpochShardCandidateRemoved: []incognitokey.CommitteePublicKey{
					*incKey, *incKey2, *incKey5, *incKey6,
				},
				RemovedStaker: []string{key1, key2, key5, key6},
			},
			want1: &instruction.ReturnStakeInstruction{
				PublicKeys: []string{key1, key2, key5, key6},
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
					assignRule: NewAssignRuleV2(),
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
			got, got1, err := b.processUnstakeInstruction(
				tt.args.unstakeInstruction,
				tt.args.env,
				tt.args.committeeChange,
				tt.args.returnStakingInstructions,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("processUnstakeInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processUnstakeInstruction() got = %v, want %v", got, tt.want)
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
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
	}{
		{
			name: "Not Found In List Subtitutes",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						autoStake: map[string]bool{
							key1: true,
						},
					},
					assignRule: NewAssignRuleV2(),
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						autoStake: map[string]bool{
							key1: true,
						},
					},
					assignRule: NewAssignRuleV2(),
				},
			},
			args: args{
				stopAutoStakeInstruction: &instruction.StopAutoStakeInstruction{
					CommitteePublicKeys: []string{key2},
				},
				env: &BeaconCommitteeStateEnvironment{
					newAllRoles: []string{key1},
				},
				committeeChange: &CommitteeChange{},
			},
			want: &CommitteeChange{},
		},
		{
			name: "Found In List Subtitutes",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						autoStake: map[string]bool{
							key1: true,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						autoStake: map[string]bool{
							key1: false,
						},
					},
				},
			},
			args: args{
				stopAutoStakeInstruction: &instruction.StopAutoStakeInstruction{
					CommitteePublicKeys: []string{key1},
				},
				env: &BeaconCommitteeStateEnvironment{
					newAllRoles: []string{key1},
				},
				committeeChange: &CommitteeChange{},
			},
			want: &CommitteeChange{
				StopAutoStake: []string{key1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV2{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					assignRule: NewAssignRuleV2(),
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
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processStopAutoStakeInstruction() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(tt.fields, tt.fieldsAfterProcess) {
				t.Errorf("processAfterSwap() tt.fields = %v, tt.fieldsAfterProcess %v", tt.fields, tt.fieldsAfterProcess)
			}
		})
	}
}

func TestBeaconCommitteeStateV2_UpdateCommitteeState(t *testing.T) {
	hash, _ := common.Hash{}.NewHashFromStr("123")
	tempHash, _ := common.Hash{}.NewHashFromStr("456")
	initTestParams()

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	paymentAddress0, err := wallet.Base58CheckDeserialize(paymentAddreessKey0)
	assert.Nil(t, err)
	rewardReceiverkey0 := incKey0.GetIncKeyBase58()
	rewardReceiverkey4 := incKey4.GetIncKeyBase58()
	rewardReceiverKey := incKey.GetIncKeyBase58()
	paymentAddress, err := wallet.Base58CheckDeserialize(paymentAddreessKey0)
	assert.Nil(t, err)

	committeeChangeProcessStakeInstruction := NewCommitteeChange()
	committeeChangeProcessStakeInstruction.NextEpochShardCandidateAdded = []incognitokey.CommitteePublicKey{
		*incKey0,
	}

	committeeChangeProcessRandomInstruction := NewCommitteeChange()
	committeeChangeProcessRandomInstruction.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey6,
	}
	committeeChangeProcessRandomInstruction.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey6,
	}

	committeeChangeProcessStopAutoStakeInstruction := NewCommitteeChange()
	committeeChangeProcessStopAutoStakeInstruction.StopAutoStake = []string{key5}

	committeeChangeProcessSwapShardInstruction2Keys := NewCommitteeChange()
	committeeChangeProcessSwapShardInstruction2Keys.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey2, *incKey3,
	}
	committeeChangeProcessSwapShardInstruction2Keys.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey2, *incKey3,
	}

	committeeChangeProcessSwapShardInstruction2Keys.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey0, *incKey,
	}
	committeeChangeProcessSwapShardInstruction2Keys.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey0, *incKey,
	}

	committeeChangeProcessSwapShardInstruction2Keys2 := NewCommitteeChange()
	committeeChangeProcessSwapShardInstruction2Keys2.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey2, *incKey3,
	}
	committeeChangeProcessSwapShardInstruction2Keys2.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey2, *incKey3,
	}

	committeeChangeProcessSwapShardInstruction2Keys2.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey12, *incKey,
	}
	committeeChangeProcessSwapShardInstruction2Keys2.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}
	committeeChangeProcessSwapShardInstruction2Keys2.RemovedStaker = []string{key12}

	committeeChangeProcessUnstakeInstruction := NewCommitteeChange()
	committeeChangeProcessUnstakeInstruction.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey0,
	}
	committeeChangeProcessUnstakeInstruction.RemovedStaker = []string{key0}

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

	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey4, *incKey11, *incKey12},
		map[string]privacy.PaymentAddress{
			rewardReceiverkey0:         paymentAddress0.KeySet.PaymentAddress,
			rewardReceiverKey:          paymentAddress.KeySet.PaymentAddress,
			rewardReceiverkey4:         paymentAddress0.KeySet.PaymentAddress,
			incKey11.GetIncKeyBase58(): paymentAddress.KeySet.PaymentAddress,
			incKey12.GetIncKeyBase58(): paymentAddress.KeySet.PaymentAddress,
		},
		map[string]bool{
			key0:  true,
			key1:  true,
			key4:  true,
			key11: true,
			key12: false,
		},
		map[string]common.Hash{
			key0:  *hash,
			key1:  *tempHash,
			key4:  *tempHash,
			key11: *hash,
			key12: *hash,
		},
		1,
	)

	finalMu := &sync.RWMutex{}

	type fields struct {
		beaconCommitteeStateV2 *BeaconCommitteeStateV2
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}
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
			name: "Process Stake Instruction",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key5,
								},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
							mu:             finalMu,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key5,
								},
							},
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
							mu: finalMu,
						},
						shardCommonPool: []string{
							key0,
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.STAKE_ACTION,
							key0,
							instruction.SHARD_INST,
							hash.String(),
							paymentAddreessKey0,
							"true",
						},
					},
					ConsensusStateDB:      sDB,
					MaxShardCommitteeSize: 4,
					ActiveShards:          1,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeProcessStakeInstruction,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Process Random Instruction with > 0 assigned candidate",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
							mu:             finalMu,
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []string{
							key6,
						},
						numberOfAssignedCandidates: 1,
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key6,
								},
							},
							mu:             finalMu,
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool:            []string{},
						numberOfAssignedCandidates: 0,
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.RANDOM_ACTION,
							"800000",
							"120000",
							"350000",
							"190000",
						},
					},
					ActiveShards:          1,
					BeaconHeight:          100,
					MaxShardCommitteeSize: 5,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeProcessRandomInstruction,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Process Random Instruction with 0 assigned candidates",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{},
							mu:              finalMu,
							autoStake:       map[string]bool{},
							rewardReceiver:  map[string]privacy.PaymentAddress{},
							stakingTx:       map[string]common.Hash{},
						},
						shardCommonPool: []string{
							key6,
						},
						numberOfAssignedCandidates: 0,
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{},
							mu:              finalMu,
							autoStake:       map[string]bool{},
							rewardReceiver:  map[string]privacy.PaymentAddress{},
							stakingTx:       map[string]common.Hash{},
						},
						shardCommonPool: []string{
							key6,
						},
						numberOfAssignedCandidates: 0,
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.RANDOM_ACTION,
							"800000",
							"120000",
							"350000",
							"190000",
						},
					},
					ActiveShards:          1,
					BeaconHeight:          100,
					MaxShardCommitteeSize: 5,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   NewCommitteeChange(),
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Process Stop Auto Stake Instruction",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key5,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key5: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								key5: paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key5: *hash,
							},
						},
						shardCommonPool: []string{},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key5,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key5: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								key5: paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key5: *hash,
							},
						},
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.STOP_AUTO_STAKE_ACTION,
							key5,
						},
					},
					MaxShardCommitteeSize: 4,
					ActiveShards:          1,
					ConsensusStateDB:      sDB,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeProcessStopAutoStakeInstruction,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Process Swap Shard Instructions",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key1, key2, key3,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key4,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								key0: paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key0,
								},
							},
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								key0: paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
							mu: finalMu,
						},
						shardCommonPool: []string{},
						swapRule:        NewSwapRuleV2(),
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key4,
							key0,
							"0",
							"1",
						},
					},
					MaxShardCommitteeSize: 4,
					ActiveShards:          1,
					ConsensusStateDB:      sDB,
					RandomNumber:          5000,
				},
			},
			want: &BeaconCommitteeStateHash{},
			want1: NewCommitteeChange().
				AddShardCommitteeAdded(0, []string{key4}).
				AddShardSubstituteAdded(0, []string{key0}).
				AddShardSubstituteRemoved(0, []string{key4}).
				AddShardCommitteeRemoved(0, []string{key0}),
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Process Unstake Instruction",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key5,
								},
							},
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
							mu: finalMu,
						},
						shardCommonPool: []string{
							key0,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key5,
								},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
							mu:             finalMu,
						},
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
					},
					newUnassignedCommonPool: []string{key0},
					ConsensusStateDB:        sDB,
					MaxShardCommitteeSize:   4,
					ActiveShards:            1,
				},
			},
			want:  &BeaconCommitteeStateHash{},
			want1: committeeChangeProcessUnstakeInstruction,
			want2: [][]string{
				[]string{
					instruction.RETURN_ACTION,
					key0,
					hash.String(),
					"100",
				},
			},
			wantErr: false,
		},
		{
			name: "Swap in 2 keys",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key2, key3, key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key0, key1,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key2, key3, key4, key5, key6, key7, key0, key1,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
							autoStake: map[string]bool{
								key0: true,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
							mu: finalMu,
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							strings.Join([]string{key0, key1}, ","),
							"",
							"0",
							"0",
						},
					},
					MaxShardCommitteeSize:            8,
					ActiveShards:                     1,
					ConsensusStateDB:                 sDB,
					NumberOfFixedShardBlockValidator: 6,
					MinShardCommitteeSize:            6,
					RandomNumber:                     5000,
				},
			},
			want: &BeaconCommitteeStateHash{},
			want1: NewCommitteeChange().
				AddShardCommitteeAdded(0, []string{key0, key1}).
				AddShardSubstituteRemoved(0, []string{key0, key1}),
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Swap out 2 keys",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key1, key4, key5, key2, key3, key6,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key1: true,
								key4: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
								incKey4.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key1: *tempHash,
								key4: *tempHash,
							},
						},
						shardCommonPool: []string{},
						swapRule:        NewSwapRuleV2(),
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key5, key2, key3, key6,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key1, key4,
								},
							},
							autoStake: map[string]bool{
								key1: true,
								key4: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
								incKey4.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key1: *tempHash,
								key4: *tempHash,
							},
							mu: finalMu,
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							"",
							strings.Join([]string{key1, key4}, ","),
							"0",
							"0",
						},
					},
					MaxShardCommitteeSize:            7,
					ActiveShards:                     1,
					ConsensusStateDB:                 sDB,
					RandomNumber:                     5000,
					NumberOfFixedShardBlockValidator: 1,
					MinShardCommitteeSize:            1,
				},
			},
			want: &BeaconCommitteeStateHash{},
			want1: NewCommitteeChange().
				AddShardCommitteeRemoved(0, []string{key1, key4}).
				AddShardSubstituteAdded(0, []string{key1, key4}),
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Swap in and out 2 keys, 2 stay",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key1, key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key2, key3,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key4, key5, key6, key7, key2, key3,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key0, key1,
								},
							},
							autoStake: map[string]bool{
								key0: true,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
							mu: finalMu,
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							strings.Join([]string{key2, key3}, ","),
							strings.Join([]string{key0, key1}, ","),
							"0",
							"0",
						},
					},
					MaxShardCommitteeSize: 6,
					ActiveShards:          1,
					ConsensusStateDB:      sDB,
					RandomNumber:          5000,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeProcessSwapShardInstruction2Keys,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Swap in and out 2 keys, 1 in, 1 stay",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key12, key1, key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key2, key3,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key12: false,
								key1:  true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey12.GetIncKeyBase58(): paymentAddress.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():   paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key12: *hash,
								key1:  *tempHash,
							},
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key4, key5, key6, key7, key2, key3,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key1,
								},
							},
							autoStake: map[string]bool{
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey.GetIncKeyBase58(): paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key1: *tempHash,
							},
							mu: finalMu,
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							strings.Join([]string{key2, key3}, ","),
							strings.Join([]string{key12, key1}, ","),
							"0",
							"0",
						},
					},
					MaxShardCommitteeSize: 6,
					ActiveShards:          1,
					ConsensusStateDB:      sDB,
					RandomNumber:          5000,
				},
			},
			want:  &BeaconCommitteeStateHash{},
			want1: committeeChangeProcessSwapShardInstruction2Keys2,
			want2: [][]string{
				[]string{
					instruction.RETURN_ACTION,
					key12,
					hash.String(),
					"100",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.fields.beaconCommitteeStateV2
			_, got1, got2, err := b.UpdateCommitteeState(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BeaconCommitteeStateV2.UpdateCommitteeState() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Fatalf("BeaconCommitteeStateV2.UpdateCommitteeState() got1 = %v, want1 = %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Fatalf("BeaconCommitteeStateV2.UpdateCommitteeState() got2 = %v, want2 = %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(b,
				tt.fieldsAfterProcess.beaconCommitteeStateV2) {
				t.Fatalf(`BeaconCommitteeStateV2.UpdateCommitteeState() b = %v, 
					tt.fieldsAfterProcess.beaconCommitteeStateV2 = %v`,
					b, tt.fieldsAfterProcess.beaconCommitteeStateV2)
			}
		})
	}
}

func TestBeaconCommitteeStateV2_UpdateCommitteeState_MultipleInstructions(t *testing.T) {
	hash, _ := common.Hash{}.NewHashFromStr("123")
	tempHash, _ := common.Hash{}.NewHashFromStr("456")
	initTestParams()

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	paymentAddress0, err := wallet.Base58CheckDeserialize(paymentAddreessKey0)
	assert.Nil(t, err)
	rewardReceiverKey := incKey.GetIncKeyBase58()
	rewardReceiverkey0 := incKey0.GetIncKeyBase58()
	rewardReceiverkey4 := incKey4.GetIncKeyBase58()
	rewardReceiverkey10 := incKey10.GetIncKeyBase58()
	rewardReceiverkey7 := incKey7.GetIncKeyBase58()
	paymentAddress, err := wallet.Base58CheckDeserialize(paymentAddreessKey0)
	assert.Nil(t, err)

	committeeChangeStakeAndAssginResult := NewCommitteeChange()
	committeeChangeStakeAndAssginResult.NextEpochShardCandidateAdded = []incognitokey.CommitteePublicKey{
		*incKey0,
	}

	committeeChangeStakeAndAssginResult.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey5,
	}

	committeeChangeStakeAndAssginResult.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}

	committeeChangeUnstakeAssign := NewCommitteeChange()
	committeeChangeUnstakeAssign.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeUnstakeAssign.StopAutoStake = []string{
		key5,
	}
	committeeChangeUnstakeAssign.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey5,
	}

	committeeChangeUnstakeAssign2 := NewCommitteeChange()
	committeeChangeUnstakeAssign2.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey6,
	}
	committeeChangeUnstakeAssign2.RemovedStaker = []string{
		key0,
	}
	committeeChangeUnstakeAssign2.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey0,
		*incKey6,
	}

	committeeChangeUnstakeAssign3 := NewCommitteeChange()
	committeeChangeUnstakeAssign3.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey6,
	}
	committeeChangeUnstakeAssign3.RemovedStaker = []string{
		key0,
	}
	committeeChangeUnstakeAssign3.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey6,
		*incKey0,
	}

	committeeChangeUnstakeAndRandomTime := NewCommitteeChange()
	committeeChangeUnstakeAndRandomTime.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey0,
	}

	committeeChangeUnstakeAndRandomTime.RemovedStaker = []string{key0}
	committeeChangeUnstakeAndRandomTime2 := NewCommitteeChange()
	committeeChangeUnstakeAndRandomTime2.StopAutoStake = []string{key0}

	committeeChangeStopAutoStakeAndRandomTime := NewCommitteeChange()
	committeeChangeStopAutoStakeAndRandomTime.StopAutoStake = []string{key0}

	committeeChangeTwoSlashing := NewCommitteeChange()
	committeeChangeTwoSlashing.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey4,
	}
	committeeChangeTwoSlashing.ShardCommitteeRemoved[1] = []incognitokey.CommitteePublicKey{
		*incKey10,
	}
	committeeChangeTwoSlashing.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeTwoSlashing.ShardCommitteeAdded[1] = []incognitokey.CommitteePublicKey{
		*incKey11,
	}
	committeeChangeTwoSlashing.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeTwoSlashing.ShardSubstituteRemoved[1] = []incognitokey.CommitteePublicKey{
		*incKey11,
	}
	committeeChangeTwoSlashing.RemovedStaker = []string{key4, key10}
	committeeChangeTwoSlashing.SlashingCommittee[0] = []string{key4}
	committeeChangeTwoSlashing.SlashingCommittee[1] = []string{key10}
	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{*incKey, *incKey0, *incKey4, *incKey10, *incKey7},
		map[string]privacy.PaymentAddress{
			rewardReceiverKey:   paymentAddress.KeySet.PaymentAddress,
			rewardReceiverkey0:  paymentAddress0.KeySet.PaymentAddress,
			rewardReceiverkey4:  paymentAddress0.KeySet.PaymentAddress,
			rewardReceiverkey10: paymentAddress0.KeySet.PaymentAddress,
			rewardReceiverkey7:  paymentAddress0.KeySet.PaymentAddress,
		},
		map[string]bool{
			key1:  true,
			key0:  true,
			key4:  true,
			key10: true,
			key7:  true,
		},
		map[string]common.Hash{
			key1:  *tempHash,
			key0:  *hash,
			key4:  *hash,
			key10: *hash,
			key7:  *hash,
		},
		1,
	)

	finalMu := &sync.RWMutex{}

	//Declare swaprule
	swapRule1 := &mocks.SwapRule{}
	swapRule1.On("Process", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey,
			},
			OutPublicKeys: []string{key1},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey0,
			},
			InPublicKeys: []string{key0},
			ChainID:      0,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key2, key3, key4, key0}, []string{}, []string{}, []string{key1})
	swapRule1.On("Version").Return(swapRuleTestVersion)
	swapRule1.On("CalculateAssignOffset", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(1)

	swapRule2 := &mocks.SwapRule{}
	swapRule2.On("Process", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey0,
			},
			OutPublicKeys: []string{key0},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey,
			},
			InPublicKeys: []string{key1},
			ChainID:      0,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key2, key3, key4, key1}, []string{}, []string{}, []string{key0})
	swapRule2.On("Version").Return(swapRuleTestVersion)
	swapRule2.On("CalculateAssignOffset", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(1)

	swapRule3 := &mocks.SwapRule{}
	swapRule3.On("Process", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey,
			},
			OutPublicKeys: []string{key1},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey5,
			},
			InPublicKeys: []string{key5},
			ChainID:      0,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key2, key3, key4, key5}, []string{}, []string{}, []string{key1})

	swapRule3.On("Process", uint8(1), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey7,
			},
			OutPublicKeys: []string{key7},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey11,
			},
			InPublicKeys: []string{key11},
			ChainID:      1,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key8, key9, key10, key11}, []string{}, []string{}, []string{key7})

	swapRule3.On("Version").Return(swapRuleTestVersion)
	swapRule3.On("CalculateAssignOffset", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(1)

	swapRule4 := &mocks.SwapRule{}
	swapRule4.On("Process", uint8(0), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey4,
			},
			OutPublicKeys: []string{key4},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey5,
			},
			InPublicKeys: []string{key5},
			ChainID:      0,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key1, key2, key3, key5}, []string{}, []string{key4}, []string{})

	swapRule4.On("Process", uint8(1), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("map[string]signaturecounter.Penalty")).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey10,
			},
			OutPublicKeys: []string{key10},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey11,
			},
			InPublicKeys: []string{key11},
			ChainID:      1,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key7, key8, key9, key11}, []string{}, []string{key10}, []string{})

	swapRule4.On("Version").Return(swapRuleTestVersion)
	swapRule4.On("CalculateAssignOffset", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(1)

	type fields struct {
		beaconCommitteeStateV2 *BeaconCommitteeStateV2
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}
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
			name: "Stake Then Assign",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
							autoStake: map[string]bool{
								key0: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
							mu: finalMu,
						},
						numberOfAssignedCandidates: 1,
						shardCommonPool: []string{
							key5,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key5,
								},
							},
							autoStake: map[string]bool{
								key0: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
							mu: finalMu,
						},
						shardCommonPool: []string{
							key0,
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.STAKE_ACTION,
							key0,
							instruction.SHARD_INST,
							hash.String(),
							paymentAddreessKey0,
							"false",
						},
						[]string{
							instruction.RANDOM_ACTION,
							"3157440766",
							"637918",
							"3157440766",
							"3157440766",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeStakeAndAssginResult,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Assign Then Stake",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
							mu: finalMu,
						},
						numberOfAssignedCandidates: 1,
						shardCommonPool: []string{
							key5,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key5,
								},
							},
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
							mu: finalMu,
						},
						shardCommonPool: []string{
							key0,
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.RANDOM_ACTION,
							"3157440766",
							"637918",
							"3157440766",
							"3157440766",
						},
						[]string{
							instruction.STAKE_ACTION,
							key0,
							instruction.SHARD_INST,
							hash.String(),
							paymentAddreessKey0,
							"true",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeStakeAndAssginResult,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Unstake And Assign 1, Fail to Unstake because Key in Current epoch Candidate, only turn off auto stake flag",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key5: true,
								key6: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								key5: paymentAddress0.KeySet.PaymentAddress,
								key6: paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key5: *hash,
								key6: *hash,
							},
						},
						shardCommonPool: []string{
							key5,
							key6,
						},
						numberOfAssignedCandidates: 1,
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key5,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key5: false,
								key6: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								key5: paymentAddress0.KeySet.PaymentAddress,
								key6: paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key5: *hash,
								key6: *hash,
							},
						},
						shardCommonPool: []string{
							key6,
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key5,
						},
						[]string{
							instruction.RANDOM_ACTION,
							"3157440766",
							"637918",
							"3157440766",
							"3157440766",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeAssign,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Assign Then Unstake 2, Fail to Unstake because Key in Current epoch Candidate, only turn off auto stake flag",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key5: true,
								key6: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								key5: paymentAddress0.KeySet.PaymentAddress,
								key6: paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key5: *hash,
								key6: *hash,
							},
						},
						shardCommonPool: []string{
							key5,
							key6,
						},
						numberOfAssignedCandidates: 1,
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key5,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key5: false,
								key6: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								key5: paymentAddress0.KeySet.PaymentAddress,
								key6: paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key5: *hash,
								key6: *hash,
							},
						},
						shardCommonPool: []string{
							key6,
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.RANDOM_ACTION,
							"3157440766",
							"637918",
							"3157440766",
							"3157440766",
						},
						[]string{
							instruction.UNSTAKE_ACTION,
							key5,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeAssign,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Unstake And Assign 3, Success to Unstake because Key in Next epoch Candidate",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
						},
						shardCommonPool: []string{
							key6,
							key0,
						},
						numberOfAssignedCandidates: 1,
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key6,
								},
							},
							mu:             finalMu,
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.RANDOM_ACTION,
							"3157440766",
							"637918",
							"3157440766",
							"3157440766",
						},
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:  &BeaconCommitteeStateHash{},
			want1: committeeChangeUnstakeAssign3,
			want2: [][]string{
				[]string{
					instruction.RETURN_ACTION,
					key0,
					hash.String(),
					"100",
				},
			},
			wantErr: false,
		},
		{
			name: "Unstake And Assign 4, Success to Unstake because Key in Next epoch Candidate",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
						},
						numberOfAssignedCandidates: 1,
						shardCommonPool: []string{
							key6,
							key0,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key6,
								},
							},
							mu:             finalMu,
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
						[]string{
							instruction.RANDOM_ACTION,
							"3157440766",
							"637918",
							"3157440766",
							"3157440766",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:  &BeaconCommitteeStateHash{},
			want1: committeeChangeUnstakeAssign2,
			want2: [][]string{
				[]string{
					instruction.RETURN_ACTION,
					key0,
					hash.String(),
					"100",
				},
			},
			wantErr: false,
		},
		{
			name: "Unstake Then Swap 1, Failed to Unstake Swap Out key1, Only turn off auto stake",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key0,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key1: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
						},
						shardCommonPool: []string{},
						swapRule:        NewSwapRuleV2(),
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key2, key3, key4, key0,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key1,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: false,
								key1: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
						},
						shardCommonPool: []string{},
						swapRule:        NewSwapRuleV2(),
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key0,
							key1,
							"0",
							"0",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want: &BeaconCommitteeStateHash{},
			want1: NewCommitteeChange().
				AddShardCommitteeRemoved(0, []string{key1}).
				AddShardCommitteeAdded(0, []string{key0}).
				AddShardSubstituteAdded(0, []string{key1}).
				AddShardSubstituteRemoved(0, []string{key0}).
				AddStopAutoStakes([]string{key0}),
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Swap Then Unstake 2, Failed to Unstake Swap Out key1, Only turn off auto stake",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key0,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
						},
						swapRule:        swapRule1,
						shardCommonPool: []string{},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key2, key3, key4, key0,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key1,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: false,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
						},
						swapRule:        swapRule1,
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key0,
							key1,
							"0",
							"0",
						},
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want: &BeaconCommitteeStateHash{},
			want1: NewCommitteeChange().
				AddShardCommitteeRemoved(0, []string{key1}).
				AddShardCommitteeAdded(0, []string{key0}).
				AddShardSubstituteAdded(0, []string{key1}).
				AddShardSubstituteRemoved(0, []string{key0}).
				AddStopAutoStakes([]string{key0}),
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Is Beacon Random Time == False And Unstake",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
						},
						shardCommonPool: []string{
							key0,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
							mu:             finalMu,
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:  &BeaconCommitteeStateHash{},
			want1: committeeChangeUnstakeAndRandomTime,
			want2: [][]string{
				[]string{
					instruction.RETURN_ACTION,
					key0,
					hash.String(),
					"100",
				},
			},
			wantErr: false,
		},
		{
			name: "Is Beacon Random Time == True And Unstake",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
						},
						shardCommonPool: []string{
							key0,
						},
						swapRule:                   swapRule1,
						numberOfAssignedCandidates: 0,
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
						},
						shardCommonPool: []string{
							key0,
						},
						swapRule:                   swapRule1,
						numberOfAssignedCandidates: 1,
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
					IsBeaconRandomTime:    true,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeAndRandomTime2,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Is Beacon Random Time And Stop Auto Stake",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
						},
						shardCommonPool: []string{
							key0,
						},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
							},
						},
						shardCommonPool: []string{
							key0,
						},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.STOP_AUTO_STAKE_ACTION,
							key0,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeStopAutoStakeAndRandomTime,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Swap Out And Unstake",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key1,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
						},
						shardCommonPool: []string{},
						swapRule:        NewSwapRuleV2(),
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key2, key3, key4, key1,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key0,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: false,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key1,
							key0,
							"0",
							"0",
						},
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want: &BeaconCommitteeStateHash{},
			want1: NewCommitteeChange().
				AddShardCommitteeRemoved(0, []string{key0}).
				AddShardCommitteeAdded(0, []string{key1}).
				AddShardSubstituteAdded(0, []string{key0}).
				AddShardSubstituteRemoved(0, []string{key1}).
				AddStopAutoStakes([]string{key0}),
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Unstake And Swap Out",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key1,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key2, key3, key4, key1,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key0,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: false,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key1,
							key0,
							"0",
							"0",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want: &BeaconCommitteeStateHash{},
			want1: NewCommitteeChange().
				AddShardCommitteeRemoved(0, []string{key0}).
				AddShardCommitteeAdded(0, []string{key1}).
				AddShardSubstituteAdded(0, []string{key0}).
				AddShardSubstituteRemoved(0, []string{key1}).
				AddStopAutoStakes([]string{key0}),
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Stop Auto Stake And Swap Out",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key1,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
						},
						shardCommonPool: []string{},
						swapRule:        NewSwapRuleV2(),
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key2, key3, key4, key1,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key0,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: false,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key1,
							key0,
							"0",
							"0",
						},
						[]string{
							instruction.UNSTAKE_ACTION,
							key0,
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want: &BeaconCommitteeStateHash{},
			want1: NewCommitteeChange().
				AddShardCommitteeRemoved(0, []string{key0}).
				AddShardCommitteeAdded(0, []string{key1}).
				AddShardSubstituteAdded(0, []string{key0}).
				AddShardSubstituteRemoved(0, []string{key1}).
				AddStopAutoStakes([]string{key0}),
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Swap Out And Stop Auto Stake",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key2, key3, key4,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key1,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key2, key3, key4, key1,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key0,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: false,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key1: *tempHash,
							},
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.STOP_AUTO_STAKE_ACTION,
							key0,
						},
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key1,
							key0,
							"0",
							"0",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want: &BeaconCommitteeStateHash{},
			want1: NewCommitteeChange().
				AddShardCommitteeRemoved(0, []string{key0}).
				AddShardCommitteeAdded(0, []string{key1}).
				AddShardSubstituteAdded(0, []string{key0}).
				AddShardSubstituteRemoved(0, []string{key1}).
				AddStopAutoStakes([]string{key0}),
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Double Swap Instruction for 2 shard, Back to substitutes",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
								1: []string{
									key7, key8, key9, key10,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key5, key13,
								},
								1: []string{
									key11, key12,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key7: true,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey7.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key7: *hash,
								key1: *tempHash,
							},
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key2, key3, key4, key5, key13,
								},
								1: []string{
									key8, key9, key10, key11, key12,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key7,
								},
								1: []string{
									key1,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key7: true,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey7.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key7: *hash,
								key1: *tempHash,
							},
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						instruction.NewSwapShardInstructionWithValue(
							[]string{key5, key13},
							[]string{key1},
							0,
							0,
						).ToString(),
						instruction.NewSwapShardInstructionWithValue(
							[]string{key11, key12},
							[]string{key7},
							1,
							0,
						).ToString(),
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          2,
					MaxShardCommitteeSize: 8,
				},
			},
			want: &BeaconCommitteeStateHash{},
			want1: NewCommitteeChange().
				AddShardCommitteeRemoved(0, []string{key1}).
				AddShardCommitteeRemoved(1, []string{key7}).
				AddShardCommitteeAdded(0, []string{key5, key13}).
				AddShardCommitteeAdded(1, []string{key11, key12}).
				AddShardSubstituteAdded(0, []string{key7}).
				AddShardSubstituteAdded(1, []string{key1}).
				AddShardSubstituteRemoved(0, []string{key5, key13}).
				AddShardSubstituteRemoved(1, []string{key11, key12}),
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Double Swap Instruction for 2 shard, Slashing",
			fields: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key4,
								},
								1: []string{
									key7, key8, key9, key10,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key5, key13,
								},
								1: []string{
									key11, key12,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0:  true,
								key4:  true,
								key10: true,
								key7:  true,
								key1:  true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
								incKey7.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
								incKey4.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
								incKey10.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():   paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0:  *hash,
								key7:  *hash,
								key4:  *hash,
								key10: *hash,
								key1:  *tempHash,
							},
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						assignRule: NewAssignRuleV2(),
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{
									key1, key2, key3, key5,
								},
								1: []string{
									key7, key8, key9, key11,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key13,
								},
								1: []string{
									key12,
								},
							},
							mu: finalMu,
							autoStake: map[string]bool{
								key0: true,
								key7: true,
								key1: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey7.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
								incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key0: *hash,
								key7: *hash,
								key1: *tempHash,
							},
						},
						swapRule:        NewSwapRuleV2(),
						shardCommonPool: []string{},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key5,
							key4,
							"0",
							"0",
						},
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key11,
							key10,
							"1",
							"0",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          2,
					MaxShardCommitteeSize: 4,
					MissingSignaturePenalty: map[string]signaturecounter.Penalty{
						key4:  samplePenalty,
						key10: samplePenalty,
					},
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeTwoSlashing,
			want2:   [][]string{instruction.NewReturnStakeInsWithValue([]string{key4, key10}, []string{hash.String(), hash.String()}).ToString()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.fields.beaconCommitteeStateV2
			_, got1, got2, err := b.UpdateCommitteeState(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BeaconCommitteeStateV2.UpdateCommitteeState() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Fatalf("BeaconCommitteeStateV2.UpdateCommitteeState() got1 = %v, want1 = %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Fatalf("BeaconCommitteeStateV2.UpdateCommitteeState() got2 = %v, want2 = %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(b,
				tt.fieldsAfterProcess.beaconCommitteeStateV2) {
				t.Fatalf(`BeaconCommitteeStateV2.UpdateCommitteeState() b = %v, 
					tt.fieldsAfterProcess.beaconCommitteeStateV2 = %v`,
					b, tt.fieldsAfterProcess.beaconCommitteeStateV2)
			}
		})
	}
}
