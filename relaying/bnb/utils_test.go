package bnb

import (
	"fmt"
	"github.com/binance-chain/go-sdk/common/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateBNBAddress(t *testing.T) {
	addr := generateBNBAddress(types.ProdNetwork)
	fmt.Printf("addr %v\n", addr)
}

func TestIsValidBNBAddress(t *testing.T) {
	testcase := []struct {
		address string
		chainID string
		isValid bool
	}{
		{"abc", TestnetBNBChainID, false},
		{"tbnb1fau9kq605jwkyfea2knw495we8cpa47r9r6uxv", TestnetBNBChainID, true},
		{"bnb149kr0gux3nfme8q60z6t7e6hv0h8nsqylff8sw", MainnetBNBChainID, true},
		{"tbnb1fau9kq605jwkyfea2knw495we8cpa47r9r6uxv", MainnetBNBChainID, false},
	}

	for _, tc := range testcase {
		fmt.Printf("Running testcase: %+v\n", tc)
		actualResult := IsValidBNBAddress(tc.address, tc.chainID)
		assert.Equal(t, tc.isValid, actualResult)
	}
}
