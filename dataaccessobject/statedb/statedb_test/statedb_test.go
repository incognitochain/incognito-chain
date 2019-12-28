package statedb_test

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	_ "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/trie"
)

var (
	warperDBStatedbTest statedb.DatabaseAccessWarper
	emptyRoot           = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
	prefixA             = "serialnumber"
	prefixB             = "serialnumberderivator"
	prefixC             = "serial"
	prefixD             = "commitment"
	prefixE             = "outputcoin"
	keysA               = []common.Hash{}
	keysB               = []common.Hash{}
	keysC               = []common.Hash{}
	keysD               = []common.Hash{}
	keysE               = []common.Hash{}
	valuesA             = [][]byte{}
	valuesB             = [][]byte{}
	valuesC             = [][]byte{}
	valuesD             = [][]byte{}
	valuesE             = [][]byte{}

	limit100000 = 100000
	limit10000  = 10000
	limit1000   = 1000
	limit100    = 100
	limit1      = 1
)
var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest = statedb.NewDatabaseAccessWarper(diskBD)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

type args struct {
	prefix []byte
}
type test struct {
	name      string
	args      args
	wantKey   map[common.Hash]bool
	wantValue map[string]bool
}

func generateKeyValuePairWithPrefix(limit int, prefix []byte) ([]common.Hash, [][]byte) {
	keys := []common.Hash{}
	values := [][]byte{}
	for i := 0; i < limit; i++ {
		value := []byte{}
		for j := 0; j < 32; j++ {
			b := rand.Int() % 256
			value = append(value, byte(b))
		}
		temp := common.HashH(value)
		key := common.BytesToHash(append(prefix, temp[:][len(prefix):]...))
		keys = append(keys, key)
		values = append(values, value)
	}
	return keys, values
}

func TestStateDB_GetTestObjectByPrefix50000(t *testing.T) {
	rootHash, tests := createAndStoreDataForTesting(limit10000)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBStatedbTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	keys, values := tempStateDB.GetAllTestObjectList()
	if len(keys) != limit10000*5 {
		t.Fatalf("number of all keys want = %+v but got = %+v", limit10000*5, len(keys))
	}
	if len(values) != limit10000*5 {
		t.Fatalf("number of all values want = %+v but got = %+v", limit10000*5, len(values))
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKeys, gotValues := tempStateDB.GetByPrefixTestObjectList(tt.args.prefix)
			if len(gotKeys) != len(tt.wantKey) {

				t.Errorf("GetByPrefixSerialNumberList() gotKey length = %v, wantKey length = %v", len(gotKeys), len(tt.wantKey))
			}
			for _, gotKey := range gotKeys {
				if _, ok := tt.wantKey[gotKey]; !ok {
					t.Logf("Got Key to Bytes %+v \n with prefix %+v", keybytesToHex(gotKey[:]), keybytesToHex(tt.args.prefix))
					t.Errorf("GetByPrefixSerialNumberList() gotKey = %v but not wanted", gotKey)
				}
			}
			if len(gotValues) != len(tt.wantValue) {
				t.Errorf("GetByPrefixSerialNumberList() gotValue length = %v, wantValues length = %v", len(gotValues), len(tt.wantValue))
			}
			for _, gotValue := range gotValues {
				if _, ok := tt.wantValue[string(gotValue)]; !ok {
					t.Errorf("GetByPrefixSerialNumberList() gotValue = %v but not wanted", gotValue)
				}
			}

		})
	}
}

func BenchmarkStateDB_NewWithPrefixTrie20000(b *testing.B) {
	rootHash, _ := createAndStoreDataForTesting(4000)
	for n := 0; n < b.N; n++ {
		statedb.NewWithPrefixTrie(rootHash, warperDBStatedbTest)
	}
}

