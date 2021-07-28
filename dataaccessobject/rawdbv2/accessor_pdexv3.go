package rawdbv2

import (
	"encoding/json"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
)

//Pdexv3Contribution Real data store to statedb
type Pdexv3Contribution struct {
	poolPairID     string
	receiveAddress string
	refundAddress  string
	tokenID        common.Hash
	amount         uint64
	amplifier      uint
	txReqID        common.Hash
	shardID        byte
}

func (contribution *Pdexv3Contribution) ShardID() byte {
	return contribution.shardID
}

func (contribution *Pdexv3Contribution) TxReqID() common.Hash {
	return contribution.txReqID
}

func (contribution *Pdexv3Contribution) Amplifier() uint {
	return contribution.amplifier
}

func (contribution *Pdexv3Contribution) PoolPairID() string {
	return contribution.poolPairID
}

func (contribution *Pdexv3Contribution) ReceiveAddress() string {
	return contribution.receiveAddress
}

func (contribution *Pdexv3Contribution) RefundAddress() string {
	return contribution.refundAddress
}

func (contribution *Pdexv3Contribution) TokenID() common.Hash {
	return contribution.tokenID
}

func (contribution *Pdexv3Contribution) Amount() uint64 {
	return contribution.amount
}

func (contribution *Pdexv3Contribution) SetAmount(amount uint64) {
	contribution.amount = amount
}

