package client

import (
	"fmt"
	"testing"
)

func TestAddNewNode(t *testing.T) {
	var tree *IncMerkleTree = new(IncMerkleTree)
	fmt.Println("initial tree:", tree)
	var hashes []MerkleHash
	for i := 1; i <= 7; i++ {
		t := [32]byte{byte(i)}
		hash := MerkleHash(t[:])
		hashes = append(hashes, hash)
	}

	for _, hash := range hashes {
		tree.AddNewNode(hash)
	}
	fmt.Println("new tree:", tree)

	fmt.Println("old hashes:")
	for _, hash := range hashes {
		fmt.Println(hash)
	}

	rt := tree.GetRoot()
	fmt.Println("rt:", rt)
}
