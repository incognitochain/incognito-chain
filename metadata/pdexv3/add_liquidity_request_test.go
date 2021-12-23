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
	"github.com/stretchr/testify/mock"
)

func TestAddLiquidity_ValidateSanityData(t *testing.T) {
	common.MaxShardNumber = 8
	tokenHash, err := common.Hash{}.NewHashFromStr("123123")
	assert.Nil(t, err)

	invalidChainRetriever := &metadataCommonMocks.ChainRetriever{}
	invalidChainRetriever.On("IsAfterPdexv3CheckPoint", mock.AnythingOfType("uint64")).Return(false)
	validChainRetriever := &metadataCommonMocks.ChainRetriever{}
	validChainRetriever.On("IsAfterPdexv3CheckPoint", mock.AnythingOfType("uint64")).Return(true)

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

	invalidOtaReceiverShardIDTx := &metadataCommonMocks.Transaction{}
	invalidOtaReceiverShardIDTx.On("GetTxBurnData").Return(true, validCoin, tokenHash, nil)
	invalidOtaReceiverShardIDTx.On("GetType").Return(common.TxCustomTokenPrivacyType)
	invalidValidationEnvironment := &metadataCommonMocks.ValidationEnviroment{}
	invalidValidationEnvironment.On("ShardID").Return(0)
	invalidOtaReceiverShardIDTx.On("GetValidationEnv").Return(invalidValidationEnvironment)

	validTx := &metadataCommonMocks.Transaction{}
	validTx.On("GetTxBurnData").Return(true, validCoin, tokenHash, nil)
	validTx.On("GetType").Return(common.TxCustomTokenPrivacyType)
	validValidationEnvironment := &metadataCommonMocks.ValidationEnviroment{}
	validValidationEnvironment.On("ShardID").Return(1)
	validTx.On("GetValidationEnv").Return(validValidationEnvironment)

	type fields struct {
		poolPairID   string
		pairHash     string
		otaReceiver  string
		tokenID      string
		AccessOption AccessOption
		tokenAmount  uint64
		amplifier    uint
		MetadataBase metadataCommon.MetadataBase
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
				pairHash: "",
			},
			args: args{
				chainRetriever: invalidChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid PairHash",
			fields: fields{
				pairHash: "",
			},
			args: args{
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid tokenID",
			fields: fields{
				pairHash: "pair hash",
				tokenID:  "asdb",
			},
			args: args{
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Empty tokenID",
			fields: fields{
				pairHash: "pair hash",
				tokenID:  "",
			},
			args: args{
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid otaReceive",
			fields: fields{
				pairHash: "pair hash",
				tokenID:  tokenHash.String(),
			},
			args: args{
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid RefundAddress",
			fields: fields{
				pairHash:    "pair hash",
				tokenID:     tokenHash.String(),
				otaReceiver: validOTAReceiver0,
			},
			args: args{
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid amplifier",
			fields: fields{
				pairHash:    "pair hash",
				tokenID:     tokenHash.String(),
				otaReceiver: validOTAReceiver0,
			},
			args: args{
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Tx is not burn tx",
			fields: fields{
				pairHash:    "pair hash",
				tokenID:     tokenHash.String(),
				otaReceiver: validOTAReceiver0,
				amplifier:   10000,
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
			name: "tokenID not match with burnCoin",
			fields: fields{
				pairHash:    "pair hash",
				tokenID:     tokenHash.String(),
				otaReceiver: validOTAReceiver0,
				amplifier:   10000,
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
			name: "Token amount = 0",
			fields: fields{
				pairHash:    "pair hash",
				tokenID:     tokenHash.String(),
				otaReceiver: validOTAReceiver0,
				amplifier:   10000,
			},
			args: args{
				tx:             notMactchAmountTx0,
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Contributed amount is not match with burn amount",
			fields: fields{
				pairHash:    "pair hash",
				tokenID:     tokenHash.String(),
				otaReceiver: validOTAReceiver0,
				amplifier:   10000,
				tokenAmount: 200,
			},
			args: args{
				tx:             notMactchAmountTx1,
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Normal tx && tokenID != prv",
			fields: fields{
				pairHash:    "pair hash",
				tokenID:     tokenHash.String(),
				otaReceiver: validOTAReceiver0,
				amplifier:   10000,
				tokenAmount: 200,
			},
			args: args{
				tx:             invalidNormalTx,
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Custom token tx && tokenID == prv",
			fields: fields{
				pairHash:    "pair hash",
				tokenID:     common.PRVCoinID.String(),
				otaReceiver: validOTAReceiver0,
				amplifier:   10000,
				tokenAmount: 200,
			},
			args: args{
				tx:             invalidCustomTx,
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Valid Input",
			fields: fields{
				pairHash:    "pair hash",
				tokenID:     tokenHash.String(),
				otaReceiver: validOTAReceiver0,
				amplifier:   10000,
				tokenAmount: 200,
				AccessOption: AccessOption{
					NftID: tokenHash,
				},
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
			al := &AddLiquidityRequest{
				poolPairID:   tt.fields.poolPairID,
				pairHash:     tt.fields.pairHash,
				otaReceiver:  tt.fields.otaReceiver,
				tokenID:      tt.fields.tokenID,
				AccessOption: tt.fields.AccessOption,
				tokenAmount:  tt.fields.tokenAmount,
				amplifier:    tt.fields.amplifier,
				MetadataBase: tt.fields.MetadataBase,
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
		poolPairID   string
		pairHash     string
		otaReceiver  string
		tokenID      string
		tokenAmount  uint64
		amplifier    uint
		MetadataBase metadataCommon.MetadataBase
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
					Type: metadataCommon.Pdexv3AddLiquidityRequestMeta,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := &AddLiquidityRequest{
				poolPairID:   tt.fields.poolPairID,
				pairHash:     tt.fields.pairHash,
				otaReceiver:  tt.fields.otaReceiver,
				tokenID:      tt.fields.tokenID,
				tokenAmount:  tt.fields.tokenAmount,
				amplifier:    tt.fields.amplifier,
				MetadataBase: tt.fields.MetadataBase,
			}
			if got := al.ValidateMetadataByItself(); got != tt.want {
				t.Errorf("AddLiquidity.ValidateMetadataByItself() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddLiquidityRequest_ValidateTxWithBlockChain(t *testing.T) {
	type fields struct {
		poolPairID   string
		pairHash     string
		otaReceiver  string
		otaReceivers map[common.Hash]privacy.OTAReceiver
		tokenID      string
		AccessOption AccessOption
		tokenAmount  uint64
		amplifier    uint
		MetadataBase metadataCommon.MetadataBase
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
			request := &AddLiquidityRequest{
				poolPairID:   tt.fields.poolPairID,
				pairHash:     tt.fields.pairHash,
				otaReceiver:  tt.fields.otaReceiver,
				otaReceivers: tt.fields.otaReceivers,
				tokenID:      tt.fields.tokenID,
				AccessOption: tt.fields.AccessOption,
				tokenAmount:  tt.fields.tokenAmount,
				amplifier:    tt.fields.amplifier,
				MetadataBase: tt.fields.MetadataBase,
			}
			got, err := request.ValidateTxWithBlockChain(tt.args.tx, tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.shardID, tt.args.transactionStateDB)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddLiquidityRequest.ValidateTxWithBlockChain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AddLiquidityRequest.ValidateTxWithBlockChain() = %v, want %v", got, tt.want)
			}
		})
	}
}
