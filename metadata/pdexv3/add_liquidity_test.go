package pdexv3

import (
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataCommonMocks "github.com/incognitochain/incognito-chain/metadata/common/mocks"
	coinMocks "github.com/incognitochain/incognito-chain/privacy/coin/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAddLiquidity_ValidateSanityData(t *testing.T) {

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
		poolPairID     string
		pairHash       string
		receiveAddress string
		refundAddress  string
		tokenID        string
		tokenAmount    uint64
		amplifier      uint
		MetadataBase   metadataCommon.MetadataBase
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
				pairHash: "",
			},
			args:    args{},
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
			args:    args{},
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
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid ReceiveAddress",
			fields: fields{
				pairHash: "pair hash",
				tokenID:  tokenHash.String(),
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid RefundAddress",
			fields: fields{
				pairHash:       "pair hash",
				tokenID:        tokenHash.String(),
				receiveAddress: validOTAReceiver0,
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Invalid amplifier",
			fields: fields{
				pairHash:       "pair hash",
				tokenID:        tokenHash.String(),
				receiveAddress: validOTAReceiver0,
				refundAddress:  validOTAReceiver1,
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Tx is not burn tx",
			fields: fields{
				pairHash:       "pair hash",
				tokenID:        tokenHash.String(),
				receiveAddress: validOTAReceiver0,
				refundAddress:  validOTAReceiver1,
				amplifier:      10000,
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
				pairHash:       "pair hash",
				tokenID:        tokenHash.String(),
				receiveAddress: validOTAReceiver0,
				refundAddress:  validOTAReceiver1,
				amplifier:      10000,
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
				pairHash:       "pair hash",
				tokenID:        tokenHash.String(),
				receiveAddress: validOTAReceiver0,
				refundAddress:  validOTAReceiver1,
				amplifier:      10000,
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
				pairHash:       "pair hash",
				tokenID:        tokenHash.String(),
				receiveAddress: validOTAReceiver0,
				refundAddress:  validOTAReceiver1,
				amplifier:      10000,
				tokenAmount:    200,
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
				pairHash:       "pair hash",
				tokenID:        tokenHash.String(),
				receiveAddress: validOTAReceiver0,
				refundAddress:  validOTAReceiver1,
				amplifier:      10000,
				tokenAmount:    200,
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
				pairHash:       "pair hash",
				tokenID:        common.PRVCoinID.String(),
				receiveAddress: validOTAReceiver0,
				refundAddress:  validOTAReceiver1,
				amplifier:      10000,
				tokenAmount:    200,
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
				pairHash:       "pair hash",
				tokenID:        tokenHash.String(),
				receiveAddress: validOTAReceiver0,
				refundAddress:  validOTAReceiver1,
				amplifier:      10000,
				tokenAmount:    200,
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
				poolPairID:     tt.fields.poolPairID,
				pairHash:       tt.fields.pairHash,
				receiveAddress: tt.fields.receiveAddress,
				refundAddress:  tt.fields.refundAddress,
				tokenID:        tt.fields.tokenID,
				tokenAmount:    tt.fields.tokenAmount,
				amplifier:      tt.fields.amplifier,
				MetadataBase:   tt.fields.MetadataBase,
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
		poolPairID     string
		pairHash       string
		receiveAddress string
		refundAddress  string
		tokenID        string
		tokenAmount    uint64
		amplifier      uint
		MetadataBase   metadataCommon.MetadataBase
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
					Type: metadataCommon.PDexv3TradeResponseMeta,
				},
			},
			want: false,
		},
		{
			name: "Valid Input",
			fields: fields{
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.Pdexv3AddLiquidityMeta,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := &AddLiquidity{
				poolPairID:     tt.fields.poolPairID,
				pairHash:       tt.fields.pairHash,
				receiveAddress: tt.fields.receiveAddress,
				refundAddress:  tt.fields.refundAddress,
				tokenID:        tt.fields.tokenID,
				tokenAmount:    tt.fields.tokenAmount,
				amplifier:      tt.fields.amplifier,
				MetadataBase:   tt.fields.MetadataBase,
			}
			if got := al.ValidateMetadataByItself(); got != tt.want {
				t.Errorf("AddLiquidity.ValidateMetadataByItself() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddLiquidity_StringSlice(t *testing.T) {
	type fields struct {
		poolPairID     string
		pairHash       string
		receiveAddress string
		refundAddress  string
		tokenID        string
		tokenAmount    uint64
		amplifier      uint
		MetadataBase   metadataCommon.MetadataBase
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "Valid Input",
			fields: fields{
				poolPairID:     "pool_pair_id",
				pairHash:       "pair_hash",
				receiveAddress: validOTAReceiver0,
				refundAddress:  validOTAReceiver1,
				tokenID:        "token_id",
				tokenAmount:    300,
				amplifier:      10000,
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.PDexV3AddLiquidityMeta,
				},
			},
			want: []string{
				strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
				"pool_pair_id",
				"pair_hash",
				validOTAReceiver0,
				validOTAReceiver1,
				"token_id",
				"300",
				"10000",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := &AddLiquidity{
				poolPairID:     tt.fields.poolPairID,
				pairHash:       tt.fields.pairHash,
				receiveAddress: tt.fields.receiveAddress,
				refundAddress:  tt.fields.refundAddress,
				tokenID:        tt.fields.tokenID,
				tokenAmount:    tt.fields.tokenAmount,
				amplifier:      tt.fields.amplifier,
				MetadataBase:   tt.fields.MetadataBase,
			}
			if got := al.StringSlice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddLiquidity.StringArr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddLiquidity_FromStringSlice(t *testing.T) {
	type fields struct {
		poolPairID     string
		pairHash       string
		receiveAddress string
		refundAddress  string
		tokenID        string
		tokenAmount    uint64
		amplifier      uint
		MetadataBase   metadataCommon.MetadataBase
	}
	type args struct {
		source []string
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		wantErr            bool
	}{
		{
			name:   "Invalid length",
			fields: fields{},
			args: args{
				source: []string{},
			},
			wantErr: true,
		},
		{
			name:   "Invalid metadata type",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddOrderRequestMeta),
					"pool_pair_id",
					"pair_hash",
					validOTAReceiver0,
					validOTAReceiver1,
					common.PRVCoinID.String(),
					"300",
					"10000",
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid status",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					"",
					"pool_pair_id",
					"pair_hash",
					validOTAReceiver0,
					validOTAReceiver1,
					common.PRVCoinID.String(),
					"300",
					"10000",
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid pair hash",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					"pool_pair_id",
					"",
					validOTAReceiver0,
					validOTAReceiver1,
					common.PRVCoinID.String(),
					"300",
					"10000",
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid receiveAddress",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					"pool_pair_id",
					"pair_hash",
					"receive_address",
					validOTAReceiver1,
					common.PRVCoinID.String(),
					"300",
					"10000",
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid refundAddress",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					"pool_pair_id",
					"pair_hash",
					validOTAReceiver0,
					"refund_address",
					common.PRVCoinID.String(),
					"300",
					"10000",
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid tokenID",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					"pool_pair_id",
					"pair_hash",
					validOTAReceiver0,
					validOTAReceiver1,
					"vzxvc",
					"300",
					"10000",
				},
			},
			wantErr: true,
		},
		{
			name:   "Empty tokenID",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					"pool_pair_id",
					"pair_hash",
					validOTAReceiver0,
					validOTAReceiver1,
					common.Hash{}.String(),
					"300",
					"10000",
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid token amount",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					"pool_pair_id",
					"pair_hash",
					validOTAReceiver0,
					validOTAReceiver1,
					common.PRVCoinID.String(),
					"token_amount",
					"10000",
				},
			},
			wantErr: true,
		},
		{
			name:   "Invalid amplifier",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					"pool_pair_id",
					"pair_hash",
					validOTAReceiver0,
					validOTAReceiver1,
					common.PRVCoinID.String(),
					"300",
					"amplifier",
				},
			},
			wantErr: true,
		},
		{
			name:   "Amplifier is smaller than default amplifier",
			fields: fields{},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					"pool_pair_id",
					"pair_hash",
					validOTAReceiver0,
					validOTAReceiver1,
					common.PRVCoinID.String(),
					"300",
					"900",
				},
			},
			wantErr: true,
		},
		{
			name:   "Valid Input",
			fields: fields{},
			fieldsAfterProcess: fields{
				poolPairID:     "pool_pair_id",
				pairHash:       "pair_hash",
				receiveAddress: validOTAReceiver0,
				refundAddress:  validOTAReceiver1,
				tokenID:        common.PRVCoinID.String(),
				tokenAmount:    300,
				amplifier:      10000,
				MetadataBase: metadataCommon.MetadataBase{
					Type: metadataCommon.PDexV3AddLiquidityMeta,
				},
			},
			args: args{
				source: []string{
					strconv.Itoa(metadataCommon.PDexV3AddLiquidityMeta),
					"pool_pair_id",
					"pair_hash",
					validOTAReceiver0,
					validOTAReceiver1,
					common.PRVCoinID.String(),
					"300",
					"10000",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := &AddLiquidity{
				poolPairID:     tt.fields.poolPairID,
				pairHash:       tt.fields.pairHash,
				receiveAddress: tt.fields.receiveAddress,
				refundAddress:  tt.fields.refundAddress,
				tokenID:        tt.fields.tokenID,
				tokenAmount:    tt.fields.tokenAmount,
				amplifier:      tt.fields.amplifier,
				MetadataBase:   tt.fields.MetadataBase,
			}
			if err := al.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("AddLiquidity.FromString() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(al.poolPairID, tt.fieldsAfterProcess.poolPairID) {
				t.Errorf("poolPairID got = %v, want %v", al.poolPairID, tt.fieldsAfterProcess.poolPairID)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(al.pairHash, tt.fieldsAfterProcess.pairHash) {
				t.Errorf("pairHash got = %v, want %v", al.pairHash, tt.fieldsAfterProcess.pairHash)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(al.receiveAddress, tt.fieldsAfterProcess.receiveAddress) {
				t.Errorf("receiveAddress got = %v, want %v", al.receiveAddress, tt.fieldsAfterProcess.receiveAddress)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(al.refundAddress, tt.fieldsAfterProcess.refundAddress) {
				t.Errorf("refundAddress got = %v, want %v", al.refundAddress, tt.fieldsAfterProcess.refundAddress)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(al.tokenID, tt.fieldsAfterProcess.tokenID) {
				t.Errorf("tokenID got = %v, want %v", al.tokenID, tt.fieldsAfterProcess.tokenID)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(al.tokenAmount, tt.fieldsAfterProcess.tokenAmount) {
				t.Errorf("tokenAmount got = %v, want %v", al.tokenAmount, tt.fieldsAfterProcess.tokenAmount)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(al.amplifier, tt.fieldsAfterProcess.amplifier) {
				t.Errorf("amplifier got = %v, want %v", al.amplifier, tt.fieldsAfterProcess.amplifier)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(al.MetadataBase, tt.fieldsAfterProcess.MetadataBase) {
				t.Errorf("fieldsAfterProcess got = %v, want %v", al, tt.fieldsAfterProcess)
				return
			}
		})
	}
}
