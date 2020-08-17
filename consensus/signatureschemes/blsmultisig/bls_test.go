package blsmultisig

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/stretchr/testify/assert"
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
	return nil
	// return CacheCommonPKs(listPKsBytes)
}

func sign(data []byte, subset []int) ([][]byte, error) {
	sigs := make([][]byte, len(subset))
	var err error
	for i := 0; i < len(subset); i++ {
		sigs[i], err = Sign(data, listSKsBytes[subset[i]], subset[i], listPKsBytes)
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
	return Verify(cSig, data, subset, listPKsBytes)
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
	t1 := time.Since(start)
	if err != nil {
		return 0, 0, 0, true, err
	}
	// fmt.Println("Sigs: ", sigs)
	start = time.Now()
	cSig, err := combine(sigs)
	// fmt.Println("sigs:", sigs)
	cSig2, err := combine(sigs)
	// fmt.Println("sigs:", sigs)
	// fmt.Println("CSig:", cSig, cSig2)
	t2 := time.Since(start)
	if err != nil {
		return 0, 0, 0, true, err
	}
	// fmt.Println("Combine sigs", cSig)
	start = time.Now()
	result, err := verify(data, cSig, committeeSign)
	result2, err := verify(data, cSig2, committeeSign)
	fmt.Println(result, result2)
	t3 := time.Since(start)
	if err != nil {
		return 0, 0, 0, true, err
	}
	return t1.Seconds(), t2.Seconds(), t3.Seconds(), result, nil
}

func Test_Verify(t *testing.T) {
	committeeSign := genSubset4Test(200, 200)
	max := 0
	for i := 1; i < len(committeeSign); i++ {
		if committeeSign[i] > committeeSign[max] {
			max = i
		}
	}
	committeeSize := committeeSign[max] + 1
	err := genKey([]byte{0, 1, 2, 3, 4}, committeeSize)
	if err != nil {
		t.Error(err)
		return
	}
	data := []byte{0, 1, 2, 3, 4}
	//start := time.Now()
	sigs, err := sign(data, committeeSign)
	//t2 := time.Since(start)
	//fmt.Println(t2.Seconds())
	if err != nil {
		t.Error(err)
		return
	}
	cSig, err := combine(sigs)
	start := time.Now()
	res, _ := verify(data, cSig, committeeSign)
	t3 := time.Since(start)
	fmt.Println(res, t3.Seconds()*1000)
	assert.Equal(t, true, res)

	start = time.Now()
	res, _ = verify(data, cSig, committeeSign)
	t3 = time.Since(start)
	fmt.Println(res, t3.Seconds()*1000)
	assert.Equal(t, true, res)
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
			want:    0.15,
			want1:   0.005,
			want2:   0.15,
			want3:   true,
			wantErr: false,
		},
		// {
		// 	name: "Test 20 of 20 committee sign",
		// 	args: args{
		// 		wantErr:       false,
		// 		rewriteKey:    true,
		// 		committeeSign: genSubset4Test(20, 20),
		// 	},
		// 	want:    2,
		// 	want1:   0.01,
		// 	want2:   2,
		// 	want3:   true,
		// 	wantErr: false,
		// },
		// {
		// 	name: "Test 10 of 20 committee sign",
		// 	args: args{
		// 		wantErr:       false,
		// 		rewriteKey:    true,
		// 		committeeSign: genSubset4Test(10, 20),
		// 	},
		// 	want:    1,
		// 	want1:   0.01,
		// 	want2:   1,
		// 	want3:   true,
		// 	wantErr: false,
		// },
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

func Test_SpecialCase(t *testing.T) {
	blkHash := []byte{204, 125, 67, 95, 25, 125, 133, 245, 212, 245, 165, 122, 161, 228, 187, 184, 187, 103, 135, 146, 135, 123, 246, 6, 86, 184, 41, 191, 177, 27, 182, 227}
	committee := make([]PublicKey, 4)
	var err error

	committee[0], _, err = base58.Base58Check{}.Decode("1XUv5rv257yzEzw4BMxevk7ZAhuvDe521CgkjDNPLCMC37VL2517oLCK6mUww9rxhrAWkRU5Dhe65w3tefZNw36W5AvMGd96vXF2UnadAjoHcUSMy1xvq77K9PouS9K8ivwmatgKVqnEtug346WvpnqzbWkZKESBA4zvE2aKP3kLh8KbpMpfQ")
	if err != nil {
		fmt.Println("err 0", err)
	}
	committee[1], _, err = base58.Base58Check{}.Decode("18q2uA89SE41hGr3L5rAwbcWWh6PDxpxKd4K32BSND1wEY2Rh1rLockytT76xdroWRHR7HYF7mcwaTtjNQthLZyXk8CvR18c6VtmzRfELUvHyC4ihWxMgHMB68fcCiD7Xpxnys8vD5SPywLpMmNKuqkXGJzagnVwB2nWfzyJs3D6Fnjjto7Nw")
	if err != nil {
		fmt.Println("err 1", err)
	}
	committee[2], _, err = base58.Base58Check{}.Decode("1JGNfsVB79nY7sjTLqucx2kgfySMYUccgGxnLRFzo3FRqeXYMr5VhH5MhfZNikHyQgGs7FX7wfzAzjY7G8Fz9EvubY7YgDestWvuNRekCZHwnYJg62SZQboNEpf5i3LQV5ML1utgVGDLSZiFQSN7RvubDa9KCyHeY6cYDVn7Wd72FTvBcPTuc")
	if err != nil {
		fmt.Println("err 2", err)
	}
	committee[3], _, err = base58.Base58Check{}.Decode("1TjN1wbra9eG4cAmjB6xNcvimPF6T3ikjimnTzy5FyJP9chDsW6YMCcbKFQ1aTz2T4kRuLVCbmMXzwduDPYRjKNuddocfUKCY2QjW25kbaS1FaNgvCPxhq5q91DnkVi5L9mKuye9FAhZymwM1cpknsv9xosVa46EyXrEMfsBaC7TYmiWQUdAE")
	if err != nil {
		fmt.Println("err 3", err)
	}
	sig, err := base64.StdEncoding.DecodeString("p25ufUnlEnXaBWaXdVr32zvbRoSrQk2PmUSHaoND+HA=")
	fmt.Println(len(sig))

	ssig, err := base64.StdEncoding.DecodeString("Afo8olSvgRxoIIlg4V/MT+DoyDLuKViCff4XqHu2kVM=")
	fmt.Println(len(ssig))
	for i := 0; i < 3; i++ {
		fmt.Println(Verify(sig, blkHash, []int{0}, []PublicKey{committee[i]}))
	}
	// if err != nil {
	// 	fmt.Println("err 4", err)
	// }
	fmt.Println(Verify(sig, blkHash, []int{1, 2, 3}, committee))
}

/*
return:
	@param1: blkHash
	@param2: wantedSig
	@param2: listPublicKey
	@param3: listPrivateKey
	@param4: signIdx
*/
func BuildTestnetEnv(i int) ([]byte, []byte, []PublicKey, [][]byte, []int) {
	switch i {
	case 0:
		pks := []PublicKey{}
		sks := [][]byte{}
		blkHash := "431907cd7ddf72e44b89474fcffc75f89ae3105abce972333e97a4a6e235db53"
		sig := []byte{27, 241, 204, 219, 4, 32, 33, 173, 251, 220, 80, 176, 177, 97, 231, 72, 144, 32, 42, 161, 72, 53, 200, 190, 53, 29, 221, 148, 52, 75, 101, 143}
		sigIdx := []int{1, 2, 3, 5, 6}
		listPrivateSeed := []string{
			"12477VBjQHsNH6wVhXhX6cRzL5yNcGw4HLHjvUjf4wibXVmNPsH",
			"1rKUZeieTgQqgcdEpYGc4krQFNu4Kmf4P5iXMuTiqKXKHEEWmh",
			"12Ms69atUaTdm2GfJFf49Q5tuFowT8xWQSACk7nEPzyT47ptaRz",
			"12m4dq4GAtxWZU7R9pm6AifxLXCQKwgiFTWAHo3JtK4SDcGKjK3",
			"1S98qbp22q1nRmkXaq3F7xF76H1fonm82wWb7V9b3eT21yeFQ4",
			"1fGk4mcjHza1T9s9PihPhoPquA5RJVGGzSVy9aPWfw544SPgKQ",
			"19Vcd7kAWyN7N8VpAeXfdbHmAeYiXaprPyxi389jVnx89eAz7h",
		}
		for _, s := range listPrivateSeed {
			seed, _, _ := base58.Base58Check{}.Decode(s)
			skRaw, pkRaw := KeyGen(seed)
			sks = append(sks, SKBytes(skRaw))
			pks = append(pks, PKBytes(pkRaw))
		}
		rawHash, _ := common.Hash{}.NewHashFromStr(blkHash)
		return rawHash.Bytes(), sig, pks, sks, sigIdx
	}
	return []byte{}, []byte{}, []PublicKey{}, [][]byte{}, []int{}
}

func TestTestnet(t *testing.T) {
	type args struct {
		data      []byte
		skBytes   [][]byte
		signIdx   []int
		committee []PublicKey
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "Testnet 01",
			wantErr: false,
		},
	}
	for j, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allSig := [][]byte{}
			tt.args.data, tt.want, tt.args.committee, tt.args.skBytes, tt.args.signIdx = BuildTestnetEnv(j)
			for _, i := range tt.args.signIdx {
				got, err := Sign(tt.args.data, tt.args.skBytes[i], i, tt.args.committee)
				if (err != nil) != tt.wantErr {
					t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				allSig = append(allSig, got)
			}
			got, err := combine(allSig)
			fmt.Println(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sign() = %v, want %v", got, tt.want)
			}
			valid, err := Verify(got, tt.args.data, tt.args.signIdx, tt.args.committee)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !valid {
				t.Errorf("Sign() = %v, want %v", err, true)
				return
			}
		})
	}
}
