package v2utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	. "github.com/stretchr/testify/assert"
)

func TestProduceAcceptedTrade(t *testing.T) {
	type TestData struct {
		AmountIn        uint64                    `json:"amountIn"`
		Fee             uint64                    `json:"fee"`
		Reserves        []*rawdbv2.Pdexv3PoolPair `json:"reserves"`
		TradeDirections []byte                    `json:"tradeDirections"`
		Orderbooks      []OrderList               `json:"orders"` // assume orders have been sorted
	}

	type TestResult struct {
		Instructions    []string                  `json:"instructions"`
		ChangedReserves []*rawdbv2.Pdexv3PoolPair `json:"changedReserves"`
	}

	// use all available testcases in testdata/
	var testcases []Testcase
	testcases = append(testcases, singleTradeTestcases...)
	testcases = append(testcases, pathTradeTestcases...)
	testcases = append(testcases, orderMatchTestcases...)

	var blankReceiver privacy.OTAReceiver
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			errParseExpectedResult := json.Unmarshal(testcase.Expected, &expected)

			// fill trade path with blank data, except for reserves
			orderbooks := make([]OrderBookIterator, len(testdata.Reserves))
			dummyTradePath := make([]string, len(testdata.Reserves))
			var lpFeesPerShares []map[common.Hash]*big.Int
			var protocolFees, stakingPoolFees []map[common.Hash]uint64
			for index, item := range testdata.Orderbooks {
				orderbooks[index] = &item
				dummyTradePath[index] = "pair" + strconv.Itoa(index)
				lpFeesPerShares = append(lpFeesPerShares, nil)
				protocolFees = append(protocolFees, nil)
				stakingPoolFees = append(stakingPoolFees, nil)
			}

			// expected outputs will have RequestTxID, shardID zeroed. Those data are out of scope for this test
			acceptedMd, changedReserves, err := MaybeAcceptTrade(
				testdata.AmountIn, 0, dummyTradePath, blankReceiver,
				testdata.Reserves, lpFeesPerShares, protocolFees, stakingPoolFees,
				testdata.TradeDirections, common.PRVCoinID, 0, orderbooks)
			acn := instruction.Action{Content: acceptedMd}
			if testcase.ExpectFailure {
				Error(t, err)
			} else {
				NoError(t, errParseExpectedResult)
				Equal(t, expected, TestResult{acn.StringSlice(), changedReserves})
			}

		})
	}
}

func (o *OrderList) NftIDs() map[string]string {
	return nil
}

type OrderList []MatchingOrder

// replica of PoolPairState.NextOrder()
func (o *OrderList) NextOrder(tradeDirection byte) (*MatchingOrder, string, error) {
	lst := []MatchingOrder(*o)
	lstLen := len(lst)
	switch tradeDirection {
	case TradeDirectionSell0:
		for i := 0; i < lstLen; i++ {
			if lst[i].TradeDirection() != tradeDirection && lst[i].Token1Balance() > 0 {
				return &lst[i], lst[i].Id(), nil
			}
		}
		// no active order
		return nil, "", nil
	case TradeDirectionSell1:
		for i := lstLen - 1; i >= 0; i-- {
			if lst[i].TradeDirection() != tradeDirection && lst[i].Token0Balance() > 0 {
				return &lst[i], lst[i].Id(), nil
			}
		}
		// no active order
		return nil, "", nil
	default:
		return nil, "", fmt.Errorf("Invalid trade direction %d", tradeDirection)
	}
}

type Testcase struct {
	Name          string          `json:"name"`
	Data          json.RawMessage `json:"data"`
	Expected      json.RawMessage `json:"expected"`
	ExpectFailure bool            `json:"fail"`
}

func mustReadTestcases(filename string) []Testcase {
	raw, err := ioutil.ReadFile("testdata/" + filename)
	if err != nil {
		panic(err)
	}
	var results []Testcase = make([]Testcase, 30)
	err = json.Unmarshal(raw, &results)
	if err != nil {
		panic(err)
	}
	return results
}

var singleTradeTestcases = mustReadTestcases("single_trade.json")
var pathTradeTestcases = mustReadTestcases("path_2.json")
var orderMatchTestcases = mustReadTestcases("order_match.json")
