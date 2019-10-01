//+build linux,386 wasm

package main

import (
	"github.com/incognitochain/incognito-chain/wasm/gomobile"
	"syscall/js"
)

//func aggregatedRangeProve(_ js.Value, args []js.Value) interface{} {
//	return gomobile.AggregatedRangeProve(args[0].String())
//}
//func oneOutOfManyProve(_ js.Value, args []js.Value) interface{} {
//	println("args.len :", len(args))
//	proof, err := gomobile.OneOutOfManyProve(args[0].String())
//	if err != nil {
//		return nil
//	}
//	return proof
//}
//func generateBLSKeyPairFromSeed(_ js.Value, args []js.Value) interface{} {
//	return gomobile.GenerateBLSKeyPairFromSeed(args[0].String())
//}

func initTx(_ js.Value, args []js.Value) interface{} {
	result, err := gomobile.InitPrivacyTx(args[0].String())


	if err != nil {
		return nil
	}

	return result
}

func main() {
	c := make(chan struct{}, 0)
	println("Hello WASM")
	//js.Global().Set("aggregatedRangeProve", js.FuncOf(aggregatedRangeProve))
	//js.Global().Set("oneOutOfManyProve", js.FuncOf(oneOutOfManyProve))
	//js.Global().Set("generateBLSKeyPairFromSeed", js.FuncOf(generateBLSKeyPairFromSeed))
	js.Global().Set("initTx", js.FuncOf(initTx))
	<-c
}
