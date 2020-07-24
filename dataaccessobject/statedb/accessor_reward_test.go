package statedb

import (
	"github.com/incognitochain/incognito-chain/common"
	"log"
	"testing"
)

func TestAddShardRewardRequest(t *testing.T) {
	sDB, _ := NewWithPrefixTrie(emptyRoot, wrarperDB)
	type args struct {
		stateDB      *StateDB
		epoch        uint64
		shardID      byte
		tokenID      common.Hash
		rewardAmount uint64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "epoch 1, add 100",
			args: args{
				stateDB:      sDB,
				epoch:        1,
				shardID:      0,
				tokenID:      common.PRVCoinID,
				rewardAmount: 100,
			},
			wantErr: false,
		},
		{
			name: "epoch 2, add 200",
			args: args{
				stateDB:      sDB,
				epoch:        2,
				shardID:      0,
				tokenID:      common.PRVCoinID,
				rewardAmount: 200,
			},
			wantErr: false,
		},
		{
			name: "epoch 2, add 200",
			args: args{
				stateDB:      sDB,
				epoch:        2,
				shardID:      0,
				tokenID:      common.PRVCoinID,
				rewardAmount: 400,
			},
			wantErr: false,
		},
		{
			name: "epoch 3, add 300",
			args: args{
				stateDB:      sDB,
				epoch:        3,
				shardID:      0,
				tokenID:      common.PRVCoinID,
				rewardAmount: 300,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddShardRewardRequest(tt.args.stateDB, tt.args.epoch, tt.args.shardID, tt.args.tokenID, tt.args.rewardAmount); (err != nil) != tt.wantErr {
				t.Errorf("AddShardRewardRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetRewardOfShardByEpoch(t *testing.T) {
	sDB, _ := NewWithPrefixTrie(emptyRoot, wrarperDB)
	type args struct {
		stateDB *StateDB
		epoch   uint64
		shardID byte
		tokenID common.Hash
	}
	type addArgs struct {
		stateDB      *StateDB
		epoch        uint64
		shardID      byte
		tokenID      common.Hash
		rewardAmount uint64
	}
	addArgss := []addArgs{
		addArgs{
			epoch:        1,
			shardID:      0,
			tokenID:      common.PRVCoinID,
			rewardAmount: 100,
		},
		addArgs{
			epoch:        2,
			shardID:      0,
			tokenID:      common.PRVCoinID,
			rewardAmount: 200,
		},
		addArgs{
			epoch:        2,
			shardID:      0,
			tokenID:      common.PRVCoinID,
			rewardAmount: 150,
		},
		addArgs{
			epoch:        3,
			shardID:      0,
			tokenID:      common.PRVCoinID,
			rewardAmount: 300,
		},
	}
	for _, add := range addArgss {
		if err := AddShardRewardRequest(sDB, add.epoch, add.shardID, add.tokenID, add.rewardAmount); err != nil {
			log.Fatal(err)
		}
	}
	rootHash, _ := sDB.Commit(true)
	_ = sDB.Database().TrieDB().Commit(rootHash, true)

	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "epoch 1, 100",
			args: args{
				stateDB: sDB,
				epoch:   1,
				shardID: 0,
				tokenID: common.PRVCoinID,
			},
			wantErr: false,
			want:    100,
		},
		{
			name: "epoch 2, 350",
			args: args{
				stateDB: sDB,
				epoch:   2,
				shardID: 0,
				tokenID: common.PRVCoinID,
			},
			wantErr: false,
			want:    350,
		},
		{
			name: "epoch 3, 300",
			args: args{
				stateDB: sDB,
				epoch:   3,
				shardID: 0,
				tokenID: common.PRVCoinID,
			},
			wantErr: false,
			want:    300,
		},
		{
			name: "epoch 4, 0",
			args: args{
				stateDB: sDB,
				epoch:   4,
				shardID: 0,
				tokenID: common.PRVCoinID,
			},
			wantErr: false,
			want:    0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRewardOfShardByEpoch(tt.args.stateDB, tt.args.epoch, tt.args.shardID, tt.args.tokenID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRewardOfShardByEpoch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetRewardOfShardByEpoch() got = %v, want %v", got, tt.want)
			}
		})
	}
}
