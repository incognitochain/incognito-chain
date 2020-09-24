package committeestate

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/instructionsprocessor"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
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

	initStateDB()
	initPublicKey()

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
		numberOfRound              map[string]int
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
				shardCommittee:  map[byte][]incognitokey.CommitteePublicKey{},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{},
				autoStake:       map[string]bool{},
				rewardReceiver:  map[string]privacy.PaymentAddress{},
				stakingTx:       map[string]common.Hash{},
				numberOfRound:   map[string]int{},
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
				numberOfRound: map[string]int{
					key: 0,
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
				numberOfRound:              tt.fields.numberOfRound,
			}
			got, err := b.processStakeInstruction(tt.args.stakeInstruction, tt.args.committeeChange, tt.args.env)
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
			if !reflect.DeepEqual(b.numberOfRound, tt.wantSideEffect.numberOfRound) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), numberOfRound = %v, want %v", got, tt.want)
			}
			stakerInfo, has, err := statedb.GetStakerInfo(tt.args.env.ConsensusStateDB, key)
			if err != nil {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), StoreStakerInfo err, %+v", err)
			}
			if !has {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), StoreStakerInfo not found, %+v", key)
			}
			if !reflect.DeepEqual(stakerInfo.RewardReceiver(), paymentAddress) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), stakerInfo.RewardReceiver() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(stakerInfo.AutoStaking(), true) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), stakerInfo.AutoStaking() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(stakerInfo.TxStakingID(), *txHash) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), stakerInfo.TxStakingID() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(stakerInfo.NumberOfRound(), 0) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction(), stakerInfo.NumberOfRound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconCommitteeStateV2_processAssignWithRandomInstruction(t *testing.T) {

	initLog()
	initPublicKey()

	type fields struct {
		beaconCommittee            []incognitokey.CommitteePublicKey
		shardCommittee             map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute            map[byte][]incognitokey.CommitteePublicKey
		shardCommonPool            []incognitokey.CommitteePublicKey
		numberOfAssignedCandidates int
		autoStake                  map[string]bool
		rewardReceiver             map[string]privacy.PaymentAddress
		stakingTx                  map[string]common.Hash
		numberOfRound              map[string]int
	}
	type args struct {
		rand            int64
		activeShards    int
		committeeChange *CommitteeChange
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
				numberOfRound: map[string]int{
					key2: 0,
					key:  1,
					key5: 1,
					key6: 1,
				},
			},
			args: args{
				rand:            10000,
				activeShards:    2,
				committeeChange: NewCommitteeChange(),
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
				numberOfRound: map[string]int{
					key2: 1,
					key:  1,
					key5: 1,
					key6: 1,
				},
			},
			want: &CommitteeChange{
				NextEpochShardCandidateRemoved: []incognitokey.CommitteePublicKey{
					*incKey2,
				},
				ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
					1: []incognitokey.CommitteePublicKey{
						*incKey2,
					},
					2: []incognitokey.CommitteePublicKey{},
					3: []incognitokey.CommitteePublicKey{},
					4: []incognitokey.CommitteePublicKey{},
					5: []incognitokey.CommitteePublicKey{},
					6: []incognitokey.CommitteePublicKey{},
					7: []incognitokey.CommitteePublicKey{},
				},
				ShardSubstituteRemoved: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
					1: []incognitokey.CommitteePublicKey{},
					2: []incognitokey.CommitteePublicKey{},
					3: []incognitokey.CommitteePublicKey{},
					4: []incognitokey.CommitteePublicKey{},
					5: []incognitokey.CommitteePublicKey{},
					6: []incognitokey.CommitteePublicKey{},
					7: []incognitokey.CommitteePublicKey{},
				},
				ShardCommitteeAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
					1: []incognitokey.CommitteePublicKey{},
					2: []incognitokey.CommitteePublicKey{},
					3: []incognitokey.CommitteePublicKey{},
					4: []incognitokey.CommitteePublicKey{},
					5: []incognitokey.CommitteePublicKey{},
					6: []incognitokey.CommitteePublicKey{},
					7: []incognitokey.CommitteePublicKey{},
				},
				ShardCommitteeRemoved: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
					1: []incognitokey.CommitteePublicKey{},
					2: []incognitokey.CommitteePublicKey{},
					3: []incognitokey.CommitteePublicKey{},
					4: []incognitokey.CommitteePublicKey{},
					5: []incognitokey.CommitteePublicKey{},
					6: []incognitokey.CommitteePublicKey{},
					7: []incognitokey.CommitteePublicKey{},
				},
				ShardCommitteeReplaced: map[byte][2][]incognitokey.CommitteePublicKey{},
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
				numberOfRound:              tt.fields.numberOfRound,
			}
			if got := b.processAssignWithRandomInstruction(tt.args.rand, tt.args.activeShards, tt.args.committeeChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV2.processAssignWithRandomInstruction() = %v, want %v", got, tt.want)
			}
			for k, gotV := range b.numberOfRound {
				wantV := tt.wantSideEffect.numberOfRound[k]
				if wantV != gotV {
					t.Errorf("BeaconCommitteeStateV2.processAssignWithRandomInstruction(), numberOfRound = %v, want %v", gotV, wantV)
				}
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

	initPublicKey()
	initLog()

	type args struct {
		shardCommonPool   []incognitokey.CommitteePublicKey
		shardCommittee    map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute   map[byte][]incognitokey.CommitteePublicKey
		maxAssignPerShard int
	}
	tests := []struct {
		name                           string
		args                           args
		wantNumberOfAssignedCandidates int
	}{
		{
			name: "maxAssignPerShard >= len(shardcommittes + subtitutes)",
			args: args{
				shardCommonPool: []incognitokey.CommitteePublicKey{
					*incKey2,
					*incKey3,
					*incKey4,
				},
				shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey,
					},
				},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey5, *incKey6,
					},
				},
				maxAssignPerShard: 5,
			},
			wantNumberOfAssignedCandidates: 1,
		},
		{
			name: "maxAssignPerShard < len(shardcommittes + subtitutes)",
			args: args{
				shardCommonPool: []incognitokey.CommitteePublicKey{
					*incKey7,
					*incKey8,
				},
				shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey,
						*incKey2,
						*incKey3,
					},
				},
				shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey4, *incKey5, *incKey6,
					},
				},
				maxAssignPerShard: 1,
			},
			wantNumberOfAssignedCandidates: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotNumberOfAssignedCandidates := SnapshotShardCommonPoolV2(tt.args.shardCommonPool, tt.args.shardCommittee, tt.args.shardSubstitute, tt.args.maxAssignPerShard); gotNumberOfAssignedCandidates != tt.wantNumberOfAssignedCandidates {
				t.Errorf("SnapshotShardCommonPoolV2() = %v, want %v", gotNumberOfAssignedCandidates, tt.wantNumberOfAssignedCandidates)
			}
		})
	}
}

