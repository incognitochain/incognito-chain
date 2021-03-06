package committeestate

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/stretchr/testify/assert"
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
		beaconCommittee            []incognitokey.CommitteePublicKey
		shardCommittee             map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute            map[byte][]incognitokey.CommitteePublicKey
		shardCommonPool            []incognitokey.CommitteePublicKey
		numberOfAssignedCandidates int
		autoStake                  map[string]bool
		rewardReceiver             map[string]privacy.PaymentAddress
		stakingTx                  map[string]common.Hash
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
				shardCommittee:  map[byte][]incognitokey.CommitteePublicKey{},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{},
				autoStake:       map[string]bool{},
				rewardReceiver:  map[string]privacy.PaymentAddress{},
				stakingTx:       map[string]common.Hash{},
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
					beaconCommittee:            []incognitokey.CommitteePublicKey{},
					shardCommittee:             map[byte][]incognitokey.CommitteePublicKey{},
					shardSubstitute:            map[byte][]incognitokey.CommitteePublicKey{},
					shardCommonPool:            []incognitokey.CommitteePublicKey{},
					numberOfAssignedCandidates: 0,
					autoStake:                  map[string]bool{},
					rewardReceiver:             map[string]privacy.PaymentAddress{},
					stakingTx:                  map[string]common.Hash{},
					mu:                         &sync.RWMutex{},
				},
			},
			want: &CommitteeChange{
				NextEpochShardCandidateAdded: []incognitokey.CommitteePublicKey{
					*incKey,
				},
			},
			wantSideEffect: &fields{
				shardCommittee:  map[byte][]incognitokey.CommitteePublicKey{},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{},
				shardCommonPool: []incognitokey.CommitteePublicKey{*incKey},
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV2{
				beaconCommittee:            tt.fields.beaconCommittee,
				shardCommittee:             tt.fields.shardCommittee,
				shardSubstitute:            tt.fields.shardSubstitute,
				shardCommonPool:            tt.fields.shardCommonPool,
				numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				autoStake:                  tt.fields.autoStake,
				rewardReceiver:             tt.fields.rewardReceiver,
				stakingTx:                  tt.fields.stakingTx,
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
			if !reflect.DeepEqual(b.shardCommonPool, tt.wantSideEffect.shardCommonPool) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), shardCommonPool = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.shardCommittee, tt.wantSideEffect.shardCommittee) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), shardCommittee = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.shardSubstitute, tt.wantSideEffect.shardSubstitute) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), shardSubstitute = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.rewardReceiver, tt.wantSideEffect.rewardReceiver) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), rewardReceiver = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.autoStake, tt.wantSideEffect.autoStake) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), autoStake = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.stakingTx, tt.wantSideEffect.stakingTx) {
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
		beaconCommittee            []incognitokey.CommitteePublicKey
		shardCommittee             map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute            map[byte][]incognitokey.CommitteePublicKey
		shardCommonPool            []incognitokey.CommitteePublicKey
		numberOfAssignedCandidates int
		autoStake                  map[string]bool
		rewardReceiver             map[string]privacy.PaymentAddress
		stakingTx                  map[string]common.Hash
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
				shardCommonPool: []incognitokey.CommitteePublicKey{
					*incKey2,
				},
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
				numberOfAssignedCandidates: 1,
			},
			args: args{
				rand:            10000,
				activeShards:    2,
				committeeChange: NewCommitteeChange(),
				oldState: &BeaconCommitteeStateV2{
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey2,
					},
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
					numberOfAssignedCandidates: 1,
				},
			},
			wantSideEffect: &fields{
				shardCommonPool: []incognitokey.CommitteePublicKey{},
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
				numberOfAssignedCandidates: 0,
			},
			want: committeeChangeValidInput,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV2{
				beaconCommittee:            tt.fields.beaconCommittee,
				shardCommittee:             tt.fields.shardCommittee,
				shardSubstitute:            tt.fields.shardSubstitute,
				shardCommonPool:            tt.fields.shardCommonPool,
				numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				autoStake:                  tt.fields.autoStake,
				rewardReceiver:             tt.fields.rewardReceiver,
				stakingTx:                  tt.fields.stakingTx,
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
				wantV := tt.wantSideEffect.shardSubstitute[shardID]
				if !reflect.DeepEqual(gotV, wantV) {
					t.Errorf("BeaconCommitteeStateV2.processAssignWithRandomInstruction(), shardSubstitute = %v, want %v", gotV, wantV)
				}
			}
			for shardID, gotV := range b.shardCommittee {
				wantV := tt.wantSideEffect.shardCommittee[shardID]
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

	type args struct {
		shardCommonPool        []incognitokey.CommitteePublicKey
		shardCommittee         map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute        map[byte][]incognitokey.CommitteePublicKey
		numberOfFixedValidator int
		minCommitteeSize       int
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
			},
			wantNumberOfAssignedCandidates: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotNumberOfAssignedCandidates := SnapshotShardCommonPoolV2(tt.args.shardCommonPool, tt.args.shardCommittee, tt.args.shardSubstitute, tt.args.numberOfFixedValidator, tt.args.minCommitteeSize); gotNumberOfAssignedCandidates != tt.wantNumberOfAssignedCandidates {
				t.Errorf("SnapshotShardCommonPoolV2() = %v, want %v", gotNumberOfAssignedCandidates, tt.wantNumberOfAssignedCandidates)
			}
		})
	}
}

