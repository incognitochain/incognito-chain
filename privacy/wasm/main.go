package main

import (
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

func main() {
	c := make(chan struct{}, 0)
	println("Hello WASM")
	RegisterCallback("add", add)
	RegisterCallback("sayHello", sayHello)
	<-c
}
