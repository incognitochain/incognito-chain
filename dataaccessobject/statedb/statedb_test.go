package statedb

import (
	"bytes"
	"github.com/incognitochain/incognito-chain/config"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
)

var (
	warperDBStatedbTest DatabaseAccessWarper
	emptyRoot           = common.HexToHash(common.HexEmptyRoot)
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
	warperDBStatedbTest = NewDatabaseAccessWarper(diskBD)
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

func TestFlatFileSerialize(t *testing.T) {

	stateDB := &StateDB{}

	objTestCase1, _ := newTestObjectWithValue(stateDB, common.HashH([]byte{0}), []byte{0, 1, 2, 3})
	objTestCase2, _ := newTestObjectWithValue(stateDB, common.HashH([]byte{0}), []byte{0, 1, 2, 3})
	objTestCase2.deleted = true

	type args struct {
		sob StateObject
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "test_state_object_delete_false",
			args: args{
				sob: objTestCase1,
			},
			want: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 93, 83, 70, 159, 32, 254, 244, 248, 234, 181, 43, 136, 4, 78, 222, 105, 199, 122, 106, 104, 166, 7, 40, 96, 159, 196, 166, 95, 245, 49, 231, 208, 0, 1, 2, 3},
		},
		{
			name: "test_state_object_delete_true",
			args: args{
				sob: objTestCase2,
			},
			want: []byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 93, 83, 70, 159, 32, 254, 244, 248, 234, 181, 43, 136, 4, 78, 222, 105, 199, 122, 106, 104, 166, 7, 40, 96, 159, 196, 166, 95, 245, 49, 231, 208, 0, 1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ByteSerialize(tt.args.sob); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ByteSerialize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlatFileDeSerialize(t *testing.T) {

	stateDB := &StateDB{}

	objTestCase1, _ := newTestObjectWithValue(stateDB, common.HashH([]byte{0}), []byte{0, 1, 2, 3})
	objTestCase2, _ := newTestObjectWithValue(stateDB, common.HashH([]byte{0}), []byte{0, 1, 2, 3})
	objTestCase2.deleted = true

	type args struct {
		stateDB *StateDB
		sobByte []byte
	}
	tests := []struct {
		name    string
		args    args
		want    StateObject
		wantErr bool
	}{
		{
			name: "test_state_object_delete_false",
			args: args{
				stateDB,
				[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 93, 83, 70, 159, 32, 254, 244, 248, 234, 181, 43, 136, 4, 78, 222, 105, 199, 122, 106, 104, 166, 7, 40, 96, 159, 196, 166, 95, 245, 49, 231, 208, 0, 1, 2, 3},
			},
			want: objTestCase1,
		},
		{
			name: "test_state_object_delete_true",
			args: args{
				stateDB,
				[]byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 93, 83, 70, 159, 32, 254, 244, 248, 234, 181, 43, 136, 4, 78, 222, 105, 199, 122, 106, 104, 166, 7, 40, 96, 159, 196, 166, 95, 245, 49, 231, 208, 0, 1, 2, 3},
			},
			want: objTestCase2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ByteDeSerialize(tt.args.stateDB, tt.args.sobByte)
			if (err != nil) != tt.wantErr {
				t.Errorf("ByteDeSerialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ByteDeSerialize() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStateDB_DeleteNotExistObject(t *testing.T) {
	keys, values := generateKeyValuePairWithPrefix(5, []byte("abc"))
	stateDB, _ := NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	stateDB.SetStateObject(TestObjectType, keys[0], values[0])
	stateDB.SetStateObject(TestObjectType, keys[1], values[1])
	stateDB.SetStateObject(TestObjectType, keys[2], values[2])
	rootHash, _, _ := stateDB.Commit(true)
	stateDB.Database().TrieDB().Commit(rootHash, false, nil)

	v0, err := stateDB.getTestObject(keys[0])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v0, values[0]) {
		t.Fatalf("want values 0 %+v, got %+v", values[0], v0)
	}
	v1, err := stateDB.getTestObject(keys[1])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v1, values[1]) {
		t.Fatalf("want values 0 %+v, got %+v", values[1], v1)
	}
	v2, err := stateDB.getTestObject(keys[2])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v2, values[2]) {
		t.Fatalf("want values 0 %+v, got %+v", values[2], v2)
	}
	v3, err := stateDB.getTestObject(keys[3])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v3, []byte{}) {
		t.Fatalf("want values 0 %+v, got %+v", []byte{}, v3)
	}
	v4, err := stateDB.getTestObject(keys[4])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v4, []byte{}) {
		t.Fatalf("want values 0 %+v, got %+v", []byte{}, v4)
	}

	stateDB.SetStateObject(TestObjectType, keys[3], values[3])
	stateDB.SetStateObject(TestObjectType, keys[4], values[4])
	stateDB.MarkDeleteStateObject(TestObjectType, keys[0])
	stateDB.MarkDeleteStateObject(TestObjectType, keys[1])
	stateDB.MarkDeleteStateObject(TestObjectType, keys[2])
	stateDB.MarkDeleteStateObject(TestObjectType, keys[3])
	stateDB.MarkDeleteStateObject(TestObjectType, keys[4])

	rootHash2, _, _ := stateDB.Commit(true)
	stateDB.Database().TrieDB().Commit(rootHash2, false, nil)

	v0, err = stateDB.getTestObject(keys[0])
	v1, err = stateDB.getTestObject(keys[1])
	v2, err = stateDB.getTestObject(keys[2])
	v3, err = stateDB.getTestObject(keys[3])
	v4, err = stateDB.getTestObject(keys[4])
	if !reflect.DeepEqual(v0, []byte{}) {
		t.Fatalf("want values 0 %+v, got %+v", []byte{}, v0)
	}
	if !reflect.DeepEqual(v1, []byte{}) {
		t.Fatalf("want values 0 %+v, got %+v", []byte{}, v1)
	}
	if !reflect.DeepEqual(v2, []byte{}) {
		t.Fatalf("want values 0 %+v, got %+v", []byte{}, v2)
	}
	if !reflect.DeepEqual(v3, []byte{}) {
		t.Fatalf("want values 0 %+v, got %+v", []byte{}, v3)
	}
	if !reflect.DeepEqual(v4, []byte{}) {
		t.Fatalf("want values 0 %+v, got %+v", []byte{}, v4)
	}
}

