package pdexv3

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

func TestWaitingAddLiquidity_FromStringSlice(t *testing.T) {
	metaData := metadataPdexv3.NewAddLiquidityWithValue(
		"pool_pair_id", "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		common.PRVIDStr, 300, 10000,
	)
	type fields struct {
		Base Base
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
			name: "Invalid Base",
			fields: fields{
				Base: Base{
					metaData: metadataPdexv3.NewAddLiquidity(),
				},
			},
			args: args{
				source: []string{
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
				source: append(metaData.StringSlice(), "tx_req_id", "1", RefundStatus),
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
				source: append(metaData.StringSlice(), "tx_req_id", "1", WaitingStatus),
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
			w := &WaitingAddLiquidity{
				Base: tt.fields.Base,
			}
			if err := w.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("WaitingAddLiquidity.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(w.metaData, tt.fieldsAfterProcess.Base.metaData) {
				t.Errorf("metaData got = %v, expected %v", w.metaData, tt.fieldsAfterProcess.Base.metaData)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(w.txReqID, tt.fieldsAfterProcess.Base.txReqID) {
				t.Errorf("txReqID got = %v, expected %v", w.txReqID, tt.fieldsAfterProcess.Base.txReqID)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(w.shardID, tt.fieldsAfterProcess.Base.shardID) {
				t.Errorf("shardID got = %v, expected %v", w.shardID, tt.fieldsAfterProcess.Base.shardID)
				return
			}
		})
	}
}

func TestWaitingAddLiquidity_StringSlice(t *testing.T) {
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
			want: append(metaData.StringSlice(), "tx_req_id", "1", WaitingStatus),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WaitingAddLiquidity{
				Base: tt.fields.Base,
			}
			if got := w.StringSlice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WaitingAddLiquidity.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
