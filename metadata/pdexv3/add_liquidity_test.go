package pdexv3

import (
	"errors"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataCommonMocks "github.com/incognitochain/incognito-chain/metadata/common/mocks"
	"github.com/incognitochain/incognito-chain/privacy"
	coinMocks "github.com/incognitochain/incognito-chain/privacy/coin/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAddLiquidity_ValidateSanityData(t *testing.T) {

	initTestParam(t)

	tokenHash, err := common.Hash{}.NewHashFromStr("123123")
	assert.Nil(t, err)

	notBurnTx := &metadataCommonMocks.Transaction{}
	notBurnTx.On("GetTxBurnData").Return(false, nil, nil, errors.New("Not tx burn"))

	notMactchTokenIDTx := &metadataCommonMocks.Transaction{}
	notMactchTokenIDTx.On("GetTxBurnData").Return(true, nil, &common.PRVCoinID, nil)

	notMatchAmountCoin := &coinMocks.Coin{}
	notMatchAmountCoin.On("GetValue").Return(uint64(100))
	notMactchAmountTx0 := &metadataCommonMocks.Transaction{}
	notMactchAmountTx0.On("GetTxBurnData").Return(true, notMatchAmountCoin, tokenHash, nil)

	validCoin := &coinMocks.Coin{}
	validCoin.On("GetValue").Return(uint64(200))
	notMactchAmountTx1 := &metadataCommonMocks.Transaction{}
	notMactchAmountTx1.On("GetTxBurnData").Return(true, notMatchAmountCoin, tokenHash, nil)

	invalidNormalTx := &metadataCommonMocks.Transaction{}
	invalidNormalTx.On("GetTxBurnData").Return(true, validCoin, tokenHash, nil)
	invalidNormalTx.On("GetType").Return(common.TxNormalType)

	invalidCustomTx := &metadataCommonMocks.Transaction{}
	invalidCustomTx.On("GetTxBurnData").Return(true, validCoin, &common.PRVCoinID, nil)
	invalidCustomTx.On("GetType").Return(common.TxCustomTokenPrivacyType)

	validTx := &metadataCommonMocks.Transaction{}
	validTx.On("GetTxBurnData").Return(true, validCoin, tokenHash, nil)
	validTx.On("GetType").Return(common.TxCustomTokenPrivacyType)

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
		{
			name: "Invalid PairHash",
			fields: fields{
				PairHash: "",
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Empty tokenID",
			fields: fields{
				PairHash: "pair hash",
				TokenID:  common.Hash{},
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid ReceiverAddress",
			fields: fields{
				PairHash: "pair hash",
				TokenID:  *tokenHash,
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid RefundAddress",
			fields: fields{
				PairHash:        "pair hash",
				TokenID:         *tokenHash,
				ReceiverAddress: validOTAReceiver0,
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid amplifier",
			fields: fields{
				PairHash:        "pair hash",
				TokenID:         *tokenHash,
				ReceiverAddress: validOTAReceiver0,
				RefundAddress:   validOTAReceiver1,
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Tx is not burn tx",
			fields: fields{
				PairHash:        "pair hash",
				TokenID:         *tokenHash,
				ReceiverAddress: validOTAReceiver0,
				RefundAddress:   validOTAReceiver1,
				Amplifier:       10000,
			},
			args: args{
				tx: notBurnTx,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "tokenID not match with burnCoin",
			fields: fields{
				PairHash:        "pair hash",
				TokenID:         *tokenHash,
				ReceiverAddress: validOTAReceiver0,
				RefundAddress:   validOTAReceiver1,
				Amplifier:       10000,
			},
			args: args{
				tx: notMactchTokenIDTx,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Token amount = 0",
			fields: fields{
				PairHash:        "pair hash",
				TokenID:         *tokenHash,
				ReceiverAddress: validOTAReceiver0,
				RefundAddress:   validOTAReceiver1,
				Amplifier:       10000,
			},
			args: args{
				tx: notMactchAmountTx0,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Contributed amount is not match with burn amount",
			fields: fields{
				PairHash:        "pair hash",
				TokenID:         *tokenHash,
				ReceiverAddress: validOTAReceiver0,
				RefundAddress:   validOTAReceiver1,
				Amplifier:       10000,
				TokenAmount:     200,
			},
			args: args{
				tx: notMactchAmountTx1,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Normal tx && tokenID != prv",
			fields: fields{
				PairHash:        "pair hash",
				TokenID:         *tokenHash,
				ReceiverAddress: validOTAReceiver0,
				RefundAddress:   validOTAReceiver1,
				Amplifier:       10000,
				TokenAmount:     200,
			},
			args: args{
				tx: invalidNormalTx,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},

		{
			name: "Custom token tx && tokenID == prv",
			fields: fields{
				PairHash:        "pair hash",
				TokenID:         common.PRVCoinID,
				ReceiverAddress: validOTAReceiver0,
				RefundAddress:   validOTAReceiver1,
				Amplifier:       10000,
				TokenAmount:     200,
			},
			args: args{
				tx: invalidCustomTx,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Valid Input",
			fields: fields{
				PairHash:        "pair hash",
				TokenID:         *tokenHash,
				ReceiverAddress: validOTAReceiver0,
				RefundAddress:   validOTAReceiver1,
				Amplifier:       10000,
				TokenAmount:     200,
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
			name: "Invalid Input",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.PDexV3TradeResponseMeta,
				},
			},
			want: false,
		},
		{
			name: "Valid Input",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.PDexV3AddLiquidityMeta,
				},
			},
			want: true,
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
