package rpcserver

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

type swapProof struct {
	inst []string

	instPath       []string
	instPathIsLeft []bool
	instRoot       string
	blkData        string
	signerSigs     []string
	sigIdxs        []int
}

// handleGetBeaconSwapProof returns a proof of a new beacon committee (for a given bridge block height)
func (httpServer *HttpServer) handleGetBeaconSwapProof(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBeaconSwapProof params: %+v", params)
	listParams := params.([]interface{})
	height := uint64(listParams[0].(float64))
	bc := httpServer.config.BlockChain
	db := *httpServer.config.Database

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

	// Decode instruction to send to Ethereum without having to decode on client
	decodedInst, err := blockchain.DecodeInstruction(bridgeInstProof.inst)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	inst := hex.EncodeToString(decodedInst)

	return buildProofResult(inst, beaconInstProof, bridgeInstProof, "", ""), nil
}

// getShardAndBeaconBlocks returns a shard block (with all of its instructions) and the included beacon blocks
func getShardAndBeaconBlocks(
	height uint64,
	bc *blockchain.BlockChain,
	db database.DatabaseInterface,
) (*blockchain.ShardBlock, []*blockchain.BeaconBlock, error) {
	bridgeID := byte(common.BRIDGE_SHARD_ID)
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
	_, instID := findCommSwapInst(insts, metadata.BeaconSwapConfirmMeta)
	if instID < 0 {
		return nil, fmt.Errorf("cannot find beacon swap instruction in bridge block")
	}

	block := &shardBlock{ShardBlock: bridgeBlock}
	return buildProofForBlock(block, insts, instID, db)
}

type block interface {
	InstructionMerkleRoot() []byte
	MetaHash() []byte
	Hash() []byte
	Sig() []string
	ValidatorsIdx() []int
}

// buildProofForBlock builds a swapProof for an instruction in a block (beacon or shard)
func buildProofForBlock(
	blk block,
	insts [][]string,
	id int,
	db database.DatabaseInterface,
) (*swapProof, error) {
	// Build merkle proof for instruction in bridge block
	instProof := buildInstProof(insts, id)

	// Get meta hash and block hash
	instRoot := hex.EncodeToString(blk.InstructionMerkleRoot())
	metaHash := blk.MetaHash()

	// Get sig data
	bSigs := blk.Sig()
	sigs := []string{}
	for _, s := range bSigs {
		sig, _, err := base58.Base58Check{}.Decode(s)
		if err != nil {
			return nil, err
		}
		sigs = append(sigs, hex.EncodeToString(sig))
	}

	// Get index of signers
	signerIdxs := blk.ValidatorsIdx()

	return &swapProof{
		inst:           insts[id],
		instPath:       instProof.getPath(),
		instPathIsLeft: instProof.left,
		instRoot:       instRoot,
		blkData:        hex.EncodeToString(metaHash[:]),
		signerSigs:     sigs,
		sigIdxs:        signerIdxs,
	}, nil
}

