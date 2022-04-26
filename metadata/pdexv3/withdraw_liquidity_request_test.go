package pdexv3

import (
	"errors"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataCommonMocks "github.com/incognitochain/incognito-chain/metadata/common/mocks"
	coinMocks "github.com/incognitochain/incognito-chain/privacy/coin/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWithdrawLiquidityRequest_ValidateSanityData(t *testing.T) {
	common.MaxShardNumber = 8
	tokenHash, err := common.Hash{}.NewHashFromStr("123123")
	assert.Nil(t, err)
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)

	invalidChainRetriever := &metadataCommonMocks.ChainRetriever{}
	invalidChainRetriever.On("IsAfterPdexv3CheckPoint", mock.AnythingOfType("uint64")).Return(false)
	validChainRetriever := &metadataCommonMocks.ChainRetriever{}
	validChainRetriever.On("IsAfterPdexv3CheckPoint", mock.AnythingOfType("uint64")).Return(true)

	validValidationEnvironment := &metadataCommonMocks.ValidationEnviroment{}
	validValidationEnvironment.On("ShardID").Return(1)

	notBurnTx := &metadataCommonMocks.Transaction{}
	notBurnTx.On("GetTxBurnData").Return(false, nil, nil, errors.New("Not tx burn"))
	notBurnTx.On("GetValidationEnv").Return(validValidationEnvironment)

	notMactchTokenIDTx := &metadataCommonMocks.Transaction{}
	notMactchTokenIDTx.On("GetTxBurnData").Return(true, nil, &common.PRVCoinID, nil)
	notMactchTokenIDTx.On("GetValidationEnv").Return(validValidationEnvironment)

	notMatchAmountCoin := &coinMocks.Coin{}
	notMatchAmountCoin.On("GetValue").Return(uint64(100))
	notMactchAmountTx := &metadataCommonMocks.Transaction{}
	notMactchAmountTx.On("GetTxBurnData").Return(true, notMatchAmountCoin, tokenHash, nil)
	notMactchAmountTx.On("GetValidationEnv").Return(validValidationEnvironment)

	validCoin := &coinMocks.Coin{}
	validCoin.On("GetValue").Return(uint64(1))

	normalTx := &metadataCommonMocks.Transaction{}
	normalTx.On("GetTxBurnData").Return(true, validCoin, tokenHash, nil)
	normalTx.On("GetType").Return(common.TxNormalType)
	normalTx.On("GetValidationEnv").Return(validValidationEnvironment)

	customTx := &metadataCommonMocks.Transaction{}
	customTx.On("GetTxBurnData").Return(true, validCoin, &common.PRVCoinID, nil)
	customTx.On("GetType").Return(common.TxCustomTokenPrivacyType)
	customTx.On("GetValidationEnv").Return(validValidationEnvironment)

	validTx := &metadataCommonMocks.Transaction{}
	validTx.On("GetTxBurnData").Return(true, validCoin, tokenHash, nil)
	validTx.On("GetType").Return(common.TxCustomTokenPrivacyType)
	validTx.On("GetValidationEnv").Return(validValidationEnvironment)

	type fields struct {
		MetadataBase metadataCommon.MetadataBase
		poolPairID   string
		nftID        string
		otaReceivers map[string]string
		shareAmount  uint64
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
			name: "Invalid chainRetriever",
			fields: fields{
				poolPairID: "",
			},
			args: args{
				chainRetriever: invalidChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid poolPairID",
			fields: fields{
				poolPairID: "",
			},
			args: args{
				chainRetriever: validChainRetriever,
			},
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
			args: args{
				chainRetriever: validChainRetriever,
			},
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
			args: args{
				chainRetriever: validChainRetriever,
			},
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
			args: args{
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid ota token 0",
			fields: fields{
				poolPairID: "123",
				nftID:      tokenHash.String(),
				otaReceivers: map[string]string{
					tokenHash.String(): validOTAReceiver0,
					token0ID.String():  "123",
				},
			},
			args: args{
				tx:             validTx,
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid ota token 1",
			fields: fields{
				poolPairID: "123",
				nftID:      tokenHash.String(),
				otaReceivers: map[string]string{
					tokenHash.String(): validOTAReceiver0,
					token0ID.String():  validOTAReceiver0,
					token1ID.String():  "123",
				},
			},
			args: args{
				tx:             validTx,
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid shareAmount",
			fields: fields{
				poolPairID: "123",
				nftID:      tokenHash.String(),
				otaReceivers: map[string]string{
					tokenHash.String(): validOTAReceiver0,
					token0ID.String():  validOTAReceiver0,
					token1ID.String():  validOTAReceiver0,
				},
				shareAmount: 0,
			},
			args: args{
				tx:             validTx,
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Tx is not burnt tx",
			fields: fields{
				poolPairID: "123",
				nftID:      tokenHash.String(),
				otaReceivers: map[string]string{
					tokenHash.String(): validOTAReceiver0,
					token0ID.String():  validOTAReceiver0,
					token1ID.String():  validOTAReceiver0,
				},
				shareAmount: 100,
			},
			args: args{
				tx:             notBurnTx,
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "tokenID not match with burn coin",
			fields: fields{
				poolPairID: "123",
				nftID:      tokenHash.String(),
				otaReceivers: map[string]string{
					tokenHash.String(): validOTAReceiver0,
					token0ID.String():  validOTAReceiver0,
					token1ID.String():  validOTAReceiver0,
				},
				shareAmount: 100,
			},
			args: args{
				tx:             notMactchTokenIDTx,
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "burn coin value is not 1",
			fields: fields{
				poolPairID: "123",
				nftID:      tokenHash.String(),
				otaReceivers: map[string]string{
					tokenHash.String(): validOTAReceiver0,
					token0ID.String():  validOTAReceiver0,
					token1ID.String():  validOTAReceiver0,
				},
				shareAmount: 100,
			},
			args: args{
				tx:             notMactchAmountTx,
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "type of tx not is custom privacy type",
			fields: fields{
				poolPairID: "123",
				nftID:      tokenHash.String(),
				otaReceivers: map[string]string{
					tokenHash.String(): validOTAReceiver0,
					token0ID.String():  validOTAReceiver0,
					token1ID.String():  validOTAReceiver0,
				},
				shareAmount: 100,
			},
			args: args{
				tx:             normalTx,
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "nftID == prv",
			fields: fields{
				poolPairID: "123",
				nftID:      tokenHash.String(),
				otaReceivers: map[string]string{
					tokenHash.String(): validOTAReceiver0,
					token0ID.String():  validOTAReceiver0,
					token1ID.String():  validOTAReceiver0,
				},
				shareAmount: 100,
			},
			args: args{
				tx:             customTx,
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Valid Input",
			fields: fields{
				poolPairID: "123",
				nftID:      tokenHash.String(),
				otaReceivers: map[string]string{
					tokenHash.String(): validOTAReceiver0,
					token0ID.String():  validOTAReceiver0,
					token1ID.String():  validOTAReceiver0,
				},
				shareAmount: 100,
			},
			args: args{
				tx:             validTx,
				chainRetriever: validChainRetriever,
			},
			want:    true,
			want1:   true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &WithdrawLiquidityRequest{
				MetadataBase: tt.fields.MetadataBase,
				poolPairID:   tt.fields.poolPairID,
				nftID:        tt.fields.nftID,
				shareAmount:  tt.fields.shareAmount,
				otaReceivers: tt.fields.otaReceivers,
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
		MetadataBase metadataCommon.MetadataBase
		poolPairID   string
		nftID        string
		otaReceivers map[string]string
		shareAmount  uint64
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
				MetadataBase: tt.fields.MetadataBase,
				poolPairID:   tt.fields.poolPairID,
				nftID:        tt.fields.nftID,
				otaReceivers: tt.fields.otaReceivers,
				shareAmount:  tt.fields.shareAmount,
			}
			if got := request.ValidateMetadataByItself(); got != tt.want {
				t.Errorf("WithdrawLiquidityRequest.ValidateMetadataByItself() = %v, want %v", got, tt.want)
			}
		})
	}
}
