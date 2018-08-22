package client

type MerkleHash []byte

type MerklePath struct {
	AuthPath []*MerkleHash
	Index    []bool
}
