package blsmultisig

import (
	"crypto/rand"
	"reflect"
	"testing"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

// func TestDecompress(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		bn   *big.Int
// 	}{}
// 	j := 0
// 	for i := 0; i < 10000; i++ {
// 		bn := privacy.RandScalar()
// 		x := bn.Mod(bn, bn256.Order)
// 		if I2Bytes(x, 32)[0] >= 0x8F {
// 			// if (x.Bit(255) == 1) && (x.Bit(254) == 1) {
// 			fmt.Println(I2Bytes(x, 32))
// 			j++
// 		}
// 	}
// 	Decompress()
// 	fmt.Println(bn256.Order.String())
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			Decompress()
// 		})
// 	}
// }

func TestCmprG1(t *testing.T) {
	type args struct {
		pn *bn256.G1
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	_, p, _ := bn256.RandomG1(rand.Reader)
	CmprG1(p)
	CmprG1(p.Neg(p))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CmprG1(tt.args.pn); (!reflect.DeepEqual(got, tt.want)) && (len(got) != CCmprPnSz) {
				t.Errorf("CmprG1() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecmprG1(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name string
		args args
		want *bn256.G1
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecmprG1(tt.args.bytes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecmprG1() = %v, want %v", got, tt.want)
			}
		})
	}
}

func cmptPnG1(oddPoint byte, loop int) (bool, *bn256.G1) {
	tests := make([]*bn256.G1, loop)
	for i := 0; i < loop; i++ {
		_, tests[i], _ = bn256.RandomG1(rand.Reader)
		for ; tests[i].Marshal()[63]&1 != oddPoint; _, tests[i], _ = bn256.RandomG1(rand.Reader) {
		}
		cmprBytesArr := CmprG1(tests[i])
		pnDeCmpr := DecmprG1(cmprBytesArr)
		if !reflect.DeepEqual(pnDeCmpr.Marshal(), tests[i].Marshal()) {
			return false, tests[i]
		}
	}
	return true, nil
}

//Test compute point in G1 group
func Test_cmptPnG1(t *testing.T) {
	type args struct {
		oddPoint byte
		loop     int
	}
	type res struct {
		success bool
		cause   *bn256.G1
	}
	tests := []struct {
		name string
		args args
		want res
	}{
		{
			name: "Test with 1000 odd point",
			args: args{
				oddPoint: 1,
				loop:     1000,
			},
			want: res{
				success: true,
				cause:   nil,
			},
		},
		{
			name: "Test with 1000 even point",
			args: args{
				oddPoint: 0,
				loop:     1000,
			},
			want: res{
				success: true,
				cause:   nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if success, cause := cmptPnG1(tt.args.oddPoint, tt.args.loop); !success {
				t.Errorf("cmptPnG1(%v, %v) failed because %v", tt.args.oddPoint, tt.args.loop, cause.Marshal())
			}
		})
	}
}
