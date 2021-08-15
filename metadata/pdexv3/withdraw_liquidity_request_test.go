package pdexv3

import (
	"testing"

	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

func TestWithdrawLiquidityRequest_ValidateSanityData(t *testing.T) {
	type fields struct {
		MetadataBase          metadataCommon.MetadataBase
		poolPairID            string
		nftID                 string
		otaReceiveNft         string
		token0Amount          uint64
		otaReceiveToken0      string
		token1Amount          uint64
		otaReceiveToken1      string
		otaReceiveTradingFees map[string]string
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
			request := &WithdrawLiquidityRequest{
				MetadataBase:          tt.fields.MetadataBase,
				poolPairID:            tt.fields.poolPairID,
				nftID:                 tt.fields.nftID,
				otaReceiveNft:         tt.fields.otaReceiveNft,
				token0Amount:          tt.fields.token0Amount,
				otaReceiveToken0:      tt.fields.otaReceiveToken0,
				token1Amount:          tt.fields.token1Amount,
				otaReceiveToken1:      tt.fields.otaReceiveToken1,
				otaReceiveTradingFees: tt.fields.otaReceiveTradingFees,
			}
			got, got1, err := request.ValidateSanityData(tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.beaconHeight, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithdrawLiquidityRequest.ValidateSanityData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("WithdrawLiquidityRequest.ValidateSanityData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("WithdrawLiquidityRequest.ValidateSanityData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
