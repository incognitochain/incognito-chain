package incognitokey

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
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
				keyString: "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",
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
				fmt.Println(pubKey.GetMiningKey(common.BLS_CONSENSUS))
				fmt.Println(pubKey.GetMiningKey(common.BRI_CONSENSUS))
			}

		})
	}
}
