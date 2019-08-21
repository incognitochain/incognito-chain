package bridgesig

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common"
)

var listPKsBytes [][]byte
var listSKsBytes [][]byte

func genKey(seed []byte, size int) error {
	internalseed := seed
	listPKsBytes = make([][]byte, size)
	listSKsBytes = make([][]byte, size)
	for i := 0; i < size; i++ {
		sk, pk := KeyGen(internalseed)
		listSKsBytes[i] = SKBytes(&sk)
		listPKsBytes[i] = PKBytes(&pk)
		internalseed = common.HashB(append(seed, append(listSKsBytes[i], listPKsBytes[i]...)...))
	}
	return nil
}

func flowECDSASignVerify() {

}

func TestSign(t *testing.T) {
	type args struct {
		keyBytes []byte
		data     []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sign(tt.args.keyBytes, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Sign() = %v, want %v", got, tt.want)
			}
		})
	}
}
