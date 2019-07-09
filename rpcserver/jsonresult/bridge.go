package jsonresult

type GetInstructionProof struct {
	Instruction  string // Hex-encoded swap inst
	BeaconHeight string // Hex encoded beacon height of the block contains the inst
	BridgeHeight string

	BeaconInstPath       []string // Hex encoded path of the inst in inst merkle tree
	BeaconInstPathIsLeft []bool   // Indicate whether each path is left or right node
	BeaconInstRoot       string   // Hex encoded root of the inst merkle tree
	BeaconBlkData        string   // Hex encoded hash of the block meta
	BeaconBlkHash        string   // Hex encoded block hash
	BeaconSignerSig      string   // Hex encoded signature

	BridgeInstPath       []string
	BridgeInstPathIsLeft []bool
	BridgeInstRoot       string
	BridgeBlkData        string
	BridgeBlkHash        string
	BridgeSignerSig      string
}
