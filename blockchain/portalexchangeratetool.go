package blockchain

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"math"
	"math/big"
)

type RateInfo struct {
	Rate    uint64
	Decimal uint8
}
type PortalExchangeRateTool struct {
	Rates map[string]RateInfo
}

// getDecimal returns decimal for portal token or collateral tokens
func getDecimal(supportPortalCollateral []PortalCollateral, tokenID string) uint8 {
	if metadata.IsPortalToken(tokenID) || tokenID == common.PRVIDStr {
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
	supportPortalCollateral []PortalCollateral,
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
func (t *PortalExchangeRateTool) Convert (tokenIDFrom string, tokenIDTo string, amount uint64) (uint64, error){
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

// ConvertToUSDT converts amount to usdt amount (in nano)
func (t *PortalExchangeRateTool) ConvertToUSDT (tokenIDFrom string, amount uint64) (uint64, error){
	rateFrom := t.Rates[tokenIDFrom]
	if rateFrom.Rate == 0 {
		return 0, errors.New("invalid exchange rate to convert to usdt")
	}

	res := new(big.Int).Mul(new(big.Int).SetUint64(amount), new(big.Int).SetUint64(rateFrom.Rate))
	res = res.Div(res, new(big.Int).SetUint64(uint64(math.Pow10(int(rateFrom.Decimal)))))

	return res.Uint64(), nil
}

// ConvertToUSDT converts amount from usdt to token amount (in nano)
func (t *PortalExchangeRateTool) ConvertFromUSDT (tokenIDTo string, amount uint64) (uint64, error){
	rateTo := t.Rates[tokenIDTo]
	if rateTo.Rate == 0 {
		return 0, errors.New("invalid exchange rate to convert to usdt")
	}

	res := new(big.Int).Mul(new(big.Int).SetUint64(amount), new(big.Int).SetUint64(uint64(math.Pow10(int(rateTo.Decimal)))))
	res = res.Div(res, new(big.Int).SetUint64(rateTo.Rate))

	return res.Uint64(), nil
}
