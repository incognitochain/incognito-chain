package committeestate

import (
	"reflect"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
)

func TestShardCommitteeStateV2_processInstructionFromBeacon(t *testing.T) {

	initPublicKey()
	initLog()

	type fields struct {
		shardCommittee  []incognitokey.CommitteePublicKey
		shardSubstitute []incognitokey.CommitteePublicKey
		mu              *sync.RWMutex
	}
	type args struct {
		env             ShardCommitteeStateEnvironment
		committeeChange *CommitteeChange
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
				shardCommittee: []incognitokey.CommitteePublicKey{
					*incKey, *incKey2, *incKey3, *incKey4,
				},
			},
			args: args{
				env: &shardCommitteeStateEnvironment{
					beaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key5,
							key,
							"0",
							"12",
							"0",
						},
					},
					numberOfFixedBlockValidators: 0,
				},
				committeeChange: &CommitteeChange{
					ShardCommitteeAdded: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
					ShardCommitteeRemoved: map[byte][]incognitokey.CommitteePublicKey{
						0: []incognitokey.CommitteePublicKey{},
					},
				},
			},
			want: &CommitteeChange{
				ShardCommitteeAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
				},
				ShardCommitteeRemoved: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShardCommitteeStateV2{
				shardCommittee:  tt.fields.shardCommittee,
				shardSubstitute: tt.fields.shardSubstitute,
				mu:              tt.fields.mu,
			}
			got, err := s.processInstructionFromBeacon(tt.args.env, tt.args.committeeChange)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShardCommitteeStateV2.processInstructionFromBeacon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ShardCommitteeStateV2.processInstructionFromBeacon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShardCommitteeEngineV2_ProcessInstructionFromBeacon(t *testing.T) {

	initPublicKey()
	initLog()

	hash, _ := common.Hash{}.NewHashFromStr("123")

	type fields struct {
		shardHeight                      uint64
		shardHash                        common.Hash
		shardID                          byte
		shardCommitteeStateV2            *ShardCommitteeStateV2
		uncommittedShardCommitteeStateV2 *ShardCommitteeStateV2
	}
	type args struct {
		env ShardCommitteeStateEnvironment
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
				shardHeight: 10,
				shardHash:   *hash,
				shardID:     0,
				shardCommitteeStateV2: &ShardCommitteeStateV2{
					shardCommittee: []incognitokey.CommitteePublicKey{
						*incKey, *incKey2, *incKey3, *incKey4,
					},
					mu: &sync.RWMutex{},
				},
				uncommittedShardCommitteeStateV2: &ShardCommitteeStateV2{
					shardCommittee: []incognitokey.CommitteePublicKey{
						*incKey, *incKey2, *incKey3, *incKey4,
					},
					mu: &sync.RWMutex{},
				},
			},
			args: args{
				env: &shardCommitteeStateEnvironment{
					beaconInstructions: [][]string{
						[]string{
							instruction.SWAP_SHARD_ACTION,
							key5,
							key,
							"0",
							"12",
							"0",
						},
					},
					numberOfFixedBlockValidators: 0,
				},
			},
			want:    NewCommitteeChange(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &ShardCommitteeEngineV2{
				shardHeight:                      tt.fields.shardHeight,
				shardHash:                        tt.fields.shardHash,
				shardID:                          tt.fields.shardID,
				shardCommitteeStateV2:            tt.fields.shardCommitteeStateV2,
				uncommittedShardCommitteeStateV2: tt.fields.uncommittedShardCommitteeStateV2,
			}
			got, err := engine.ProcessInstructionFromBeacon(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShardCommitteeEngineV2.ProcessInstructionFromBeacon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ShardCommitteeEngineV2.ProcessInstructionFromBeacon() = %v, want %v", got, tt.want)
			}
		})
	}
}
