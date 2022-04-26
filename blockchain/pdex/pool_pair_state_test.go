package pdex

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/stretchr/testify/assert"
)

func TestPoolPairState_updateReserveAndCalculateShare(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)

	type fields struct {
		state           rawdbv2.Pdexv3PoolPair
		shares          map[string]*Share
		lpFeesPerShare  map[common.Hash]*big.Int
		protocolFees    map[common.Hash]uint64
		stakingPoolFees map[common.Hash]uint64
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
					200, 0, 100, 400,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:             200,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 0, 150, 600,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:             200,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
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
					200, 0, 100, 400,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:             200,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 0, 150, 600,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:             200,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
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
				state:           tt.fields.state,
				shares:          tt.fields.shares,
				lpFeesPerShare:  tt.fields.lpFeesPerShare,
				protocolFees:    tt.fields.protocolFees,
				stakingPoolFees: tt.fields.stakingPoolFees,
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
		state           rawdbv2.Pdexv3PoolPair
		shares          map[string]*Share
		lpFeesPerShare  map[common.Hash]*big.Int
		protocolFees    map[common.Hash]uint64
		stakingPoolFees map[common.Hash]uint64
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
					200, 0, 100, 400,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:             200,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
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
				state:           tt.fields.state,
				shares:          tt.fields.shares,
				lpFeesPerShare:  tt.fields.lpFeesPerShare,
				protocolFees:    tt.fields.protocolFees,
				stakingPoolFees: tt.fields.stakingPoolFees,
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
		state           rawdbv2.Pdexv3PoolPair
		shares          map[string]*Share
		lpFeesPerShare  map[common.Hash]*big.Int
		protocolFees    map[common.Hash]uint64
		stakingPoolFees map[common.Hash]uint64
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
					200, 0, 150, 600,
					big.NewInt(0).SetUint64(100),
					big.NewInt(0).SetUint64(400),
					metadataPdexv3.BaseAmplifier,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:             200,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 0, 150, 600,
					big.NewInt(0).SetUint64(150),
					big.NewInt(0).SetUint64(600),
					metadataPdexv3.BaseAmplifier,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:             200,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
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
					200, 0, 150, 600,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:             200,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 0, 150, 600,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					"123": &Share{
						amount:             200,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
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
				state:           tt.fields.state,
				shares:          tt.fields.shares,
				lpFeesPerShare:  tt.fields.lpFeesPerShare,
				protocolFees:    tt.fields.protocolFees,
				stakingPoolFees: tt.fields.stakingPoolFees,
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
		state           rawdbv2.Pdexv3PoolPair
		shares          map[string]*Share
		orderbook       Orderbook
		lpFeesPerShare  map[common.Hash]*big.Int
		protocolFees    map[common.Hash]uint64
		stakingPoolFees map[common.Hash]uint64
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
					300, 0, 150, 600,
					big.NewInt(0).SetUint64(150),
					big.NewInt(0).SetUint64(600),
					metadataPdexv3.BaseAmplifier,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             300,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
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
					200, 0, 100, 400,
					big.NewInt(0).SetUint64(100),
					big.NewInt(0).SetUint64(400),
					metadataPdexv3.BaseAmplifier,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                200,
						tradingFees:           map[common.Hash]uint64{},
						lastLPFeesPerShare:    map[common.Hash]*big.Int{},
						lastLmRewardsPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
			wantErr: false,
		},
		{
			name: "Not BaseAmplifier",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					300, 0, 150, 600,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             300,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
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
					200, 0, 100, 400,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:                200,
						tradingFees:           map[common.Hash]uint64{},
						lastLPFeesPerShare:    map[common.Hash]*big.Int{},
						lastLmRewardsPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PoolPairState{
				state:           tt.fields.state,
				shares:          tt.fields.shares,
				orderbook:       tt.fields.orderbook,
				lpFeesPerShare:  tt.fields.lpFeesPerShare,
				protocolFees:    tt.fields.protocolFees,
				stakingPoolFees: tt.fields.stakingPoolFees,
			}
			got, got1, got2, err := p.deductShare(tt.args.nftID, tt.args.shareAmount)
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
				t.Errorf("fieldsAfterProcess.shares got = %v, want %v", p.shares, tt.fieldsAfterProcess.shares)
			}
			if !reflect.DeepEqual(p.orderbook, tt.fieldsAfterProcess.orderbook) {
				t.Errorf("fieldsAfterProcess.orderbook got = %v, want %v", p.orderbook, tt.fieldsAfterProcess.orderbook)
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
		state           rawdbv2.Pdexv3PoolPair
		shares          map[string]*Share
		orderbook       Orderbook
		lpFeesPerShare  map[common.Hash]*big.Int
		protocolFees    map[common.Hash]uint64
		stakingPoolFees map[common.Hash]uint64
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
					300, 0, 150, 600,
					big.NewInt(0).SetUint64(150),
					big.NewInt(0).SetUint64(600),
					metadataPdexv3.BaseAmplifier,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             300,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
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
					300, 0, 100, 400,
					big.NewInt(0).SetUint64(100),
					big.NewInt(0).SetUint64(400),
					metadataPdexv3.BaseAmplifier,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             300,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
		},
		{
			name: "Not BaseAmplifier",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					300, 0, 150, 600,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             300,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
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
					300, 0, 100, 400,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             300,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PoolPairState{
				state:           tt.fields.state,
				shares:          tt.fields.shares,
				orderbook:       tt.fields.orderbook,
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
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

	temp, ok := big.NewInt(0).SetString("36893488147419103220", 10)
	if ok != true {
		panic(ok)
	}
	temp0, ok := big.NewInt(0).SetString("36893488147419103230", 10)
	if ok != true {
		panic(ok)
	}

	type fields struct {
		state           rawdbv2.Pdexv3PoolPair
		shares          map[string]*Share
		orderbook       Orderbook
		lpFeesPerShare  map[common.Hash]*big.Int
		protocolFees    map[common.Hash]uint64
		stakingPoolFees map[common.Hash]uint64
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
					300, 0, 150, 600,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             300,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					300, 0, 100, 600,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             300,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
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
					300, 0, 150, 600,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             300,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					300, 0, 150, 400,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             300,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
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
					200, 0, 100, 400,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             200,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 0, 150, 400,
					big.NewInt(0).SetUint64(300),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             200,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
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
					200, 0, 100, 400,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(800),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             200,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					200, 0, 100, 600,
					big.NewInt(0).SetUint64(200),
					big.NewInt(0).SetUint64(1200),
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             200,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
			args: args{
				tokenID:     *token1ID,
				amount:      200,
				shareAmount: 100,
				operator:    addOperator,
			},
			wantErr: false,
		},
		{
			name: "Out of range",
			fields: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					18446744073709551610, 0, 18446744073709551610, 18446744073709551610,
					temp,
					temp,
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             200,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
			fieldsAfterProcess: fields{
				state: *rawdbv2.NewPdexv3PoolPairWithValue(
					*token0ID, *token1ID,
					18446744073709551610, 0, 18446744073709551610, 18446744073709551615,
					temp, temp0,
					20000,
				),
				shares: map[string]*Share{
					nftID: &Share{
						amount:             200,
						tradingFees:        map[common.Hash]uint64{},
						lastLPFeesPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
			args: args{
				tokenID:     *token1ID,
				amount:      5,
				shareAmount: 5,
				operator:    addOperator,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PoolPairState{
				state:           tt.fields.state,
				shares:          tt.fields.shares,
				orderbook:       tt.fields.orderbook,
				lpFeesPerShare:  tt.fields.lpFeesPerShare,
				protocolFees:    tt.fields.protocolFees,
				stakingPoolFees: tt.fields.stakingPoolFees,
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

func TestPoolPairState_getDiff(t *testing.T) {
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)

	state := rawdbv2.NewPdexv3PoolPairWithValue(
		*token0ID, *token1ID, 200, 0, 100, 400,
		big.NewInt(200), big.NewInt(800), 20000,
	)
	compareState := rawdbv2.NewPdexv3PoolPairWithValue(
		*token0ID, *token1ID, 200, 0, 100, 400,
		big.NewInt(200), big.NewInt(800), 20000,
	)

	type fields struct {
		makingVolume      map[common.Hash]*MakingVolume
		state             rawdbv2.Pdexv3PoolPair
		shares            map[string]*Share
		orderRewards      map[string]*OrderReward
		orderbook         Orderbook
		lpFeesPerShare    map[common.Hash]*big.Int
		lmRewardsPerShare map[common.Hash]*big.Int
		protocolFees      map[common.Hash]uint64
		stakingPoolFees   map[common.Hash]uint64
		lmLockedShare     map[string]map[uint64]uint64
	}
	type args struct {
		poolPairID      string
		comparePoolPair *PoolPairState
		poolPairChange  *v2utils.PoolPairChange
		stateChange     *v2utils.StateChange
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *v2utils.PoolPairChange
		want1  *v2utils.StateChange
	}{
		{
			name: "Delete share, share reward, order reward, making volume",
			fields: fields{
				makingVolume: map[common.Hash]*MakingVolume{
					common.PRVCoinID: &MakingVolume{
						volume: map[string]*big.Int{
							common.PRVIDStr: big.NewInt(300),
						},
					},
					common.PDEXCoinID: &MakingVolume{
						volume: map[string]*big.Int{},
					},
				},
				state: *state,
				shares: map[string]*Share{
					common.PRVIDStr: &Share{
						amount: 100,
						tradingFees: map[common.Hash]uint64{
							common.PRVCoinID: 200,
							*token0ID:        300,
						},
						lastLPFeesPerShare: map[common.Hash]*big.Int{
							common.PRVCoinID: big.NewInt(200),
							*token0ID:        big.NewInt(300),
						},
						lastLmRewardsPerShare: map[common.Hash]*big.Int{},
					},
				},
				orderRewards: map[string]*OrderReward{
					common.PRVIDStr: &OrderReward{
						uncollectedRewards: Reward{
							common.PRVCoinID: 300,
						},
					},
				},
				orderbook:         Orderbook{},
				lpFeesPerShare:    map[common.Hash]*big.Int{},
				lmRewardsPerShare: map[common.Hash]*big.Int{},
				protocolFees:      map[common.Hash]uint64{},
				stakingPoolFees:   map[common.Hash]uint64{},
				lmLockedShare:     map[string]map[uint64]uint64{},
			},
			args: args{
				poolPairID:     "id",
				poolPairChange: v2utils.NewPoolPairChange(),
				stateChange:    v2utils.NewStateChange(),
				comparePoolPair: &PoolPairState{
					makingVolume: map[common.Hash]*MakingVolume{
						common.PRVCoinID: &MakingVolume{
							volume: map[string]*big.Int{
								common.PRVIDStr: big.NewInt(200),
							},
						},
						common.PDEXCoinID: &MakingVolume{
							volume: map[string]*big.Int{
								common.PDEXIDStr: big.NewInt(200),
							},
						},
					},
					state: *compareState,
					shares: map[string]*Share{
						common.PRVIDStr: &Share{
							amount: 100,
							tradingFees: map[common.Hash]uint64{
								common.PRVCoinID:  100,
								common.PDEXCoinID: 200,
							},
							lastLPFeesPerShare: map[common.Hash]*big.Int{
								common.PRVCoinID:  big.NewInt(100),
								common.PDEXCoinID: big.NewInt(200),
							},
							lastLmRewardsPerShare: map[common.Hash]*big.Int{},
						},
					},
					orderRewards: map[string]*OrderReward{
						common.PRVIDStr: &OrderReward{
							uncollectedRewards: Reward{
								common.PRVCoinID: 100,
							},
						},
						common.PDEXIDStr: &OrderReward{
							uncollectedRewards: Reward{
								common.PDEXCoinID: 100,
							},
						},
					},
					orderbook:         Orderbook{},
					lpFeesPerShare:    map[common.Hash]*big.Int{},
					lmRewardsPerShare: map[common.Hash]*big.Int{},
					protocolFees:      map[common.Hash]uint64{},
					stakingPoolFees:   map[common.Hash]uint64{},
					lmLockedShare:     map[string]map[uint64]uint64{},
				},
			},
			want: &v2utils.PoolPairChange{
				IsChanged: false,
				Shares: map[string]*v2utils.ShareChange{
					common.PRVIDStr: &v2utils.ShareChange{
						IsChanged: false,
						TradingFees: map[string]bool{
							common.PRVIDStr:   true,
							common.PDEXIDStr:  true,
							token0ID.String(): true,
						},
						LastLPFeesPerShare: map[string]bool{
							common.PRVIDStr:   true,
							common.PDEXIDStr:  true,
							token0ID.String(): true,
						},
						LastLmRewardsPerShare: map[string]bool{},
					},
				},
				OrderIDs:        map[string]bool{},
				LpFeesPerShare:  map[string]bool{},
				ProtocolFees:    map[string]bool{},
				StakingPoolFees: map[string]bool{},
				MakingVolume: map[string]*v2utils.MakingVolumeChange{
					common.PRVIDStr: &v2utils.MakingVolumeChange{
						Volume: map[string]bool{
							common.PRVIDStr: true,
						},
					},
					common.PDEXIDStr: &v2utils.MakingVolumeChange{
						Volume: map[string]bool{
							common.PDEXIDStr: true,
						},
					},
				},
				OrderRewards: map[string]*v2utils.OrderRewardChange{
					common.PRVIDStr: &v2utils.OrderRewardChange{
						UncollectedReward: map[string]bool{
							common.PRVIDStr: true,
						},
					},
					common.PDEXIDStr: &v2utils.OrderRewardChange{
						UncollectedReward: map[string]bool{
							common.PDEXIDStr: true,
						},
					},
				},
				LmLockedShare:     map[string]map[uint64]bool{},
				LmRewardsPerShare: map[string]bool{},
			},
			want1: &v2utils.StateChange{
				PoolPairs:    map[string]*v2utils.PoolPairChange{},
				StakingPools: map[string]*v2utils.StakingPoolChange{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PoolPairState{
				makingVolume:    tt.fields.makingVolume,
				state:           tt.fields.state,
				shares:          tt.fields.shares,
				orderRewards:    tt.fields.orderRewards,
				orderbook:       tt.fields.orderbook,
				lpFeesPerShare:  tt.fields.lpFeesPerShare,
				protocolFees:    tt.fields.protocolFees,
				stakingPoolFees: tt.fields.stakingPoolFees,
			}
			got, got1 := p.getDiff(tt.args.poolPairID, tt.args.comparePoolPair, tt.args.poolPairChange, tt.args.stateChange)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PoolPairState.getDiff() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("PoolPairState.getDiff() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestPoolPairState_updateToDB(t *testing.T) {
	initDB()
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)
	token0ID, err := common.Hash{}.NewHashFromStr("123")
	assert.Nil(t, err)
	token1ID, err := common.Hash{}.NewHashFromStr("456")
	assert.Nil(t, err)

	state := rawdbv2.NewPdexv3PoolPairWithValue(
		*token0ID, *token1ID, 200, 0, 100, 400,
		big.NewInt(200), big.NewInt(800), 20000,
	)

	err = statedb.StorePdexv3PoolPairMakingVolume(
		sDB, "id",
		statedb.NewPdexv3PoolPairMakingVolumeStateWithValue(
			common.PDEXIDStr, common.PDEXCoinID, big.NewInt(200),
		),
	)
	assert.Nil(t, err)

	err = statedb.StorePdexv3PoolPairOrderReward(
		sDB, "id",
		statedb.NewPdexv3PoolPairOrderRewardStateWithValue(
			common.PDEXCoinID, common.PDEXIDStr, 100,
		),
	)
	assert.Nil(t, err)

	err = statedb.StorePdexv3ShareTradingFee(
		sDB, "id", common.PRVIDStr,
		statedb.NewPdexv3ShareTradingFeeStateWithValue(common.PDEXCoinID, 100),
	)

	assert.Nil(t, err)
	err = statedb.StorePdexv3ShareLastLpFeePerShare(
		sDB, "id", common.PRVIDStr,
		statedb.NewPdexv3ShareLastLpFeePerShareStateWithValue(common.PDEXCoinID, big.NewInt(100)),
	)
	assert.Nil(t, err)

	type fields struct {
		makingVolume    map[common.Hash]*MakingVolume
		state           rawdbv2.Pdexv3PoolPair
		shares          map[string]*Share
		orderRewards    map[string]*OrderReward
		orderbook       Orderbook
		lpFeesPerShare  map[common.Hash]*big.Int
		protocolFees    map[common.Hash]uint64
		stakingPoolFees map[common.Hash]uint64
	}
	type args struct {
		env            StateEnvironment
		poolPairID     string
		poolPairChange *v2utils.PoolPairChange
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "delete makingVolume and orderRewards",
			fields: fields{
				makingVolume: map[common.Hash]*MakingVolume{
					common.PRVCoinID: &MakingVolume{
						volume: map[string]*big.Int{
							common.PRVIDStr: big.NewInt(300),
						},
					},
					*token0ID: &MakingVolume{
						volume: map[string]*big.Int{
							token0ID.String(): big.NewInt(200),
						},
					},
					common.PDEXCoinID: &MakingVolume{
						volume: map[string]*big.Int{},
					},
				},
				state: *state,
				shares: map[string]*Share{
					common.PRVIDStr: &Share{
						amount: 100,
						tradingFees: map[common.Hash]uint64{
							common.PRVCoinID: 200,
							*token0ID:        300,
						},
						lastLPFeesPerShare: map[common.Hash]*big.Int{
							common.PRVCoinID: big.NewInt(200),
							*token0ID:        big.NewInt(300),
						},
					},
				},
				orderRewards: map[string]*OrderReward{
					common.PRVIDStr: &OrderReward{
						uncollectedRewards: Reward{
							common.PRVCoinID: 300,
						},
					},
					token0ID.String(): &OrderReward{
						uncollectedRewards: Reward{
							*token0ID: 300,
						},
					},
				},
				orderbook:       Orderbook{},
				lpFeesPerShare:  map[common.Hash]*big.Int{},
				protocolFees:    map[common.Hash]uint64{},
				stakingPoolFees: map[common.Hash]uint64{},
			},
			args: args{
				poolPairID: "id",
				env: &stateEnvironment{
					stateDB: sDB,
				},
				poolPairChange: &v2utils.PoolPairChange{
					IsChanged: false,
					Shares: map[string]*v2utils.ShareChange{
						common.PRVIDStr: &v2utils.ShareChange{
							IsChanged: false,
							TradingFees: map[string]bool{
								common.PRVIDStr:   true,
								common.PDEXIDStr:  true,
								token0ID.String(): true,
							},
							LastLPFeesPerShare: map[string]bool{
								common.PRVIDStr:   true,
								common.PDEXIDStr:  true,
								token0ID.String(): true,
							},
						},
					},
					OrderIDs:        map[string]bool{},
					LpFeesPerShare:  map[string]bool{},
					ProtocolFees:    map[string]bool{},
					StakingPoolFees: map[string]bool{},
					MakingVolume: map[string]*v2utils.MakingVolumeChange{
						token0ID.String(): &v2utils.MakingVolumeChange{
							Volume: map[string]bool{
								token0ID.String(): true,
							},
						},
						common.PRVIDStr: &v2utils.MakingVolumeChange{
							Volume: map[string]bool{
								common.PRVIDStr: true,
							},
						},
						common.PDEXIDStr: &v2utils.MakingVolumeChange{
							Volume: map[string]bool{
								common.PDEXIDStr: true,
							},
						},
					},
					OrderRewards: map[string]*v2utils.OrderRewardChange{
						token0ID.String(): &v2utils.OrderRewardChange{
							UncollectedReward: map[string]bool{
								token0ID.String(): true,
							},
						},
						common.PRVIDStr: &v2utils.OrderRewardChange{
							UncollectedReward: map[string]bool{
								common.PRVIDStr: true,
							},
						},
						common.PDEXIDStr: &v2utils.OrderRewardChange{
							UncollectedReward: map[string]bool{
								common.PDEXIDStr: true,
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PoolPairState{
				makingVolume:    tt.fields.makingVolume,
				state:           tt.fields.state,
				shares:          tt.fields.shares,
				orderRewards:    tt.fields.orderRewards,
				orderbook:       tt.fields.orderbook,
				lpFeesPerShare:  tt.fields.lpFeesPerShare,
				protocolFees:    tt.fields.protocolFees,
				stakingPoolFees: tt.fields.stakingPoolFees,
			}
			if err := p.updateToDB(tt.args.env, tt.args.poolPairID, tt.args.poolPairChange); (err != nil) != tt.wantErr {
				t.Errorf("PoolPairState.updateToDB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
