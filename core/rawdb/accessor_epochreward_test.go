package rawdb

import (
	"github.com/incognitochain/incognito-chain/incdb"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/privacy"

	"github.com/incognitochain/incognito-chain/common"
)

func TestDbAddShardRewardRequest(t *testing.T) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %+v", err)
	}
	t.Log(dbPath)
	db, err := incdb.Open("leveldb", dbPath)
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
			if err := AddShardRewardRequest(db, tt.args.epoch, tt.args.shardID, tt.args.rewardAmount, tt.args.tokenID, &[]incdb.BatchData{}); (err != nil) != tt.wantErr {
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
	db, err := incdb.Open("leveldb", dbPath)

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
			got, err := GetRewardOfShardByEpoch(db, tt.args.epoch, tt.args.shardID, tt.args.tokenID)
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
	db, err := incdb.Open("leveldb", dbPath)
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
			if err := AddCommitteeReward(db, tt.args.committeeAddress, tt.args.amount, tt.args.tokenID); (err != nil) != tt.wantErr {
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
	db, err := incdb.Open("leveldb", dbPath)
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
			got, err := GetCommitteeReward(db, tt.args.committeeAddress, tt.args.tokenID)
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
	db, err := incdb.Open("leveldb", dbPath)
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
			if err := RemoveCommitteeReward(db, tt.args.committeeAddress, tt.args.amount, tt.args.tokenID, &[]incdb.BatchData{}); (err != nil) != tt.wantErr {
				t.Errorf("db.RemoveCommitteeReward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewKeyAddShardRewardRequest(t *testing.T) {
	type args struct {
		epoch   uint64
		shardID byte
		tokenID common.Hash
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"Happy case",
			*&args{
				epoch:   100,
				shardID: 0x03,
				tokenID: common.PRVCoinID,
			},
			[]byte{115, 104, 97, 114, 100, 114, 101, 113, 117, 101, 115, 116, 114, 101, 119, 97, 114, 100, 45, 100, 0, 0, 0, 0, 0, 0, 0, 3, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addShardRewardRequestKey(tt.args.epoch, tt.args.shardID, tt.args.tokenID)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKeyAddShardRewardRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewKeyAddCommitteeReward(t *testing.T) {
	type args struct {
		committeeAddress []byte
		tokenID          common.Hash
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"Happy case",
			*&args{
				committeeAddress: privacy.GeneratePublicKey([]byte{0x01, 0x02, 0x03, 0x09}),
				tokenID:          common.PRVCoinID,
			},
			[]byte{99, 111, 109, 109, 105, 116, 116, 101, 101, 45, 114, 101, 119, 97, 114, 100, 45, 3, 56, 13, 86, 127, 3, 39, 177, 103, 184, 120, 161, 45, 149, 151, 62, 170, 220, 161, 75, 142, 240, 5, 118, 84, 26, 240, 71, 179, 132, 127, 88, 63, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addCommitteeRewardKey(tt.args.committeeAddress, tt.args.tokenID)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKeyAddCommitteeReward() = %v, want %v", got, tt.want)
			}
		})
	}
}
