package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

type swapProof struct {
	inst []string

	instPath         []string
	instPathIsLeft   []bool
	instRoot         string
	blkData          string
	blkHash          string
	signerPubkeys    []string
	signerSig        string
	signerPaths      [][]string
	signerPathIsLeft [][]bool
}

// handleGetBeaconSwapProof returns a proof of a new beacon committee (for a given bridge block height)
func (rpcServer RpcServer) handleGetBeaconSwapProof(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBeaconSwapProof params: %+v", params)
	listParams := params.([]interface{})
	height := uint64(listParams[0].(float64))
	bc := rpcServer.config.BlockChain
	db := *rpcServer.config.Database

	// Get bridge block and corresponding beacon blocks
	bridgeBlock, beaconBlocks, err := getShardAndBeaconBlocks(height-1, bc, db)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// Get proof of instruction on bridge
	bridgeInstProof, err := getBeaconSwapProofOnBridge(bridgeBlock, bc, db)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// Get proof of instruction on beacon
	beaconInstProof, err := getBeaconSwapProofOnBeacon(bridgeInstProof.inst, beaconBlocks, db)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// Save instruction as a single slice of byte
	flattenBridgeInst := []byte{}
	for _, part := range bridgeInstProof.inst {
		flattenBridgeInst = append(flattenBridgeInst, []byte(part)...)
	}

	return jsonresult.GetSwapProof{
		Instruction: hex.EncodeToString(flattenBridgeInst),

		BeaconInstPath:         beaconInstProof.instPath,
		BeaconInstPathIsLeft:   beaconInstProof.instPathIsLeft,
		BeaconInstRoot:         beaconInstProof.instRoot,
		BeaconBlkData:          beaconInstProof.blkData,
		BeaconBlkHash:          beaconInstProof.blkHash,
		BeaconSignerPubkeys:    beaconInstProof.signerPubkeys,
		BeaconSignerSig:        beaconInstProof.signerSig,
		BeaconSignerPaths:      beaconInstProof.signerPaths,
		BeaconSignerPathIsLeft: beaconInstProof.signerPathIsLeft,

		BridgeInstPath:         bridgeInstProof.instPath,
		BridgeInstPathIsLeft:   bridgeInstProof.instPathIsLeft,
		BridgeInstRoot:         bridgeInstProof.instRoot,
		BridgeBlkData:          bridgeInstProof.blkData,
		BridgeBlkHash:          bridgeInstProof.blkHash,
		BridgeSignerPubkeys:    bridgeInstProof.signerPubkeys,
		BridgeSignerSig:        bridgeInstProof.signerSig,
		BridgeSignerPaths:      bridgeInstProof.signerPaths,
		BridgeSignerPathIsLeft: bridgeInstProof.signerPathIsLeft,
	}, nil
}

// getShardAndBeaconBlocks returns a shard block (with all of its instructions) and the included beacon blocks
func getShardAndBeaconBlocks(
	height uint64,
	bc *blockchain.BlockChain,
	db database.DatabaseInterface,
) (*blockchain.ShardBlock, []*blockchain.BeaconBlock, error) {
	bridgeID := byte(1) // TODO(@0xbunyip): replace with bridge shardID
	bridgeBlock, err := bc.GetShardBlockByHeight(height, bridgeID)
	if err != nil {
		return nil, nil, err
	}
	beaconBlocks, err := getIncludedBeaconBlocks(
		bridgeBlock.Header.Height,
		bridgeBlock.Header.BeaconHeight,
		bridgeBlock.Header.ShardID,
		bc,
		db,
	)
	if err != nil {
		return nil, nil, err
	}
	bridgeInsts, err := extractInstsFromShardBlock(bridgeBlock, beaconBlocks, bc)
	if err != nil {
		return nil, nil, err
	}
	bridgeBlock.Body.Instructions = bridgeInsts
	return bridgeBlock, beaconBlocks, nil
}

