package common

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

/*
	Unit test for InterfaceSlice function
*/

func TestCommonInterfaceSlice(t *testing.T) {
	data := []struct {
		slice interface{}
		len   int
	}{
		{[]byte{1}, 1},
		{[]byte{1, 2, 3}, 3},
		{[]byte{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5}, 25},
		{[]byte{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5}, 30},
	}

	for _, item := range data {
		slice := InterfaceSlice(item.slice)
		assert.Equal(t, item.len, len(slice))
	}
}

func TestCommonInterfaceSliceWithInvalidSliceInterface(t *testing.T) {
	data := []struct {
		slice interface{}
	}{
		{"abc"},
		{123},
		{struct{ a int }{12}},
		{nil},
	}

	for _, item := range data {
		slice := InterfaceSlice(item.slice)
		assert.Equal(t, []interface{}(nil), slice)
	}
}

/*
	Unit test for ParseListener function
*/

func TestCommonParseListener(t *testing.T) {
	data := []struct {
		addr    string
		netType string
	}{
		{"1.2.3.4:9934", "test"},
		{"100.255.3.4:9934", "main"},
		{"192.168.3.4:9934", "main1"},
		{"0.0.0.0:9934", "main"},
		{":9934", "main"},    // empty host
		{"1.2.3.4:9934", ""}, // empty netType
	}

	for _, item := range data {
		SimpleAddr, err := ParseListener(item.addr, item.netType)
		fmt.Printf("SimpleAddr.Addr: %v\n", SimpleAddr.Addr)
		fmt.Printf("SimpleAddr.Net: %v\n", SimpleAddr.Net)

		assert.Equal(t, nil, err)
		assert.Equal(t, item.addr, SimpleAddr.Addr)
		assert.Equal(t, item.netType+"4", SimpleAddr.Net)
	}
}

func TestCommonParseListenerWithInvalidIPAddr(t *testing.T) {
	data := []struct {
		addr    string
		netType string
	}{
		{"256.2.3.4:9934", "test"},
		{"1.2.3:9934", "main1"},
		{"1.2:9934", "main1"},
		{"1:9934", "main1"},
		{"*:9934", "main1"},
		{"a.2.3.4:9934", "test"},
		{"-.2.3.4:9934", "test"},
	}

	for _, item := range data {
		_, err := ParseListener(item.addr, item.netType)
		assert.Equal(t, errors.New("IP address is invalid").Error(), err.Error())
	}
}

func TestCommonParseListenerWithInvalidPort(t *testing.T) {
	data := []struct {
		addr    string
		netType string
	}{
		{"100.255.3.4:-2", "main"},
		{"192.168.3.4:a", "main1"},
		{"0.0.0.0:?", "main"},
		{":", "main"},       // empty port
		{"1.2.3.4:...", ""}, // empty netType
	}

	for _, item := range data {
		_, err := ParseListener(item.addr, item.netType)
		assert.Equal(t, errors.New("port is invalid").Error(), err.Error())
	}
}

/*
	Unit test for ParseListeners function
*/

func TestCommonParseListeners(t *testing.T) {
	addrs := []string{
		"1.2.3.4:9934",
		"100.255.3.4:9934",
		"100.255.3.4:9934",
		"0.0.0.0:9934",
		":9934",
		"1.2.3.4:9934",
	}

	netType := "test"

	simpleAddrs, err := ParseListeners(addrs, netType)

	assert.Equal(t, nil, err)
	assert.Equal(t, 6, len(simpleAddrs))
}

func TestCommonParseListenersWithInvalidIPAddr(t *testing.T) {
	addrs := []string{
		"256.2.3.4:9934",
		"100.255.3.4:9934",
		"100.255.3.4:9934",
		"0.0.0.0:9934",
		":9934",
		"1.2.3.4:9934",
	}

	netType := "test"

	simpleAddrs, err := ParseListeners(addrs, netType)

	assert.Equal(t, errors.New("IP address is invalid").Error(), err.Error())
	assert.Equal(t, 0, len(simpleAddrs))
}

