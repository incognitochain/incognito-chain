package v3utils

import(

)

type OrderMatchingInfo struct {
	Token0Rate       uint64
	Token1Rate       uint64
	Token0Balance    uint64
	Token1Balance    uint64
	TradeDirection int
}

// Quote() computes the theoretical (by rate) output amount given input
func (ordInf OrderMatchingInfo) Quote(){}

// MatchOwnRate() swaps this order's selling balance with token sold in a trade, following this order's rate
func (ordInf *OrderMatchingInfo) MatchOwnRate(){}

// Match() is the default order matcher. It uses pool price instead of this order's rate
func (ordInf *OrderMatchingInfo) Match(){}

type OrderBookIterator interface {
	NextOrder(tradeDirection int) (OrderMatchingInfo, error)
}
