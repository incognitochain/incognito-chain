package v3utils

import (
	"fmt"
	"math/big"
)

type OrderMatchingInfo struct {
	Token0Rate     uint64
	Token1Rate     uint64
	Token0Balance  uint64
	Token1Balance  uint64
	TradeDirection int
}

type OrderBookIterator interface {
	NextOrder(tradeDirection int) (*OrderMatchingInfo, string, error)
}


// BuyAmount() computes the theoretical (by rate) output amount given incoming trade's sell amount. Order balance only needs to be non-zero since it can be partially matched
func (ordInf OrderMatchingInfo) BuyAmountFromOrder(incomingTradeSellAmount uint64, incomingTradeDirection int) (uint64, error) {
	if ordInf.TradeDirection == incomingTradeDirection {
		return 0, fmt.Errorf("Cannot match order with trade of same direction")
	}
	// sell / buy rates from incoming trade's view
	var sellRate, buyRate *big.Int
	switch incomingTradeDirection {
	case TradeDirectionSell0:
		sellRate = big.NewInt(0).SetUint64(ordInf.Token0Rate)
		buyRate = big.NewInt(0).SetUint64(ordInf.Token1Rate)
	case TradeDirectionSell1:
		sellRate = big.NewInt(0).SetUint64(ordInf.Token1Rate)
		buyRate = big.NewInt(0).SetUint64(ordInf.Token0Rate)
	default:
		return 0, fmt.Errorf("Invalid trade direction %d", incomingTradeDirection)
	}
	amount := big.NewInt(0).SetUint64(incomingTradeSellAmount)
	num := big.NewInt(0).Mul(amount, buyRate)
	result := num.Div(num, sellRate)
	if !result.IsUint64() {
		return 0, fmt.Errorf("Buy-from-order amount out of uint64 range")
	}
	return result.Uint64(), nil
}

func (ordInf OrderMatchingInfo) SellAmountToOrder(incomingTradeBuyAmount uint64, incomingTradeDirection int) (uint64, error) {
	if ordInf.TradeDirection == incomingTradeDirection {
		return 0, fmt.Errorf("Cannot match order with trade of same direction")
	}
	// sell / buy rates from incoming trade's view
	var sellRate, buyRate *big.Int
	var maxBuyingAmount uint64
	switch incomingTradeDirection {
	case TradeDirectionSell0:
		sellRate = big.NewInt(0).SetUint64(ordInf.Token0Rate)
		buyRate = big.NewInt(0).SetUint64(ordInf.Token1Rate)
		maxBuyingAmount = ordInf.Token1Balance
	case TradeDirectionSell1:
		sellRate = big.NewInt(0).SetUint64(ordInf.Token1Rate)
		buyRate = big.NewInt(0).SetUint64(ordInf.Token0Rate)
		maxBuyingAmount = ordInf.Token0Balance
	default:
		return 0, fmt.Errorf("Invalid trade direction %d", incomingTradeDirection)
	}
	if maxBuyingAmount < incomingTradeBuyAmount {
		return 0, fmt.Errorf("Insufficient order balance for swap")
	}
	amount := big.NewInt(0).SetUint64(incomingTradeBuyAmount)
	num := big.NewInt(0).Mul(amount, sellRate)
	result := num.Div(num, buyRate)
	if !result.IsUint64() {
		return 0, fmt.Errorf("Sell-to-order amount out of uint64 range")
	}
	return result.Uint64(), nil
}

