package client

import (
	"fmt"
	"testing"

	"github.com/ninjadotorg/cash-prototype/common"
)

func buildByteSequence(k int) [][]byte {
	var bytes [][]byte
	for i := 1; i <= k; i++ {
		t := [32]byte{byte(i)}
		bytes = append(bytes, t[:])
	}
	return bytes
}

func buildMerkleHashSequence(k int) []MerkleHash {
	var hashes []MerkleHash
	bytes := buildByteSequence(k)
	for _, b := range bytes {
		var h MerkleHash = make([]byte, 32)
		copy(h, b)
		hashes = append(hashes, h)
	}
	return hashes
}

func TestAddNewNode(t *testing.T) {
	var tree = new(IncMerkleTree)
	hashes := buildMerkleHashSequence(7)
	fmt.Println("initial tree:", tree)
	for _, hash := range hashes {
		tree.AddNewNode(hash)
	}
	fmt.Println("new tree:", tree)

	fmt.Println("old hashes:")
	for _, hash := range hashes {
		fmt.Println(hash)
	}

	rt := tree.GetRoot(common.IncMerkleTreeHeight)
	fmt.Println("rt:", rt)
}

func TestBuildWitness(t *testing.T) {
	var tree = new(IncMerkleTree)
	n := 7
	bytes := buildByteSequence(n)
	for _, b := range bytes {
		tree.AddNewNode(b)
	}
	notes := []*JSInput{&JSInput{InputNote: &Note{Cm: bytes[1]}}}
	err := BuildWitnessPath(notes, bytes)
	if err != nil {
		t.Errorf("error: %s", err.Error())
	}
	// fmt.Printf("left: %x\n", tree.left)
	// fmt.Printf("right: %x\n", tree.right)
	// fmt.Printf("nodes: %v\n", tree.nodes)
	fmt.Printf("witness path: %x\n\n", notes[0].WitnessPath.AuthPath)
	fmt.Printf("2+3: %x\n", combineAndHash(bytes[2], bytes[3]))
	fmt.Printf("4+5: %x\n", combineAndHash(bytes[4], bytes[5]))
	fmt.Printf("6+nil: %x\n", combineAndHash(bytes[6], uncommittedNodeHash))
	fmt.Printf("(4+5)+(6+nil): %x\n", combineAndHash(combineAndHash(bytes[4], bytes[5]), combineAndHash(bytes[6], uncommittedNodeHash)))
	fmt.Printf("padding@3: %x\n", getPaddingAtDepth(27))
}
