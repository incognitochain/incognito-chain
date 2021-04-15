package main

import (
	b64 "encoding/base64"

	"github.com/btcsuite/btcutil/base58"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"

	"encoding/json"
	"fmt"
	"io/ioutil"
)

type EmptyStruct struct{}

type CoinDetails struct {
	PublicKey      string `json:"PublicKey"`
	CoinCommitment string `json:"CoinCommitment"`
	SNDerivator    struct {
	} `json:"SNDerivator"`
	SerialNumber string `json:"SerialNumber"`
	Randomness   struct {
	} `json:"Randomness"`
	Value int    `json:"Value"`
	Info  string `json:"Info"`
}

type Txver2 struct {
	ID     int `json:"Id"`
	Result struct {
		BlockHash   string `json:"BlockHash"`
		BlockHeight int    `json:"BlockHeight"`
		TxSize      int    `json:"TxSize"`
		Index       int    `json:"Index"`
		ShardID     int    `json:"ShardID"`
		Hash        string `json:"Hash"`
		Version     int    `json:"Version"`
		Type        string `json:"Type"`
		LockTime    string `json:"LockTime"`
		Fee         int    `json:"Fee"`
		Image       string `json:"Image"`
		IsPrivacy   bool   `json:"IsPrivacy"`
		Proof       string `json:"Proof"`
		ProofDetail struct {
			InputCoins []struct {
				Detail               CoinDetails `json:"CoinDetails"`
				CoinDetailsEncrypted string      `json:"CoinDetailsEncrypted"`
			} `json:"InputCoins"`
			OutputCoins []struct {
				Detail               CoinDetails `json:"CoinDetails"`
				CoinDetailsEncrypted string      `json:"CoinDetailsEncrypted"`
			} `json:"OutputCoins"`
		} `json:"ProofDetail"`
		InputCoinPubKey               string `json:"InputCoinPubKey"`
		SigPubKey                     string `json:"SigPubKey"`
		Sig                           string `json:"Sig"`
		Metadata                      string `json:"Metadata"`
		CustomTokenData               string `json:"CustomTokenData"`
		PrivacyCustomTokenID          string `json:"PrivacyCustomTokenID"`
		PrivacyCustomTokenName        string `json:"PrivacyCustomTokenName"`
		PrivacyCustomTokenSymbol      string `json:"PrivacyCustomTokenSymbol"`
		PrivacyCustomTokenData        string `json:"PrivacyCustomTokenData"`
		PrivacyCustomTokenProofDetail struct {
			InputCoins  interface{} `json:"InputCoins"`
			OutputCoins interface{} `json:"OutputCoins"`
		} `json:"PrivacyCustomTokenProofDetail"`
		PrivacyCustomTokenIsPrivacy bool   `json:"PrivacyCustomTokenIsPrivacy"`
		PrivacyCustomTokenFee       int    `json:"PrivacyCustomTokenFee"`
		IsInMempool                 bool   `json:"IsInMempool"`
		IsInBlock                   bool   `json:"IsInBlock"`
		Info                        string `json:"Info"`
	} `json:"Result"`
	Error   interface{} `json:"Error"`
	Params  []string    `json:"Params"`
	Method  string      `json:"Method"`
	Jsonrpc string      `json:"Jsonrpc"`
}

func decodeBase64(s string) []byte {
	res, _ := b64.StdEncoding.DecodeString(s)
	return res
}

func decodeBase58(s string) []byte {
	return base58.Decode(s)
}

func parseCoinDetails(d CoinDetails) *coin.CoinV1 {
	c := new(coin.CoinV1)
	c.SetPublicKey(nil)
	if len(d.PublicKey) != 0 {
		c.SetPublicKey(operation.RandomPoint())
	}
	c.SetCoinCommitment(nil)
	if len(d.CoinCommitment) != 0 {
		c.SetCoinCommitment(operation.RandomPoint())
	}
	c.SetSNDerivator(nil)
	if (d.SNDerivator == EmptyStruct{}) {
		c.SetSNDerivator(operation.RandomScalar())
	}
	c.SetRandomness(nil)
	if (d.Randomness == EmptyStruct{}) {
		c.SetRandomness(operation.RandomScalar())
	}
	c.SetSerialNumber(nil)
	if len(d.SerialNumber) != 0 {
		c.SetSerialNumber(operation.RandomPoint())
	}
	c.SetValue(uint64(d.Value))
	c.SetInfo([]byte(d.Info))
	return c
}

func getStatistic(filename string) {
	fmt.Println("Getting statistic of", filename)
	dat, _ := ioutil.ReadFile(filename)
	// fmt.Println(dat)

	tx := new(Txver2)
	_ = json.Unmarshal(dat, tx)
	proof := decodeBase64(tx.Result.Proof)
	sigPubKey := decodeBase58(tx.Result.SigPubKey)
	sig := decodeBase58(tx.Result.Sig)

	fmt.Printf("Proof length in bytes = %d\n", len(proof))
	fmt.Printf("Signature length in bytes = %d\n", len(sig))
	fmt.Printf("Signature public key in bytes = %d\n", len(sigPubKey))

	inputCoins := tx.Result.ProofDetail.InputCoins
	outputCoins := tx.Result.ProofDetail.OutputCoins

	fmt.Println("-")
	fmt.Println("Number of InputCoins: ", len(inputCoins))
	sumBytesInputCoins := 0
	for i, inp := range inputCoins {
		c := parseCoinDetails(inp.Detail)
		cBytes := c.Bytes()
		lenEnc := len(decodeBase58(inp.CoinDetailsEncrypted))
		fmt.Printf("InputCoin[%d] length in byte = %d\n", i, len(cBytes)+lenEnc)

		sumBytesInputCoins += len(cBytes) + lenEnc
	}
	fmt.Printf("Sum bytes of inputCoins = %d\n", sumBytesInputCoins)
	fmt.Println("-")

	fmt.Println("Number of OutputCoins: ", len(outputCoins))
	sumBytesOutputCoins := 0
	for i, inp := range outputCoins {
		c := parseCoinDetails(inp.Detail)
		cBytes := c.Bytes()
		lenEnc := len(decodeBase58(inp.CoinDetailsEncrypted))
		fmt.Printf("OutputCoin[%d] length in byte = %d\n", i, len(cBytes)+lenEnc)
		sumBytesOutputCoins += len(cBytes) + lenEnc
	}
	fmt.Printf("Sum bytes of outputCoins = %d\n", sumBytesOutputCoins)
	fmt.Println("-")

	fmt.Println("Done ==================")
}

func main() {
	getStatistic("./tx1.json")
	getStatistic("./tx2.json")

	// s, _ := new(big.Int).SetString("1000000000000000000", 10)
	// fmt.Println(len(s.Bytes()))
	// fmt.Println(s.Bytes())
	// fmt.Println("?")
	// a := new(int)
	// *a = 10

	// b := *a

	// c := &b
	// *c = 20

	// fmt.Println(*a)
	// fmt.Println(*c)
}
