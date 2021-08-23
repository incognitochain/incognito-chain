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
	nftID          common.Hash
	shardID        byte
}

func (contribution *Pdexv3Contribution) NftID() common.Hash {
	return contribution.nftID
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
		NftID          common.Hash `json:"NftID"`
		ShardID        byte        `json:"ShardID"`
	}{
		PoolPairID:     contribution.poolPairID,
		ReceiveAddress: contribution.receiveAddress,
		RefundAddress:  contribution.refundAddress,
		TokenID:        contribution.tokenID,
		Amount:         contribution.amount,
		TxReqID:        contribution.txReqID,
		Amplifier:      contribution.amplifier,
		NftID:          contribution.nftID,
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
		NftID          common.Hash `json:"NftID"`
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
	contribution.nftID = temp.NftID
	return nil
}

func (contribution *Pdexv3Contribution) Clone() *Pdexv3Contribution {
	return NewPdexv3ContributionWithValue(
		contribution.poolPairID, contribution.receiveAddress, contribution.refundAddress,
		contribution.tokenID, contribution.txReqID, contribution.nftID,
		contribution.amount, contribution.amplifier, contribution.shardID,
	)
}

func NewPdexv3Contribution() *Pdexv3Contribution {
	return &Pdexv3Contribution{}
}

func NewPdexv3ContributionWithValue(
	poolPairID, receiveAddress, refundAddress string,
	tokenID, txReqID, nftID common.Hash,
	amount uint64, amplifier uint, shardID byte,
) *Pdexv3Contribution {
	return &Pdexv3Contribution{
		poolPairID:     poolPairID,
		refundAddress:  refundAddress,
		receiveAddress: receiveAddress,
		tokenID:        tokenID,
		amount:         amount,
		txReqID:        txReqID,
		nftID:          nftID,
		amplifier:      amplifier,
		shardID:        shardID,
	}
}

type Pdexv3PoolPair struct {
	shareAmount         uint64
	token0ID            common.Hash
	token1ID            common.Hash
	token0RealAmount    uint64
	token1RealAmount    uint64
	token0VirtualAmount *big.Int
	token1VirtualAmount *big.Int
	amplifier           uint
	lpFeesPerShare      map[common.Hash]*big.Int
	protocolFees        map[common.Hash]uint64
	stakingPoolFees     map[common.Hash]uint64
}

