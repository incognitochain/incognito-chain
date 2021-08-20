package pdex

import (
	"encoding/json"
	"fmt"
	"testing"

	// "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	. "github.com/stretchr/testify/assert"
)

var _ = fmt.Print

func TestProduceOrder(t *testing.T) {
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

			env := skipToProduce(&testdata.Metadata, 0)
			testState := mustReadState("test_state.json")
			// manually add nftID
			testState.nftIDs[testdata.Metadata.NftID.String()] = true
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
