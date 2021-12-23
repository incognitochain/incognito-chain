package pdexv3

import (
	"errors"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
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
		AccessOption AccessOption
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
			name: "Empty NftID",
			fields: fields{
				poolPairID: "123",
				AccessOption: AccessOption{
					NftID: &common.Hash{},
				},
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
				AccessOption: AccessOption{
					NftID: tokenHash,
				},
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
				AccessOption: AccessOption{
					NftID: tokenHash,
				},
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
				AccessOption: AccessOption{
					NftID: tokenHash,
				},
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
				AccessOption: AccessOption{
					NftID: tokenHash,
				},
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
			name: "Valid Input",
			fields: fields{
				poolPairID: "123",
				AccessOption: AccessOption{
					NftID: tokenHash,
				},
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
				AccessOption: tt.fields.AccessOption,
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
		AccessOption AccessOption
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
				AccessOption: tt.fields.AccessOption,
				otaReceivers: tt.fields.otaReceivers,
				shareAmount:  tt.fields.shareAmount,
			}
			if got := request.ValidateMetadataByItself(); got != tt.want {
				t.Errorf("WithdrawLiquidityRequest.ValidateMetadataByItself() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithdrawLiquidityRequest_ValidateTxWithBlockChain(t *testing.T) {
	type fields struct {
		MetadataBase metadataCommon.MetadataBase
		poolPairID   string
		AccessOption AccessOption
		otaReceivers map[string]string
		shareAmount  uint64
	}
	type args struct {
		tx                  metadataCommon.Transaction
		chainRetriever      metadataCommon.ChainRetriever
		shardViewRetriever  metadataCommon.ShardViewRetriever
		beaconViewRetriever metadataCommon.BeaconViewRetriever
		shardID             byte
		transactionStateDB  *statedb.StateDB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &WithdrawLiquidityRequest{
				MetadataBase: tt.fields.MetadataBase,
				poolPairID:   tt.fields.poolPairID,
				AccessOption: tt.fields.AccessOption,
				otaReceivers: tt.fields.otaReceivers,
				shareAmount:  tt.fields.shareAmount,
			}
			got, err := request.ValidateTxWithBlockChain(tt.args.tx, tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.shardID, tt.args.transactionStateDB)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithdrawLiquidityRequest.ValidateTxWithBlockChain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("WithdrawLiquidityRequest.ValidateTxWithBlockChain() = %v, want %v", got, tt.want)
			}
		})
	}
}