func TestBeaconCommitteeEngineV2_GenerateAllSwapShardInstructions(t *testing.T) {

	initTestParams()
	initLog()

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
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &BeaconCommitteeEngineV2{
				beaconHeight:                      tt.fields.beaconHeight,
				beaconHash:                        tt.fields.beaconHash,
				finalBeaconCommitteeStateV2:       tt.fields.finalBeaconCommitteeStateV2,
				uncommittedBeaconCommitteeStateV2: tt.fields.uncommittedBeaconCommitteeStateV2,
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
	committeeChangeValidInputSwapOut.SlashingCommittee[0] = nil

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
	committeeChangeValidInputSwapOut2.SlashingCommittee[0] = nil

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
	committeeChangeValidInputBackToSub.SlashingCommittee[0] = nil

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
	committeeChangeValidInputBackToSub2.SlashingCommittee[0] = nil

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
	type fields struct {
		beaconCommittee            []incognitokey.CommitteePublicKey
		shardCommittee             map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute            map[byte][]incognitokey.CommitteePublicKey
		shardCommonPool            []incognitokey.CommitteePublicKey
		numberOfAssignedCandidates int
		autoStake                  map[string]bool
		rewardReceiver             map[string]privacy.PaymentAddress
		stakingTx                  map[string]common.Hash
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
			},
			want1:   nil,
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: true,
		},
		{
			name: "Swap In Not Valid In List Substitutes Public Key",
			fields: fields{
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
			},
			want1:   nil,
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: true,
		},
		{
			name: "Swap Out But Not found In Committee List",
			fields: fields{
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
				swapShardInstruction: instruction.NewSwapShardInstructionWithValue(
					[]string{key5},
					[]string{key7},
					0,
					instruction.SWAP_BY_END_EPOCH,
				),
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					NumberOfFixedShardBlockValidator: 0,
					MaxShardCommitteeSize:            4,
				},
				committeeChange:           NewCommitteeChange(),
				returnStakingInstructions: new(instruction.ReturnStakeInstruction),
				oldState: &BeaconCommitteeStateV2{
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
			},
			want1:   nil,
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: true,
		},
		{
			name: "Valid Input [Back To Common Pool And Re-assign]",
			fields: fields{
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
			},
			want1:   committeeChangeValidInputBackToSub,
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: false,
		},
		{
			name: "Valid Input [Back To Common Pool And Re-assign], Data in 2 Shard",
			fields: fields{
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
			},
			want1:   committeeChangeValidInputBackToSub2,
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: false,
		},
		{
			name: "Valid Input [Swap Out]",
			fields: fields{
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
			},
			want1: committeeChangeSlashingForceSwapOut,
			want2: instruction.NewReturnStakeInsWithValue(
				[]string{key},
				[]string{hash.String()},
			),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV2{
				beaconCommittee:            tt.fields.beaconCommittee,
				shardCommittee:             tt.fields.shardCommittee,
				shardSubstitute:            tt.fields.shardSubstitute,
				shardCommonPool:            tt.fields.shardCommonPool,
				numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				autoStake:                  tt.fields.autoStake,
				rewardReceiver:             tt.fields.rewardReceiver,
				stakingTx:                  tt.fields.stakingTx,
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

func TestBeaconCommitteeEngineV2_UpdateCommitteeState(t *testing.T) {
	hash, _ := common.Hash{}.NewHashFromStr("123")
	tempHash, _ := common.Hash{}.NewHashFromStr("456")
	initTestParams()
	initLog()

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

	committeeChangeProcessSwapShardInstruction := NewCommitteeChange()
	committeeChangeProcessSwapShardInstruction.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey4,
	}
	committeeChangeProcessSwapShardInstruction.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey0,
	}
	committeeChangeProcessSwapShardInstruction.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey4,
	}
	committeeChangeProcessSwapShardInstruction.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey0,
	}
	committeeChangeProcessSwapShardInstruction.SlashingCommittee[0] = nil

	committeeChangeProcessSwapShardInstruction2KeysIn := NewCommitteeChange()
	committeeChangeProcessSwapShardInstruction2KeysIn.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey0, *incKey,
	}
	committeeChangeProcessSwapShardInstruction2KeysIn.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey0, *incKey,
	}
	committeeChangeProcessSwapShardInstruction2KeysIn.SlashingCommittee[0] = nil
	committeeChangeProcessSwapShardInstruction2KeysIn.ShardCommitteeRemoved[0] = nil

	committeeChangeProcessSwapShardInstruction2KeysOut := NewCommitteeChange()
	committeeChangeProcessSwapShardInstruction2KeysOut.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey, *incKey4,
	}
	committeeChangeProcessSwapShardInstruction2KeysOut.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey, *incKey4,
	}
	committeeChangeProcessSwapShardInstruction2KeysOut.SlashingCommittee[0] = nil
	committeeChangeProcessSwapShardInstruction2KeysOut.ShardSubstituteRemoved[0] = nil
	committeeChangeProcessSwapShardInstruction2KeysOut.ShardCommitteeAdded[0] = nil

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
	committeeChangeProcessSwapShardInstruction2Keys.SlashingCommittee[0] = nil

	committeeChangeProcessUnstakeInstruction := NewCommitteeChange()
	committeeChangeProcessUnstakeInstruction.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{*incKey0}
	committeeChangeProcessUnstakeInstruction.RemovedStaker = []string{key0}

	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{*incKey0, *incKey, *incKey4},
		map[string]privacy.PaymentAddress{
			rewardReceiverkey0: paymentAddress0.KeySet.PaymentAddress,
			rewardReceiverKey:  paymentAddress.KeySet.PaymentAddress,
			rewardReceiverkey4: paymentAddress0.KeySet.PaymentAddress,
		},
		map[string]bool{
			key0: true,
			key:  true,
			key4: true,
		},
		map[string]common.Hash{
			key0: *hash,
			key:  *tempHash,
			key4: *tempHash,
		},
	)

	finalMu := &sync.RWMutex{}
	unCommitteedMu := &sync.RWMutex{}

	type fields struct {
		beaconHeight                      uint64
		beaconHash                        common.Hash
		finalBeaconCommitteeStateV2       *BeaconCommitteeStateV2
		uncommittedBeaconCommitteeStateV2 *BeaconCommitteeStateV2
		version                           uint
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
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
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
					mu:             finalMu,
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				beaconHash:                  *hash,
				version:                     SLASHING_VERSION,
				beaconHeight:                10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
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
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey0,
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
					mu: unCommitteedMu,
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
			name: "Process Random Instruction",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey6,
					},
					mu:                         finalMu,
					autoStake:                  map[string]bool{},
					rewardReceiver:             map[string]privacy.PaymentAddress{},
					stakingTx:                  map[string]common.Hash{},
					numberOfAssignedCandidates: 1,
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					numberOfAssignedCandidates: 0,
					beaconCommittee:            []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey6,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              unCommitteedMu,
					autoStake:       map[string]bool{},
					rewardReceiver:  map[string]privacy.PaymentAddress{},
					stakingTx:       map[string]common.Hash{},
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
			name: "Process Stop Auto Stake Instruction",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
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
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              finalMu,
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
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
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
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              unCommitteedMu,
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
				beaconHeight: 5,
				beaconHash:   *hash,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey2, *incKey3,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey4,
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
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
				version: SLASHING_VERSION,
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0,
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
					mu: unCommitteedMu,
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
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeProcessSwapShardInstruction,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Process Unstake Instruction",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
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
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey0,
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
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
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
					shardCommonPool: make([]incognitokey.CommitteePublicKey, 0, 1),
					autoStake:       map[string]bool{},
					rewardReceiver:  map[string]privacy.PaymentAddress{},
					stakingTx:       map[string]common.Hash{},
					mu:              unCommitteedMu,
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
				beaconHeight: 5,
				beaconHash:   *hash,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey2, *incKey3, *incKey4, *incKey5, *incKey6, *incKey7,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey,
						},
					},
					mu: finalMu,
					autoStake: map[string]bool{
						key0: true,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
				version: SLASHING_VERSION,
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey2, *incKey3, *incKey4, *incKey5, *incKey6, *incKey7, *incKey0, *incKey,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
					autoStake: map[string]bool{
						key0: true,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
					},
					mu: unCommitteedMu,
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							strings.Join([]string{key0, key}, ","),
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
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeProcessSwapShardInstruction2KeysIn,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Swap out 2 keys",
			fields: fields{
				beaconHeight: 5,
				beaconHash:   *hash,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey4, *incKey5, *incKey2, *incKey3, *incKey6,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
					mu: finalMu,
					autoStake: map[string]bool{
						key:  true,
						key4: true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
						incKey4.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key:  *tempHash,
						key4: *tempHash,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
				version: SLASHING_VERSION,
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey5, *incKey2, *incKey3, *incKey6,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey4,
						},
					},
					autoStake: map[string]bool{
						key:  true,
						key4: true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
						incKey4.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key:  *tempHash,
						key4: *tempHash,
					},
					mu: unCommitteedMu,
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							"",
							strings.Join([]string{key, key4}, ","),
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
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeProcessSwapShardInstruction2KeysOut,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Swap in and out 2 keys",
			fields: fields{
				beaconHeight: 5,
				beaconHash:   *hash,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey, *incKey4, *incKey5, *incKey6, *incKey7,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey2, *incKey3,
						},
					},
					mu: finalMu,
					autoStake: map[string]bool{
						key0: true,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
				version: SLASHING_VERSION,
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey4, *incKey5, *incKey6, *incKey7, *incKey2, *incKey3,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey,
						},
					},
					autoStake: map[string]bool{
						key0: true,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
					},
					mu: unCommitteedMu,
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							strings.Join([]string{key2, key3}, ","),
							strings.Join([]string{key0, key}, ","),
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &BeaconCommitteeEngineV2{
				beaconHeight:                      tt.fields.beaconHeight,
				beaconHash:                        tt.fields.beaconHash,
				finalBeaconCommitteeStateV2:       tt.fields.finalBeaconCommitteeStateV2,
				uncommittedBeaconCommitteeStateV2: tt.fields.uncommittedBeaconCommitteeStateV2,
			}
			_, got1, got2, err := engine.UpdateCommitteeState(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BeaconCommitteeEngineV2.UpdateCommitteeState() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Fatalf("BeaconCommitteeEngineV2.UpdateCommitteeState() got1 = %v, want1 = %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Fatalf("BeaconCommitteeEngineV2.UpdateCommitteeState() got2 = %v, want2 = %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(tt.fields.uncommittedBeaconCommitteeStateV2,
				tt.fieldsAfterProcess.uncommittedBeaconCommitteeStateV2) {
				t.Fatalf(`BeaconCommitteeEngineV2.UpdateCommitteeState() tt.fields.uncommittedBeaconCommitteeStateV2 = %v, 
					tt.fieldsAfterProcess.uncommittedBeaconCommitteeStateV2 = %v`,
					tt.fields.uncommittedBeaconCommitteeStateV2, tt.fieldsAfterProcess.uncommittedBeaconCommitteeStateV2)
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

	rootHash, _ := sDB.Commit(true)
	sDB.Database().TrieDB().Commit(rootHash, false)

	rootHash2, _ := sDB2.Commit(true)
	sDB2.Database().TrieDB().Commit(rootHash2, false)

	rootHash3, _ := sDB3.Commit(true)
	sDB3.Database().TrieDB().Commit(rootHash3, false)

	type fields struct {
		beaconCommittee            []incognitokey.CommitteePublicKey
		shardCommittee             map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute            map[byte][]incognitokey.CommitteePublicKey
		shardCommonPool            []incognitokey.CommitteePublicKey
		numberOfAssignedCandidates int
		autoStake                  map[string]bool
		rewardReceiver             map[string]privacy.PaymentAddress
		stakingTx                  map[string]common.Hash
		mu                         *sync.RWMutex
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
			fieldsAfterProcess: fields{
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
			want1:   &CommitteeChange{},
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: true,
		},
		{
			name: "[Swap Out] Return Staking Amount",
			fields: fields{
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
			fieldsAfterProcess: fields{
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
			},
			want1: &CommitteeChange{RemovedStaker: []string{key}},
			want2: instruction.NewReturnStakeInsWithValue(
				[]string{key},
				[]string{hash.String()},
			),
			wantErr: false,
		},
		{
			name: "[Swap Out] Not Found Staker Info",
			fields: fields{
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
			fieldsAfterProcess: fields{
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
			want1:   &CommitteeChange{},
			want2:   new(instruction.ReturnStakeInstruction),
			wantErr: true,
		},
		{
			name: "[Back To Substitute] Valid Input",
			fields: fields{
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
			fieldsAfterProcess: fields{
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
			fieldsAfterProcess: fields{
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
				beaconCommittee:            tt.fields.beaconCommittee,
				shardCommittee:             tt.fields.shardCommittee,
				shardSubstitute:            tt.fields.shardSubstitute,
				shardCommonPool:            tt.fields.shardCommonPool,
				numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				autoStake:                  tt.fields.autoStake,
				rewardReceiver:             tt.fields.rewardReceiver,
				stakingTx:                  tt.fields.stakingTx,
				mu:                         tt.fields.mu,
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
			if !reflect.DeepEqual(tt.fields, tt.fieldsAfterProcess) {
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

	type fields struct {
		beaconCommittee            []incognitokey.CommitteePublicKey
		shardCommittee             map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute            map[byte][]incognitokey.CommitteePublicKey
		shardCommonPool            []incognitokey.CommitteePublicKey
		numberOfAssignedCandidates int
		autoStake                  map[string]bool
		rewardReceiver             map[string]privacy.PaymentAddress
		stakingTx                  map[string]common.Hash
		mu                         *sync.RWMutex
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
				shardCommonPool:            []incognitokey.CommitteePublicKey{*incKey},
				numberOfAssignedCandidates: 0,
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
					shardCommonPool:            []incognitokey.CommitteePublicKey{*incKey},
					numberOfAssignedCandidates: 0,
				},
			},
			want:    &CommitteeChange{},
			want1:   &instruction.ReturnStakeInstruction{},
			wantErr: true,
		},
		{
			name: "[Subtitute List] Can't find staker info in database",
			fields: fields{
				shardCommonPool: []incognitokey.CommitteePublicKey{*incKey},
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
					shardCommonPool: []incognitokey.CommitteePublicKey{*incKey},
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
			},
			want:    &CommitteeChange{},
			want1:   new(instruction.ReturnStakeInstruction),
			wantErr: true,
		},
		{
			name: "Valid Input Key In Candidates List",
			fields: fields{
				shardCommonPool: []incognitokey.CommitteePublicKey{*incKey},
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
					shardCommonPool: []incognitokey.CommitteePublicKey{*incKey},
					autoStake: map[string]bool{
						key: true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						rewardReceiverkey: paymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key: *hash,
					}},
			},
			want: &CommitteeChange{
				NextEpochShardCandidateRemoved: []incognitokey.CommitteePublicKey{*incKey},
				RemovedStaker:                  []string{key},
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
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{*incKey},
				},
				autoStake: map[string]bool{
					key: true,
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
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{*incKey},
					},
					autoStake: map[string]bool{
						key: true,
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
				shardCommonPool: []incognitokey.CommitteePublicKey{
					*incKey, *incKey2, *incKey3, *incKey4, *incKey5, *incKey6,
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
			},
			want: &CommitteeChange{
				NextEpochShardCandidateRemoved: []incognitokey.CommitteePublicKey{
					*incKey, *incKey2, *incKey5, *incKey6,
				},
				RemovedStaker: []string{key, key2, key5, key6},
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
				beaconCommittee:            tt.fields.beaconCommittee,
				shardCommittee:             tt.fields.shardCommittee,
				shardSubstitute:            tt.fields.shardSubstitute,
				shardCommonPool:            tt.fields.shardCommonPool,
				numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				autoStake:                  tt.fields.autoStake,
				rewardReceiver:             tt.fields.rewardReceiver,
				stakingTx:                  tt.fields.stakingTx,
				mu:                         tt.fields.mu,
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
		beaconCommittee            []incognitokey.CommitteePublicKey
		shardCommittee             map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute            map[byte][]incognitokey.CommitteePublicKey
		shardCommonPool            []incognitokey.CommitteePublicKey
		numberOfAssignedCandidates int
		autoStake                  map[string]bool
		rewardReceiver             map[string]privacy.PaymentAddress
		stakingTx                  map[string]common.Hash
		mu                         *sync.RWMutex
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
				autoStake: map[string]bool{
					key: true,
				},
			},
			fieldsAfterProcess: fields{
				autoStake: map[string]bool{
					key: false,
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
					autoStake: map[string]bool{
						key: true,
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
				beaconCommittee:            tt.fields.beaconCommittee,
				shardCommittee:             tt.fields.shardCommittee,
				shardSubstitute:            tt.fields.shardSubstitute,
				shardCommonPool:            tt.fields.shardCommonPool,
				numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				autoStake:                  tt.fields.autoStake,
				rewardReceiver:             tt.fields.rewardReceiver,
				stakingTx:                  tt.fields.stakingTx,
				mu:                         tt.fields.mu,
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

func TestBeaconCommitteeStateV2_clone(t *testing.T) {

	initTestParams()
	initLog()

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	hash, _ := common.Hash{}.NewHashFromStr("123")
	hash6, _ := common.Hash{}.NewHashFromStr("456")

	type fields struct {
		beaconCommittee            []incognitokey.CommitteePublicKey
		shardCommittee             map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute            map[byte][]incognitokey.CommitteePublicKey
		shardCommonPool            []incognitokey.CommitteePublicKey
		numberOfAssignedCandidates int
		autoStake                  map[string]bool
		rewardReceiver             map[string]privacy.PaymentAddress
		stakingTx                  map[string]common.Hash
		hashes                     *BeaconCommitteeStateHash
	}
	type args struct {
		newB *BeaconCommitteeStateV2
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "[valid input] full data",
			fields: fields{
				beaconCommittee: []incognitokey.CommitteePublicKey{
					*incKey6, *incKey7, *incKey8,
				},
				shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey3, *incKey4, *incKey5,
					},
				},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey,
					},
				},
				shardCommonPool: []incognitokey.CommitteePublicKey{
					*incKey2,
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
				numberOfAssignedCandidates: 1,
				hashes:                     NewBeaconCommitteeStateHash(),
			},
			args: args{
				newB: NewBeaconCommitteeStateV2(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV2{
				beaconCommittee:            tt.fields.beaconCommittee,
				shardCommittee:             tt.fields.shardCommittee,
				shardSubstitute:            tt.fields.shardSubstitute,
				shardCommonPool:            tt.fields.shardCommonPool,
				numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				autoStake:                  tt.fields.autoStake,
				rewardReceiver:             tt.fields.rewardReceiver,
				stakingTx:                  tt.fields.stakingTx,
				hashes:                     tt.fields.hashes,
			}
			tt.args.newB.mu = nil
			if b.clone(tt.args.newB); !reflect.DeepEqual(b, tt.args.newB) {
				t.Errorf("clone() = %v, \n"+
					"want %v", tt.args.newB, b)
			}
		})
	}
}

func TestBeaconCommitteeEngineV2_UpdateCommitteeState_MultipleInstructions(t *testing.T) {
	hash, _ := common.Hash{}.NewHashFromStr("123")
	tempHash, _ := common.Hash{}.NewHashFromStr("456")
	initTestParams()
	initLog()

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

	committeeChangeUnstakeSwap := NewCommitteeChange()
	committeeChangeUnstakeSwap.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}
	committeeChangeUnstakeSwap.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey0,
	}
	committeeChangeUnstakeSwap.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}

	committeeChangeUnstakeSwap.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey0,
	}
	committeeChangeUnstakeSwap.StopAutoStake = []string{
		key0,
	}
	committeeChangeUnstakeSwap.SlashingCommittee[0] = nil

	committeeChangeUnstakeSwapOut := NewCommitteeChange()
	committeeChangeUnstakeSwapOut.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey0,
	}
	committeeChangeUnstakeSwapOut.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}
	committeeChangeUnstakeSwapOut.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey0,
	}
	committeeChangeUnstakeSwapOut.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}
	committeeChangeUnstakeSwap.SlashingCommittee[0] = nil

	committeeChangeUnstakeSwapOut.StopAutoStake = []string{
		key0,
	}
	committeeChangeUnstakeSwapOut.SlashingCommittee[0] = nil

	committeeChangeUnstakeAndRandomTime := NewCommitteeChange()
	committeeChangeUnstakeAndRandomTime.NextEpochShardCandidateRemoved = []incognitokey.CommitteePublicKey{
		*incKey0,
	}

	committeeChangeUnstakeAndRandomTime.RemovedStaker = []string{key0}
	committeeChangeUnstakeAndRandomTime2 := NewCommitteeChange()
	committeeChangeUnstakeAndRandomTime2.StopAutoStake = []string{key0}

	committeeChangeStopAutoStakeAndRandomTime := NewCommitteeChange()
	committeeChangeStopAutoStakeAndRandomTime.StopAutoStake = []string{key0}

	committeeChangeTwoSwapOut := NewCommitteeChange()
	committeeChangeTwoSwapOut.ShardCommitteeRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey,
	}
	committeeChangeTwoSwapOut.ShardCommitteeRemoved[1] = []incognitokey.CommitteePublicKey{
		*incKey7,
	}
	committeeChangeTwoSwapOut.ShardCommitteeAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeTwoSwapOut.ShardCommitteeAdded[1] = []incognitokey.CommitteePublicKey{
		*incKey11,
	}
	committeeChangeTwoSwapOut.ShardSubstituteAdded[0] = []incognitokey.CommitteePublicKey{
		*incKey7,
	}
	committeeChangeTwoSwapOut.ShardSubstituteAdded[1] = []incognitokey.CommitteePublicKey{
		*incKey,
	}
	committeeChangeTwoSwapOut.ShardSubstituteRemoved[0] = []incognitokey.CommitteePublicKey{
		*incKey5,
	}
	committeeChangeTwoSwapOut.ShardSubstituteRemoved[1] = []incognitokey.CommitteePublicKey{
		*incKey11,
	}
	committeeChangeTwoSwapOut.SlashingCommittee[0] = nil
	committeeChangeTwoSwapOut.SlashingCommittee[1] = nil

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
			key:   true,
			key0:  true,
			key4:  true,
			key10: true,
			key7:  true,
		},
		map[string]common.Hash{
			key:   *tempHash,
			key0:  *hash,
			key4:  *hash,
			key10: *hash,
			key7:  *hash,
		},
	)

	finalMu := &sync.RWMutex{}
	unCommitteedMu := &sync.RWMutex{}

	type fields struct {
		beaconHeight                      uint64
		beaconHash                        common.Hash
		finalBeaconCommitteeStateV2       *BeaconCommitteeStateV2
		uncommittedBeaconCommitteeStateV2 *BeaconCommitteeStateV2
		version                           uint
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
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey5,
					},
					autoStake:                  map[string]bool{},
					rewardReceiver:             map[string]privacy.PaymentAddress{},
					stakingTx:                  map[string]common.Hash{},
					mu:                         finalMu,
					numberOfAssignedCandidates: 1,
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				beaconHash:                  *hash,
				version:                     SLASHING_VERSION,
				beaconHeight:                10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
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
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey0,
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
					mu: unCommitteedMu,
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
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey5,
					},
					autoStake:                  map[string]bool{},
					rewardReceiver:             map[string]privacy.PaymentAddress{},
					stakingTx:                  map[string]common.Hash{},
					mu:                         finalMu,
					numberOfAssignedCandidates: 1,
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				beaconHash:                  *hash,
				version:                     SLASHING_VERSION,
				beaconHeight:                10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
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
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey0,
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
					mu: unCommitteedMu,
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
			name: "Unstake And Assign 1, Fail to Unstake because Key in Current Epoch Candidate, only turn off auto stake flag",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey5,
						*incKey6,
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
					numberOfAssignedCandidates: 1,
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
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
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey6,
					},
					mu: unCommitteedMu,
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
			name: "Assign Then Unstake 2, Fail to Unstake because Key in Current Epoch Candidate, only turn off auto stake flag",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey5,
						*incKey6,
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
					numberOfAssignedCandidates: 1,
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
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
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey6,
					},
					mu: unCommitteedMu,
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
			name: "Unstake And Assign 3, Success to Unstake because Key in Next Epoch Candidate",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey6,
						*incKey0,
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
					numberOfAssignedCandidates: 1,
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey6,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              unCommitteedMu,
					autoStake:       map[string]bool{},
					rewardReceiver:  map[string]privacy.PaymentAddress{},
					stakingTx:       map[string]common.Hash{},
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
			name: "Unstake And Assign 4, Success to Unstake because Key in Next Epoch Candidate",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey6,
						*incKey0,
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
					numberOfAssignedCandidates: 1,
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey6,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              unCommitteedMu,
					autoStake:       map[string]bool{},
					rewardReceiver:  map[string]privacy.PaymentAddress{},
					stakingTx:       map[string]common.Hash{},
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
			name: "Unstake Then Swap 1, Failed to Unstake Swap Out key, Only turn off auto stake",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              finalMu,
					autoStake: map[string]bool{
						key0: true,
						key:  false,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey2, *incKey3, *incKey4, *incKey0,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              unCommitteedMu,
					autoStake: map[string]bool{
						key0: false,
						key:  false,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
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
							key,
							"0",
							"0",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          1,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeSwap,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Swap Then Unstake 2, Failed to Unstake Swap Out key, Only turn off auto stake",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              finalMu,
					autoStake: map[string]bool{
						key0: true,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey2, *incKey3, *incKey4, *incKey0,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              unCommitteedMu,
					autoStake: map[string]bool{
						key0: false,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key0,
							key,
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
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeSwap,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Is Beacon Random Time == False And Unstake",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey0,
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
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              unCommitteedMu,
					autoStake:       map[string]bool{},
					rewardReceiver:  map[string]privacy.PaymentAddress{},
					stakingTx:       map[string]common.Hash{},
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
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey0,
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
					numberOfAssignedCandidates: 0,
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey0,
					},
					mu: unCommitteedMu,
					autoStake: map[string]bool{
						key0: false,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
					},
					numberOfAssignedCandidates: 1,
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
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey0,
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
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{
						*incKey0,
					},
					mu: unCommitteedMu,
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
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              finalMu,
					autoStake: map[string]bool{
						key0: true,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey2, *incKey3, *incKey4, *incKey,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              unCommitteedMu,
					autoStake: map[string]bool{
						key0: false,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key,
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
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeSwapOut,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Unstake And Swap Out",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              finalMu,
					autoStake: map[string]bool{
						key0: true,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey2, *incKey3, *incKey4, *incKey,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              unCommitteedMu,
					autoStake: map[string]bool{
						key0: false,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
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
							key,
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
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeSwapOut,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Stop Auto Stake And Swap Out",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              finalMu,
					autoStake: map[string]bool{
						key0: true,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey2, *incKey3, *incKey4, *incKey,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              unCommitteedMu,
					autoStake: map[string]bool{
						key0: false,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key,
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
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeSwapOut,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Swap Out And Stop Auto Stake",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0, *incKey2, *incKey3, *incKey4,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              finalMu,
					autoStake: map[string]bool{
						key0: true,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey2, *incKey3, *incKey4, *incKey,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey0,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              unCommitteedMu,
					autoStake: map[string]bool{
						key0: false,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress0.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key:  *tempHash,
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
							key,
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
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeUnstakeSwapOut,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Double Swap Instruction for 2 shard, Back to substitutes",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
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
							*incKey5, *incKey13,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey11, *incKey12,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              finalMu,
					autoStake: map[string]bool{
						key0: true,
						key7: true,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey7.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key7: *hash,
						key:  *tempHash,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey2, *incKey3, *incKey4, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey8, *incKey9, *incKey10, *incKey11,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey13, *incKey7,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey12, *incKey,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              unCommitteedMu,
					autoStake: map[string]bool{
						key0: true,
						key7: true,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey7.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key7: *hash,
						key:  *tempHash,
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key5,
							key,
							"0",
							"0",
						},
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key11,
							key7,
							"1",
							"0",
						},
					},
					ConsensusStateDB:      sDB,
					ActiveShards:          2,
					MaxShardCommitteeSize: 4,
				},
			},
			want:    &BeaconCommitteeStateHash{},
			want1:   committeeChangeTwoSwapOut,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Double Swap Instruction for 2 shard, Slashing",
			fields: fields{
				beaconHash:   *hash,
				version:      SLASHING_VERSION,
				beaconHeight: 10,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
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
							*incKey5, *incKey13,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey11, *incKey12,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              finalMu,
					autoStake: map[string]bool{
						key0:  true,
						key4:  true,
						key10: true,
						key7:  true,
						key:   true,
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
						key:   *tempHash,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					mu: unCommitteedMu,
				},
			},
			fieldsAfterProcess: fields{
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					beaconCommittee: []incognitokey.CommitteePublicKey{},
					shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey, *incKey2, *incKey3, *incKey5,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey7, *incKey8, *incKey9, *incKey11,
						},
					},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{
							*incKey13,
						},
						1: []incognitokey.CommitteePublicKey{
							*incKey12,
						},
					},
					shardCommonPool: []incognitokey.CommitteePublicKey{},
					mu:              unCommitteedMu,
					autoStake: map[string]bool{
						key0: true,
						key7: true,
						key:  true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey0.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey7.GetIncKeyBase58(): paymentAddress0.KeySet.PaymentAddress,
						incKey.GetIncKeyBase58():  paymentAddress.KeySet.PaymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key0: *hash,
						key7: *hash,
						key:  *tempHash,
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
			engine := &BeaconCommitteeEngineV2{
				beaconHeight:                      tt.fields.beaconHeight,
				beaconHash:                        tt.fields.beaconHash,
				finalBeaconCommitteeStateV2:       tt.fields.finalBeaconCommitteeStateV2,
				uncommittedBeaconCommitteeStateV2: tt.fields.uncommittedBeaconCommitteeStateV2,
			}
			_, got1, got2, err := engine.UpdateCommitteeState(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV2.UpdateCommitteeState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeEngineV2.UpdateCommitteeState() got1 = %v, want1 = %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("BeaconCommitteeEngineV2.UpdateCommitteeState() got2 = %v, want2 = %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(tt.fields.uncommittedBeaconCommitteeStateV2,
				tt.fieldsAfterProcess.uncommittedBeaconCommitteeStateV2) {
				t.Errorf(`BeaconCommitteeEngineV2.UpdateCommitteeState() tt.fields.uncommittedBeaconCommitteeStateV2 = %v,
			tt.fieldsAfterProcess.uncommittedBeaconCommitteeStateV2 = %v`,
					tt.fields.uncommittedBeaconCommitteeStateV2, tt.fieldsAfterProcess.uncommittedBeaconCommitteeStateV2)
			}
		})
	}
}
