package types

import (
	"strconv"
	"strings"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
)

const MAX_SHARD_NUMBER = 255

func CreateHashes() ([]common.Hash, []*common.Hash) {
	hashesPointer := []*common.Hash{}
	hashes := []common.Hash{}
	for i := 0; i < 5; i++ {
		newHash, _ := common.Hash{}.NewHashFromStr(strconv.Itoa(i))
		hashesPointer = append(hashesPointer, newHash)
	}

	for i := 0; i < 5; i++ {
		hashes = append(hashes, *hashesPointer[i])
	}
	return hashes, hashesPointer
}
func TestConvertHashAndHashPointer(t *testing.T) {
	hashes, hashesPointer := CreateHashes()
	merkleTree1 := Merkle{}.BuildMerkleTreeOfHashes(hashesPointer, MAX_SHARD_NUMBER)
	tempMerkleTree1 := []common.Hash{}
	for _, value := range merkleTree1 {
		tempMerkleTree1 = append(tempMerkleTree1, *value)
	}
	merkleTree2 := Merkle{}.BuildMerkleTreeOfHashes2(hashes, MAX_SHARD_NUMBER)
	tempMerkleTree2 := []*common.Hash{}
	for _, value := range merkleTree2 {
		newHash, _ := common.Hash{}.NewHashFromStr(value.String())
		tempMerkleTree2 = append(tempMerkleTree2, newHash)
	}

	if len(tempMerkleTree1) != len(merkleTree2) {
		panic("len not compatible")
	}

	if len(tempMerkleTree2) != len(merkleTree1) {
		panic("len not compatible")
	}

	for i := range tempMerkleTree1 {
		if strings.Compare(tempMerkleTree1[i].String(), merkleTree2[i].String()) != 0 {
			panic("value not compatible")
		}
	}

	for i := range tempMerkleTree2 {
		if strings.Compare(tempMerkleTree2[i].String(), merkleTree1[i].String()) != 0 {
			panic("value not compatible")
		}
	}
}

func TestGetMerkleShardPath(t *testing.T) {
	hashes, _ := CreateHashes()
	merkleTree := Merkle{}.BuildMerkleTreeOfHashes2(hashes, MAX_SHARD_NUMBER)
	for shardID := 0; shardID < len(hashes); shardID++ {
		merklePathShard, merkleShardRoot := Merkle{}.GetMerklePathForCrossShard(MAX_SHARD_NUMBER, merkleTree, byte(shardID))
		ok := Merkle{}.VerifyMerkleRootFromMerklePath(hashes[shardID], merklePathShard, merkleShardRoot, byte(shardID))
		if !ok {
			panic("Fail To Verify Shard Merkle Path")
		}
	}
}
