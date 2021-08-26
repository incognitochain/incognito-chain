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

func TestUserMintNftRequest_ValidateSanityData(t *testing.T) {
	tokenHash, err := common.Hash{}.NewHashFromStr("123123")
	assert.Nil(t, err)

	notBurnTx := &metadataCommonMocks.Transaction{}
	notBurnTx.On("GetTxBurnData").Return(false, nil, nil, errors.New("Not tx burn"))

	notMactchTokenIDTx := &metadataCommonMocks.Transaction{}
	notMactchTokenIDTx.On("GetTxBurnData").Return(true, nil, tokenHash, nil)

	notMatchAmountCoin := &coinMocks.Coin{}
	notMatchAmountCoin.On("GetValue").Return(uint64(100))
	notMactchAmountTx := &metadataCommonMocks.Transaction{}
	notMactchAmountTx.On("GetTxBurnData").Return(true, notMatchAmountCoin, &common.PRVCoinID, nil)

	validCoin := &coinMocks.Coin{}
	validCoin.On("GetValue").Return(uint64(1))

	customTx := &metadataCommonMocks.Transaction{}
	customTx.On("GetTxBurnData").Return(true, validCoin, &common.PRVCoinID, nil)
	customTx.On("GetType").Return(common.TxCustomTokenPrivacyType)

	validTx := &metadataCommonMocks.Transaction{}
	validTx.On("GetTxBurnData").Return(true, validCoin, &common.PRVCoinID, nil)
	validTx.On("GetType").Return(common.TxNormalType)

	type fields struct {
		MetadataBase metadataCommon.MetadataBase
		otaReceiver  string
		amount       uint64
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
				otaReceiver: "123",
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "amount = 0",
			fields: fields{
				otaReceiver: validOTAReceiver0,
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Not tx burnt",
			fields: fields{
				otaReceiver: validOTAReceiver0,
				amount:      1,
			},
			args: args{
				tx: notBurnTx,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Burnt coin is not prv",
			fields: fields{
				otaReceiver: validOTAReceiver0,
				amount:      1,
			},
			args: args{
				tx: notMactchTokenIDTx,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Burn coin is not similar to metadata amount",
			fields: fields{
				otaReceiver: validOTAReceiver0,
				amount:      1,
			},
			args: args{
				tx: notMactchAmountTx,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Tx type is not normal type",
			fields: fields{
				otaReceiver: validOTAReceiver0,
				amount:      1,
			},
			args: args{
				tx: customTx,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Valid input",
			fields: fields{
				otaReceiver: validOTAReceiver0,
				amount:      1,
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
			request := &UserMintNftRequest{
				MetadataBase: tt.fields.MetadataBase,
				otaReceiver:  tt.fields.otaReceiver,
				amount:       tt.fields.amount,
			}
			got, got1, err := request.ValidateSanityData(tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.beaconHeight, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserMintNftRequest.ValidateSanityData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UserMintNftRequest.ValidateSanityData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("UserMintNftRequest.ValidateSanityData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestUserMintNftRequest_ValidateMetadataByItself(t *testing.T) {
	type fields struct {
		MetadataBase metadataCommon.MetadataBase
		otaReceiver  string
		amount       uint64
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
					Type: metadataCommon.Pdexv3UserMintNftRequestMeta,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &UserMintNftRequest{
				MetadataBase: tt.fields.MetadataBase,
				otaReceiver:  tt.fields.otaReceiver,
				amount:       tt.fields.amount,
			}
			if got := request.ValidateMetadataByItself(); got != tt.want {
				t.Errorf("UserMintNftRequest.ValidateMetadataByItself() = %v, want %v", got, tt.want)
			}
		})
	}
}
