package pdex

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/stretchr/testify/assert"
)

func TestPoolPairState_updateReserveAndCalculateShare(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)

	type fields struct {
		state  rawdbv2.Pdexv3PoolPair
		shares map[string]*Share
	}
	type args struct {
		token0ID     string
		token1ID     string
		token0Amount uint64
		token1Amount uint64
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               uint64
	}{
		{
			name: "token0ID < token1ID",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 100, 400,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:                  200,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 150, 600,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:                  200,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
			},
			args: args{
				token0ID:     token0ID.String(),
				token1ID:     token1ID.String(),
				token0Amount: 50,
				token1Amount: 200,
			},
			want: 100,
		},
		{
			name: "token0ID >= token1ID",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 100, 400,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:                  200,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 150, 600,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:                  200,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
			},
			args: args{
				token0ID:     token1ID.String(),
				token1ID:     token0ID.String(),
				token0Amount: 200,
				token1Amount: 50,
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PoolPairState{
				state:  tt.fields.state,
				shares: tt.fields.shares,
			}
			if got, _ := p.addReserveDataAndCalculateShare(tt.args.token0ID, tt.args.token1ID, tt.args.token0Amount, tt.args.token1Amount); got != tt.want {
				t.Errorf("PoolPairState.addReserveDataAndCalculateShare() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(tt.fieldsAfterProcess.state.Token0VirtualAmount(), p.state.Token0VirtualAmount()) {
				t.Errorf("token0VirtualAmount expect %v but get %v", tt.fieldsAfterProcess.state.Token0VirtualAmount(), p.state.Token0VirtualAmount())
				return
			}
			if !reflect.DeepEqual(tt.fieldsAfterProcess.state.Token1VirtualAmount(), p.state.Token1VirtualAmount()) {
				t.Errorf("token1VirtualAmount expect %v but get %v", tt.fieldsAfterProcess.state.Token1VirtualAmount(), p.state.Token1VirtualAmount())
				return
			}
			if !reflect.DeepEqual(tt.fieldsAfterProcess.state.Token0RealAmount(), p.state.Token0RealAmount()) {
				t.Errorf("token0RealAmount expect %v but get %v", tt.fieldsAfterProcess.state.Token0RealAmount(), p.state.Token0RealAmount())
				return
			}
			if !reflect.DeepEqual(tt.fieldsAfterProcess.state.Token1RealAmount(), p.state.Token1RealAmount()) {
				t.Errorf("token1RealAmount expect %v but get %v", tt.fieldsAfterProcess.state.Token1RealAmount(), p.state.Token1RealAmount())
				return
			}
		})
	}
}