func (pp *Pdexv3PoolPair) ShareAmount() uint64 {
	return pp.shareAmount
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

func (pp *Pdexv3PoolPair) Token0VirtualAmount() *big.Int {
	return pp.token0VirtualAmount
}

func (pp *Pdexv3PoolPair) Token1VirtualAmount() *big.Int {
	return pp.token1VirtualAmount
}

func (pp *Pdexv3PoolPair) LPFeesPerShare() map[common.Hash]*big.Int {
	return pp.lpFeesPerShare
}

func (pp *Pdexv3PoolPair) ProtocolFees() map[common.Hash]uint64 {
	return pp.protocolFees
}

func (pp *Pdexv3PoolPair) StakingPoolFees() map[common.Hash]uint64 {
	return pp.stakingPoolFees
}

func (pp *Pdexv3PoolPair) SetShareAmount(shareAmount uint64) {
	pp.shareAmount = shareAmount
}

func (pp *Pdexv3PoolPair) SetToken0RealAmount(amount uint64) {
	pp.token0RealAmount = amount
}

func (pp *Pdexv3PoolPair) SetToken1RealAmount(amount uint64) {
	pp.token1RealAmount = amount
}

func (pp *Pdexv3PoolPair) SetToken0VirtualAmount(amount *big.Int) {
	pp.token0VirtualAmount = amount
}

func (pp *Pdexv3PoolPair) SetToken1VirtualAmount(amount *big.Int) {
	pp.token1VirtualAmount = amount
}

func (pp *Pdexv3PoolPair) SetLPFeesPerShare(fees map[common.Hash]*big.Int) {
	pp.lpFeesPerShare = fees
}

func (pp *Pdexv3PoolPair) SetProtocolFees(fees map[common.Hash]uint64) {
	pp.protocolFees = fees
}

func (pp *Pdexv3PoolPair) SetStakingPoolFees(fees map[common.Hash]uint64) {
	pp.stakingPoolFees = fees
}

func (pp *Pdexv3PoolPair) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Token0ID            common.Hash              `json:"Token0ID"`
		Token1ID            common.Hash              `json:"Token1ID"`
		Token0RealAmount    uint64                   `json:"Token0RealAmount"`
		Token1RealAmount    uint64                   `json:"Token1RealAmount"`
		Token0VirtualAmount *big.Int                 `json:"Token0VirtualAmount"`
		Token1VirtualAmount *big.Int                 `json:"Token1VirtualAmount"`
		Amplifier           uint                     `json:"Amplifier"`
		ShareAmount         uint64                   `json:"ShareAmount"`
		LPFeesPerShare      map[common.Hash]*big.Int `json:"LPFeesPerShare"`
		ProtocolFees        map[common.Hash]uint64   `json:"ProtocolFees"`
		StakingPoolFees     map[common.Hash]uint64   `json:"StakingPoolFees"`
	}{
		Token0ID:            pp.token0ID,
		Token1ID:            pp.token1ID,
		Token0RealAmount:    pp.token0RealAmount,
		Token1RealAmount:    pp.token1RealAmount,
		Token0VirtualAmount: pp.token0VirtualAmount,
		Token1VirtualAmount: pp.token1VirtualAmount,
		Amplifier:           pp.amplifier,
		ShareAmount:         pp.shareAmount,
		LPFeesPerShare:      pp.lpFeesPerShare,
		ProtocolFees:        pp.protocolFees,
		StakingPoolFees:     pp.stakingPoolFees,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (pp *Pdexv3PoolPair) UnmarshalJSON(data []byte) error {
	temp := struct {
		Token0ID            common.Hash              `json:"Token0ID"`
		Token1ID            common.Hash              `json:"Token1ID"`
		Token0RealAmount    uint64                   `json:"Token0RealAmount"`
		Token1RealAmount    uint64                   `json:"Token1RealAmount"`
		Token0VirtualAmount *big.Int                 `json:"Token0VirtualAmount"`
		Token1VirtualAmount *big.Int                 `json:"Token1VirtualAmount"`
		Amplifier           uint                     `json:"Amplifier"`
		ShareAmount         uint64                   `json:"ShareAmount"`
		LPFeesPerShare      map[common.Hash]*big.Int `json:"LPFeesPerShare"`
		ProtocolFees        map[common.Hash]uint64   `json:"ProtocolFees"`
		StakingPoolFees     map[common.Hash]uint64   `json:"StakingPoolFees"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	pp.token0ID = temp.Token0ID
	pp.token1ID = temp.Token1ID
	pp.token0RealAmount = temp.Token0RealAmount
	pp.token1RealAmount = temp.Token1RealAmount
	pp.token0VirtualAmount = temp.Token0VirtualAmount
	pp.token1VirtualAmount = temp.Token1VirtualAmount
	pp.amplifier = temp.Amplifier
	pp.shareAmount = temp.ShareAmount
	pp.lpFeesPerShare = temp.LPFeesPerShare
	pp.protocolFees = temp.ProtocolFees
	pp.stakingPoolFees = temp.StakingPoolFees
	return nil
}

func (pp *Pdexv3PoolPair) Clone() *Pdexv3PoolPair {
	return NewPdexv3PoolPairWithValue(
		pp.token0ID, pp.token1ID, pp.shareAmount,
		pp.token0RealAmount, pp.token1RealAmount,
		pp.token0VirtualAmount, pp.token1VirtualAmount, pp.amplifier,
		pp.lpFeesPerShare, pp.protocolFees, pp.stakingPoolFees,
	)
}

func NewPdexv3PoolPair() *Pdexv3PoolPair {
	return &Pdexv3PoolPair{}
}

func NewPdexv3PoolPairWithValue(
	token0ID, token1ID common.Hash,
	shareAmount, token0RealAmount, token1RealAmount uint64,
	token0VirtualAmount, token1VirtualAmount *big.Int,
	amplifier uint, lpFeesPerShare map[common.Hash]*big.Int,
	protocolFees map[common.Hash]uint64, stakingPoolFees map[common.Hash]uint64,
) *Pdexv3PoolPair {
	return &Pdexv3PoolPair{
		token0ID:            token0ID,
		token1ID:            token1ID,
		token0RealAmount:    token0RealAmount,
		token1RealAmount:    token1RealAmount,
		token0VirtualAmount: token0VirtualAmount,
		token1VirtualAmount: token1VirtualAmount,
		amplifier:           amplifier,
		shareAmount:         shareAmount,
		lpFeesPerShare:      lpFeesPerShare,
		protocolFees:        protocolFees,
		stakingPoolFees:     stakingPoolFees,
	}
}

type Pdexv3Order struct {
	id             string
	nftID          common.Hash
	token0Rate     uint64
	token1Rate     uint64
	token0Balance  uint64
	token1Balance  uint64
	tradeDirection byte
	fee            uint64
}

func (o *Pdexv3Order) Id() string            { return o.id }
func (o *Pdexv3Order) NftID() common.Hash    { return o.nftID }
func (o *Pdexv3Order) Token0Rate() uint64    { return o.token0Rate }
func (o *Pdexv3Order) Token1Rate() uint64    { return o.token1Rate }
func (o *Pdexv3Order) Token0Balance() uint64 { return o.token0Balance }
func (o *Pdexv3Order) Token1Balance() uint64 { return o.token1Balance }
func (o *Pdexv3Order) TradeDirection() byte  { return o.tradeDirection }
func (o *Pdexv3Order) Fee() uint64           { return o.fee }

// SetToken0Balance() changes the token0 balance of this order. Only balances & fee can be updated,
// while rates, id & trade direction cannot
func (o *Pdexv3Order) SetToken0Balance(b uint64) { o.token0Balance = b }
func (o *Pdexv3Order) SetToken1Balance(b uint64) { o.token1Balance = b }
func (o *Pdexv3Order) SetFee(fee uint64)         { o.fee = fee }

func NewPdexv3OrderWithValue(
	id string, nftID common.Hash,
	token0Rate, token1Rate, token0Balance, token1Balance uint64,
	tradeDirection byte,
	fee uint64,
) *Pdexv3Order {
	return &Pdexv3Order{
		id:             id,
		nftID:          nftID,
		token0Rate:     token0Rate,
		token1Rate:     token1Rate,
		token0Balance:  token0Balance,
		token1Balance:  token1Balance,
		tradeDirection: tradeDirection,
		fee:            fee,
	}
}

func (o *Pdexv3Order) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Id             string      `json:"Id"`
		NftID          common.Hash `json:"NftID"`
		Token0Rate     uint64      `json:"Token0Rate"`
		Token1Rate     uint64      `json:"Token1Rate"`
		Token0Balance  uint64      `json:"Token0Balance"`
		Token1Balance  uint64      `json:"Token1Balance"`
		TradeDirection byte        `json:"TradeDirection"`
		Fee            uint64      `json:"Fee"`
	}{
		Id:             o.id,
		NftID:          o.nftID,
		Token0Rate:     o.token0Rate,
		Token1Rate:     o.token1Rate,
		Token0Balance:  o.token0Balance,
		Token1Balance:  o.token1Balance,
		TradeDirection: o.tradeDirection,
		Fee:            o.fee,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (o *Pdexv3Order) UnmarshalJSON(data []byte) error {
	var temp struct {
		Id             string      `json:"Id"`
		NftID          common.Hash `json:"NftID"`
		Token0Rate     uint64      `json:"Token0Rate"`
		Token1Rate     uint64      `json:"Token1Rate"`
		Token0Balance  uint64      `json:"Token0Balance"`
		Token1Balance  uint64 `json:"Token1Balance"`
		TradeDirection byte   `json:"TradeDirection"`
		Fee            uint64 `json:"Fee"`
	}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	*o = Pdexv3Order{
		id:             temp.Id,
		nftID:          temp.NftID,
		token0Rate:     temp.Token0Rate,
		token1Rate:     temp.Token1Rate,
		token0Balance:  temp.Token0Balance,
		token1Balance:  temp.Token1Balance,
		tradeDirection: temp.TradeDirection,
		fee:            temp.Fee,
	}
	return nil
}

func (o *Pdexv3Order) Clone() *Pdexv3Order {
	return NewPdexv3OrderWithValue(o.id, o.nftID, o.token0Rate, o.token1Rate,
		o.token0Balance, o.token1Balance, o.tradeDirection, o.fee)
}
