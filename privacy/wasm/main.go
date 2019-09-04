//+build !test

package main

// Uncomment before build
/*import (
	"strconv"
	"syscall/js"

	"github.com/incognitochain/incognito-chain/privacy/wasm/gomobile"
)

func add(_ js.Value, args []js.Value) interface{} {
	a, _ := strconv.Atoi(args[0].String())
	b, _ := strconv.Atoi(args[1].String())
	return gomobile.Add(a, b)
}

func sayHello(_ js.Value, args []js.Value) interface{} {
	return gomobile.SayHello(args[0].String())
}

func aggregatedRangeProve(_ js.Value, args []js.Value) interface{} {
	return gomobile.AggregatedRangeProve(args[0].String())
}

func oneOutOfManyProve(_ js.Value, args []js.Value) interface{} {
	println("args.len :", len(args))
	proof, err := gomobile.OneOutOfManyProve(args[0].String())
	if err != nil {
		return nil
	}
	return proof
}

func generateBLSKeyPairFromSeed(_ js.Value, args []js.Value) interface{} {
	return gomobile.GenerateBLSKeyPairFromSeed(args[0].String())
}

func main() {
	c := make(chan struct{}, 0)
	println("Hello WASM")
	js.Global().Set("add", js.FuncOf(add))
	js.Global().Set("sayHello", js.FuncOf(sayHello))
	js.Global().Set("aggregatedRangeProve", js.FuncOf(aggregatedRangeProve))
	js.Global().Set("oneOutOfManyProve", js.FuncOf(oneOutOfManyProve))
	js.Global().Set("generateBLSKeyPairFromSeed", js.FuncOf(generateBLSKeyPairFromSeed))
	<-c
}*/
