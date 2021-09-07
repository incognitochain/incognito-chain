package pdexv3

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

func TestMintNft_ValidateSanityData(t *testing.T) {
	type fields struct {
		nftID        string
		otaReceiver  string
		MetadataBase metadataCommon.MetadataBase
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
			name: "Invalid ota receive",
			fields: fields{
				otaReceiver: "132",
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid nftID",
			fields: fields{
				otaReceiver: validOTAReceiver0,
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Empty nftID",
			fields: fields{
				otaReceiver: validOTAReceiver0,
				nftID:       common.Hash{}.String(),
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Valid Input",
			fields: fields{
				otaReceiver: validOTAReceiver0,
				nftID:       common.PRVIDStr,
			},
			args:    args{},
			want:    true,
			want1:   true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mintNft := &MintNftResponse{
				nftID:        tt.fields.nftID,
				otaReceiver:  tt.fields.otaReceiver,
				MetadataBase: tt.fields.MetadataBase,
			}
			got, got1, err := mintNft.ValidateSanityData(tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.beaconHeight, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("MintNft.ValidateSanityData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MintNft.ValidateSanityData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("MintNft.ValidateSanityData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
