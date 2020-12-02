package portalprocess

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/portal"
	portalMeta "github.com/incognitochain/incognito-chain/portal/metadata"
	"math"
	"math/big"
	"sort"
)

type RateInfo struct {
	Rate    uint64
	Decimal uint8
}
type PortalExchangeRateTool struct {
	Rates map[string]RateInfo
}

// getDecimal returns decimal for portal token or collateral tokens
func getDecimal(supportPortalCollateral []portal.PortalCollateral, tokenID string) uint8 {
	if portalMeta.IsPortalToken(tokenID) || tokenID == common.PRVIDStr {
		return 9
	}
	for _, col := range supportPortalCollateral {
		if tokenID == col.ExternalTokenID {
			return col.Decimal
		}
	}

	return 0
}

func NewPortalExchangeRateTool(
	finalExchangeRate *statedb.FinalExchangeRatesState,
	supportPortalCollateral []portal.PortalCollateral,
) *PortalExchangeRateTool {
	t := new(PortalExchangeRateTool)
	t.Rates = map[string]RateInfo{}

	rates := finalExchangeRate.Rates()

	for tokenID, detail := range rates {
		decimal := getDecimal(supportPortalCollateral, tokenID)
		if decimal > 0 {
			t.Rates[tokenID] = RateInfo{
				Rate:    detail.Amount,
				Decimal: decimal,
			}
		}
	}

	return t
}

// convert converts amount in nano unit from tokenIDFrom to tokenIDTo
// result in nano unit (smallest unit of token)
func (t *PortalExchangeRateTool) Convert(tokenIDFrom string, tokenIDTo string, amount uint64) (uint64, error) {
	rateFrom := t.Rates[tokenIDFrom]
	rateTo := t.Rates[tokenIDTo]
	if rateFrom.Rate == 0 || rateTo.Rate == 0 {
		return 0, errors.New("invalid exchange rate to convert")
	}

	res := new(big.Int).Mul(new(big.Int).SetUint64(amount), new(big.Int).SetUint64(rateFrom.Rate))
	res = res.Mul(res, new(big.Int).SetUint64(uint64(math.Pow10(int(rateTo.Decimal)))))
	res = res.Div(res, new(big.Int).SetUint64(uint64(math.Pow10(int(rateFrom.Decimal)))))
	res = res.Div(res, new(big.Int).SetUint64(rateTo.Rate))

	return res.Uint64(), nil
}

// ConvertToUSD converts amount to usdt amount (in nano)
func (t *PortalExchangeRateTool) ConvertToUSD(tokenIDFrom string, amount uint64) (uint64, error) {
	rateFrom := t.Rates[tokenIDFrom]
	if rateFrom.Rate == 0 {
		return 0, errors.New("invalid exchange rate to convert to usdt")
	}

	res := new(big.Int).Mul(new(big.Int).SetUint64(amount), new(big.Int).SetUint64(rateFrom.Rate))
	res = res.Div(res, new(big.Int).SetUint64(uint64(math.Pow10(int(rateFrom.Decimal)))))

	return res.Uint64(), nil
}

// ConvertToUSD converts amount from usdt to token amount (in nano)
func (t *PortalExchangeRateTool) ConvertFromUSD(tokenIDTo string, amount uint64) (uint64, error) {
	rateTo := t.Rates[tokenIDTo]
	if rateTo.Rate == 0 {
		return 0, errors.New("invalid exchange rate to convert to usdt")
	}

	res := new(big.Int).Mul(new(big.Int).SetUint64(amount), new(big.Int).SetUint64(uint64(math.Pow10(int(rateTo.Decimal)))))
	res = res.Div(res, new(big.Int).SetUint64(rateTo.Rate))

	return res.Uint64(), nil
}

func (t *PortalExchangeRateTool) ConvertMapTokensToUSD(tokens map[string]uint64) (uint64, error) {
	if len(tokens) == 0 {
		return 0, nil
	}

	res := uint64(0)
	for tokenID, amount := range tokens {
		amountInUSDT, err := t.ConvertToUSD(tokenID, amount)
		if err != nil {
			return 0, nil
		}
		res += amountInUSDT
	}
	return res, nil
}

func (t *PortalExchangeRateTool) ConvertMapTokensFromUSD(amountInUSDT uint64, maxPRVAmount uint64, maxTokenAmounts map[string]uint64) (uint64, map[string]uint64, error) {
	if amountInUSDT == 0 {
		return 0, nil, nil
	}

	prvAmountRes := uint64(0)
	tokenAmountsRes := map[string]uint64{}

	// convert to prv amount first
	maxPRVInUSDT, err := t.ConvertToUSD(common.PRVIDStr, maxPRVAmount)
	if err != nil {
		return 0, nil, nil
	}

	if maxPRVInUSDT <= amountInUSDT {
		prvAmountRes = maxPRVAmount
		amountInUSDT -= maxPRVInUSDT
	} else {
		prvAmountRes, err = t.ConvertFromUSD(common.PRVIDStr, amountInUSDT)
		if err != nil {
			return 0, nil, nil
		}
		amountInUSDT = 0
	}

	if amountInUSDT == 0 {
		return prvAmountRes, tokenAmountsRes, nil
	}

	// sort token by amountInUSDT descending
	type tokenInfo struct {
		tokenID      string
		amount       uint64
		amountInUSDT uint64
	}
	tokenInfos := make([]tokenInfo, 0)
	for tokenID, maxAmount := range maxTokenAmounts {
		maxAmountInUSDT, err := t.ConvertToUSD(tokenID, maxAmount)
		if err != nil {
			return 0, nil, nil
		}

		tokenInfos = append(tokenInfos, tokenInfo{
			tokenID:      tokenID,
			amount:       maxAmount,
			amountInUSDT: maxAmountInUSDT,
		})
	}
	sort.SliceStable(tokenInfos, func(i, j int) bool {
		return tokenInfos[i].amountInUSDT > tokenInfos[j].amountInUSDT
	})

	for _, tInfo := range tokenInfos {
		if tInfo.amountInUSDT <= amountInUSDT {
			tokenAmountsRes[tInfo.tokenID] = tInfo.amount
			amountInUSDT -= tInfo.amountInUSDT
		} else {
			tokenAmountsRes[tInfo.tokenID], err = t.ConvertFromUSD(tInfo.tokenID, amountInUSDT)
			if err != nil {
				return 0, nil, nil
			}
			amountInUSDT = 0
		}

		if amountInUSDT == 0 {
			return prvAmountRes, tokenAmountsRes, nil
		}
	}

	return 0, nil, errors.New("Not enough token to convert")
}
