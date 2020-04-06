package blockchain

// Basic imports
import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type PDEProcessSuite struct {
	suite.Suite
	currentPDEState *CurrentPDEState
}

func (suite *PDEProcessSuite) SetupTest() {
	suite.currentPDEState = &CurrentPDEState{
		WaitingPDEContributions: make(map[string]*rawdbv2.PDEContribution),
		PDEPoolPairs:            make(map[string]*rawdbv2.PDEPoolForPair),
		PDEShares:               make(map[string]uint64),
	}
}

func buildPDEContributionActionContent(
	pdeContributionPairID string,
	contributorAddressStr string,
	contributedAmount uint64,
	tokenIDStr string,
) string {
	metadataBase := metadata.MetadataBase{
		Type: metadata.PDEContributionMeta,
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

func buildPDEContributionAction(
	pdeContributionPairID string,
	contributorAddressStr string,
	contributedAmount uint64,
	tokenIDStr string,
) [][]string {
	actionContentBase64Str := buildPDEContributionActionContent(
		pdeContributionPairID,
		contributorAddressStr,
		contributedAmount,
		tokenIDStr,
	)
	action := []string{strconv.Itoa(metadata.PDEContributionMeta), actionContentBase64Str}
	return [][]string{action}
}

func buildPDEContributionInst(
	pdeContributionPairID string,
	contributorAddressStr string,
	contributedAmount uint64,
	tokenIDStr string,
) [][]string {
	actionContentBase64Str := buildPDEContributionActionContent(
		pdeContributionPairID,
		contributorAddressStr,
		contributedAmount,
		tokenIDStr,
	)
	shardID := byte(1)
	inst := []string{
		strconv.Itoa(metadata.PDEContributionMeta),
		strconv.Itoa(int(shardID)),
		"accepted",
		actionContentBase64Str,
	}
	return [][]string{inst}
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *PDEProcessSuite) TestPDEContributionOnUnexistedWaitingUniqID() {
	fmt.Println("Running testcase: TestPDEContributionOnUnexistedWaitingUniqID")
	uniqPairID := "123"
	contributorAddr := "12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj"
	contributedAmt := uint64(10000000000)
	contribTokenIDStr := "0000000000000000000000000000000000000000000000000000000000000005"
	contribInsts := buildPDEContributionInst(
		uniqPairID,
		contributorAddr,
		contributedAmt,
		contribTokenIDStr,
	)
	beaconHeight := uint64(1001)
	bc := &BlockChain{}
	err := bc.processPDEContributionV2(&statedb.StateDB{}, beaconHeight-1, contribInsts[0], suite.currentPDEState)
	suite.Equal(err, nil)
	waitingContribKey := string(rawdbv2.BuildWaitingPDEContributionKey(
		beaconHeight-1,
		uniqPairID,
	))

	suite.Equal(len(suite.currentPDEState.PDEPoolPairs), 0)
	suite.Equal(len(suite.currentPDEState.PDEShares), 0)
	suite.Equal(len(suite.currentPDEState.WaitingPDEContributions), 1)
	suite.Equal(suite.currentPDEState.WaitingPDEContributions[waitingContribKey].ContributorAddressStr, contributorAddr)
	suite.Equal(suite.currentPDEState.WaitingPDEContributions[waitingContribKey].TokenIDStr, contribTokenIDStr)
	suite.Equal(suite.currentPDEState.WaitingPDEContributions[waitingContribKey].Amount, contributedAmt)
}

func (suite *PDEProcessSuite) TestPDEContributionOnUnexistedPairForExistedWaitingUniqID() {
	fmt.Println("Running testcase: TestPDEContributionOnUnexistedPairForExistedWaitingUniqID")
	uniqPairID := "123"
	contributorAddr := "12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj"
	contributedAmt := uint64(10000000000)
	contribToken1IDStr := "0000000000000000000000000000000000000000000000000000000000000005"
	contribToken2IDStr := "0000000000000000000000000000000000000000000000000000000000000007"
	beaconHeight := uint64(1001)

	existedWaitingContribKey := string(rawdbv2.BuildWaitingPDEContributionKey(
		beaconHeight-1,
		uniqPairID,
	))
	currentPDEState := suite.currentPDEState
	currentPDEState.WaitingPDEContributions[existedWaitingContribKey] = &rawdbv2.PDEContribution{
		ContributorAddressStr: contributorAddr,
		TokenIDStr:            contribToken1IDStr,
		Amount:                20000000000,
	}

	contribInsts := buildPDEContributionInst(
		uniqPairID,
		contributorAddr,
		contributedAmt,
		contribToken2IDStr,
	)
	bc := &BlockChain{}
	err := bc.processPDEContributionV2(&statedb.StateDB{}, beaconHeight-1, contribInsts[0], suite.currentPDEState)
	suite.Equal(err, nil)
	_, found := currentPDEState.WaitingPDEContributions[existedWaitingContribKey]
	suite.Equal(found, false)
	suite.Equal(len(suite.currentPDEState.PDEPoolPairs), 1)
	suite.Equal(len(suite.currentPDEState.PDEShares), 2)
	suite.Equal(len(suite.currentPDEState.WaitingPDEContributions), 0)

	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight-1, contribToken1IDStr, contribToken2IDStr))
	newPair := suite.currentPDEState.PDEPoolPairs[pairKey]
	suite.Equal(newPair.Token1IDStr, contribToken1IDStr)
	suite.Equal(newPair.Token2IDStr, contribToken2IDStr)
	suite.Equal(newPair.Token1PoolValue, uint64(20000000000))
	suite.Equal(newPair.Token2PoolValue, uint64(10000000000))

	shareKey1 := string(rawdbv2.BuildPDESharesKey(beaconHeight-1, contribToken1IDStr, contribToken2IDStr, contribToken1IDStr, contributorAddr))
	shareKey2 := string(rawdbv2.BuildPDESharesKey(beaconHeight-1, contribToken1IDStr, contribToken2IDStr, contribToken2IDStr, contributorAddr))

	suite.Equal(suite.currentPDEState.PDEShares[shareKey1], uint64(20000000000))
	suite.Equal(suite.currentPDEState.PDEShares[shareKey2], uint64(10000000000))
}

