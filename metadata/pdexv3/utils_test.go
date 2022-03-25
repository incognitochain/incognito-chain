package pdexv3

import (
	"errors"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataCommonMocks "github.com/incognitochain/incognito-chain/metadata/common/mocks"
	"github.com/incognitochain/incognito-chain/privacy"
	coinMocks "github.com/incognitochain/incognito-chain/privacy/coin/mocks"
	"github.com/stretchr/testify/assert"
)

const (
	validOTAReceiver0 = "15sXoyo8kCZCHjurNC69b8WV2jMCvf5tVrcQ5mT1eH9Nm351XRjE1BH4WHHLGYPZy9dxTSLiKQd6KdfoGq4yb4gP1AU2oaJTeoGymsEzonyi1XSW2J2U7LeAVjS1S2gjbNDk1t3f9QUg2gk4"
	validOTAReceiver1 = "15ujixNQY1Qc5wyX9UYQW3s6cbcecFPNhrWjWiFCggeN5HukPVdjbKyRE3goUpFgZhawtBtRUK3ZSZb5LtH7bevhGzz3UTh1muzLHG3pvsE6RNB81y8xNGhyHdpHZfjwmSWDdwDe74Tg2CUP"
	validAccessOTA    = "5xbO6s+gO8pn/Irevhdy6l7S3A64oKGKkAENpRTI5MA="
)

var (
	otaReceiver0 = privacy.OTAReceiver{}
	otaReceiver1 = privacy.OTAReceiver{}
	accessOTA    = new(AccessOTA)
)

func initTestParam(t *testing.T) {
	common.MaxShardNumber = 8
	err := otaReceiver0.FromString(validOTAReceiver0)
	assert.Nil(t, err)
	err = otaReceiver1.FromString(validOTAReceiver1)
	assert.Nil(t, err)
	err = accessOTA.FromString(validAccessOTA)
	assert.Nil(t, err)
}