func TestPoolPairState_calculateShareAmount(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)

	type fields struct {
		state  rawdbv2.Pdexv3PoolPair
		shares map[string]*Share
	}
	type args struct {
		amount0 uint64
		amount1 uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
	}{
		{
			name: "liquidityToken0.Uint64() >= liquidityToken1.Uint64()",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 100, 400,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:                  200,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
			},
			args: args{
				amount0: 50,
				amount1: 200,
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PoolPairState{
				state:  tt.fields.state,
				shares: tt.fields.shares,
			}
			if got := p.calculateShareAmount(tt.args.amount0, tt.args.amount1); got != tt.want {
				t.Errorf("PoolPairState.calculateShareAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoolPairState_updateReserveData(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)

	type fields struct {
		state  rawdbv2.Pdexv3PoolPair
		shares map[string]*Share
	}
	type args struct {
		amount0     uint64
		amount1     uint64
		shareAmount uint64
		operator    byte
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		wantErr            bool
	}{
		{
			name: "Base Amplifier",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 150, 600,
					big.NewInt(0).SetUint64(100),
					big.NewInt(0).SetUint64(400),
					metadataPdexv3.BaseAmplifier,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:                  200,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 150, 600,
					big.NewInt(0).SetUint64(150),
					big.NewInt(0).SetUint64(600),
					metadataPdexv3.BaseAmplifier,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:                  200,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
			},
			args: args{
				amount0:     50,
				amount1:     200,
				shareAmount: 100,
				operator:    addOperator,
			},
			wantErr: false,
		},
		{
			name: "Amplifier = 20000",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 150, 600,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:                  200,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 150, 600,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:                  200,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
			},
			args: args{
				amount0:     50,
				amount1:     200,
				shareAmount: 100,
				operator:    addOperator,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PoolPairState{
				state:  tt.fields.state,
				shares: tt.fields.shares,
			}
			err := p.updateReserveData(tt.args.amount0, tt.args.amount1, tt.args.shareAmount, tt.args.operator)
			if (err != nil) != tt.wantErr {
				t.Errorf("PoolPairState.deductShare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.fieldsAfterProcess.state.Token0VirtualAmount(), p.state.Token0VirtualAmount()) {
				t.Errorf("token0VirtualAmount expect %v but get %v", tt.fieldsAfterProcess.state.Token0VirtualAmount(), p.state.Token0VirtualAmount())
				return
			}
			if !reflect.DeepEqual(tt.fieldsAfterProcess.state.Token1VirtualAmount(), p.state.Token1VirtualAmount()) {
				t.Errorf("token1VirtualAmount expect %v but get %v", tt.fieldsAfterProcess.state.Token1VirtualAmount(), p.state.Token1VirtualAmount())
				return
			}
		})
	}
}

func TestPoolPairState_deductShare(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)

	type fields struct {
		state     rawdbv2.Pdexv3PoolPair
		shares    map[string]*Share
		orderbook Orderbook
	}
	type args struct {
		nftID        string
		shareAmount  uint64
		beaconHeight uint64
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               uint64
		want1              uint64
		want2              uint64
		wantErr            bool
	}{
		{
			name: "BaseAmplifier",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					300, 150, 600,
					big.NewInt(0).SetUint64(150),
					big.NewInt(0).SetUint64(600),
					metadataPdexv3.BaseAmplifier,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  300,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
				orderbook: Orderbook{},
			},
			args: args{
				nftID:        nftID,
				shareAmount:  100,
				beaconHeight: 20,
			},
			want:  50,
			want1: 200,
			want2: 100,
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 100, 400,
					big.NewInt(0).SetUint64(100),
					big.NewInt(0).SetUint64(400),
					metadataPdexv3.BaseAmplifier,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  200,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 20,
					},
				},
				orderbook: Orderbook{},
			},
			wantErr: false,
		},
		{
			name: "Not BaseAmplifier",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					300, 150, 600,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  300,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
				orderbook: Orderbook{},
			},
			args: args{
				nftID:        nftID,
				shareAmount:  100,
				beaconHeight: 20,
			},
			want:  50,
			want1: 200,
			want2: 100,
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 100, 400,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  200,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 20,
					},
				},
				orderbook: Orderbook{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PoolPairState{
				state:     tt.fields.state,
				shares:    tt.fields.shares,
				orderbook: tt.fields.orderbook,
			}
			got, got1, got2, err := p.deductShare(tt.args.nftID, tt.args.shareAmount, tt.args.beaconHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("PoolPairState.deductShare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PoolPairState.deductShare() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("PoolPairState.deductShare() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("PoolPairState.deductShare() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(p.state, tt.fieldsAfterProcess.state) {
				t.Errorf("fieldsAfterProcess.state got = %v, want %v", p.state, tt.fieldsAfterProcess.state)
			}
			if !reflect.DeepEqual(p.shares, tt.fieldsAfterProcess.shares) {
				t.Errorf("fieldsAfterProcess.state got = %v, want %v", p.shares, tt.fieldsAfterProcess.shares)
			}
			if !reflect.DeepEqual(p.orderbook, tt.fieldsAfterProcess.orderbook) {
				t.Errorf("fieldsAfterProcess.state got = %v, want %v", p.orderbook, tt.fieldsAfterProcess.orderbook)
			}
		})
	}
}

