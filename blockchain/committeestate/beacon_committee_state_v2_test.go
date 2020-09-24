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

// func TestBeaconCommitteeEngine_AssignShardsPoolUsingRandomInstruction(t *testing.T) {
// 	type fields struct {
// 		beaconHeight                      uint64
// 		beaconHash                        common.Hash
// 		beaconCommitteeStateV1            *BeaconCommitteeStateV1
// 		uncommittedBeaconCommitteeStateV1 *BeaconCommitteeStateV1
// 	}
// 	type args struct {
// 		seed        int64
// 		numShards   int
// 		subsSizeMap map[byte]int
// 		epoches     int
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   int
// 	}{
// 		{
// 			name:   "imbalance",
// 			fields: fields{},
// 			args: args{
// 				seed:      500,
// 				numShards: 4,
// 				subsSizeMap: map[byte]int{
// 					0: 5,
// 					1: 5,
// 					2: 5,
// 					3: 5,
// 				},
// 				epoches: 100,
// 			},
// 			want: 2,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			b := &BeaconCommitteeEngineV1{
// 				beaconHeight:                      tt.fields.beaconHeight,
// 				beaconHash:                        tt.fields.beaconHash,
// 				beaconCommitteeStateV1:            tt.fields.beaconCommitteeStateV1,
// 				uncommittedBeaconCommitteeStateV1: tt.fields.uncommittedBeaconCommitteeStateV1,
// 			}
// 			for i := 0; i < tt.args.epoches; i++ {
// 				_ = b.AssignShardsPoolUsingRandomInstruction(rand.Int63(), tt.args.numShards, SampleCandidateList(12), tt.args.subsSizeMap)

// 				fmt.Println(tt.args.subsSizeMap)
// 				diff := GetMinMaxRange(tt.args.subsSizeMap)
// 				if diff > tt.want {
// 					t.Errorf("BeaconCommitteeEngineV1.AssignShardsPoolUsingRandomInstruction() = %v, want %v", diff, tt.want)
// 				}
// 			}

