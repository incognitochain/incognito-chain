package pdex

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

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
		Orders []*Order `json:"orders"`
	}

	var testcases []Testcase
	testcases = append(testcases, sortOrderTestcases...)

	testState := newStateV2WithValue(nil, nil, make(map[string]*PoolPairState),
		&Params{}, nil, map[string]uint64{})
	blankPairID := "pair0"
	testState.poolPairs[blankPairID] = &PoolPairState{orderbook: Orderbook{[]*Order{}}}
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal([]byte(testcase.Data), &testdata)
			NoError(t, err)
			// get a random permutation of orders in test data for inserting
			// since this test inserts items at random order, it is not compatible for testing equality-breaking of orders
			perm := rand.Perm(len(testdata.Orders))
			var orderbookPerm []*Order
			for _, newInd := range perm {
				orderbookPerm = append(orderbookPerm, testdata.Orders[newInd])
			}
			testdata.Orders = orderbookPerm

			for _, item := range testdata.Orders {
				pair := testState.poolPairs[blankPairID]
				pair.orderbook.InsertOrder(item)
				testState.poolPairs[blankPairID] = pair
			}
			encodedResult, _ := json.Marshal(TestResult{testState.poolPairs[blankPairID].orderbook.orders})
			Equal(t, testcase.Expected, string(encodedResult))
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
			err := json.Unmarshal([]byte(testcase.Data), &testdata)
			NoError(t, err)

			env := skipToProduce([]metadataCommon.Metadata{&testdata.Metadata}, 0)
			testState := mustReadState("test_state.json")
			// manually add nftID
			testState.nftIDs[testdata.Metadata.NftID.String()] = 100
			temp := &StateFormatter{}
			temp.FromState(testState)

			instructions, err := testState.BuildInstructions(env)
			NoError(t, err)

			encodedResult, _ := json.Marshal(TestResult{instructions})
			Equal(t, testcase.Expected, string(encodedResult))
		})
	}
}

func TestProcessOrder(t *testing.T) {
	setTestTradeConfig()
	type TestData struct {
		Instructions [][]string `json:"instructions"`
	}

	type TestResult StateFormatter

	var testcases []Testcase = mustReadTestcases("process_order.json")
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var testdata TestData
			err := json.Unmarshal([]byte(testcase.Data), &testdata)
			NoError(t, err)

			env := skipToProcess(testdata.Instructions)
			testState := mustReadState("test_state.json")
			err = testState.Process(env)
			NoError(t, err)

			temp := (&StateFormatter{}).FromState(testState)
			encodedResult, _ := json.Marshal(TestResult(*temp))
			Equal(t, testcase.Expected, string(encodedResult))
		})
	}
}