func TestStateDB_GetTestObjectByPrefix50000(t *testing.T) {
	rootHash, tests := createAndStoreDataForTesting(limit10000)
	tempStateDB, err := NewWithPrefixTrie(rootHash, warperDBStatedbTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	keys, values := tempStateDB.getAllTestObjectList()
	if len(keys) != limit10000*5 {
		t.Fatalf("number of all keys want = %+v but got = %+v", limit10000*5, len(keys))
	}
	if len(values) != limit10000*5 {
		t.Fatalf("number of all values want = %+v but got = %+v", limit10000*5, len(values))
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKeys, gotValues := tempStateDB.getByPrefixTestObjectList(tt.args.prefix)
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
	rootHash, _ := createAndStoreDataForTesting(limit100000)
	sDB, _ := NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
	for n := 0; n < b.N; n++ {
		sDB.Reset(rootHash)
	}
}

func BenchmarkStateDB_GetAllTestObjectList500000(b *testing.B) {
	rootHash, _ := createAndStoreDataForTesting(limit100000)
	tempStateDB, err := NewWithPrefixTrie(rootHash, warperDBStatedbTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.getAllTestObjectList()
	}
}
func BenchmarkStateDB_GetAllTestObjectList50000(b *testing.B) {
	rootHash, _ := createAndStoreDataForTesting(limit10000)
	for n := 0; n < b.N; n++ {
		tempStateDB, err := NewWithPrefixTrie(rootHash, warperDBStatedbTest)
		if err != nil || tempStateDB == nil {
			panic(err)
		}
		tempStateDB.getAllTestObjectList()
	}
}
func BenchmarkStateDB_GetAllTestObjectList5000(b *testing.B) {
	rootHash, _ := createAndStoreDataForTesting(limit1000)
	for n := 0; n < b.N; n++ {
		tempStateDB, err := NewWithPrefixTrie(rootHash, warperDBStatedbTest)
		if err != nil || tempStateDB == nil {
			panic(err)
		}
		tempStateDB.getAllTestObjectList()
	}
}
func BenchmarkStateDB_GetAllTestObjectList500(b *testing.B) {
	rootHash, _ := createAndStoreDataForTesting(limit100)
	for n := 0; n < b.N; n++ {
		tempStateDB, err := NewWithPrefixTrie(rootHash, warperDBStatedbTest)
		if err != nil || tempStateDB == nil {
			panic(err)
		}
		tempStateDB.getAllTestObjectList()
	}
}
func BenchmarkStateDB_GetAllTestObjectList5(b *testing.B) {
	rootHash, _ := createAndStoreDataForTesting(limit1)
	for n := 0; n < b.N; n++ {
		tempStateDB, err := NewWithPrefixTrie(rootHash, warperDBStatedbTest)
		if err != nil || tempStateDB == nil {
			panic(err)
		}
		tempStateDB.getAllTestObjectList()
	}
}

func BenchmarkStateDB_GetTestObject500000(b *testing.B) {
	var sampleKey common.Hash
	rootHash, m := createAndStoreDataForTesting(limit100000)
	for key := range m[0].wantKey {
		sampleKey = key
		break
	}
	tempStateDB, err := NewWithPrefixTrie(rootHash, warperDBStatedbTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.getTestObject(sampleKey)
	}
}
func BenchmarkStateDB_GetTestObject50000(b *testing.B) {
	var sampleKey common.Hash
	rootHash, m := createAndStoreDataForTesting(limit10000)
	for key := range m[0].wantKey {
		sampleKey = key
		break
	}
	tempStateDB, err := NewWithPrefixTrie(rootHash, warperDBStatedbTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.getTestObject(sampleKey)
	}
}
func BenchmarkStateDB_GetTestObject5000(b *testing.B) {
	var sampleKey common.Hash
	rootHash, m := createAndStoreDataForTesting(limit1000)
	for key := range m[0].wantKey {
		sampleKey = key
		break
	}
	tempStateDB, err := NewWithPrefixTrie(rootHash, warperDBStatedbTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.getTestObject(sampleKey)
	}
}

func BenchmarkStateDB_GetByPrefixTestObjectList50000(b *testing.B) {
	rootHash, tests := createAndStoreDataForTesting(limit10000)
	for n := 0; n < b.N; n++ {
		tempStateDB, err := NewWithPrefixTrie(rootHash, warperDBStatedbTest)
		if err != nil || tempStateDB == nil {
			panic(err)
		}
		tempStateDB.getByPrefixTestObjectList(tests[0].args.prefix)
	}
}

func BenchmarkStateDB_GetByPrefixTestObjectList5000(b *testing.B) {
	rootHash, tests := createAndStoreDataForTesting(limit1000)
	for n := 0; n < b.N; n++ {
		tempStateDB, err := NewWithPrefixTrie(rootHash, warperDBStatedbTest)
		if err != nil || tempStateDB == nil {
			panic(err)
		}
		tempStateDB.getByPrefixTestObjectList(tests[0].args.prefix)
	}
}

func BenchmarkStateDB_GetByPrefixTestObjectList500(b *testing.B) {
	rootHash, tests := createAndStoreDataForTesting(limit100)
	for n := 0; n < b.N; n++ {
		tempStateDB, err := NewWithPrefixTrie(rootHash, warperDBStatedbTest)
		if err != nil || tempStateDB == nil {
			panic(err)
		}
		tempStateDB.getByPrefixTestObjectList(tests[0].args.prefix)
	}
}

func createAndStoreDataForTesting(limit int) (common.Hash, []test) {
	sDB, err := NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)
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
		sDB.SetStateObject(TestObjectType, keysA[i], valuesA[i])
	}
	for i := 0; i < len(keysB); i++ {
		sDB.SetStateObject(TestObjectType, keysB[i], valuesB[i])
	}
	for i := 0; i < len(keysC); i++ {
		sDB.SetStateObject(TestObjectType, keysC[i], valuesC[i])
	}
	for i := 0; i < len(keysD); i++ {
		sDB.SetStateObject(TestObjectType, keysD[i], valuesD[i])
	}
	for i := 0; i < len(keysE); i++ {
		sDB.SetStateObject(TestObjectType, keysE[i], valuesE[i])
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
	rootHash, _, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	if bytes.Compare(rootHash.Bytes(), emptyRoot.Bytes()) == 0 {
		panic("root hash is empty")
	}
	err = warperDBStatedbTest.TrieDB().Commit(rootHash, false, nil)
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

//
//func Test_BatchStateDB(t *testing.T) {
//	config.LoadConfig()
//	config.LoadParam()
//
//	//init DB and txDB
//	os.RemoveAll("./tmp")
//	db, err := incdb.Open("leveldb", "./tmp")
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	//generate data
//	var randKey []common.Hash
//	var randValue [][]byte
//	rand.Seed(1)
//	for i := 0; i < 100; i++ {
//		k, v := genRandomKV()
//		randKey = append(randKey, k)
//		randValue = append(randValue, v)
//	}
//	dbName := "test"
//	stateDB, err := NewWithMode(dbName, common.STATEDB_BATCH_COMMIT_MODE, db, *NewEmptyRebuildInfo(""), nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//	// check set & get object
//	stateDB.getOrNewStateObjectWithValue(TestObjectType, randKey[0], randValue[0])
//	getData, _ := stateDB.getTestObject(randKey[0])
//	if !bytes.Equal(getData, randValue[0]) { // must return equal
//		t.Fatal(errors.New("Cannot store live object to newTxDB"))
//	}
//	stateDB.getOrNewStateObjectWithValue(TestObjectType, randKey[1], randValue[1])
//	getData, _ = stateDB.getTestObject(randKey[1])
//	if !bytes.Equal(getData, randValue[1]) { // must return equal
//		t.Fatal(errors.New("Cannot store live object to newTxDB"))
//	}
//
//	//clone new txDB: must remove old live state
//	newTxDB := stateDB.Copy()
//	getData, _ = newTxDB.getTestObject(randKey[0])
//	if len(getData) != 0 { // must return empty
//		t.Fatal(errors.New("Copy stateDB but data of other live state still exist"))
//	}
//
//	newTxDB.getOrNewStateObjectWithValue(TestObjectType, randKey[1], randValue[1])
//	getData, _ = newTxDB.getTestObject(randKey[1])
//	if !bytes.Equal(getData, randValue[1]) { // must return equal
//		t.Fatal(errors.New("Cannot store live object to newTxDB"))
//	}
//
//	for i := 0; i < 10; i++ {
//		newTxDB.getOrNewStateObjectWithValue(TestObjectType, randKey[i+1], randValue[i+1])
//		agg, rebuildRoot, err := newTxDB.Commit(true)
//		if err != nil {
//			t.Fatal(err)
//		}
//		fmt.Println(agg, rebuildRoot)
//	}
//	//double commit
//	newAgg, newAggRoot, _ := newTxDB.Commit(true)
//	fmt.Println("double commit", newAgg, newAggRoot)
//
//	newTxDB.getOrNewStateObjectWithValue(TestObjectType, randKey[12], randValue[12])
//	_, rebuildRoot, err := newTxDB.Commit(true)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	newTxDB.Finalized(true, *newAggRoot)
//	pivotCommit, _ := GetLatestPivotCommit(db, dbName)
//	if strings.Index(pivotCommit, newAgg.String()) != 0 {
//		fmt.Println(pivotCommit, newAgg.String())
//		t.Fatal(errors.New("Store wrong pivot point"))
//	}
//
//	fmt.Println(rebuildRoot)
//	newStateDB, err := NewWithMode(dbName, common.STATEDB_BATCH_COMMIT_MODE, db, *rebuildRoot, nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//	getData, _ = newStateDB.getTestObject(randKey[12])
//	if !bytes.Equal(getData, randValue[12]) { // must return equal
//		t.Fatal(errors.New("Cannot store live object to newStateDB"))
//	}
//}

func TestBatchCommitFinalizeNoFFRebuild(t *testing.T) {

	config.LoadConfig()
	config.LoadParam()

	//generate data
	var randKey []common.Hash
	var randValue [][]byte
	rand.Seed(1)
	for i := 0; i < 100; i++ {
		k, v := genRandomKV()
		randKey = append(randKey, k)
		randValue = append(randValue, v)
	}
	testObjects := []StateObject{}
	//init DB and txDB
	os.RemoveAll("./tmp")
	db, err := incdb.Open("leveldb", "./tmp")
	if err != nil {
		t.Fatal(err)
	}

	dbName := "test"

	stateDB, err := NewWithMode(dbName, common.STATEDB_BATCH_COMMIT_MODE, db, *NewEmptyRebuildInfo(""), nil)
	if err != nil {
		t.Fatal(err)
	}
	sObj1, _ := newTestObjectWithValue(stateDB, randKey[0], randValue[0])
	sObj2, _ := newTestObjectWithValue(stateDB, randKey[1], randValue[1])
	sObj3, _ := newTestObjectWithValue(stateDB, randKey[2], randValue[2])
	testObjects = append(testObjects, sObj1)
	testObjects = append(testObjects, sObj2)
	testObjects = append(testObjects, sObj3)

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[0].GetHash(), testObjects[0].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[1].GetHash(), testObjects[1].GetValue()); err != nil {
		t.Fatal(err)
	}

	rootHash, rebuildInfo1, err := stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	if err := stateDB.Finalized(true, *rebuildInfo1); err != nil {
		t.Fatal(err)
	}

	rebuildInfo2 := RebuildInfo{
		mode:            common.STATEDB_BATCH_COMMIT_MODE,
		rebuildRootHash: rootHash,
		pivotRootHash:   rootHash,
		rebuildFFIndex:  rebuildInfo1.rebuildFFIndex,
		pivotFFIndex:    rebuildInfo1.pivotFFIndex,
	}
	stateDB2, err := NewWithMode(dbName, common.STATEDB_BATCH_COMMIT_MODE, db, rebuildInfo2, nil)
	if v0, err := stateDB2.getStateObject(TestObjectType, testObjects[0].GetHash()); err != nil {
		t.Fatal(err)
	} else {
		if !bytes.Equal(v0.GetValueBytes(), testObjects[0].GetValueBytes()) {
			t.Fatalf("expect %+v, got %+v", testObjects[0].GetValueBytes(), v0.GetValueBytes())
		}
	}

	if v1, err := stateDB2.getStateObject(TestObjectType, testObjects[1].GetHash()); err != nil {
		t.Fatal(err)
	} else {
		if !bytes.Equal(v1.GetValueBytes(), testObjects[1].GetValueBytes()) {
			t.Fatalf("expect %+v, got %+v", testObjects[1].GetValueBytes(), v1.GetValueBytes())
		}
	}

	if v2, err := stateDB2.getStateObject(TestObjectType, testObjects[2].GetHash()); err != nil {
		t.Fatal("expect err but got ", err)
	} else if v2 != nil {
		t.Fatal("expect nil but got ", v2)
	}

}

func TestBatchCommitNoFinalizeFFRebuild(t *testing.T) {

	config.LoadConfig()
	config.LoadParam()

	//generate data
	var randKey []common.Hash
	var randValue [][]byte
	rand.Seed(1)
	for i := 0; i < 100; i++ {
		k, v := genRandomKV()
		randKey = append(randKey, k)
		randValue = append(randValue, v)
	}
	testObjects := []StateObject{}
	//init DB and txDB
	os.RemoveAll("./tmp")
	db, err := incdb.Open("leveldb", "./tmp")
	if err != nil {
		t.Fatal(err)
	}

	dbName := "test"

	stateDB, err := NewWithMode(dbName, common.STATEDB_BATCH_COMMIT_MODE, db, *NewEmptyRebuildInfo(""), nil)
	if err != nil {
		t.Fatal(err)
	}
	sObj1, _ := newTestObjectWithValue(stateDB, randKey[0], randValue[0])
	sObj2, _ := newTestObjectWithValue(stateDB, randKey[1], randValue[1])
	sObj3, _ := newTestObjectWithValue(stateDB, randKey[2], randValue[2])
	testObjects = append(testObjects, sObj1)
	testObjects = append(testObjects, sObj2)
	testObjects = append(testObjects, sObj3)

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[0].GetHash(), testObjects[0].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[1].GetHash(), testObjects[1].GetValue()); err != nil {
		t.Fatal(err)
	}

	rootHash, rebuildInfo1, err := stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	//if err := stateDB.Finalized(true, *rebuildInfo1); err != nil {
	//	t.Fatal(err)
	//}

	rebuildInfo2 := RebuildInfo{
		mode:            common.STATEDB_BATCH_COMMIT_MODE,
		rebuildRootHash: rootHash,
		pivotRootHash:   emptyRoot,
		rebuildFFIndex:  rebuildInfo1.rebuildFFIndex,
		pivotFFIndex:    rebuildInfo1.pivotFFIndex,
	}
	stateDB2, err := NewWithMode(dbName, common.STATEDB_BATCH_COMMIT_MODE, db, rebuildInfo2, nil)
	if v0, err := stateDB2.getStateObject(TestObjectType, testObjects[0].GetHash()); err != nil {
		t.Fatal(err)
	} else {
		if !bytes.Equal(v0.GetValueBytes(), testObjects[0].GetValueBytes()) {
			t.Fatalf("expect %+v, got %+v", testObjects[0].GetValueBytes(), v0.GetValueBytes())
		}
	}

	if v1, err := stateDB2.getStateObject(TestObjectType, testObjects[1].GetHash()); err != nil {
		t.Fatal(err)
	} else {
		if !bytes.Equal(v1.GetValueBytes(), testObjects[1].GetValueBytes()) {
			t.Fatalf("expect %+v, got %+v", testObjects[1].GetValueBytes(), v1.GetValueBytes())
		}
	}

	if v2, err := stateDB2.getStateObject(TestObjectType, testObjects[2].GetHash()); err != nil {
		t.Fatal("expect err but got ", err)
	} else if v2 != nil {
		t.Fatal("expect nil but got ", v2)
	}

	rootHash2, _, err := stateDB2.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rootHash, rootHash2) {
		t.Fatalf("expect two root hash equal but 1 = %+v, 2 = %+v", rootHash, rootHash2)
	}
}

