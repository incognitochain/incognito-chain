package bridgeagg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"testing"

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

func TestCalculateShieldActualAmount(t *testing.T) {
	type args struct {
		x        uint64
		y        uint64
		deltaX   uint64
		operator byte
		isPaused bool
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "Cannot recognize operator",
			args: args{
				operator: 3,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "y == 0, first shield",
			args: args{
				deltaX:   100,
				operator: AddOperator,
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "y == 0, second shield",
			args: args{
				x:        100,
				deltaX:   300,
				operator: AddOperator,
			},
			want:    300,
			wantErr: false,
		},
		{
			name: "y != 0, first shield",
			args: args{
				y:        100,
				deltaX:   100,
				operator: AddOperator,
			},
			want:    199,
			wantErr: false,
		},
		{
			name: "y != 0, second shield",
			args: args{
				y:        100,
				deltaX:   100,
				x:        1000,
				operator: AddOperator,
			},
			want:    109,
			wantErr: false,
		},
		{
			name: "isPaused shield",
			args: args{
				y:        10,
				deltaX:   100,
				x:        1000,
				operator: AddOperator,
				isPaused: true,
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "shield with large amount",
			args: args{
				y:        1e10,
				deltaX:   1234567,
				x:        1234567,
				operator: AddOperator,
				isPaused: false,
			},
			want:    5001234567,
			wantErr: false,
		},
		{
			name: "shield with large amount - 2",
			args: args{
				y:        5e9,
				deltaX:   609,
				x:        2469134,
				operator: SubOperator,
				isPaused: false,
			},
			want:    1233530,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateShieldActualAmount(tt.args.x, tt.args.y, tt.args.deltaX, tt.args.isPaused)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateActualAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CalculateActualAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstimateActualAmountByBurntAmount(t *testing.T) {
	type args struct {
		x           uint64
		y           uint64
		burntAmount uint64
		isPaused    bool
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "x == 0",
			args: args{
				x: 0,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "x == 1",
			args: args{
				x: 1,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "burntAmount == 0",
			args: args{
				x:           100,
				burntAmount: 0,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "y == 0, burntAmount > x",
			args: args{
				x:           10,
				burntAmount: 100,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "y == 0, burntAmount < x",
			args: args{
				x:           150,
				burntAmount: 100,
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "y == 0, burntAmount = x",
			args: args{
				x:           100,
				burntAmount: 100,
			},
			want:    99,
			wantErr: false,
		},
		{
			name: "y != 0, burntAmount <= x",
			args: args{
				x:           1000,
				y:           100,
				burntAmount: 111,
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "y != 0, burntAmount <= x - example 2",
			args: args{
				x:           50000000,
				y:           1000000,
				burntAmount: 300,
			},
			want:    294,
			wantErr: false,
		},
		{
			name: "y != 0, burntAmount > x - valid",
			args: args{
				x:           1000,
				y:           100,
				burntAmount: 1050,
			},
			want:    750,
			wantErr: false,
		},
		{
			name: "unshield after shield",
			args: args{
				x:           1100,
				y:           91,
				burntAmount: 109,
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "isPaused",
			args: args{
				x:           1000,
				y:           10,
				burntAmount: 100,
				isPaused:    true,
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "unshield with large amount",
			args: args{
				x:           2469134,
				y:           5e9,
				burntAmount: 1234567,
			},
			want:    609,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EstimateActualAmountByBurntAmount(tt.args.x, tt.args.y, tt.args.burntAmount, tt.args.isPaused)
			if (err != nil) != tt.wantErr {
				t.Errorf("EstimateActualAmountByBurntAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EstimateActualAmountByBurntAmount() = %v, want %v", got, tt.want)
			}
		})
	}
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
		decimal            uint
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

func Test_updateRewardReserve(t *testing.T) {
	type args struct {
		lastUpdatedRewardReserve uint64
		currentRewardReserve     uint64
		newRewardReserve         uint64
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		want1   uint64
		wantErr bool
	}{
		{
			name: "First time modify",
			args: args{
				lastUpdatedRewardReserve: 0,
				currentRewardReserve:     0,
				newRewardReserve:         100,
			},
			want:    100,
			want1:   100,
			wantErr: false,
		},
		{
			name: "Second time modify - not yet update reward",
			args: args{
				lastUpdatedRewardReserve: 100,
				currentRewardReserve:     100,
				newRewardReserve:         200,
			},
			want:    200,
			want1:   200,
			wantErr: false,
		},
		{
			name: "Second time modify - has updated reward - deltaY > 0",
			args: args{
				lastUpdatedRewardReserve: 100,
				currentRewardReserve:     90,
				newRewardReserve:         200,
			},
			want:    200,
			want1:   190,
			wantErr: false,
		},
		{
			name: "Second time modify - has updated reward - deltaY < 0",
			args: args{
				lastUpdatedRewardReserve: 100,
				currentRewardReserve:     110,
				newRewardReserve:         200,
			},
			want:    200,
			want1:   210,
			wantErr: false,
		},
		{
			name: "newY < deltaY",
			args: args{
				lastUpdatedRewardReserve: 100,
				currentRewardReserve:     10,
				newRewardReserve:         50,
			},
			want:    0,
			want1:   0,
			wantErr: true,
		},
		{
			name: "deltaY < 0",
			args: args{
				lastUpdatedRewardReserve: 100,
				currentRewardReserve:     200,
				newRewardReserve:         50,
			},
			want:    50,
			want1:   150,
			wantErr: false,
		},
		{
			name: "deltaY < 0",
			args: args{
				lastUpdatedRewardReserve: 100,
				currentRewardReserve:     200,
				newRewardReserve:         50,
			},
			want:    50,
			want1:   150,
			wantErr: false,
		},
		{
			name: "Set reward to 0",
			args: args{
				lastUpdatedRewardReserve: 100,
				currentRewardReserve:     10,
				newRewardReserve:         90,
			},
			want:    0,
			want1:   0,
			wantErr: true,
		},
		{
			name: "Not change value",
			args: args{
				lastUpdatedRewardReserve: 100,
				currentRewardReserve:     10,
				newRewardReserve:         100,
			},
			want:    100,
			want1:   10,
			wantErr: false,
		},
		{
			name: "newReward = deltaY + 1",
			args: args{
				lastUpdatedRewardReserve: 5000000000,
				currentRewardReserve:     88,
				newRewardReserve:         4999999913,
			},
			want:    4999999913,
			want1:   1,
			wantErr: false,
		},
		{
			name: "newReward == 0 && lastUpdatedRewardReserve == 0",
			args: args{
				lastUpdatedRewardReserve: 0,
				currentRewardReserve:     0,
				newRewardReserve:         0,
			},
			want:    0,
			want1:   0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := updateRewardReserve(tt.args.lastUpdatedRewardReserve, tt.args.currentRewardReserve, tt.args.newRewardReserve)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateRewardReserve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("updateRewardReserve() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("updateRewardReserve() got1 = %v, want %v", got1, tt.want1)
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

func TestCalculateDeltaY(t *testing.T) {
	type args struct {
		x        uint64
		y        uint64
		deltaX   uint64
		operator byte
		isPaused bool
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "unshield y != 0",
			args: args{
				y:        100,
				deltaX:   100,
				x:        1000,
				operator: SubOperator,
			},
			want:    11,
			wantErr: false,
		},
		{
			name: "unshield y != 0, isPaused",
			args: args{
				y:        100,
				deltaX:   100,
				x:        1000,
				operator: SubOperator,
				isPaused: true,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "unshield y == 0",
			args: args{
				y:        0,
				deltaX:   100,
				x:        1000,
				operator: SubOperator,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "Try to shield to y = 0",
			args: args{
				y:        100,
				deltaX:   10000000000000000000,
				x:        1000,
				operator: AddOperator,
			},
			want:    99,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateDeltaY(tt.args.x, tt.args.y, tt.args.deltaX, tt.args.operator, tt.args.isPaused)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateDeltaY() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CalculateDeltaY() = %v, want %v", got, tt.want)
			}
		})
	}
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
