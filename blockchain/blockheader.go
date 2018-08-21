package blockchain

import (
	"time"
	"crypto/sha256"
	"fmt"

	"github.com/ninjadotorg/money-prototype/common"
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

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint32 on the wire and therefore is limited to 2106.
	Timestamp time.Time

	// Difficulty target for the block.
	Difficulty int

	// Nonce used to generate the block.
	Nonce int
}

/**
 BlockHash computes the block identifier hash for the given block header.
 */
func (h BlockHeader) BlockHash() (common.Hash) {
	record := fmt.Sprint(h.Version) + h.Timestamp.String() + h.MerkleRoot.String() + h.PrevBlockHash.String() + fmt.Sprint(h.Nonce)
	hash256 := sha256.New()
	hash256.Write([]byte(record))
	hashed := hash256.Sum(nil)
	hash, _ := common.Hash{}.NewHash(hashed)
	return *hash
}