func TestBatchCommitFinalizeSwitchToArchive(t *testing.T) {

	config.LoadConfig()
	config.LoadParam()

	//generate data
	var randKey []common.Hash
	var randValue [][]byte
	rand.Seed(1)
	for i := 0; i < 100; i++ {
		k, v := genRandomKV()
		randKey = append(randKey, k)
		randValue = append(randValue, v)
	}
	testObjects := []StateObject{}
	//init DB and txDB
	os.RemoveAll("./tmp")
	db, err := incdb.Open("leveldb", "./tmp")
	if err != nil {
		t.Fatal(err)
	}

	dbName := "test"

	stateDB, err := NewWithMode(dbName, common.STATEDB_BATCH_COMMIT_MODE, db, *NewEmptyRebuildInfo(""), nil)
	if err != nil {
		t.Fatal(err)
	}
	sObj1, _ := newTestObjectWithValue(stateDB, randKey[0], randValue[0])
	sObj2, _ := newTestObjectWithValue(stateDB, randKey[1], randValue[1])
	sObj3, _ := newTestObjectWithValue(stateDB, randKey[2], randValue[2])
	testObjects = append(testObjects, sObj1)
	testObjects = append(testObjects, sObj2)
	testObjects = append(testObjects, sObj3)

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[0].GetHash(), testObjects[0].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[1].GetHash(), testObjects[1].GetValue()); err != nil {
		t.Fatal(err)
	}

	rootHash, rebuildInfo1, err := stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	if err := stateDB.Finalized(true, *rebuildInfo1); err != nil {
		t.Fatal(err)
	}

	rebuildInfo2 := RebuildInfo{
		mode:            common.STATEDB_BATCH_COMMIT_MODE,
		rebuildRootHash: rootHash,
		pivotRootHash:   rootHash,
		rebuildFFIndex:  rebuildInfo1.rebuildFFIndex,
		pivotFFIndex:    rebuildInfo1.pivotFFIndex,
	}
	stateDB2, err := NewWithMode(dbName, common.STATEDB_ARCHIVE_MODE, db, rebuildInfo2, nil)
	if v0, err := stateDB2.getStateObject(TestObjectType, testObjects[0].GetHash()); err != nil {
		t.Fatal(err)
	} else {
		if !bytes.Equal(v0.GetValueBytes(), testObjects[0].GetValueBytes()) {
			t.Fatalf("expect %+v, got %+v", testObjects[0].GetValueBytes(), v0.GetValueBytes())
		}
	}

	if v1, err := stateDB2.getStateObject(TestObjectType, testObjects[1].GetHash()); err != nil {
		t.Fatal(err)
	} else {
		if !bytes.Equal(v1.GetValueBytes(), testObjects[1].GetValueBytes()) {
			t.Fatalf("expect %+v, got %+v", testObjects[1].GetValueBytes(), v1.GetValueBytes())
		}
	}

	if v2, err := stateDB2.getStateObject(TestObjectType, testObjects[2].GetHash()); err != nil {
		t.Fatal("expect err but got ", err)
	} else if v2 != nil {
		t.Fatal("expect nil but got ", v2)
	}

}

