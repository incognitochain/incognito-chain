package bridgeagg

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	"github.com/stretchr/testify/suite"
)

type ShieldTestCase struct {
	Metadatas             []*metadataBridge.ShieldRequest `json:"metadatas"`
	ExpectedInstructions  [][]string                      `json:"expected_instructions"`
	UnifiedTokens         map[common.Hash]map[uint]*Vault `json:"unified_tokens"`
	ExpectedUnifiedTokens map[common.Hash]map[uint]*Vault `json:"expected_unified_tokens"`
	TxID                  common.Hash                     `json:"tx_id"`
}

type ShieldTestSuite struct {
	suite.Suite
	testCases            []ShieldTestCase
	currentTestCaseIndex int
	actualResults        []ActualResult

	sdb *statedb.StateDB
}
