package blsmultisig

import (
	"math/rand"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

var listPKsBytes []PublicKey
var listSKsBytes []SecretKey

func genSubset4Test(k, n int) []int {
	res := make([]int, k)
	if k == n {
		for i := 0; i < k; i++ {
			res[i] = i
		}
		return res
	}
	chk := make([]bool, n)
	res[k-1] = n - 1
	chk[n-1] = true
	for i := k - 2; i >= 0; i-- {
		res[i] = rand.Intn(n)
		for {
			if chk[res[i]] {
				res[i] = rand.Intn(n)
			} else {
				chk[res[i]] = true
				break
			}
		}
	}
	return res
}

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
	sigs := make([][]byte, len(subset))
	var err error
	for i := 0; i < len(subset); i++ {
		sigs[i], err = Sign(data, B2I(listSKsBytes[subset[i]]), subset[i])
		if err != nil {
			return [][]byte{[]byte{0}}, err
		}
	}
	return sigs, nil
}

func combine(sigs [][]byte) ([]byte, error) {
	return Combine(sigs)
}

func verify(data, cSig []byte, subset []int) (bool, error) {
	return Verify(cSig, data, subset)
}

// return time sign, combine, verify
func fullBLSSignFlow(wantErr, rewriteKey bool, committeeSign []int) (float64, float64, float64, bool, error) {
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
	// fmt.Println("Sigs: ", sigs)
	start = time.Now()
	cSig, err := combine(sigs)
	t2 := time.Now().Sub(start)
	if err != nil {
		return 0, 0, 0, true, err
	}
	// fmt.Println("Combine sigs", cSig)
	start = time.Now()
	result, err := verify(data, cSig, committeeSign)
	t3 := time.Now().Sub(start)
	if err != nil {
		return 0, 0, 0, true, err
	}
	return t1.Seconds(), t2.Seconds(), t3.Seconds(), result, nil
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
		want    float64
		want1   float64
		want2   float64
		want3   bool
		wantErr bool
	}{
		{
			name: "Test single committee sign",
			args: args{
				wantErr:       false,
				rewriteKey:    true,
				committeeSign: []int{0},
			},
			want:    0.05,
			want1:   0.005,
			want2:   0.05,
			want3:   true,
			wantErr: false,
		},
		{
			name: "Test 100 of 100 committee sign",
			args: args{
				wantErr:       false,
				rewriteKey:    true,
				committeeSign: genSubset4Test(100, 100),
			},
			want:    0.2,
			want1:   0.01,
			want2:   0.2,
			want3:   true,
			wantErr: false,
		},
		{
			name: "Test 50 of 100 committee sign",
			args: args{
				wantErr:       false,
				rewriteKey:    true,
				committeeSign: genSubset4Test(50, 100),
			},
			want:    0.08,
			want1:   0.005,
			want2:   0.08,
			want3:   true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, got3, err := fullBLSSignFlow(tt.args.wantErr, tt.args.rewriteKey, tt.args.committeeSign)
			if (err != nil) != tt.wantErr {
				t.Errorf("fullBLSSignFlow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got > tt.want {
				t.Errorf("fullBLSSignFlow() got = %v, want %v", got, tt.want)
			}
			if got1 > tt.want1 {
				t.Errorf("fullBLSSignFlow() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 > tt.want2 {
				t.Errorf("fullBLSSignFlow() got2 = %v, want %v", got2, tt.want2)
			}
			if got3 != tt.want3 {
				t.Errorf("fullBLSSignFlow() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}

func Test_genSubset4Test(t *testing.T) {
	type args struct {
		k int
		n int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "100 of 100",
			args: args{
				k: 100,
				n: 100,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genSubset4Test(tt.args.k, tt.args.n); len(got) != tt.args.k {
				t.Errorf("len(genSubset4Test(%v, %v)) = %v, want %v", tt.args.k, tt.args.n, len(got), tt.args.k)
			}
		})
	}
}
