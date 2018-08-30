package lvdb_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/database"
	_ "github.com/ninjadotorg/cash-prototype/database/lvdb"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

func setup(t *testing.T) (database.DB, func()) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %+v", err)
	}
	t.Log(dbPath)
	db, err := database.Open("leveldb", dbPath)
	if err != nil {
		t.Fatalf("could not open db path: %s, %+v", dbPath, err)
	}
	return db, func() {
		if err := db.Close(); err != nil {
			t.Fatalf("db.close %+v", err)
		}
		os.RemoveAll(dbPath)
	}
}

func TestBlock(t *testing.T) {
	db, teardown := setup(t)
	defer teardown()

	block := &blockchain.Block{
		Header:       blockchain.BlockHeader{},
		Transactions: []transaction.Transaction{},
	}

	err := db.StoreBlock(block)
	if err != nil {
		t.Errorf("db.StoreBlock returns err: %+v", err)
	}

	exists, err := db.HasBlock(block.Hash())
	if err != nil {
		t.Errorf("db.HasBlock returns err: %+v", err)
	}
	if !exists {
		t.Errorf("block should exists")
	}

	fetched, err := db.FetchBlock(block.Hash())
	if err != nil {
		t.Errorf("db.FetchBlock returns err: %+v", err)
	}
	blockJSON, _ := json.Marshal(block)
	if !bytes.Equal(blockJSON, fetched) {
		t.Logf("should equal")
	}
}

func TestStoreTxOut(t *testing.T) {
	db, teardown := setup(t)
	defer teardown()

	tx := []byte("abcd")
	err := db.StoreTx(tx)
	if err != nil {
		t.Errorf("db.StoreTx %v", err)
	}

	tx = []byte("efgh")
	err = db.StoreTx(tx)
	if err != nil {
		t.Errorf("db.StoreTx %v", err)
	}
}
