package blsmultisig

import (
	"math/big"
	"reflect"
	"testing"

	"golang.org/x/crypto/bn256"
)

func TestHash4Block(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Hash4Block(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Hash4Block() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestI2P(t *testing.T) {
	type args struct {
		bigInt *big.Int
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
			if got := I2P(tt.args.bigInt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("I2P() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestP2I(t *testing.T) {
	type args struct {
		point *bn256.G1
	}
	tests := []struct {
		name string
		args args
		want *big.Int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := P2I(tt.args.point); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2I() = %v, want %v", got, tt.want)
			}
		})
	}
}