// getBeaconSwapProofOnBridge finds a beacon committee swap instruction in a given bridge block and returns its proof
func getBeaconSwapProofOnBridge(
	bridgeBlock *blockchain.ShardBlock,
	bc *blockchain.BlockChain,
	db database.DatabaseInterface,
) (*swapProof, error) {
	insts := bridgeBlock.Body.Instructions
	_, id := findCommSwapInst(insts, metadata.BeaconPubkeyRootMeta)
	if id < 0 {
		return nil, fmt.Errorf("cannot find beacon swap instruction in bridge block")
	}

	proof, err := buildProofOnBridge(bridgeBlock, insts, id, db)
	if err != nil {
		return nil, err
	}
	return proof, nil
}

// buildProofOnBridge builds a swapProof for an instruction in a bridge block
func buildProofOnBridge(
	bridgeBlock *blockchain.ShardBlock,
	insts [][]string,
	id int,
	db database.DatabaseInterface,
) (*swapProof, error) {
	// Get committee pubkey and signature
	pubkeys, signerIdxs, err := getBridgeSignerPubkeys(bridgeBlock, db)
	if err != nil {
		return nil, err
	}
	signerPubkeys := make([]string, len(pubkeys))
	for i, pk := range pubkeys {
		signerPubkeys[i] = hex.EncodeToString(pk)
	}

	// Build merkle proof for signer pubkeys
	signerProof := buildSignersProof(pubkeys, signerIdxs)
	signerPaths := make([][]string, len(pubkeys))
	signerPathIsLeft := make([][]bool, len(pubkeys))
	for i, p := range signerProof {
		signerPaths[i] = p.getPath()
		signerPathIsLeft[i] = p.left
	}

	// Build merkle proof for instruction in bridge block
	instProof := buildInstProof(insts, id)

	// Get meta hash and block hash
	instRoot := hex.EncodeToString(bridgeBlock.Header.InstructionMerkleRoot[:])
	metaHash := bridgeBlock.Header.MetaHash()
	blkHash := bridgeBlock.Header.Hash()

	return &swapProof{
		inst:             insts[id],
		instPath:         instProof.getPath(),
		instPathIsLeft:   instProof.left,
		instRoot:         instRoot,
		blkData:          hex.EncodeToString(metaHash[:]),
		blkHash:          hex.EncodeToString(blkHash[:]),
		signerPubkeys:    signerPubkeys,
		signerSig:        bridgeBlock.AggregatedSig,
		signerPaths:      signerPaths,
		signerPathIsLeft: signerPathIsLeft,
	}, nil
}

// buildProofOnBeacon builds a swapProof for an instruction in a beacon block
func buildProofOnBeacon(
	beaconBlock *blockchain.BeaconBlock,
	insts [][]string,
	id int,
	db database.DatabaseInterface,
) (*swapProof, error) {
	// Get committee pubkey and signature
	pubkeys, signerIdxs, err := getBeaconSignerPubkeys(beaconBlock, db)
	if err != nil {
		return nil, err
	}
	signerPubkeys := make([]string, len(pubkeys))
	for i, pk := range pubkeys {
		signerPubkeys[i] = hex.EncodeToString(pk)
	}

	// Build merkle proof for signer pubkeys
	signerProof := buildSignersProof(pubkeys, signerIdxs)
	signerPaths := make([][]string, len(pubkeys))
	signerPathIsLeft := make([][]bool, len(pubkeys))
	for i, p := range signerProof {
		signerPaths[i] = p.getPath()
		signerPathIsLeft[i] = p.left
	}

	// Build merkle proof for instruction in beacon block
	instProof := buildInstProof(insts, id)

	// Get meta hash and block hash
	instRoot := hex.EncodeToString(beaconBlock.Header.InstructionMerkleRoot[:])
	metaHash := beaconBlock.Header.MetaHash()
	blkHash := beaconBlock.Header.Hash()

	return &swapProof{
		inst:             insts[id],
		instPath:         instProof.getPath(),
		instPathIsLeft:   instProof.left,
		instRoot:         instRoot,
		blkData:          hex.EncodeToString(metaHash[:]),
		blkHash:          hex.EncodeToString(blkHash[:]),
		signerPubkeys:    signerPubkeys,
		signerSig:        beaconBlock.AggregatedSig,
		signerPaths:      signerPaths,
		signerPathIsLeft: signerPathIsLeft,
	}, nil
}

