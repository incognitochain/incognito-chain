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

func TestAcceptUserMintNft_FromStringSlice(t *testing.T) {
	txReqID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	data := AcceptUserMintNft{
		otaReceiver: validOTAReceiver0,
		nftID:       common.PRVCoinID,
		burntAmount: 100,
		shardID:     1,
		txReqID:     *txReqID,
	}
	dataBytes, err := json.Marshal(&data)
	assert.Nil(t, err)

	type fields struct {
		nftID       common.Hash
		burntAmount uint64
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
			name: "Invalid length of instructions",
			args: args{
				source: []string{},
			},
			wantErr: true,
		},
		{
			name: "Invalid metaType",
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
			name: "Invalid status",
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.Pdexv3UserMintNftRequestMeta),
					common.Pdexv3RejectStringStatus,
					string(dataBytes),
				},
			},
			wantErr: true,
		},
		{
			name: "Valid input",
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.Pdexv3UserMintNftRequestMeta),
					common.Pdexv3AcceptStringStatus,
					string(dataBytes),
				},
			},
			fieldsAfterProcess: fields{
				nftID:       common.PRVCoinID,
				burntAmount: 100,
				otaReceiver: validOTAReceiver0,
				shardID:     1,
				txReqID:     *txReqID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AcceptUserMintNft{
				nftID:       tt.fields.nftID,
				burntAmount: tt.fields.burntAmount,
				otaReceiver: tt.fields.otaReceiver,
				shardID:     tt.fields.shardID,
				txReqID:     tt.fields.txReqID,
			}
			if err := a.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("AcceptUserMintNft.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(a.burntAmount, tt.fieldsAfterProcess.burntAmount) {
				t.Errorf("burntAmount got = %v, want %v", a.burntAmount, tt.fieldsAfterProcess.burntAmount)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(a.nftID, tt.fieldsAfterProcess.nftID) {
				t.Errorf("burntAmount got = %v, want %v", a.nftID, tt.fieldsAfterProcess.nftID)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(a.otaReceiver, tt.fieldsAfterProcess.otaReceiver) {
				t.Errorf("otaReceive got = %v, want %v", a.otaReceiver, tt.fieldsAfterProcess.otaReceiver)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(a.shardID, tt.fieldsAfterProcess.shardID) {
				t.Errorf("shardID got = %v, want %v", a.shardID, tt.fieldsAfterProcess.shardID)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(a.txReqID, tt.fieldsAfterProcess.txReqID) {
				t.Errorf("txReqID got = %v, want %v", a.txReqID, tt.fieldsAfterProcess.txReqID)
				return
			}
		})
	}
}

func TestAcceptUserMintNft_StringSlice(t *testing.T) {
	txReqID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	data := AcceptUserMintNft{
		otaReceiver: validOTAReceiver0,
		nftID:       common.PRVCoinID,
		burntAmount: 100,
		shardID:     1,
		txReqID:     *txReqID,
	}
	dataBytes, err := json.Marshal(&data)
	assert.Nil(t, err)

	type fields struct {
		nftID       common.Hash
		burntAmount uint64
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
			name: "Valid input",
			fields: fields{
				nftID:       common.PRVCoinID,
				otaReceiver: validOTAReceiver0,
				burntAmount: 100,
				shardID:     1,
				txReqID:     *txReqID,
			},
			want: []string{
				strconv.Itoa(metadataCommon.Pdexv3UserMintNftRequestMeta),
				common.Pdexv3AcceptStringStatus,
				string(dataBytes),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AcceptUserMintNft{
				nftID:       tt.fields.nftID,
				burntAmount: tt.fields.burntAmount,
				otaReceiver: tt.fields.otaReceiver,
				shardID:     tt.fields.shardID,
				txReqID:     tt.fields.txReqID,
			}
			got, err := a.StringSlice()
			if (err != nil) != tt.wantErr {
				t.Errorf("AcceptUserMintNft.StringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AcceptUserMintNft.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