// TODO: @tin write unit test
func TestBeaconCommitteeEngineV2_GenerateAllSwapShardInstructions(t *testing.T) {

	initPublicKey()
	initLog()

	type fields struct {
		beaconHeight                      uint64
		beaconHash                        common.Hash
		finalBeaconCommitteeStateV2       *BeaconCommitteeStateV2
		uncommittedBeaconCommitteeStateV2 *BeaconCommitteeStateV2
		insProcessor                      *instructionsprocessor.BInsProcessor
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
					NumberOfFixedShardBlockValidators: 0,
					ActiveShards:                      2,
				},
			},
			want:    []*instruction.SwapShardInstruction{},
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeEngineV2.GenerateAllRequestShardSwapInstruction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconCommitteeStateV2_processSwapShardInstruction(t *testing.T) {

	initPublicKey()
	initLog()
	initStateDB()

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	rewardReceiverkey := incKey.GetIncKeyBase58()

	hash, err := common.Hash{}.NewHashFromStr("123")
	hash6, err := common.Hash{}.NewHashFromStr("456")
	statedb.StoreStakerInfoV2(
		sDB,
		[]incognitokey.CommitteePublicKey{*incKey, *incKey6},
		map[string]privacy.PaymentAddress{
			rewardReceiverkey: paymentAddress,
			rewardReceiverkey: paymentAddress,
		},
		map[string]bool{
			key:  true,
			key6: false,
		},
		map[string]common.Hash{
			key:  *hash,
			key6: *hash6,
		},
		map[string]int{
			key:  0,
			key6: 0,
		},
	)

	type fields struct {
		beaconCommittee            []incognitokey.CommitteePublicKey
		shardCommittee             map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute            map[byte][]incognitokey.CommitteePublicKey
		shardCommonPool            []incognitokey.CommitteePublicKey
		numberOfAssignedCandidates int
		autoStake                  map[string]bool
		rewardReceiver             map[string]privacy.PaymentAddress
		stakingTx                  map[string]common.Hash
		numberOfRound              map[string]int
		mu                         *sync.RWMutex
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
		wantErr bool
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
				numberOfRound:  map[string]int{},
				autoStake:      map[string]bool{},
				rewardReceiver: map[string]privacy.PaymentAddress{},
				stakingTx:      map[string]common.Hash{},
			},
			args: args{
				swapShardInstruction: &instruction.SwapShardInstruction{
					OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey6,
					},
				},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                  sDB,
					NumberOfFixedShardBlockValidators: 0,
				},
				committeeChange: &CommitteeChange{
					ShardSubstituteAdded:   map[byte][]incognitokey.CommitteePublicKey{},
					ShardSubstituteRemoved: map[byte][]incognitokey.CommitteePublicKey{},
					ShardCommitteeAdded:    map[byte][]incognitokey.CommitteePublicKey{},
					ShardCommitteeRemoved:  map[byte][]incognitokey.CommitteePublicKey{},
				},
			},
			want:    NewCommitteeChange(),
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
				numberOfRound:  map[string]int{},
				autoStake:      map[string]bool{},
				rewardReceiver: map[string]privacy.PaymentAddress{},
				stakingTx:      map[string]common.Hash{},
			},
			args: args{
				swapShardInstruction: &instruction.SwapShardInstruction{
					InPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey6,
					},
				},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                  sDB,
					NumberOfFixedShardBlockValidators: 0,
				},
				committeeChange: &CommitteeChange{
					ShardSubstituteAdded:   map[byte][]incognitokey.CommitteePublicKey{},
					ShardSubstituteRemoved: map[byte][]incognitokey.CommitteePublicKey{},
					ShardCommitteeAdded:    map[byte][]incognitokey.CommitteePublicKey{},
					ShardCommitteeRemoved:  map[byte][]incognitokey.CommitteePublicKey{},
				},
			},
			want:    NewCommitteeChange(),
			wantErr: true,
		},
		{
			name: "Valid Input [Back Directly To This Shard Pool]",
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
				numberOfRound: map[string]int{
					key: 0,
				},
				autoStake:      map[string]bool{},
				rewardReceiver: map[string]privacy.PaymentAddress{},
				stakingTx:      map[string]common.Hash{},
			},
			args: args{
				swapShardInstruction: &instruction.SwapShardInstruction{
					OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey,
					},
					InPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey5,
					},
					OutPublicKeys: []string{key},
				},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                  sDB,
					NumberOfFixedShardBlockValidators: 0,
				},
				committeeChange: &CommitteeChange{
					ShardSubstituteAdded:   map[byte][]incognitokey.CommitteePublicKey{},
					ShardSubstituteRemoved: map[byte][]incognitokey.CommitteePublicKey{},
					ShardCommitteeAdded:    map[byte][]incognitokey.CommitteePublicKey{},
					ShardCommitteeRemoved:  map[byte][]incognitokey.CommitteePublicKey{},
				},
			},
			want: &CommitteeChange{
				ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey,
					},
				},
				ShardSubstituteRemoved: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey5,
					},
				},
				ShardCommitteeAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey5,
					},
				},
				ShardCommitteeRemoved: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey,
					},
				},
			},
			wantErr: false,
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
				numberOfRound: map[string]int{
					key: 1,
				},
				autoStake:      map[string]bool{},
				rewardReceiver: map[string]privacy.PaymentAddress{},
				stakingTx:      map[string]common.Hash{},
			},
			args: args{
				swapShardInstruction: &instruction.SwapShardInstruction{
					OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey,
					},
					InPublicKeyStructs: []incognitokey.CommitteePublicKey{
						*incKey5,
					},
					OutPublicKeys: []string{key},
				},
				env: &BeaconCommitteeStateEnvironment{
					NumberOfFixedShardBlockValidators: 0,
					ConsensusStateDB:                  sDB,
					RandomNumber:                      5000,
					ActiveShards:                      1,
				},
				committeeChange: &CommitteeChange{
					ShardSubstituteAdded:   map[byte][]incognitokey.CommitteePublicKey{},
					ShardSubstituteRemoved: map[byte][]incognitokey.CommitteePublicKey{},
					ShardCommitteeAdded:    map[byte][]incognitokey.CommitteePublicKey{},
					ShardCommitteeRemoved:  map[byte][]incognitokey.CommitteePublicKey{},
				},
			},
			want: &CommitteeChange{
				ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey,
					},
				},
				ShardSubstituteRemoved: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey5,
					},
				},
				ShardCommitteeAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey5,
					},
				},
				ShardCommitteeRemoved: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey,
					},
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
				numberOfRound:              tt.fields.numberOfRound,
				mu:                         tt.fields.mu,
			}
			got, err := b.processSwapShardInstruction(tt.args.swapShardInstruction, tt.args.env, tt.args.committeeChange)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV2.processSwapShardInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV2.processSwapShardInstruction() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TODO: @tin write unit test