func TestBatchCommitNoFinalizeSwitchToArchive(t *testing.T) {

	config.LoadConfig()
	config.LoadParam()

	//generate data
	var randKey []common.Hash
	var randValue [][]byte
	rand.Seed(1)
	for i := 0; i < 100; i++ {
		k, v := genRandomKV()
		randKey = append(randKey, k)
		randValue = append(randValue, v)
	}
	testObjects := []StateObject{}
	//init DB and txDB
	os.RemoveAll("./tmp")
	db, err := incdb.Open("leveldb", "./tmp")
	if err != nil {
		t.Fatal(err)
	}

	dbName := "test"

	stateDB, err := NewWithMode(dbName, common.STATEDB_BATCH_COMMIT_MODE, db, *NewEmptyRebuildInfo(""), nil)
	if err != nil {
		t.Fatal(err)
	}
	sObj1, _ := newTestObjectWithValue(stateDB, randKey[0], randValue[0])
	sObj2, _ := newTestObjectWithValue(stateDB, randKey[1], randValue[1])
	sObj3, _ := newTestObjectWithValue(stateDB, randKey[2], randValue[2])
	testObjects = append(testObjects, sObj1)
	testObjects = append(testObjects, sObj2)
	testObjects = append(testObjects, sObj3)

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[0].GetHash(), testObjects[0].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[1].GetHash(), testObjects[1].GetValue()); err != nil {
		t.Fatal(err)
	}

	rootHash, rebuildInfo1, err := stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	//if err := stateDB.Finalized(true, *rebuildInfo1); err != nil {
	//	t.Fatal(err)
	//}

	rebuildInfo2 := RebuildInfo{
		mode:            common.STATEDB_BATCH_COMMIT_MODE,
		rebuildRootHash: rootHash,
		pivotRootHash:   rootHash,
		rebuildFFIndex:  rebuildInfo1.rebuildFFIndex,
		pivotFFIndex:    rebuildInfo1.pivotFFIndex,
	}
	stateDB2, err := NewWithMode(dbName, common.STATEDB_ARCHIVE_MODE, db, rebuildInfo2, nil)
	if v0, err := stateDB2.getStateObject(TestObjectType, testObjects[0].GetHash()); err != nil {
		t.Fatal(err)
	} else {
		if !bytes.Equal(v0.GetValueBytes(), testObjects[0].GetValueBytes()) {
			t.Fatalf("expect %+v, got %+v", testObjects[0].GetValueBytes(), v0.GetValueBytes())
		}
	}

	if v1, err := stateDB2.getStateObject(TestObjectType, testObjects[1].GetHash()); err != nil {
		t.Fatal(err)
	} else {
		if !bytes.Equal(v1.GetValueBytes(), testObjects[1].GetValueBytes()) {
			t.Fatalf("expect %+v, got %+v", testObjects[1].GetValueBytes(), v1.GetValueBytes())
		}
	}

	if v2, err := stateDB2.getStateObject(TestObjectType, testObjects[2].GetHash()); err != nil {
		t.Fatal("expect err but got ", err)
	} else if v2 != nil {
		t.Fatal("expect nil but got ", v2)
	}

}

