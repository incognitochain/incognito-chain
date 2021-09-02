package v2utils

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
)

const (
	TradeDirectionSell0 = iota
	TradeDirectionSell1
)

type TradingPair struct {
	*rawdbv2.Pdexv3PoolPair
}

func (tp *TradingPair) UnmarshalJSON(data []byte) error {
	tp.Pdexv3PoolPair = &rawdbv2.Pdexv3PoolPair{}
	return json.Unmarshal(data, tp.Pdexv3PoolPair)
}

// BuyAmount() computes the output amount given input, based on reserve amounts. Deduct fees before calling this
func (tp TradingPair) BuyAmount(sellAmount uint64, tradeDirection byte) (uint64, error) {
	if tradeDirection == TradeDirectionSell0 {
		return calculateBuyAmount(sellAmount, tp.Token0RealAmount(), tp.Token1RealAmount(), tp.Token0VirtualAmount(), tp.Token1VirtualAmount())
	} else {
		return calculateBuyAmount(sellAmount, tp.Token1RealAmount(), tp.Token0RealAmount(), tp.Token1VirtualAmount(), tp.Token0VirtualAmount())
	}
}

// BuyAmount() computes the input amount given output, based on reserve amounts
func (tp TradingPair) AmountToSell(buyAmount uint64, tradeDirection byte) (uint64, error) {
	if tradeDirection == TradeDirectionSell0 {
		return calculateAmountToSell(buyAmount, tp.Token0RealAmount(), tp.Token1RealAmount(), tp.Token0VirtualAmount(), tp.Token1VirtualAmount())
	} else {
		return calculateAmountToSell(buyAmount, tp.Token1RealAmount(), tp.Token0RealAmount(), tp.Token1VirtualAmount(), tp.Token0VirtualAmount())
	}
}

