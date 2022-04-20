package pdexv3

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/stretchr/testify/assert"
)

func TestUnstakingResponse_ValidateSanityData(t *testing.T) {
	txReqID, err := common.Hash{}.NewHashFromStr("123456")
	assert.Nil(t, err)
	type fields struct {
		MetadataBase metadataCommon.MetadataBase
		status       string
		txReqID      string
	}
	type args struct {
		chainRetriever      metadataCommon.ChainRetriever
		shardViewRetriever  metadataCommon.ShardViewRetriever
		beaconViewRetriever metadataCommon.BeaconViewRetriever
		beaconHeight        uint64
		tx                  metadataCommon.Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		want1   bool
		wantErr bool
	}{
		{
			name: "status != common.Pdexv3AcceptStringStatus",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingResponseMeta,
				},
				status:  common.Pdexv3RejectStringStatus,
				txReqID: txReqID.String(),
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "txReqID is invalid",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingResponseMeta,
				},
				status:  common.Pdexv3AcceptStringStatus,
				txReqID: "abc12312bdas",
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "txReqID is empty",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingResponseMeta,
				},
				status:  common.Pdexv3AcceptStringStatus,
				txReqID: common.Hash{}.String(),
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Valid input",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingResponseMeta,
				},
				status:  common.Pdexv3AcceptStringStatus,
				txReqID: txReqID.String(),
			},
			args:    args{},
			want:    true,
			want1:   true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &UnstakingResponse{
				MetadataBase: tt.fields.MetadataBase,
				status:       tt.fields.status,
				txReqID:      tt.fields.txReqID,
			}
			got, got1, err := response.ValidateSanityData(tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.beaconHeight, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnstakingResponse.ValidateSanityData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UnstakingResponse.ValidateSanityData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("UnstakingResponse.ValidateSanityData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestUnstakingResponse_ValidateMetadataByItself(t *testing.T) {
	type fields struct {
		MetadataBase metadataCommon.MetadataBase
		status       string
		txReqID      string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Invalid Input",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3AddOrderRequestMeta,
				},
			},
			want: false,
		},
		{
			name: "Valid Input",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingResponseMeta,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &UnstakingResponse{
				MetadataBase: tt.fields.MetadataBase,
				status:       tt.fields.status,
				txReqID:      tt.fields.txReqID,
			}
			if got := response.ValidateMetadataByItself(); got != tt.want {
				t.Errorf("UnstakingResponse.ValidateMetadataByItself() = %v, want %v", got, tt.want)
			}
		})
	}
}
