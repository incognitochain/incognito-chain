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

func TestRejectWithdrawLiquidity_FromStringSlice(t *testing.T) {
	inst := RejectWithdrawLiquidity{
		txReqID: common.PRVCoinID,
		shardID: 1,
	}
	data, err := json.Marshal(&inst)
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
			name:   "Valid Input",
			fields: fields{},
			fieldsAfterProcess: fields{
				txReqID: common.PRVCoinID,
				shardID: 1,
			},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.Pdexv3WithdrawLiquidityRequestMeta),
					common.PDEWithdrawalRejectedChainStatus,
					string(data),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RejectWithdrawLiquidity{
				txReqID: tt.fields.txReqID,
				shardID: tt.fields.shardID,
			}
			if err := r.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("RejectWithdrawLiquidity.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(inst, *r) {
				t.Errorf("fieldsAfterProcess got = %v, expect %v", r, inst)
			}
		})
	}
}

func TestRejectWithdrawLiquidity_StringSlice(t *testing.T) {
	inst := RejectWithdrawLiquidity{
		txReqID: common.PRVCoinID,
		shardID: 1,
	}
	data, err := json.Marshal(&inst)
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
				strconv.Itoa(metadataCommon.Pdexv3WithdrawLiquidityRequestMeta),
				common.PDEWithdrawalRejectedChainStatus,
				string(data),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RejectWithdrawLiquidity{
				txReqID: tt.fields.txReqID,
				shardID: tt.fields.shardID,
			}
			got, err := r.StringSlice()
			if (err != nil) != tt.wantErr {
				t.Errorf("RejectWithdrawLiquidity.StringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RejectWithdrawLiquidity.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
