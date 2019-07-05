package lvdb

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestDbAddShardRewardRequest(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		epoch        uint64
		shardID      byte
		rewardAmount uint64
		tokenID      common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.AddShardRewardRequest(tt.args.epoch, tt.args.shardID, tt.args.rewardAmount, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("db.AddShardRewardRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDbGetRewardOfShardByEpoch(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		epoch   uint64
		shardID byte
		tokenID common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
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
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		committeeAddress []byte
		amount           uint64
		tokenID          common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.AddCommitteeReward(tt.args.committeeAddress, tt.args.amount, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("db.AddCommitteeReward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDbGetCommitteeReward(t *testing.T) {
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		committeeAddress []byte
		tokenID          common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
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
	type fields struct {
		lvdb *leveldb.DB
	}
	type args struct {
		committeeAddress []byte
		amount           uint64
		tokenID          common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &db{
				lvdb: tt.fields.lvdb,
			}
			if err := db.RemoveCommitteeReward(tt.args.committeeAddress, tt.args.amount, tt.args.tokenID); (err != nil) != tt.wantErr {
				t.Errorf("db.RemoveCommitteeReward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