func TestAccessOption_ValidateOtaReceivers(t *testing.T) {
	initTestParam(t)
	nftID, err := common.Hash{}.NewHashFromStr("123456")
	assert.Nil(t, err)
	invalidValEnv := &metadataCommonMocks.ValidationEnviroment{}
	invalidValEnv.On("ShardID").Return(2)
	validValEnv := &metadataCommonMocks.ValidationEnviroment{}
	validValEnv.On("ShardID").Return(1)
	invalidShardIDTx := &metadataCommonMocks.Transaction{}
	invalidShardIDTx.On("GetValidationEnv").Return(invalidValEnv)
	validTx := &metadataCommonMocks.Transaction{}
	validTx.On("GetValidationEnv").Return(validValEnv)

	type fields struct {
		NftID    *common.Hash
		BurntOTA *AccessOTA
		AccessID *common.Hash
	}
	type args struct {
		tx             metadataCommon.Transaction
		otaReceiver    string
		otaReceivers   map[common.Hash]privacy.OTAReceiver
		tokenHash      common.Hash
		isNewLpRequest bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "otaReceiver and otaReceivers null at the same time",
			fields:  fields{},
			args:    args{},
			wantErr: true,
		},
		{
			name:   "otaReceiver and otaReceivers null at the same time - 1",
			fields: fields{},
			args: args{
				otaReceivers: map[common.Hash]privacy.OTAReceiver{},
			},
			wantErr: true,
		},
		{
			name:   "otaReceiver and otaReceivers exist at the same time ",
			fields: fields{},
			args: args{
				otaReceivers: map[common.Hash]privacy.OTAReceiver{},
				otaReceiver:  validOTAReceiver0,
			},
			wantErr: true,
		},
		{
			name: "otaReceiver is not valid",
			fields: fields{
				NftID: nftID,
			},
			args: args{
				otaReceiver: "123",
			},
			wantErr: true,
		},
		{
			name: "otaReceiver from other shard with tx's shard",
			fields: fields{
				NftID: nftID,
			},
			args: args{
				otaReceiver: validOTAReceiver0,
				tx:          invalidShardIDTx,
			},
			wantErr: true,
		},
		{
			name: "AccessID - otaReceivers == nil",
			fields: fields{
				AccessID: nftID,
			},
			args: args{
				tx: invalidShardIDTx,
			},
			wantErr: true,
		},
		{
			name: "AccessID - len(otaReceivers) == 0",
			fields: fields{
				AccessID: nftID,
			},
			args: args{
				otaReceivers: map[common.Hash]privacy.OTAReceiver{},
				tx:           invalidShardIDTx,
			},
			wantErr: true,
		},
		{
			name: "AccessID - not found otaReceiver for tokenID",
			fields: fields{
				AccessID: nftID,
			},
			args: args{
				otaReceivers: map[common.Hash]privacy.OTAReceiver{
					common.PdexAccessCoinID: otaReceiver0,
				},
				tx:        invalidShardIDTx,
				tokenHash: common.PRVCoinID,
			},
			wantErr: true,
		},
		{
			name: "AccessID - Invalid shard ID",
			fields: fields{
				AccessID: nftID,
			},
			args: args{
				otaReceivers: map[common.Hash]privacy.OTAReceiver{
					common.PRVCoinID: otaReceiver0,
				},
				tx:        invalidShardIDTx,
				tokenHash: common.PRVCoinID,
			},
			wantErr: true,
		},
		{
			name:   "Add request - no pdex access receiver",
			fields: fields{},
			args: args{
				otaReceivers: map[common.Hash]privacy.OTAReceiver{
					common.PRVCoinID: otaReceiver0,
				},
				tx:        validTx,
				tokenHash: common.PRVCoinID,
			},
			wantErr: true,
		},
		{
			name:   "Add request - Valid input",
			fields: fields{},
			args: args{
				otaReceivers: map[common.Hash]privacy.OTAReceiver{
					common.PRVCoinID:        otaReceiver0,
					common.PdexAccessCoinID: otaReceiver0,
				},
				tx:        validTx,
				tokenHash: common.PRVCoinID,
			},
			wantErr: false,
		},
		{
			name: "Add request - Valid input",
			fields: fields{
				AccessID: nftID,
			},
			args: args{
				otaReceivers: map[common.Hash]privacy.OTAReceiver{
					common.PRVCoinID: otaReceiver0,
				},
				tx:        validTx,
				tokenHash: common.PRVCoinID,
			},
			wantErr: false,
		},
		{
			name: "Add request - Valid input",
			fields: fields{
				AccessID: nil,
			},
			args: args{
				otaReceiver: validOTAReceiver0,
				otaReceivers: map[common.Hash]privacy.OTAReceiver{
					common.PRVCoinID: otaReceiver0,
				},
				isNewLpRequest: true,
				tx:             validTx,
				tokenHash:      common.PRVCoinID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AccessOption{
				NftID:    tt.fields.NftID,
				BurntOTA: tt.fields.BurntOTA,
				AccessID: tt.fields.AccessID,
			}
			if err := a.ValidateOtaReceivers(tt.args.tx, tt.args.otaReceiver, tt.args.otaReceivers, tt.args.tokenHash, tt.args.isNewLpRequest); (err != nil) != tt.wantErr {
				t.Errorf("AccessOption.ValidateOtaReceivers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccessOption_IsValid(t *testing.T) {
	initTestParam(t)
	nftID, err := common.Hash{}.NewHashFromStr("123456")
	assert.Nil(t, err)
	notFoundNftIDView := &metadataCommonMocks.BeaconViewRetriever{}
	notFoundNftIDView.On("IsValidPdexv3NftID", nftID.String()).Return(false, errors.New("123"))
	foundNftIDView := &metadataCommonMocks.BeaconViewRetriever{}
	foundNftIDView.On("IsValidPdexv3NftID", nftID.String()).Return(true, nil)
	tempPrivacyPoint := &privacy.Point{}
	tempAccessOTA := accessOTA.ToBytes()
	tempPrivacyPoint, err = tempPrivacyPoint.FromBytes(tempAccessOTA)
	assert.Nil(t, err)
	invalidTx := &metadataCommonMocks.Transaction{}
	coin := &coinMocks.Coin{}
	coin.On("GetValue").Return(1)
	invalidTx.On("GetTxFullBurnData").Return(true, &coinMocks.Coin{}, coin, common.PdexAccessCoinID, errors.New("errror"))
	invalidTx.On("DerivableBurnInput").Return(map[common.Hash]privacy.Point{
		common.PdexAccessCoinID: *tempPrivacyPoint,
	}, nil)
	invalidTx.On("Hash").Return(&common.PRVCoinID)

	type fields struct {
		NftID    *common.Hash
		BurntOTA *AccessOTA
		AccessID *common.Hash
	}
	type args struct {
		tx                      metadataCommon.Transaction
		receivers               map[common.Hash]privacy.OTAReceiver
		beaconViewRetriever     metadataCommon.BeaconViewRetriever
		transactionStateDB      *statedb.StateDB
		isWithdrawalRequest     bool
		isNewAccessOTALpRequest bool
		accessReceiverStr       string
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
			name: "Use nftID - Not found nftID",
			fields: fields{
				NftID: nftID,
			},
			args: args{
				beaconViewRetriever: notFoundNftIDView,
			},
			wantErr: true,
		},
		{
			name: "Use nftID - burnt ota or access id is not null",
			fields: fields{
				NftID:    nftID,
				AccessID: nftID,
			},
			args: args{
				beaconViewRetriever: foundNftIDView,
			},
			wantErr: true,
		},
		{
			name: "Use nftID - valid input",
			fields: fields{
				NftID: nftID,
			},
			args: args{
				beaconViewRetriever: foundNftIDView,
			},
			wantErr: false,
		},
		{
			name:   "Use accessID - withdrawal request - accessID is null",
			fields: fields{},
			args: args{
				isWithdrawalRequest: true,
			},
			wantErr: true,
		},
		{
			name: "Use accessID - withdrawal request - accessID is zero value",
			fields: fields{
				AccessID: &common.Hash{},
			},
			args: args{
				isWithdrawalRequest: true,
			},
			wantErr: true,
		},
		{
			name: "Use accessID - withdrawal request - burnt ota = null",
			fields: fields{
				AccessID: nftID,
			},
			args: args{
				isWithdrawalRequest: true,
			},
			wantErr: true,
		},
		{
			name: "Use accessID - add request - null receivers",
			fields: fields{
				AccessID: nftID,
			},
			args: args{
				isWithdrawalRequest: false,
			},
			wantErr: true,
		},
		{
			name:   "Use accessID - add request - null access receiver",
			fields: fields{},
			args: args{
				isWithdrawalRequest: false,
				receivers: map[common.Hash]privacy.OTAReceiver{
					common.PRVCoinID: otaReceiver0,
				},
			},
			wantErr: true,
		},
		{
			name:   "Use accessID - add request - Success",
			fields: fields{},
			args: args{
				isWithdrawalRequest: false,
				receivers: map[common.Hash]privacy.OTAReceiver{
					common.PdexAccessCoinID: otaReceiver0,
				},
			},
			wantErr: false,
		},
		{
			name:   "AccessID is null and NftID is null",
			fields: fields{},
			args: args{
				isWithdrawalRequest: false,
				receivers: map[common.Hash]privacy.OTAReceiver{
					common.PdexAccessCoinID: otaReceiver0,
				},
			},
			wantErr: false,
		},
		{
			name: "Use accessID - withdrawal request - invalid burntOTA",
			fields: fields{
				AccessID: nftID,
				BurntOTA: accessOTA,
			},
			args: args{
				isWithdrawalRequest: true,
				receivers: map[common.Hash]privacy.OTAReceiver{
					common.PdexAccessCoinID: otaReceiver0,
				},
				tx: invalidTx,
			},
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
			if err := a.IsValid(tt.args.tx, tt.args.receivers, tt.args.beaconViewRetriever, tt.args.transactionStateDB, tt.args.isWithdrawalRequest, tt.args.isNewAccessOTALpRequest, tt.args.accessReceiverStr); (err != nil) != tt.wantErr {
				t.Errorf("AccessOption.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
