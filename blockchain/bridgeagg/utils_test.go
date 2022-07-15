package bridgeagg

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/metadata/evmcaller"
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/jrick/logrotate/rotator"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	currentTestCaseName string
	actualResults       map[string]ActualResult
	sDB                 *statedb.StateDB
}

type TestData struct {
	State             *State                                        `json:"state"`
	TxIDs             []common.Hash                                 `json:"tx_ids"`
	BridgeTokensInfo  map[common.Hash]*statedb.BridgeTokenInfoState `json:"bridge_tokens_info"`
	AccumulatedValues *metadata.AccumulatedValues                   `json:"accumulated_values"`
	env               *stateEnvironment
}

type ExpectedResult struct {
	State             *State                                   `json:"state"`
	Instructions      [][]string                               `json:"instructions"`
	AccumulatedValues *metadata.AccumulatedValues              `json:"accumulated_values"`
	BridgeTokensInfo  map[common.Hash]*rawdbv2.BridgeTokenInfo `json:"bridge_tokens_info"`
}

type ActualResult struct {
	Instructions      [][]string
	ProducerState     *State
	ProcessorState    *State
	AccumulatedValues *metadata.AccumulatedValues
	BridgeTokensInfo  map[common.Hash]*rawdbv2.BridgeTokenInfo
}

var _ = func() (_ struct{}) {
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	evmcaller.Logger.Init(common.NewBackend(nil).Logger("test", true))
	metadataCommon.Logger.Init(common.NewBackend(nil).Logger("test", true))
	Logger.log.Info("Init logger")
	return
}()

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

func initLog() {
	initLogRotator("./bridgeagg.log")
	bridgeAggLogger := common.NewBackend(logWriter{}).Logger("BRIDGEAGG log ", false)
	Logger.Init(bridgeAggLogger)
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

func TestCalculateAmountByDecimal(t *testing.T) {
	config.AbortParam()
	config.Param().BridgeAggParam.BaseDecimal = 9
	type args struct {
		amount             *big.Int
		decimal            uint8
		isToUnifiedDecimal bool
	}
	tests := []struct {
		name    string
		args    args
		want    *big.Int
		wantErr bool
	}{
		{
			name: "Decimal < base decimal - Add",
			args: args{
				amount:             big.NewInt(100000),
				decimal:            6,
				isToUnifiedDecimal: true,
			},
			want:    big.NewInt(100000000),
			wantErr: false,
		},
		{
			name: "Decimal < base decimal - Sub",
			args: args{
				amount:             big.NewInt(100000000),
				decimal:            6,
				isToUnifiedDecimal: false,
			},
			want:    big.NewInt(100000),
			wantErr: false,
		},
		{
			name: "Convert",
			args: args{
				amount:             big.NewInt(100),
				decimal:            config.Param().BridgeAggParam.BaseDecimal,
				isToUnifiedDecimal: true,
			},
			want:    big.NewInt(100),
			wantErr: false,
		},
		{
			name: "Shield",
			args: args{
				amount:             big.NewInt(1234567890000000),
				decimal:            18,
				isToUnifiedDecimal: true,
			},
			want:    big.NewInt(1234567),
			wantErr: false,
		},
		{
			name: "Shield - 2",
			args: args{
				amount:             big.NewInt(50000000000000000),
				decimal:            18,
				isToUnifiedDecimal: true,
			},
			want:    big.NewInt(50000000),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertAmountByDecimal(tt.args.amount, tt.args.decimal, tt.args.isToUnifiedDecimal)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateAmountByDecimal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalculateAmountByDecimal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func readTestCases(fileName string) ([]byte, error) {
	raw, err := ioutil.ReadFile("testdata/" + fileName)
	if err != nil {
		panic(err)
	}
	return raw, nil
}

func CheckInterfacesIsEqual(expected interface{}, actual interface{}) error {
	expectedData, err := json.Marshal(expected)
	if err != nil {
		return err
	}
	actualData, err := json.Marshal(actual)
	if err != nil {
		return err
	}
	if !bytes.Equal(expectedData, actualData) {
		return fmt.Errorf("expected %s but get %s", string(expectedData), string(actualData))
	}
	return nil
}

func TestA(t *testing.T) {
	externalTokenIDs := []string{
		"0x07de306FF27a2B630B1141956844eB1552B956B5",
		"0x337610d27c682E347C9cD60BD4b3b107C9d34dDd",
		"0x75b0622cec14130172eae9cf166b92e5c112faff",
		"0x64544969ed7EBf5f083679233325356EbE738930",
		"0x0000000000000000000000000000000000000000",
		"0xd66c6b4f0be8ce5b39d52e0fd1344c389929b378",
		"0xa36085F69e2889c224210F603D836748e7dC0088",
		"0x84b9B910527Ad5C03A9Ca831909E21e236EA7b06",
		"0x326c977e6efc84e512bb9c30f76e30c160ed06fb",
		"0xfaFedb041c0DD4fA2Dc0d87a6B0979Ee6FA7af5F",
		"0x4f96fe3b7a6cf9725f59d353f723c1bdb64ca6aa",
		"0x001b3b4d0f3714ca98ba10f6042daebf0b1b7b6f",
		"0x9440c3bB6Adb5F0D5b8A460d8a8c010690daC2E8",
		"0x0000000000000000000000000000000000000000",
		"0x0000000000000000000000000000000000000000",
		"0x0000000000000000000000000000000000000000",
		"0xed24fc36d5ee211ea25a80239fb8c4cfd80f12ee",
		"0x3813e82e6f7098b9583FC0F33a962D02018B6803",
	}

	prefixs := []string{
		"",
		common.BSCPrefix,
		"",
		common.BSCPrefix,
		"",
		common.BSCPrefix,
		"",
		common.BSCPrefix,
		common.PLGPrefix,
		common.FTMPrefix,
		"",
		common.PLGPrefix,
		common.FTMPrefix,
		common.PLGPrefix,
		common.FTMPrefix,
		common.BSCPrefix,
		common.BSCPrefix,
		common.PLGPrefix,
	}

	for i := 0; i < len(prefixs); i++ {
		tokenAddr := rCommon.HexToAddress(externalTokenIDs[i])
		res := append([]byte(prefixs[i]), tokenAddr.Bytes()...)

		bytes := base64.StdEncoding.EncodeToString(res)
		fmt.Printf("bytes: %v\n", bytes)
	}

}