func TestMultipleBatchCommitNoFinalize(t *testing.T) {

	config.LoadConfig()
	config.LoadParam()

	//generate data
	var randKey []common.Hash
	var randValue [][]byte
	rand.Seed(1)
	for i := 0; i < 100; i++ {
		k, v := genRandomKV()
		randKey = append(randKey, k)
		randValue = append(randValue, v)
	}
	//init DB and txDB
	os.RemoveAll("./tmp")
	db, err := incdb.Open("leveldb", "./tmp")
	if err != nil {
		t.Fatal(err)
	}

	dbName := "test"

	stateDB, err := NewWithMode(dbName, common.STATEDB_BATCH_COMMIT_MODE, db, *NewEmptyRebuildInfo(""), nil)
	if err != nil {
		t.Fatal(err)
	}
	testObjects := make([]StateObject, 7)
	testObjects[0], _ = newTestObjectWithValue(stateDB, randKey[0], randValue[0])
	testObjects[1], _ = newTestObjectWithValue(stateDB, randKey[1], randValue[1])
	testObjects[2], _ = newTestObjectWithValue(stateDB, randKey[2], randValue[2])
	testObjects[3], _ = newTestObjectWithValue(stateDB, randKey[3], randValue[3])
	testObjects[4], _ = newTestObjectWithValue(stateDB, randKey[4], randValue[4])
	testObjects[5], _ = newTestObjectWithValue(stateDB, randKey[5], randValue[5])
	testObjects[6], _ = newTestObjectWithValue(stateDB, randKey[6], randValue[6])

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[0].GetHash(), testObjects[0].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[1].GetHash(), testObjects[1].GetValue()); err != nil {
		t.Fatal(err)
	}

	_, _, err = stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[2].GetHash(), testObjects[2].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[3].GetHash(), testObjects[3].GetValue()); err != nil {
		t.Fatal(err)
	}

	_, _, err = stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[4].GetHash(), testObjects[4].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[5].GetHash(), testObjects[5].GetValue()); err != nil {
		t.Fatal(err)
	}

	rootHash3, rebuildInfo3, err := stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	rebuildInfo4 := RebuildInfo{
		mode:            common.STATEDB_BATCH_COMMIT_MODE,
		rebuildRootHash: rootHash3,
		pivotRootHash:   emptyRoot,
		rebuildFFIndex:  rebuildInfo3.rebuildFFIndex,
		pivotFFIndex:    rebuildInfo3.pivotFFIndex,
	}
	stateDB2, err := NewWithMode(dbName, common.STATEDB_BATCH_COMMIT_MODE, db, rebuildInfo4, nil)
	for i := 0; i <= 5; i++ {
		if v, err := stateDB2.getStateObject(TestObjectType, testObjects[i].GetHash()); err != nil {
			t.Fatal(err)
		} else {
			if !bytes.Equal(v.GetValueBytes(), testObjects[i].GetValueBytes()) {
				t.Fatalf("expect %+v, got %+v", testObjects[i].GetValueBytes(), v.GetValueBytes())
			}
		}
	}

	for i := 6; i < len(testObjects); i++ {
		if v1, err := stateDB2.getStateObject(TestObjectType, testObjects[i].GetHash()); err != nil {
			t.Fatal(err)
		} else if v1 != nil {
			t.Fatalf("expect nil, but got %+v", v1)
		}
	}

}

