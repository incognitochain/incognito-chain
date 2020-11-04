package main

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/devframework"
)

func main() {

	//package main
	//
	//type rpc interface{
	//	rpc()
	//}
	//
	//
	//func API(r rpc,i int) {
	//
	//}
	//
	//type bar struct{}
	//func(s *bar) rpc(){
	//	API(s,1)
	//}
	//
	//func main(){
	//	b := &bar{}
	//	b.
	//}
	//
	keyID := 0
	account1, _ := devframework.GenerateAccountByShard(0, keyID, "")
	keyID++
	account2, _ := devframework.GenerateAccountByShard(1, keyID, "")
	keyID++
	account3, _ := devframework.GenerateAccountByShard(2, keyID, "")
	keyID++
	account4, _ := devframework.GenerateAccountByShard(3, keyID, "")
	keyID++
	account5, _ := devframework.GenerateAccountByShard(4, keyID, "")
	keyID++
	account6, _ := devframework.GenerateAccountByShard(5, keyID, "")
	keyID++
	account7, _ := devframework.GenerateAccountByShard(6, keyID, "")
	keyID++
	account8, _ := devframework.GenerateAccountByShard(7, keyID, "")
	keyID++

	json1, _ := json.MarshalIndent(account1, "", "")
	json2, _ := json.MarshalIndent(account2, "", "")
	json3, _ := json.MarshalIndent(account3, "", "")
	json4, _ := json.MarshalIndent(account4, "", "")
	json5, _ := json.MarshalIndent(account5, "", "")
	json6, _ := json.MarshalIndent(account6, "", "")
	json7, _ := json.MarshalIndent(account7, "", "")
	json8, _ := json.MarshalIndent(account8, "", "")

	fmt.Printf("%+v", string(json1))
	fmt.Printf("%+v", string(json2))
	fmt.Printf("%+v", string(json3))
	fmt.Printf("%+v", string(json4))
	fmt.Printf("%+v", string(json5))
	fmt.Printf("%+v", string(json6))
	fmt.Printf("%+v", string(json7))
	fmt.Printf("%+v", string(json8))

}
