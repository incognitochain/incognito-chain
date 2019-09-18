package incognitokey

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNewCommitteeKeyFromSeed(t *testing.T) {
	type args struct {
		seed      []byte
		incPubKey []byte
	}
	tests := []struct {
		name    string
		args    args
		want    CommitteePublicKey
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	xxx := []byte{123, 34, 73, 110, 99, 80, 117, 98, 75, 101, 121, 34, 58, 34, 65, 81, 73, 68, 34, 44, 34, 77, 105, 110, 105, 110, 103, 80, 117, 98, 75, 101, 121, 34, 58, 123, 34, 98, 108, 115, 34, 58, 34, 65, 68, 47, 53, 118, 89, 71, 114, 115, 72, 117, 80, 43, 113, 111, 66, 102, 56, 54, 51, 100, 72, 69, 108, 43, 88, 97, 43, 70, 120, 114, 111, 109, 108, 121, 84, 51, 68, 77, 122, 77, 67, 119, 113, 90, 67, 52, 43, 88, 103, 80, 113, 73, 65, 85, 74, 100, 97, 111, 113, 79, 79, 84, 68, 90, 69, 98, 48, 68, 98, 65, 84, 104, 75, 89, 104, 97, 108, 52, 68, 55, 75, 119, 49, 66, 81, 98, 79, 80, 79, 121, 79, 79, 103, 100, 55, 47, 97, 97, 78, 90, 97, 65, 104, 75, 77, 119, 119, 101, 68, 56, 66, 71, 85, 74, 81, 102, 121, 108, 47, 118, 80, 90, 73, 89, 100, 55, 67, 72, 81, 57, 80, 74, 121, 117, 89, 54, 117, 120, 43, 81, 75, 47, 89, 118, 109, 68, 110, 65, 81, 52, 114, 98, 65, 118, 99, 120, 115, 55, 84, 121, 50, 122, 83, 120, 49, 102, 121, 112, 119, 85, 61, 34, 44, 34, 100, 115, 97, 34, 58, 34, 65, 118, 109, 110, 104, 112, 69, 115, 50, 66, 67, 43, 102, 83, 69, 88, 69, 121, 88, 49, 101, 52, 97, 50, 88, 56, 117, 81, 57, 81, 67, 54, 114, 66, 68, 118, 98, 101, 103, 77, 78, 43, 81, 56, 34, 125, 125}
	for i := 0; i < 5000; i++ {
		x, _ := NewCommitteeKeyFromSeed([]byte{1, 2, 3}, []byte{1, 2, 3})
		// fmt.Println(x.By tes())
		xBytes, _ := x.Bytes()
		xNew := new(CommitteePublicKey)
		xNew.FromBytes(xBytes)
		// fmt.Println(xNew.Bytes())
		xNewBytes, _ := xNew.Bytes()
		if !reflect.DeepEqual(xBytes, xNewBytes) {
			panic("vvv")
		}
		if !reflect.DeepEqual(xBytes, xxx) {
			panic("vvv")
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCommitteeKeyFromSeed(tt.args.seed, tt.args.incPubKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCommitteeKeyFromSeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCommitteeKeyFromSeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommitteePublicKey_FromString(t *testing.T) {
	type fields struct {
		IncPubKey    []byte
		MiningPubKey map[string][]byte
	}
	type args struct {
		keyString string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "From string which is created by js",
			fields: fields{
				IncPubKey:    []byte{3, 252, 240, 233, 139, 254, 238, 141, 127, 76, 168, 143, 100, 230, 199, 235, 57, 130, 189, 57, 169, 247, 183, 198, 214, 253, 170, 32, 132, 32, 14, 58, 248},
				MiningPubKey: map[string][]byte{},
			},
			args: args{
				keyString: "121VhftSAygpEJZ6i9jGk9dPK5DE41KmqUhp2EFJXyyHNrQomEh7QW7icEh2zymM5953C2HN2kYYCbepkL37Qny6mgyxGiAxxqEEkCHnqK4FsNBDbGwpHAFon6CZayaLFHW5Uo9QxwxHWNAE2ZGwKEDrbJxNTkh5U9xGbaN5BRNGecTYXvqdmjmtsbNrayN8sz3PtS9KVFcY1CveDJTucsgtG7UtdNvKwDsshgT4tj4KY5L8mHqRd4c4nnPLjL8FKgdcuCwsP7sENJs6zyZEwrhjN8WZmXD4cvfautTt8soTr2fNvCNsXLcoFFBsC4ZSiPeCGub73A41oPMd7fp4yr8og8V3mxfDjvjq9MUeNwkyAAZYjM9fL3hYv7FPbAkgGniAAXtokUAWEKU923Ttbq5UcGujoc5V3r5CyjRPVBrxPnRP",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pubKey := new(CommitteePublicKey)
			if err := pubKey.FromString(tt.args.keyString); (err != nil) != tt.wantErr {
				t.Errorf("CommitteePublicKey.FromString() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				fmt.Println(pubKey.GetNormalKey())
				// blsBytes, _ := pubKey.GetMiningKey(common.BlsConsensus)
				// fmt.Println(base58.Base58Check{}.Encode(blsBytes, common.Base58Version))
				// fmt.Println(pubKey.GetMiningKey(common.BridgeConsensus))
			}

		})
	}
}
