package pdex

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/jrick/logrotate/rotator"
)

const (
	paymentAddress0 = "12Rs8bHvYZELqHrv28bYezBQQpteZUEbYjUf2oqV9pJm6Gx4sD4n9mr4UgQe5cDeP9A2x1DsB4mbJ9LT8x2ShaY41cZJWrL7RpFpp2v"
	tempPToken      = "41fe8c2f89cce24c0b798ec0fa10ac9cd6f0d273249922e92cb26412df989830"
)

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
