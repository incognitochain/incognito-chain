package lvdb_test

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/stretchr/testify/assert"
)

var db incdb.Database

var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		log.Fatalf("failed to create temp dir: %+v", err)
	}
	log.Println(dbPath)
	db, err = incdb.Open("leveldb", dbPath)
	if err != nil {
		log.Fatalf("could not open db path: %s, %+v", dbPath, err)
	}
	incdb.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestDb_Setup(t *testing.T) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %+v", err)
	}
	t.Log(dbPath)
	db, err := incdb.Open("leveldb", dbPath)
	if err != nil {
		t.Fatalf("could not open db path: %s, %+v", dbPath, err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("db.close %+v", err)
	}
	os.RemoveAll(dbPath)
}

func TestDb_Base(t *testing.T) {
	if db != nil {
		db.Put([]byte("a"), []byte{1})
		result, err := db.Get([]byte("a"))
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, result[0], []byte{1}[0])
		has, err := db.Has([]byte("a"))
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, has, true)

		err = db.Delete([]byte("a"))
		assert.Equal(t, nil, err)
		err = db.Delete([]byte("b"))
		assert.Equal(t, nil, err)
		has, err = db.Has([]byte("a"))
		assert.Equal(t, err, nil)
		assert.Equal(t, has, false)

		batchData := []incdb.BatchData{}
		batchData = append(batchData, incdb.BatchData{
			Key:   []byte("abc1"),
			Value: []byte("abc1"),
		})
		batchData = append(batchData, incdb.BatchData{
			Key:   []byte("abc2"),
			Value: []byte("abc2"),
		})
		err = db.PutBatch(batchData)
		assert.Equal(t, err, nil)
		v, err := db.Get([]byte("abc2"))
		assert.Equal(t, err, nil)
		assert.Equal(t, "abc2", string(v))
	} else {
		t.Error("DB is not open")
	}
}
