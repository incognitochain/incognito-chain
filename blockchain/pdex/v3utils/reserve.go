package v3utils

import ()

type OrderBookIterator interface {
	NextOrder(isSellingToken0 bool) (sellAmount uint64, minBuyAmount uint64, err error)
}

type PoolReserve struct {
	SellToken        uint64
	BuyToken         uint64
	SellTokenVirtual uint64
	BuyTokenVirtual  uint64
}

func AcceptOrRefundTrade(amountIn uint64, reserves []PoolReserve, orderbooks []OrderBookIterator) (bool, []string, []PoolReserve, error)  {
	return false, nil, nil, nil
}
