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

			clonedState, ok := testState.Clone().(*stateV2)
			True(t, ok)
			blankState := newStateV2()
			sc := NewStateChange()
			_, sc, err = testState.GetDiff(blankState, sc)
			NoError(t, err)
			err = testState.StoreToDB(env, sc)
			NoError(t, err)
			
			loadedState, err := initStateV2(testDB, 0)
			NoError(t, err)
			temp := (&StateFormatter{}).FromState(clonedState)
			encodedClonedState, _ := json.Marshal(TestResult(*temp))
			temp = (&StateFormatter{}).FromState(loadedState)
			encodedLoadedState, _ := json.Marshal(TestResult(*temp))
			Equal(t, string(encodedClonedState), string(encodedLoadedState))

			// err = newState.Process(env)
			// NoError(t, err)

			// temp = (&StateFormatter{}).FromState(testState)
			// encodedResult, _ := json.Marshal(TestResult(*temp))
			// Equal(t, testcase.Expected, string(encodedResult))
		})
	}
}

// func TestBuildResponseOrder(t *testing.T) {
// 	type TestData struct {
// 		Instructions [][]string `json:"instructions"`
// 	}

// 	type TestResult struct {
// 		Tx metadataCommon.Transaction `json:"tx"`
// 	}

// 	var testcases []Testcase
// 	testcases = append(testcases, buildResponseTradeTestcases...)
// 	var blankPrivateKey privacy.PrivateKey = make([]byte, 32)
// 	// use a fixed, non-zero private key for testing
// 	blankPrivateKey[3] = 10

// 	var blankShardID byte = 0
// 	for _, testcase := range testcases {
// 		t.Run(testcase.Name, func(t *testing.T) {
// 			var testdata TestData
// 			err := json.Unmarshal([]byte(testcase.Data), &testdata)
// 			NoError(t, err)

// 			myInstruction := testdata.Instructions[0]
// 			metaType, err := strconv.Atoi(myInstruction[0])
// 			NoError(t, err)

// 			tx, err := (&TxBuilderV2{}).Build(
// 				metaType,
// 				myInstruction,
// 				&blankPrivateKey,
// 				blankShardID,
// 				testDB,
// 				testDB,
// 			)
// 			NoError(t, err)
// 			txv2, ok := tx.(*transaction.TxTokenVersion2)
// 			True(t, ok)
// 			mintedCoin, ok := txv2.TokenData.
// 				Proof.GetOutputCoins()[0].(*privacy.CoinV2)
// 			True(t, ok)

// 			var expectedTx transaction.TxTokenVersion2
// 			err = json.Unmarshal([]byte(testcase.Expected), &expectedTx)
// 			NoError(t, err)
// 			expectedMintedCoin, ok := expectedTx.TokenData.Proof.GetOutputCoins()[0].(*privacy.CoinV2)
// 			True(t, ok)
// 			// check token id, receiver & value
// 			Equal(t, expectedTx.TokenData.PropertyID, txv2.TokenData.PropertyID)
// 			True(t, bytes.Equal(expectedMintedCoin.GetPublicKey().ToBytesS(),
// 				mintedCoin.GetPublicKey().ToBytesS()))
// 			Equal(t, expectedMintedCoin.GetValue(), mintedCoin.GetValue())
// 		})
// 	}
// }