func (contribution *Pdexv3Contribution) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PoolPairID     string      `json:"PoolPairID"`
		ReceiveAddress string      `json:"ReceiveAddress"`
		RefundAddress  string      `json:"RefundAddress"`
		TokenID        common.Hash `json:"TokenID"`
		Amount         uint64      `json:"Amount"`
		Amplifier      uint        `json:"Amplifier"`
		TxReqID        common.Hash `json:"TxReqID"`
		ShardID        byte        `json:"ShardID"`
	}{
		PoolPairID:     contribution.poolPairID,
		ReceiveAddress: contribution.receiveAddress,
		RefundAddress:  contribution.refundAddress,
		TokenID:        contribution.tokenID,
		Amount:         contribution.amount,
		TxReqID:        contribution.txReqID,
		Amplifier:      contribution.amplifier,
		ShardID:        contribution.shardID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (contribution *Pdexv3Contribution) UnmarshalJSON(data []byte) error {
	temp := struct {
		PoolPairID     string      `json:"PoolPairID"`
		ReceiveAddress string      `json:"ReceiveAddress"`
		RefundAddress  string      `json:"RefundAddress"`
		TokenID        common.Hash `json:"TokenID"`
		Amount         uint64      `json:"Amount"`
		Amplifier      uint        `json:"Amplifier"`
		TxReqID        common.Hash `json:"TxReqID"`
		ShardID        byte        `json:"ShardID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	contribution.poolPairID = temp.PoolPairID
	contribution.receiveAddress = temp.ReceiveAddress
	contribution.refundAddress = temp.RefundAddress
	contribution.tokenID = temp.TokenID
	contribution.amount = temp.Amount
	contribution.txReqID = temp.TxReqID
	contribution.amplifier = temp.Amplifier
	contribution.shardID = temp.ShardID
	return nil
}

func (contribution *Pdexv3Contribution) Clone() *Pdexv3Contribution {
	return NewPdexv3ContributionWithValue(
		contribution.poolPairID, contribution.receiveAddress, contribution.refundAddress,
		contribution.tokenID, contribution.txReqID,
		contribution.amount, contribution.amplifier, contribution.shardID,
	)
}

func NewPdexv3Contribution() *Pdexv3Contribution {
	return &Pdexv3Contribution{}
}

func NewPdexv3ContributionWithValue(
	poolPairID, receiveAddress, refundAddress string,
	tokenID, txReqID common.Hash,
	amount uint64, amplifier uint, shardID byte,
) *Pdexv3Contribution {
	return &Pdexv3Contribution{
		poolPairID:     poolPairID,
		refundAddress:  refundAddress,
		receiveAddress: receiveAddress,
		tokenID:        tokenID,
		amount:         amount,
		txReqID:        txReqID,
		amplifier:      amplifier,
		shardID:        shardID,
	}
}

type Pdexv3PoolPair struct {
	token0ID              common.Hash
	token1ID              common.Hash
	token0RealAmount      uint64
	token1RealAmount      uint64
	currentContributionID uint64
	token0VirtualAmount   big.Int
	token1VirtualAmount   big.Int
	amplifier             uint
}

func (pp *Pdexv3PoolPair) Amplifier() uint {
	return pp.amplifier
}

func (pp *Pdexv3PoolPair) Token0ID() common.Hash {
	return pp.token0ID
}

func (pp *Pdexv3PoolPair) Token1ID() common.Hash {
	return pp.token1ID
}

func (pp *Pdexv3PoolPair) Token0RealAmount() uint64 {
	return pp.token0RealAmount
}

func (pp *Pdexv3PoolPair) Token1RealAmount() uint64 {
	return pp.token1RealAmount
}

func (pp *Pdexv3PoolPair) CurrentContributionID() uint64 {
	return pp.currentContributionID
}

func (pp *Pdexv3PoolPair) Token0VirtualAmount() big.Int {
	return pp.token0VirtualAmount
}

func (pp *Pdexv3PoolPair) Token1VirtualAmount() big.Int {
	return pp.token1VirtualAmount
}

func (pp *Pdexv3PoolPair) SetToken0RealAmount(amount uint64) {
	pp.token0RealAmount = amount
}

func (pp *Pdexv3PoolPair) SetToken1RealAmount(amount uint64) {
	pp.token1RealAmount = amount
}

func (pp *Pdexv3PoolPair) SetCurrentContributionID(id uint64) {
	pp.currentContributionID = id
}

func (pp *Pdexv3PoolPair) SetToken0VirtualAmount(amount big.Int) {
	pp.token0VirtualAmount = amount
}

func (pp *Pdexv3PoolPair) SetToken1VirtualAmount(amount big.Int) {
	pp.token1VirtualAmount = amount
}

func (pp *Pdexv3PoolPair) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Token0ID              common.Hash `json:"Token0ID"`
		Token1ID              common.Hash `json:"Token1ID"`
		Token0RealAmount      uint64      `json:"Token0RealAmount"`
		Token1RealAmount      uint64      `json:"Token1RealAmount"`
		CurrentContributionID uint64      `json:"CurrentContributionID"`
		Token0VirtualAmount   big.Int     `json:"Token0VirtualAmount"`
		Token1VirtualAmount   big.Int     `json:"Token1VirtualAmount"`
		Amplifier             uint        `json:"Amplifier"`
	}{
		Token0ID:              pp.token0ID,
		Token1ID:              pp.token1ID,
		Token0RealAmount:      pp.token0RealAmount,
		Token1RealAmount:      pp.token1RealAmount,
		CurrentContributionID: pp.currentContributionID,
		Token0VirtualAmount:   pp.token0VirtualAmount,
		Token1VirtualAmount:   pp.token1VirtualAmount,
		Amplifier:             pp.amplifier,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (pp *Pdexv3PoolPair) UnmarshalJSON(data []byte) error {
	temp := struct {
		Token0ID              common.Hash `json:"Token0ID"`
		Token1ID              common.Hash `json:"Token1ID"`
		Token0RealAmount      uint64      `json:"Token0RealAmount"`
		Token1RealAmount      uint64      `json:"Token1RealAmount"`
		CurrentContributionID uint64      `json:"CurrentContributionID"`
		Token0VirtualAmount   big.Int     `json:"Token0VirtualAmount"`
		Token1VirtualAmount   big.Int     `json:"Token1VirtualAmount"`
		Amplifier             uint        `json:"Amplifier"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	pp.token0ID = temp.Token0ID
	pp.token1ID = temp.Token1ID
	pp.token0RealAmount = temp.Token0RealAmount
	pp.token1RealAmount = temp.Token1RealAmount
	pp.currentContributionID = temp.CurrentContributionID
	pp.token0VirtualAmount = temp.Token0VirtualAmount
	pp.token1VirtualAmount = temp.Token1VirtualAmount
	pp.amplifier = temp.Amplifier
	return nil
}

func (pp *Pdexv3PoolPair) Clone() *Pdexv3PoolPair {
	return NewPdexv3PoolPairWithValue(
		pp.token0ID, pp.token1ID,
		pp.token0RealAmount, pp.token1RealAmount, pp.currentContributionID,
		pp.token0VirtualAmount, pp.token1VirtualAmount, pp.amplifier,
	)
}

func NewPdexv3PoolPair() *Pdexv3PoolPair {
	return &Pdexv3PoolPair{}
}

func NewPdexv3PoolPairWithValue(
	token0ID, token1ID common.Hash,
	token0RealAmount, token1RealAmount, currentContributionID uint64,
	token0VirtualAmount, token1VirtualAmount big.Int,
	amplifier uint,
) *Pdexv3PoolPair {
	return &Pdexv3PoolPair{
		token0ID:              token0ID,
		token1ID:              token1ID,
		token0RealAmount:      token0RealAmount,
		token1RealAmount:      token1RealAmount,
		currentContributionID: currentContributionID,
		token0VirtualAmount:   token0VirtualAmount,
		token1VirtualAmount:   token1VirtualAmount,
		amplifier:             amplifier,
	}
}
