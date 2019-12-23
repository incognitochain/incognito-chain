package statedb_test

import (
	"github.com/incognitochain/incognito-chain/common"
	"math/rand"
	"strconv"
)

func generateTokenIDs(max int) []common.Hash {
	hashes := []common.Hash{}
	for i := 0; i < max; i++ {
		temp := []byte(strconv.Itoa(i))
		hashes = append(hashes, common.HashH(temp))
	}
	return hashes
}

func generateSerialNumberList(max int) [][]byte {
	list := [][]byte{}
	for i := 0; i < max; i++ {
		temp := []byte{}
		for j := 0; j < 32; j++ {
			v := byte(rand.Int() % 256)
			temp = append(temp, v)
		}
		list = append(list, temp)
	}
	return list
}

func generateSNDList(max int) [][]byte {
	list := [][]byte{}
	for i := 0; i < max; i++ {
		temp := []byte{}
		for j := 0; j < 32; j++ {
			v := byte(rand.Int() % 256)
			temp = append(temp, v)
		}
		list = append(list, temp)
	}
	return list
}

func generateCommitmentList(max int) [][]byte {
	list := [][]byte{}
	for i := 0; i < max; i++ {
		temp := []byte{}
		for j := 0; j < 32; j++ {
			v := byte(rand.Int() % 256)
			temp = append(temp, v)
		}
		list = append(list, temp)
	}
	return list
}

func generatePublicKeyList(max int) [][]byte {
	list := [][]byte{}
	for i := 0; i < max; i++ {
		temp := []byte{}
		for j := 0; j < 32; j++ {
			v := byte(rand.Int() % 256)
			temp = append(temp, v)
		}
		list = append(list, temp)
	}
	return list
}

func generateOutputCoinList(max int) [][]byte {
	list := [][]byte{}
	for i := 0; i < max; i++ {
		temp := []byte{}
		for j := 0; j < 100; j++ {
			v := byte(rand.Int() % 256)
			temp = append(temp, v)
		}
		list = append(list, temp)
	}
	return list
}

func generateTokenMapWithAmount() map[common.Hash]uint64 {
	reward := make(map[common.Hash]uint64)
	for _, temp := range tokenIDs {
		tokenID := common.BytesToHash([]byte(temp))
		reward[tokenID] = uint64(rand.Int() % 1000000000)
	}
	return reward
}

func generatePunishedDuration() uint8 {
	return uint8(rand.Int() % 256)
}