func BenchmarkStateDB_GetAllTestObjectList500000(b *testing.B) {
	rootHash, _ := createAndStoreDataForTesting(limit100000)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBStatedbTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.GetAllTestObjectList()
	}
}
func BenchmarkStateDB_GetAllTestObjectList50000(b *testing.B) {
	rootHash, _ := createAndStoreDataForTesting(limit10000)
	for n := 0; n < b.N; n++ {
		tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBStatedbTest)
		if err != nil || tempStateDB == nil {
			panic(err)
		}
		tempStateDB.GetAllTestObjectList()
	}
}
func BenchmarkStateDB_GetAllTestObjectList5000(b *testing.B) {
	rootHash, _ := createAndStoreDataForTesting(limit1000)
	for n := 0; n < b.N; n++ {
		tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBStatedbTest)
		if err != nil || tempStateDB == nil {
			panic(err)
		}
		tempStateDB.GetAllTestObjectList()
	}
}
func BenchmarkStateDB_GetAllTestObjectList500(b *testing.B) {
	rootHash, _ := createAndStoreDataForTesting(limit100)
	for n := 0; n < b.N; n++ {
		tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBStatedbTest)
		if err != nil || tempStateDB == nil {
			panic(err)
		}
		tempStateDB.GetAllTestObjectList()
	}
}
func BenchmarkStateDB_GetAllTestObjectList5(b *testing.B) {
	rootHash, _ := createAndStoreDataForTesting(limit1)
	for n := 0; n < b.N; n++ {
		tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBStatedbTest)
		if err != nil || tempStateDB == nil {
			panic(err)
		}
		tempStateDB.GetAllTestObjectList()
	}
}

func BenchmarkStateDB_GetTestObject500000(b *testing.B) {
	var sampleKey common.Hash
	rootHash, m := createAndStoreDataForTesting(limit100000)
	for key, _ := range m[0].wantKey {
		sampleKey = key
		break
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBStatedbTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.GetTestObject(sampleKey)
	}
}
func BenchmarkStateDB_GetTestObject50000(b *testing.B) {
	var sampleKey common.Hash
	rootHash, m := createAndStoreDataForTesting(limit10000)
	for key, _ := range m[0].wantKey {
		sampleKey = key
		break
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBStatedbTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.GetTestObject(sampleKey)
	}
}
func BenchmarkStateDB_GetTestObject5000(b *testing.B) {
	var sampleKey common.Hash
	rootHash, m := createAndStoreDataForTesting(limit1000)
	for key, _ := range m[0].wantKey {
		sampleKey = key
		break
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBStatedbTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.GetTestObject(sampleKey)
	}
}

func BenchmarkStateDB_GetByPrefixTestObjectList50000(b *testing.B) {
	rootHash, tests := createAndStoreDataForTesting(limit10000)
	for n := 0; n < b.N; n++ {
		tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBStatedbTest)
		if err != nil || tempStateDB == nil {
			panic(err)
		}
		tempStateDB.GetByPrefixTestObjectList(tests[0].args.prefix)
	}
}

func BenchmarkStateDB_GetByPrefixTestObjectList5000(b *testing.B) {
	rootHash, tests := createAndStoreDataForTesting(limit1000)
	for n := 0; n < b.N; n++ {
		tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBStatedbTest)
		if err != nil || tempStateDB == nil {
			panic(err)
		}
		tempStateDB.GetByPrefixTestObjectList(tests[0].args.prefix)
	}
}

func BenchmarkStateDB_GetByPrefixTestObjectList500(b *testing.B) {
	rootHash, tests := createAndStoreDataForTesting(limit100)
	for n := 0; n < b.N; n++ {
		tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBStatedbTest)
		if err != nil || tempStateDB == nil {
			panic(err)
		}
		tempStateDB.GetByPrefixTestObjectList(tests[0].args.prefix)
	}
}

