package lvdb_test

import (
	"bytes"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/databasemp"
	"github.com/incognitochain/incognito-chain/databasemp/lvdb"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/transaction"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	dbmp databasemp.DatabaseInterface
	now = time.Now()
	tx1 = transaction.Tx{
		LockTime: now.Unix(),
	}
	tx2 = transaction.Tx{
		LockTime: now.Unix() + 10,
	}
)

var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		log.Fatalf("failed to create temp dir: %+v", err)
	}
	log.Println(dbPath)
	dbmp, err = databasemp.Open("leveldbmempool", dbPath)
	if err != nil {
		log.Fatalf("could not open db path: %s, %+v", dbPath, err)
	}
	databasemp.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()
func TestDb_AddTransaction(t *testing.T) {
	tempDesc := mempool.TempDesc{
		StartTime: time.Now(),
		IsPushMessage: false,
		Height: uint64(1),
		Fee: uint64(1),
		FeePerKB: int32(1),
	}
	valueTx, _ := json.Marshal(&tx1)
	valueTxDesc, _ := json.Marshal(&tempDesc)
	err := dbmp.AddTransaction(tx1.Hash(), common.TxNormalType, valueTx, valueTxDesc)
	if err != nil {
		t.Fatal("Expect no error but get ", err)
	}
}
func TestDb_HasTransaction(t *testing.T) {
	tempDesc := mempool.TempDesc{
		StartTime: time.Now(),
		IsPushMessage: false,
		Height: uint64(1),
		Fee: uint64(1),
		FeePerKB: int32(1),
	}
	valueTx, _ := json.Marshal(&tx1)
	valueTxDesc, _ := json.Marshal(&tempDesc)
	err := dbmp.AddTransaction(tx1.Hash(), common.TxNormalType, valueTx, valueTxDesc)
	if err != nil {
		t.Fatal("Expect no error but get ", err)
	}
	isHas, err := dbmp.HasTransaction(tx1.Hash())
	if err != nil {
		t.Fatal("Expect no error but get ", err)
	}
	if !isHas {
		t.Fatalf("Expect tx %+v in db", *tx1.Hash())
	}
	isHas, err = dbmp.HasTransaction(tx2.Hash())
	if err != nil {
		t.Fatal("Expect no error but get ", err)
	}
	if isHas {
		t.Fatalf("Expect tx %+v NOT in db", *tx2.Hash())
	}
}

func TestDb_GetTransaction(t *testing.T) {
	tempDesc := mempool.TempDesc{
		StartTime: time.Now(),
		IsPushMessage: false,
		Height: uint64(1),
		Fee: uint64(1),
		FeePerKB: int32(1),
	}
	valueTx, err := json.Marshal(&tx1)
	if err != nil {
		t.Fatal("Fail to marshal tx")
	}
	valueTxDesc, err := json.Marshal(&tempDesc)
	if err != nil {
		t.Fatal("Fail to marshal tx")
	}
	err = dbmp.AddTransaction(tx1.Hash(), common.TxNormalType, valueTx, valueTxDesc)
	if err != nil {
		t.Fatal("Expect no error but get ", err)
	}
	data, err := dbmp.GetTransaction(tx1.Hash())
	if err != nil {
		t.Fatal("Expect no error but get ", err)
	}
	strs := strings.Split(string(data),string(lvdb.Splitter))
	if strs[0] != common.TxNormalType {
		t.Fatal("Wrong Tx Type")
	}
	valueTxDb := []byte(strs[1])
	if bytes.Compare(valueTxDb, valueTx) != 0 {
		t.Fatal("Value tx before store and get from db is different")
	}
	valueTxDescDb := []byte(strs[2])
	if bytes.Compare(valueTxDescDb, valueTxDesc) != 0 {
		t.Fatal("Value tx desc before store and get from db is different")
	}
	tx1Db := &transaction.Tx{}
	tempDescDb := &mempool.TempDesc{}
	if err := json.Unmarshal(valueTxDb, tx1Db); err != nil {
		t.Fatal("Fail to unmarshal tx data to tx struct", err)
	} else {
		if!tx1Db.Hash().IsEqual(tx1.Hash()) {
			t.Fatal("Tx before store and get from  pool has different hash")
		}
	}
	if err := json.Unmarshal(valueTxDescDb, tempDescDb); err != nil {
		t.Fatal("Fail to unmarshal tx desc data to temp desc struct", err)
	} else {
		if tempDescDb.StartTime.Second() != tempDesc.StartTime.Second() || tempDesc.FeePerKB != tempDescDb.FeePerKB || tempDesc.Fee != tempDescDb.Fee || tempDescDb.Height != tempDesc.Height || tempDesc.IsPushMessage != tempDescDb.IsPushMessage {
			t.Fatal("Tx DESC before store and get from  pool has different value")
		}
	}
}

