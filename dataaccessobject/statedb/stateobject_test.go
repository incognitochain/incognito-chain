package statedb

import (
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
	"testing"
)

func TestMapByteSerialize(t *testing.T) {

	stateDB := new(StateDB)

	testObjectKey1 := common.HashH([]byte{1})
	testObjectValue1 := []byte("test1")
	testObject1, _ := newTestObjectWithValue(stateDB, testObjectKey1, testObjectValue1)

	testObjectKey2 := common.HashH([]byte{2})
	testObjectValue2 := []byte("test2")
	testObject2, _ := newTestObjectWithValue(stateDB, testObjectKey2, testObjectValue2)
	testObject2.MarkDelete()

	testObjectKey3 := common.HashH([]byte{3})
	testObjectValue3 := []byte("test3")
	testObject3, _ := newTestObjectWithValue(stateDB, testObjectKey3, testObjectValue3)

	m := make(map[common.Hash]StateObject)
	m[testObjectKey1] = testObject1
	m[testObjectKey2] = testObject2
	m[testObjectKey3] = testObject3

	res := MapByteSerialize(m)

	gotM, err := MapByteDeserialize(stateDB, res)
	if err != nil {
		t.Fatal(err)
	}

	if len(gotM) != len(m) {
		t.Fatalf("len(gotM) = %d, len(m) = %d", len(gotM), len(m))
	}

	for k, v := range m {
		gotV := gotM[k]
		if !reflect.DeepEqual(v, gotV) {
			t.Fatalf("want = %+v, got %+v", v, gotV)
		}
	}
}

func TestMapByteSerialize1(t *testing.T) {

	stateDB := new(StateDB)

	testObjectKey1 := common.HashH([]byte{1})
	testObjectValue1 := []byte("test1")
	testObject1, _ := newTestObjectWithValue(stateDB, testObjectKey1, testObjectValue1)

	testObjectKey2 := common.HashH([]byte{2})
	testObjectValue2 := []byte("test2")
	testObject2, _ := newTestObjectWithValue(stateDB, testObjectKey2, testObjectValue2)
	testObject2.MarkDelete()

	m := make(map[common.Hash]StateObject)
	m[testObjectKey1] = testObject1

	res := MapByteSerialize(m)

	gotM, err := MapByteDeserialize(stateDB, res)
	if err != nil {
		t.Fatal(err)
	}

	if len(gotM) != len(m) {
		t.Fatalf("len(gotM) = %d, len(m) = %d", len(gotM), len(m))
	}

	for k, v := range m {
		gotV := gotM[k]
		if !reflect.DeepEqual(v, gotV) {
			t.Fatalf("want = %+v, got %+v", v, gotV)
		}
	}
}
