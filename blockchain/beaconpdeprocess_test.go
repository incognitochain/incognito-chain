package blockchain

// Basic imports
import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database/lvdb"
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
		WaitingPDEContributions: make(map[string]*lvdb.PDEContribution),
		PDEPoolPairs:            make(map[string]*lvdb.PDEPoolForPair),
		PDEShares:               make(map[string]uint64),
	}
}

func buildPDEContributionInst(
	pdeContributionPairID string,
	contributorAddressStr string,
	contributedAmount uint64,
	tokenIDStr string,
) [][]string {
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
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)

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
	err := bc.processPDEContribution(beaconHeight-1, contribInsts[0], suite.currentPDEState)
	suite.Equal(err, nil)
	waitingContribKey := string(lvdb.BuildWaitingPDEContributionKey(
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

func (suite *PDEProcessSuite) TestPDEContributionOnExistedWaitingUniqID() {
	fmt.Println("Running testcase: TestPDEContributionOnExistedWaitingUniqID")
	uniqPairID := "123"
	contributorAddr := "12S2jM1TBbX2V5TBTvpJkJmsdaYxbCspGNedQkvJpYcbnV4gad7FDEbzY9P3zbpZRJTsGD5vxJRia3UiiUwMUbXbjfgezewq6rtPNtj"
	contributedAmt := uint64(10000000000)
	contribToken1IDStr := "0000000000000000000000000000000000000000000000000000000000000005"
	contribToken2IDStr := "0000000000000000000000000000000000000000000000000000000000000007"
	beaconHeight := uint64(1001)

	existedWaitingContribKey := string(lvdb.BuildWaitingPDEContributionKey(
		beaconHeight-1,
		uniqPairID,
	))
	currentPDEState := suite.currentPDEState
	currentPDEState.WaitingPDEContributions[existedWaitingContribKey] = &lvdb.PDEContribution{
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
	err := bc.processPDEContribution(beaconHeight-1, contribInsts[0], suite.currentPDEState)
	suite.Equal(err, nil)
	_, found := currentPDEState.WaitingPDEContributions[existedWaitingContribKey]
	suite.Equal(found, false)
	suite.Equal(len(suite.currentPDEState.PDEPoolPairs), 1)
	suite.Equal(len(suite.currentPDEState.PDEShares), 2)
	suite.Equal(len(suite.currentPDEState.WaitingPDEContributions), 0)

	pairKey := string(lvdb.BuildPDEPoolForPairKey(beaconHeight-1, contribToken1IDStr, contribToken2IDStr))
	newPair := suite.currentPDEState.PDEPoolPairs[pairKey]
	suite.Equal(newPair.Token1IDStr, contribToken1IDStr)
	suite.Equal(newPair.Token2IDStr, contribToken2IDStr)
	suite.Equal(newPair.Token1PoolValue, uint64(20000000000))
	suite.Equal(newPair.Token2PoolValue, uint64(10000000000))

	shareKey1 := string(lvdb.BuildPDESharesKey(beaconHeight-1, contribToken1IDStr, contribToken2IDStr, contribToken1IDStr, contributorAddr))
	shareKey2 := string(lvdb.BuildPDESharesKey(beaconHeight-1, contribToken1IDStr, contribToken2IDStr, contribToken2IDStr, contributorAddr))

	suite.Equal(suite.currentPDEState.PDEShares[shareKey1], uint64(20000000000))
	suite.Equal(suite.currentPDEState.PDEShares[shareKey2], uint64(10000000000))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestPDEProcessSuite(t *testing.T) {
	fmt.Println("Initialized...")
	suite.Run(t, new(PDEProcessSuite))
}
