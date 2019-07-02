package lvdb_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/incognitochain/incognito-chain/database"
	_ "github.com/incognitochain/incognito-chain/database/lvdb"
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
