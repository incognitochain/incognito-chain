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

func TestRejectUnstaking_FromStringSlice(t *testing.T) {
	rejectUnstaking := NewRejectUnstakingWithValue(common.PRVCoinID, 1, "", nil, nil)
	data, err := json.Marshal(&rejectUnstaking)
	assert.Nil(t, err)

	type fields struct {
		txReqID common.Hash
		shardID byte
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
					strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta),
					common.Pdexv3RejectStringStatus,
					string(data),
				},
			},
			fieldsAfterProcess: fields{
				shardID: 1,
				txReqID: common.PRVCoinID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RejectUnstaking{
				txReqID: tt.fields.txReqID,
				shardID: tt.fields.shardID,
			}
			if err := r.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("RejectUnstaking.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
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

func TestRejectUnstaking_StringSlice(t *testing.T) {
	rejectUnstaking := NewRejectUnstakingWithValue(common.PRVCoinID, 1, "", nil, nil)
	data, err := json.Marshal(&rejectUnstaking)
	assert.Nil(t, err)

	type fields struct {
		txReqID common.Hash
		shardID byte
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
				txReqID: common.PRVCoinID,
				shardID: 1,
			},
			want: []string{
				strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta),
				common.Pdexv3RejectStringStatus,
				string(data),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RejectUnstaking{
				txReqID: tt.fields.txReqID,
				shardID: tt.fields.shardID,
			}
			got, err := r.StringSlice()
			if (err != nil) != tt.wantErr {
				t.Errorf("RejectUnstaking.StringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RejectUnstaking.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
