package internal

import (
	"errors"
	// "fmt"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	// "github.com/incognitochain/incognito-chain/wallet"

	// "github.com/incognitochain/incognito-chain/metadata"
	// "github.com/pkg/errors"
	// "math/big"
)

func CreateTransaction(args string, serverTime int64) (string, error) {
	params := &InitParamsAsm{}
	println("Before parse")
	println(args)
	err := json.Unmarshal([]byte(args), params)
	if err!=nil{
		println(err.Error())
		return "", err
	}
	println("After parse")
	thoseBytesAgain, _ := json.Marshal(params)
	println(string(thoseBytesAgain))

	var txJson []byte
	if params.TokenParams==nil{			
		tx := &Tx{}
		err = tx.InitASM(params)

		if err != nil {
			println("Can not create tx: ", err.Error())
			return "", err
		}

		// serialize tx json
		txJson, err = json.Marshal(tx)
		if err != nil {
			println("Can not marshal tx: ", err)
			return "", err
		}
	}else{
		tx := &TxToken{}
		err = tx.InitASM(params)

		if err != nil {
			println("Can not create tx: ", err.Error())
			return "", err
		}

		// serialize tx json
		txJson, err = json.Marshal(tx)
		if err != nil {
			println("Error marshalling tx: ", err)
			return "", err
		}
	}

	// lockTimeBytes := common.AddPaddingBigInt(new(big.Int).SetInt64(tx.LockTime), 8)
	// resBytes := append(txJson, lockTimeBytes...)

	res := b58.Encode(txJson, common.ZeroByte)

	return res, nil
}

func DecompressCoins(args string) (string, error) {
	buf := make([][]byte, 1)
	err := json.Unmarshal([]byte(args), &buf)
	if err!=nil{
		println(err.Error())
		return "", err
	}
	
	var cis []CoinInter
	for _, ele := range buf{
		temp := &CoinInter{}
		err := temp.SetBytes(ele)
		if err!=nil{
			println(err.Error)
			return "", err
		}

		cis = append(cis, *temp)
	}
	// serialize tx json
	txJson, err := json.Marshal(cis)
	if err != nil {
		println("Error marshalling coin array: ", err.Error())
		return "", err
	}

	return string(txJson), nil
}

func CacheCoins(coinsString string, indexesString string) (string, error) {
	coinBytesArray := make([][]byte, 1)
	err := json.Unmarshal([]byte(coinsString), &coinBytesArray)
	if err!=nil{
		println(err.Error())
		return "", err
	}

	indexes := make([]uint64, 1)
	err = json.Unmarshal([]byte(indexesString), &indexes)
	if err!=nil{
		println(err.Error())
		return "", err
	}
	if len(coinBytesArray)!=len(indexes){
		return "", errors.New("Mismatched input lengths")
	}
	
	cache := MakeCoinCache()
	var isCA  bool = true
	for i, cb := range coinBytesArray{
		temp := &CoinInter{}
		err := temp.SetBytes(cb)
		if err!=nil{
			println(err.Error)
			return "", err
		}

		cache.PublicKeys 	= append(cache.PublicKeys, temp.PublicKey)
		cache.Commitments 	= append(cache.Commitments, temp.Commitment)
		if temp.AssetTag==nil{
			isCA = false
		}else{
			if isCA{
				cache.AssetTags = append(cache.AssetTags, temp.AssetTag)
			}
		}
		mapKey := b58.Encode(temp.PublicKey, common.ZeroByte)
		cache.PkToIndexMap[mapKey] = indexes[i]
	}
	// serialize tx json

	txJson, err := json.Marshal(cache)
	if err != nil {
		println("Error marshalling coin cache: ", err)
		return "", err
	}

	return string(txJson), nil
}