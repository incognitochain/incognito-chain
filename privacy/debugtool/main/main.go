package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	//"os"
	//"strconv"

	"github.com/incognitochain/incognito-chain/privacy/debugtool"
	"github.com/incognitochain/incognito-chain/wallet"
)

func sendTx(tool *debugtool.DebugTool) {
	b, _ := tool.CreateAndSendTransaction()
	fmt.Println(string(b))
}

func InitToken(tool *debugtool.DebugTool, privKey, tokenName string) {
	fmt.Println("========== INIT TOKEN ==========")
	b, _ := tool.CreateAndSendPrivacyCustomTokenTransaction(privKey, tokenName)
	fmt.Println(string(b))
	fmt.Println("========== END INIT TOKEN ==========")
}

func ListTokens(tool *debugtool.DebugTool) *debugtool.ListCustomToken {
	fmt.Println("========== LIST ALL TOKEN ==========")
	b, _ := tool.ListPrivacyCustomToken()
	res := new(debugtool.ListCustomToken)
	_ = json.Unmarshal(b, res)
	fmt.Println("Number of Token: ", len(res.Result.ListCustomToken))
	if len(res.Result.ListCustomToken) > 0 {
		for _, token := range res.Result.ListCustomToken {
			fmt.Println("Token ", token.Name, token.ID)
		}
		fmt.Println("========== END LIST ALL TOKEN ==========")
		return res
	}
	fmt.Println("========== END LIST ALL TOKEN ==========")
	return nil
}

func TransferToken(tool *debugtool.DebugTool, fromPrivKey, toPrivKey, tokenID, amount string) {
	fmt.Println("========== TRANSFER TOKEN ==========")
	b, _ := tool.TransferPrivacyCustomToken(fromPrivKey, toPrivKey, tokenID, amount)
	fmt.Println(string(b))
	fmt.Println("========== END TRANSFER TOKEN ==========")
}

