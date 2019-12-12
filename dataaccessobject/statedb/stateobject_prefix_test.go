package statedb_test

import (
	"bytes"
	"github.com/incognitochain/incognito-chain/trie"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
)

var (
	prefix                  = []byte("ser")
	serialNumber1HashPrefix = common.BytesToHash(append(prefix, serialNumber1Hash[:][len(prefix):]...))
	serialNumber2HashPrefix = common.BytesToHash(append(prefix, serialNumber2Hash[:][len(prefix):]...))
)
var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "testIT_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDB = statedb.NewDatabaseAccessWarper(diskBD)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestStoreAndGetStateObjectByPrefix(t *testing.T) {
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, warperDB)
	if err != nil {
		t.Fatal(err)
	}
	if sDB == nil {
		t.Fatal("statedb is nil")
	}

	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber1HashPrefix, serialNumber1)
	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber2HashPrefix, serialNumber2)
	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber3Hash, serialNumber3)

	rootHash1, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(rootHash1)
	if bytes.Compare(rootHash1.Bytes(), emptyRoot.Bytes()) == 0 {
		t.Fatal("root hash is empty")
	}
	err = warperDB.TrieDB().Commit(rootHash1, false)
	if err != nil {
		t.Fatal(err)
	}

	tempStateDB, err := statedb.New(rootHash1, warperDB)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	keys, values := tempStateDB.GetSerialNumberListByPrefix(prefix)
	//keys, values := tempStateDB.GetSerialNumberAllList()
	log.Println(keys)
	log.Println(values)
}