func TestCommonParseListenersWithInvalidPort(t *testing.T) {
	addrs := []string{
		"100.2.3.4:a",
		"100.255.3.4:9934",
		"100.255.3.4:9934",
		"0.0.0.0:9934",
		":9934",
		"1.2.3.4:9934",
	}

	netType := "test"

	simpleAddrs, err := ParseListeners(addrs, netType)

	assert.Equal(t, errors.New("port is invalid").Error(), err.Error())
	assert.Equal(t, 0, len(simpleAddrs))
}

/*
	Unit test for SliceExists function
*/
func TestCommonSliceExists(t *testing.T) {
	data := []struct {
		slice     interface{}
		item      interface{}
		isContain bool
	}{
		{[]byte{1, 2, 3, 4, 5, 6}, byte(6), true},
		{[]int{1, 2, 3, 4, 5, 6}, int(10), false},
		{[]byte{1, 2, 3, 4, 5, 6}, 6, false},
		{[]string{"a", "b", "c", "d", "e"}, "E", false},
		{[]*big.Int{big.NewInt(int64(100)), big.NewInt(int64(1000)), big.NewInt(int64(10000)), big.NewInt(int64(100000)), big.NewInt(int64(10000000))}, big.NewInt(int64(100001)), false},
	}

	for _, dataItem := range data {
		isContain, err := SliceExists(dataItem.slice, dataItem.item)
		assert.Equal(t, nil, err)
		assert.Equal(t, dataItem.isContain, isContain)
	}
}

func TestCommonSliceExistsWithInvalidSlice(t *testing.T) {
	data := []struct {
		slice interface{}
		item  interface{}
	}{
		{"abc", "a"},
		{123456, 4},
	}

	for _, dataItem := range data {
		isContain, err := SliceExists(dataItem.slice, dataItem.item)
		assert.Equal(t, errors.New("SliceExists() given a non-slice type").Error(), err.Error())
		assert.Equal(t, false, isContain)
	}
}

/*
	Unit test for GetShardIDFromLastByte function
*/

func TestCommonGetShardIDFromLastByte(t *testing.T) {
	data := []byte{
		1,
		2,
		108,
		203,
		255,
	}

	for _, item := range data {
		shardID := GetShardIDFromLastByte(item)
		assert.Equal(t, item%MaxShardNumber, shardID)
	}
}

/*
	Unit test for IndexOfStr function
*/

func TestCommonIndexOfStr(t *testing.T) {
	data := []struct {
		list  []string
		item  string
		index int
	}{
		{[]string{"a", "b", "c", "d", "e"}, "E", -1},
		{[]string{"Incognito", "Constant", "Decentralized", "Privacy", "Incognito", "Stable"}, "Incognito", 0},
		{[]string{"Constant", "Decentralized", "Privacy", "Incognito", "Stable"}, "Incognito", 3},
	}

	for _, dataItem := range data {
		index := IndexOfStr(dataItem.item, dataItem.list)
		assert.Equal(t, dataItem.index, index)
	}
}

/*
	Unit test for IndexOfStrInHashMap function
*/

func TestCommonIndexOfStrInHashMap(t *testing.T) {
	bytes := []byte{1, 2, 3}
	hash1 := HashH(bytes)

	bytes2 := []byte{1, 2, 3, 4}
	hash2 := HashH(bytes2)

	bytes3 := []byte{1, 2, 3, 4, 5}
	hash3 := HashH(bytes3)

	data := []struct {
		m      map[Hash]string
		v      string
		result int
	}{
		{map[Hash]string{hash1: "abc", hash2: "abcd", hash3: "lala"}, "lala", 1},
		{map[Hash]string{hash1: "Incognito", hash2: "Constant", hash3: "Decentralized"}, "Privacy", -1},
	}

	for _, dataItem := range data {
		index := IndexOfStrInHashMap(dataItem.v, dataItem.m)
		assert.Equal(t, dataItem.result, index)
	}
}

/*
	Unit test for RandBigIntMaxRange function
*/

func TestCommonRandBigIntMaxRange(t *testing.T) {
	data := []*big.Int{
		big.NewInt(int64(1234567890)),
		big.NewInt(int64(100000000)),
		big.NewInt(int64(1)),
	}

	for _, item := range data {
		number, err := RandBigIntMaxRange(item)
		//fmt.Printf("number: %v\n", number)
		cmp := number.Cmp(item)

		assert.Equal(t, nil, err)
		assert.Equal(t, -1, cmp)
	}
}

/*
	Unit test for CompareStringArray function
*/

