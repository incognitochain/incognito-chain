package blockchain

import (
	"time"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/database"
)

// MaxBlockHeaderPayload is the maximum number of bytes a block header can be.
// Version 4 bytes + Timestamp 4 bytes + Bits 4 bytes + Nonce 4 bytes +
// PrevBlockHash and MerkleRoot hashes.
const MaxBlockHeaderPayload = 16 + (common.HashSize * 2)

type BlockHeader struct {
	// Version of the block.  This is not the same as the protocol version.
	Version int

	Db database.DB
	// Hash of the previous block header in the block chain.
	PrevBlockHash string

	// Merkle tree reference to hash of all transactions for the block.
	MerkleRoot common.Hash

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint32 on the wire and therefore is limited to 2106.
	Timestamp time.Time

	// Difficulty target for the block.
	Difficulty int

	// Nonce used to generate the block.
	Nonce int
}

func (h *BlockChain) GetBlockByHash(hash common.Hash) *Block {
	return nil
}
