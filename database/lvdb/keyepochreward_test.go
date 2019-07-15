package lvdb

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
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
