package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	zkp "github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"strconv"
	"syscall/js"
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
	fmt.Printf("Hello %s \n", i[0].String())
	return i[0].String(), nil
}

func aggregatedRangeProve(this js.Value, args []js.Value) (interface{}, error) {
	bytes := []byte(args[0].String())
	fmt.Println("Bytes: %v\n", bytes)

	wit := zkp.AggregatedRangeWitness{}
	_ = wit

	fmt.Println("Wit: ", wit)

	json.Unmarshal(bytes, &wit)

	fmt.Println("wit after unmarshal : %v\n", wit)

	start := time.Now()
	proof, err := wit.Prove()
	if err != nil {
		fmt.Println("Err: %v\n", err)
	}
	end := time.Since(start)
	fmt.Println("Aggregated range proving time: %v\n", end)

	//tln("Proof json marshal: %v\n", proofMarshal)proofMarshal, _ :=  json.Marshal(proof)
	//	//
	//	//prin

	proofBase64 := base64.StdEncoding.EncodeToString(proof.Bytes())

	fmt.Println("proofBase64: %v\n", proofBase64)

	return proofBase64, nil
}

func main() {
	c := make(chan struct{}, 0)
	fmt.Println("Hello WASM")
	RegisterCallback("add", add)
	RegisterCallback("sayHello", sayHello)
	//RegisterCallback("aggregatedRangeProve", aggregatedRangeProve)
	<-c
}
