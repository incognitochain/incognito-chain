package pdexv3

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

func TestRefundAddLiquidity_FromStringSlice(t *testing.T) {
	metaData := metadataPdexv3.NewAddLiquidityWithValue(
		"pool_pair_id", "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		common.PRVIDStr, 300, 10000,
	)
	type fields struct {
		Base               Base
		existedTokenID     string
		existedTokenAmount uint64
		refundAddress      string
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
			name: "Length of source < 2",
			fields: fields{
				Base: Base{
					metaData: metadataPdexv3.NewAddLiquidity(),
				},
			},
			args: args{
				source: []string{},
			},
			wantErr: true,
		},
		{
			name: "Invalid Base Instruction",
			fields: fields{
				Base: Base{
					metaData: metadataPdexv3.NewAddLiquidity(),
				},
			},
			args: args{
				source: []string{
					"",
					"",
					"",
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid status",
			fields: fields{
				Base: Base{
					metaData: metadataPdexv3.NewAddLiquidity(),
				},
			},
			args: args{
				source: append(metaData.StringSlice(), "tx_req_id", "1", WaitingStatus),
			},
			wantErr: true,
		},
		{
			name: "Valid Input",
			fields: fields{
				Base: Base{
					metaData: metadataPdexv3.NewAddLiquidity(),
				},
			},
			args: args{
				source: append(metaData.StringSlice(), "tx_req_id", "1", RefundStatus),
			},
			fieldsAfterProcess: fields{
				Base: Base{
					metaData: metaData,
					txReqID:  "tx_req_id",
					shardID:  1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RefundAddLiquidity{
				Base: tt.fields.Base,
			}
			if err := r.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("RefundAddLiquidity.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(r.metaData, tt.fieldsAfterProcess.Base.metaData) {
				t.Errorf("metaData got = %v, expected %v", r.metaData, tt.fieldsAfterProcess.Base.metaData)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(r.txReqID, tt.fieldsAfterProcess.Base.txReqID) {
				t.Errorf("txReqID got = %v, expected %v", r.txReqID, tt.fieldsAfterProcess.Base.txReqID)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(r.shardID, tt.fieldsAfterProcess.Base.shardID) {
				t.Errorf("shardID got = %v, expected %v", r.shardID, tt.fieldsAfterProcess.Base.shardID)
				return
			}
		})
	}
}

func TestRefundAddLiquidity_StringSlice(t *testing.T) {
	metaData := metadataPdexv3.NewAddLiquidityWithValue(
		"pool_pair_id", "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		common.PRVIDStr, 300, 10000,
	)
	type fields struct {
		Base Base
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "Valid Input",
			fields: fields{
				Base: Base{
					metaData: metaData,
					txReqID:  "tx_req_id",
					shardID:  1,
				},
			},
			want: append(metaData.StringSlice(), "tx_req_id", "1", RefundStatus),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RefundAddLiquidity{
				Base: tt.fields.Base,
			}
			if got := r.StringSlice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RefundAddLiquidity.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