func TestCommonCompareStringArray(t *testing.T) {
	data := []struct {
		src     []string
		dst     []string
		isEqual bool
	}{
		{[]string{"a", "b", "c", "d", "e"}, []string{"a", "b", "c", "d", "e"}, true},
		{[]string{"a", "b", "c", "d", "e"}, []string{"a", "b", "c", "d", "f"}, false},
		{[]string{"a", "b", "c", "d", "e", "a", "b", "c", "d", "e"}, []string{"a", "b", "c", "d", "e"}, false},
	}

	for _, item := range data {
		isEqual := CompareStringArray(item.src, item.dst)
		assert.Equal(t, item.isEqual, isEqual)
	}
}

/*
	Unit test for BytesToInt32 function
*/

func TestCommonBytesToInt32(t *testing.T) {
	data := []struct {
		bytes  []byte
		number int32
	}{
		{[]byte{1, 2, 3, 4}, 67305985},
		{[]byte{1, 2, 3, 0}, 197121},
		{[]byte{1, 2, 3, 10}, 167969281},
		{[]byte{1, 7, 8, 9}, 151521025},
		{[]byte{1, 2, 10, 4}, 67764737},
	}

	for _, item := range data {
		number, err := BytesToInt32(item.bytes)

		assert.Equal(t, nil, err)
		assert.Equal(t, item.number, number)
	}
}

func TestCommonBytesToInt32WithInvalidInput(t *testing.T) {
	data := [][]byte{
		{1, 2, 3, 4, 5},
		{1, 2, 3},
	}

	for _, item := range data {
		_, err := BytesToInt32(item)
		assert.Equal(t, errors.New("invalid length of input BytesToInt32").Error(), err.Error())
	}
}

/*
	Unit test for Int32ToBytes function
*/

func TestCommonInt32ToBytes(t *testing.T) {
	data := []struct {
		bytes  []byte
		number int32
	}{
		{[]byte{1, 2, 3, 4}, 67305985},
		{[]byte{1, 2, 3, 0}, 197121},
		{[]byte{1, 2, 3, 10}, 167969281},
		{[]byte{1, 7, 8, 9}, 151521025},
		{[]byte{1, 2, 10, 4}, 67764737},
	}

	for _, item := range data {
		bytes := Int32ToBytes(item.number)
		assert.Equal(t, item.bytes, bytes)
	}
}

/*
	Unit test for BytesToUint32 function
*/

func TestCommonBytesToUint32(t *testing.T) {
	data := []struct {
		bytes  []byte
		number uint32
	}{
		{[]byte{1, 2, 3, 4}, 16909060},
		{[]byte{1, 2, 3, 0}, 16909056},
		{[]byte{1, 2, 3, 10}, 16909066},
		{[]byte{1, 7, 8, 9}, 17238025},
		{[]byte{1, 2, 10, 4}, 16910852},
	}

	for _, item := range data {
		number, err := BytesToUint32(item.bytes)
		//fmt.Printf("number: %v\n", number)

		assert.Equal(t, nil, err)
		assert.Equal(t, item.number, number)
	}
}

func TestCommonBytesToUint32WithInvalidInput(t *testing.T) {
	data := [][]byte{
		{1, 2, 3, 4, 5},
		{1, 2, 3},
	}

	for _, item := range data {
		_, err := BytesToUint32(item)
		assert.Equal(t, errors.New("invalid length of input BytesToUint32").Error(), err.Error())
	}
}

/*
	Unit test for Uint32ToBytes function
*/

func TestCommonUint32ToBytes(t *testing.T) {
	data := []struct {
		bytes  []byte
		number uint32
	}{
		{[]byte{1, 2, 3, 4}, 16909060},
		{[]byte{1, 2, 3, 0}, 16909056},
		{[]byte{1, 2, 3, 10}, 16909066},
		{[]byte{1, 7, 8, 9}, 17238025},
		{[]byte{1, 2, 10, 4}, 16910852},
	}

	for _, item := range data {
		bytes := Uint32ToBytes(item.number)
		assert.Equal(t, item.bytes, bytes)
	}
}

/*
	Unit test for BytesToUint64 function
*/

