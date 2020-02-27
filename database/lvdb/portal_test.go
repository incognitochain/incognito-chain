package lvdb

import (
	"math"
	"testing"
)
import "github.com/stretchr/testify/assert"

func TestFinalExchangeRates(t *testing.T)  {
	ratesDetail := make(map[string]FinalExchangeRatesDetail)
	ratesDetail["BTC"] = FinalExchangeRatesDetail{Amount: uint64(8000 * math.Pow10(6))}
	ratesDetail["BNB"] = FinalExchangeRatesDetail{Amount: uint64(20 * math.Pow10(6))}
	ratesDetail["PRV"] = FinalExchangeRatesDetail{Amount: uint64(0.5 * math.Pow10(6))}

	finalExchangeRates := FinalExchangeRates {
		Rates: ratesDetail,
	}

	bnb2PRV := finalExchangeRates.ExchangeBNB2PRV(20)
	assert.Equal(t, bnb2PRV, uint64(800))

	prv2BNB := finalExchangeRates.ExchangePRV2BNB(800)
	assert.Equal(t, prv2BNB, uint64(20))

	btc2PRV := finalExchangeRates.ExchangeBTC2PRV(1)
	assert.Equal(t, btc2PRV, uint64(16000))

	prv2BTC := finalExchangeRates.ExchangePRV2BTC(16000)
	assert.Equal(t, prv2BTC, uint64(1))
}