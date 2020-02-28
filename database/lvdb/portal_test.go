package lvdb

import (
	"math"
	"testing"
)
import "github.com/stretchr/testify/assert"

func TestFinalExchangeRates(t *testing.T)  {
	ratesDetail := make(map[string]FinalExchangeRatesDetail)
	ratesDetail["BTC"] = FinalExchangeRatesDetail{Amount: 10}
	ratesDetail["BNB"] = FinalExchangeRatesDetail{Amount: 20}
	ratesDetail["PRV"] = FinalExchangeRatesDetail{Amount: 5}

	finalExchangeRates := FinalExchangeRates {
		Rates: ratesDetail,
	}

	bnb2PRV := finalExchangeRates.ExchangeBNB2PRV(1)
	assert.Equal(t, bnb2PRV, uint64(4000000000))
	assert.Equal(t, uint64(4),  uint64(4000000000 / math.Pow10(9)))

	prv2BNB := finalExchangeRates.ExchangePRV2BNB(4000000000)
	assert.Equal(t, prv2BNB, uint64(1000000000))
	assert.Equal(t, uint64(1),  uint64(1000000000 / math.Pow10(9)))

	btc2PRV := finalExchangeRates.ExchangeBTC2PRV(1)
	assert.Equal(t, btc2PRV, uint64(2000000000))
	assert.Equal(t, uint64(2),  uint64(2000000000 / math.Pow10(9)))

	prv2BTC := finalExchangeRates.ExchangePRV2BTC(2000000000)
	assert.Equal(t, prv2BTC, uint64(1000000000))
	assert.Equal(t, uint64(1),  uint64(1000000000 / math.Pow10(9)))
}