func (suite *PDEProcessSuite) TestPDEContributionOnExistedPairForExistedWaitingUniqID() {
	fmt.Println("Running testcase: TestPDEContributionOnExistedPairForExistedWaitingUniqID")
	uniqPairID1 := "123"
	uniqPairID2 := "456"
	contributorAddr := "12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj"
	oldContribTokenIDStr := "0000000000000000000000000000000000000000000000000000000000000003"
	contribToken1IDStr := "0000000000000000000000000000000000000000000000000000000000000005"
	contribToken2IDStr := "0000000000000000000000000000000000000000000000000000000000000007"
	contributedAmt := uint64(10000000000)
	beaconHeight := uint64(1001)

	currentPDEState := suite.currentPDEState
	// waiting contribution
	existedWaitingContribKey1 := string(rawdbv2.BuildWaitingPDEContributionKey(
		beaconHeight-1,
		uniqPairID1,
	))
	currentPDEState.WaitingPDEContributions[existedWaitingContribKey1] = &rawdbv2.PDEContribution{
		ContributorAddressStr: contributorAddr,
		TokenIDStr:            contribToken1IDStr,
		Amount:                20000000000,
	}
	existedWaitingContribKey2 := string(rawdbv2.BuildWaitingPDEContributionKey(
		beaconHeight-1,
		uniqPairID2,
	))
	currentPDEState.WaitingPDEContributions[existedWaitingContribKey2] = &rawdbv2.PDEContribution{
		ContributorAddressStr: contributorAddr,
		TokenIDStr:            contribToken1IDStr,
		Amount:                80000000000,
	}

	// pool pairs
	existedPoolPairKey1 := string(rawdbv2.BuildPDEPoolForPairKey(
		beaconHeight-1,
		contribToken1IDStr,
		contribToken2IDStr,
	))
	existedPoolPairKey2 := string(rawdbv2.BuildPDEPoolForPairKey(
		beaconHeight-1,
		oldContribTokenIDStr,
		contribToken1IDStr,
	))
	currentPDEState.PDEPoolPairs[existedPoolPairKey1] = &rawdbv2.PDEPoolForPair{
		Token1IDStr:     contribToken1IDStr,
		Token1PoolValue: 50000000000,
		Token2IDStr:     contribToken2IDStr,
		Token2PoolValue: 80000000000,
	}
	currentPDEState.PDEPoolPairs[existedPoolPairKey2] = &rawdbv2.PDEPoolForPair{
		Token1IDStr:     oldContribTokenIDStr,
		Token1PoolValue: 10000000000,
		Token2IDStr:     contribToken1IDStr,
		Token2PoolValue: 90000000000,
	}

	// shares
	shareKey1 := string(rawdbv2.BuildPDESharesKey(
		beaconHeight-1,
		contribToken1IDStr,
		contribToken2IDStr,
		contribToken1IDStr,
		contributorAddr,
	))
	shareKey2 := string(rawdbv2.BuildPDESharesKey(
		beaconHeight-1,
		contribToken1IDStr,
		contribToken2IDStr,
		contribToken2IDStr,
		contributorAddr,
	))
	shareKey3 := string(rawdbv2.BuildPDESharesKey(
		beaconHeight-1,
		oldContribTokenIDStr,
		contribToken2IDStr,
		oldContribTokenIDStr,
		contributorAddr+"-new",
	))
	currentPDEState.PDEShares[shareKey1] = 10000000000
	currentPDEState.PDEShares[shareKey2] = 10000000000
	currentPDEState.PDEShares[shareKey3] = 10000000000

	contribInsts := buildPDEContributionInst(
		uniqPairID1,
		contributorAddr,
		contributedAmt,
		contribToken2IDStr,
	)
	bc := &BlockChain{}
	err := bc.processPDEContributionV2(&statedb.StateDB{}, beaconHeight-1, contribInsts[0], suite.currentPDEState)
	suite.Equal(err, nil)
	newWaitingPDEContributions := suite.currentPDEState.WaitingPDEContributions
	suite.Equal(len(newWaitingPDEContributions), 1)
	waitingContrib, found := newWaitingPDEContributions[existedWaitingContribKey2]
	suite.Equal(found, true)
	suite.Equal(waitingContrib.Amount, uint64(80000000000))

	newPoolPairs := suite.currentPDEState.PDEPoolPairs
	suite.Equal(len(newPoolPairs), 2)
	suite.Equal(newPoolPairs[existedPoolPairKey1].Token1PoolValue, uint64(50000000000+20000000000))
	suite.Equal(newPoolPairs[existedPoolPairKey1].Token2PoolValue, uint64(80000000000+contributedAmt))

	newShares := suite.currentPDEState.PDEShares
	suite.Equal(len(newShares), 3)
	suite.Equal(newShares[shareKey1], uint64(14000000000))
	suite.Equal(newShares[shareKey2], uint64(11250000000))
	suite.Equal(newShares[shareKey3], uint64(10000000000))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestPDEProcessSuite(t *testing.T) {
	fmt.Println("Initialized...")
	suite.Run(t, new(PDEProcessSuite))
}
