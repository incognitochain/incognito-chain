package statedb_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
)

var (
	emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
)

func TestStateDB(t *testing.T) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		t.Fatal(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDB := statedb.NewDatabaseAccessWarper(diskBD)
	sdb, err := statedb.New(emptyRoot, warperDB)
	if err != nil {
		t.Fatal(err)
	}
	if sdb == nil {
		t.Fatal("statedb is nil")
	}
}
