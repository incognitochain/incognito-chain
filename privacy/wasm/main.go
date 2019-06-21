package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"math/big"
	"time"

	"strconv"
	"syscall/js"
)

func add(this js.Value, i []js.Value) interface{} {
	ret := 0

	for _, item := range i {
		val, _ := strconv.Atoi(item.String())
		ret += val
	}

	return ret
}

func sayHello(this js.Value, i []js.Value) interface{} {
	println("Hello %s \n", i[0].String())
	return i[0].String()
}

func randomScalar(this js.Value, i []js.Value) interface{} {
	res := privacy.RandBytes(1)
	return res
}

func aggregatedRangeProve(this js.Value, args []js.Value) interface{} {
	println("args:", args[0].String())
	bytes := []byte(args[0].String())
	println("Bytes:", bytes)
	temp := make(map[string][]string)

	err := json.Unmarshal(bytes, &temp)
	if err != nil {
		println(err)
		return nil
	}
	println("temp values", temp["values"])
	println("temp rands", temp["rands"])

	if len(temp["values"]) != len(temp["rands"]) {
		println("Wrong args")
	}

	values := make([]*big.Int, len(temp["values"]))
	rands := make([]*big.Int, len(temp["values"]))

	for i := 0; i < len(temp["values"]); i++ {
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

	return proofBase64
}

func main() {
	c := make(chan struct{}, 0)
	println("Hello WASM")
	js.Global().Set("add", js.FuncOf(add))
	js.Global().Set("sayHello", js.FuncOf(sayHello))
	js.Global().Set("randomScalar", js.FuncOf(randomScalar))
	js.Global().Set("aggregatedRangeProve", js.FuncOf(aggregatedRangeProve))
	<-c
}
