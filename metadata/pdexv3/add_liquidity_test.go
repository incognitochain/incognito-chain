package metadata

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

func TestAddLiquidity_ValidateSanityData(t *testing.T) {
	type fields struct {
		PoolPairID      string
		PairHash        string
		ReceiverAddress privacy.OTAReceiver
		RefundAddress   privacy.OTAReceiver
		TokenID         common.Hash
		TokenAmount     uint64
		Amplifier       uint
		MetadataBase    metadataCommon.MetadataBase
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := &AddLiquidity{
				PoolPairID:      tt.fields.PoolPairID,
				PairHash:        tt.fields.PairHash,
				ReceiverAddress: tt.fields.ReceiverAddress,
				RefundAddress:   tt.fields.RefundAddress,
				TokenID:         tt.fields.TokenID,
				TokenAmount:     tt.fields.TokenAmount,
				Amplifier:       tt.fields.Amplifier,
				MetadataBase:    tt.fields.MetadataBase,
			}
			got, got1, err := al.ValidateSanityData(tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.beaconHeight, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddLiquidity.ValidateSanityData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AddLiquidity.ValidateSanityData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("AddLiquidity.ValidateSanityData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestAddLiquidity_ValidateMetadataByItself(t *testing.T) {
	type fields struct {
		PoolPairID      string
		PairHash        string
		ReceiverAddress privacy.OTAReceiver
		RefundAddress   privacy.OTAReceiver
		TokenID         common.Hash
		TokenAmount     uint64
		Amplifier       uint
		MetadataBase    metadataCommon.MetadataBase
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "",
			fields: fields{},
			want:   false,
		},
		{
			name:   "",
			fields: fields{},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := &AddLiquidity{
				PoolPairID:      tt.fields.PoolPairID,
				PairHash:        tt.fields.PairHash,
				ReceiverAddress: tt.fields.ReceiverAddress,
				RefundAddress:   tt.fields.RefundAddress,
				TokenID:         tt.fields.TokenID,
				TokenAmount:     tt.fields.TokenAmount,
				Amplifier:       tt.fields.Amplifier,
				MetadataBase:    tt.fields.MetadataBase,
			}
			if got := al.ValidateMetadataByItself(); got != tt.want {
				t.Errorf("AddLiquidity.ValidateMetadataByItself() = %v, want %v", got, tt.want)
			}
		})
	}
}
