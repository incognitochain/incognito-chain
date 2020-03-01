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

	bnb2PRV := finalExchangeRates.ExchangeBNB2PRV(uint64(math.Pow10(9)))
	assert.Equal(t, bnb2PRV, uint64(4000000000))


	prv2BNB := finalExchangeRates.ExchangePRV2BNB(4000000000)
	assert.Equal(t, prv2BNB, uint64(math.Pow10(9)))

	btc2PRV := finalExchangeRates.ExchangeBTC2PRV(uint64(math.Pow10(9)))
	assert.Equal(t, btc2PRV, 2 * uint64(math.Pow10(9)))

	prv2BTC := finalExchangeRates.ExchangePRV2BTC(2 * uint64(math.Pow10(9)))
	assert.Equal(t, prv2BTC, uint64(math.Pow10(9)))
}

func TestRealFinalExchangeRates(t *testing.T)  {
	ratesDetail := make(map[string]FinalExchangeRatesDetail)
	ratesDetail["BTC"] = FinalExchangeRatesDetail{Amount: 9000 * uint64(math.Pow10(6))}
	ratesDetail["BNB"] = FinalExchangeRatesDetail{Amount: 20 * uint64(math.Pow10(6))}
	ratesDetail["PRV"] = FinalExchangeRatesDetail{Amount: uint64(0.5 * math.Pow10(6))}

	finalExchangeRates := FinalExchangeRates {
		Rates: ratesDetail,
	}


	bnb2PRV := finalExchangeRates.ExchangeBNB2PRV(1000) //nano BNB
	assert.Equal(t, bnb2PRV, uint64(40000))

	prv2BNB := finalExchangeRates.ExchangePRV2BNB(40000) //nano PRV
	assert.Equal(t, prv2BNB,  uint64(1000))

}