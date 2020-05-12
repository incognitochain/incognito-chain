package main

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/privacy/debugtool"
)

func sendTx(tool *debugtool.DebugTool) {
	b, _ := tool.CreateAndSendTransaction()
	fmt.Println(string(b))
}

func main() {
	privateKeys := []string{"112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or", "112t8rnZDRztVgPjbYQiXS7mJgaTzn66NvHD7Vus2SrhSAY611AzADsPFzKjKQCKWTgbkgYrCPo9atvSMoCf9KT23Sc7Js9RKhzbNJkxpJU6", "112t8rne7fpTVvSgZcSgyFV23FYEv3sbRRJZzPscRcTo8DsdZwstgn6UyHbnKHmyLJrSkvF13fzkZ4e8YD5A2wg8jzUZx6Yscdr4NuUUQDAt", "112t8rnXoBXrThDTACHx2rbEq7nBgrzcZhVZV4fvNEcGJetQ13spZRMuW5ncvsKA1KvtkauZuK2jV8pxEZLpiuHtKX3FkKv2uC5ZeRC8L6we"}
	////paymentKeys := []string{"", "12RuhVZQtGgYmCVzVi49zFZD7gR8SQx8Uuz8oHh6eSZ8PwB2MwaNE6Kkhd6GoykfkRnHNSHz1o2CzMiQBCyFPikHmjvvrZkLERuhcVE", "12RxDSnQVjPojzf7uju6dcgC2zkKkg85muvQh347S76wKSSsKPAqXkvfpSeJzyEH3PREHZZ6SKsXLkDZbs3BSqwEdxqprqih4VzANK9", "12S6m2LpzN17jorYnLb2ApNKaV2EVeZtd6unvrPT1GH8yHGCyjYzKbywweQDZ7aAkhD31gutYAgfQizb2JhJTgBb3AJ8aB4hyppm2ax"}
	//
	tool := new(debugtool.DebugTool).InitLocal()
	sendTx(tool)
	//
	//fmt.Println("===========================")
	//fmt.Println("Printing output coins after create tx")
	b, _ := tool.GetListOutputCoins(privateKeys[0])
	fmt.Println(string(b))

	//b, _ := tool.GetRawMempool()
	//fmt.Println(string(b))
	//
	//tool := new(debugtool.DebugTool).InitLocal()
	//b, _ := tool.GetTransactionByHash("640bb10fd9781db5e878f7a1bc212e88847791861ce0b0dcc2ff22a8f5905fc5")
	//fmt.Println(string(b))
}

//11
//[161 0 17 98 4 7 232 46 247 220 169 220]