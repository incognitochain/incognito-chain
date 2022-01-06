package pdexv3

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/stretchr/testify/assert"
)

func TestAcceptUnstaking_FromStringSlice(t *testing.T) {
	acceptUnstakingInst := NewAcceptUnstakingWithValue(
		common.PRVCoinID, common.PRVCoinID, 50, validOTAReceiver0,
		common.PRVCoinID, 1,
	)
	data, err := json.Marshal(&acceptUnstakingInst)
	assert.Nil(t, err)

	type fields struct {
		stakingPoolID common.Hash
		nftID         common.Hash
		amount        uint64
		otaReceiver   string
		txReqID       common.Hash
		shardID       byte
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
					common.Pdexv3AcceptUnstakingStatus,
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
					strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta),
					common.Pdexv3AcceptUnstakingStatus,
					string(data),
				},
			},
			wantErr: false,
			fieldsAfterProcess: fields{
				nftID:         common.PRVCoinID,
				stakingPoolID: common.PRVCoinID,
				amount:        50,
				otaReceiver:   validOTAReceiver0,
				txReqID:       common.PRVCoinID,
				shardID:       1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AcceptUnstaking{
				stakingPoolID: tt.fields.stakingPoolID,
				nftID:         tt.fields.nftID,
				amount:        tt.fields.amount,
				otaReceiver:   tt.fields.otaReceiver,
				txReqID:       tt.fields.txReqID,
				shardID:       tt.fields.shardID,
			}
			if err := a.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("AcceptUnstaking.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(tt.fieldsAfterProcess.nftID, a.nftID) {
					t.Errorf("nftID got = %v, err %v", a.nftID, tt.fieldsAfterProcess.nftID)
					return
				}
				if !reflect.DeepEqual(tt.fieldsAfterProcess.stakingPoolID, a.stakingPoolID) {
					t.Errorf("stakingPoolID got = %v, err %v", a.stakingPoolID, tt.fieldsAfterProcess.stakingPoolID)
					return
				}
				if !reflect.DeepEqual(tt.fieldsAfterProcess.amount, a.amount) {
					t.Errorf("amount got = %v, err %v", a.amount, tt.fieldsAfterProcess.amount)
					return
				}
				if !reflect.DeepEqual(tt.fieldsAfterProcess.otaReceiver, a.otaReceiver) {
					t.Errorf("otaReceiver got = %v, err %v", a.otaReceiver, tt.fieldsAfterProcess.otaReceiver)
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

func TestAcceptUnstaking_StringSlice(t *testing.T) {
	acceptUnstakingInst := NewAcceptUnstakingWithValue(
		common.PRVCoinID, common.PRVCoinID, 50, validOTAReceiver0,
		common.PRVCoinID, 1,
	)
	data, err := json.Marshal(&acceptUnstakingInst)
	assert.Nil(t, err)

	type fields struct {
		stakingPoolID common.Hash
		nftID         common.Hash
		amount        uint64
		otaReceiver   string
		txReqID       common.Hash
		shardID       byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "Valid input",
			fields: fields{
				stakingPoolID: common.PRVCoinID,
				nftID:         common.PRVCoinID,
				amount:        50,
				otaReceiver:   validOTAReceiver0,
				txReqID:       common.PRVCoinID,
				shardID:       1,
			},
			want: []string{
				strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta),
				common.Pdexv3AcceptUnstakingStatus,
				string(data),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AcceptUnstaking{
				stakingPoolID: tt.fields.stakingPoolID,
				nftID:         tt.fields.nftID,
				amount:        tt.fields.amount,
				otaReceiver:   tt.fields.otaReceiver,
				txReqID:       tt.fields.txReqID,
				shardID:       tt.fields.shardID,
			}
			got, err := a.StringSlice()
			if (err != nil) != tt.wantErr {
				t.Errorf("AcceptUnstaking.StringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AcceptUnstaking.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