// SwapToReachOrderRate() does a *partial* swap using liquidity in the pool, such that the price afterwards does not exceed an order's rate
// It returns an error when the pool runs out of liquidity
// Upon success, it updates the reserve values and returns (buyAmount, sellAmountRemain, token0Change, token1Change)
func (tp *TradingPair) SwapToReachOrderRate(maxSellAmountAfterFee uint64, tradeDirection byte, ord *MatchingOrder) (uint64, uint64, *big.Int, *big.Int, error) {
	token0Change := big.NewInt(0)
	token1Change := big.NewInt(0)

	maxDeltaX := big.NewInt(0).SetUint64(maxSellAmountAfterFee)

	// x, y represent selling & buying reserves, respectively
	var xV, yV *big.Int
	switch tradeDirection {
	case TradeDirectionSell0:
		xV = big.NewInt(0).Set(tp.Token0VirtualAmount())
		yV = big.NewInt(0).Set(tp.Token1VirtualAmount())
	case TradeDirectionSell1:
		xV = big.NewInt(0).Set(tp.Token1VirtualAmount())
		yV = big.NewInt(0).Set(tp.Token0VirtualAmount())
	}

	var xOrd, yOrd, L, targetDeltaX *big.Int
	if ord != nil {
		if tradeDirection == ord.TradeDirection() {
			return 0, 0, nil, nil, fmt.Errorf("Cannot match trade with order of same direction")
		}
		if tradeDirection == TradeDirectionSell0 {
			xOrd = big.NewInt(0).SetUint64(ord.Token0Rate())
			yOrd = big.NewInt(0).SetUint64(ord.Token1Rate())
		} else {
			xOrd = big.NewInt(0).SetUint64(ord.Token1Rate())
			yOrd = big.NewInt(0).SetUint64(ord.Token0Rate())
		}
		L = big.NewInt(0).Mul(xV, yV)

		targetDeltaX = big.NewInt(0).Mul(L, xOrd)
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
	buyAmount, err := tp.BuyAmount(finalSellAmount, tradeDirection)
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
	err = tp.ApplyReserveChanges(token0Change, token1Change)
	if err != nil {
		return 0, 0, nil, nil, err
	}

	return buyAmount, sellAmountRemain, token0Change, token1Change, err
}

func (tp *TradingPair) ApplyReserveChanges(change0, change1 *big.Int) error {
	// sign check : changes must have opposite signs or both be zero
	if change0.Sign()*change1.Sign() >= 0 {
		if !(change0.Sign() == 0 && change1.Sign() == 0) {
			return fmt.Errorf("Invalid signs for reserve changes")
		}
	}

	resv := big.NewInt(0).SetUint64(tp.Token0RealAmount())
	temp := big.NewInt(0).Add(resv, change0)
	if temp.Cmp(big.NewInt(0)) == -1 {
		return fmt.Errorf("Not enough token0 liquidity for trade")
	}
	if !temp.IsUint64() {
		return fmt.Errorf("Cannot set real token0 reserve out of uint64 range")
	}
	tp.SetToken0RealAmount(temp.Uint64())

	resv.Set(tp.Token0VirtualAmount())
	temp.Add(resv, change0)
	tp.SetToken0VirtualAmount(big.NewInt(0).Set(temp))

	resv.SetUint64(tp.Token1RealAmount())
	temp.Add(resv, change1)
	if temp.Cmp(big.NewInt(0)) == -1 {
		return fmt.Errorf("Not enough token1 liquidity for trade")
	}
	if !temp.IsUint64() {
		return fmt.Errorf("Cannot set real token1 reserve out of uint64 range")
	}
	tp.SetToken1RealAmount(temp.Uint64())

	resv.Set(tp.Token1VirtualAmount())
	temp.Add(resv, change1)
	tp.SetToken1VirtualAmount(big.NewInt(0).Set(temp))

	return nil
}

// MaybeAcceptTrade() performs a trade determined by input amount, path, directions & order book state. Upon success, state changes are applied in memory & collected in an instruction.
// A returned error means the trade is refunded
func MaybeAcceptTrade(amountIn, fee uint64, tradePath []string, receiver privacy.OTAReceiver, 
	reserves []*rawdbv2.Pdexv3PoolPair, tradeDirections []byte, 
	tokenToBuy common.Hash, minAmount uint64, orderbooks []OrderBookIterator,
) (*metadataPdexv3.AcceptedTrade, []*rawdbv2.Pdexv3PoolPair, error) {
	mutualLen := len(reserves)
	if len(tradeDirections) != mutualLen || len(orderbooks) != mutualLen {
		return nil, nil, fmt.Errorf("Trade path vs directions vs orderbooks length mismatch")
	}
	if amountIn < fee {
		return nil, nil, fmt.Errorf("Trade input insufficient for trading fee")
	}
	sellAmountRemain := amountIn - fee
	acceptedMeta := metadataPdexv3.AcceptedTrade{
		Receiver:     receiver,
		TradePath:    tradePath,
		PairChanges:  make([][2]*big.Int, mutualLen),
		OrderChanges: make([]map[string][2]*big.Int, mutualLen),
		TokenToBuy:   tokenToBuy,
	}

	var totalBuyAmount uint64
	for i := 0; i < mutualLen; i++ {
		acceptedMeta.OrderChanges[i] = make(map[string][2]*big.Int)

		accumulatedToken0Change := big.NewInt(0)
		accumulatedToken1Change := big.NewInt(0)
		totalBuyAmount = uint64(0)

		for order, ordID, err := orderbooks[i].NextOrder(tradeDirections[i]); err == nil; order, ordID, err = orderbooks[i].NextOrder(tradeDirections[i]) {
			buyAmount, temp, token0Change, token1Change, err := (&TradingPair{reserves[i]}).SwapToReachOrderRate(sellAmountRemain, tradeDirections[i], order)
			if err != nil {
				return nil, nil, err
			}
			sellAmountRemain = temp
			if totalBuyAmount+buyAmount < totalBuyAmount {
				return nil, nil, fmt.Errorf("Sum exceeds uint64 range after swapping in pool")
			}
			totalBuyAmount += buyAmount
			accumulatedToken0Change.Add(accumulatedToken0Change, token0Change)
			accumulatedToken1Change.Add(accumulatedToken1Change, token1Change)
			if sellAmountRemain == 0 {
				break
			}
			if order != nil {
				buyAmount, temp, token0Change, token1Change, err := order.Match(sellAmountRemain, tradeDirections[i])
				if err != nil {
					return nil, nil, err
				}
				sellAmountRemain = temp
				if totalBuyAmount+buyAmount < totalBuyAmount {
					return nil, nil, fmt.Errorf("Sum exceeds uint64 range after matching order")
				}
				totalBuyAmount += buyAmount
				// add order balance changes to "accepted" instruction
				acceptedMeta.OrderChanges[i][ordID] = [2]*big.Int{token0Change, token1Change}
				if sellAmountRemain == 0 {
					break
				}
			}
		}

		// add pair changes to "accepted" instruction
		acceptedMeta.PairChanges[i] = [2]*big.Int{accumulatedToken0Change, accumulatedToken1Change}
		// set sell amount before moving on to next pair
		sellAmountRemain = totalBuyAmount
	}

	if totalBuyAmount < minAmount {
		return nil, nil, fmt.Errorf("Min acceptable amount %d not reached - trade output %d", minAmount, totalBuyAmount)
	}
	acceptedMeta.Amount = totalBuyAmount
	return &acceptedMeta, reserves, nil
}