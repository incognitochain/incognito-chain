package blsmultisig

import (
	"crypto/rand"
	"testing"
	"time"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

func testAvgTimeP2I(loop int64) int64 {
	sum := int64(0)
	for i := int64(0); i < loop; i++ {
		_, randPoint, _ := bn256.RandomG1(rand.Reader)
		start := time.Now()
		P2I(randPoint)
		sum += -(start.Sub(time.Now())).Nanoseconds()
	}
	return sum / loop
}

func testAvgTimeI2P(loop int64) int64 {
	sum := int64(0)
	for i := int64(0); i < loop; i++ {
		_, randPoint, _ := bn256.RandomG1(rand.Reader)
		start := time.Now()
		P2I(randPoint)
		sum += -(start.Sub(time.Now())).Nanoseconds()
	}
	return sum / loop
}

func Test_testAvgTimeP2I(t *testing.T) {
	type args struct {
		loop int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "Test 10000 loop and UpperBound for function execution time is 0.0001s",
			args: args{
				loop: 10000,
			},
			want: 100000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testAvgTimeP2I(tt.args.loop); got > tt.want {
				t.Errorf("Execution time of testAvgTimeP2I() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_testAvgTimeI2P(t *testing.T) {
	type args struct {
		loop int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "Test 10000 loop and UpperBound for function execution time is 0.0001s",
			args: args{
				loop: 10000,
			},
			want: 100000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testAvgTimeI2P(tt.args.loop); got > tt.want {
				t.Errorf("Execution time of testAvgTimeI2P() = %v, want %v", got, tt.want)
			}
		})
	}
}
