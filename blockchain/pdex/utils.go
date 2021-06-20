package pdex

import "github.com/incognitochain/incognito-chain/common"

func InitVersionByBeaconHeight(beaconHeight uint64) State {
	var state State
	return state
}

func isTradingPairContainsPRV(
	tokenIDToSellStr string,
	tokenIDToBuyStr string,
) bool {
	return tokenIDToSellStr == common.PRVCoinID.String() ||
		tokenIDToBuyStr == common.PRVCoinID.String()
}

type tradingFeeForContributorByPair struct {
	ContributorAddressStr string
	FeeAmt                uint64
	Token1IDStr           string
	Token2IDStr           string
}

type tradeInfo struct {
	tokenIDToBuyStr         string
	tokenIDToSellStr        string
	sellAmount              uint64
	newTokenPoolValueToBuy  uint64
	newTokenPoolValueToSell uint64
	receiveAmount           uint64
}

type shareInfo struct {
	shareKey string
	shareAmt uint64
}

type deductingAmountsByWithdrawal struct {
	Token1IDStr string
	PoolValue1  uint64
	Token2IDStr string
	PoolValue2  uint64
	Shares      uint64
}
