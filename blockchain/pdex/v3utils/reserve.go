package v3utils

import (
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/privacy"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
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

// BuyAmount() computes the input amount given output, based on reserve amounts
func (pr PoolReserve) AmountToSell(buyAmount uint64, tradeDirection int) (uint64, error) {
	if tradeDirection == TradeDirectionSell0 {
		return calculateAmountToSell(buyAmount, pr.Token0, pr.Token1, pr.Token0Virtual, pr.Token1Virtual, 0)
	} else {
		return calculateAmountToSell(buyAmount, pr.Token1, pr.Token0, pr.Token1Virtual, pr.Token0Virtual, 0)
	}
}

// SwapToReachOrderRate() does a *partial* swap using liquidity in the pool, such that the price afterwards does not exceed an order's rate
// It returns an error when the pool runs out of liquidity
// Upon success, it updates the reserve values and returns (buyAmount, sellAmountRemain, token0Change, token1Change)
func (pr *PoolReserve) SwapToReachOrderRate(maxSellAmountAfterFee uint64, tradeDirection int, ord *OrderMatchingInfo) (uint64, uint64, *big.Int, *big.Int, error) {
	if tradeDirection == ord.TradeDirection {
		return 0, 0, nil, nil, fmt.Errorf("Cannot match trade with order of same direction")
	}
	token0Change := big.NewInt(0)
	token1Change := big.NewInt(0)

	maxDeltaX := big.NewInt(0).SetUint64(maxSellAmountAfterFee)

	// x, y represent selling & buying reserves, respectively
	var xV, yV *big.Int
	switch tradeDirection {
	case TradeDirectionSell0:
		xV = big.NewInt(0).SetUint64(pr.Token0Virtual)
		yV = big.NewInt(0).SetUint64(pr.Token1Virtual)
	case TradeDirectionSell1:
		xV = big.NewInt(0).SetUint64(pr.Token1Virtual)
		yV = big.NewInt(0).SetUint64(pr.Token0Virtual)
	}

	var xOrd, yOrd, L, targetDeltaX *big.Int
	if ord != nil {
		if tradeDirection == TradeDirectionSell0 {
			xOrd = big.NewInt(0).SetUint64(ord.Token0Rate)
			yOrd = big.NewInt(0).SetUint64(ord.Token1Rate)
		} else {
			xOrd = big.NewInt(0).SetUint64(ord.Token1Rate)
			yOrd = big.NewInt(0).SetUint64(ord.Token0Rate)
		}
		L = big.NewInt(0).Mul(xV, yV)

		targetDeltaX := big.NewInt(0).Mul(L, xOrd)
		targetDeltaX.Div(targetDeltaX, yOrd)
		targetDeltaX.Sqrt(targetDeltaX)
		targetDeltaX.Sub(targetDeltaX, xV)
	}

	var finalSellAmount, sellAmountRemain uint64
	if ord == nil || targetDeltaX.Cmp(maxDeltaX) >= 0 {
		// able to trade fully in pool before reaching order rate
		finalSellAmount = maxSellAmountAfterFee
		sellAmountRemain = 0
	} else if targetDeltaX.Cmp(big.NewInt(0)) <= 0 {
		// pool price already reached (or surpassed) order rate -> exit
		return 0, maxSellAmountAfterFee, token0Change, token1Change, nil
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

	resv.SetUint64(pr.Token1)
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

// MaybeAcceptTrade() performs a trade determined by input amount, path, directions & order book state. Upon success, state changes are applied in memory & collected in an instruction.
// A returned error means the trade is refunded
func MaybeAcceptTrade(amountIn, fee uint64, receiver privacy.OTAReceiver, reserves []*PoolReserve, tradeDirections []int, orderbooks []OrderBookIterator) ([]string, []*PoolReserve, error) {
	mutualLen := len(reserves)
	if len(tradeDirections) != mutualLen || len(orderbooks) != mutualLen {
		return nil, nil, fmt.Errorf("Trade path vs directions vs orderbooks length mismatch")
	}
	if amountIn > fee {
		return nil, nil, fmt.Errorf("Trade input insufficient for trading fee")
	}
	sellAmountRemain := amountIn - fee
	acceptedMeta := metadataPdexv3.AcceptedTrade{
		Receiver:     receiver,
		PairChanges:  make([][2]big.Int, mutualLen),
		OrderChanges: make([]map[string][2]big.Int, mutualLen),
	}

	for i := 0; i < mutualLen; i++ {
		acceptedMeta.OrderChanges[i] = make(map[string][2]big.Int)

		accumulatedToken0Change := big.NewInt(0)
		accumulatedToken1Change := big.NewInt(0)
		var order *OrderMatchingInfo
		var ordID string
		var err error
		totalBuyAmount := uint64(0)

		for order, ordID, err = orderbooks[i].NextOrder(tradeDirections[i]); err == nil; {
			buyAmount, temp, token0Change, token1Change, err := reserves[i].SwapToReachOrderRate(sellAmountRemain, tradeDirections[i], order)
			if err != nil {
				return nil, nil, err
			}
			sellAmountRemain = temp
			if totalBuyAmount+buyAmount < totalBuyAmount {
				return nil, nil, fmt.Errorf("Sum exceeds uint64 range")
			}
			totalBuyAmount += buyAmount
			accumulatedToken0Change.Add(accumulatedToken0Change, token0Change)
			accumulatedToken1Change.Add(accumulatedToken1Change, token1Change)
			if sellAmountRemain == 0 {
				break
			}
			if order != nil {
				buyAmount, temp, token0Change, token1Change, err := order.Match(sellAmountRemain, tradeDirections[i], *reserves[i])
				if err != nil {
					return nil, nil, err
				}
				sellAmountRemain = temp
				if totalBuyAmount+buyAmount < totalBuyAmount {
					return nil, nil, fmt.Errorf("Sum exceeds uint64 range")
				}
				totalBuyAmount += buyAmount
				// add order balance changes to "accepted" instruction
				acceptedMeta.OrderChanges[i][ordID] = [2]big.Int{*token0Change, *token1Change}
				if sellAmountRemain == 0 {
					acn := &instruction.Action{Content: acceptedMeta}
					return acn.Strings(), reserves, nil
				}
			}
		}

		// add pair changes to "accepted" instruction
		acceptedMeta.PairChanges[i] = [2]big.Int{*accumulatedToken0Change, *accumulatedToken1Change}
		// set sell amount before moving on to next pair
		sellAmountRemain = totalBuyAmount
	}
	return nil, nil, fmt.Errorf("Trade handling ended unexpectedly")
}
