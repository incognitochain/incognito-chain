package gomobile

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/oneoutofmany"
	"math/big"
	"time"

	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/aggregaterange"
)

func Add(a int, b int) int {
	return a + b
}

func SayHello(i string) string {
	println("Hello %s \n", i)
	return i
}

// args {
//      "values": valueStrs,
//      "rands": randStrs
//    }
// convert object to JSON string (JSON.stringify)
func AggregatedRangeProve(args string) string {
	println("args:", args)
	bytes := []byte(args)
	println("Bytes:", bytes)
	temp := make(map[string][]string)

	err := json.Unmarshal(bytes, &temp)
	if err != nil {
		println("Can not unmarshal", err)
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

	wit := new(aggregaterange.AggregatedRangeWitness)
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

// args {
//      "commitments": commitments,   // list of bytes arrays
//      "rand": rand,					// string
// 		"indexiszero" 					//number
//    }
// convert object to JSON string (JSON.stringify)
func OneOutOfManyProve(args string) (string, error) {
	bytes := []byte(args)
	//println("Bytes:", bytes)
	temp := make(map[string][]string)

	err := json.Unmarshal(bytes, &temp)
	if err != nil {
		println(err)
		return "", err
	}

	// list of commitments
	commitmentStrs := temp["commitments"]
	//fmt.Printf("commitmentStrs: %v\n", commitmentStrs)

	if len(commitmentStrs) != privacy.CommitmentRingSize {
		println(err)
		return "", errors.New("the number of Commitment list's elements must be equal to CMRingSize")
	}

	commitmentPoints := make([]*privacy.EllipticPoint, len(commitmentStrs))

	for i := 0; i < len(commitmentStrs); i++ {
		//fmt.Printf("commitments[i]: %v\n", commitmentStrs[i])
		tmp, _ := new(big.Int).SetString(commitmentStrs[i], 16)
		tmpByte := tmp.Bytes()

		commitmentPoints[i] = new(privacy.EllipticPoint)
		err = commitmentPoints[i].Decompress(tmpByte)
		if err != nil {
			println(err)
			return "", err
		}
	}

	// rand
	randBN, _ := new(big.Int).SetString(temp["rand"][0], 10)
	//println("randBN: ", randBN)

	// indexIsZero
	indexIsZero, _ := new(big.Int).SetString(temp["indexiszero"][0], 10)
	indexIsZeroUint64 := indexIsZero.Uint64()

	//println("indexIsZeroUint64: ", indexIsZeroUint64)

	// set witness for One out of many protocol
	wit := new(oneoutofmany.OneOutOfManyWitness)
	wit.Set(commitmentPoints, randBN, indexIsZeroUint64)
	println("Wit: ", wit)

	// proving
	//start := time.Now()
	proof, err := wit.Prove()
	//fmt.Printf("Proof go: %v\n", proof)
	if err != nil {
		println("Err: %v\n", err)
	}
	//end := time.Since(start)
	//fmt.Printf("One out of many proving time: %v\n", end)

	// convert proof to bytes array
	proofBytes := proof.Bytes()
	//println("Proof bytes: ", proofBytes)

	proofBase64 := base64.StdEncoding.EncodeToString(proofBytes)
	//println("proofBase64: %v\n", proofBase64)

	return proofBase64, nil
}

// GenerateBLSKeyPairFromSeed generates BLS key pair from seed
func GenerateBLSKeyPairFromSeed(args string) string {
	// convert seed from string to bytes array
	//fmt.Printf("args: %v\n", args)
	seed, _ := base64.StdEncoding.DecodeString(args)
	//fmt.Printf("bls seed: %v\n", seed)

	// generate  bls key
	privateKey, publicKey := blsmultisig.KeyGen(seed)

	// append key pair to one bytes array
	keyPairBytes := []byte{}
	keyPairBytes = append(keyPairBytes, privateKey.Bytes()...)
	keyPairBytes = append(keyPairBytes, blsmultisig.CmprG2(publicKey)...)

	//  base64.StdEncoding.EncodeToString()
	keyPairEncode := base64.StdEncoding.EncodeToString(keyPairBytes)

	return keyPairEncode
}
