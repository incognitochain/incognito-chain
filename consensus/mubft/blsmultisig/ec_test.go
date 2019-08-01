package blsmultisig

import (
	"crypto/rand"
	"errors"
	"reflect"
	"testing"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

func TestDecmprG1(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *bn256.G1
		wantErr bool
	}{
		// {
		// 	name: "Decompre"
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecmprG1(tt.args.bytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecmprG1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecmprG1() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCmprG1(t *testing.T) {
	type args struct {
		pn *bn256.G1
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// {
		// 	name: "Compress generator element of G1",
		// 	args: args{
		// 		pn: new(bn256.G1).ScalarBaseMult(big.NewInt(1)),
		// 	},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CmprG1(tt.args.pn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CmprG1() = %v, want %v", got, tt.want)
			}
		})
	}
}

func cmptPnG1(oddPoint byte, loop int) (*bn256.G1, error) {
	tests := make([]*bn256.G1, loop)
	var err error
	for i := 0; i < loop; i++ {
		_, tests[i], err = bn256.RandomG1(rand.Reader)
		if err != nil {
			return tests[i], err
		}
		for ; tests[i].Marshal()[63]&1 != oddPoint; _, tests[i], _ = bn256.RandomG1(rand.Reader) {
		}
		cmprBytesArr := CmprG1(tests[i])
		pnDeCmpr, err := DecmprG1(cmprBytesArr)
		if err != nil {
			return tests[i], err
		}
		if !reflect.DeepEqual(pnDeCmpr.Marshal(), tests[i].Marshal()) {
			return tests[i], errors.New("Not equal")
		}
	}
	return nil, nil
}

func Test_cmptPnG1(t *testing.T) {
	type args struct {
		oddPoint byte
		loop     int
	}
	tests := []struct {
		name    string
		args    args
		want    *bn256.G1
		wantErr bool
	}{
		{
			name: "Test with 10000 odd point",
			args: args{
				oddPoint: 1,
				loop:     10000,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Test with 10000 even point",
			args: args{
				oddPoint: 0,
				loop:     10000,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cmptPnG1(tt.args.oddPoint, tt.args.loop)
			if (err != nil) != tt.wantErr {
				t.Errorf("cmptPnG1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cmptPnG1() = %v, want %v", got, tt.want)
			}
		})
	}
}
