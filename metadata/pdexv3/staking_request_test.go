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

func TestStakingRequest_ValidateSanityData(t *testing.T) {
	tokenHash, err := common.Hash{}.NewHashFromStr("123123")
	assert.Nil(t, err)
	nftHash, err := common.Hash{}.NewHashFromStr("123456")
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

	invalidTypeTx := &metadataCommonMocks.Transaction{}
	invalidTypeTx.On("GetTxBurnData").Return(true, validCoin, &common.PRVCoinID, nil)
	invalidTypeTx.On("GetType").Return(common.TxTokenConversionType)

	validTx := &metadataCommonMocks.Transaction{}
	validTx.On("GetTxBurnData").Return(true, validCoin, tokenHash, nil)
	validTx.On("GetType").Return(common.TxCustomTokenPrivacyType)

	type fields struct {
		MetadataBase metadataCommon.MetadataBase
		tokenID      string
		otaReceiver  string
		AccessOption AccessOption
		tokenAmount  uint64
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
				tokenID: "asdb",
			},
			args: args{
				chainRetriever: invalidChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid tokenID",
			fields: fields{
				tokenID: "asdb",
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
				tokenID: common.Hash{}.String(),
			},
			args: args{
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Not burn tx",
			fields: fields{
				tokenID: tokenHash.String(),
				AccessOption: AccessOption{
					NftID: nftHash,
				},
				otaReceiver: validOTAReceiver0,
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
			name: "Burnt Token != tokenID",
			fields: fields{
				tokenID: tokenHash.String(),
				AccessOption: AccessOption{
					NftID: nftHash,
				},
				otaReceiver: validOTAReceiver0,
			},
			args: args{
				tx: notMactchTokenIDTx,

				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Token amount = 0",
			fields: fields{
				tokenID: tokenHash.String(),
				AccessOption: AccessOption{
					NftID: nftHash,
				},
				otaReceiver: validOTAReceiver0,
			},
			args: args{
				tx: notMactchAmountTx0,

				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "requet.tokenAmount != burnCoin.GetValue()",
			fields: fields{
				tokenID: tokenHash.String(),
				AccessOption: AccessOption{
					NftID: nftHash,
				},
				otaReceiver: validOTAReceiver0,
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
			name: "normatl tx && tokenID != prv",
			fields: fields{
				tokenID: tokenHash.String(),
				AccessOption: AccessOption{
					NftID: nftHash,
				},
				otaReceiver: validOTAReceiver0,
			},
			args: args{
				tx: invalidNormalTx,

				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "custom token tx && tokenID == prv",
			fields: fields{
				tokenID: tokenHash.String(),
				AccessOption: AccessOption{
					NftID: nftHash,
				},
				otaReceiver: validOTAReceiver0,
			},
			args: args{
				tx: invalidCustomTx,

				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "invalid tx type",
			fields: fields{
				tokenID: tokenHash.String(),
				AccessOption: AccessOption{
					NftID: nftHash,
				},
				otaReceiver: validOTAReceiver0,
			},
			args: args{
				tx:             invalidTypeTx,
				chainRetriever: validChainRetriever,
			},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Valid input",
			fields: fields{
				tokenID: tokenHash.String(),
				AccessOption: AccessOption{
					NftID: nftHash,
				},
				otaReceiver: validOTAReceiver0,
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
			request := &StakingRequest{
				MetadataBase: tt.fields.MetadataBase,
				tokenID:      tt.fields.tokenID,
				otaReceiver:  tt.fields.otaReceiver,
				AccessOption: tt.fields.AccessOption,
				tokenAmount:  tt.fields.tokenAmount,
			}
			got, got1, err := request.ValidateSanityData(tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.beaconHeight, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("StakingRequest.ValidateSanityData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StakingRequest.ValidateSanityData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("StakingRequest.ValidateSanityData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestStakingRequest_ValidateMetadataByItself(t *testing.T) {
	type fields struct {
		MetadataBase metadataCommon.MetadataBase
		tokenID      string
		otaReceiver  string
		otaReceivers map[common.Hash]privacy.OTAReceiver // receive tokens
		AccessOption
		tokenAmount uint64
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
					Type: metadataCommon.Pdexv3StakingRequestMeta,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &StakingRequest{
				MetadataBase: tt.fields.MetadataBase,
				tokenID:      tt.fields.tokenID,
				otaReceiver:  tt.fields.otaReceiver,
				otaReceivers: tt.fields.otaReceivers,
				AccessOption: tt.fields.AccessOption,
				tokenAmount:  tt.fields.tokenAmount,
			}
			if got := request.ValidateMetadataByItself(); got != tt.want {
				t.Errorf("StakingRequest.ValidateMetadataByItself() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStakingRequest_ValidateTxWithBlockChain(t *testing.T) {
	initTestParam(t)
	nftID, err := common.Hash{}.NewHashFromStr("123456")
	assert.Nil(t, err)
	type fields struct {
		MetadataBase metadataCommon.MetadataBase
		tokenID      string
		otaReceiver  string
		otaReceivers map[common.Hash]privacy.OTAReceiver
		AccessOption AccessOption
		tokenAmount  uint64
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
		{
			name: "NftID and AccessID exist at same time",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3StakingRequestMeta,
				},
				AccessOption: AccessOption{
					NftID:    nftID,
					AccessID: nftID,
				},
			},
			args:    args{},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &StakingRequest{
				MetadataBase: tt.fields.MetadataBase,
				tokenID:      tt.fields.tokenID,
				otaReceiver:  tt.fields.otaReceiver,
				otaReceivers: tt.fields.otaReceivers,
				AccessOption: tt.fields.AccessOption,
				tokenAmount:  tt.fields.tokenAmount,
			}
			got, err := request.ValidateTxWithBlockChain(tt.args.tx, tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.shardID, tt.args.transactionStateDB)
			if (err != nil) != tt.wantErr {
				t.Errorf("StakingRequest.ValidateTxWithBlockChain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StakingRequest.ValidateTxWithBlockChain() = %v, want %v", got, tt.want)
			}
		})
	}
}
