package jsonresult

type GetBeaconSwapProof struct {
	Instruction            string     // Hex encoded of the swap inst
	BeaconInstPath         []string   // Hex encoded path of the inst in inst merkle tree
	BeaconInstPathIsLeft   []bool     // Indicate whether each path is left or right node
	BeaconInstRoot         string     // Hex encoded root of the inst merkle tree
	BeaconBlkData          string     // Hex encoded hash of the block meta
	BeaconBlkHash          string     // Hex encoded block hash
	BeaconSignerPubkeys    []string   // Hex encoded pubkeys of all signers
	BeaconSignerSig        string     // Hex encoded signature
	BeaconSignerPaths      [][]string // Hex encoded path of each pubkey in pubkey merkle tree for each signer
	BeaconSignerPathIsLeft [][]bool   // Indicate whether each signer's path is left or right node
}
