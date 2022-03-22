package statedb

import (
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/incdb"
	"math/rand"
	"os"
	"sort"
	"testing"
)

func IndexByteSlice(b []byte, bs [][]byte) int {
	for i, v := range bs {
		if bytes.Equal(b, v) {
			return i
		}
	}
	return -1
}
func TestLiteStateDBIterator(t *testing.T) {
	os.RemoveAll("./tmp")
	db, err := incdb.Open("leveldb", "./tmp")
	if err != nil {
		t.Error(err)
	}

	for testID := 0; testID < 10000; testID++ {

		keys := [][]byte{}
		values := [][]byte{}

		var randKey [][]byte
		var randValue [][]byte
		PREFIX := fmt.Sprintf("%v-", testID)
		rand.Seed(int64(testID))

		for i := 0; i < 20; i++ {
			k, v := genRandomKV()
			randKey = append(randKey, append([]byte(PREFIX), k.Bytes()...))
			randValue = append(randValue, v)
		}
		//fmt.Println("--------------------- DB")
		dbSize := 0
		for i := 0; i < 5; i++ {
			dbSize++
			db.Put(append([]byte("abc"), randKey[i]...), randValue[i])
			//fmt.Println(IndexByteSlice(randKey[i], randKey), randKey[i])
		}

		//fmt.Println("--------------------- Test 1: total different key")
		kvMap := map[string][]byte{}
		for i := 0; i < 5; i++ {
			keys = append(keys, randKey[dbSize+i])
			values = append(values, randValue[dbSize+i])
			//fmt.Println(IndexByteSlice(randKey[dbSize+i], randKey), randKey[dbSize+i])
			kvMap[string(randKey[dbSize+i])] = randValue[dbSize+i]
		}

		//fmt.Println("--------------------- Test 1: result")
		expectResult := append([][]byte{}, randKey[:dbSize+len(kvMap)]...)

		sort.Slice(expectResult, func(i, j int) bool {
			for index := range randKey[i] {
				if expectResult[i][index] < expectResult[j][index] {
					return true
				}
				if expectResult[i][index] > expectResult[j][index] {
					return false
				}
			}
			return false
		})
		iter := NewLiteStateDBIterator(db, []byte("abc"), []byte(PREFIX), kvMap)
		cnt := 0
		for iter.Next() {
			//fmt.Println(IndexByteSlice(iter.Key(), randKey), iter.Key())
			if !bytes.Equal(iter.Key(), expectResult[cnt]) {
				t.Error("Not expected at test " + PREFIX)
				t.Error("Expect", expectResult[cnt])
				t.Error("Get", iter.Key())
				t.FailNow()
			}
			cnt++
		}
	}

}
