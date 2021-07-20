package pdexv3

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

func TestWaitingAddLiquidity_FromStringArr(t *testing.T) {
	type fields struct {
		poolPairID      string
		pairHash        string
		receiverAddress string
		refundAddress   string
		tokenID         string
		tokenAmount     uint64
		amplifier       uint
		txReqID         string
		shardID         byte
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
			name:   "Invalid length",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddOrderRequestMeta),
					WaitingStatus,
					common.PRVCoinID.String(),
					"300",
					"10000",
					validOTAReceiver0,
					validOTAReceiver1,
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid metadata type",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddOrderRequestMeta),
					WaitingStatus,
					"pool_pair_id",
					"pair_hash",
					common.PRVCoinID.String(),
					"300",
					"10000",
					validOTAReceiver0,
					validOTAReceiver1,
					"tx_req_id",
					"1",
				},
			},
			wantErr: true,
		},
		{
			name:   "invalid status",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					"",
					"pool_pair_id",
					"pair_hash",
					common.PRVCoinID.String(),
					"300",
					"10000",
					validOTAReceiver0,
					validOTAReceiver1,
					"tx_req_id",
					"1",
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid pair hash",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					WaitingStatus,
					"pool_pair_id",
					"",
					common.PRVCoinID.String(),
					"300",
					"10000",
					validOTAReceiver0,
					validOTAReceiver1,
					"tx_req_id",
					"1",
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid tokenID",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					WaitingStatus,
					"pool_pair_id",
					"pair_hash",
					"vzxvc",
					"300",
					"10000",
					validOTAReceiver0,
					validOTAReceiver1,
					"tx_req_id",
					"1",
				},
			},
			wantErr: true,
		},
		{
			name:   "Empty tokenID",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					WaitingStatus,
					"pool_pair_id",
					"pair_hash",
					common.Hash{}.String(),
					"300",
					"10000",
					validOTAReceiver0,
					validOTAReceiver1,
					"tx_req_id",
					"1",
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid token amount",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					WaitingStatus,
					"pool_pair_id",
					"pair_hash",
					common.PRVCoinID.String(),
					"token_amount",
					"10000",
					validOTAReceiver0,
					validOTAReceiver1,
					"tx_req_id",
					"1",
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid amplifier",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					WaitingStatus,
					"pool_pair_id",
					"pair_hash",
					common.PRVCoinID.String(),
					"300",
					"amplifier",
					validOTAReceiver0,
					validOTAReceiver1,
					"tx_req_id",
					"1",
				},
			},
			wantErr: true,
		},
		{
			name:   "Amplifier is smaller than default amplifier",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					WaitingStatus,
					"pool_pair_id",
					"pair_hash",
					common.PRVCoinID.String(),
					"300",
					"900",
					validOTAReceiver0,
					validOTAReceiver1,
					"tx_req_id",
					"1",
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid receiver Address",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					WaitingStatus,
					"pool_pair_id",
					"pair_hash",
					common.PRVCoinID.String(),
					"300",
					"10000",
					"receiver_address",
					validOTAReceiver1,
					"tx_req_id",
					"1",
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid refund Address",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					WaitingStatus,
					"pool_pair_id",
					"pair_hash",
					common.PRVCoinID.String(),
					"300",
					"10000",
					validOTAReceiver0,
					"refund_address",
					"tx_req_id",
					"1",
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid shardID",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					WaitingStatus,
					"pool_pair_id",
					"pair_hash",
					common.PRVCoinID.String(),
					"300",
					"10000",
					validOTAReceiver0,
					validOTAReceiver1,
					"tx_req_id",
					"shard_id",
				},
			},
			wantErr: true,
		},
		{
			name:   "Valid input",
			fields: fields{},
			fieldsAfterProcess: fields{
				poolPairID:      "pool_pair_id",
				pairHash:        "pair_hash",
				receiverAddress: validOTAReceiver0,
				refundAddress:   validOTAReceiver1,
				tokenID:         common.PRVCoinID.String(),
				tokenAmount:     300,
				amplifier:       10000,
				txReqID:         "tx_req_id",
				shardID:         1,
			},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					WaitingStatus,
					"pool_pair_id",
					"pair_hash",
					common.PRVCoinID.String(),
					"300",
					"10000",
					validOTAReceiver0,
					validOTAReceiver1,
					"tx_req_id",
					"1",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WaitingAddLiquidity{
				poolPairID:      tt.fields.poolPairID,
				pairHash:        tt.fields.pairHash,
				receiverAddress: tt.fields.receiverAddress,
				refundAddress:   tt.fields.refundAddress,
				tokenID:         tt.fields.tokenID,
				tokenAmount:     tt.fields.tokenAmount,
				amplifier:       tt.fields.amplifier,
				txReqID:         tt.fields.txReqID,
				shardID:         tt.fields.shardID,
			}
			if err := w.FromStringArr(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("WaitingAddLiquidity.FromStringArr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(w.poolPairID, tt.fieldsAfterProcess.poolPairID) {
					t.Errorf("poolPairID got = %v, want %v", w.poolPairID, tt.fieldsAfterProcess.poolPairID)
					return
				}
				if !reflect.DeepEqual(w.pairHash, tt.fieldsAfterProcess.pairHash) {
					t.Errorf("pairHash got = %v, want %v", w.pairHash, tt.fieldsAfterProcess.pairHash)
					return
				}
				if !reflect.DeepEqual(w.receiverAddress, tt.fieldsAfterProcess.receiverAddress) {
					t.Errorf("receiverAddress got = %v, want %v", w.receiverAddress, tt.fieldsAfterProcess.receiverAddress)
					return
				}
				if !reflect.DeepEqual(w.refundAddress, tt.fieldsAfterProcess.refundAddress) {
					t.Errorf("refundAddress got = %v, want %v", w.refundAddress, tt.fieldsAfterProcess.refundAddress)
					return
				}
				if !reflect.DeepEqual(w.tokenID, tt.fieldsAfterProcess.tokenID) {
					t.Errorf("tokenID got = %v, want %v", w.tokenID, tt.fieldsAfterProcess.tokenID)
					return
				}
				if !reflect.DeepEqual(w.tokenAmount, tt.fieldsAfterProcess.tokenAmount) {
					t.Errorf("tokenAmount got = %v, want %v", w.tokenAmount, tt.fieldsAfterProcess.tokenAmount)
					return
				}
				if !reflect.DeepEqual(w.amplifier, tt.fieldsAfterProcess.amplifier) {
					t.Errorf("amplifier got = %v, want %v", w.amplifier, tt.fieldsAfterProcess.amplifier)
					return
				}
				if !reflect.DeepEqual(w.txReqID, tt.fieldsAfterProcess.txReqID) {
					t.Errorf("txReqID got = %v, want %v", w.txReqID, tt.fieldsAfterProcess.txReqID)
					return
				}
				if !reflect.DeepEqual(w.shardID, tt.fieldsAfterProcess.shardID) {
					t.Errorf("shardID got = %v, want %v", w.shardID, tt.fieldsAfterProcess.shardID)
					return
				}
			}
		})
	}
}

