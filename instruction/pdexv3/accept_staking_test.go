package pdexv3

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/stretchr/testify/assert"
)

func TestAcceptStaking_FromStringSlice(t *testing.T) {
	initTestParam(t)
	acceptStakingInst := NewAcceptStakingWithValue(
		common.PRVCoinID, common.PRVCoinID, 1, 100, accessOTA.ToBytesS(),
		metadataPdexv3.AccessOption{
			NftID: &common.PRVCoinID,
		},
	)
	data, err := json.Marshal(&acceptStakingInst)
	assert.Nil(t, err)

	type fields struct {
		metadataPdexv3.AccessOption
		stakingPoolID common.Hash
		liquidity     uint64
		accessOTA     []byte
		shardID       byte
		txReqID       common.Hash
	}
	type args struct {
		source []string
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		wantErr            bool
	}{
		{
			name:   "Length of instruction != 3",
			fields: fields{},
			args: args{
				source: []string{},
			},
			wantErr: true,
		},
		{
			name:   "Invalid metaType",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityResponseMeta),
					common.Pdexv3AcceptStringStatus,
					string(data),
				},
			},
			wantErr: true,
		},
		{
			name:   "Valid input",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.Pdexv3StakingRequestMeta),
					common.Pdexv3AcceptStringStatus,
					string(data),
				},
			},
			wantErr: false,
			fieldsAfterProcess: fields{
				AccessOption: metadataPdexv3.AccessOption{
					NftID: &common.PRVCoinID,
				},
				accessOTA:     accessOTA.ToBytesS(),
				stakingPoolID: common.PRVCoinID,
				txReqID:       common.PRVCoinID,
				liquidity:     100,
				shardID:       1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AcceptStaking{
				AccessOption:  tt.fields.AccessOption,
				stakingPoolID: tt.fields.stakingPoolID,
				liquidity:     tt.fields.liquidity,
				accessOTA:     tt.fields.accessOTA,
				shardID:       tt.fields.shardID,
				txReqID:       tt.fields.txReqID,
			}
			if err := a.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("AcceptStaking.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(tt.fieldsAfterProcess.AccessOption, a.AccessOption) {
					t.Errorf("accessOption got = %v, err %v", a.AccessOption, tt.fieldsAfterProcess.AccessOption)
					return
				}
				if !reflect.DeepEqual(tt.fieldsAfterProcess.stakingPoolID, a.stakingPoolID) {
					t.Errorf("stakingPoolID got = %v, err %v", a.stakingPoolID, tt.fieldsAfterProcess.stakingPoolID)
					return
				}
				if !reflect.DeepEqual(tt.fieldsAfterProcess.liquidity, a.liquidity) {
					t.Errorf("liquidity got = %v, err %v", a.liquidity, tt.fieldsAfterProcess.liquidity)
					return
				}
				if !reflect.DeepEqual(tt.fieldsAfterProcess.shardID, a.shardID) {
					t.Errorf("shardID got = %v, err %v", a.shardID, tt.fieldsAfterProcess.shardID)
					return
				}
				if !reflect.DeepEqual(tt.fieldsAfterProcess.txReqID, a.txReqID) {
					t.Errorf("txReqID got = %v, err %v", a.txReqID, tt.fieldsAfterProcess.txReqID)
					return
				}
			}
		})
	}
}

func TestAcceptStaking_StringSlice(t *testing.T) {
	initTestParam(t)
	acceptStakingInst := NewAcceptStakingWithValue(
		common.PRVCoinID, common.PRVCoinID, 1, 100, accessOTA.ToBytesS(),
		metadataPdexv3.AccessOption{
			NftID: &common.PRVCoinID,
		},
	)
	data, err := json.Marshal(&acceptStakingInst)
	assert.Nil(t, err)
	type fields struct {
		AccessOption  metadataPdexv3.AccessOption
		stakingPoolID common.Hash
		liquidity     uint64
		accessOTA     []byte
		shardID       byte
		txReqID       common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				AccessOption: metadataPdexv3.AccessOption{
					NftID: &common.PRVCoinID,
				},
				stakingPoolID: common.PRVCoinID,
				liquidity:     100,
				accessOTA:     accessOTA.ToBytesS(),
				shardID:       1,
				txReqID:       common.PRVCoinID,
			},
			want: []string{
				strconv.Itoa(metadataCommon.Pdexv3StakingRequestMeta),
				common.Pdexv3AcceptStringStatus,
				string(data),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AcceptStaking{
				AccessOption:  tt.fields.AccessOption,
				stakingPoolID: tt.fields.stakingPoolID,
				liquidity:     tt.fields.liquidity,
				accessOTA:     tt.fields.accessOTA,
				shardID:       tt.fields.shardID,
				txReqID:       tt.fields.txReqID,
			}
			got, err := a.StringSlice()
			if (err != nil) != tt.wantErr {
				t.Errorf("AcceptStaking.StringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AcceptStaking.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