func TestMultipleBatchCommitFinalize(t *testing.T) {

	config.LoadConfig()
	config.LoadParam()

	//generate data
	var randKey []common.Hash
	var randValue [][]byte
	rand.Seed(1)
	for i := 0; i < 100; i++ {
		k, v := genRandomKV()
		randKey = append(randKey, k)
		randValue = append(randValue, v)
	}
	//init DB and txDB
	os.RemoveAll("./tmp")
	db, err := incdb.Open("leveldb", "./tmp")
	if err != nil {
		t.Fatal(err)
	}

	dbName := "test"

	stateDB, err := NewWithMode(dbName, common.STATEDB_BATCH_COMMIT_MODE, db, *NewEmptyRebuildInfo(""), nil)
	if err != nil {
		t.Fatal(err)
	}
	testObjects := make([]StateObject, 7)
	testObjects[0], _ = newTestObjectWithValue(stateDB, randKey[0], randValue[0])
	testObjects[1], _ = newTestObjectWithValue(stateDB, randKey[1], randValue[1])
	testObjects[2], _ = newTestObjectWithValue(stateDB, randKey[2], randValue[2])
	testObjects[3], _ = newTestObjectWithValue(stateDB, randKey[3], randValue[3])
	testObjects[4], _ = newTestObjectWithValue(stateDB, randKey[4], randValue[4])
	testObjects[5], _ = newTestObjectWithValue(stateDB, randKey[5], randValue[5])
	testObjects[6], _ = newTestObjectWithValue(stateDB, randKey[6], randValue[6])

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[0].GetHash(), testObjects[0].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[1].GetHash(), testObjects[1].GetValue()); err != nil {
		t.Fatal(err)
	}

	_, _, err = stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[2].GetHash(), testObjects[2].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[3].GetHash(), testObjects[3].GetValue()); err != nil {
		t.Fatal(err)
	}

	_, _, err = stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[4].GetHash(), testObjects[4].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[5].GetHash(), testObjects[5].GetValue()); err != nil {
		t.Fatal(err)
	}

	rootHash3, rebuildInfo3, err := stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	if err := stateDB.Finalized(true, *NewEmptyRebuildInfo("")); err != nil {
		t.Fatal(err)
	}

	rebuildInfo4 := RebuildInfo{
		mode:            common.STATEDB_BATCH_COMMIT_MODE,
		rebuildRootHash: rootHash3,
		pivotRootHash:   rebuildInfo3.pivotRootHash,
		rebuildFFIndex:  rebuildInfo3.rebuildFFIndex,
		pivotFFIndex:    rebuildInfo3.pivotFFIndex,
	}
	stateDB2, err := NewWithMode(dbName, common.STATEDB_BATCH_COMMIT_MODE, db, rebuildInfo4, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= 5; i++ {
		if v, err := stateDB2.getStateObject(TestObjectType, testObjects[i].GetHash()); err != nil {
			t.Fatal(err)
		} else {
			if !bytes.Equal(v.GetValueBytes(), testObjects[i].GetValueBytes()) {
				t.Fatalf("expect %+v, got %+v", testObjects[i].GetValueBytes(), v.GetValueBytes())
			}
		}
	}

	for i := 6; i < len(testObjects); i++ {
		if v1, err := stateDB2.getStateObject(TestObjectType, testObjects[i].GetHash()); err != nil {
			t.Fatal(err)
		} else if v1 != nil {
			t.Fatalf("expect nil, but got %+v", v1)
		}
	}

}

