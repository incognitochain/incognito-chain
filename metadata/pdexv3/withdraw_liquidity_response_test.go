package pdexv3

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

func TestWithdrawLiquidityResponse_ValidateSanityData(t *testing.T) {
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
			name: "Invalid status",
			fields: fields{
				status: common.PDEWithdrawalRejectedChainStatus,
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "txReqID is invalid",
			fields: fields{
				status: common.PDEWithdrawalAcceptedChainStatus,
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "txReqID is empty",
			fields: fields{
				status:  common.PDEWithdrawalAcceptedChainStatus,
				txReqID: common.Hash{}.String(),
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Valid Input",
			fields: fields{
				status:  common.PDEWithdrawalAcceptedChainStatus,
				txReqID: common.PRVIDStr,
			},
			args:    args{},
			want:    true,
			want1:   true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &WithdrawLiquidityResponse{
				MetadataBase: tt.fields.MetadataBase,
				status:       tt.fields.status,
				txReqID:      tt.fields.txReqID,
			}
			got, got1, err := response.ValidateSanityData(tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.beaconHeight, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithdrawLiquidityResponse.ValidateSanityData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("WithdrawLiquidityResponse.ValidateSanityData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("WithdrawLiquidityResponse.ValidateSanityData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
