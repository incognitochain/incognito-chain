package gomobile

import (
	"encoding/base64"
	"encoding/json"
	zkp "github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"math/big"
	"time"
)

func Add(a int, b int) int {
	return a + b
}

func SayHello(i string) string {
	println("Hello %s \n", i)
	return i
}

func AggregatedRangeProve(args string) string {
	println("args:", args)
	bytes := []byte(args)
	println("Bytes:", bytes)
	temp := make(map[string][]string)

	err := json.Unmarshal(bytes, &temp)
	if err != nil {
		println(err)
		return ""
	}
	println("temp values", temp["values"])
	println("temp rands", temp["rands"])

	if len(temp["values"]) != len(temp["rands"]) {
		println("Wrong args")
	}

	values := make([]*big.Int, len(temp["values"]))
	rands := make([]*big.Int, len(temp["values"]))

	for i := 0; i < len(temp["values"]); i++ {
		values[i], _ = new(big.Int).SetString(temp["values"][i], 10)
		rands[i], _ = new(big.Int).SetString(temp["rands"][i], 10)
	}

	wit := new(zkp.AggregatedRangeWitness)
	wit.Set(values, rands)

	start := time.Now()
	proof, err := wit.Prove()
	if err != nil {
		println("Err: %v\n", err)
	}
	end := time.Since(start)
	println("Aggregated range proving time: %v\n", end)

	proofBytes := proof.Bytes()
	println("Proof bytes: ", proofBytes)

	proofBase64 := base64.StdEncoding.EncodeToString(proofBytes)
	println("proofBase64: %v\n", proofBase64)

	return proofBase64
}
