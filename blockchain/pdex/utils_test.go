package pdex

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/jrick/logrotate/rotator"
)

var (
	wrarperDB statedb.DatabaseAccessWarper
	diskDB    incdb.Database
)

const (
	paymentAddress0   = "12Rs8bHvYZELqHrv28bYezBQQpteZUEbYjUf2oqV9pJm6Gx4sD4n9mr4UgQe5cDeP9A2x1DsB4mbJ9LT8x2ShaY41cZJWrL7RpFpp2v"
	tempPToken        = "41fe8c2f89cce24c0b798ec0fa10ac9cd6f0d273249922e92cb26412df989830"
	validOTAReceiver0 = "15sXoyo8kCZCHjurNC69b8WV2jMCvf5tVrcQ5mT1eH9Nm351XRjE1BH4WHHLGYPZy9dxTSLiKQd6KdfoGq4yb4gP1AU2oaJTeoGymsEzonyi1XSW2J2U7LeAVjS1S2gjbNDk1t3f9QUg2gk4"
	validOTAReceiver1 = "15ujixNQY1Qc5wyX9UYQW3s6cbcecFPNhrWjWiFCggeN5HukPVdjbKyRE3goUpFgZhawtBtRUK3ZSZb5LtH7bevhGzz3UTh1muzLHG3pvsE6RNB81y8xNGhyHdpHZfjwmSWDdwDe74Tg2CUP"
	nftID             = "9b2966a3ef898ebcd1fc41369cd00165624e6d8f555cdd7a2aae274380b1ea79"
	nftID1            = "4c11a0f6b7aaf8e4ba0c42969fe8ba369adfc409aa3a1ef94db234f835fd3d67"
	newNftID          = "662a1d89dc4134f32f06799a3e65cdeed4ff84aa4070b16e37c7208426d12c53"
	poolPairID        = "0000000000000000000000000000000000000000000000000000000000000123-0000000000000000000000000000000000000000000000000000000000000456-0000000000000000000000000000000000000000000000000000000000000abc"
	poolPairPRV       = "0000000000000000000000000000000000000000000000000000000000000004-0000000000000000000000000000000000000000000000000000000000000123-0000000000000000000000000000000000000000000000000000000000000bcd"
	newPoolPairID     = "0000000000000000000000000000000000000000000000000000000000000123-0000000000000000000000000000000000000000000000000000000000000456-0000000000000000000000000000000000000000000000000000000000111000"
)

func initDB() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "data")
	if err != nil {
		panic(err)
	}
	diskDB, _ = incdb.Open("leveldb", dbPath)
	wrarperDB = statedb.NewDatabaseAccessWarper(diskDB)
}

// initLogRotator initializes the logging rotater to write logs to logFile and
// create roll files in the same directory.  It must be called before the
// package-global log rotater variables are used.
func initLogRotator(logFile string) {
	logDir, _ := filepath.Split(logFile)
	err := os.MkdirAll(logDir, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log directory: %v\n", err)
		os.Exit(utils.ExitByLogging)
	}
	r, err := rotator.New(logFile, 10*1024, false, 3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create file rotator: %v\n", err)
		os.Exit(utils.ExitByLogging)
	}

	logRotator = r
}

// logWriter implements an io.Writer that outputs to both standard output and
// the write-end pipe of an initialized log rotator.
type logWriter struct{}

var logRotator *rotator.Rotator

func (logWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write(p)
	logRotator.Write(p)
	return len(p), nil
}

func initLog() {
	initLogRotator("./test-pdex.log")
	pdexLogger := common.NewBackend(logWriter{}).Logger("PDEX log ", false)
	Logger.Init(pdexLogger)
}

func buildFeeWithdrawalRequestActionForTest(
	withdrawerAddressStr string,
	withdrawalToken1IDStr string,
	withdrawalToken2IDStr string,
	withdrawalFeeAmt uint64,
) []string {
	feeWithdrawalRequest := metadata.PDEFeeWithdrawalRequest{
		WithdrawerAddressStr:  withdrawerAddressStr,
		WithdrawalToken1IDStr: withdrawalToken1IDStr,
		WithdrawalToken2IDStr: withdrawalToken2IDStr,
		WithdrawalFeeAmt:      withdrawalFeeAmt,
		MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
			MetadataBase: metadata.MetadataBase{
				Type: metadata.PDEFeeWithdrawalRequestMeta,
			},
		},
	}
	actionContent := metadata.PDEFeeWithdrawalRequestAction{
		Meta:    feeWithdrawalRequest,
		TxReqID: common.Hash{},
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(metadata.PDEFeeWithdrawalRequestMeta), actionContentBase64Str}
	return action
}

func convertToLPFeesPerShare(totalFee uint64, totalShare uint64) *big.Int {
	result := new(big.Int).Mul(new(big.Int).SetUint64(totalFee), BaseLPFeesPerShare)
	result = new(big.Int).Div(result, new(big.Int).SetUint64(totalShare))
	return result
}

func getStrPoolPairState(p *PoolPairState) string {
	result := "\n"
	result += fmt.Sprintf("Pool Pair: %+v\n", p)
	result += fmt.Sprintf("Virtual amount: token0 %v, token1 %v\n", p.state.Token0VirtualAmount().String(), p.state.Token1VirtualAmount().String())
	result += fmt.Sprintf("LP Fees Per Share\n")
	for tokenID := range p.lpFeesPerShare {
		result += fmt.Sprintf(" %v: %v\n", tokenID, p.lpFeesPerShare[tokenID].String())
	}
	result += fmt.Sprintf("LM Rewards Per Share\n")
	for tokenID := range p.lmRewardsPerShare {
		result += fmt.Sprintf(" %v: %v\n", tokenID, p.lmRewardsPerShare[tokenID].String())
	}
	result += fmt.Sprintf("Shares\n")
	for tokenID, value := range p.shares {
		result += fmt.Sprintf(" %v: %+v\n", tokenID, value)
		for _tokenID := range value.lastLPFeesPerShare {
			result += fmt.Sprintf(" Fee %v: %v\n", _tokenID, value.lastLPFeesPerShare[_tokenID])
		}
		for _tokenID := range value.lastLmRewardsPerShare {
			result += fmt.Sprintf(" LM %v: %v\n", _tokenID, value.lastLmRewardsPerShare[_tokenID])
		}
	}
	result += fmt.Sprintf("Order Rewards\n")
	for nftID, value := range p.orderRewards {
		result += fmt.Sprintf(" %v: %v\n", nftID, value)
		for tokenID, amount := range value.uncollectedRewards {
			result += fmt.Sprintf("  %v: %v\n", tokenID, amount)
		}
	}
	result += fmt.Sprintf("Making Volume\n")
	for tokenID, value := range p.makingVolume {
		result += fmt.Sprintf(" %v: %v\n", tokenID.String(), value)
		for nftID, amount := range value.volume {
			result += fmt.Sprintf("  %v: %v\n", nftID, amount.String())
		}
	}
	result += fmt.Sprintf("Locked Record\n")
	for shareID, record := range p.lmLockedShare {
		result += fmt.Sprintf(" Share %v: %+v\n", shareID, record)
	}
	return result
}

func getStrStakingPoolState(p *StakingPoolState) string {
	result := "\n"
	result += fmt.Sprintf("Staking Pool: %+v\n", p)
	result += fmt.Sprintf("Total liquidity: %v\n", p.liquidity)
	for tokenID, value := range p.stakers {
		result += fmt.Sprintf(" %v: %v\n", tokenID, value)
	}
	return result
}
