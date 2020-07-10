package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

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

func GetOutputToken(tool *debugtool.DebugTool, privKey string, tokenID string){
	fmt.Println("========== GET OUTPUT TOKEN ==========")
	b, _ := tool.GetListOutputTokens(privKey, tokenID)
	fmt.Println(string(b))
	fmt.Println("========== END OUTPUT TOKEN ==========")
}
func GetUnspentOutputToken(tool *debugtool.DebugTool, privKey string, tokenID string){
	fmt.Println("========== GET UNSPENT OUTPUT TOKEN ==========")
	b, _ := tool.GetListUnspentOutputTokens(privKey, tokenID)
	fmt.Println(string(b))
	fmt.Println("========== END UNSPENT OUTPUT TOKEN ==========")
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

func ConvertTokenCoinVersion(tool *debugtool.DebugTool, privKey string, tokenID string) {
	fmt.Println("CONVERT TOKEN COIN")
	b, _ := tool.SwitchTokenCoinVersion(privKey, tokenID)
	fmt.Println(string(b))
	fmt.Println("END CONVERT TOKEN COIN")
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

// PDE
func PDEContributePRV(tool *debugtool.DebugTool, privKey, amount string) {
	fmt.Println("========== PDE CONTRIBUTE PRV  ==========")
	b, _ := tool.PDEContributePRV(privKey, amount)
	fmt.Println(string(b))
	fmt.Println("========== END PDE CONTRIBUTE PRV  ==========")
}
func PDEContributeToken(tool *debugtool.DebugTool, privKey, tokenID, amount string) {
	fmt.Println("========== PDE CONTRIBUTE TOKEN  ==========")
	b, _ := tool.PDEContributeToken(privKey, tokenID, amount)
	fmt.Println(string(b))
	fmt.Println("========== END PDE CONTRIBUTE TOKEN  ==========")
}

func PDEWithdrawContribution(tool *debugtool.DebugTool, privKey, tokenID1, tokenID2, amountShare string) {
	fmt.Println("========== PDE WITHDRAW   ==========")
	b, _ := tool.PDEWithdrawContribution(privKey, tokenID1, tokenID2, amountShare)
	fmt.Println(string(b))
	fmt.Println("========== END PDE WITHDRAW  ==========")
}


func PDETradePRV(tool *debugtool.DebugTool, privKey, token, amount string) {
	fmt.Println("========== PDE TRADE PRV   ==========")
	b, _ := tool.PDETradePRV(privKey, token, amount)
	fmt.Println(string(b))
	fmt.Println("========== END TRADE PRV  ==========")
}

func PDETradeToken(tool *debugtool.DebugTool, privKey, token, amount string) {
	fmt.Println("========== PDE TRADE TOKEN   ==========")
	b, _ := tool.PDETradeToken(privKey, token, amount)
	fmt.Println(string(b))
	fmt.Println("========== END PDE TRADE TOKEN  ==========")
}

func main() {
	privateKeys := []string{
		"112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or",
		"112t8rnZDRztVgPjbYQiXS7mJgaTzn66NvHD7Vus2SrhSAY611AzADsPFzKjKQCKWTgbkgYrCPo9atvSMoCf9KT23Sc7Js9RKhzbNJkxpJU6",
		"112t8rne7fpTVvSgZcSgyFV23FYEv3sbRRJZzPscRcTo8DsdZwstgn6UyHbnKHmyLJrSkvF13fzkZ4e8YD5A2wg8jzUZx6Yscdr4NuUUQDAt",
		"112t8rnXoBXrThDTACHx2rbEq7nBgrzcZhVZV4fvNEcGJetQ13spZRMuW5ncvsKA1KvtkauZuK2jV8pxEZLpiuHtKX3FkKv2uC5ZeRC8L6we",
		"112t8rnbcZ92v5omVfbXf1gu7j7S1xxr2eppxitbHfjAMHWdLLBjBcQSv1X1cKjarJLffrPGwBhqZzBvEeA9PhtKeM8ALWiWjhUzN5Fi6WVC",
		"112t8rnZUQXxcbayAZvyyZyKDhwVJBLkHuTKMhrS51nQZcXKYXGopUTj22JtZ8KxYQcak54KUQLhimv1GLLPFk1cc8JCHZ2JwxCRXGsg4gXU",
		"112t8rnXDS4cAjFVgCDEw4sWGdaqQSbKLRH1Hu4nUPBFPJdn29YgUei2KXNEtC8mhi1sEZb1V3gnXdAXjmCuxPa49rbHcH9uNaf85cnF3tMw",
		"112t8rnYoioTRNsM8gnUYt54ThWWrRnG4e1nRX147MWGbEazYP7RWrEUB58JLnBjKhh49FMS5o5ttypZucfw5dFYMAsgDUsHPa9BAasY8U1i",
		"112t8rnXtw6pWwowv88Ry4XxukFNLfbbY2PLh2ph38ixbCbZKwf9ZxVjd4s7jU3RSdKctC7gGZp9piy8nZoLqHwqDBWcsMHWsQg27S5WCdm4",
	}
	privateSeeds := []string{
		"12MZ4QiFoETNbdLKgRQWPMQMqsceWPKo71Jma9NzwvLTabpcDhn",
		"158ZGK5EHmoyrHEd8aA2HCaqbNQ4r45ZsnwL4Zh8mH8dueWHWs",
		"1mYRSzV7yigD7qNpuQwnyKeVMQcnenjSxAB1L8MEpDuT3RRbZc",
		"1G5Q9uGSxekPSgC1w1ZFaDJ8RxeYrekk2FtFLF33QCKNbg2V88",
		"1mYRSzV7yigD7qNpuQwnyKeVMQcnenjSxAB1L8MEpDuT3RRbZc",
		"12MZ4QiFoETNbdLKgRQWPMQMqsceWPKo71Jma9NzwvLTabpcDhn",
		"1G5Q9uGSxekPSgC1w1ZFaDJ8RxeYrekk2FtFLF33QCKNbg2V88",
		"1cQCTV1m33LxBKpNW2SisbuJfp5VcBSEau7PE5aD16gGLAN7eq",
		"1DRtHReKFQqPQzd689EsjivNgUScPBUwvbw8azgxaLRUBtmFL2",
		"12YCyPu6KyBToSaaQkw7hzbWkJnUi78DLkfvWokgi4dCtkbVusC",
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
	//ConvertTokenCoinVersion(tool, privateKeys[0], tokenID)

	//TransferToken(tool, privateKeys[0], privateKeys[1], tokenID,"1000")

	//GetBalanceToken(tool, privateKeys[0], tokenID)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter your choice (arguments separated by ONLY ONE space) and hit ENTER: ")
		text, _ := reader.ReadString('\n')
		args := strings.Split(text[:len(text)-1], " ")
		if len(args) < 1 {
			return
		}
		if args[0] == "port" {
			tool = SwitchPort(args[1])
		}
		if args[0] == "convert" {
			index, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				panic(err)
			}
			ConvertCoinVersion(tool, privateKeys[index])
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
				fmt.Println(err)
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
			if args[2] != "" && len(args[2]) > 0 {
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

		//TOKEN RPC
		if args[0] == "inittoken" {
			if len(args) < 3 {
				panic("Not enough params for initToken")
			}
			index, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				panic(err)
			}
			fmt.Println("[BUGLOG2] privKey =", privateKeys[index])
			InitToken(tool, privateKeys[index], args[2])
		}
		if args[0] == "listtoken" {
			ListTokens(tool)
		}
		if args[0] == "converttoken"{
			if len(args) < 3 {
				panic("Not enough params for converttoken")
			}
			index, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				panic(err)
			}
			tokenID := args[2]
			ConvertTokenCoinVersion(tool, privateKeys[index], tokenID)
		}
		if args[0] == "transfertoken"{
			if len(args) < 5 {
				panic("Not enough params for transfertoken")
			}
			indexFrom, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				panic(err)
			}
			indexTo, err := strconv.ParseInt(args[2], 10, 32)
			if err != nil {
				panic(err)
			}
			TransferToken(tool, privateKeys[indexFrom], privateKeys[indexTo], args[3], args[4])

		}
		if args[0] == "balancetoken" {
			index, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				panic(err)
			}
			GetBalanceToken(tool, privateKeys[index], args[2])
		}
		if args[0] == "outtoken" {
			if len(args) < 3 {
				panic("Not enough param for outtoken")
			}
			index, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				panic(err)
			}
			GetOutputToken(tool, privateKeys[index], args[2])
		}
		if args[0] == "unspentouttoken" {
			if len(args) < 3 {
				panic("Not enough param for unspentouttoken")
			}
			index, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				panic(err)
			}
			GetUnspentOutputToken(tool, privateKeys[index], args[2])
		}
		// PDE
		if args[0] == "pdecontributeprv" {
			index, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				panic(err)
			}
			PDEContributePRV(tool, privateKeys[index], args[2])
		}
		if args[0] == "pdecontributetoken" {
			index, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				panic(err)
			}
			PDEContributeToken(tool, privateKeys[index], args[2], args[3])
		}
		if args[0] == "pdewithdraw" {
			index, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				panic(err)
			}
			PDEWithdrawContribution(tool, privateKeys[index], args[2], args[3], args[4])
		}

		if args[0] == "pdetradeprv" {
			index, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				panic(err)
			}
			PDETradePRV(tool, privateKeys[index], args[2], args[3])
		}

		if args[0] == "pdetradetoken" {
			index, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				panic(err)
			}
			PDETradeToken(tool, privateKeys[index], args[2], args[3])
		}
	}
}
