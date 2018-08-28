package lvdb

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/database"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

func setup(t *testing.T) (database.DB, func()) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %+v", err)
	}
	t.Log(dbPath)
	db, err := open(dbPath)
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

	hash := common.Hash([32]byte{})
	block := &blockchain.Block{
		Header:       blockchain.BlockHeader{},
		BlockHash:    &hash,
		Transactions: []transaction.Transaction{},
	}

	err := db.StoreBlock(block)
	if err != nil {
		t.Errorf("db.StoreBlock returns err: %+v", err)
	}

	exists, err := db.HasBlock(&hash)
	if err != nil {
		t.Errorf("db.HasBlock returns err: %+v", err)
	}
	if !exists {
		t.Errorf("block should exists")
	}

	fetched, err := db.FetchBlock(&hash)
	if err != nil {
		t.Errorf("db.FetchBlock returns err: %+v", err)
	}
	blockJSON, _ := json.Marshal(block)
	if !reflect.DeepEqual(blockJSON, fetched) {
		t.Logf("should equal")
	}
}