// getBeaconSwapProofOnBeacon finds in given beacon blocks a beacon committee swap instruction and returns its proof
func getBeaconSwapProofOnBeacon(
	inst []string,
	beaconBlocks []*blockchain.BeaconBlock,
	db database.DatabaseInterface,
) (*swapProof, error) {
	// Get beacon block and check if it contains beacon swap instruction
	b, instID := findBeaconBlockWithInst(beaconBlocks, inst)
	if b == nil {
		return nil, fmt.Errorf("cannot find corresponding beacon block that includes swap instruction")
	}

	insts := b.Body.Instructions
	block := &beaconBlock{BeaconBlock: b}
	return buildProofForBlock(block, insts, instID, db)
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
	instructions, err := blockchain.CreateShardInstructionsFromTransactionAndInstruction(
		shardBlock.Body.Transactions,
		bc,
		shardBlock.Header.ShardID,
	//	&shardBlock.Header.ProducerAddress,
	//	shardBlock.Header.Height,
	//	beaconBlocks,
	//	shardBlock.Header.BeaconHeight,
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
			BLogger.log.Debug("CommSwap inst:", inst)
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
	BLogger.log.Debugf("BuildProof: %x", merkles[id])
	BLogger.log.Debugf("BuildProof merkles: %x", merkles)
	return buildProofFromTree(merkles, id)
}

// buildInstProof receives a list of instructions (as string) and returns a merkle proof for one instruction in the list
func buildInstProof(insts [][]string, id int) *keccak256MerkleProof {
	flattenInsts, err := blockchain.FlattenAndConvertStringInst(insts)
	if err != nil {
		BLogger.log.Errorf("Cannot flatten instructions: %+v", err)
		return nil
	}
	BLogger.log.Debugf("insts: %v", insts)
	return buildProof(flattenInsts, id)
}

type beaconBlock struct {
	*blockchain.BeaconBlock
}

func (bb *beaconBlock) InstructionMerkleRoot() []byte {
	return bb.Header.InstructionMerkleRoot[:]
}

func (bb *beaconBlock) MetaHash() []byte {
	h := bb.Header.MetaHash()
	return h[:]
}

func (bb *beaconBlock) Hash() []byte {
	h := bb.Header.Hash()
	return h[:]
}

func (bb *beaconBlock) Sig() []string {
	// return bb.BeaconBlock.AggregatedSig
	return []string{}
}

func (bb *beaconBlock) ValidatorsIdx() []int {
	// return bb.BeaconBlock.ValidatorsIdx[idx]
	return []int{}
}

type shardBlock struct {
	*blockchain.ShardBlock
}

func (sb *shardBlock) InstructionMerkleRoot() []byte {
	return sb.Header.InstructionMerkleRoot[:]
}

func (sb *shardBlock) MetaHash() []byte {
	h := sb.Header.MetaHash()
	return h[:]
}

func (sb *shardBlock) Hash() []byte {
	h := sb.Header.Hash()
	return h[:]
}

func (sb *shardBlock) Sig() []string {
	// return sb.ShardBlock.AggregatedSig
	return []string{}
}

func (sb *shardBlock) ValidatorsIdx() []int {
	// return sb.ShardBlock.ValidatorsIdx[idx]
	return []int{}
}

// buildSignersProof builds the merkle proofs for some elements in a list of pubkeys
func buildSignersProof(pubkeys [][]byte, idxs []int) []*keccak256MerkleProof {
	merkles := blockchain.BuildKeccak256MerkleTree(pubkeys)
	BLogger.log.Debugf("pubkeys: %x", pubkeys)
	BLogger.log.Debugf("merkles: %x", merkles)
	proofs := make([]*keccak256MerkleProof, len(pubkeys))
	for i, pid := range idxs {
		proofs[i] = buildProofFromTree(merkles, pid)
	}
	return proofs
}

// findBeaconBlockWithInst finds a beacon block with a specific instruction and the instruction's index; nil if not found
func findBeaconBlockWithInst(beaconBlocks []*blockchain.BeaconBlock, inst []string) (*blockchain.BeaconBlock, int) {
	for _, b := range beaconBlocks {
		for k, blkInst := range b.Body.Instructions {
			diff := false
			for i, part := range blkInst {
				if part != inst[i] {
					diff = true
					break
				}
			}
			if !diff {
				return b, k
			}
		}
	}
	return nil, -1
}

func buildProofResult(
	decodedInst string,
	beaconInstProof *swapProof,
	bridgeInstProof *swapProof,
	beaconHeight string,
	bridgeHeight string,
) jsonresult.GetInstructionProof {
	return jsonresult.GetInstructionProof{
		Instruction:  decodedInst,
		BeaconHeight: beaconHeight,
		BridgeHeight: bridgeHeight,

		BeaconInstPath:       beaconInstProof.instPath,
		BeaconInstPathIsLeft: beaconInstProof.instPathIsLeft,
		BeaconInstRoot:       beaconInstProof.instRoot,
		BeaconBlkData:        beaconInstProof.blkData,
		BeaconSigs:           beaconInstProof.signerSigs,
		BeaconSigIdxs:        beaconInstProof.sigIdxs,

		BridgeInstPath:       bridgeInstProof.instPath,
		BridgeInstPathIsLeft: bridgeInstProof.instPathIsLeft,
		BridgeInstRoot:       bridgeInstProof.instRoot,
		BridgeBlkData:        bridgeInstProof.blkData,
		BridgeSigs:           bridgeInstProof.signerSigs,
		BridgeSigIdxs:        bridgeInstProof.sigIdxs,
	}
}
