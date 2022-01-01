package v2utils

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

type MatchingOrder struct {
	*rawdbv2.Pdexv3Order
}

func (order *MatchingOrder) UnmarshalJSON(data []byte) error {
	order.Pdexv3Order = &rawdbv2.Pdexv3Order{}
	return json.Unmarshal(data, order.Pdexv3Order)
}

type OrderBookIterator interface {
	NextOrder(tradeDirection byte) (*MatchingOrder, string, error)
	NftIDs() map[string]string
}

// BuyAmountFromOrder() computes the theoretical (by rate) output amount given incoming trade's sell amount. Order balance only needs to be non-zero since it can be partially matched
func (order MatchingOrder) BuyAmountFromOrder(incomingTradeSellAmount uint64, incomingTradeDirection byte) (uint64, error) {
	if order.TradeDirection() == incomingTradeDirection {
		return 0, fmt.Errorf("Cannot match order with trade of same direction")
	}
	// sell / buy rates from incoming trade's view
	var sellRate, buyRate *big.Int
	switch incomingTradeDirection {
	case TradeDirectionSell0:
		sellRate = big.NewInt(0).SetUint64(order.Token0Rate())
		buyRate = big.NewInt(0).SetUint64(order.Token1Rate())
	case TradeDirectionSell1:
		sellRate = big.NewInt(0).SetUint64(order.Token1Rate())
		buyRate = big.NewInt(0).SetUint64(order.Token0Rate())
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

func (order MatchingOrder) SellAmountToOrder(incomingTradeBuyAmount uint64, incomingTradeDirection byte) (uint64, error) {
	if order.TradeDirection() == incomingTradeDirection {
		return 0, fmt.Errorf("Cannot match order with trade of same direction")
	}
	// sell / buy rates from incoming trade's view
	var sellRate, buyRate *big.Int
	var maxBuyingAmount uint64
	switch incomingTradeDirection {
	case TradeDirectionSell0:
		sellRate = big.NewInt(0).SetUint64(order.Token0Rate())
		buyRate = big.NewInt(0).SetUint64(order.Token1Rate())
		maxBuyingAmount = order.Token1Balance()
	case TradeDirectionSell1:
		sellRate = big.NewInt(0).SetUint64(order.Token1Rate())
		buyRate = big.NewInt(0).SetUint64(order.Token0Rate())
		maxBuyingAmount = order.Token0Balance()
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

// Match() swaps this order's selling balance with token sold in a trade, following this order's rate
func (order *MatchingOrder) Match(maxSellAmountAfterFee uint64, tradeDirection byte) (uint64, uint64, *big.Int, *big.Int, error) {
	finalSellAmount := maxSellAmountAfterFee
	sellAmountRemain := uint64(0)
	buyAmount, err := order.BuyAmountFromOrder(maxSellAmountAfterFee, tradeDirection)
	if err != nil {
		return 0, 0, nil, nil, err
	}
	var maxBuyingAmount uint64
	if tradeDirection == TradeDirectionSell0 {
		maxBuyingAmount = order.Token1Balance()
	} else {
		maxBuyingAmount = order.Token0Balance()
	}
	if maxBuyingAmount < buyAmount {
		buyAmount = maxBuyingAmount
		finalSellAmount, err = order.SellAmountToOrder(buyAmount, tradeDirection)
		if err != nil {
			return 0, 0, nil, nil, err
		}
		if maxSellAmountAfterFee > finalSellAmount {
			sellAmountRemain = maxSellAmountAfterFee - finalSellAmount
		} else {
			return 0, 0, nil, nil, fmt.Errorf("Final sell amount %d exceeds maximum %d", finalSellAmount, maxSellAmountAfterFee)
		}
	}

	token0Change := big.NewInt(0)
	token1Change := big.NewInt(0)
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
	err = order.ApplyBalanceChanges(token0Change, token1Change)
	if err != nil {
		return 0, 0, nil, nil, err
	}

	return buyAmount, sellAmountRemain, token0Change, token1Change, nil
}

// MatchPoolAmount() uses pool values instead of this order's rate
func (order *MatchingOrder) MatchPoolAmount(maxSellAmountAfterFee uint64, tradeDirection byte, pr TradingPair) (uint64, uint64, *big.Int, *big.Int, error) {
	finalSellAmount := maxSellAmountAfterFee
	sellAmountRemain := uint64(0)
	buyAmount, err := pr.BuyAmount(maxSellAmountAfterFee, tradeDirection)
	if err != nil {
		return 0, 0, nil, nil, err
	}
	var maxBuyingAmount uint64
	if tradeDirection == TradeDirectionSell0 {
		maxBuyingAmount = order.Token1Balance()
	} else {
		maxBuyingAmount = order.Token0Balance()
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

	token0Change := big.NewInt(0)
	token1Change := big.NewInt(0)
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
	err = order.ApplyBalanceChanges(token0Change, token1Change)
	if err != nil {
		return 0, 0, nil, nil, err
	}

	return buyAmount, sellAmountRemain, token0Change, token1Change, nil
}

func (order *MatchingOrder) ApplyBalanceChanges(change0, change1 *big.Int) error {
	// sign check : changes must have opposite signs or both be zero
	if change0.Sign()*change1.Sign() >= 0 {
		if !(change0.Sign() == 0 && change1.Sign() == 0) {
			return fmt.Errorf("Invalid signs for order changes %v, %v", change0, change1)
		}
	}

	balance := big.NewInt(0).SetUint64(order.Token0Balance())
	temp := big.NewInt(0).Add(balance, change0)
	if temp.Cmp(big.NewInt(0)) == -1 {
		return fmt.Errorf("Not enough token0 balance in order for trade")
	} else if !temp.IsUint64() {
		return fmt.Errorf("Cannot set token0 balance out of uint64 range")
	}
	order.SetToken0Balance(temp.Uint64())

	balance.SetUint64(order.Token1Balance())
	temp.Add(balance, change1)
	if temp.Cmp(big.NewInt(0)) == -1 {
		return fmt.Errorf("Not enough token1 balance in order for trade")
	} else if !temp.IsUint64() {
		return fmt.Errorf("Cannot set token1 balance out of uint64 range")
	}
	order.SetToken1Balance(temp.Uint64())

	return nil
}

// CanMatch() returns true if
// - incoming trade is of opposite direction (sell0 vs sell1)
// - outstanding order balance is not too small (would exchange for at least 1 unit of TokenBuy using this order's own rate)
func (order *MatchingOrder) CanMatch(incomingTradeDirection byte) (bool, error) {
	if order.TradeDirection() == incomingTradeDirection {
		return false, nil
	}

	var amountBuyAllFromOrder uint64
	switch incomingTradeDirection {
	case TradeDirectionSell0:
		amountBuyAllFromOrder = order.Token1Balance()
	case TradeDirectionSell1:
		amountBuyAllFromOrder = order.Token0Balance()
	}

	sellAmountToOrder, err := order.SellAmountToOrder(amountBuyAllFromOrder, incomingTradeDirection)
	return sellAmountToOrder >= 1, err
}
