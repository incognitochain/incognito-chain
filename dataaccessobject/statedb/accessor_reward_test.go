package statedb

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"log"
	"reflect"
	"sort"
	"testing"
)

func TestAddShardRewardRequestV3(t *testing.T) {
	sDB, _ := NewWithPrefixTrie(emptyRoot, wrarperDB)
	type args struct {
		stateDB      *StateDB
		epoch        uint64
		shardID      byte
		subsetID     byte
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
				subsetID:     0,
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
				subsetID:     1,
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
				subsetID:     0,
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
				subsetID:     1,
				tokenID:      common.PRVCoinID,
				rewardAmount: 300,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddShardRewardRequestV3(tt.args.stateDB, tt.args.epoch, tt.args.shardID, tt.args.subsetID, tt.args.tokenID, tt.args.rewardAmount); (err != nil) != tt.wantErr {
				t.Errorf("AddShardRewardRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetRewardOfShardByEpochV3(t *testing.T) {
	sDB, _ := NewWithPrefixTrie(emptyRoot, wrarperDB)
	type addArgs struct {
		stateDB      *StateDB
		epoch        uint64
		shardID      byte
		subsetID     byte
		tokenID      common.Hash
		rewardAmount uint64
	}
	addArgss := []addArgs{
		addArgs{
			epoch:        1,
			shardID:      0,
			subsetID:     0,
			tokenID:      common.PRVCoinID,
			rewardAmount: 100,
		},
		addArgs{
			epoch:        1,
			shardID:      0,
			subsetID:     1,
			tokenID:      common.PRVCoinID,
			rewardAmount: 150,
		},
		addArgs{
			epoch:        2,
			shardID:      0,
			subsetID:     0,
			tokenID:      common.PRVCoinID,
			rewardAmount: 300,
		},
		addArgs{
			epoch:        2,
			shardID:      0,
			subsetID:     0,
			tokenID:      common.PRVCoinID,
			rewardAmount: 150,
		},
		addArgs{
			epoch:        2,
			shardID:      0,
			subsetID:     1,
			tokenID:      common.PRVCoinID,
			rewardAmount: 300,
		},
		addArgs{
			epoch:        2,
			shardID:      0,
			subsetID:     1,
			tokenID:      common.PRVCoinID,
			rewardAmount: 450,
		},
		addArgs{
			epoch:        3,
			shardID:      0,
			subsetID:     0,
			tokenID:      common.PRVCoinID,
			rewardAmount: 300,
		},
		addArgs{
			epoch:        3,
			shardID:      0,
			subsetID:     1,
			tokenID:      common.PRVCoinID,
			rewardAmount: 400,
		},
		addArgs{
			epoch:        3,
			shardID:      1,
			subsetID:     0,
			tokenID:      common.PRVCoinID,
			rewardAmount: 350,
		},
		addArgs{
			epoch:        3,
			shardID:      1,
			subsetID:     1,
			tokenID:      common.PRVCoinID,
			rewardAmount: 450,
		},
	}
	for _, add := range addArgss {
		if err := AddShardRewardRequestV3(sDB, add.epoch, add.shardID, add.subsetID, add.tokenID, add.rewardAmount); err != nil {
			log.Fatal(err)
		}
	}
	rootHash, _ := sDB.Commit(true)
	_ = sDB.Database().TrieDB().Commit(rootHash, true)

	type args struct {
		stateDB  *StateDB
		epoch    uint64
		shardID  byte
		subsetID byte
		tokenID  common.Hash
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "epoch 1, subset 0, 100",
			args: args{
				stateDB:  sDB,
				epoch:    1,
				shardID:  0,
				subsetID: 0,
				tokenID:  common.PRVCoinID,
			},
			wantErr: false,
			want:    100,
		},
		{
			name: "epoch 1, subset 1, 150",
			args: args{
				stateDB:  sDB,
				epoch:    1,
				shardID:  0,
				subsetID: 1,
				tokenID:  common.PRVCoinID,
			},
			wantErr: false,
			want:    150,
		},
		{
			name: "epoch 2, subset 0, 450",
			args: args{
				stateDB:  sDB,
				epoch:    2,
				shardID:  0,
				subsetID: 0,
				tokenID:  common.PRVCoinID,
			},
			wantErr: false,
			want:    450,
		},
		{
			name: "epoch 2, subset 1, 750",
			args: args{
				stateDB:  sDB,
				epoch:    2,
				shardID:  0,
				subsetID: 1,
				tokenID:  common.PRVCoinID,
			},
			wantErr: false,
			want:    750,
		},
		{
			name: "epoch 3, subset 0, 300",
			args: args{
				stateDB:  sDB,
				epoch:    3,
				shardID:  0,
				subsetID: 0,
				tokenID:  common.PRVCoinID,
			},
			wantErr: false,
			want:    300,
		}, {
			name: "epoch 3, subset 1, 400",
			args: args{
				stateDB:  sDB,
				epoch:    3,
				shardID:  0,
				subsetID: 1,
				tokenID:  common.PRVCoinID,
			},
			wantErr: false,
			want:    400,
		},
		{
			name: "epoch 3, subset 0, shard 1, 350",
			args: args{
				stateDB:  sDB,
				epoch:    3,
				shardID:  1,
				subsetID: 0,
				tokenID:  common.PRVCoinID,
			},
			wantErr: false,
			want:    350,
		}, {
			name: "epoch 3, subset 1, shard 1, 450",
			args: args{
				stateDB:  sDB,
				epoch:    3,
				shardID:  1,
				subsetID: 1,
				tokenID:  common.PRVCoinID,
			},
			wantErr: false,
			want:    450,
		},
		{
			name: "epoch 4, subset 0, 0",
			args: args{
				stateDB:  sDB,
				epoch:    4,
				shardID:  0,
				subsetID: 0,
				tokenID:  common.PRVCoinID,
			},
			wantErr: false,
			want:    0,
		},
		{
			name: "epoch 4, subset 1, 0",
			args: args{
				stateDB:  sDB,
				epoch:    4,
				shardID:  0,
				subsetID: 1,
				tokenID:  common.PRVCoinID,
			},
			wantErr: false,
			want:    0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRewardOfShardByEpochV3(tt.args.stateDB, tt.args.epoch, tt.args.shardID, tt.args.subsetID, tt.args.tokenID)
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

func TestGetAllTokenIDForRewardV3(t *testing.T) {
	coin1 := common.HashH([]byte{1})
	sDB, _ := NewWithPrefixTrie(emptyRoot, wrarperDB)
	type addArgs struct {
		stateDB      *StateDB
		epoch        uint64
		shardID      byte
		subsetID     byte
		tokenID      common.Hash
		rewardAmount uint64
	}
	addArgss := []addArgs{
		addArgs{
			epoch:        1,
			shardID:      0,
			subsetID:     0,
			tokenID:      common.PRVCoinID,
			rewardAmount: 100,
		},
		addArgs{
			epoch:        1,
			shardID:      0,
			subsetID:     1,
			tokenID:      common.PRVCoinID,
			rewardAmount: 150,
		},
		addArgs{
			epoch:        2,
			shardID:      0,
			subsetID:     0,
			tokenID:      common.PRVCoinID,
			rewardAmount: 300,
		},
		addArgs{
			epoch:        2,
			shardID:      0,
			subsetID:     0,
			tokenID:      coin1,
			rewardAmount: 150,
		},
		addArgs{
			epoch:        2,
			shardID:      0,
			subsetID:     1,
			tokenID:      common.PRVCoinID,
			rewardAmount: 300,
		},
		addArgs{
			epoch:        2,
			shardID:      0,
			subsetID:     1,
			tokenID:      coin1,
			rewardAmount: 450,
		},
		addArgs{
			epoch:        3,
			shardID:      0,
			subsetID:     0,
			tokenID:      common.PRVCoinID,
			rewardAmount: 300,
		},
		addArgs{
			epoch:        3,
			shardID:      0,
			subsetID:     1,
			tokenID:      common.PRVCoinID,
			rewardAmount: 400,
		},
		addArgs{
			epoch:        3,
			shardID:      1,
			subsetID:     0,
			tokenID:      common.PRVCoinID,
			rewardAmount: 350,
		},
		addArgs{
			epoch:        3,
			shardID:      1,
			subsetID:     1,
			tokenID:      common.PRVCoinID,
			rewardAmount: 450,
		},
	}
	for _, add := range addArgss {
		if err := AddShardRewardRequestV3(sDB, add.epoch, add.shardID, add.subsetID, add.tokenID, add.rewardAmount); err != nil {
			log.Fatal(err)
		}
	}
	rootHash, _ := sDB.Commit(true)
	_ = sDB.Database().TrieDB().Commit(rootHash, true)

	type args struct {
		stateDB *StateDB
		epoch   uint64
	}
	tests := []struct {
		name string
		args args
		want []common.Hash
	}{
		{
			name: "get only prv",
			args: args{
				stateDB: sDB,
				epoch:   1,
			},
			want: []common.Hash{common.PRVCoinID},
		},
		{
			name: "get only prv and coin 1",
			args: args{
				stateDB: sDB,
				epoch:   2,
			},
			want: []common.Hash{common.PRVCoinID, coin1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAllTokenIDForRewardV3(tt.args.stateDB, tt.args.epoch)
			sort.Slice(got, func(i, j int) bool {
				return got[i].String() < got[j].String()
			})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllTokenIDForRewardV3() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestGetAllTokenIDForRewardV3_1(t *testing.T) {
	//diskDB, _ = incdb.Open("leveldb", "./data/beacon0")
	//wrarperDB = NewDatabaseAccessWarper(diskDB)
	//hash, _ := common.Hash{}.NewHashFromStr("79dcb9012102b7f2efd2c486bfb951b5b09a0f4a1648ef611477b0798717dc80")
	//t.Log(emptyRoot.String())
	//sDB, _ := NewWithPrefixTrie(*hash, wrarperDB)
	//got := GetAllTokenIDForRewardV3(sDB, 798)
	//t.Log(got)
}

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

	type args struct {
		stateDB *StateDB
		epoch   uint64
		shardID byte
		tokenID common.Hash
	}
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

func TestGetAllTokenIDForReward(t *testing.T) {
	sDB, _ := NewWithPrefixTrie(emptyRoot, wrarperDB)
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
			epoch:        1,
			shardID:      0,
			tokenID:      common.HashH([]byte{0}),
			rewardAmount: 100,
		},
		addArgs{
			epoch:        2,
			shardID:      0,
			tokenID:      common.HashH([]byte{0}),
			rewardAmount: 200,
		},
		addArgs{
			epoch:        2,
			shardID:      0,
			tokenID:      common.HashH([]byte{1}),
			rewardAmount: 150,
		},
		addArgs{
			epoch:        2,
			shardID:      0,
			tokenID:      common.HashH([]byte{2}),
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

	type args struct {
		stateDB *StateDB
		epoch   uint64
	}
	tests := []struct {
		name string
		args args
		want []common.Hash
	}{
		{
			name: "epoch 1",
			args: args{
				stateDB: sDB,
				epoch:   1,
			},
			want: []common.Hash{
				common.PRVCoinID,
				common.HashH([]byte{0}),
			},
		},
		{
			name: "epoch 1",
			args: args{
				stateDB: sDB,
				epoch:   2,
			},
			want: []common.Hash{
				common.HashH([]byte{0}),
				common.HashH([]byte{1}),
				common.HashH([]byte{2}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAllTokenIDForReward(tt.args.stateDB, tt.args.epoch); len(got) == len(tt.want) {
				sort.Slice(got, func(i, j int) bool {
					return string(got[i][:]) < string(got[j][:])
				})
				sort.Slice(tt.want, func(i, j int) bool {
					return string(tt.want[i][:]) < string(tt.want[j][:])
				})
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetAllTokenIDForReward() = %v, want %v", got, tt.want)
				}
			} else {
				t.Errorf("GetAllTokenIDForReward() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStateDB_AddCommitteeReward(t *testing.T) {
	stateDB, err := NewWithPrefixTrie(common.EmptyRoot, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	incognitoPublicKey := incognitoPublicKeys[0]
	amount := uint64(10000)
	withdraw := uint64(5000)
	tokenID := common.Hash{5}
	err = AddCommitteeReward(stateDB, incognitoPublicKey, amount, common.PRVCoinID)
	if err != nil {
		t.Fatal(err)
	}
	err = AddCommitteeReward(stateDB, incognitoPublicKey, amount, tokenID)
	if err != nil {
		t.Fatal(err)
	}
	rootHash, err := stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	err = stateDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		t.Fatal(err)
	}
	gotAmount0, err := GetCommitteeReward(stateDB, incognitoPublicKey, common.PRVCoinID)
	if err != nil {
		t.Fatal(err)
	}
	if gotAmount0 != amount {
		t.Fatalf("want %+v but got %+v", amount, gotAmount0)
	}
	err = AddCommitteeReward(stateDB, incognitoPublicKey, amount, common.PRVCoinID)
	if err != nil {
		t.Fatal(err)
	}
	rootHash, err = stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	err = stateDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		t.Fatal(err)
	}
	gotAmount1, err := GetCommitteeReward(stateDB, incognitoPublicKey, common.PRVCoinID)
	if err != nil {
		t.Fatal(err)
	}
	if gotAmount1 != amount*2 {
		t.Fatalf("want %+v but got %+v", amount*2, gotAmount1)
	}
	incognitoPublicKeyBytes, _, _ := base58.Base58Check{}.Decode(incognitoPublicKey)
	err = RemoveCommitteeReward(stateDB, incognitoPublicKeyBytes, withdraw, common.PRVCoinID)
	if err != nil {
		t.Fatal(err)
	}
	err = RemoveCommitteeReward(stateDB, incognitoPublicKeyBytes, withdraw, tokenID)
	if err != nil {
		t.Fatal(err)
	}
	rootHash, err = stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	err = stateDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		t.Fatal(err)
	}
	gotAmount2, err := GetCommitteeReward(stateDB, incognitoPublicKey, common.PRVCoinID)
	if err != nil {
		t.Fatal(err)
	}
	if gotAmount2 != amount*2-withdraw {
		t.Fatalf("want %+v but got %+v", amount*2-withdraw, gotAmount1)
	}
	gotAmount3, err := GetCommitteeReward(stateDB, incognitoPublicKey, tokenID)
	if err != nil {
		t.Fatal(err)
	}
	if gotAmount3 != amount-withdraw {
		t.Fatalf("want %+v but got %+v", amount-withdraw, gotAmount1)
	}
}
