package lvdb

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
)

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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewKeyAddShardRewardRequest(tt.args.epoch, tt.args.shardID, tt.args.tokenID)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKeyAddShardRewardRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewKeyAddCommitteeReward(tt.args.committeeAddress, tt.args.tokenID)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKeyAddCommitteeReward() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKeyAddCommitteeReward() = %v, want %v", got, tt.want)
			}
		})
	}
}