func GetBalanceToken(tool *debugtool.DebugTool, privkey, tokenID string) {
	fmt.Println("========== GET TOKEN BALANCE ==========")
	b, _ := tool.GetBalancePrivacyCustomToken(privkey, tokenID)
	fmt.Println(string(b))
	fmt.Println("========== END GET TOKEN BALANCE ==========")
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

func ConvertCoinVersion(tool *debugtool.DebugTool, privKey string) {
	fmt.Println("CONVERT COIN")
	b, _ := tool.SwitchCoinVersion(privKey)
	fmt.Println(string(b))
	fmt.Println("END CONVERT COIN")
}

func GetPRVOutPutCoin(tool *debugtool.DebugTool, privkey string) {
	fmt.Println("========== GET PRV OUTPUT COIN ==========")
	b, _ := tool.GetListOutputCoins(privkey)
	fmt.Println(string(b))
	fmt.Println("========== END GET PRV OUTPUT COIN ==========")
}

func GetPRVBalance(tool *debugtool.DebugTool, privkey string) {
	fmt.Println("========== GET PRV BALANCE ==========")
	b, _ := tool.GetBalanceByPrivatekey(privkey)
	fmt.Println(string(b))
	fmt.Println("========== END GET PRV BALANCE ==========")
}

func GetRawMempool(tool *debugtool.DebugTool) {
	fmt.Println("========== GET RAW MEMPOOL ==========")
	b, _ := tool.GetRawMempool()
	fmt.Println(string(b))
	fmt.Println("========== END GET RAW MEMPOOL ==========")
}

func GetTxByHash(tool *debugtool.DebugTool, txHash string) {
	fmt.Println("========== GET TX BY HASH ==========")
	b, _ := tool.GetTransactionByHash(txHash)
	fmt.Println(string(b))
	fmt.Println("========== END GET TX BY HASH ==========")
}

func Staking(tool *debugtool.DebugTool, privKey string, privSeed string) {
	fmt.Println("========== STAKING  ==========")
	b, _ := tool.Stake(privKey, privSeed)
	fmt.Println(string(b))
	fmt.Println("========== END STAKING ==========")
}

func UnStaking(tool *debugtool.DebugTool, privKey string, privSeed string) {
	fmt.Println("========== UN-STAKING  ==========")
	b, _ := tool.Unstake(privKey, privSeed)
	fmt.Println(string(b))
	fmt.Println("========== END UN-STAKING ==========")
}

func WithdrawReward(tool *debugtool.DebugTool, privKey string, tokenID string) {
	fmt.Println("========== WITHDRAW REWARD  ==========")
	b, _ := tool.WithdrawReward(privKey, tokenID)
	fmt.Println(string(b))
	fmt.Println("========== END WITHDRAW REWARD  ==========")
}

func SwitchPort(newPort string) *debugtool.DebugTool {
	tool := new(debugtool.DebugTool).InitLocal(newPort)
	return tool
}

func TransferPRV(tool *debugtool.DebugTool, fromPrivKey, toPrivKey, amount string) {
	fmt.Println("========== TRANSFER PRV  ==========")
	b, _ := tool.CreateAndSendTransactionFromAToB(fromPrivKey, toPrivKey, amount)
	fmt.Println(string(b))
	fmt.Println("========== END TRANSFER PRV  ==========")
}

func main() {
	privateKeys := []string{
		"112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or",
		"112t8rnZDRztVgPjbYQiXS7mJgaTzn66NvHD7Vus2SrhSAY611AzADsPFzKjKQCKWTgbkgYrCPo9atvSMoCf9KT23Sc7Js9RKhzbNJkxpJU6",
		"112t8rne7fpTVvSgZcSgyFV23FYEv3sbRRJZzPscRcTo8DsdZwstgn6UyHbnKHmyLJrSkvF13fzkZ4e8YD5A2wg8jzUZx6Yscdr4NuUUQDAt",
		"112t8rnXoBXrThDTACHx2rbEq7nBgrzcZhVZV4fvNEcGJetQ13spZRMuW5ncvsKA1KvtkauZuK2jV8pxEZLpiuHtKX3FkKv2uC5ZeRC8L6we",
		"112t8rnXWRThUTJQgoyH6evV8w19dFZfKWpCh8rZpfymW9JTgKPEVQS44nDRPpsooJiGStHxu81m3HA84t9DBVobz8hgBKRMcz2hddPWNX9N",
		"112t8rnXfotCdLe7Gb8z13Qr2QqAneRYQPYN8faYTp8WLZCbd2JY2iSdB2qgxhhrkQ2PNZxrxZAj8944TjEEdjzPbhUTKUspxhA4vTFDW6aV",
		"112t8rnX2FzscSBNqNAMuQfeSQhMPamSAKDX9f7X7Nft6JR4hfxNh5WJ8r9jeAmbmzTytW8hvpeXVn9FwLD8fjuvn4k7oHtXeh2YxnMDUzZv",
		"112t8rnXoEWG5H8x1odKxSj6sbLXowTBsVVkAxNWr5WnsbSTDkRiVrSdPy8QfMujntKRYBqywKMJCyhMpdr93T3XiUD5QJR1QFtTpYKpjBEx",
	}
	privateSeeds := []string{
		"12MZ4QiFoETNbdLKgRQWPMQMqsceWPKo71Jma9NzwvLTabpcDhn",
		"158ZGK5EHmoyrHEd8aA2HCaqbNQ4r45ZsnwL4Zh8mH8dueWHWs",
		"1mYRSzV7yigD7qNpuQwnyKeVMQcnenjSxAB1L8MEpDuT3RRbZc",
		"1G5Q9uGSxekPSgC1w1ZFaDJ8RxeYrekk2FtFLF33QCKNbg2V88",
		"1mYRSzV7yigD7qNpuQwnyKeVMQcnenjSxAB1L8MEpDuT3RRbZc",
		"158ZGK5EHmoyrHEd8aA2HCaqbNQ4r45ZsnwL4Zh8mH8dueWHWs",
		"1G5Q9uGSxekPSgC1w1ZFaDJ8RxeYrekk2FtFLF33QCKNbg2V88",
		"1cQCTV1m33LxBKpNW2SisbuJfp5VcBSEau7PE5aD16gGLAN7eq",
	}

	//paymentKeys := []string{
	//	"",
	//	"12RuhVZQtGgYmCVzVi49zFZD7gR8SQx8Uuz8oHh6eSZ8PwB2MwaNE6Kkhd6GoykfkRnHNSHz1o2CzMiQBCyFPikHmjvvrZkLERuhcVE",
	//	"12RxDSnQVjPojzf7uju6dcgC2zkKkg85muvQh347S76wKSSsKPAqXkvfpSeJzyEH3PREHZZ6SKsXLkDZbs3BSqwEdxqprqih4VzANK9",
	//	"12S6m2LpzN17jorYnLb2ApNKaV2EVeZtd6unvrPT1GH8yHGCyjYzKbywweQDZ7aAkhD31gutYAgfQizb2JhJTgBb3AJ8aB4hyppm2ax"}

	tool := new(debugtool.DebugTool).InitLocal("9334")

	//InitToken(tool, privateKeys[0], "something")

	//a := ListTokens(tool)
	//tokenID := a.Result.ListCustomToken[0].ID

	//TransferToken(tool, privateKeys[0], privateKeys[1], tokenID,"1000")

	//GetBalanceToken(tool, privateKeys[1], tokenID)

	if len(os.Args) <= 1 {
		return
	}

	args := os.Args[1:]
	if args[0] == "port" {
		tool = SwitchPort(args[1])
	}
	if args[0] == "convert" {
		ConvertCoinVersion(tool, privateKeys[0])
	}
	if args[0] == "send" {
		sendTx(tool)
	}
	if args[0] == "outcoin" {
		index, err := strconv.ParseInt(args[1], 10, 32)
		if err != nil {
			panic(err)
		}
		GetPRVOutPutCoin(tool, privateKeys[index])
	}
	if args[0] == "balance" {
		index, err := strconv.ParseInt(args[1], 10, 32)
		if err != nil {
			panic(err)
		}
		GetPRVBalance(tool, privateKeys[index])
	}
	if args[0] == "mempool" {
		GetRawMempool(tool)
	}
	if args[0] == "txhash" {
		GetTxByHash(tool, args[1])
	}
	if args[0] == "staking" {
		index, err := strconv.ParseInt(args[1], 10, 32)
		if err != nil {
			panic(err)
		}
		Staking(tool, privateKeys[index], privateSeeds[index])
	}
	if args[0] == "unstaking" {
		index, err := strconv.ParseInt(args[1], 10, 32)
		if err != nil {
			panic(err)
		}
		UnStaking(tool, privateKeys[index], privateSeeds[index])
	}

	if args[0] == "reward" {
		index, err := strconv.ParseInt(args[1], 10, 32)
		if err != nil {
			panic(err)
		}
		tokenID := "0000000000000000000000000000000000000000000000000000000000000004"
		if len(args[2]) > 0 {
			tokenID = args[2]
		}
		WithdrawReward(tool, privateKeys[index], tokenID)
	}
	if args[0] == "transfer" {
		indexFrom, err := strconv.ParseInt(args[1], 10, 32)
		if err != nil {
			panic(err)
		}
		indexTo, err := strconv.ParseInt(args[2], 10, 32)
		if err != nil {
			panic(err)
		}
		TransferPRV(tool, privateKeys[indexFrom], privateKeys[indexTo], args[3])
	}

	if args[0] == "payment" {
		fmt.Println("Payment Address", privateKeyToPaymentAddress(args[1]))
	}

	if args[0] == "public" {
		fmt.Println("Public Key", privateKeyToPublicKey(args[1]))
	}

}
