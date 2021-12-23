package pdexv3

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

const (
	validOTAReceiver0 = "15sXoyo8kCZCHjurNC69b8WV2jMCvf5tVrcQ5mT1eH9Nm351XRjE1BH4WHHLGYPZy9dxTSLiKQd6KdfoGq4yb4gP1AU2oaJTeoGymsEzonyi1XSW2J2U7LeAVjS1S2gjbNDk1t3f9QUg2gk4"
	validOTAReceiver1 = "15ujixNQY1Qc5wyX9UYQW3s6cbcecFPNhrWjWiFCggeN5HukPVdjbKyRE3goUpFgZhawtBtRUK3ZSZb5LtH7bevhGzz3UTh1muzLHG3pvsE6RNB81y8xNGhyHdpHZfjwmSWDdwDe74Tg2CUP"
)

/*var (*/
//validOTAReceiver0 = privacy.OTAReceiver{}
//validOTAReceiver1 = privacy.OTAReceiver{}
/*)*/

/*func initTestParam(t *testing.T) {*/
//err := validOTAReceiver0.FromString(validOTAReceiver0Str)
//assert.Nil(t, err)
//err = validOTAReceiver1.FromString(validOTAReceiver1Str)
//assert.Nil(t, err)
/*}*/

func TestAccessOption_ValidateOtaReceivers(t *testing.T) {
	type fields struct {
		NftID    *common.Hash
		BurntOTA *AccessOTA
		AccessID *common.Hash
	}
	type args struct {
		tx           metadataCommon.Transaction
		otaReceiver  string
		otaReceivers map[common.Hash]privacy.OTAReceiver
		tokenHash    common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AccessOption{
				NftID:    tt.fields.NftID,
				BurntOTA: tt.fields.BurntOTA,
				AccessID: tt.fields.AccessID,
			}
			if err := a.ValidateOtaReceivers(tt.args.tx, tt.args.otaReceiver, tt.args.otaReceivers, tt.args.tokenHash); (err != nil) != tt.wantErr {
				t.Errorf("AccessOption.ValidateOtaReceivers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccessOption_IsValid(t *testing.T) {
	type fields struct {
		NftID    *common.Hash
		BurntOTA *AccessOTA
		AccessID *common.Hash
	}
	type args struct {
		tx                  metadataCommon.Transaction
		receivers           map[common.Hash]privacy.OTAReceiver
		beaconViewRetriever metadataCommon.BeaconViewRetriever
		transactionStateDB  *statedb.StateDB
		isWithdrawalRequest bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Use nftID - Empty nftID",
			fields: fields{
				NftID: &common.Hash{},
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "Use nftID - Empty nftID",
			fields: fields{
				NftID: &common.Hash{},
			},
			args:    args{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AccessOption{
				NftID:    tt.fields.NftID,
				BurntOTA: tt.fields.BurntOTA,
				AccessID: tt.fields.AccessID,
			}
			if err := a.IsValid(tt.args.tx, tt.args.receivers, tt.args.beaconViewRetriever, tt.args.transactionStateDB, tt.args.isWithdrawalRequest); (err != nil) != tt.wantErr {
				t.Errorf("AccessOption.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
