package main

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/privacy/debugtool"
)

func main() {
	tool := new(debugtool.DebugTool).InitMainnet()
	b, err := tool.GetSigTransactionByHash("b2d114d576f12898c360de11df67b1f9c2c91e4c32c2b75911c3c544b5476dc5")
	fmt.Println(err)
	fmt.Println(b)
}
