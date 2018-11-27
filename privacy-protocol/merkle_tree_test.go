package privacy

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ninjadotorg/constant/common"
)

// buildByteSequence creates bytes array with array size is k
// and value of element at index i is i
func buildByteSequence(k int) [][]byte {
	var bytes [][]byte
	for i := 1; i <= k; i++ {
		t := [32]byte{byte(i)}
		bytes = append(bytes, t[:])
	}
	return bytes
}

// buildMerkleHashSequence creates MerkleHash array with array size is k
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
	fmt.Printf("rt: %v\n", rt)
}

//func TestBuildWitness(t *testing.T) {
//	var tree = new(IncMerkleTree)
//	n := 7
//	bytes := buildByteSequence(n)
//	for _, b := range bytes {
//		tree.AddNewNode(b)
//	}
//	notes := []*JSInput{&JSInput{InputNote: &Note{Cm: bytes[1]}}}
//	err := BuildWitnessPath(notes, bytes)
//	if err != nil {
//		t.Errorf("error: %s", err.Error())
//	}
//	// fmt.Printf("left: %x\n", tree.left)
//	// fmt.Printf("right: %x\n", tree.right)
//	// fmt.Printf("nodes: %v\n", tree.nodes)
//	fmt.Printf("witness path: %x\n\n", notes[0].WitnessPath.AuthPath)
//	fmt.Printf("2+3: %x\n", combineAndHash(bytes[2], bytes[3]))
//	fmt.Printf("4+5: %x\n", combineAndHash(bytes[4], bytes[5]))
//	fmt.Printf("6+nil: %x\n", combineAndHash(bytes[6], uncommittedNodeHash))
//	fmt.Printf("(4+5)+(6+nil): %x\n", combineAndHash(combineAndHash(bytes[4], bytes[5]), combineAndHash(bytes[6], uncommittedNodeHash)))
//	fmt.Printf("padding@3: %x\n", getPaddingAtDepth(27))
//}

func TestMerkleTreeRoot(t *testing.T) {
	cmHex := []string{
		"d26356e6f726dfb4c0a395f3af134851139ce1c64cfed3becc3530c8c8ad5660",
		"5aaf71f995db014006d630dedf7ffcbfa8854055e6a8cc9ef153629e3045b7e1",
		"5349bfb53f3720042268ab0d6d617bf56ad54042f956a8869ae59d9ef89801bd",
		"6a012707716b08c7430508a23e0fcb105373be8df5383d0c181566ed5f44ef5e",
		"d52b90e834aa7c3d5add4ee71daf005cf9756acd8a6a87992174069401ce66b4",
		"2daae4da90d52734a611de096d160ff0f7bc03f1f05a4145b5dedaae84acf637",
		"e22dc31f03ff60a1ea3fb3503b5761681d8d0e74f19709338321e292750d88bc",
		"a21b653d8c6f218bc75512ba51fd9180cb6e886f2661d6e16812f0b5dbaa79a1",
		"c910cd2529c24894f5bc99dc1fecf86dad19cb09817169671e02d64fba91451b",
		"2ca641d9046821932fb74376184d769d9aa3ae1947c0207251aae67da471772b",
	}
	commitments := [][]byte{}
	for _, cm := range cmHex {
		c, err := hex.DecodeString(cm)
		if err != nil {
			t.Error(err)
		}
		commitments = append(commitments, c)
	}

	tree := &IncMerkleTree{}
	for _, cm := range commitments {
		tree.AddNewNode(cm)
	}

	rt := tree.GetRoot(common.IncMerkleTreeHeight)
	fmt.Printf("Root: %x\n", rt)
}
