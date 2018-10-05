package blockchain

import (
	"github.com/ninjadotorg/cash-prototype/common"
)

type BlockHeader struct {
	// Version of the block.  This is not the same as the protocol version.
	Version int
	// Hash of the previous block header in the block chain.
	PrevBlockHash common.Hash

	// Merkle tree reference to hash of all transactions for the block.
	MerkleRoot common.Hash

	// Merkle tree reference to hash of all commitments to the current block.
	MerkleRootCommitments common.Hash

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint64 on the wire and therefore is limited to 2106.
	Timestamp int64

	// POS
	// BlockCommitteeSigs []string          //Include sealer and validators signature
	// Committee          []string          //Voted committee for the next block
	CommitteeSigs map[string]string // Committee and its sigs
	// Parallel PoS
	ChainID      byte
	ChainsHeight []int //height of 20 chain when this block is created
}
