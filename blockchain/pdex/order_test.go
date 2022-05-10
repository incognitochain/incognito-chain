package pdex

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	. "github.com/stretchr/testify/assert"
)

var _ = fmt.Print

func TestSortOrder(t *testing.T) {
	type TestData struct {
		Orders []*Order `json:"orders"`
	}

	type TestResult struct {
		Orders         []*Order `json:"orders"`
		MatchTradeBuy0 string
		MatchTradeBuy1 string
	}

	var testcases []Testcase
	testcases = append(testcases, sortOrderTestcases...)

	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			// initialize test state & order book
			testState := newStateV2WithValue(nil, nil, make(map[string]*PoolPairState),
				&Params{}, nil, map[string]uint64{})
			blankPairID := "pair0"
			testState.poolPairs[blankPairID] = &PoolPairState{orderbook: Orderbook{[]*Order{}}}

			var testdata TestData
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			err = json.Unmarshal(testcase.Expected, &expected)
			NoError(t, err)

			// get a random permutation of orders in test data for inserting
			perm := rand.Perm(len(testdata.Orders))
			var orderbookPerm []*Order
			for _, newInd := range perm {
				orderbookPerm = append(orderbookPerm, testdata.Orders[newInd])
			}
			testdata.Orders = orderbookPerm
			// insert the orders. Result will be sorted
			for _, item := range testdata.Orders {
				pair := testState.poolPairs[blankPairID]
				pair.orderbook.InsertOrder(item)
				testState.poolPairs[blankPairID] = pair
			}

			result := TestResult{Orders: testState.poolPairs[blankPairID].orderbook.orders}
			// test the outputs of NextOrder()
			ord, id, err := testState.poolPairs[blankPairID].orderbook.NextOrder(v2utils.TradeDirectionSell0)
			NoError(t, err)
			Equal(t, ord.Id(), id)
			result.MatchTradeBuy1 = id
			ord, id, err = testState.poolPairs[blankPairID].orderbook.NextOrder(v2utils.TradeDirectionSell1)
			NoError(t, err)
			Equal(t, ord.Id(), id)
			result.MatchTradeBuy0 = id

			Equal(t, expected, result)
		})
	}
}

func TestProduceOrder(t *testing.T) {
	setTestTradeConfig()
	type TestData struct {
		Metadata metadataPdexv3.AddOrderRequest `json:"metadata"`
	}

	type TestResult struct {
		Instructions [][]string `json:"instructions"`
	}

	var testcases []Testcase = mustReadTestcases("produce_order.json")
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			err = json.Unmarshal(testcase.Expected, &expected)
			NoError(t, err)
			testState := mustReadState("test_state.json")

			env := skipToProduce([]metadataCommon.Metadata{&testdata.Metadata}, 0)
			// manually add nftID
			if testdata.Metadata.NftID != nil {
				testState.nftIDs[testdata.Metadata.NftID.String()] = 100
			}

			instructions, err := testState.BuildInstructions(env)
			NoError(t, err)
			Equal(t, expected, TestResult{instructions})
		})
	}
}

func TestProduceWithdrawOrder(t *testing.T) {
	setTestTradeConfig()
	type TestData struct {
		Metadata metadataPdexv3.WithdrawOrderRequest `json:"metadata"`
	}

	type TestResult struct {
		Instructions [][]string `json:"instructions"`
	}

	var testcases []Testcase = mustReadTestcases("produce_withdraw_order.json")
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			err = json.Unmarshal(testcase.Expected, &expected)
			NoError(t, err)
			testState := mustReadState("test_state.json")

			env := skipToProduce([]metadataCommon.Metadata{&testdata.Metadata}, 0)
			// manually add nftID
			if testdata.Metadata.NftID != nil {
				testState.nftIDs[testdata.Metadata.NftID.String()] = 100
			}

			instructions, err := testState.BuildInstructions(env)
			NoError(t, err)
			Equal(t, expected, TestResult{instructions})
		})
	}
}

func TestAutoWithdraw(t *testing.T) {
	setTestTradeConfig()
	type TestData struct {
		State StateFormatter `json:"state"`
		Limit uint           `json:"limit"`
	}

	type TestResult struct {
		Instructions [][]string `json:"instructions"`
	}

	var testcases []Testcase = mustReadTestcases("auto_withdraw_order.json")
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			err = json.Unmarshal(testcase.Expected, &expected)
			NoError(t, err)
			testState := testdata.State.State(&Params{
				MaxOrdersPerNft:   DefaultTestMaxOrdersPerNft,
				DefaultFeeRateBPS: 30,
			})

			instructions, _, _, err := testState.producer.withdrawAllMatchedOrders(testState.poolPairs, testdata.Limit)
			NoError(t, err)
			Equal(t, expected, TestResult{instructions})
		})
	}
}

func TestOrderOverNftIDLimit(t *testing.T) {
	setTestTradeConfig()

	type TestData struct {
		Metadata metadataPdexv3.AddOrderRequest `json:"metadata"`
		Repeat   uint                           `json:"repeat"`
	}

	type TestResult struct {
		Instructions [][]string `json:"instructions"`
	}

	var testcases []Testcase = mustReadTestcases("produce_order_over_limit.json")
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			err = json.Unmarshal(testcase.Expected, &expected)
			NoError(t, err)
			testState := mustReadState("test_state.json")

			// repeat the same metadata to simulate producing multiple orders of the same NftID in 1 block
			var mds []metadataCommon.Metadata
			for i := 0; i < int(testdata.Repeat); i++ {
				var temp metadataPdexv3.AddOrderRequest = testdata.Metadata
				mds = append(mds, &temp)
			}

			env := skipToProduce(mds, 0)
			// manually add nftID
			if testdata.Metadata.NftID != nil {
				testState.nftIDs[testdata.Metadata.NftID.String()] = 100
			}
			// set order count per NFT to 2 for this test
			testState.params.MaxOrdersPerNft = 2

			instructions, err := testState.BuildInstructions(env)
			NoError(t, err)
			Equal(t, expected, TestResult{instructions})
		})
	}
}

func TestProcessOrder(t *testing.T) {
	setTestTradeConfig()
	type TestData struct {
		Instructions [][]string `json:"instructions"`
	}

	type TestResult = StateFormatter

	var testcases []Testcase = mustReadTestcases("process_order.json")
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			err = json.Unmarshal(testcase.Expected, &expected)
			NoError(t, err)
			testState := mustReadState("test_state.json")

			env := skipToProcess(testdata.Instructions)
			err = testState.Process(env)
			NoError(t, err)

			var result TestResult
			result.FromState(testState)
			Equal(t, expected, result)
		})
	}
}

func TestProcessOrderReward(t *testing.T) {
	setTestTradeConfig()
	type TestData struct {
		Instructions [][]string `json:"instructions"`
	}

	type TestResult StateFormatter

	var testcases []Testcase = mustReadTestcases("process_trade_order_reward.json")
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal(testcase.Data, &testdata)
			NoError(t, err)
			var expected TestResult
			err = json.Unmarshal(testcase.Expected, &expected)
			NoError(t, err)
			testState := mustReadState("test_state_order_reward.json", "params.json")

			env := skipToProcess(testdata.Instructions)
			err = testState.Process(env)
			NoError(t, err)
			result := (*TestResult)((&StateFormatter{}).FromState(testState))

			Equal(t, expected, *result)
		})
	}
}