// MatchOwnRate() swaps this order's selling balance with token sold in a trade, following this order's rate
func (ordInf *OrderMatchingInfo) MatchOwnRate(maxSellAmountAfterFee uint64, tradeDirection int) (uint64, uint64, *big.Int, *big.Int, error) {
	finalSellAmount := maxSellAmountAfterFee
	sellAmountRemain := uint64(0)
	buyAmount, err := ordInf.BuyAmountFromOrder(maxSellAmountAfterFee, tradeDirection)
	if err != nil {
		return 0, 0, nil, nil, err
	}
	var maxBuyingAmount uint64
	if tradeDirection == TradeDirectionSell0 {
		maxBuyingAmount = ordInf.Token1Rate
	} else {
		maxBuyingAmount = ordInf.Token0Rate
	}
	if maxBuyingAmount < buyAmount {
		buyAmount = maxBuyingAmount
		finalSellAmount, err = ordInf.SellAmountToOrder(buyAmount, tradeDirection)
		if err != nil {
			return 0, 0, nil, nil, err
		}
		if maxSellAmountAfterFee > finalSellAmount {
			sellAmountRemain = maxSellAmountAfterFee - finalSellAmount
		} else {
			return 0, 0, nil, nil, fmt.Errorf("Final sell amount %d exceeds maximum %d", finalSellAmount, maxSellAmountAfterFee)
		}
	}

	var token0Change, token1Change *big.Int
	// sell / buy are from incoming trade's view
	if tradeDirection == TradeDirectionSell0 {
		token0Change.SetUint64(finalSellAmount)
		token1Change.SetUint64(buyAmount)
		token1Change.Neg(token1Change)
	} else {
		token1Change.SetUint64(finalSellAmount)
		token0Change.SetUint64(buyAmount)
		token0Change.Neg(token0Change)
	}
	err = ordInf.ApplyBalanceChanges(token0Change, token1Change)
	if err != nil {
		return 0, 0, nil, nil, err
	}

	return buyAmount, sellAmountRemain, token0Change, token1Change, nil
}

// Match() is the default order matcher. It uses pool price instead of this order's rate
func (ordInf *OrderMatchingInfo) Match(maxSellAmountAfterFee uint64, tradeDirection int, pr PoolReserve) (uint64, uint64, *big.Int, *big.Int, error) {
	finalSellAmount := maxSellAmountAfterFee
	sellAmountRemain := uint64(0)
	buyAmount, err := pr.BuyAmount(maxSellAmountAfterFee, tradeDirection)
	if err != nil {
		return 0, 0, nil, nil, err
	}
	var maxBuyingAmount uint64
	if tradeDirection == TradeDirectionSell0 {
		maxBuyingAmount = ordInf.Token1Rate
	} else {
		maxBuyingAmount = ordInf.Token0Rate
	}
	if maxBuyingAmount < buyAmount {
		buyAmount = maxBuyingAmount
		finalSellAmount, err = pr.AmountToSell(buyAmount, tradeDirection)
		if err != nil {
			return 0, 0, nil, nil, err
		}
		if maxSellAmountAfterFee > finalSellAmount {
			sellAmountRemain = maxSellAmountAfterFee - finalSellAmount
		} else {
			return 0, 0, nil, nil, fmt.Errorf("Final sell amount %d exceeds maximum %d", finalSellAmount, maxSellAmountAfterFee)
		}
	}

	var token0Change, token1Change *big.Int
	// sell / buy are from incoming trade's view
	if tradeDirection == TradeDirectionSell0 {
		token0Change.SetUint64(finalSellAmount)
		token1Change.SetUint64(buyAmount)
		token1Change.Neg(token1Change)
	} else {
		token1Change.SetUint64(finalSellAmount)
		token0Change.SetUint64(buyAmount)
		token0Change.Neg(token0Change)
	}
	err = ordInf.ApplyBalanceChanges(token0Change, token1Change)
	if err != nil {
		return 0, 0, nil, nil, err
	}

	return buyAmount, sellAmountRemain, token0Change, token1Change, nil
}

func (ordInf *OrderMatchingInfo) ApplyBalanceChanges(change0, change1 *big.Int) error {
	// sign check : changes must have opposite signs or both be zero
	if change0.Sign()*change1.Sign() >= 0 {
		if !(change0.Sign() == 0 && change1.Sign() == 0) {
			return fmt.Errorf("Invalid signs for reserve changes")
		}
	}

	balance := big.NewInt(0).SetUint64(ordInf.Token0Balance)
	temp := big.NewInt(0).Add(balance, change0)
	if temp.Cmp(big.NewInt(0)) == -1 {
		return fmt.Errorf("Not enough liquidity for trade")
	} else if !temp.IsUint64() {
		return fmt.Errorf("Cannot set token0 balance out of uint64 range")
	}
	ordInf.Token0Balance = temp.Uint64()

	balance.SetUint64(ordInf.Token1Balance)
	temp.Add(balance, change1)
	if temp.Cmp(big.NewInt(0)) == -1 {
		return fmt.Errorf("Not enough liquidity for trade")
	} else if !temp.IsUint64() {
		return fmt.Errorf("Cannot set token1 balance out of uint64 range")
	}
	ordInf.Token1Balance = temp.Uint64()

	return nil
}
