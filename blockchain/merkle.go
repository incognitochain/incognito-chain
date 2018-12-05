package blockchain

import (
	"fmt"
	"math"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy-protocol/client"
	"github.com/ninjadotorg/constant/transaction"
)

type Merkle struct {
}

// BuildMerkleTreeStore creates a merkle tree from a slice of transactions,
// stores it using a linear array, and returns a slice of the backing array.  A
// linear array was chosen as opposed to an actual tree structure since it uses
// about half as much memory.  The following describes a merkle tree and how it
// is stored in a linear array.
//
// A merkle tree is a tree in which every non-leaf node is the hash of its
// children nodes.  A diagram depicting how this works for constant transactions
// where h(x) is a double sha256 follows:
//
//	         root = h1234 = h(h12 + h34)
//	        /                           \
//	  h12 = h(h1 + h2)            h34 = h(h3 + h4)
//	   /            \              /            \
//	h1 = h(tx1)  h2 = h(tx2)    h3 = h(tx3)  h4 = h(tx4)
//
// The above stored as a linear array is as follows:
//
// 	[h1 h2 h3 h4 h12 h34 root]
//
// As the above shows, the merkle root is always the last element in the array.
//
// The number of inputs is not always a power of two which results in a
// balanced tree structure as above.  In that case, parent nodes with no
// children are also zero and parent nodes with only a single left node
// are calculated by concatenating the left node with itself before hashing.
// Since this function uses nodes that are pointers to the hashes, empty nodes
// will be nil.
//
// The additional bool parameter indicates if we are generating the merkle tree
// using witness transaction id's rather than regular transaction id's. This
// also presents an additional case wherein the wtxid of the salary transaction
// is the zeroHash.
func (self Merkle) BuildMerkleTreeStore(transactions []metadata.Transaction) []*common.Hash {
	// Calculate how many entries are required to hold the binary merkle
	// tree as a linear array and create an array of that size.
	nextPoT := self.nextPowerOfTwo(len(transactions))
	arraySize := nextPoT*2 - 1
	merkles := make([]*common.Hash, arraySize)

	// Create the base transaction hashes and populate the array with them.
	for i, tx := range transactions {
		// If we're computing a witness merkle root, instead of the
		// regular txid, we use the modified wtxid which includes a
		// transaction's witness data within the digest. Additionally,
		// the salary's wtxid is all zeroes.
		witness := false
		switch {
		case witness && i == 0:
			var zeroHash common.Hash
			merkles[i] = &zeroHash
		case witness:
			//wSha := tx.MsgTx().WitnessHash()
			//merkles[i] = &wSha
			continue
		default:
			merkles[i] = tx.Hash()
		}

	}

	// Start the array offset after the last transaction and adjusted to the
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
			newHash := self.hashMerkleBranches(merkles[i], merkles[i])
			merkles[offset] = newHash

			// The normal case sets the parent node to the double sha256
			// of the concatentation of the left and right children.
		default:
			newHash := self.hashMerkleBranches(merkles[i], merkles[i+1])
			merkles[offset] = newHash
		}
		offset++
	}

	return merkles
}

// nextPowerOfTwo returns the next highest power of two from a given number if
// it is not already a power of two.  This is a helper function used during the
// calculation of a merkle tree.
func (self Merkle) nextPowerOfTwo(n int) int {
	// Return the number if it's already a power of 2.
	if n&(n-1) == 0 {
		return n
	}

	// Figure out and return the next power of two.
	exponent := uint(math.Log2(float64(n))) + 1
	return 1 << exponent // 2^exponent
}

/*
hashMerkleBranches takes two hashes, treated as the left and right tree
nodes, and returns the hash of their concatenation.  This is a helper
function used to aid in the generation of a merkle tree.
*/
func (self Merkle) hashMerkleBranches(left *common.Hash, right *common.Hash) *common.Hash {
	// Concatenate the left and right nodes.
	var hash [common.HashSize * 2]byte
	copy(hash[:common.HashSize], left[:])
	copy(hash[common.HashSize:], right[:])

	newHash := common.DoubleHashH(hash[:])
	return &newHash
}

/*
// UpdateMerkleTreeForBlock adds all transaction's commitments in a block to the newest merkle tree
*/
func UpdateMerkleTreeForBlock(tree *client.IncMerkleTree, block *Block) error {
	for _, blockTx := range block.Transactions {
		if blockTx.GetType() == common.TxNormalType || blockTx.GetType() == common.TxSalaryType {
			tx, ok := blockTx.(*transaction.Tx)
			if ok == false {
				return NewBlockChainError(UnExpectedError, fmt.Errorf("Transaction in block not valid"))
			}

			for _, desc := range tx.Descs {
				for _, cm := range desc.Commitments {
					tree.AddNewNode(cm[:])
				}
			}
		}
	}
	return nil
}
