package jsonresult

type GetBeaconSwapProof struct {
	Instruction            string     // Hex encoded of the swap inst
	BridgeInstPath         []string   // Hex encoded path of the inst in inst merkle tree
	BridgeInstPathIsLeft   []bool     // Indicate whether each path is left or right node
	BridgeInstRoot         string     // Hex encoded root of the inst merkle tree
	BridgeBlkData          string     // Hex encoded hash of the block meta
	BridgeBlkHash          string     // Hex encoded block hash
	BridgeSignerPubkeys    []string   // Hex encoded pubkeys of all signers
	BridgeSignerSig        string     // Hex encoded signature
	BridgeSignerPaths      [][]string // Hex encoded path of each pubkey in pubkey merkle tree for each signer
	BridgeSignerPathIsLeft [][]bool   // Indicate whether each signer's path is left or right node
}
