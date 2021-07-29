package v3utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	. "github.com/stretchr/testify/assert"
)

func TestProduceAcceptedTrade(t *testing.T) {
	type TestData struct {
		AmountIn        uint64         `json:"amountIn"`
		Fee             uint64         `json:"fee"`
		Reserves        []*PoolReserve `json:"reserves"`
		TradeDirections []int          `json:"tradeDirections"`
		Orderbooks      []OrderList    `json:"orders"` // assume orders have been sorted
	}

	type TestResult struct {
		Instructions    []string       `json:"instructions"`
		ChangedReserves []*PoolReserve `json:"changedReserves"`
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
			err := json.Unmarshal([]byte(testcase.Data), &testdata)
			NoError(t, err)
			// s, _ := json.Marshal(testdata)
			// fmt.Println(string(s))

			orderbooks := make([]OrderBookIterator, len(testdata.Reserves))
			dummyTradePath := make([]string, len(testdata.Reserves))
			for index, item := range testdata.Orderbooks {
				orderbooks[index] = &item
				dummyTradePath[index] = "pair" + strconv.Itoa(index)
			}

			// expected outputs will have RequestTxID, shardID zeroed. Those data are out of scope for this test
			var blankAction instruction.Action
			instructions, changedReserves, err := MaybeAcceptTrade(&blankAction, testdata.AmountIn, testdata.Fee, dummyTradePath, blankReceiver, testdata.Reserves, testdata.TradeDirections, common.PRVCoinID, orderbooks)
			encodedResult, _ := json.Marshal(TestResult{instructions, changedReserves})
			// fmt.Println(string(encodedResult))
			if testcase.ExpectSuccess {
				NoError(t, err)
				Equal(t, string(encodedResult), testcase.Expected)
			} else {
				Errorf(t, err, testcase.Expected)
			}

		})
	}
}

type Order struct {
	OrderMatchingInfo
	Id string
}
type OrderList []Order

// replica of PoolPairState.NextOrder()
func (o *OrderList) NextOrder(tradeDirection int) (*OrderMatchingInfo, string, error) {
	lst := []Order(*o)
	lstLen := len(lst)
	switch tradeDirection {
	case TradeDirectionSell0:
		for i := 0; i < lstLen; i++ {
			if lst[i].Token1Balance > 0 {
				return &lst[i].OrderMatchingInfo, lst[i].Id, nil
			}
		}
		// no active order
		return nil, "", nil
	case TradeDirectionSell1:
		for i := lstLen - 1; i >= 0; i-- {
			if lst[i].Token0Balance > 0 {
				return &lst[i].OrderMatchingInfo, lst[i].Id, nil
			}
		}
		// no active order
		return nil, "", nil
	default:
		return nil, "", fmt.Errorf("Invalid trade direction %d", tradeDirection)
	}
}

type Testcase struct {
	Name          string `json:"name"`
	Data          string `json:"data"`
	Expected      string `json:"expected"`
	ExpectSuccess bool   `json:"expectSuccess"`
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
