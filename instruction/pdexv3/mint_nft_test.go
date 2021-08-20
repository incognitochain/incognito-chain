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

func TestMintNft_FromStringSlice(t *testing.T) {
	mintNftTest := MintNft{
		nftID:       common.PRVCoinID,
		otaReceiver: validOTAReceiver0,
		shardID:     1,
	}
	data, err := json.Marshal(&mintNftTest)
	assert.Nil(t, err)

	type fields struct {
		nftID       common.Hash
		otaReceiver string
		shardID     byte
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
				nftID:       common.PRVCoinID,
				otaReceiver: validOTAReceiver0,
				shardID:     1,
			},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.Pdexv3MintNftRequestMeta),
					strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
					string(data),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MintNft{
				nftID:       tt.fields.nftID,
				otaReceiver: tt.fields.otaReceiver,
				shardID:     tt.fields.shardID,
			}
			if err := m.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("MintNft.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(mintNftTest, *m) {
				t.Errorf("fieldsAfterProcess got = %v, expect %v", *m, mintNftTest)
			}
		})
	}
}

func TestMintNft_StringSlice(t *testing.T) {
	mintNftTest := MintNft{
		nftID:       common.PRVCoinID,
		otaReceiver: validOTAReceiver0,
		shardID:     1,
	}
	data, err := json.Marshal(&mintNftTest)
	assert.Nil(t, err)
	type fields struct {
		nftID       common.Hash
		otaReceiver string
		shardID     byte
	}
	type args struct {
		action string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Valid Input",
			fields: fields{
				nftID:       common.PRVCoinID,
				otaReceiver: validOTAReceiver0,
				shardID:     1,
			},
			args: args{
				action: strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
			},
			want: []string{
				strconv.Itoa(metadataCommon.Pdexv3MintNftRequestMeta),
				strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta),
				string(data),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MintNft{
				nftID:       tt.fields.nftID,
				otaReceiver: tt.fields.otaReceiver,
				shardID:     tt.fields.shardID,
			}
			got, err := m.StringSlice(tt.args.action)
			if (err != nil) != tt.wantErr {
				t.Errorf("MintNft.StringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MintNft.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