// getBeaconSwapProofOnBeacon finds in given beacon blocks a beacon committee swap instruction and returns its proof
func getBeaconSwapProofOnBeacon(
	inst []string,
	beaconBlocks []*blockchain.BeaconBlock,
	db database.DatabaseInterface,
) (*swapProof, error) {
	// Get beacon block and check if it contains beacon swap instruction
	beaconBlock := findBeaconBlockWithInst(beaconBlocks, inst)
	if beaconBlock == nil {
		return nil, fmt.Errorf("cannot find corresponding beacon block that includes swap instruction")
	}

	insts := beaconBlock.Body.Instructions
	_, instID := findCommSwapInst(insts, metadata.BeaconPubkeyRootMeta)
	if instID < 0 {
		return nil, fmt.Errorf("cannot find beacon swap instruction in beacon block")
	}

	return buildProofOnBeacon(beaconBlock, insts, instID, db)
}

// getIncludedBeaconBlocks retrieves all beacon blocks included in a shard block
func getIncludedBeaconBlocks(
	shardHeight uint64,
	beaconHeight uint64,
	shardID byte,
	bc *blockchain.BlockChain,
	db database.DatabaseInterface,
) ([]*blockchain.BeaconBlock, error) {
	prevShardBlock, err := bc.GetShardBlockByHeight(shardHeight-1, shardID)
	if err != nil {
		return nil, err
	}
	beaconBlocks, err := blockchain.FetchBeaconBlockFromHeight(
		db,
		prevShardBlock.Header.BeaconHeight+1,
		beaconHeight,
	)
	if err != nil {
		return nil, err
	}
	return beaconBlocks, nil
}

// extractInstsFromShardBlock returns all instructions in a shard block as a slice of []string
func extractInstsFromShardBlock(
	shardBlock *blockchain.ShardBlock,
	beaconBlocks []*blockchain.BeaconBlock,
	bc *blockchain.BlockChain,
) ([][]string, error) {
	instructions, err := blockchain.CreateShardInstructionsFromTransactionAndIns(
		shardBlock.Body.Transactions,
		bc,
		shardBlock.Header.ShardID,
		&shardBlock.Header.ProducerAddress,
		shardBlock.Header.Height,
		beaconBlocks,
		shardBlock.Header.BeaconHeight,
	)
	if err != nil {
		return nil, err
	}
	shardInsts := append(instructions, shardBlock.Body.Instructions...)
	return shardInsts, nil
}

// findCommSwapInst finds a swap instruction in a list, returns it along with its index
func findCommSwapInst(insts [][]string, meta int) ([]string, int) {
	for i, inst := range insts {
		if strconv.Itoa(meta) == inst[0] {
			fmt.Println("[db] BeaconPubkeyRootMeta inst:", inst)
			return inst, i
		}
	}
	return nil, -1
}

type keccak256MerkleProof struct {
	path [][]byte
	left []bool
}

// getPath encodes the path of merkle proof as string and returns
func (p *keccak256MerkleProof) getPath() []string {
	path := make([]string, len(p.path))
	for i, h := range p.path {
		path[i] = hex.EncodeToString(h)
	}
	return path
}

// buildProof builds a merkle proof for one element in a merkle tree
func buildProofFromTree(merkles [][]byte, id int) *keccak256MerkleProof {
	path, left := blockchain.GetKeccak256MerkleProofFromTree(merkles, id)
	return &keccak256MerkleProof{path: path, left: left}
}

