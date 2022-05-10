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

func TestRejectStaking_FromStringSlice(t *testing.T) {
	rejectStaking := NewRejectStakingWithValue(
		validOTAReceiver0, common.PRVCoinID, common.PRVCoinID, 1, 100,
	)
	data, err := json.Marshal(&rejectStaking)
	assert.Nil(t, err)

	type fields struct {
		tokenID     common.Hash
		amount      uint64
		otaReceiver string
		shardID     byte
		txReqID     common.Hash
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
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					common.Pdexv3RejectStringStatus,
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
					common.Pdexv3RejectStringStatus,
					string(data),
				},
			},
			fieldsAfterProcess: fields{
				tokenID:     common.PRVCoinID,
				amount:      100,
				otaReceiver: validOTAReceiver0,
				shardID:     1,
				txReqID:     common.PRVCoinID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RejectStaking{
				tokenID:     tt.fields.tokenID,
				amount:      tt.fields.amount,
				otaReceiver: tt.fields.otaReceiver,
				shardID:     tt.fields.shardID,
				txReqID:     tt.fields.txReqID,
			}
			if err := r.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("RejectStaking.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(tt.fieldsAfterProcess.tokenID, r.tokenID) {
					t.Errorf("tokenID got = %v, want %v", r.tokenID, tt.fieldsAfterProcess.tokenID)
					return
				}
				if !reflect.DeepEqual(tt.fieldsAfterProcess.amount, r.amount) {
					t.Errorf("amount got = %v, want %v", r.amount, tt.fieldsAfterProcess.amount)
					return
				}
				if !reflect.DeepEqual(tt.fieldsAfterProcess.otaReceiver, r.otaReceiver) {
					t.Errorf("otaReceiver got = %v, want %v", r.otaReceiver, tt.fieldsAfterProcess.otaReceiver)
					return
				}
				if !reflect.DeepEqual(tt.fieldsAfterProcess.shardID, r.shardID) {
					t.Errorf("shardID got = %v, want %v", r.shardID, tt.fieldsAfterProcess.shardID)
					return
				}
				if !reflect.DeepEqual(tt.fieldsAfterProcess.txReqID, r.txReqID) {
					t.Errorf("txReqID got = %v, want %v", r.txReqID, tt.fieldsAfterProcess.txReqID)
					return
				}
			}
		})
	}
}

func TestRejectStaking_StringSlice(t *testing.T) {
	rejectStaking := NewRejectStakingWithValue(
		validOTAReceiver0, common.PRVCoinID, common.PRVCoinID, 1, 100,
	)
	data, err := json.Marshal(&rejectStaking)
	assert.Nil(t, err)
	type fields struct {
		tokenID     common.Hash
		amount      uint64
		otaReceiver string
		shardID     byte
		txReqID     common.Hash
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
				tokenID:     common.PRVCoinID,
				amount:      100,
				otaReceiver: validOTAReceiver0,
				shardID:     1,
				txReqID:     common.PRVCoinID,
			},
			want: []string{
				strconv.Itoa(metadataCommon.Pdexv3StakingRequestMeta),
				common.Pdexv3RejectStringStatus,
				string(data),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RejectStaking{
				tokenID:     tt.fields.tokenID,
				amount:      tt.fields.amount,
				otaReceiver: tt.fields.otaReceiver,
				shardID:     tt.fields.shardID,
				txReqID:     tt.fields.txReqID,
			}
			got, err := r.StringSlice()
			if (err != nil) != tt.wantErr {
				t.Errorf("RejectStaking.StringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RejectStaking.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