func TestBeaconCommitteeEngineV2_UpdateCommitteeState(t *testing.T) {
	hash, _ := common.Hash{}.NewHashFromStr("123")

	initPublicKey()
	initLog()

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
		name    string
		fields  fields
		args    args
		want    *BeaconCommitteeStateHash
		want1   *CommitteeChange
		want2   [][]string
		wantErr bool
	}{
		{
			name: "Process Swap Shard Instructions",
			fields: fields{
				beaconHeight: 5,
				beaconHash:   *hash,
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
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
					mu:             &sync.RWMutex{},
					autoStake:      map[string]bool{},
					rewardReceiver: map[string]privacy.PaymentAddress{},
					stakingTx:      map[string]common.Hash{},
					numberOfRound:  map[string]int{},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
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
					mu:             &sync.RWMutex{},
					autoStake:      map[string]bool{},
					rewardReceiver: map[string]privacy.PaymentAddress{},
					stakingTx:      map[string]common.Hash{},
					numberOfRound:  map[string]int{},
				},
				version: SLASHING_VERSION,
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key5,
							key,
							"0",
							"120",
							"0",
						},
					},
					RandomNumber: 5000,
				},
			},
			want: &BeaconCommitteeStateHash{},
			want1: &CommitteeChange{
				ShardCommitteeAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey5,
					},
				},
				ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey,
					},
				},
				ShardSubstituteRemoved: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey5,
					},
				},
				ShardCommitteeRemoved: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{
						*incKey,
					},
				},
			},
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
			_, _, got2, err := engine.UpdateCommitteeState(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV2.UpdateCommitteeState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("BeaconCommitteeEngineV2.UpdateCommitteeState() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

// TODO: @tin write unit test
func TestBeaconCommitteeStateV2_processAfterSwap(t *testing.T) {

	initPublicKey()
	initLog()
	initStateDB()

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	rewardReceiverkey := incKey.GetIncKeyBase58()

	hash, err := common.Hash{}.NewHashFromStr("123")
	hash6, err := common.Hash{}.NewHashFromStr("456")
	statedb.StoreStakerInfoV2(
		sDB,
		[]incognitokey.CommitteePublicKey{*incKey, *incKey6},
		map[string]privacy.PaymentAddress{
			rewardReceiverkey: paymentAddress,
			rewardReceiverkey: paymentAddress,
		},
		map[string]bool{
			key:  true,
			key6: false,
		},
		map[string]common.Hash{
			key:  *hash,
			key6: *hash6,
		},
		map[string]int{
			key:  0,
			key6: 0,
		},
	)
	type fields struct {
		beaconCommittee            []incognitokey.CommitteePublicKey
		shardCommittee             map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute            map[byte][]incognitokey.CommitteePublicKey
		shardCommonPool            []incognitokey.CommitteePublicKey
		numberOfAssignedCandidates int
		autoStake                  map[string]bool
		rewardReceiver             map[string]privacy.PaymentAddress
		stakingTx                  map[string]common.Hash
		numberOfRound              map[string]int
		mu                         *sync.RWMutex
	}
	type args struct {
		env                 *BeaconCommitteeStateEnvironment
		outPublicKeys       []string
		outPublicKeyStructs []incognitokey.CommitteePublicKey
		shardID             byte
		committeeChange     *CommitteeChange
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CommitteeChange
		wantErr bool
	}{
		// TODO: Add test cases.
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
				numberOfRound:              tt.fields.numberOfRound,
				mu:                         tt.fields.mu,
			}
			got, err := b.processAfterSwap(tt.args.env, tt.args.outPublicKeys, tt.args.outPublicKeyStructs, tt.args.shardID, tt.args.committeeChange)
			if (err != nil) != tt.wantErr {
				t.Errorf("processAfterSwap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processAfterSwap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconCommitteeEngineV2_BuildIncurredInstructions(t *testing.T) {

	initPublicKey()
	initStateDB()
	initLog()

	sDB, _ := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	rewardReceiverkey := incKey.GetIncKeyBase58()
	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})

	hash, _ := common.Hash{}.NewHashFromStr("123")
	err := statedb.StoreStakerInfoV2(
		sDB,
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
		map[string]int{
			key: 0,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
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
		want    [][]string
		wantErr bool
	}{
		{
			name: "Environment Is Null",
			fields: fields{
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
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
						key: *hash,
					},
					numberOfRound: map[string]int{
						key: 0,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
			},
			args:    args{},
			want:    [][]string{},
			wantErr: true,
		},
		{
			name: "Length Of Beacon Instructions Is 0",
			fields: fields{
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
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
						key: *hash,
					},
					numberOfRound: map[string]int{
						key: 0,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{},
				},
			},
			want:    [][]string{},
			wantErr: false,
		},

		{
			name: "Invalid Unstake Instruction Format",
			fields: fields{
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
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
						key: *hash,
					},
					numberOfRound: map[string]int{
						key: 0,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{instruction.UNSTAKE_ACTION},
					},
				},
			},
			want:    [][]string{},
			wantErr: true,
		},
		{
			name: "Error In Processing Unstake Instruction",
			fields: fields{
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
					shardCommittee:  map[byte][]incognitokey.CommitteePublicKey{},
					shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{},
					shardCommonPool: []incognitokey.CommitteePublicKey{*incKey2},
					autoStake: map[string]bool{
						key2: true,
					},
					rewardReceiver: map[string]privacy.PaymentAddress{
						incKey2.GetIncKeyBase58(): paymentAddress,
					},
					stakingTx: map[string]common.Hash{
						key2: *hash,
					},
					numberOfRound: map[string]int{
						key2: 0,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key2,
						},
					},
					ConsensusStateDB:     sDB,
					unassignedCommonPool: []string{key2},
				},
			},
			want:    [][]string{},
			wantErr: true,
		},
		{
			name: "Valid Input",
			fields: fields{
				finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
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
						key: *hash,
					},
					numberOfRound: map[string]int{
						key: 0,
					},
				},
				uncommittedBeaconCommitteeStateV2: &BeaconCommitteeStateV2{},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							key,
						},
					},
					ConsensusStateDB:     sDB,
					unassignedCommonPool: []string{key},
				},
			},
			want: [][]string{
				[]string{
					instruction.RETURN_ACTION,
					key,
					"0",
					hash.String(),
					"100",
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
			got, err := engine.BuildIncurredInstructions(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildIncurredInstructions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildIncurredInstructions() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconCommitteeStateV2_processUnstakeInstruction(t *testing.T) {

	// Init data for testcases
	initStateDB()
	initPublicKey()

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	rewardReceiverkey := incKey.GetIncKeyBase58()
	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})

	validSDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreStakerInfoV2(
		validSDB,
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
		map[string]int{
			key: 0,
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
		numberOfRound              map[string]int
		mu                         *sync.RWMutex
	}
	type args struct {
		unstakeInstruction *instruction.UnstakeInstruction
		env                *BeaconCommitteeStateEnvironment
		committeeChange    *CommitteeChange
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CommitteeChange
		want1   [][]string
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
					unassignedCommonPool: []string{"123"},
				},
				committeeChange: &CommitteeChange{},
			},
			want:    &CommitteeChange{},
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
				numberOfRound: map[string]int{
					key: 0,
				},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys:       []string{key2},
					CommitteePublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey2},
				},
				env: &BeaconCommitteeStateEnvironment{
					unassignedCommonPool: []string{key2},
					ConsensusStateDB:     sDB,
				},
				committeeChange: &CommitteeChange{},
			},
			want:    &CommitteeChange{},
			wantErr: true,
		},
		{
			name: "Valid Input Key In Subtitutes List",
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
				numberOfRound: map[string]int{
					key: 0,
				},
			},
			args: args{
				committeeChange: &CommitteeChange{},
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys:       []string{key},
					CommitteePublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey},
				},
				env: &BeaconCommitteeStateEnvironment{
					unassignedCommonPool: []string{key},
					ConsensusStateDB:     validSDB,
				},
			},
			want: &CommitteeChange{
				NextEpochShardCandidateRemoved: []incognitokey.CommitteePublicKey{*incKey},
			},
			want1: [][]string{
				[]string{
					instruction.RETURN_ACTION,
					key,
					"0",
					hash.String(),
					"100",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid Input Key In Validators List",
			fields: fields{
				shardCommonPool: []incognitokey.CommitteePublicKey{*incKey},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys:       []string{key},
					CommitteePublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey},
				},
				env: &BeaconCommitteeStateEnvironment{
					allSubstituteCommittees: []string{key},
				},
				committeeChange: &CommitteeChange{},
			},
			want:    &CommitteeChange{},
			want1:   [][]string{},
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
				numberOfRound:              tt.fields.numberOfRound,
				mu:                         tt.fields.mu,
			}
			got, got1, err := b.processUnstakeInstruction(tt.args.unstakeInstruction, tt.args.env, tt.args.committeeChange)
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

