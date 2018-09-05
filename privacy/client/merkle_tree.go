package client

import (
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/privacy/client/crypto/sha256"
)

type MerkleHash []byte

type MerklePath struct {
	AuthPath []*MerkleHash
	Index    []bool
}

var UncommittedNodeHash = [32]byte{}
var arrayPadding []MerkleHash

type IncMerkleTree struct {
	node []MerkleHash
}

func getPaddingAtDepth(d int) MerkleHash {
	if len(arrayPadding) >= common.IncMerkleTreeHeight {
		return arrayPadding[d]
	}
	var paddings []MerkleHash
	paddings = append(paddings, UncommittedNodeHash[:])
	lastHash := UncommittedNodeHash
	for i := 1; i < common.IncMerkleTreeHeight; i++ {
		data := append(lastHash[:], lastHash[:]...)
		lastHash := sha256.Sum256NoPad(data)
		paddings = append(paddings, lastHash[:])
	}
	arrayPadding = paddings
	return arrayPadding[d]
}

// AddNewNode incrementally adds a new leaf node data to the merkle tree
func (tree *IncMerkleTree) AddNewNode(hash MerkleHash) {
	prevHash := hash
	leftOverHash := true // Found new hash for parent node, push to array if no parents found
	for d, node := range tree.node {
		if node == nil {
			tree.node[d] = prevHash // Save hash for current depth
			leftOverHash = false    // Hash already saved in array
			break
		}

		data := append(node, prevHash...)
		tree.node[d] = nil // Clear data at current depth for next hashes
		newHash := sha256.Sum256NoPad(data)
		prevHash = newHash[:]

	}
	if leftOverHash {
		if len(tree.node) >= common.IncMerkleTreeHeight {
			panic("Too many hashes for IncMerkleTree")
		}
		tree.node = append(tree.node, prevHash)
	}
}

// GetRoot returns the merkle root of an unfinished tree; empty nodes are padded with default values
func (tree *IncMerkleTree) GetRoot() MerkleHash {
	prevHash := UncommittedNodeHash
	for i := 0; i < common.IncMerkleTreeHeight; i++ {
		var data []byte
		if i >= len(tree.node) || tree.node[i] == nil {
			data = append(prevHash[:], getPaddingAtDepth(i)[:]...) // Previous hash is the left child
		} else {
			data = append(tree.node[i], prevHash[:]...) // Previous hash is the right child
		}
		prevHash = sha256.Sum256NoPad(data)
	}
	return MerkleHash(prevHash[:])
}
