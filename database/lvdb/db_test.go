package lvdb_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/database"
	_ "github.com/ninjadotorg/constant/database/lvdb"
	"github.com/ninjadotorg/constant/metadata"
)

func setup(t *testing.T) (database.DatabaseInterface, func()) {
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
		Transactions: []metadata.Transaction{},
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
	err := db.StoreNullifiers(tx)
	if err != nil {
		t.Errorf("db.StoreNullifiers %+v", err)
	}

	tx = []byte("efgh")
	err = db.StoreNullifiers(tx)
	if err != nil {
		t.Errorf("db.StoreNullifiers %+v", err)
	}
}
