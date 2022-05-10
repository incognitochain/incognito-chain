package rawdbv2

import (
	"encoding/json"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/utils"
)

//Pdexv3Contribution Real data store to statedb
type Pdexv3Contribution struct {
	poolPairID   string
	otaReceiver  string
	otaReceivers map[common.Hash]string
	tokenID      common.Hash
	amount       uint64
	amplifier    uint
	txReqID      common.Hash
	nftID        common.Hash // prsent for nftID or accessID
	shardID      byte
	accessOTA    []byte
}

func (contribution *Pdexv3Contribution) NftID() common.Hash {
	return contribution.nftID
}

func (contribution *Pdexv3Contribution) SetNftID(nftID common.Hash) {
	contribution.nftID = nftID
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

func (contribution *Pdexv3Contribution) OtaReceiver() string {
	return contribution.otaReceiver
}

func (contribution *Pdexv3Contribution) AccessOTA() []byte {
	return contribution.accessOTA
}

func (contribution *Pdexv3Contribution) SetAccessOTA(accessOTA []byte) {
	contribution.accessOTA = accessOTA
}

func (contribution *Pdexv3Contribution) TokenID() common.Hash {
	return contribution.tokenID
}

func (contribution *Pdexv3Contribution) Amount() uint64 {
	return contribution.amount
}

//OtaReceivers read only function
func (contribution *Pdexv3Contribution) OtaReceivers() map[common.Hash]string {
	return contribution.otaReceivers
}

func (contribution *Pdexv3Contribution) SetAmount(amount uint64) {
	contribution.amount = amount
}

func (contribution *Pdexv3Contribution) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PoolPairID   string                 `json:"PoolPairID"`
		OtaReceiver  string                 `json:"OtaReceiver,omitempty"`
		TokenID      common.Hash            `json:"TokenID"`
		Amount       uint64                 `json:"Amount"`
		Amplifier    uint                   `json:"Amplifier"`
		TxReqID      common.Hash            `json:"TxReqID"`
		NftID        common.Hash            `json:"NftID"`
		ShardID      byte                   `json:"ShardID"`
		AccessOTA    []byte                 `json:"AccessOTA,omitempty"`
		OtaReceivers map[common.Hash]string `json:"OtaReceivers,omitempty"`
	}{
		PoolPairID:   contribution.poolPairID,
		OtaReceiver:  contribution.otaReceiver,
		TokenID:      contribution.tokenID,
		Amount:       contribution.amount,
		TxReqID:      contribution.txReqID,
		Amplifier:    contribution.amplifier,
		NftID:        contribution.nftID,
		ShardID:      contribution.shardID,
		AccessOTA:    contribution.accessOTA,
		OtaReceivers: contribution.otaReceivers,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (contribution *Pdexv3Contribution) UnmarshalJSON(data []byte) error {
	temp := struct {
		PoolPairID   string                 `json:"PoolPairID"`
		OtaReceiver  string                 `json:"OtaReceiver,omitempty"`
		TokenID      common.Hash            `json:"TokenID"`
		Amount       uint64                 `json:"Amount"`
		Amplifier    uint                   `json:"Amplifier"`
		TxReqID      common.Hash            `json:"TxReqID"`
		NftID        common.Hash            `json:"NftID"`
		ShardID      byte                   `json:"ShardID"`
		AccessOTA    []byte                 `json:"AccessOTA,omitempty"`
		OtaReceivers map[common.Hash]string `json:"OtaReceivers,omitempty"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	contribution.poolPairID = temp.PoolPairID
	contribution.otaReceiver = temp.OtaReceiver
	contribution.tokenID = temp.TokenID
	contribution.amount = temp.Amount
	contribution.txReqID = temp.TxReqID
	contribution.amplifier = temp.Amplifier
	contribution.shardID = temp.ShardID
	contribution.nftID = temp.NftID
	contribution.accessOTA = temp.AccessOTA
	contribution.otaReceivers = temp.OtaReceivers
	return nil
}

func (contribution *Pdexv3Contribution) Clone() *Pdexv3Contribution {
	return NewPdexv3ContributionWithValue(
		contribution.poolPairID, contribution.otaReceiver,
		contribution.tokenID, contribution.txReqID, contribution.nftID,
		contribution.amount, contribution.amplifier, contribution.shardID,
		contribution.accessOTA, contribution.otaReceivers,
	)
}

func NewPdexv3Contribution() *Pdexv3Contribution {
	return &Pdexv3Contribution{}
}

func NewPdexv3ContributionWithValue(
	poolPairID, otaReceiver string,
	tokenID, txReqID, nftID common.Hash,
	amount uint64, amplifier uint, shardID byte,
	accessOTA []byte, otaReceivers map[common.Hash]string,
) *Pdexv3Contribution {
	return &Pdexv3Contribution{
		poolPairID:   poolPairID,
		otaReceiver:  otaReceiver,
		tokenID:      tokenID,
		amount:       amount,
		txReqID:      txReqID,
		nftID:        nftID,
		amplifier:    amplifier,
		shardID:      shardID,
		accessOTA:    accessOTA,
		otaReceivers: otaReceivers,
	}
}

// This contribution use nftID to contribute
func (contribution *Pdexv3Contribution) UseNft() bool {
	return len(contribution.accessOTA) == 0 && len(contribution.otaReceivers) == 0 && contribution.otaReceiver != utils.EmptyString
}

// This contribution use accessOTA to contribute for creating new LP
func (contribution *Pdexv3Contribution) UseAccessOTANewLP() bool {
	return len(contribution.otaReceivers) != 0 && contribution.otaReceiver != utils.EmptyString
}

// This contribution use accessOTA to contribute for adding liquidity to old LP
func (contribution *Pdexv3Contribution) UseAccessOTAOldLP() bool {
	return len(contribution.accessOTA) == 0 && len(contribution.otaReceivers) != 0 && contribution.otaReceiver == utils.EmptyString
}

type Pdexv3PoolPair struct {
	shareAmount         uint64
	lmLockedShareAmount uint64
	token0ID            common.Hash
	token1ID            common.Hash
	token0RealAmount    uint64
	token1RealAmount    uint64
	token0VirtualAmount *big.Int
	token1VirtualAmount *big.Int
	amplifier           uint
}

func (pp *Pdexv3PoolPair) ShareAmount() uint64 {
	return pp.shareAmount
}

func (pp *Pdexv3PoolPair) LmLockedShareAmount() uint64 {
	return pp.lmLockedShareAmount
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

func (pp *Pdexv3PoolPair) SetShareAmount(shareAmount uint64) {
	pp.shareAmount = shareAmount
}

func (pp *Pdexv3PoolPair) SetLmLockedShareAmount(lmLockedShareAmount uint64) {
	pp.lmLockedShareAmount = lmLockedShareAmount
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

func (pp *Pdexv3PoolPair) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Token0ID            common.Hash `json:"Token0ID"`
		Token1ID            common.Hash `json:"Token1ID"`
		Token0RealAmount    uint64      `json:"Token0RealAmount"`
		Token1RealAmount    uint64      `json:"Token1RealAmount"`
		Token0VirtualAmount *big.Int    `json:"Token0VirtualAmount"`
		Token1VirtualAmount *big.Int    `json:"Token1VirtualAmount"`
		Amplifier           uint        `json:"Amplifier"`
		ShareAmount         uint64      `json:"ShareAmount"`
		LmLockedShareAmount uint64      `json:"LmLockedShareAmount,omitempty"`
	}{
		Token0ID:            pp.token0ID,
		Token1ID:            pp.token1ID,
		Token0RealAmount:    pp.token0RealAmount,
		Token1RealAmount:    pp.token1RealAmount,
		Token0VirtualAmount: pp.token0VirtualAmount,
		Token1VirtualAmount: pp.token1VirtualAmount,
		Amplifier:           pp.amplifier,
		ShareAmount:         pp.shareAmount,
		LmLockedShareAmount: pp.lmLockedShareAmount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (pp *Pdexv3PoolPair) UnmarshalJSON(data []byte) error {
	temp := struct {
		Token0ID            common.Hash `json:"Token0ID"`
		Token1ID            common.Hash `json:"Token1ID"`
		Token0RealAmount    uint64      `json:"Token0RealAmount"`
		Token1RealAmount    uint64      `json:"Token1RealAmount"`
		Token0VirtualAmount *big.Int    `json:"Token0VirtualAmount"`
		Token1VirtualAmount *big.Int    `json:"Token1VirtualAmount"`
		Amplifier           uint        `json:"Amplifier"`
		ShareAmount         uint64      `json:"ShareAmount"`
		LmLockedShareAmount uint64      `json:"LmLockedShareAmount,omitempty"`
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
	pp.lmLockedShareAmount = temp.LmLockedShareAmount
	return nil
}

func (pp *Pdexv3PoolPair) Clone() *Pdexv3PoolPair {
	res := NewPdexv3PoolPairWithValue(
		pp.token0ID, pp.token1ID, pp.shareAmount, pp.lmLockedShareAmount,
		pp.token0RealAmount, pp.token1RealAmount,
		pp.token0VirtualAmount, pp.token1VirtualAmount, pp.amplifier,
	)
	res.token0VirtualAmount = new(big.Int).Set(pp.token0VirtualAmount)
	res.token1VirtualAmount = new(big.Int).Set(pp.token1VirtualAmount)
	return res
}

func NewPdexv3PoolPair() *Pdexv3PoolPair {
	return &Pdexv3PoolPair{}
}

func NewPdexv3PoolPairWithValue(
	token0ID, token1ID common.Hash,
	shareAmount, lmLockedShareAmount,
	token0RealAmount, token1RealAmount uint64,
	token0VirtualAmount, token1VirtualAmount *big.Int,
	amplifier uint,
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
		lmLockedShareAmount: lmLockedShareAmount,
	}
}

type Pdexv3Order struct {
	id             string
	nftID          common.Hash
	accessOTA      []byte
	token0Rate     uint64
	token1Rate     uint64
	token0Balance  uint64
	token1Balance  uint64
	tradeDirection byte
	receiver       [2]string
}

func (o *Pdexv3Order) Id() string             { return o.id }
func (o *Pdexv3Order) NftID() common.Hash     { return o.nftID }
func (o *Pdexv3Order) Token0Rate() uint64     { return o.token0Rate }
func (o *Pdexv3Order) Token1Rate() uint64     { return o.token1Rate }
func (o *Pdexv3Order) Token0Balance() uint64  { return o.token0Balance }
func (o *Pdexv3Order) Token1Balance() uint64  { return o.token1Balance }
func (o *Pdexv3Order) TradeDirection() byte   { return o.tradeDirection }
func (o *Pdexv3Order) Token0Receiver() string { return o.receiver[0] }
func (o *Pdexv3Order) Token1Receiver() string { return o.receiver[1] }
func (o *Pdexv3Order) AccessOTA() []byte      { return o.accessOTA }
func (o *Pdexv3Order) RewardKey() common.Hash { return o.nftID }

// SetToken0Balance() changes the token0 balance of this order. Only balances can be updated,
// while rates, id & trade direction cannot
func (o *Pdexv3Order) SetToken0Balance(b uint64) { o.token0Balance = b }
func (o *Pdexv3Order) SetToken1Balance(b uint64) { o.token1Balance = b }
func (o *Pdexv3Order) SetAccessOTA(b []byte)     { o.accessOTA = b }

func NewPdexv3OrderWithValue(
	id string, nftID common.Hash, accessOTA []byte,
	token0Rate, token1Rate, token0Balance, token1Balance uint64,
	tradeDirection byte,
	receiver [2]string,
) *Pdexv3Order {
	return &Pdexv3Order{
		id:             id,
		nftID:          nftID,
		accessOTA:      accessOTA,
		token0Rate:     token0Rate,
		token1Rate:     token1Rate,
		token0Balance:  token0Balance,
		token1Balance:  token1Balance,
		tradeDirection: tradeDirection,
		receiver:       receiver,
	}
}

func (o *Pdexv3Order) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Id             string      `json:"Id"`
		NftID          common.Hash `json:"NftID"`
		AccessOTA      []byte      `json:"AccessOTA,omitempty"`
		Token0Rate     uint64      `json:"Token0Rate"`
		Token1Rate     uint64      `json:"Token1Rate"`
		Token0Balance  uint64      `json:"Token0Balance"`
		Token1Balance  uint64      `json:"Token1Balance"`
		TradeDirection byte        `json:"TradeDirection"`
		Receiver       [2]string   `json:"Receiver"`
	}{
		Id:             o.id,
		NftID:          o.nftID,
		AccessOTA:      o.accessOTA,
		Token0Rate:     o.token0Rate,
		Token1Rate:     o.token1Rate,
		Token0Balance:  o.token0Balance,
		Token1Balance:  o.token1Balance,
		TradeDirection: o.tradeDirection,
		Receiver:       o.receiver,
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
		AccessOTA      []byte      `json:"AccessOTA,omitempty"`
		Token0Rate     uint64      `json:"Token0Rate"`
		Token1Rate     uint64      `json:"Token1Rate"`
		Token0Balance  uint64      `json:"Token0Balance"`
		Token1Balance  uint64      `json:"Token1Balance"`
		TradeDirection byte        `json:"TradeDirection"`
		Receiver       [2]string   `json:"Receiver"`
	}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	*o = Pdexv3Order{
		id:             temp.Id,
		nftID:          temp.NftID,
		accessOTA:      temp.AccessOTA,
		token0Rate:     temp.Token0Rate,
		token1Rate:     temp.Token1Rate,
		token0Balance:  temp.Token0Balance,
		token1Balance:  temp.Token1Balance,
		tradeDirection: temp.TradeDirection,
		receiver:       temp.Receiver,
	}
	return nil
}

func (o *Pdexv3Order) Clone() *Pdexv3Order {
	return NewPdexv3OrderWithValue(o.id, o.nftID, o.accessOTA, o.token0Rate, o.token1Rate,
		o.token0Balance, o.token1Balance, o.tradeDirection, o.receiver)
}

func (o *Pdexv3Order) IsEmpty() bool {
	return o.token0Balance == 0 && o.token1Balance == 0
}
