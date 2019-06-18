package main

import (
	"fmt"
	"syscall/js"
)

func AggregatedRangeWitnessProve(this js.Value, args []js.Value) interface{} {
	return nil
}

func main() {
	c := make(chan struct{}, 0)
	fmt.Println("Hello, WebAssembly!")

	// register add function in onclick event with golang add funtion
	js.Global().Set("add", js.FuncOf(AggregatedRangeWitnessProve))
	<-c
}
