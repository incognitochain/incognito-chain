package pdexv3

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

func TestMatchAndReturnAddLiquidity_FromStringSlice(t *testing.T) {
	metaData := metadataPdexv3.NewAddLiquidityWithValue(
		"pool_pair_id", "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		common.PRVIDStr, 300, 10000,
	)
	type fields struct {
		Base         Base
		returnAmount uint64
		nfctID       string
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
			name: "Length of source < 4",
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
					"",
					"",
				},
			},
			wantErr: true,
		},
		{
			name: "Pool pair id is empty",
			fields: fields{
				Base: Base{
					metaData: metadataPdexv3.NewAddLiquidity(),
				},
			},
			args: args{
				source: append(metaData.StringSlice(),
					"tx_req_id", "1",
					"", common.PRVCoinID.String(),
					RefundStatus),
			},
			wantErr: true,
		},
		{
			name: "Invalid Token ID",
			fields: fields{
				Base: Base{
					metaData: metadataPdexv3.NewAddLiquidity(),
				},
			},
			args: args{
				source: append(metaData.StringSlice(),
					"tx_req_id", "1",
					"new_pool_pair_id", "basv",
					RefundStatus),
			},
			wantErr: true,
		},
		{
			name: "Empty Token ID",
			fields: fields{
				Base: Base{
					metaData: metadataPdexv3.NewAddLiquidity(),
				},
			},
			args: args{
				source: append(metaData.StringSlice(),
					"tx_req_id", "1",
					"new_pool_pair_id", common.Hash{}.String(),
					RefundStatus),
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
				source: append(metaData.StringSlice(),
					"tx_req_id", "1",
					"new_pool_pair_id", common.PRVCoinID.String(),
					RefundStatus),
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
				source: append(metaData.StringSlice(),
					"tx_req_id", "1",
					"100", common.PRVCoinID.String(),
					MatchAndReturnStatus,
				),
			},
			fieldsAfterProcess: fields{
				Base: Base{
					metaData: metaData,
					txReqID:  "tx_req_id",
					shardID:  1,
				},
				returnAmount: 100,
				nfctID:       common.PRVCoinID.String(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MatchAndReturnAddLiquidity{
				Base:         tt.fields.Base,
				returnAmount: tt.fields.returnAmount,
				nfctID:       tt.fields.nfctID,
			}
			if err := m.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("MatchAndReturnAddLiquidity.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(m.metaData, tt.fieldsAfterProcess.Base.metaData) {
				t.Errorf("metaData got = %v, expected = %v", m.metaData, tt.fieldsAfterProcess.Base.metaData)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(m.txReqID, tt.fieldsAfterProcess.Base.txReqID) {
				t.Errorf("txReqID got = %v, expected = %v", m.txReqID, tt.fieldsAfterProcess.Base.txReqID)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(m.shardID, tt.fieldsAfterProcess.Base.shardID) {
				t.Errorf("shardID got = %v, expected = %v", m.shardID, tt.fieldsAfterProcess.Base.shardID)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(m.returnAmount, tt.fieldsAfterProcess.returnAmount) {
				t.Errorf("newPoolPairID got = %v, expected = %v", m.returnAmount, tt.fieldsAfterProcess.returnAmount)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(m.nfctID, tt.fieldsAfterProcess.nfctID) {
				t.Errorf("nfctID got = %v, expected = %v", m.nfctID, tt.fieldsAfterProcess.nfctID)
				return
			}
		})
	}
}

func TestMatchAndReturnAddLiquidity_StringSlice(t *testing.T) {
	metaData := metadataPdexv3.NewAddLiquidityWithValue(
		"pool_pair_id", "pair_hash",
		validOTAReceiver0, validOTAReceiver1,
		common.PRVIDStr, 300, 10000,
	)
	type fields struct {
		Base         Base
		returnAmount uint64
		nfctID       string
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
				returnAmount: 100,
				nfctID:       "nfct_id",
			},
			want: append(metaData.StringSlice(),
				"tx_req_id", "1", "100", "nfct_id", MatchAndReturnStatus),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MatchAndReturnAddLiquidity{
				Base:         tt.fields.Base,
				returnAmount: tt.fields.returnAmount,
				nfctID:       tt.fields.nfctID,
			}
			if got := m.StringSlice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MatchAndReturnAddLiquidity.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
