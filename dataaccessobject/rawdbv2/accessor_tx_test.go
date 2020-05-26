package rawdbv2_test

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"testing"
)

var (
	dbTx incdb.Database
)
var _ = func() (_ struct{}) {
	dbSPath, err := ioutil.TempDir(os.TempDir(), "test_rawdbv2")
	if err != nil {
		panic(err)
	}
	dbTx, err = incdb.Open("leveldb", dbSPath)
	if err != nil {
		panic(err)
	}
	return
}()

func generatePublicKey(limit int) [][]byte {
	publicKeys := [][]byte{}
	for i := 0; i < limit; i++ {
		publicKey := []byte{}
		for j := 0; j < 32; j++ {
			n := rand.Int() % 32
			publicKey = append(publicKey, byte(n))
		}
		publicKeys = append(publicKeys, publicKey)
	}
	return publicKeys
}

func generateTxHash(limit int) []common.Hash {
	txHashes := []common.Hash{}
	for i := 0; i < limit; i++ {
		temp := []byte{}
		for j := 0; j < 32; j++ {
			n := rand.Int() % 32
			temp = append(temp, byte(n))
		}
		h := common.BytesToHash(temp)
		txHashes = append(txHashes, h)
	}
	return txHashes
}

func resetDatabaseTx() {
	dbSPath, err := ioutil.TempDir(os.TempDir(), "test_rawdbv2")
	if err != nil {
		panic(err)
	}
	dbTx, err = incdb.Open("leveldb", dbSPath)
	if err != nil {
		panic(err)
	}
}

func TestStoreTxByPublicKey(t *testing.T) {
	publicKey1s := generatePublicKey(5)
	shardID0 := byte(0)
	for _, publicKey := range publicKey1s {
		txHashes := generateTxHash(10)
		for _, txHash := range txHashes {
			err := rawdbv2.StoreTxByPublicKey(dbTx, publicKey, txHash, shardID0)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestGetTxByPublicKey(t *testing.T) {
	publicKey1s := generatePublicKey(5)
	m := make(map[string][]common.Hash)
	shardID0 := byte(0)
	for _, publicKey := range publicKey1s {
		txHashes := generateTxHash(10)
		m[string(publicKey)] = txHashes
		for _, txHash := range txHashes {
			err := rawdbv2.StoreTxByPublicKey(dbTx, publicKey, txHash, shardID0)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	for k, want := range m {
		publicKey := []byte(k)
		sort.Slice(want, func(i, j int) bool {
			return want[i].String() < want[j].String()
		})
		res, err := rawdbv2.GetTxByPublicKey(dbTx, publicKey)
		if err != nil {
			t.Fatal(err)
		}
		got, ok := res[shardID0]
		if !ok {
			t.Fatalf("want shard %+v but got none", shardID0)
		}
		sort.Slice(got, func(i, j int) bool {
			return got[i].String() < got[j].String()
		})
		if len(want) != len(got) {
			t.Fatalf("want number of txHash %+v but got %+v", len(want), len(got))
		}
		for i := 0; i < len(want); i++ {
			if want[i] != got[i] {
				t.Fatalf("want txHash %+v but got %+v", want[i], got[i])
			}
		}
	}
}
