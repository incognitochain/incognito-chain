package main

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

func main() {
	initEnvironment()

	var transactions []string
	db, err := incdb.Open("leveldb", filepath.Join("./data", "./"))
	if err != nil {
		fmt.Print("could not open connection to leveldb")
		fmt.Print(err)
		panic(err)
	}
	stateDB, err := statedb.NewWithPrefixTrie(common.EmptyRoot, statedb.NewDatabaseAccessWarper(db))
	if err != nil {
		panic(err)
	}
	privateKeys := []string{
		"1111111CL65f56Khc7NhYHcxFabrWrNE3tbEHJpd4diMJiKd9NHmLAYgdHRDV23rb4qwjYBMohicB5hDhhFb9qxTbBprCt19yKiUu4ajTLB",
	}

	for _, privateKey := range privateKeys {
		txs, err := initTx("1000", privateKey, stateDB)
		if err != nil {
			panic(err)
		}
		transactions = append(transactions, txs[0])
	}
	file, err := json.MarshalIndent(transactions, "", " ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("shard2-2-init-txs.json", file, 0644)
	if err != nil {
		panic(err)
	}
}

func initEnvironment() {
	common.MaxShardNumber = 8
}

func readTxsFromFile(filename string) ([]string, error) {
	// Open our jsonFile
	jsonFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, err
	}
	fmt.Println("Successfully Opened ", filename)
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer func() {
		err = jsonFile.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	var result []string
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		 return nil, err
	}
	return result, nil
}

func initTx(amount string, privateKey string, stateDB *statedb.StateDB) ([]string, error) {
	testKeyList := []string{
		privateKey,
	}

	initTxs := make([]string, 0)
	var initAmount, err = strconv.Atoi(amount) // amount init
	if err != nil {
		return nil, err
	}

	for _, val := range testKeyList {
		w, err := wallet.Base58CheckDeserialize(val)
		if err != nil {
			return nil, err
		}
		err = w.KeySet.InitFromPrivateKey(&w.KeySet.PrivateKey)
		if err != nil {
			return nil, err
		}

		// Generate a new OTA coin
		otaCoin, err := privacy.NewCoinFromPaymentInfo(&privacy.PaymentInfo{PaymentAddress: w.KeySet.PaymentAddress, Amount: uint64(initAmount)})
		if err != nil {
			return nil, err
		}

		// Initialize the tx
		testSalaryTX := transaction.TxVersion2{}
		err = testSalaryTX.InitTxSalary(otaCoin, &(w.KeySet.PrivateKey), stateDB, nil)
		if err != nil {
			return nil, err
		}

		initTx, err := json.Marshal(testSalaryTX)
		if err != nil {
			return nil, err
		}
		initTxs = append(initTxs, string(initTx))
	}

	return initTxs, nil
}
