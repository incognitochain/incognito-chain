package pdexv3

import (
	"errors"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataCommonMocks "github.com/incognitochain/incognito-chain/metadata/common/mocks"
	coinMocks "github.com/incognitochain/incognito-chain/privacy/coin/mocks"
	"github.com/stretchr/testify/assert"
)

func TestWithdrawLiquidityRequest_ValidateSanityData(t *testing.T) {
	tokenHash, err := common.Hash{}.NewHashFromStr("123123")
	assert.Nil(t, err)

	validationEnv := &metadataCommonMocks.ValidationEnviroment{}
	validationEnv.On("ShardID").Return(0)

	notBurnTx := &metadataCommonMocks.Transaction{}
	notBurnTx.On("GetTxBurnData").Return(false, nil, nil, errors.New("Not tx burn"))
	notBurnTx.On("GetValidationEnv").Return(validationEnv)

	notMactchTokenIDTx := &metadataCommonMocks.Transaction{}
	notMactchTokenIDTx.On("GetTxBurnData").Return(true, nil, &common.PRVCoinID, nil)
	notMactchTokenIDTx.On("GetValidationEnv").Return(validationEnv)

	notMatchAmountCoin := &coinMocks.Coin{}
	notMatchAmountCoin.On("GetValue").Return(uint64(100))
	notMactchAmountTx := &metadataCommonMocks.Transaction{}
	notMactchAmountTx.On("GetTxBurnData").Return(true, notMatchAmountCoin, tokenHash, nil)
	notMactchAmountTx.On("GetValidationEnv").Return(validationEnv)

	validCoin := &coinMocks.Coin{}
	validCoin.On("GetValue").Return(uint64(1))

	normalTx := &metadataCommonMocks.Transaction{}
	normalTx.On("GetTxBurnData").Return(true, validCoin, tokenHash, nil)
	normalTx.On("GetType").Return(common.TxNormalType)
	normalTx.On("GetValidationEnv").Return(validationEnv)

	customTx := &metadataCommonMocks.Transaction{}
	customTx.On("GetTxBurnData").Return(true, validCoin, &common.PRVCoinID, nil)
	customTx.On("GetType").Return(common.TxCustomTokenPrivacyType)
	customTx.On("GetValidationEnv").Return(validationEnv)

	validTx := &metadataCommonMocks.Transaction{}
	validTx.On("GetTxBurnData").Return(true, validCoin, tokenHash, nil)
	validTx.On("GetType").Return(common.TxCustomTokenPrivacyType)
	validTx.On("GetValidationEnv").Return(validationEnv)

	type fields struct {
		MetadataBase     metadataCommon.MetadataBase
		poolPairID       string
		nftID            string
		otaReceiveNft    string
		otaReceiveToken0 string
		otaReceiveToken1 string
		shareAmount      uint64
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
			name: "Invalid poolPairID",
			fields: fields{
				poolPairID: "",
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid NftID",
			fields: fields{
				poolPairID: "123",
				nftID:      "abc",
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Empty NftID",
			fields: fields{
				poolPairID: "123",
				nftID:      common.Hash{}.String(),
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid ota receive nft",
			fields: fields{
				poolPairID: "123",
				nftID:      tokenHash.String(),
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid ota token 0",
			fields: fields{
				poolPairID:       "123",
				nftID:            tokenHash.String(),
				otaReceiveNft:    validOTAReceiver0,
				otaReceiveToken0: "123",
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid ota token 1",
			fields: fields{
				poolPairID:       "123",
				nftID:            tokenHash.String(),
				otaReceiveNft:    validOTAReceiver0,
				otaReceiveToken0: validOTAReceiver1,
				otaReceiveToken1: "123",
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid shareAmount",
			fields: fields{
				poolPairID:       "123",
				nftID:            tokenHash.String(),
				otaReceiveNft:    validOTAReceiver0,
				otaReceiveToken0: validOTAReceiver1,
				otaReceiveToken1: validOTAReceiver1,
				shareAmount:      0,
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Tx is not burnt tx",
			fields: fields{
				poolPairID:       "123",
				nftID:            tokenHash.String(),
				otaReceiveNft:    validOTAReceiver0,
				otaReceiveToken0: validOTAReceiver1,
				otaReceiveToken1: validOTAReceiver1,
				shareAmount:      100,
			},
			args: args{
				tx: notBurnTx,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "tokenID not match with burn coin",
			fields: fields{
				poolPairID:       "123",
				nftID:            tokenHash.String(),
				otaReceiveNft:    validOTAReceiver0,
				otaReceiveToken0: validOTAReceiver1,
				otaReceiveToken1: validOTAReceiver1,
				shareAmount:      100,
			},
			args: args{
				tx: notMactchTokenIDTx,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "burn coin value is not 1",
			fields: fields{
				poolPairID:       "123",
				nftID:            tokenHash.String(),
				otaReceiveNft:    validOTAReceiver0,
				otaReceiveToken0: validOTAReceiver1,
				otaReceiveToken1: validOTAReceiver1,
				shareAmount:      100,
			},
			args: args{
				tx: notMactchAmountTx,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "type of tx not is custom privacy type",
			fields: fields{
				poolPairID:       "123",
				nftID:            tokenHash.String(),
				otaReceiveNft:    validOTAReceiver0,
				otaReceiveToken0: validOTAReceiver1,
				otaReceiveToken1: validOTAReceiver1,
				shareAmount:      100,
			},
			args: args{
				tx: normalTx,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "nftID == prv",
			fields: fields{
				poolPairID:       "123",
				nftID:            tokenHash.String(),
				otaReceiveNft:    validOTAReceiver0,
				otaReceiveToken0: validOTAReceiver1,
				otaReceiveToken1: validOTAReceiver1,
				shareAmount:      100,
			},
			args: args{
				tx: customTx,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Valid Input",
			fields: fields{
				poolPairID:       "123",
				nftID:            tokenHash.String(),
				otaReceiveNft:    validOTAReceiver0,
				otaReceiveToken0: validOTAReceiver1,
				otaReceiveToken1: validOTAReceiver1,
				shareAmount:      100,
			},
			args: args{
				tx: validTx,
			},
			want:    true,
			want1:   true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &WithdrawLiquidityRequest{
				MetadataBase:     tt.fields.MetadataBase,
				poolPairID:       tt.fields.poolPairID,
				nftID:            tt.fields.nftID,
				otaReceiveNft:    tt.fields.otaReceiveNft,
				otaReceiveToken0: tt.fields.otaReceiveToken0,
				shareAmount:      tt.fields.shareAmount,
				otaReceiveToken1: tt.fields.otaReceiveToken1,
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

func TestWithdrawLiquidityRequest_ValidateMetadataByItself(t *testing.T) {
	type fields struct {
		MetadataBase     metadataCommon.MetadataBase
		poolPairID       string
		nftID            string
		otaReceiveNft    string
		otaReceiveToken0 string
		otaReceiveToken1 string
		shareAmount      uint64
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
					Type: metadataCommon.Pdexv3WithdrawLiquidityRequestMeta,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &WithdrawLiquidityRequest{
				MetadataBase:     tt.fields.MetadataBase,
				poolPairID:       tt.fields.poolPairID,
				nftID:            tt.fields.nftID,
				otaReceiveNft:    tt.fields.otaReceiveNft,
				otaReceiveToken0: tt.fields.otaReceiveToken0,
				otaReceiveToken1: tt.fields.otaReceiveToken1,
				shareAmount:      tt.fields.shareAmount,
			}
			if got := request.ValidateMetadataByItself(); got != tt.want {
				t.Errorf("WithdrawLiquidityRequest.ValidateMetadataByItself() = %v, want %v", got, tt.want)
			}
		})
	}
}
