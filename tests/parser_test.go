package main

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"
)

func TestReadFile(t *testing.T) {
	res, err := readfile("./testsdata/normal_transaction.json")
	if err != nil {
		t.Fatal()
	}
	log.Println(res)
}
func TestParseResultNumber(t *testing.T) {
	var (
		numberInt int = 10
		numberInt64 int64 = 10
		numberInt32 int32 = 10
		numberUint uint = 10
		numberUint32 uint32 = 10
		numberUint64 uint64 = 10
		numberByte byte = 10
		numberFloat32 float32 = 10
		numberFloat64 float64 = 10
		
		expectNumber interface{} = float64(10)
		notExpectNumber interface{} = float64(9)
	)
	if numberBytes, err := json.Marshal(numberInt); err != nil {
		t.Fatal(err)
	} else {
		result := parseResult(numberBytes)
		if!reflect.DeepEqual(result, expectNumber) {
			t.Fatalf("Expect %+v, get %+v", expectNumber, result)
		}
		if reflect.DeepEqual(result, notExpectNumber) {
			t.Fatalf("Expect %+v, get %+v", notExpectNumber, result)
		}
	}
	if numberBytes, err := json.Marshal(numberInt64); err != nil {
		t.Fatal(err)
	} else {
		result := parseResult(numberBytes)
		if!reflect.DeepEqual(result, expectNumber) {
			t.Fatalf("Expect %+v, get %+v", expectNumber, result)
		}
		if reflect.DeepEqual(result, notExpectNumber) {
			t.Fatalf("Expect %+v, get %+v", notExpectNumber, result)
		}
	}
	if numberBytes, err := json.Marshal(numberInt32); err != nil {
		t.Fatal(err)
	} else {
		result := parseResult(numberBytes)
		if!reflect.DeepEqual(result, expectNumber) {
			t.Fatalf("Expect %+v, get %+v", expectNumber, result)
		}
		if reflect.DeepEqual(result, notExpectNumber) {
			t.Fatalf("Expect %+v, get %+v", notExpectNumber, result)
		}
	}
	if numberBytes, err := json.Marshal(numberUint); err != nil {
		t.Fatal(err)
	} else {
		result := parseResult(numberBytes)
		if!reflect.DeepEqual(result, expectNumber) {
			t.Fatalf("Expect %+v, get %+v", expectNumber, result)
		}
		if reflect.DeepEqual(result, notExpectNumber) {
			t.Fatalf("Expect %+v, get %+v", notExpectNumber, result)
		}
	}
	if numberBytes, err := json.Marshal(numberUint32); err != nil {
		t.Fatal(err)
	} else {
		result := parseResult(numberBytes)
		if!reflect.DeepEqual(result, expectNumber) {
			t.Fatalf("Expect %+v, get %+v", expectNumber, result)
		}
		if reflect.DeepEqual(result, notExpectNumber) {
			t.Fatalf("Expect %+v, get %+v", notExpectNumber, result)
		}
	}
	if numberBytes, err := json.Marshal(numberUint64); err != nil {
		t.Fatal(err)
	} else {
		result := parseResult(numberBytes)
		if!reflect.DeepEqual(result, expectNumber) {
			t.Fatalf("Expect %+v, get %+v", expectNumber, result)
		}
		if reflect.DeepEqual(result, notExpectNumber) {
			t.Fatalf("Expect %+v, get %+v", notExpectNumber, result)
		}
	}
	if numberBytes, err := json.Marshal(numberByte); err != nil {
		t.Fatal(err)
	} else {
		result := parseResult(numberBytes)
		if!reflect.DeepEqual(result, expectNumber) {
			t.Fatalf("Expect %+v, get %+v", expectNumber, result)
		}
		if reflect.DeepEqual(result, notExpectNumber) {
			t.Fatalf("Expect %+v, get %+v", notExpectNumber, result)
		}
	}
	if numberBytes, err := json.Marshal(numberFloat32); err != nil {
		t.Fatal(err)
	} else {
		result := parseResult(numberBytes)
		if!reflect.DeepEqual(result, expectNumber) {
			t.Fatalf("Expect %+v, get %+v", expectNumber, result)
		}
		if reflect.DeepEqual(result, notExpectNumber) {
			t.Fatalf("Expect %+v, get %+v", notExpectNumber, result)
		}
	}
	if numberBytes, err := json.Marshal(numberFloat64); err != nil {
		t.Fatal(err)
	} else {
		result := parseResult(numberBytes)
		if!reflect.DeepEqual(result, expectNumber) {
			t.Fatalf("Expect %+v, get %+v", expectNumber, result)
		}
		if reflect.DeepEqual(result, notExpectNumber) {
			t.Fatalf("Expect %+v, get %+v", notExpectNumber, result)
		}
	}
}

