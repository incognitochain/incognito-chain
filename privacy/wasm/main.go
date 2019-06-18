package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"syscall/js"
	"time"
)


func sayHello(this js.Value, args []js.Value) interface{} {
	fmt.Printf("Hello, %v\n", args[0].String())
	return nil
}

func AggregatedRangeWitnessProve(this js.Value, args []js.Value) interface{} {
	bytes := []byte(args[0].String())

	wit := new(zkp.AggregatedRangeWitness)

	json.Unmarshal(bytes, &wit)

	start := time.Now()
	proof, err := wit.Prove()
	if err != nil {
		fmt.Printf("Err: %v\n", err)
	}
	end := time.Since(start)
	fmt.Printf("Aggregated range proving time: %v\n", end)

	fmt.Printf("Proof json marshal: %v\n", json.Marshal(proof))

	proofBase64 := base64.StdEncoding.EncodeToString(proof.Bytes())

	fmt.Printf("proofBase64: %v\n", proofBase64)

	return proofBase64
}

func main() {
	c := make(chan struct{}, 0)
	fmt.Println("Hello, WebAssembly!")

	// register add function in onclick event with golang add funtion
	js.Global().Set("add", js.FuncOf(AggregatedRangeWitnessProve))

	// register sayHello function
	js.Global().Set("sayHello", js.FuncOf(sayHello))
	<-c
}