// buildProof receives a list of data (as bytes) and returns a merkle proof for one element in the list
func buildProof(data [][]byte, id int) *keccak256MerkleProof {
	merkles := blockchain.BuildKeccak256MerkleTree(data)
	return buildProofFromTree(merkles, id)
}

// buildInstProof receives a list of instructions (as string) and returns a merkle proof for one instruction in the list
func buildInstProof(insts [][]string, id int) *keccak256MerkleProof {
	flattenInsts := blockchain.FlattenAndConvertStringInst(insts)
	fmt.Printf("[db] insts: %v\n", insts)
	fmt.Printf("[db] flattenInsts: %x\n", flattenInsts)
	return buildProof(flattenInsts, id)
}

// getBeaconSignerPubkeys finds the pubkeys of all signers of a beacon block
func getBeaconSignerPubkeys(beaconBlock *blockchain.BeaconBlock, db database.DatabaseInterface) ([][]byte, []int, error) {
	commsRaw, err := db.FetchBeaconCommitteeByEpoch(beaconBlock.Header.Epoch)
	if err != nil {
		return nil, nil, err
	}

	comm := []string{}
	err = json.Unmarshal(commsRaw, &comm)
	if err != nil {
		return nil, nil, err
	}

	signerIdxs := beaconBlock.ValidatorsIdx[1] // List of signers
	pubkeys := make([][]byte, len(signerIdxs))
	for i, signerID := range signerIdxs {
		pubkey, _, err := base58.Base58Check{}.Decode(comm[signerID])
		if err != nil {
			return nil, nil, err
		}
		pubkeys[i] = pubkey
	}
	return pubkeys, signerIdxs, nil
}

// getBridgeSignerPubkeys finds the pubkeys of all signers of a shard block
func getBridgeSignerPubkeys(shardBlock *blockchain.ShardBlock, db database.DatabaseInterface) ([][]byte, []int, error) {
	commsRaw, err := db.FetchCommitteeByEpoch(shardBlock.Header.Epoch)
	if err != nil {
		return nil, nil, err
	}

	comms := make(map[byte][]string)
	err = json.Unmarshal(commsRaw, &comms)
	if err != nil {
		return nil, nil, err
	}

	comm, ok := comms[shardBlock.Header.ShardID]
	if !ok {
		return nil, nil, fmt.Errorf("no committee member found for shard block %d", shardBlock.Header.ShardID)
	}

	signerIdxs := shardBlock.ValidatorsIdx[1] // List of signers
	pubkeys := make([][]byte, len(signerIdxs))
	for i, signerID := range signerIdxs {
		pubkey, _, err := base58.Base58Check{}.Decode(comm[signerID])
		if err != nil {
			return nil, nil, err
		}
		pubkeys[i] = pubkey
	}
	return pubkeys, signerIdxs, nil
}

// buildSignersProof builds the merkle proofs for some elements in a list of pubkeys
func buildSignersProof(pubkeys [][]byte, idxs []int) []*keccak256MerkleProof {
	merkles := blockchain.BuildKeccak256MerkleTree(pubkeys)
	fmt.Printf("[db] pubkeys: %x\n", pubkeys)
	fmt.Printf("[db] merkles: %x\n", merkles)
	proofs := make([]*keccak256MerkleProof, len(pubkeys))
	for i, pid := range idxs {
		proofs[i] = buildProofFromTree(merkles, pid)
	}
	return proofs
}

// findBeaconBlockWithInst finds a beacon block with a specific instruction; nil if not found
func findBeaconBlockWithInst(beaconBlocks []*blockchain.BeaconBlock, inst []string) *blockchain.BeaconBlock {
	for _, b := range beaconBlocks {
		for _, blkInst := range b.Body.Instructions {
			diff := false
			for i, part := range blkInst {
				if part != inst[i] {
					diff = true
					break
				}
			}
			if !diff {
				return b
			}
		}
	}
	return nil
}
