package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"math/big"
	"time"

	//"github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	//"math/big"
	"strconv"
	"syscall/js"
)

func add(this js.Value, i []js.Value) (interface{}, error) {
	ret := 0

	for _, item := range i {
		val, _ := strconv.Atoi(item.String())
		ret += val
	}

	return ret, nil
}

func sayHello(this js.Value, i []js.Value) (interface{}, error) {
	println("Hello %s", i[0].String())
	return i[0].String(), nil
}

func randomScalar (this js.Value, i []js.Value) (interface{}, error) {
	res :=  privacy.RandBytes(1)
	return res, nil
}

func aggregatedRangeProve(this js.Value, args []js.Value) (interface{}, error) {
	println("args:", args[0].String())
	bytes := []byte(args[0].String())
	println("Bytes:", bytes)
	temp := make(map[string] []string)

	err := json.Unmarshal(bytes, &temp)
	if err != nil {
		println(err)
		return nil, nil
	}
	println("asfasf", temp["values"][0])

	if len(temp["values"]) != len(temp["rands"]) {
		println("Wrong args")
	}

	values := make([]*big.Int, len(temp["values"]))
	rands := make([]*big.Int, len(temp["values"]))

	for i:=0; i < len(temp["values"]); i++{
		values[i], _ = new(big.Int).SetString(temp["values"][i], 10)
		rands[i], _ = new(big.Int).SetString(temp["rands"][i], 10)
	}

	wit := new(zkp.AggregatedRangeWitness)
	wit.Set(values, rands)

	start := time.Now()
	proof, err := wit.Prove()
	if err != nil {
		println("Err: %v\n", err)
	}
	end := time.Since(start)
	println("Aggregated range proving time: %v\n", end)

	proofBytes := proof.Bytes()
	println("Proof bytes: ", proofBytes)

	res := proof.Verify()
	println("Res Verify: ", res)


	proofBase64 := base64.StdEncoding.EncodeToString(proofBytes)
	println("proofBase64: %v\n", proofBase64)

	return proofBase64, nil
	//return nil, nil
}

func main() {
	c := make(chan struct{}, 0)
	println("Hello WASM")
	RegisterCallback("add", add)
	RegisterCallback("sayHello", sayHello)
	RegisterCallback("randomScalar", randomScalar)
	RegisterCallback("aggregatedRangeProve", aggregatedRangeProve)
	<-c
}
