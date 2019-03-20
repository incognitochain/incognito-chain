package main

import (
	"encoding/json"
	"fmt"
	"github.com/constant-money/constant-chain/database"
	_ "github.com/constant-money/constant-chain/database/lvdb"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wallet"
	"os"
	"path/filepath"
)

func main() {
	initTx()
}

func initTx() {
	db, err := database.Open("leveldb", filepath.Join("./", "./"))
	if err != nil {
		fmt.Print("could not open connection to leveldb")
		fmt.Print(err)
		panic(err)
	}
	var initTxs []string
	testUserkeyList := []string{
		"112t8rqnMrtPkJ4YWzXfG82pd9vCe2jvWGxqwniPM5y4hnimki6LcVNfXxN911ViJS8arTozjH4rTpfaGo5i1KKcG1ayjiMsa4E3nABGAqQh",
		"112t8rqGc71CqjrDCuReGkphJ4uWHJmiaV7rVczqNhc33pzChmJRvikZNc3Dt5V7quhdzjWW9Z4BrB2BxdK5VtHzsG9JZdZ5M7yYYGidKKZV",
	}
	for _, val := range testUserkeyList {

		testUserKey, _ := wallet.Base58CheckDeserialize(val)
		testUserKey.KeySet.ImportFromPrivateKey(&testUserKey.KeySet.PrivateKey)

		testSalaryTX := transaction.Tx{}
		testSalaryTX.InitTxSalary(1000000, &testUserKey.KeySet.PaymentAddress, &testUserKey.KeySet.PrivateKey,
			db,
			nil,
		)
		initTx, _ := json.Marshal(testSalaryTX)
		initTxs = append(initTxs, string(initTx))
	}
	fmt.Println(initTxs)
	os.Exit(1)
}
