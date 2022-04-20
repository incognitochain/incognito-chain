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

func TestRejectUserMintNft_FromStringSlice(t *testing.T) {
	txReqID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	data := RejectUserMintNft{
		otaReceiver: validOTAReceiver0,
		amount:      100,
		shardID:     1,
		txReqID:     *txReqID,
	}
	dataBytes, err := json.Marshal(&data)
	assert.Nil(t, err)

	type fields struct {
		otaReceiver string
		amount      uint64
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
			name:   "Length of instructions != 2",
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
					"",
					"",
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid status",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.Pdexv3UserMintNftRequestMeta),
					common.Pdexv3AcceptStringStatus,
					string(dataBytes),
				},
			},
			wantErr: true,
		},
		{
			name:   "Valid input",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.Pdexv3UserMintNftRequestMeta),
					common.Pdexv3RejectStringStatus,
					string(dataBytes),
				},
			},
			fieldsAfterProcess: fields{
				otaReceiver: validOTAReceiver0,
				amount:      100,
				shardID:     1,
				txReqID:     *txReqID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RejectUserMintNft{
				otaReceiver: tt.fields.otaReceiver,
				amount:      tt.fields.amount,
				shardID:     tt.fields.shardID,
				txReqID:     tt.fields.txReqID,
			}
			if err := r.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("RejectUserMintNft.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.fieldsAfterProcess.otaReceiver, r.otaReceiver) {
				t.Errorf("otaReceive got = %v, want %v", r.otaReceiver, tt.fieldsAfterProcess.otaReceiver)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.fieldsAfterProcess.amount, r.amount) {
				t.Errorf("amount got = %v, want %v", r.amount, tt.fieldsAfterProcess.amount)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.fieldsAfterProcess.shardID, r.shardID) {
				t.Errorf("shardID got = %v, want %v", r.shardID, tt.fieldsAfterProcess.shardID)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.fieldsAfterProcess.txReqID, r.txReqID) {
				t.Errorf("txReqID got = %v, want %v", r.txReqID, tt.fieldsAfterProcess.txReqID)
			}
		})
	}
}

func TestRejectUserMintNft_StringSlice(t *testing.T) {
	txReqID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	data := RejectUserMintNft{
		otaReceiver: validOTAReceiver0,
		amount:      100,
		shardID:     1,
		txReqID:     *txReqID,
	}
	dataBytes, err := json.Marshal(&data)
	assert.Nil(t, err)

	type fields struct {
		otaReceiver string
		amount      uint64
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
			name: "Valid input",
			fields: fields{
				otaReceiver: validOTAReceiver0,
				amount:      100,
				shardID:     1,
				txReqID:     *txReqID,
			},
			want: []string{
				strconv.Itoa(metadataCommon.Pdexv3UserMintNftRequestMeta),
				common.Pdexv3RejectStringStatus,
				string(dataBytes),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RejectUserMintNft{
				otaReceiver: tt.fields.otaReceiver,
				amount:      tt.fields.amount,
				shardID:     tt.fields.shardID,
				txReqID:     tt.fields.txReqID,
			}
			got, err := r.StringSlice()
			if (err != nil) != tt.wantErr {
				t.Errorf("RejectUserMintNft.StringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RejectUserMintNft.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
