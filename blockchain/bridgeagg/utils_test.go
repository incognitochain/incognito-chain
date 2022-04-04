package bridgeagg

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/jrick/logrotate/rotator"
)

func TestCalculateActualAmount(t *testing.T) {
	type args struct {
		x        uint64
		y        uint64
		deltaX   uint64
		operator byte
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateActualAmount(tt.args.x, tt.args.y, tt.args.deltaX, tt.args.operator)
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
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "burntAmount == 0",
			args: args{
				burntAmount: 0,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "y == 0, burntAmount > x",
			args: args{
				burntAmount: 10,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EstimateActualAmountByBurntAmount(tt.args.x, tt.args.y, tt.args.burntAmount)
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
		amount      big.Int
		decimal     uint
		operator    byte
		prefix      string
		networkType uint
		token       []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *big.Int
		wantErr bool
	}{
		{
			name: "Convert",
			args: args{
				amount:      *big.NewInt(100),
				decimal:     config.Param().BridgeAggParam.BaseDecimal,
				operator:    AddOperator,
				prefix:      "",
				networkType: 0,
				token:       []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
			want:    big.NewInt(100),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateAmountByDecimal(tt.args.amount, tt.args.decimal, tt.args.operator, tt.args.prefix, tt.args.networkType, tt.args.token)
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