// TODO: @tin write unittest
func TestBeaconCommitteeStateV2_processStopAutoStakeInstruction(t *testing.T) {
	type fields struct {
		beaconCommittee            []incognitokey.CommitteePublicKey
		shardCommittee             map[byte][]incognitokey.CommitteePublicKey
		shardSubstitute            map[byte][]incognitokey.CommitteePublicKey
		shardCommonPool            []incognitokey.CommitteePublicKey
		numberOfAssignedCandidates int
		autoStake                  map[string]bool
		rewardReceiver             map[string]privacy.PaymentAddress
		stakingTx                  map[string]common.Hash
		numberOfRound              map[string]int
		mu                         *sync.RWMutex
	}
	type args struct {
		stopAutoStakeInstruction *instruction.StopAutoStakeInstruction
		env                      *BeaconCommitteeStateEnvironment
		committeeChange          *CommitteeChange
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
			b := &BeaconCommitteeStateV2{
				beaconCommittee:            tt.fields.beaconCommittee,
				shardCommittee:             tt.fields.shardCommittee,
				shardSubstitute:            tt.fields.shardSubstitute,
				shardCommonPool:            tt.fields.shardCommonPool,
				numberOfAssignedCandidates: tt.fields.numberOfAssignedCandidates,
				autoStake:                  tt.fields.autoStake,
				rewardReceiver:             tt.fields.rewardReceiver,
				stakingTx:                  tt.fields.stakingTx,
				numberOfRound:              tt.fields.numberOfRound,
				mu:                         tt.fields.mu,
			}
			if got := b.processStopAutoStakeInstruction(tt.args.stopAutoStakeInstruction, tt.args.env, tt.args.committeeChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processStopAutoStakeInstruction() = %v, want %v", got, tt.want)
			}
		})
	}
}