func TestWaitingAddLiquidity_StringArr(t *testing.T) {
	type fields struct {
		poolPairID      string
		pairHash        string
		receiverAddress string
		refundAddress   string
		tokenID         string
		tokenAmount     uint64
		amplifier       uint
		txReqID         string
		shardID         byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "Valid Input",
			fields: fields{
				poolPairID:      "pool_pair_id",
				pairHash:        "pair_hash",
				receiverAddress: validOTAReceiver0,
				refundAddress:   validOTAReceiver1,
				tokenID:         common.PRVCoinID.String(),
				tokenAmount:     300,
				amplifier:       10000,
				txReqID:         "tx_req_id",
				shardID:         1,
			},
			want: []string{
				strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
				WaitingStatus,
				"pool_pair_id",
				"pair_hash",
				common.PRVCoinID.String(),
				"300",
				"10000",
				validOTAReceiver0,
				validOTAReceiver1,
				"tx_req_id",
				"1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WaitingAddLiquidity{
				poolPairID:      tt.fields.poolPairID,
				pairHash:        tt.fields.pairHash,
				receiverAddress: tt.fields.receiverAddress,
				refundAddress:   tt.fields.refundAddress,
				tokenID:         tt.fields.tokenID,
				tokenAmount:     tt.fields.tokenAmount,
				amplifier:       tt.fields.amplifier,
				txReqID:         tt.fields.txReqID,
				shardID:         tt.fields.shardID,
			}
			if got := w.StringArr(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WaitingAddLiquidity.StringArr() = %v, want %v", got, tt.want)
			}
		})
	}
}
