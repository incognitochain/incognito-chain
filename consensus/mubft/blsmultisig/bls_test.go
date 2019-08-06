package blsmultisig

import (
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

var listPKsBytes []PublicKey
var listSKsBytes []SecretKey

func genKey(seed []byte, size int) error {
	internalseed := seed
	listPKsBytes = make([]PublicKey, size)
	listSKsBytes = make([]SecretKey, size)
	for i := 0; i < size; i++ {
		sk, pk := KeyGen(internalseed)
		listSKsBytes[i] = SKBytes(sk)
		listPKsBytes[i] = PKBytes(pk)
		internalseed = common.HashB(append(seed, append(listSKsBytes[i], listPKsBytes[i]...)...))
	}
	return CacheCommonPKs(listPKsBytes)
}

func sign(data []byte, subset []int) ([][]byte, error) {
	return [][]byte{[]byte{0}}, nil
}

func combine(sigs [][]byte) ([]byte, error) {
	return []byte{0}, nil
}

func verify(data []byte, sigs []byte, subset []int) bool {
	return true
}

// return time sign, combine, verify
func fullBLSSignFlow(wantErr, rewriteKey bool, committeeSign []int) (int64, int64, int64, bool, error) {
	if rewriteKey {
		max := 0
		for i := 1; i < len(committeeSign); i++ {
			if committeeSign[i] > committeeSign[max] {
				max = i
			}
		}
		committeeSize := committeeSign[max] + 1
		err := genKey([]byte{0, 1, 2, 3, 4}, committeeSize)
		if err != nil {
			return 0, 0, 0, true, err
		}
	}
	data := []byte{0, 1, 2, 3, 4}
	start := time.Now()
	sigs, err := sign(data, committeeSign)
	t1 := time.Now().Sub(start)
	if err != nil {
		return 0, 0, 0, true, err
	}
	start = time.Now()
	cSig, err := combine(sigs)
	t2 := time.Now().Sub(start)
	if err != nil {
		return 0, 0, 0, true, err
	}
	start = time.Now()
	result := verify(data, cSig, committeeSign)
	t3 := time.Now().Sub(start)
	return t1.Nanoseconds(), t2.Nanoseconds(), t3.Nanoseconds(), result, nil
}

func Test_fullBLSSignFlow(t *testing.T) {
	type args struct {
		wantErr       bool
		rewriteKey    bool
		committeeSign []int
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		want1   int64
		want2   int64
		want3   bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, got3, err := fullBLSSignFlow(tt.args.wantErr, tt.args.rewriteKey, tt.args.committeeSign)
			if (err != nil) != tt.wantErr {
				t.Errorf("fullBLSSignFlow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("fullBLSSignFlow() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("fullBLSSignFlow() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("fullBLSSignFlow() got2 = %v, want %v", got2, tt.want2)
			}
			if got3 != tt.want3 {
				t.Errorf("fullBLSSignFlow() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}
