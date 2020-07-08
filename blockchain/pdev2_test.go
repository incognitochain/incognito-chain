package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/stretchr/testify/suite"
	"strconv"
	"testing"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type PDETestSuiteV2 struct {
	suite.Suite
	currentPDEStateForProducer CurrentPDEState
	currentPDEStateForProcess  CurrentPDEState
}

func (s *PDETestSuiteV2) SetupTest() {
	s.currentPDEStateForProducer = CurrentPDEState{
		WaitingPDEContributions: make(map[string]*rawdbv2.PDEContribution),
		PDEPoolPairs:            make(map[string]*rawdbv2.PDEPoolForPair),
		PDEShares:               make(map[string]uint64),
		PDETradingFees:          make(map[string]uint64),
	}
	s.currentPDEStateForProcess = CurrentPDEState{
		WaitingPDEContributions: make(map[string]*rawdbv2.PDEContribution),
		PDEPoolPairs:            make(map[string]*rawdbv2.PDEPoolForPair),
		PDEShares:               make(map[string]uint64),
		PDETradingFees:          make(map[string]uint64),
	}
}

func buildPDEPRVRequiredContributionActionContent(
	pdeContributionPairID string,
	contributorAddressStr string,
	contributedAmount uint64,
	tokenIDStr string,
) string {
	metadataBase := metadata.MetadataBase{
		Type: metadata.PDEPRVRequiredContributionRequestMeta,
	}
	pdeContribution := metadata.PDEContribution{
		PDEContributionPairID: pdeContributionPairID,
		ContributorAddressStr: contributorAddressStr,
		ContributedAmount:     contributedAmount,
		TokenIDStr:            tokenIDStr,
	}
	pdeContribution.MetadataBase = metadataBase
	actionContent := metadata.PDEContributionAction{
		Meta:    pdeContribution,
		TxReqID: common.Hash{},
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	return base64.StdEncoding.EncodeToString(actionContentBytes)
}

func buildPDEPRVRequiredContributionAction(
	pdeContributionPairID string,
	contributorAddressStr string,
	contributedAmount uint64,
	tokenIDStr string,
) [][]string {
	actionContentBase64Str := buildPDEPRVRequiredContributionActionContent(
		pdeContributionPairID,
		contributorAddressStr,
		contributedAmount,
		tokenIDStr,
	)
	action := []string{strconv.Itoa(metadata.PDEPRVRequiredContributionRequestMeta), actionContentBase64Str}
	return [][]string{action}
}

// All methods that begin with "Test" are run as tests within a
// suite.
// test contribution
func (s *PDETestSuiteV2) TestSimulatedBeaconBlock1001() {
	fmt.Println("Running testcase: TestSimulatedBeaconBlock1001")
	bc := &BlockChain{}
	shardID := byte(1)
	beaconHeight := uint64(1001)

	// case 1: (valid) one of two contributions (in an unique ID) is PRV contribution
	contribInst1 := buildPDEPRVRequiredContributionAction(
		"unique-pair-1",
		"contributor-address-1",
		1000000000000,
		"token-id-1",
	)
	contribInst2 := buildPDEPRVRequiredContributionAction(
		"unique-pair-1",
		"contributor-address-1",
		2000000000000,
		common.PRVIDStr,
	)

	// case 2: (invalid) there is no PRV contribution in an unique ID contribution
	contribInst3 := buildPDEPRVRequiredContributionAction(
		"unique-pair-2",
		"contributor-address-1",
		1000000000000,
		"token-id-1",
	)
	contribInst4 := buildPDEPRVRequiredContributionAction(
		"unique-pair-2",
		"contributor-address-1",
		2000000000000,
		"token-id-2",
	)

	// case 3: (valid) contribute on the existing pool pair with correct proportion
	contribInst5 := buildPDEPRVRequiredContributionAction(
		"unique-pair-1",
		"contributor-address-1",
		500000000000,
		"token-id-1",
	)
	contribInst6 := buildPDEPRVRequiredContributionAction(
		"unique-pair-1",
		"contributor-address-1",
		1000000000000,
		common.PRVIDStr,
	)

	// case 4: (valid) contribute on the existing pool pair with incorrect proportion
	contribInst7 := buildPDEPRVRequiredContributionAction(
		"unique-pair-1",
		"contributor-address-1",
		1000000000000,
		"token-id-1",
	)
	contribInst8 := buildPDEPRVRequiredContributionAction(
		"unique-pair-1",
		"contributor-address-1",
		3000000000000,
		common.PRVIDStr,
	)

	insts := [][]string{
		contribInst1[0], contribInst2[0],
		contribInst3[0], contribInst4[0],
		contribInst5[0], contribInst6[0],
		contribInst7[0], contribInst8[0],
	}

	newInsts := [][]string{}
	for _, inst := range insts {
		metaType, _ := strconv.Atoi(inst[0])
		contentStr := inst[1]
		newInst := [][]string{}
		var err error
		switch metaType {
		case metadata.PDEPRVRequiredContributionRequestMeta:
			newInst, err = bc.buildInstructionsForPDEContribution(contentStr, shardID, metaType, &s.currentPDEStateForProducer, beaconHeight, true)
		//case metadata.PDETradeRequestMeta:
		//	newInst, err = bc.buildInstructionsForPDETrade(contentStr, shardID, metaType, &suite.currentPDEStateForProducer, beaconHeight-1)
		//case metadata.PDEWithdrawalRequestMeta:
		//	newInst, err = bc.buildInstructionsForPDEWithdrawal(contentStr, shardID, metaType, &suite.currentPDEStateForProducer, beaconHeight-1)
		default:
			continue
		}
		s.Equal(err, nil)
		newInsts = append(newInsts, newInst...)
	}

	//s.Equal(len(newInsts), 3)

	// check result of functions
	// case 1:
	s.Equal(newInsts[0][2], "waiting")
	s.Equal(newInsts[1][2], "matched")

	// case 2: to be continued ...
}

func TestPDETestSuiteV2(t *testing.T) {
	suite.Run(t, new(PDETestSuiteV2))
}
