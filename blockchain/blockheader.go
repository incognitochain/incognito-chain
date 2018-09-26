package blockchain

import (
	"github.com/ninjadotorg/cash-prototype/common"
)

// MaxBlockHeaderPayload is the maximum number of bytes a block header can be.
// Version 4 bytes + Timestamp 4 bytes + Bits 4 bytes + Nonce 4 bytes +
// PrevBlockHash and MerkleRoot hashes.
const MaxBlockHeaderPayload = 16 + (common.HashSize * 2)

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

	// Difficulty target for the block.
	Difficulty uint32

	// Nonce used to generate the block.
	Nonce int

	// POS
	BlockCommitteeSigs []string //Include sealer and validators signature
	Committee          []string //Voted committee for the next block
	// Parallel PoS
	ChainID      byte
	ChainsHeight []int //height of 20 chain when this block is created
}
