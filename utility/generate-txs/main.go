package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/incognitochain/incognito-chain/database"
	_ "github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

func main() {
	//==========Write
	transactions := []string{}
	db, err := database.Open("leveldb", filepath.Join("./", "./"))
	if err != nil {
		fmt.Print("could not open connection to leveldb")
		fmt.Print(err)
		panic(err)
	}
	//privateKeys := readTxsFromFile("private-keys-shard-1-1.json")
	//fmt.Println(len(privateKeys))
	privateKeys := []string{
		"112t8rpnK9Hq2EqZbGJpoS2t7rB3z46sFeRwogUjvzMYHhfBvB3B2X1Mx9W1jCahUZ9aXnbcmrhLXQJsjKDzMvX8vbsA8wjKDArxXfknsVy9",
		"112t8rq19Uu7UGbTApZzZwCAvVszAgRNAzHzr3p8Cu75jPH3h5AUtRXMKiqF3hw8NbEfeLcjtbpeUvJfw4tGj7pbqwDYngc8wB13Gf77o33f",
		"112t8rrEW3NPNgU8xzbeqE7cr4WTT8JvyaQqSZyczA5hBJVvpQMTBVqNfcCdzhvquWCHH11jHihZtgyJqbdWPhWYbmmsw5aV29WSXBEsgbVX",
		"112t8roHikeAFyuBpdCU76kXurEqrC9VYWyRyfFb6PwX6nip9KGYbwpXL78H92mUoWK2GWkA2WysgXbHqwSxnC6XCkmtxBVb3zJeCXgfcYyL",
		"112t8rr4sE2L8WzsVNEN9WsiGcMTDCmEH9TC1ZK8517cxURRFNoWoStYQTgqXpiAMU4gzmkmnWahHdGvQqFaY1JTVsn3nHfD5Ppgz8hQDiVC",
		"112t8rtt9Kd5LUcfXNmd7aMnQehCnKabArVB3BUk2RHVjeh88x5MJnJY4okB8JdFm4JNm4A2WjSe58qWNVkJPEFjpLHNYfKHpWfRdqyfDD9f",
	}
	for _, privateKey := range privateKeys {
		txs := initTx("1000000000000000", privateKey, db)
		transactions = append(transactions, txs[0])
	}
	fmt.Println(len(transactions))
	file, _ := json.MarshalIndent(transactions, "", " ")
	_ = ioutil.WriteFile("shard0-0-init-txs.json", file, 0644)
}
func readTxsFromFile(filename string) []string {
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
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	return result
}
func initTx(amount string, privateKey string, db database.DatabaseInterface) []string {
	var initTxs []string
	var initAmount, _ = strconv.Atoi(amount) // amount init
	testUserkeyList := []string{
		privateKey,
	}
	for _, val := range testUserkeyList {

		testUserKey, _ := wallet.Base58CheckDeserialize(val)
		testUserKey.KeySet.InitFromPrivateKey(&testUserKey.KeySet.PrivateKey)

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
