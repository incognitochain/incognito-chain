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

func NewTradingPair() *TradingPair {
	return &TradingPair{
		Pdexv3PoolPair: rawdbv2.NewPdexv3PoolPair(),
	}
}

func NewTradingPairWithValue(
	reserve *rawdbv2.Pdexv3PoolPair,
) *TradingPair {
	return &TradingPair{
		Pdexv3PoolPair: reserve,
	}
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

	if HasInsufficientLiquidity(*tp.Pdexv3PoolPair) {
		return 0, 0, nil, nil, fmt.Errorf("No liquidity in pool for swap")
	}

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

	var finalSellAmount, sellAmountRemain, buyAmount uint64
	var err error
	if ord == nil || targetDeltaX.Cmp(maxDeltaX) >= 0 {
		// able to trade fully in pool before reaching order rate
		finalSellAmount = maxSellAmountAfterFee
		sellAmountRemain = 0
		buyAmount, err = tp.BuyAmount(finalSellAmount, tradeDirection)
		if err != nil {
			return 0, 0, nil, nil, err
		}
	} else {
		if targetDeltaX.Cmp(big.NewInt(0)) <= 0 {
			// pool price already surpassed order rate -> exit
			return 0, maxSellAmountAfterFee, big.NewInt(0), big.NewInt(0), nil
		}
		// only swap the target delta x
		// maxDeltaX is valid uint64, while 0 < targetDeltaX < maxDeltaX
		finalSellAmount = targetDeltaX.Uint64()
		sellAmountRemain = big.NewInt(0).Sub(maxDeltaX, targetDeltaX).Uint64()
		buyAmount, err = tp.BuyAmount(finalSellAmount, tradeDirection)
		if err != nil {
			return 0, 0, nil, nil, err
		}
		if buyAmount == 0 {
			// pool price close enough to order rate -> exit
			return 0, maxSellAmountAfterFee, big.NewInt(0), big.NewInt(0), nil
		}
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
			return fmt.Errorf("Invalid signs for reserve changes %v, %v", change0, change1)
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
	reserves []*rawdbv2.Pdexv3PoolPair, lpFeesPerShares []map[common.Hash]*big.Int,
	protocolFees, stakingPoolFees []map[common.Hash]uint64,
	tradeDirections []byte,
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
			buyAmount, temp, token0Change, token1Change, err := NewTradingPairWithValue(
				reserves[i],
			).SwapToReachOrderRate(sellAmountRemain, tradeDirections[i], order)
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

func TrackFee(
	fee uint64, feeInPRV bool, sellingTokenID common.Hash, baseLPPerShare *big.Int, bps uint,
	tradePath []string, reserves []*rawdbv2.Pdexv3PoolPair,
	lpFeesPerShares []map[common.Hash]*big.Int, protocolFees, stakingPoolFees []map[common.Hash]uint64,
	tradeDirections []byte, orderbooks []OrderBookIterator,
	poolFees []uint, feeRateBPS uint,
	acceptedMeta *metadataPdexv3.AcceptedTrade,
	protocolFeePercent, stakingPoolRewardPercent uint, stakingRewardTokens []common.Hash,
	defaultOrderTradingRewardRatioBPS uint, orderTradingRewardRatioBPS map[string]uint,
) (*metadataPdexv3.AcceptedTrade, []map[string]map[common.Hash]uint64, []map[common.Hash]map[string]*big.Int, error) {
	mutualLen := len(reserves)
	if len(tradeDirections) != mutualLen || len(orderbooks) != mutualLen {
		return nil, nil, nil, fmt.Errorf("Trade path vs directions vs orderbooks length mismatch")
	}

	acceptedMeta.RewardEarned = make([]map[common.Hash]uint64, mutualLen)
	for i := 0; i < mutualLen; i++ {
		acceptedMeta.RewardEarned[i] = make(map[common.Hash]uint64)
	}

	orderRewardChanges := make([]map[string]map[common.Hash]uint64, mutualLen)
	for i := 0; i < mutualLen; i++ {
		orderRewardChanges[i] = make(map[string]map[common.Hash]uint64)
	}

	orderMakingChanges := make([]map[common.Hash]map[string]*big.Int, mutualLen)
	for i := 0; i < mutualLen; i++ {
		orderMakingChanges[i] = make(map[common.Hash]map[string]*big.Int)
	}

	if feeInPRV || sellingTokenID == common.PRVCoinID {
		// weighted divide fee into reserves
		sumPoolFees := feeRateBPS
		feeRemain := fee
		for i := 0; i < mutualLen; i++ {
			// reward for this pool = feeRemain * feeRate / sumPoolFeesRemain
			reward := new(big.Int).Mul(new(big.Int).SetUint64(feeRemain), new(big.Int).SetUint64(uint64(poolFees[i])))
			reward.Div(reward, new(big.Int).SetUint64(uint64(sumPoolFees)))

			// split reward between LPs and LOPs by weighted ratio
			ratio := defaultOrderTradingRewardRatioBPS
			if orderTradingRewardRatioBPS != nil {
				bps, ok := orderTradingRewardRatioBPS[tradePath[i]]
				if ok {
					ratio = bps
				}
			}

			remain := new(big.Int).SetUint64(0)

			// add staking pools and protocol fees
			protocolFees[i], stakingPoolFees[i], remain = NewTradingPairWithValue(
				reserves[i],
			).AddStakingAndProtocolFee(
				common.PRVCoinID, reward, protocolFees[i], stakingPoolFees[i],
				protocolFeePercent, stakingPoolRewardPercent, stakingRewardTokens,
			)

			ammMakingVolume, orderMakingVolumes, tradeDirection := GetMakingVolumes(
				acceptedMeta.PairChanges[i], acceptedMeta.OrderChanges[i],
				orderbooks[i].NftIDs(),
			)

			makingToken := reserves[i].Token0ID()
			if tradeDirection == TradeDirectionSell0 {
				makingToken = reserves[i].Token1ID()
			}
			orderMakingChanges[i][makingToken] = orderMakingVolumes

			ammReward, orderRewards := SplitTradingReward(
				remain, ratio, bps,
				ammMakingVolume, orderMakingVolumes,
			)

			// add reward to LOPs
			for nftID, reward := range orderRewards {
				if _, ok := orderRewardChanges[i][nftID]; !ok {
					orderRewardChanges[i][nftID] = make(map[common.Hash]uint64)
				}
				if _, ok := orderRewardChanges[i][nftID][common.PRVCoinID]; !ok {
					orderRewardChanges[i][nftID][common.PRVCoinID] = 0
				}
				orderRewardChanges[i][nftID][common.PRVCoinID] += reward
			}

			// add reward to LPs
			lpFeesPerShares[i] = NewTradingPairWithValue(
				reserves[i],
			).AddLPFee(
				common.PRVCoinID, new(big.Int).SetUint64(ammReward), baseLPPerShare,
				lpFeesPerShares[i],
			)
			acceptedMeta.RewardEarned[i][common.PRVCoinID] = reward.Uint64()

			sumPoolFees -= poolFees[i]
			feeRemain -= reward.Uint64()
		}
		return acceptedMeta, orderRewardChanges, orderMakingChanges, nil
	}

	sumPoolFees := feeRateBPS
	sellAmountRemain := fee

	var totalBuyAmount uint64
	for i := 0; i < mutualLen; i++ {
		// reward for this pool = feeRemain * feeRate / sumPoolFeesRemain
		reward := new(big.Int).Mul(new(big.Int).SetUint64(sellAmountRemain), new(big.Int).SetUint64(uint64(poolFees[i])))
		reward.Div(reward, new(big.Int).SetUint64(uint64(sumPoolFees)))
		rewardAmount := reward.Uint64()

		rewardToken := reserves[i].Token0ID()
		if tradeDirections[i] == TradeDirectionSell1 {
			rewardToken = reserves[i].Token1ID()
		}

		// split reward between LPs and LOPs by weighted ratio
		ratio := defaultOrderTradingRewardRatioBPS
		if orderTradingRewardRatioBPS != nil {
			bps, ok := orderTradingRewardRatioBPS[tradePath[i]]
			if ok {
				ratio = bps
			}
		}

		remain := new(big.Int).SetUint64(0)

		// add staking pools and protocol fees
		protocolFees[i], stakingPoolFees[i], remain = NewTradingPairWithValue(
			reserves[i],
		).AddStakingAndProtocolFee(
			rewardToken, reward, protocolFees[i], stakingPoolFees[i],
			protocolFeePercent, stakingPoolRewardPercent, stakingRewardTokens,
		)

		ammMakingVolume, orderMakingVolumes, tradeDirection := GetMakingVolumes(
			acceptedMeta.PairChanges[i], acceptedMeta.OrderChanges[i],
			orderbooks[i].NftIDs(),
		)

		makingToken := reserves[i].Token0ID()
		if tradeDirection == TradeDirectionSell0 {
			makingToken = reserves[i].Token1ID()
		}
		orderMakingChanges[i][makingToken] = orderMakingVolumes

		ammReward, orderRewards := SplitTradingReward(
			remain, ratio, bps,
			ammMakingVolume, orderMakingVolumes,
		)

		// add reward to LOPs
		for nftID, reward := range orderRewards {
			if _, ok := orderRewardChanges[i][nftID]; !ok {
				orderRewardChanges[i][nftID] = make(map[common.Hash]uint64)
			}
			if _, ok := orderRewardChanges[i][nftID][rewardToken]; !ok {
				orderRewardChanges[i][nftID][rewardToken] = 0
			}
			orderRewardChanges[i][nftID][rewardToken] += reward
		}

		// add reward to LPs
		lpFeesPerShares[i] = NewTradingPairWithValue(
			reserves[i],
		).AddLPFee(
			rewardToken, new(big.Int).SetUint64(ammReward), baseLPPerShare,
			lpFeesPerShares[i],
		)
		acceptedMeta.RewardEarned[i][rewardToken] = rewardAmount

		sumPoolFees -= poolFees[i]
		sellAmountRemain -= rewardAmount
		if i == mutualLen-1 {
			break
		}

		accumulatedToken0Change := big.NewInt(0)
		accumulatedToken1Change := big.NewInt(0)
		totalBuyAmount = uint64(0)

		for order, ordID, err := orderbooks[i].NextOrder(tradeDirections[i]); err == nil; order, ordID, err = orderbooks[i].NextOrder(tradeDirections[i]) {
			buyAmount, temp, token0Change, token1Change, err :=
				NewTradingPairWithValue(
					reserves[i],
				).SwapToReachOrderRate(sellAmountRemain, tradeDirections[i], order)
			if err != nil {
				return nil, nil, nil, err
			}
			sellAmountRemain = temp
			if totalBuyAmount+buyAmount < totalBuyAmount {
				return nil, nil, nil, fmt.Errorf("Sum exceeds uint64 range after swapping in pool")
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
					return nil, nil, nil, err
				}
				sellAmountRemain = temp
				if totalBuyAmount+buyAmount < totalBuyAmount {
					return nil, nil, nil, fmt.Errorf("Sum exceeds uint64 range after matching order")
				}
				totalBuyAmount += buyAmount
				// add order balance changes to "accepted" instruction
				prevToken0Change := new(big.Int).SetUint64(0)
				prevToken1Change := new(big.Int).SetUint64(0)
				if _, ok := acceptedMeta.OrderChanges[i][ordID]; ok {
					prevToken0Change = acceptedMeta.OrderChanges[i][ordID][0]
					prevToken1Change = acceptedMeta.OrderChanges[i][ordID][1]
				}
				acceptedMeta.OrderChanges[i][ordID] = [2]*big.Int{new(big.Int).SetUint64(0), new(big.Int).SetUint64(0)}
				acceptedMeta.OrderChanges[i][ordID][0].Add(prevToken0Change, token0Change)
				acceptedMeta.OrderChanges[i][ordID][1].Add(prevToken1Change, token1Change)
				if sellAmountRemain == 0 {
					break
				}
			}
		}

		// add pair changes to "accepted" instruction
		acceptedMeta.PairChanges[i][0] = new(big.Int).Add(acceptedMeta.PairChanges[i][0], accumulatedToken0Change)
		acceptedMeta.PairChanges[i][1] = new(big.Int).Add(acceptedMeta.PairChanges[i][1], accumulatedToken1Change)

		// set sell amount before moving on to next pair
		sellAmountRemain = totalBuyAmount
	}

	return acceptedMeta, orderRewardChanges, orderMakingChanges, nil
}

func (tp *TradingPair) AddStakingAndProtocolFee(
	tokenID common.Hash, amount *big.Int,
	rootProtocolFees, rootStakingPoolFees map[common.Hash]uint64,
	protocolFeePercent, stakingPoolRewardPercent uint, stakingRewardTokens []common.Hash,
) (map[common.Hash]uint64, map[common.Hash]uint64, *big.Int) {
	isStakingRewardToken := false
	for _, stakingRewardToken := range stakingRewardTokens {
		if tokenID == stakingRewardToken {
			isStakingRewardToken = true
			break
		}
	}

	if !isStakingRewardToken {
		stakingPoolRewardPercent = 0
	}

	// if there is no LP for this pair, then there is no LP fee to add
	if tp.ShareAmount() == 0 {
		if !isStakingRewardToken {
			// move all LP fee to protocol fee
			protocolFeePercent = 100
		} else {
			// move all LP fee to staking pool fee
			stakingPoolRewardPercent = 100 - protocolFeePercent
		}
	}

	protocolFees := new(big.Int).Mul(amount, new(big.Int).SetUint64(uint64(protocolFeePercent)))
	protocolFees = new(big.Int).Div(protocolFees, new(big.Int).SetUint64(100))

	if protocolFees.IsUint64() && protocolFees.Uint64() != 0 {
		oldProtocolFees, isExisted := rootProtocolFees[tokenID]
		if !isExisted {
			oldProtocolFees = uint64(0)
		}
		tempProtocolFees := rootProtocolFees
		tempProtocolFees[tokenID] = oldProtocolFees + protocolFees.Uint64()

		rootProtocolFees = tempProtocolFees
	}

	stakingRewards := new(big.Int).Mul(amount, new(big.Int).SetUint64(uint64(stakingPoolRewardPercent)))
	stakingRewards = new(big.Int).Div(stakingRewards, new(big.Int).SetUint64(100))

	if stakingRewards.IsUint64() && stakingRewards.Uint64() != 0 {
		oldStakingRewards, isExisted := rootStakingPoolFees[tokenID]
		if !isExisted {
			oldStakingRewards = uint64(0)
		}
		tempStakingRewards := rootStakingPoolFees
		tempStakingRewards[tokenID] = oldStakingRewards + stakingRewards.Uint64()

		rootStakingPoolFees = tempStakingRewards
	}

	remain := new(big.Int).Sub(amount, protocolFees)
	remain.Sub(remain, stakingRewards)

	return rootProtocolFees, rootStakingPoolFees, remain
}

func (tp *TradingPair) AddLPFee(
	tokenID common.Hash, amount *big.Int, baseLPPerShare *big.Int,
	rootLpFeesPerShare map[common.Hash]*big.Int,
) map[common.Hash]*big.Int {
	if tp.ShareAmount() == 0 {
		return rootLpFeesPerShare
	}
	oldLPFeesPerShare, isExisted := rootLpFeesPerShare[tokenID]
	if !isExisted {
		oldLPFeesPerShare = big.NewInt(0)
	}

	// delta (fee / LP share) = LP Reward * BASE / totalLPShare
	deltaLPFeesPerShare := new(big.Int).Mul(amount, baseLPPerShare)
	deltaLPFeesPerShare = new(big.Int).Div(deltaLPFeesPerShare, new(big.Int).SetUint64(tp.ShareAmount()))

	// update accumulated sum of (fee / LP share)
	newLPFeesPerShare := new(big.Int).Add(oldLPFeesPerShare, deltaLPFeesPerShare)
	tempLPFeesPerShare := rootLpFeesPerShare
	tempLPFeesPerShare[tokenID] = newLPFeesPerShare

	rootLpFeesPerShare = tempLPFeesPerShare
	return rootLpFeesPerShare
}

func (tp *TradingPair) AddLMRewards(
	tokenID common.Hash, amount *big.Int, baseLPPerShare *big.Int,
	rootLmRewardsPerShare map[common.Hash]*big.Int,
) map[common.Hash]*big.Int {
	if tp.ShareAmount() == tp.LmLockedShareAmount() {
		return rootLmRewardsPerShare
	}
	oldLMRewardsPerShare, isExisted := rootLmRewardsPerShare[tokenID]
	if !isExisted {
		oldLMRewardsPerShare = big.NewInt(0)
	}

	unlockedShareAmount := new(big.Int).SetUint64(tp.ShareAmount() - tp.LmLockedShareAmount())

	// delta (fee / LP share) = LM Reward * BASE / totalLPShare
	deltaLMRewardsPerShare := new(big.Int).Mul(amount, baseLPPerShare)
	deltaLMRewardsPerShare = new(big.Int).Div(deltaLMRewardsPerShare, unlockedShareAmount)

	// update accumulated sum of (fee / LP share)
	newLPFeesPerShare := new(big.Int).Add(oldLMRewardsPerShare, deltaLMRewardsPerShare)
	tempLPFeesPerShare := rootLmRewardsPerShare
	tempLPFeesPerShare[tokenID] = newLPFeesPerShare

	rootLmRewardsPerShare = tempLPFeesPerShare
	return rootLmRewardsPerShare
}

func HasInsufficientLiquidity(poolPair rawdbv2.Pdexv3PoolPair) bool {
	return poolPair.Token0RealAmount() <= 0 || poolPair.Token1RealAmount() <= 0
}