func TestCommonBytesToUint64(t *testing.T) {
	data := []struct {
		bytes  []byte
		number uint64
	}{
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8}, 578437695752307201},
		{[]byte{1, 2, 3, 0, 0, 0, 0, 0}, 197121},
		{[]byte{1, 2, 3, 10, 1, 2, 3, 10}, 721422568795603457},
		{[]byte{1, 7, 8, 9, 1, 7, 8, 9}, 650777847182919425},
		{[]byte{1, 2, 10, 4, 1, 2, 10, 4}, 291047329304805889},
	}

	for _, item := range data {
		number, err := BytesToUint64(item.bytes)
		//fmt.Printf("number: %v\n", number)

		assert.Equal(t, nil, err)
		assert.Equal(t, item.number, number)
	}
}

func TestCommonBytesToUint64WithInvalidInput(t *testing.T) {
	data := [][]byte{
		{1, 2, 3, 4, 5},
		{1, 2, 3},
	}

	for _, item := range data {
		_, err := BytesToUint64(item)
		assert.Equal(t, errors.New("invalid length of input BytesToUint64").Error(), err.Error())
	}
}

/*
	Unit test for Uint64ToBytes function
*/

func TestCommonUint64ToBytes(t *testing.T) {
	data := []struct {
		bytes  []byte
		number uint64
	}{
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8}, 578437695752307201},
		{[]byte{1, 2, 3, 0, 0, 0, 0, 0}, 197121},
		{[]byte{1, 2, 3, 10, 1, 2, 3, 10}, 721422568795603457},
		{[]byte{1, 7, 8, 9, 1, 7, 8, 9}, 650777847182919425},
		{[]byte{1, 2, 10, 4, 1, 2, 10, 4}, 291047329304805889},
	}

	for _, item := range data {
		bytes := Uint64ToBytes(item.number)
		assert.Equal(t, item.bytes, bytes)
	}
}

/*
	Unit test for Int64ToBytes function
*/

func TestCommonInt64ToBytes(t *testing.T) {
	data := []struct {
		bytes  []byte
		number int64
	}{
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8}, 578437695752307201},
		{[]byte{1, 2, 3, 0, 0, 0, 0, 0}, 197121},
		{[]byte{1, 2, 3, 10, 1, 2, 3, 10}, 721422568795603457},
		{[]byte{1, 7, 8, 9, 1, 7, 8, 9}, 650777847182919425},
		{[]byte{1, 2, 10, 4, 1, 2, 10, 4}, 291047329304805889},
	}

	for _, item := range data {
		bytes := Int64ToBytes(item.number)
		assert.Equal(t, item.bytes, bytes)
	}
}

/*
	Unit test for BoolToByte function
*/

func TestCommonBoolToByte(t *testing.T) {
	data := []struct {
		boolValue bool
		byteValue byte
	}{
		{true, 1},
		{false, 0},
	}

	for _, item := range data {
		byteValue := BoolToByte(item.boolValue)
		assert.Equal(t, item.byteValue, byteValue)
	}
}

/*
	Unit test for IndexOfByte function
*/

func TestCommonIndexOfByte(t *testing.T) {
	data := []struct {
		list  []byte
		item  byte
		index int
	}{
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8, 5}, byte(5), 4},
		{[]byte{145, 23, 3, 44, 52, 6, 47, 28}, byte(23), 1},
		{[]byte{145, 23, 3, 44, 52, 6, 47, 28}, byte(5), -1},
	}

	for _, dataItem := range data {
		index := IndexOfByte(dataItem.item, dataItem.list)
		assert.Equal(t, dataItem.index, index)
	}
}

/*
	Unit test for AppendSliceString function
*/

func TestCommonAppendSliceString(t *testing.T) {
	arr1 := [][]string{
		{"a", "b", "c"},
		{"1", "2", "3"},
	}
	arr2 := [][]string{
		{"d", "e", "f"},
		{"4", "5", "6"},
	}
	arr3 := [][]string{
		{"g", "h", "k"},
		{"7", "8", "9"},
	}

	finalArr := AppendSliceString(arr1, arr2, arr3)
	assert.Equal(t, 6, len(finalArr))
}

//func TestCommonHashToString(t *testing.T){
//	for i:=0; i< 1000; i++{
//		hash := new(Hash)
//		hash.SetBytes([]byte{1,2,3,4})
//		fmt.Printf("Hash string len: %v\n", len(hash.String()))
//	}
//}
