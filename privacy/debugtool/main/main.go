package main

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy/debugtool"
	"github.com/incognitochain/incognito-chain/wallet"
)

func sendTx(tool *debugtool.DebugTool) {
	b, _ := tool.CreateAndSendTransaction()
	fmt.Println(string(b))
}

func testInitToken(tool *debugtool.DebugTool, privateKeys []string, privateSeeds []string) {
	//b, _ := tool.CreateAndSendTransaction()
	//fmt.Println(string(b))

	//b, _ := tool.CreateAndSendPrivacyCustomTokenTransaction(privateKeys[0] )
	//fmt.Println(string(b))
	//
	//b, _ := tool.ListPrivacyCustomToken()
	//res := new(debugtool.ListCustomToken)
	//_ = json.Unmarshal(b, res)
	//fmt.Println(len(res.Result.ListCustomToken))
	//tokenID := res.Result.ListCustomToken[0].ID
	//fmt.Println(string(tokenID))
	//b, _ = tool.TransferPrivacyCustomToken(privateKeys[0], privateKeys[1], tokenID, "4000")
	//fmt.Println(string(b))

	b, _ := tool.ListPrivacyCustomToken()
	res := new(debugtool.ListCustomToken)
	_ = json.Unmarshal(b, res)
	tokenID := res.Result.ListCustomToken[0].ID
	//b, _ = tool.TransferPrivacyCustomToken(privateKeys[2], privateKeys[3], tokenID, "1000")
	//fmt.Println(string(b))

	b, _ = tool.GetBalancePrivacyCustomToken(privateKeys[0], tokenID)
	fmt.Println(string(b))
}

//TokenID of coin = [191 233 71 70 219 216 161 134 234 186 130 184 9 0 113 206 223 234 207 61 3 200 250 114 247 71 54 72 233 45 223 116]

//With tokenID = [191 233 71 70 219 216 161 134 234 186 130 184 9 0 113 206 223 234 207 61 3 200 250 114 247 71 54 72 233 45 223 116]

func privateKeyToPaymentAddress(privkey string) string {
	keyWallet, _ := wallet.Base58CheckDeserialize(privkey)
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	paymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	return paymentAddStr
}

func privateKeyToPublicKey(privkey string) []byte {
	keyWallet, _ := wallet.Base58CheckDeserialize(privkey)
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	return keyWallet.KeySet.PaymentAddress.Pk
}

