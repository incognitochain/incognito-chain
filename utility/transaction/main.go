package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/constant-money/constant-chain/database"
	_ "github.com/constant-money/constant-chain/database/lvdb"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wallet"
)

func main() {

	//==========Write
	if os.Args[1] == "write" {
		transactions := []string{}
		db, err := database.Open("leveldb", filepath.Join("./", "./"))
		if err != nil {
			fmt.Print("could not open connection to leveldb")
			fmt.Print(err)
			panic(err)
		}
		for i := 0; i < 500; i++ {
			txs := initTx("1000", "112t8rsq5Xx45T1ZKH4N45aBztqBJiDAR9Nw5wMb8Fe5PnFCqDiUAgVzoMr3xBznNJTfu2CSW3HC6M9rGHxTyUzUBbZHjv6wCMnucDDKbHT4", db)
			transactions = append(transactions, txs[0])
		}
		file, _ := json.MarshalIndent(transactions, "", " ")
		_ = ioutil.WriteFile("test0.json", file, 0644)
	}
	//==========Read
	if os.Args[1] == "read" {
		readTxsFromFile("test-read-write.json")
	}

}
func readTxsFromFile(filename string) {
	// Open our jsonFile
	jsonFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened ", filename)
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result []string
	json.Unmarshal([]byte(byteValue), &result)
	fmt.Println(result)

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
}
func initTx(amount string, privateKey string, db database.DatabaseInterface) []string {
	var initTxs []string
	var initAmount, _ = strconv.Atoi(amount) // amount init
	var spendingKey = privateKey             // spending key str
	testUserkeyList := []string{
		spendingKey,
	}
	for _, val := range testUserkeyList {

		testUserKey, _ := wallet.Base58CheckDeserialize(val)
		testUserKey.KeySet.ImportFromPrivateKey(&testUserKey.KeySet.PrivateKey)

		testSalaryTX := transaction.Tx{}
		testSalaryTX.InitTxSalary(uint64(initAmount), &testUserKey.KeySet.PaymentAddress, &testUserKey.KeySet.PrivateKey,
			db,
			nil,
		)
		initTx, _ := json.Marshal(testSalaryTX)
		initTxs = append(initTxs, string(initTx))
	}
	return initTxs
}
