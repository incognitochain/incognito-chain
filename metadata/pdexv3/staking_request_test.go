package pdexv3

import (
	"testing"

	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

func TestStakingRequest_ValidateSanityData(t *testing.T) {
	type fields struct {
		MetadataBase metadataCommon.MetadataBase
		tokenID      string
		otaReceivers map[string]string
		nftID        string
		tokenAmount  uint64
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
			request := &StakingRequest{
				MetadataBase: tt.fields.MetadataBase,
				tokenID:      tt.fields.tokenID,
				otaReceivers: tt.fields.otaReceivers,
				nftID:        tt.fields.nftID,
				tokenAmount:  tt.fields.tokenAmount,
			}
			got, got1, err := request.ValidateSanityData(tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.beaconHeight, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("StakingRequest.ValidateSanityData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StakingRequest.ValidateSanityData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("StakingRequest.ValidateSanityData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestStakingRequest_ValidateMetadataByItself(t *testing.T) {
	type fields struct {
		MetadataBase metadataCommon.MetadataBase
		tokenID      string
		otaReceivers map[string]string
		nftID        string
		tokenAmount  uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &StakingRequest{
				MetadataBase: tt.fields.MetadataBase,
				tokenID:      tt.fields.tokenID,
				otaReceivers: tt.fields.otaReceivers,
				nftID:        tt.fields.nftID,
				tokenAmount:  tt.fields.tokenAmount,
			}
			if got := request.ValidateMetadataByItself(); got != tt.want {
				t.Errorf("StakingRequest.ValidateMetadataByItself() = %v, want %v", got, tt.want)
			}
		})
	}
}
