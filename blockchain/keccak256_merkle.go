package blockchain

import "github.com/incognitochain/incognito-chain/common"

// BuildKeccak256MerkleTree creates a merkle tree using Keccak256 hash func.
// This merkle tree is used for storing all beacon (and bridge) data to relay them to Ethereum.
func BuildKeccak256MerkleTree(data [][]byte) [][]byte {
	if len(data) == 0 {
		emptyRoot := [32]byte{}
		return [][]byte{emptyRoot[:]}
	}
	// Calculate how many entries are required to hold the binary merkle
	// tree as a linear array and create an array of that size.
	nextPoT := NextPowerOfTwo(len(data))
	arraySize := nextPoT*2 - 1
	merkles := make([][]byte, arraySize)

	// Create the base data hashes and populate the array with them.
	for i, d := range data {
		h := common.Keccak256(d)
		merkles[i] = h[:]
	}

	// Start the array offset after the last data and adjusted to the
	// next power of two.
	offset := nextPoT
	for i := 0; i < arraySize-1; i += 2 {
		switch {
		// When there is no left child node, the parent is nil too.
		case merkles[i] == nil:
			merkles[offset] = nil

			// When there is no right child, the parent is generated by
			// hashing the concatenation of the left child with itself.
		case merkles[i+1] == nil:
			newHash := keccak256MerkleBranches(merkles[i], merkles[i])
			merkles[offset] = newHash

			// The normal case sets the parent node to the keccak256
			// of the concatentation of the left and right children.
		default:
			newHash := keccak256MerkleBranches(merkles[i], merkles[i+1])
			merkles[offset] = newHash
		}
		offset++
	}

	return merkles
}

func GetKeccak256MerkleRoot(data [][]byte) []byte {
	merkles := BuildKeccak256MerkleTree(data)
	return merkles[len(merkles)-1]
}

// keccak256MerkleBranches concatenates the 2 branches of a Merkle tree and hash it to create the parent node using Keccak256 hash function
func keccak256MerkleBranches(left []byte, right []byte) []byte {
	// Concatenate the left and right nodes.
	hash := append(left, right...)
	newHash := common.Keccak256(hash)
	return newHash[:]
}
