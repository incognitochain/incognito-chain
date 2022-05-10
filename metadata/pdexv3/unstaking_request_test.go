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

func TestUnstakingRequest_ValidateSanityData(t *testing.T) {
	common.MaxShardNumber = 8
	nftID, err := common.Hash{}.NewHashFromStr("12345678")
	assert.Nil(t, err)
	tokenHash, err := common.Hash{}.NewHashFromStr("123123")
	assert.Nil(t, err)

	invalidChainRetriever := &metadataCommonMocks.ChainRetriever{}
	invalidChainRetriever.On("IsAfterPdexv3CheckPoint", mock.AnythingOfType("uint64")).Return(false)
	validChainRetriever := &metadataCommonMocks.ChainRetriever{}
	validChainRetriever.On("IsAfterPdexv3CheckPoint", mock.AnythingOfType("uint64")).Return(true)

	validValidationEnvironment := &metadataCommonMocks.ValidationEnviroment{}
	validValidationEnvironment.On("ShardID").Return(1)

	invalidValidationEnvironment := &metadataCommonMocks.ValidationEnviroment{}
	invalidValidationEnvironment.On("ShardID").Return(3)

	invalidShardIDTx := &metadataCommonMocks.Transaction{}
	invalidShardIDTx.On("GetValidationEnv").Return(invalidValidationEnvironment)

	notBurnTx := &metadataCommonMocks.Transaction{}
	notBurnTx.On("GetTxBurnData").Return(false, nil, nil, errors.New("Not tx burn"))
	notBurnTx.On("GetValidationEnv").Return(validValidationEnvironment)

	notMactchTokenIDTx := &metadataCommonMocks.Transaction{}
	notMactchTokenIDTx.On("GetTxBurnData").Return(true, nil, tokenHash, nil)
	notMactchTokenIDTx.On("GetValidationEnv").Return(validValidationEnvironment)

	notMatchAmountCoin := &coinMocks.Coin{}
	notMatchAmountCoin.On("GetValue").Return(uint64(100))
	notMactchAmountTx := &metadataCommonMocks.Transaction{}
	notMactchAmountTx.On("GetTxBurnData").Return(true, notMatchAmountCoin, nftID, nil)
	notMactchAmountTx.On("GetValidationEnv").Return(validValidationEnvironment)

	validCoin := &coinMocks.Coin{}
	validCoin.On("GetValue").Return(uint64(1))

	normalTx := &metadataCommonMocks.Transaction{}
	normalTx.On("GetTxBurnData").Return(true, validCoin, nftID, nil)
	normalTx.On("GetType").Return(common.TxNormalType)
	normalTx.On("GetValidationEnv").Return(validValidationEnvironment)

	customTx := &metadataCommonMocks.Transaction{}
	customTx.On("GetTxBurnData").Return(true, validCoin, &common.PRVCoinID, nil)
	customTx.On("GetType").Return(common.TxCustomTokenPrivacyType)
	customTx.On("GetValidationEnv").Return(validValidationEnvironment)

	validTx := &metadataCommonMocks.Transaction{}
	validTx.On("GetTxBurnData").Return(true, validCoin, nftID, nil)
	validTx.On("GetType").Return(common.TxCustomTokenPrivacyType)
	validTx.On("GetValidationEnv").Return(validValidationEnvironment)

	type fields struct {
		MetadataBase    metadataCommon.MetadataBase
		stakingPoolID   string
		otaReceivers    map[string]string
		AccessOption    AccessOption
		unstakingAmount uint64
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
			name: "invalid chainRetriever",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingRequestMeta,
				},
				stakingPoolID: "abc",
			},
			args: args{
				chainRetriever: invalidChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "stakingPoolID is invalid",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingRequestMeta,
				},
				stakingPoolID: "abc",
			},
			args: args{

				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "stakingPoolID is empty",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingRequestMeta,
				},
				stakingPoolID: common.Hash{}.String(),
			},
			args: args{

				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "nftID is empty",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingRequestMeta,
				},
				stakingPoolID: common.PRVIDStr,
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
			name: "otaReceivers are invalid",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingRequestMeta,
				},
				stakingPoolID: common.PRVIDStr,
				AccessOption: AccessOption{
					NftID: nftID,
				},
				otaReceivers: map[string]string{
					common.PRVIDStr: "abcd",
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
			name: "otaReceivers' shardid is invalid",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingRequestMeta,
				},
				stakingPoolID: common.PRVIDStr,
				AccessOption: AccessOption{
					NftID: nftID,
				},
				otaReceivers: map[string]string{
					common.PRVIDStr: validOTAReceiver0,
					nftID.String():  validOTAReceiver0,
				},
			},
			args: args{
				tx:             invalidShardIDTx,
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "unstakingAmount == 0",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingRequestMeta,
				},
				stakingPoolID: common.PRVIDStr,
				AccessOption: AccessOption{
					NftID: nftID,
				},
				otaReceivers: map[string]string{
					common.PRVIDStr: validOTAReceiver0,
					nftID.String():  validOTAReceiver0,
				},
				unstakingAmount: 0,
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
			name: "burnAmount is not 1",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingRequestMeta,
				},
				stakingPoolID: common.PRVIDStr,
				AccessOption: AccessOption{
					NftID: nftID,
				},
				otaReceivers: map[string]string{
					common.PRVIDStr: validOTAReceiver0,
					nftID.String():  validOTAReceiver0,
				},
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
			name: "Tx type is not customTokenPrivacy",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingRequestMeta,
				},
				stakingPoolID: common.PRVIDStr,
				AccessOption: AccessOption{
					NftID: nftID,
				},
				otaReceivers: map[string]string{
					common.PRVIDStr: validOTAReceiver0,
					nftID.String():  validOTAReceiver0,
				},
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
			name: "nftID is PRVID",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingRequestMeta,
				},
				stakingPoolID: common.PRVIDStr,
				AccessOption: AccessOption{
					NftID: nftID,
				},
				otaReceivers: map[string]string{
					common.PRVIDStr: validOTAReceiver0,
					nftID.String():  validOTAReceiver0,
				},
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
			name: "Valid input",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3UnstakingRequestMeta,
				},
				stakingPoolID: common.PRVIDStr,
				AccessOption: AccessOption{
					NftID: nftID,
				},
				otaReceivers: map[string]string{
					common.PRVIDStr: validOTAReceiver0,
					nftID.String():  validOTAReceiver0,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &UnstakingRequest{
				MetadataBase:    tt.fields.MetadataBase,
				stakingPoolID:   tt.fields.stakingPoolID,
				otaReceivers:    tt.fields.otaReceivers,
				AccessOption:    tt.fields.AccessOption,
				unstakingAmount: tt.fields.unstakingAmount,
			}
			got, got1, err := request.ValidateSanityData(tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.beaconHeight, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnstakingRequest.ValidateSanityData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UnstakingRequest.ValidateSanityData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("UnstakingRequest.ValidateSanityData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestUnstakingRequest_ValidateMetadataByItself(t *testing.T) {
	type fields struct {
		MetadataBase    metadataCommon.MetadataBase
		stakingPoolID   string
		otaReceivers    map[string]string
		AccessOption    AccessOption
		unstakingAmount uint64
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
					Type: metadataCommon.Pdexv3UnstakingRequestMeta,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &UnstakingRequest{
				MetadataBase:    tt.fields.MetadataBase,
				stakingPoolID:   tt.fields.stakingPoolID,
				otaReceivers:    tt.fields.otaReceivers,
				AccessOption:    tt.fields.AccessOption,
				unstakingAmount: tt.fields.unstakingAmount,
			}
			if got := request.ValidateMetadataByItself(); got != tt.want {
				t.Errorf("UnstakingRequest.ValidateMetadataByItself() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnstakingRequest_ValidateTxWithBlockChain(t *testing.T) {
	type fields struct {
		MetadataBase    metadataCommon.MetadataBase
		stakingPoolID   string
		otaReceivers    map[string]string
		AccessOption    AccessOption
		unstakingAmount uint64
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
			request := &UnstakingRequest{
				MetadataBase:    tt.fields.MetadataBase,
				stakingPoolID:   tt.fields.stakingPoolID,
				otaReceivers:    tt.fields.otaReceivers,
				AccessOption:    tt.fields.AccessOption,
				unstakingAmount: tt.fields.unstakingAmount,
			}
			got, err := request.ValidateTxWithBlockChain(tt.args.tx, tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.shardID, tt.args.transactionStateDB)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnstakingRequest.ValidateTxWithBlockChain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UnstakingRequest.ValidateTxWithBlockChain() = %v, want %v", got, tt.want)
			}
		})
	}
}
