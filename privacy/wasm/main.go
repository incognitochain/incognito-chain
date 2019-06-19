package main

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"syscall/js"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"time"
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

func aggregatedRangeProve(this js.Value, args []js.Value) (interface{}, error) {
	bytes := []byte(args[0].String())
	println("Bytes: %v\n", bytes)

	wit := new(zkp.AggregatedRangeWitness)

	println("Wit: ", wit)

	json.Unmarshal(bytes, &wit)

	println("wit after unmarshal : %v\n", wit)

	start := time.Now()
	proof, err := wit.Prove()
	if err != nil {
		println("Err: %v\n", err)
	}
	end := time.Since(start)
	println("Aggregated range proving time: %v\n", end)

	//tln("Proof json marshal: %v\n", proofMarshal)proofMarshal, _ :=  json.Marshal(proof)
	//	//
	//	//prin

	proofBase64 := base64.StdEncoding.EncodeToString(proof.Bytes())

	println("proofBase64: %v\n", proofBase64)

	return proofBase64, nil
}

func main() {
	c := make(chan struct{}, 0)
	//println("Hello WASM")
	RegisterCallback("add", add)
	RegisterCallback("sayHello", sayHello)
	RegisterCallback("aggregatedRangeProve", aggregatedRangeProve)
	<-c
}
