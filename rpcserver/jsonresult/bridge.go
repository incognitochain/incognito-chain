package jsonresult

type GetInstructionProof struct {
	Instruction  string // Hex-encoded swap inst
	BeaconHeight string // Hex encoded height of the block contains the inst
	BridgeHeight string

	BeaconInstPath       []string // Hex encoded path of the inst in merkle tree
	BeaconInstPathIsLeft []bool   // Indicate if it is the left or right node
	BeaconInstRoot       string   // Hex encoded root of the inst merkle tree
	BeaconBlkData        string   // Hex encoded hash of the block meta
	BeaconSigs           []string // Hex encoded signature (r, s, v)
	BeaconSigIdxs        []int    // Idxs of signer

	BridgeInstPath       []string
	BridgeInstPathIsLeft []bool
	BridgeInstRoot       string
	BridgeBlkData        string
	BridgeSigs           []string
	BridgeSigIdxs        []int
}