func main() {
	privateKeys := []string{"112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or", "112t8rnZDRztVgPjbYQiXS7mJgaTzn66NvHD7Vus2SrhSAY611AzADsPFzKjKQCKWTgbkgYrCPo9atvSMoCf9KT23Sc7Js9RKhzbNJkxpJU6", "112t8rne7fpTVvSgZcSgyFV23FYEv3sbRRJZzPscRcTo8DsdZwstgn6UyHbnKHmyLJrSkvF13fzkZ4e8YD5A2wg8jzUZx6Yscdr4NuUUQDAt", "112t8rnXoBXrThDTACHx2rbEq7nBgrzcZhVZV4fvNEcGJetQ13spZRMuW5ncvsKA1KvtkauZuK2jV8pxEZLpiuHtKX3FkKv2uC5ZeRC8L6we"}
	privateSeeds := []string{"12MZ4QiFoETNbdLKgRQWPMQMqsceWPKo71Jma9NzwvLTabpcDhn", "158ZGK5EHmoyrHEd8aA2HCaqbNQ4r45ZsnwL4Zh8mH8dueWHWs", "1mYRSzV7yigD7qNpuQwnyKeVMQcnenjSxAB1L8MEpDuT3RRbZc", "1G5Q9uGSxekPSgC1w1ZFaDJ8RxeYrekk2FtFLF33QCKNbg2V88", "1cQCTV1m33LxBKpNW2SisbuJfp5VcBSEau7PE5aD16gGLAN7eq"}
	////paymentKeys := []string{"", "12RuhVZQtGgYmCVzVi49zFZD7gR8SQx8Uuz8oHh6eSZ8PwB2MwaNE6Kkhd6GoykfkRnHNSHz1o2CzMiQBCyFPikHmjvvrZkLERuhcVE", "12RxDSnQVjPojzf7uju6dcgC2zkKkg85muvQh347S76wKSSsKPAqXkvfpSeJzyEH3PREHZZ6SKsXLkDZbs3BSqwEdxqprqih4VzANK9", "12S6m2LpzN17jorYnLb2ApNKaV2EVeZtd6unvrPT1GH8yHGCyjYzKbywweQDZ7aAkhD31gutYAgfQizb2JhJTgBb3AJ8aB4hyppm2ax"}

	//fmt.Println(privateKeyToPaymentAddress(privateKeys[2]))
	//tool := new(debugtool.DebugTool).InitLocal()
	tool := new(debugtool.DebugTool).InitLocal()
	testInitToken(tool, privateKeys, privateSeeds)

	//b, _ := tool.SwitchCoinVersion(privateKeys[0])
	//fmt.Println(string(b))

	//sendTx(tool)

	//fmt.Println("===========================")
	//fmt.Println("Printing output coins after create tx")

	//
	//b, _ := tool.GetListOutputCoins(privateKeys[0])
	//fmt.Println(string(b))
	//b, _ := tool.GetListOutputCoins(privateKeys[1])
	//fmt.Println(string(b))
	//b, _ = tool.GetListOutputCoins(privateKeys[2])
	//fmt.Println(string(b))
	//b, _ = tool.GetListOutputCoins(privateKeys[3])
	//fmt.Println(string(b))

	//b, _ := tool.CreateAndSendTransactionFromAToB(privateKeys[1], privateKeys[3], "10")
	//fmt.Println(string(b))

	//b, _ := tool.GetRawMempool()
	//fmt.Println(string(b))

	//tool := new(debugtool.DebugTool).InitLocal()
	//b, _ := tool.GetTransactionByHash("84a958bedab870c9354aab7f60a9463ecbf7f40ff9eb03a8e0fb3022662f9620")
	//fmt.Println(string(b))

	//b, _ := tool.GetBalanceByPrivatekey(privateKeys[0])
	//fmt.Println(string(b))
	//b, _ = tool.GetBalanceByPrivatekey(privateKeys[1])
	//fmt.Println(string(b))
	//b, _ = tool.GetBalanceByPrivatekey(privateKeys[2])
	//fmt.Println(string(b))
	//b, _ = tool.GetBalanceByPrivatekey(privateKeys[3])
	//fmt.Println(string(b))
	//b, _ := tool.Stake(privateKeys[3], privateSeeds[3])
	//fmt.Println(string(b))
	//b, _ = tool.Stake(privateKeys[2], privateSeeds[2])
	//fmt.Println(string(b))
	//b, _ = tool.Stake(privateKeys[1], privateSeeds[1])
	//fmt.Println(string(b))

	//b, _ := tool.Unstake(privateKeys[3], privateSeeds[3])
	//fmt.Println(string(b))
	//b, _ = tool.Unstake(privateKeys[2], privateSeeds[2])
	//fmt.Println(string(b))
	//b, _ = tool.Unstake(privateKeys[1], privateSeeds[1])
	//fmt.Println(string(b))

	//b, _ := tool.WithdrawReward(privateKeys[1], "0000000000000000000000000000000000000000000000000000000000000040")
	//fmt.Println(string(b))

}

//&{[97 178 112 109 99 221 186 221 159 127 25 42 190 184 219 222 203 101 49 169 3 36 251 148 227 74 0 167 82 51 60 85]}
//
//Found  2  coins
//&{[251 216 135 215 176 83 199 103 193 155 4 16 242 114 73 1 180 212 12 147 197 65 68 94 116 125 57 122 100 143 9 114]}
//&{[36 52 58 141 30 143 208 227 212 204 44 186 78 60 117 217 137 138 27 99 80 35 230 24 74 123 180 40 222 253 28 173]}
