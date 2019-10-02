package random

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"testing"
)

func TestConvertHexToUint64(t *testing.T) {
	res, err := hexutil.DecodeUint64("0x1b4")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}

func TestConvertHexToBigInt(t *testing.T) {
	type args struct {
		nonce string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test 8860467",
			args: args{nonce: "0x6436aeb426dc0344"},
		},
		{
			name: "Test 8860466",
			args: args{nonce: "0x9744c0380e23299e"},
		},
		{
			name: "Test 8860465",
			args: args{nonce: "0xe95ff5741b9c972d"},
		},
		{
			name: "Test 8860464",
			args: args{nonce: "0xd29f0188041f147f"},
		},
		{
			name: "Test 8860463",
			args: args{nonce: "0x8e3569392f7bd979"},
		},
		{
			name: "Test 8860462",
			args: args{nonce: "0xf1d36fc80e76945b"},
		},
	}
	result := []int64{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := hexutil.DecodeBig(tt.args.nonce)
			if err != nil {
				t.Fatal(err)
			}
			newBigInt := new(big.Int).SetInt64(MaxInt64)
			resBigInt := new(big.Int)
			resBigInt.Mod(res, newBigInt)
			t.Log(resBigInt.Int64())
			result = append(result, resBigInt.Int64())
		})
	}
	for index1, value1 := range result {
		for index2, value2 := range result {
			if index1 != index2 {
				if value1 == value2 {
					t.Fatal()
				}
			}
		}
	}
}
func TestConvertGetBlockNumberResult(t *testing.T) {
	res := `{
      "id":83,
	  "jsonrpc": "2.0",
	  "result": "0x8426B3"
	}`
	//1207
	getBlockNumberResult := &GetBlockNumberResult{}
	err := json.Unmarshal([]byte(res), getBlockNumberResult)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(getBlockNumberResult.Result)
	res1, err1 := hexutil.DecodeBig(getBlockNumberResult.Result)
	t.Log(res1, err1)
	t.Log(res1.Int64())
	if res1.Int64() != 8660659 {
		t.Fatal("Unexpected value ", res1)
	}
	res2, err := hexutil.DecodeUint64(getBlockNumberResult.Result)
	if res2 != 8660659 {
		t.Fatal("Unexpected value ", res2)
	}
}
