package v3utils

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	. "github.com/stretchr/testify/assert"
)

type orderless struct{}

func (o *orderless) NextOrder(tradeDirection int) (*OrderMatchingInfo, string, error) {
	return nil, "", nil
}

func TestTradeSinglePair(t *testing.T) {
	type TestData struct {
		AmountIn        uint64
		Fee             uint64
		Reserves        []*PoolReserve
		TradeDirections []int
	}

	type TestResult struct {
		Instructions    []string
		ChangedReserves []*PoolReserve
	}

	type Testcase struct {
		name          string
		data          string
		expected      string
		expectSuccess bool
	}

	testcases := []Testcase{
		Testcase{"New pool - Valid trade", "{\"AmountIn\":80,\"Fee\":30,\"Reserves\":[{\"Token0\":200,\"Token1\":2000,\"Token0Virtual\":400,\"Token1Virtual\":4000}],\"TradeDirections\":[0]}", "{\"Instructions\":[\"273\",\"accepted\",\"0\",\"{\\\"Content\\\":{\\\"Receiver\\\":\\\"15VRBi1S7Pme5bUpHW12HZVjXCTg1FwDM3yoSUjWgGEhVJoLudKwtpQk3iSmwe27ffsj76LLZvyJ9x5tbX44SmKBSzdegpLsNKFc71j4jDxQh1PYSVbVHtTbgMoLvUUReaQWyyoXPvuarV5E\\\",\\\"Amount\\\":444,\\\"TokenToBuy\\\":\\\"0000000000000000000000000000000000000000000000000000000000000004\\\",\\\"PairChanges\\\":[[50,-444]],\\\"OrderChanges\\\":[{}]},\\\"RequestTxID\\\":\\\"0000000000000000000000000000000000000000000000000000000000000000\\\"}\"],\"ChangedReserves\":[{\"Token0\":250,\"Token1\":1556,\"Token0Virtual\":450,\"Token1Virtual\":3556}]}", true},
		Testcase{"New pool - Insufficient trade", "{\"AmountIn\":30030,\"Fee\":30,\"Reserves\":[{\"Token0\":200,\"Token1\":2000,\"Token0Virtual\":400,\"Token1Virtual\":4000}],\"TradeDirections\":[1]}", "Not enough liquidity for trade", false},
	}

	var blankReceiver privacy.OTAReceiver
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal([]byte(testcase.data), &testdata)
			NoError(t, err)
			// s, _ := json.Marshal(testdata)
			// fmt.Println(string(s))
			
			instructions, changedReserves, err := MaybeAcceptTrade(testdata.AmountIn, testdata.Fee, blankReceiver, testdata.Reserves, testdata.TradeDirections, common.PRVCoinID, []OrderBookIterator{&orderless{}})
			encodedResult, _ := json.Marshal(TestResult{instructions, changedReserves})
			// fmt.Println(string(encodedResult))
			if testcase.expectSuccess {
				NoError(t, err)
				Equal(t, string(encodedResult), testcase.expected)
			} else {
				Errorf(t, err, testcase.expected)
			}
			
		})
	}
}