func TestParseResultString(t *testing.T) {
	var (
		str string = "0xabc"
		
		expectString interface{} = string("0x456")
		notExpectString interface{} = string("0x4567")
	)
	if strBytes, err := json.Marshal(str); err != nil {
		t.Fatal(err)
	} else {
		result := parseResult(strBytes)
		if!reflect.DeepEqual(result, expectString) {
			t.Fatalf("Expect %+v, get %+v", expectString, result)
		}
		if reflect.DeepEqual(result, notExpectString) {
			t.Fatalf("Expect %+v, get %+v", notExpectString, result)
		}
	}
}

func TestParseResultBool(t *testing.T) {
	var (
		boolean bool = true
		
		expectBoolean interface{} = bool(true)
		notExpectBoolean interface{} = bool(false)
	)
	if booleanBytes, err := json.Marshal(boolean); err != nil {
		t.Fatal(err)
	} else {
		result := parseResult(booleanBytes)
		if!reflect.DeepEqual(result, expectBoolean) {
			t.Fatalf("Expect %+v, get %+v", expectBoolean, result)
		}
		if reflect.DeepEqual(result, notExpectBoolean) {
			t.Fatalf("Expect %+v, get %+v", notExpectBoolean, result)
		}
	}
}

func TestParseResultArray(t *testing.T) {
	var (
		arrString []string = []string{"a","b","c","d"}
		arrNumber []int = []int{1,2,3,4,5}
		arrBool []bool = []bool{true, true, false, false}
		
		expectArrString interface{} = []interface{}{"a","b","c","d"}
		notExpectArrString interface{} = []interface{}{"a","c","b"}
		expectArrNumber interface{} = []interface{}{float64(1),float64(2),float64(3),float64(4),float64(5)}
		notExpectArrNumber interface{} = []interface{}{float64(1),float64(2),float64(3),float64(5),float64(4)}
		expectArrBool interface{} = []interface{}{true, true, false, false}
		notExpectArrBool interface{} = []interface{}{true, true, false, true}
	)
	if arrBytes, err := json.Marshal(arrString); err != nil {
		t.Fatal(err)
	} else {
		result := parseResult(arrBytes)
		if!reflect.DeepEqual(result, expectArrString) {
			t.Fatalf("Expect %+v, get %+v", expectArrString, result)
		}
		if reflect.DeepEqual(result, notExpectArrString) {
			t.Fatalf("Expect %+v, get %+v", notExpectArrString, result)
		}
	}
	if arrBytes, err := json.Marshal(arrNumber); err != nil {
		t.Fatal(err)
	} else {
		result := parseResult(arrBytes)
		if!reflect.DeepEqual(result, expectArrNumber) {
			t.Fatalf("Expect %+v, get %+v", expectArrNumber, result)
		}
		if reflect.DeepEqual(result, notExpectArrNumber) {
			t.Fatalf("Expect %+v, get %+v", notExpectArrNumber, result)
		}
	}
	if arrBytes, err := json.Marshal(arrBool); err != nil {
		t.Fatal(err)
	} else {
		result := parseResult(arrBytes)
		if!reflect.DeepEqual(result, expectArrBool) {
			t.Fatalf("Expect %+v, get %+v", expectArrBool, result)
		}
		if reflect.DeepEqual(result, notExpectArrBool) {
			t.Fatalf("Expect %+v, get %+v", notExpectArrBool, result)
		}
	}
}

func TestParseResultObject(t *testing.T) {
	var (
		obj = map[string]interface{}{
			"tx": "abc",
			"id": 10,
			"done": true,
		}
		
		expectObj = map[string]interface{}{
			"tx": "abc",
			"id": float64(10),
			"done": true,
		}
		notExpectObj = map[string]interface{}{
			"tx": "abc",
			"id": int(10),
			"done": true,
		}
	)
	if objBytes, err := json.Marshal(obj); err != nil {
		t.Fatal(err)
	} else {
		result := parseResult(objBytes)
		if!reflect.DeepEqual(result, expectObj) {
			t.Fatalf("Expect %+v, get %+v", expectObj, result)
		}
		if reflect.DeepEqual(result, notExpectObj) {
			t.Fatalf("Expect %+v, get %+v", notExpectObj, result)
		}
	}
}