package lvdb_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/incognitochain/incognito-chain/privacy"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
)

func TestDbAddShardRewardRequest(t *testing.T) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %+v", err)
	}
	t.Log(dbPath)
	db, err := database.Open("leveldb", dbPath)
	type args struct {
		epoch        uint64
		shardID      byte
		rewardAmount uint64
		tokenID      common.Hash
	}
	tests := []struct {
		name string
		// fields  fields
		args    args
		wantErr bool
	}{
		{
			"Happy case: Write on new key",
			*&args{
				epoch:        10,
				shardID:      0x00,
				rewardAmount: 1000,
				tokenID:      common.PRVCoinID,
			},
			false,
		},
		{
			"Happy case: Write on existed key",
			*&args{
				epoch:        10,
				shardID:      0x00,
				rewardAmount: 1000,
				tokenID:      common.PRVCoinID,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// db := &db{
			// 	lvdb: tt.fields.lvdb,
			// }
			if err := db.AddShardRewardRequest(tt.args.epoch, tt.args.shardID, tt.args.rewardAmount, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("db.AddShardRewardRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDbGetRewardOfShardByEpoch(t *testing.T) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %+v", err)
	}
	t.Log(dbPath)
	db, err := database.Open("leveldb", dbPath)

	type args struct {
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
			"Happy case: Get value of non-existed key",
			*&args{
				epoch:   10,
				shardID: 0x00,
				tokenID: common.PRVCoinID,
			},
			0,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetRewardOfShardByEpoch(tt.args.epoch, tt.args.shardID, tt.args.tokenID)
			if (err != nil) != tt.wantErr {
				t.Errorf("db.GetRewardOfShardByEpoch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("db.GetRewardOfShardByEpoch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDbAddCommitteeReward(t *testing.T) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %+v", err)
	}
	t.Log(dbPath)
	db, err := database.Open("leveldb", dbPath)
	type args struct {
		committeeAddress []byte
		amount           uint64
		tokenID          common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Happy case: Write on new key",
			*&args{
				committeeAddress: privacy.GeneratePublicKey([]byte{0x01, 0x02, 0x03, 0x09}),
				amount:           1000,
				tokenID:          common.PRVCoinID,
			},
			false,
		},
		{
			"Happy case: Write on existed key",
			*&args{
				committeeAddress: privacy.GeneratePublicKey([]byte{0x01, 0x02, 0x03, 0x09}),
				amount:           100,
				tokenID:          common.PRVCoinID,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := db.AddCommitteeReward(tt.args.committeeAddress, tt.args.amount, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("db.AddCommitteeReward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDbGetCommitteeReward(t *testing.T) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %+v", err)
	}
	t.Log(dbPath)
	db, err := database.Open("leveldb", dbPath)
	type args struct {
		committeeAddress []byte
		tokenID          common.Hash
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			"Happy case",
			*&args{
				committeeAddress: privacy.GeneratePublicKey([]byte{0x01, 0x02, 0x03, 0x09}),
				tokenID:          common.PRVCoinID,
			},
			0,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetCommitteeReward(tt.args.committeeAddress, tt.args.tokenID)
			if (err != nil) != tt.wantErr {
				t.Errorf("db.GetCommitteeReward() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("db.GetCommitteeReward() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDbRemoveCommitteeReward(t *testing.T) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %+v", err)
	}
	t.Log(dbPath)
	db, err := database.Open("leveldb", dbPath)
	type args struct {
		committeeAddress []byte
		amount           uint64
		tokenID          common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Happy case",
			*&args{
				committeeAddress: privacy.GeneratePublicKey([]byte{0x01, 0x02, 0x03, 0x09}),
				amount:           100,
				tokenID:          common.PRVCoinID,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := db.RemoveCommitteeReward(tt.args.committeeAddress, tt.args.amount, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("db.RemoveCommitteeReward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
