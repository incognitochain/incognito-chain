package jsonresult

type GetInstructionProof struct {
	Instruction  string // Hex-encoded swap inst
	BeaconHeight string // Hex encoded height of the block contains the inst
	BridgeHeight string

	BeaconInstPath       []string // Hex encoded path of the inst in merkle tree
	BeaconInstPathIsLeft []bool   // Indicate if it is the left or right node
	BeaconInstRoot       string   // Hex encoded root of the inst merkle tree
	BeaconBlkData        string   // Hex encoded hash of the block meta
	BeaconBlkHash        string   // Hex encoded block hash
	BeaconSignerSig      string   // Hex encoded signature
	BeaconPubkeys        []string // To decompress and send to contract
	BeaconRIdxs          []int    // Idxs of R's aggregators
	BeaconSigIdxs        []int    // Idxs of signer
	BeaconR              string   // Random number (33 bytes)

	BridgeInstPath       []string
	BridgeInstPathIsLeft []bool
	BridgeInstRoot       string
	BridgeBlkData        string
	BridgeBlkHash        string
	BridgeSignerSig      string
	BridgePubkeys        []string
	BridgeRIdxs          []int
	BridgeSigIdxs        []int
	BridgeR              string
}