func TestPoolPairState_deductReserveData(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)

	type fields struct {
		state     rawdbv2.Pdexv3PoolPair
		shares    map[string]*Share
		orderbook Orderbook
	}
	type args struct {
		amount0     uint64
		amount1     uint64
		shareAmount uint64
		operator    byte
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		wantErr            bool
	}{
		{
			name: "BaseAmplifier",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					300, 150, 600,
					big.NewInt(0).SetUint64(150),
					big.NewInt(0).SetUint64(600),
					metadataPdexv3.BaseAmplifier,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  300,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
				orderbook: Orderbook{},
			},
			args: args{
				amount0:     50,
				amount1:     200,
				shareAmount: 100,
				operator:    subOperator,
			},
			wantErr: false,
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					300, 100, 400,
					big.NewInt(0).SetUint64(100),
					big.NewInt(0).SetUint64(400),
					metadataPdexv3.BaseAmplifier,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  300,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
				orderbook: Orderbook{},
			},
		},
		{
			name: "Not BaseAmplifier",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					300, 150, 600,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  300,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
				orderbook: Orderbook{},
			},
			args: args{
				amount0:     50,
				amount1:     200,
				shareAmount: 100,
				operator:    subOperator,
			},
			wantErr: false,
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					300, 100, 400,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  300,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
				orderbook: Orderbook{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PoolPairState{
				state:     tt.fields.state,
				shares:    tt.fields.shares,
				orderbook: tt.fields.orderbook,
			}
			if err := p.updateReserveData(
				tt.args.amount0, tt.args.amount1, tt.args.shareAmount, tt.args.operator,
			); (err != nil) != tt.wantErr {
				t.Errorf("PoolPairState.deductReserveData() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(p.state, tt.fieldsAfterProcess.state) {
				t.Errorf("fieldsAfterProcess.state got = %v, expect %v", p.state, tt.fieldsAfterProcess.state)
			}
			if !reflect.DeepEqual(p.shares, tt.fieldsAfterProcess.shares) {
				t.Errorf("fieldsAfterProcess.shares got = %v, expect %v", p.shares, tt.fieldsAfterProcess.shares)
			}
			if !reflect.DeepEqual(p.orderbook, tt.fieldsAfterProcess.orderbook) {
				t.Errorf("fieldsAfterProcess.orderbook got = %v, expect %v", p.orderbook, tt.fieldsAfterProcess.orderbook)
			}
		})
	}
}

func TestPoolPairState_updateSingleTokenAmount(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)

	type fields struct {
		state     rawdbv2.Pdexv3PoolPair
		shares    map[string]*Share
		orderbook Orderbook
	}
	type args struct {
		tokenID     common.Hash
		amount      uint64
		shareAmount uint64
		operator    byte
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		wantErr            bool
	}{
		{
			name: "Sub operator - token 0",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					300, 150, 600,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  300,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
				orderbook: Orderbook{},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					300, 100, 600,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  300,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
				orderbook: Orderbook{},
			},
			args: args{
				tokenID:     *token0ID,
				amount:      50,
				shareAmount: 100,
				operator:    subOperator,
			},
			wantErr: false,
		},
		{
			name: "Sub operator - token 1",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					300, 150, 600,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  300,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
				orderbook: Orderbook{},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					300, 150, 400,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  300,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
				orderbook: Orderbook{},
			},
			args: args{
				tokenID:     *token1ID,
				amount:      200,
				shareAmount: 100,
				operator:    subOperator,
			},
			wantErr: false,
		},
		{
			name: "Add operator - token 0",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 100, 400,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  200,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
				orderbook: Orderbook{},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 150, 400,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  200,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
				orderbook: Orderbook{},
			},
			args: args{
				tokenID:     *token0ID,
				amount:      50,
				shareAmount: 100,
				operator:    addOperator,
			},
			wantErr: false,
		},
		{
			name: "Add operator - token 1",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 100, 400,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  200,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
				orderbook: Orderbook{},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 100, 600,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                  200,
						tradingFees:             map[string]uint64{},
						lastUpdatedBeaconHeight: 10,
					},
				},
				orderbook: Orderbook{},
			},
			args: args{
				tokenID:     *token1ID,
				amount:      200,
				shareAmount: 100,
				operator:    addOperator,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PoolPairState{
				state:     tt.fields.state,
				shares:    tt.fields.shares,
				orderbook: tt.fields.orderbook,
			}
			if err := p.updateSingleTokenAmount(tt.args.tokenID, tt.args.amount, tt.args.shareAmount, tt.args.operator); (err != nil) != tt.wantErr {
				t.Errorf("PoolPairState.updateSingleTokenAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(p.state, tt.fieldsAfterProcess.state) {
				t.Errorf("fieldsAfterProcess got = %v, want %v", p.state, tt.fieldsAfterProcess.state)
			}
			if !reflect.DeepEqual(p.shares, tt.fieldsAfterProcess.shares) {
				t.Errorf("fieldsAfterProcess got = %v, want %v", p.shares, tt.fieldsAfterProcess.shares)
			}
			if !reflect.DeepEqual(p.orderbook, tt.fieldsAfterProcess.orderbook) {
				t.Errorf("fieldsAfterProcess got = %v, want %v", p.orderbook, tt.fieldsAfterProcess.orderbook)
			}
		})
	}
}