// 		})
// 	}
// }

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
		mu                         *sync.RWMutex
	}
	type args struct {
		stakeInstruction *instruction.StakeInstruction
		committeeChange  *CommitteeChange
		env              *BeaconCommitteeStateEnvironment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CommitteeChange
		wantErr bool
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
			got, err := b.processStakeInstruction(tt.args.stakeInstruction, tt.args.committeeChange, tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV2.processStakeInstruction() = %v, want %v", got, tt.want)
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
		mu                         *sync.RWMutex
	}
	type args struct {
		rand            int64
		activeShards    int
		committeeChange *CommitteeChange
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *CommitteeChange
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
				numberOfRound:              map[string]int{},
			},
			args: args{
				rand:            10000,
				activeShards:    2,
				committeeChange: NewCommitteeChange(),
			},
			want: &CommitteeChange{
				NextEpochShardCandidateRemoved: []incognitokey.CommitteePublicKey{
					*incKey2,
				},
				ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
					1: []incognitokey.CommitteePublicKey{
						*incKey2,
						// *incKey3,
						// *incKey4,
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
				mu:                         tt.fields.mu,
			}
			if got := b.processAssignWithRandomInstruction(tt.args.rand, tt.args.activeShards, tt.args.committeeChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV2.processAssignWithRandomInstruction() = %v, want %v", got, tt.want)
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

func TestBeaconCommitteeEngineV2_GenerateAllRequestShardSwapInstruction(t *testing.T) {

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
		// {
		// 	name: "int((len(committees) + len(subtitutes)) / 3) < maxCommitteeSize",
		// 	fields: fields{
		// 		finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
		// 			shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
		// 				0: []incognitokey.CommitteePublicKey{
		// 					*incKey, *incKey2, *incKey3, *incKey4},
		// 			},
		// 			shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
		// 				0: []incognitokey.CommitteePublicKey{
		// 					*incKey5, *incKey6},
		// 			},
		// 		},
		// 	},
		// 	args: args{
		// 		env: &BeaconCommitteeStateEnvironment{
		// 			NumberOfFixedBlockNumberOfFixedShardBlockValidatorsValidator: 0,
		// 			Epoch:                       200,
		// 			RandomNumber:                10000,
		// 			MaxCommitteeSize:            5,
		// 			ActiveShards:                2,
		// 		},
		// 	},
		// 	want: []*instruction.RequestShardSwapInstruction{
		// 		&instruction.RequestShardSwapInstruction{
		// 			Epoch:         200,
		// 			ChainID:       0,
		// 			RandomNumber:  10000,
		// 			InPublicKeys:  []string{key5, key6},
		// 			OutPublicKeys: []string{key, key2},
		// 		},
		// 	},
		// 	wantErr: false,
		// },
		// {
		// 	name: "int((len(committees) + len(subtitutes)) / 3) >= maxCommitteeSize",
		// 	fields: fields{
		// 		finalBeaconCommitteeStateV2: &BeaconCommitteeStateV2{
		// 			shardCommittee: map[byte][]incognitokey.CommitteePublicKey{
		// 				0: []incognitokey.CommitteePublicKey{
		// 					*incKey, *incKey2, *incKey3, *incKey4},
		// 			},
		// 			shardSubstitute: map[byte][]incognitokey.CommitteePublicKey{
		// 				0: []incognitokey.CommitteePublicKey{
		// 					*incKey5, *incKey6},
		// 			},
		// 		},
		// 	},
		// 	args: args{
		// 		env: &BeaconCommitteeStateEnvironment{
		// 			NumberOfFixedShardBlockValidators: 0,
		// 			Epoch:                       200,
		// 			RandomNumber:                10000,
		// 			MaxCommitteeSize:            1,
		// 			ActiveShards:                2,
		// 		},
		// 	},
		// 	want: []*instruction.RequestShardSwapInstruction{
		// 		&instruction.RequestShardSwapInstruction{
		// 			Epoch:         200,
		// 			ChainID:       0,
		// 			RandomNumber:  10000,
		// 			InPublicKeys:  []string{key5},
		// 			OutPublicKeys: []string{key},
		// 		},
		// 	},
		// 	wantErr: false,
		// },
	}
	//TODO: @hung check this testcase pls
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
	statedb.StoreStakerInfoV1(
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
	)

	Logger.log.Info("key:", key)
	Logger.log.Info("key5:", key5)

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
					ConsensusStateDB: sDB,
				},
				committeeChange: &CommitteeChange{
					ShardSubstituteAdded:   map[byte][]incognitokey.CommitteePublicKey{},
					ShardSubstituteRemoved: map[byte][]incognitokey.CommitteePublicKey{},
					ShardCommitteeAdded:    map[byte][]incognitokey.CommitteePublicKey{},
					ShardCommitteeRemoved:  map[byte][]incognitokey.CommitteePublicKey{},
				},
			},
			want: &CommitteeChange{
				ShardSubstituteAdded:   map[byte][]incognitokey.CommitteePublicKey{},
				ShardSubstituteRemoved: map[byte][]incognitokey.CommitteePublicKey{},
				ShardCommitteeAdded:    map[byte][]incognitokey.CommitteePublicKey{},
				ShardCommitteeRemoved:  map[byte][]incognitokey.CommitteePublicKey{},
			},
			wantErr: true,
		},
		{
			name: "Swap In Not Valid In List Subtitutes Public Key",
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
					ConsensusStateDB: sDB,
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
					0: []incognitokey.CommitteePublicKey{},
				},
				ShardSubstituteRemoved: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
				},
				ShardCommitteeAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
				},
				ShardCommitteeRemoved: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
				},
				ShardCommitteeReplaced: map[byte][2][]incognitokey.CommitteePublicKey{
					// 0: []incognitokey.CommitteePublicKey{},
				},
				BeaconCommitteeReplaced: [2][]incognitokey.CommitteePublicKey{},
			},
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
					key: 2,
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

func TestBeaconCommitteeEngineV2_UpdateCommitteeState(t *testing.T) {
	hash, _ := common.Hash{}.NewHashFromStr("123")

	initPublicKey()
	initStateDB()
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
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("BeaconCommitteeEngineV2.UpdateCommitteeState() got = %v, want %v", got, tt.want)
			// }
			// if !reflect.DeepEqual(got1, tt.want1) {
			// 	t.Errorf("BeaconCommitteeEngineV2.UpdateCommitteeState() got1 = %v, want %v", got1, tt.want1)
			// }
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("BeaconCommitteeEngineV2.UpdateCommitteeState() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
