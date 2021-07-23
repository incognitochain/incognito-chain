package v3utils

import (
	"fmt"
	"math/big"
)

const (
	TradeDirectionSell0 = iota
	TradeDirectionSell1
)

type PoolReserve struct {
	Token0        uint64
	Token1        uint64
	Token0Virtual uint64
	Token1Virtual uint64
}

// BuyAmount() computes the output amount given input, based on reserve amounts. Deduct fees before calling this
func (pr PoolReserve) BuyAmount(sellAmount uint64, tradeDirection int) (uint64, error) {
	if tradeDirection == TradeDirectionSell0 {
		return calculateBuyAmount(sellAmount, pr.Token0, pr.Token1, pr.Token0Virtual, pr.Token1Virtual, 0)
	} else {
		return calculateBuyAmount(sellAmount, pr.Token1, pr.Token0, pr.Token1Virtual, pr.Token0Virtual, 0)
	}
}

// SwapToReachOrderRate() does a *partial* swap using liquidity in the pool, such that the price afterwards does not exceed an order's rate
// It returns an error when the pool runs out of liquidity
// Upon success, it updates the reserve values and returns (buyAmount, sellAmountRemain, token0Change, token1Change)
func (pr *PoolReserve) SwapToReachOrderRate(maxSellAmountAfterFee uint64, tradeDirection int, ord OrderMatchingInfo) (uint64, uint64, *big.Int, *big.Int, error) {
	if tradeDirection == ord.TradeDirection {
		return 0, 0, nil, nil, fmt.Errorf("Cannot match trade with order of same direction")
	}
	token0Change := big.NewInt(0)
	token1Change := big.NewInt(0)

	maxDeltaX := big.NewInt(0).SetUint64(maxSellAmountAfterFee)
	xV := big.NewInt(0).SetUint64(pr.Token0Virtual)
	yV := big.NewInt(0).SetUint64(pr.Token1Virtual)
	xOrd := big.NewInt(0).SetUint64(ord.Token0Rate)
	yOrd := big.NewInt(0).SetUint64(ord.Token1Rate)
	L := big.NewInt(0).Mul(xV, yV)

	targetDeltaX := big.NewInt(0).Mul(L, xOrd)
	targetDeltaX.Div(targetDeltaX, yOrd)
	targetDeltaX.Sqrt(targetDeltaX)
	targetDeltaX.Sub(targetDeltaX, xV)

	// pool price already reached (or surpassed) order rate -> exit
	if targetDeltaX.Cmp(big.NewInt(0)) <= 0 {
		return 0, maxSellAmountAfterFee, token0Change, token1Change, nil
	}

	var finalSellAmount, sellAmountRemain uint64
	// able to trade fully in pool before reaching order rate
	if targetDeltaX.Cmp(maxDeltaX) >= 0 {
		finalSellAmount = maxSellAmountAfterFee
		sellAmountRemain = 0
	} else {
		// only swap the target delta x
		// maxDeltaX is valid uint64, while 0 < targetDeltaX < maxDeltaX
		finalSellAmount = targetDeltaX.Uint64()
		sellAmountRemain = big.NewInt(0).Sub(maxDeltaX, targetDeltaX).Uint64()
	}
	buyAmount, err := pr.BuyAmount(finalSellAmount, tradeDirection)
	if err != nil {
		return 0, 0, nil, nil, err
	}

	if tradeDirection == TradeDirectionSell0 {
		token0Change.SetUint64(finalSellAmount)
		token1Change.SetUint64(buyAmount)
		token1Change.Neg(token1Change)
	} else {
		token1Change.SetUint64(finalSellAmount)
		token0Change.SetUint64(buyAmount)
		token0Change.Neg(token0Change)
	}
	err = pr.ApplyReserveChanges(token0Change, token1Change)
	if err != nil {
		return 0, 0, nil, nil, err
	}

	return buyAmount, sellAmountRemain, token0Change, token1Change, err
}

func (pr *PoolReserve) ApplyReserveChanges(change0, change1 *big.Int) error {
	// sign check : changes must have opposite signs or both be zero
	if change0.Sign()*change1.Sign() >= 0 {
		if !(change0.Sign() == 0 && change1.Sign() == 0) {
			return fmt.Errorf("Invalid signs for reserve changes")
		}
	}

	resv := big.NewInt(0).SetUint64(pr.Token0)
	temp := big.NewInt(0).Add(resv, change0)
	if temp.Cmp(big.NewInt(0)) == -1 {
		return fmt.Errorf("Not enough liquidity for trade")
	}
	pr.Token0 = temp.Uint64()

	resv.SetUint64(pr.Token0Virtual)
	temp.Add(resv, change0)
	if !temp.IsUint64() {
		return fmt.Errorf("Cannot set reserve out of uint64 range")
	}
	pr.Token0Virtual = temp.Uint64()

	resv.SetUint64(pr.Token1Virtual)
	temp.Add(resv, change1)
	if temp.Cmp(big.NewInt(0)) == -1 {
		return fmt.Errorf("Not enough liquidity for trade")
	}
	pr.Token1 = temp.Uint64()

	resv.SetUint64(pr.Token1Virtual)
	temp.Add(resv, change1)
	if !temp.IsUint64() {
		return fmt.Errorf("Cannot set reserve out of uint64 range")
	}
	pr.Token1Virtual = temp.Uint64()

	return nil
}

func AcceptOrRefundTrade(amountIn uint64, reserves []*PoolReserve, tradeDirections []int, orderbooks []OrderBookIterator) (bool, []string, []*PoolReserve, error) {
	mutualLen := len(reserves)
	if len(tradeDirections) != mutualLen || len(orderbooks) != mutualLen {
		return false, nil, nil, nil
	}
	return false, nil, nil, nil
}