func TestDb_RemoveTransaction(t *testing.T) {
	tempDesc := mempool.TempDesc{
		StartTime: time.Now(),
		IsPushMessage: false,
		Height: uint64(1),
		Fee: uint64(1),
		FeePerKB: int32(1),
	}
	valueTx, err := json.Marshal(&tx1)
	if err != nil {
		t.Fatal("Fail to marshal tx")
	}
	valueTxDesc, err := json.Marshal(&tempDesc)
	if err != nil {
		t.Fatal("Fail to marshal tx")
	}
	err = dbmp.AddTransaction(tx1.Hash(), common.TxNormalType, valueTx, valueTxDesc)
	if err != nil {
		t.Fatal("Expect no error but get ", err)
	}
	if err := dbmp.RemoveTransaction(tx1.Hash()); err != nil {
		t.Fatal("Fail to remove transaction")
	}
	if isHas, err := dbmp.HasTransaction(tx1.Hash()); err != nil || isHas {
		t.Fatalf("Expect no error and no transaction but get error %+v or has tx %+v", err, isHas)
	}
}
func TestDb_Load(t *testing.T) {
	tempDesc := mempool.TempDesc{
		StartTime: time.Now(),
		IsPushMessage: false,
		Height: uint64(1),
		Fee: uint64(1),
		FeePerKB: int32(1),
	}
	valueTxDesc, err := json.Marshal(&tempDesc)
	if err != nil {
		t.Fatal("Fail to marshal tx")
	}
	// add tx1
	valueTx, err := json.Marshal(&tx1)
	if err != nil {
		t.Fatal("Fail to marshal tx")
	}
	err = dbmp.AddTransaction(tx1.Hash(), common.TxNormalType, valueTx, valueTxDesc)
	// add tx2
	valueTx2, err := json.Marshal(&tx2)
	if err != nil {
		t.Fatal("Fail to marshal tx")
	}
	err = dbmp.AddTransaction(tx2.Hash(), common.TxNormalType, valueTx2, valueTxDesc)
	// load all tx
	txHashesByte, txsByte, err := dbmp.Load()
	if err != nil {
		t.Fatal("Fail to load transaction from db")
	}
	if len(txsByte) != 2 {
		t.Fatal("Expect only 2 tx from db but get ", len(txsByte))
	}
	if len(txHashesByte) != 2 {
		t.Fatal("Expect only 2 tx from db but get ", len(txHashesByte))
	}
}
func TestDb_Reset(t *testing.T) {
	tempDesc := mempool.TempDesc{
		StartTime: time.Now(),
		IsPushMessage: false,
		Height: uint64(1),
		Fee: uint64(1),
		FeePerKB: int32(1),
	}
	valueTxDesc, err := json.Marshal(&tempDesc)
	if err != nil {
		t.Fatal("Fail to marshal tx")
	}
	// add tx1
	valueTx, err := json.Marshal(&tx1)
	if err != nil {
		t.Fatal("Fail to marshal tx")
	}
	err = dbmp.AddTransaction(tx1.Hash(), common.TxNormalType, valueTx, valueTxDesc)
	// add tx2
	valueTx2, err := json.Marshal(&tx2)
	if err != nil {
		t.Fatal("Fail to marshal tx")
	}
	err = dbmp.AddTransaction(tx2.Hash(), common.TxNormalType, valueTx2, valueTxDesc)
	// reset
	err = dbmp.Reset()
	if err != nil {
		t.Fatal("Fail to reset database")
	}
	// detect if tx exist in mempool
	isHas, err := dbmp.HasTransaction(tx1.Hash())
	if err != nil {
		t.Fatal("Expect no error but get ", err)
	}
	if isHas {
		t.Fatalf("Expect tx %+v NOT in db", *tx1.Hash())
	}
	isHas, err = dbmp.HasTransaction(tx2.Hash())
	if err != nil {
		t.Fatal("Expect no error but get ", err)
	}
	if isHas {
		t.Fatalf("Expect tx %+v NOT in db", *tx2.Hash())
	}
	// load all after reset
	txHashesByte, txsByte, err := dbmp.Load()
	if err != nil {
		t.Fatal("Fail to load transaction from db")
	}
	if len(txsByte) != 0 {
		t.Fatal("Expect only 2 tx from db but get ", len(txsByte))
	}
	if len(txHashesByte) != 0 {
		t.Fatal("Expect only 2 tx from db but get ", len(txHashesByte))
	}
}