func TestMultipleBatchCommitFinalizeSwitchToArchive(t *testing.T) {

	config.LoadConfig()
	config.LoadParam()

	//generate data
	var randKey []common.Hash
	var randValue [][]byte
	rand.Seed(1)
	for i := 0; i < 100; i++ {
		k, v := genRandomKV()
		randKey = append(randKey, k)
		randValue = append(randValue, v)
	}
	//init DB and txDB
	os.RemoveAll("./tmp")
	db, err := incdb.Open("leveldb", "./tmp")
	if err != nil {
		t.Fatal(err)
	}

	dbName := "test"

	stateDB, err := NewWithMode(dbName, common.STATEDB_BATCH_COMMIT_MODE, db, *NewEmptyRebuildInfo(""), nil)
	if err != nil {
		t.Fatal(err)
	}
	testObjects := make([]StateObject, 7)
	testObjects[0], _ = newTestObjectWithValue(stateDB, randKey[0], randValue[0])
	testObjects[1], _ = newTestObjectWithValue(stateDB, randKey[1], randValue[1])
	testObjects[2], _ = newTestObjectWithValue(stateDB, randKey[2], randValue[2])
	testObjects[3], _ = newTestObjectWithValue(stateDB, randKey[3], randValue[3])
	testObjects[4], _ = newTestObjectWithValue(stateDB, randKey[4], randValue[4])
	testObjects[5], _ = newTestObjectWithValue(stateDB, randKey[5], randValue[5])
	testObjects[6], _ = newTestObjectWithValue(stateDB, randKey[6], randValue[6])

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[0].GetHash(), testObjects[0].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[1].GetHash(), testObjects[1].GetValue()); err != nil {
		t.Fatal(err)
	}

	_, _, err = stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[2].GetHash(), testObjects[2].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[3].GetHash(), testObjects[3].GetValue()); err != nil {
		t.Fatal(err)
	}

	_, _, err = stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[4].GetHash(), testObjects[4].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[5].GetHash(), testObjects[5].GetValue()); err != nil {
		t.Fatal(err)
	}

	rootHash3, rebuildInfo3, err := stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	if err := stateDB.Finalized(true, *NewEmptyRebuildInfo("")); err != nil {
		t.Fatal(err)
	}

	rebuildInfo4 := RebuildInfo{
		mode:            common.STATEDB_BATCH_COMMIT_MODE,
		rebuildRootHash: rootHash3,
		pivotRootHash:   rebuildInfo3.pivotRootHash,
		rebuildFFIndex:  rebuildInfo3.rebuildFFIndex,
		pivotFFIndex:    rebuildInfo3.pivotFFIndex,
	}
	stateDB2, err := NewWithMode(dbName, common.STATEDB_ARCHIVE_MODE, db, rebuildInfo4, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= 5; i++ {
		if v, err := stateDB2.getStateObject(TestObjectType, testObjects[i].GetHash()); err != nil {
			t.Fatal(err)
		} else {
			if !bytes.Equal(v.GetValueBytes(), testObjects[i].GetValueBytes()) {
				t.Fatalf("expect %+v, got %+v", testObjects[i].GetValueBytes(), v.GetValueBytes())
			}
		}
	}

	for i := 6; i < len(testObjects); i++ {
		if v1, err := stateDB2.getStateObject(TestObjectType, testObjects[i].GetHash()); err != nil {
			t.Fatal(err)
		} else if v1 != nil {
			t.Fatalf("expect nil, but got %+v", v1)
		}
	}

}

