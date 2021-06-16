package debugtool

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/wallet"
	"testing"
	"time"
)

func TestGetBalance(t *testing.T) {
	tool := new(DebugTool)
	tool.SetNetwork("http://51.161.119.66:9334")

	for index := range privJKeyList {

		keyWallet, _ := wallet.Base58CheckDeserialize(privJKeyList[index])
		keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
		shardID := byte(int(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1]) % NoOfShard)
		viewingKeyStr := keyWallet.Base58CheckSerialize(wallet.ReadonlyKeyType)
		paymentAddressStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
		paymentAddressStr, _ = wallet.GetPaymentAddressV1(paymentAddressStr, false)

		listTokens, err := GetListToken(viewingKeyStr)
		if listTokens == nil && err != nil{
			fmt.Println("Cannot get list token", viewingKeyStr, privJKeyList[index])
			return
		}
		total1 := time.Duration(0) * time.Millisecond
		total2 := time.Duration(0) * time.Millisecond
		for tokenID, tokenDetail := range listTokens {
			start := time.Now()
			balance, err := GetBalanceFromCS(privJKeyList[index], viewingKeyStr, tokenID, tokenDetail.Total, shardID)
			if err != nil {
				fmt.Println("error ", err)
			}
			elapsed1 := time.Since(start)
			total1 += elapsed1

			start = time.Now()
			balancePrime, err := GetBalanceFromRPC(tool, privJKeyList[index], paymentAddressStr, viewingKeyStr, tokenID, shardID, 0)
			if err != nil {
				fmt.Println(err)
			}
			elapsed2 := time.Since(start)
			total2 += elapsed2

			if balance != balancePrime {
				panic(fmt.Sprintf("Balance %v %v- %v", tokenID, balance, balancePrime))
			}
			fmt.Printf("Balance token %v: %v (%v) - %v (%v)\n", tokenID, balance, elapsed1, balancePrime, elapsed2)
		}
		fmt.Println("=======", total1, total2, "=======")
		return
	}
}


func BenchmarkCheckKeyImageFromCS(b *testing.B) {
	tool := new(DebugTool)
	tool.SetNetwork("http://51.161.119.66:9334")

	index := 0
	keyWallet, _ := wallet.Base58CheckDeserialize(privJKeyList[index])
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	shardID := byte(int(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1]) % NoOfShard)
	viewingKeyStr := keyWallet.Base58CheckSerialize(wallet.ReadonlyKeyType)
	paymentAddressStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	paymentAddressStr, _ = wallet.GetPaymentAddressV1(paymentAddressStr, false)
	tokenID := "0000000000000000000000000000000000000000000000000000000000000004"

	listOutputCoins, _, err :=GetOutputCoins(viewingKeyStr, tokenID)
	if err != nil {
		fmt.Printf("cannot get list output coins from CS. Error %v", err)
		return
	}
	_, listKeyImages, err := GetListDecryptedCoins(privJKeyList[index], listOutputCoins)
	if err != nil {
		fmt.Printf("cannot get plain coins from CS. Error %v", err)
	}

	b.ResetTimer()
	_, err = CheckCoinsSpent(shardID, listKeyImages)
	if err != nil {
		fmt.Printf("cannot get key image status. Error %v", err)
	}
}

func BenchmarkCheckKeyImageFromRPC(b *testing.B) {
	tool := new(DebugTool)
	tool.SetNetwork("http://51.161.119.66:9334")

	index := 0
	keyWallet, _ := wallet.Base58CheckDeserialize(privJKeyList[index])
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	//shardID := byte(int(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1]) % NoOfShard)
	viewingKeyStr := keyWallet.Base58CheckSerialize(wallet.ReadonlyKeyType)
	paymentAddressStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	paymentAddressStr, _ = wallet.GetPaymentAddressV1(paymentAddressStr, false)
	tokenID := "0000000000000000000000000000000000000000000000000000000000000004"

	listOutputCoins, _, err :=GetOutputCoins(viewingKeyStr, tokenID)
	if err != nil {
		fmt.Printf("cannot get list output coins from CS. Error %v", err)
		return
	}
	_, listKeyImages, err := GetListDecryptedCoins(privJKeyList[index], listOutputCoins)
	if err != nil {
		fmt.Printf("cannot get plain coins from CS. Error %v", err)
	}

	b.ResetTimer()
	_, err = CheckCoinSpentFromRPC(tool, listKeyImages, paymentAddressStr, tokenID)
	if err != nil {
		fmt.Printf("cannot get key image status. Error %v", err)
	}
}

func BenchmarkGetCoinFromCS(b *testing.B) {
	tool := new(DebugTool)
	tool.SetNetwork("http://51.161.119.66:9334")

	index := 0
	keyWallet, _ := wallet.Base58CheckDeserialize(privJKeyList[index])
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	//shardID := byte(int(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1]) % NoOfShard)
	viewingKeyStr := keyWallet.Base58CheckSerialize(wallet.ReadonlyKeyType)
	paymentAddressStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	paymentAddressStr, _ = wallet.GetPaymentAddressV1(paymentAddressStr, false)
	tokenID := "0000000000000000000000000000000000000000000000000000000000000004"

	b.ResetTimer()
	_, _, err :=GetOutputCoins(viewingKeyStr, tokenID)
	if err != nil {
		fmt.Printf("cannot get list output coins from CS. Error %v", err)
		return
	}
}

func BenchmarkGetCoinFromRPC(b *testing.B) {
	tool := new(DebugTool)
	tool.SetNetwork("http://51.161.119.66:9334")

	index := 0
	keyWallet, _ := wallet.Base58CheckDeserialize(privJKeyList[index])
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	//shardID := byte(int(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1]) % NoOfShard)
	viewingKeyStr := keyWallet.Base58CheckSerialize(wallet.ReadonlyKeyType)
	paymentAddressStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	paymentAddressStr, _ = wallet.GetPaymentAddressV1(paymentAddressStr, false)
	tokenID := "0000000000000000000000000000000000000000000000000000000000000004"

	b.ResetTimer()
	_, _, err :=GetOutputCoins(viewingKeyStr, tokenID)
	if err != nil {
		fmt.Printf("cannot get list output coins from CS. Error %v", err)
		return
	}
}