func createAndStoreDataForTesting(limit int) (common.Hash, []test) {
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	if err != nil {
		panic(err)
	}
	tests := []test{
		{
			name: prefixA,
			args: args{
				prefix: []byte(prefixA),
			},
			wantKey:   make(map[common.Hash]bool),
			wantValue: make(map[string]bool),
		},
		{
			name: prefixB,
			args: args{
				prefix: []byte(prefixB),
			},
			wantKey:   make(map[common.Hash]bool),
			wantValue: make(map[string]bool),
		},
		{
			name: prefixC,
			args: args{
				prefix: []byte(prefixC),
			},
			wantKey:   make(map[common.Hash]bool),
			wantValue: make(map[string]bool),
		},
		{
			name: prefixD,
			args: args{
				prefix: []byte(prefixD),
			},
			wantKey:   make(map[common.Hash]bool),
			wantValue: make(map[string]bool),
		},
		{
			name: prefixE,
			args: args{
				prefix: []byte(prefixE),
			},
			wantKey:   make(map[common.Hash]bool),
			wantValue: make(map[string]bool),
		},
	}
	keysA, valuesA = generateKeyValuePairWithPrefix(limit, []byte(prefixA))
	keysB, valuesB = generateKeyValuePairWithPrefix(limit, []byte(prefixB))
	keysC, valuesC = generateKeyValuePairWithPrefix(limit, []byte(prefixC))
	keysD, valuesD = generateKeyValuePairWithPrefix(limit, []byte(prefixD))
	keysE, valuesE = generateKeyValuePairWithPrefix(limit, []byte(prefixE))
	for i := 0; i < len(keysA); i++ {
		sDB.SetStateObject(statedb.TestObjectType, keysA[i], valuesA[i])
	}
	for i := 0; i < len(keysB); i++ {
		sDB.SetStateObject(statedb.TestObjectType, keysB[i], valuesB[i])
	}
	for i := 0; i < len(keysC); i++ {
		sDB.SetStateObject(statedb.TestObjectType, keysC[i], valuesC[i])
	}
	for i := 0; i < len(keysD); i++ {
		sDB.SetStateObject(statedb.TestObjectType, keysD[i], valuesD[i])
	}
	for i := 0; i < len(keysE); i++ {
		sDB.SetStateObject(statedb.TestObjectType, keysE[i], valuesE[i])
	}
	for _, tt := range tests {
		if tt.name == prefixA {
			for _, key := range keysA {
				tt.wantKey[key] = true
			}
			for _, value := range valuesA {
				tt.wantValue[string(value)] = true
			}
			for _, key := range keysB {
				tt.wantKey[key] = true
			}
			for _, value := range valuesB {
				tt.wantValue[string(value)] = true
			}
		}
		if tt.name == prefixB {
			for _, key := range keysB {
				tt.wantKey[key] = true
			}
			for _, value := range valuesB {
				tt.wantValue[string(value)] = true
			}
		}
		if tt.name == prefixC {
			for _, key := range keysA {
				tt.wantKey[key] = true
			}
			for _, value := range valuesA {
				tt.wantValue[string(value)] = true
			}
			for _, key := range keysB {
				tt.wantKey[key] = true
			}
			for _, value := range valuesB {
				tt.wantValue[string(value)] = true
			}
			for _, key := range keysC {
				tt.wantKey[key] = true
			}
			for _, value := range valuesC {
				tt.wantValue[string(value)] = true
			}
			//for _, key := range keysD {
			//	tt.wantKey[key] = true
			//}
			//for _, value := range valuesD {
			//	tt.wantValue[string(value)] = true
			//}
		}
		if tt.name == prefixD {
			for _, key := range keysD {
				tt.wantKey[key] = true
			}
			for _, value := range valuesD {
				tt.wantValue[string(value)] = true
			}
		}
		if tt.name == prefixE {
			for _, key := range keysE {
				tt.wantKey[key] = true
			}
			for _, value := range valuesE {
				tt.wantValue[string(value)] = true
			}
		}
	}
	rootHash, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	if bytes.Compare(rootHash.Bytes(), emptyRoot.Bytes()) == 0 {
		panic("root hash is empty")
	}
	err = warperDBStatedbTest.TrieDB().Commit(rootHash, false)
	if err != nil {
		panic(err)
	}

	return rootHash, tests
}
func keybytesToHex(str []byte) []byte {
	l := len(str)*2 + 1
	var nibbles = make([]byte, l)
	for i, b := range str {
		nibbles[i*2] = b / 16
		nibbles[i*2+1] = b % 16
	}
	nibbles[l-1] = 16
	return nibbles
}

// hexToKeybytes turns hex nibbles into key bytes.
// This can only be used for keys of even length.
func hexToKeybytes(hex []byte) []byte {
	if hasTerm(hex) {
		hex = hex[:len(hex)-1]
	}
	if len(hex)&1 != 0 {
		panic("can't convert hex key of odd length")
	}
	key := make([]byte, len(hex)/2)
	decodeNibbles(hex, key)
	return key
}

func decodeNibbles(nibbles []byte, bytes []byte) {
	for bi, ni := 0, 0; ni < len(nibbles); bi, ni = bi+1, ni+2 {
		bytes[bi] = nibbles[ni]<<4 | nibbles[ni+1]
	}
}

// prefixLen returns the length of the common prefix of a and b.
func prefixLen(a, b []byte) int {
	var i, length = 0, len(a)
	if len(b) < length {
		length = len(b)
	}
	for ; i < length; i++ {
		if a[i] != b[i] {
			break
		}
	}
	return i
}

// hasTerm returns whether a hex key has the terminator flag.
func hasTerm(s []byte) bool {
	return len(s) > 0 && s[len(s)-1] == 16
}