func TestBatchCommitIterator(t *testing.T) {

	config.LoadConfig()
	config.LoadParam()

	//generate data
	var randKey []common.Hash
	var randValue [][]byte
	rand.Seed(1)
	for i := 0; i < 100; i++ {
		k, v := genRandomKV()
		randKey = append(randKey, k)
		randValue = append(randValue, v)
	}

	wantKeys := [][]byte{}
	temp := common.HashH([]byte("test-prefix"))
	prefix := temp[:12]
	for i := 0; i < 5; i++ {
		randKey[i] = common.BytesToHash(append(prefix, randKey[i][12:]...))
		wantKeys = append(wantKeys, randKey[i][:])
	}
	//init DB and txDB
	os.RemoveAll("./tmp")
	db, err := incdb.Open("leveldb", "./tmp")
	if err != nil {
		t.Fatal(err)
	}

	dbName := "test"

	stateDB, err := NewWithMode(dbName, common.STATEDB_BATCH_COMMIT_MODE, db, *NewEmptyRebuildInfo(""), nil)
	if err != nil {
		t.Fatal(err)
	}
	testObjects := make([]StateObject, 7)
	testObjects[0], _ = newTestObjectWithValue(stateDB, randKey[0], randValue[0])
	testObjects[1], _ = newTestObjectWithValue(stateDB, randKey[1], randValue[1])
	testObjects[2], _ = newTestObjectWithValue(stateDB, randKey[2], randValue[2])
	testObjects[3], _ = newTestObjectWithValue(stateDB, randKey[3], randValue[3])
	testObjects[4], _ = newTestObjectWithValue(stateDB, randKey[4], randValue[4])
	testObjects[5], _ = newTestObjectWithValue(stateDB, randKey[5], randValue[5])
	testObjects[6], _ = newTestObjectWithValue(stateDB, randKey[6], randValue[6])

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[0].GetHash(), testObjects[0].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[1].GetHash(), testObjects[1].GetValue()); err != nil {
		t.Fatal(err)
	}

	_, _, err = stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[2].GetHash(), testObjects[2].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[3].GetHash(), testObjects[3].GetValue()); err != nil {
		t.Fatal(err)
	}

	_, _, err = stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	// check set & get object
	if err := stateDB.SetStateObject(TestObjectType, testObjects[4].GetHash(), testObjects[4].GetValue()); err != nil {
		t.Fatal(err)
	}
	if err := stateDB.SetStateObject(TestObjectType, testObjects[5].GetHash(), testObjects[5].GetValue()); err != nil {
		t.Fatal(err)
	}

	_, _, err = stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	if err := stateDB.Finalized(true, *NewEmptyRebuildInfo("")); err != nil {
		t.Fatal(err)
	}

	stateDB2 := stateDB.Copy()

	nodeIter := stateDB2.trie.NodeIterator(prefix)
	iter := trie.NewIterator(nodeIter)
	gotKeys := [][]byte{}
	for iter.Next() {
		gotKey := make([]byte, len(iter.Key))
		copy(gotKey, iter.Key)
		gotKeys = append(gotKeys, gotKey)
	}

	sort.Slice(wantKeys, func(i, j int) bool {
		return string(wantKeys[i]) < string(wantKeys[j])
	})

	sort.Slice(gotKeys, func(i, j int) bool {
		return string(gotKeys[i]) < string(gotKeys[j])
	})

	if !reflect.DeepEqual(wantKeys, gotKeys) {
		t.Fatal("iterator test fail")
	}
}
