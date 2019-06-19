package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

// handleGetBestBlock implements the getbestblock command.
func (rpcServer RpcServer) handleGetBeaconSwapProof(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBeaconSwapProof params: %+v", params)
	listParams := params.([]interface{})
	height := uint64(listParams[0].(float64))
	bc := rpcServer.config.BlockChain
	db := *rpcServer.config.Database

	// Get proof of instruction on bridge
	beaconInstProof, beaconBlocks, err := getBeaconSwapProofOnBridge(height-1, bc, db)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// Get proof of instruction on beacon
	bridgeInstProof, err := getBeaconSwapProofOnBeacon(beaconInstProof.inst, beaconBlocks, db)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// Save instruction as a single slice of byte
	flattenBridgeInst := []byte{}
	for _, part := range beaconInstProof.inst {
		flattenBridgeInst = append(flattenBridgeInst, []byte(part)...)
	}

	return jsonresult.GetBeaconSwapProof{
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

type instructionProof struct {
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

func getBeaconSwapProofOnBridge(
	height uint64,
	bc *blockchain.BlockChain,
	db database.DatabaseInterface,
) (*instructionProof, []*blockchain.BeaconBlock, error) {
	// Get bridge block and check if it contains beacon swap instruction
	bridgeID := byte(1) // TODO(@0xbunyip): replace with bridge shardID
	bridgeBlock, err := bc.GetShardBlockByHeight(height, bridgeID)
	if err != nil {
		return nil, nil, err
	}
	beaconBlocks, err := getIncludedBeaconBlocks(
		bc,
		db,
		bridgeBlock.Header.Height,
		bridgeBlock.Header.BeaconHeight,
		bridgeBlock.Header.ShardID,
	)
	if err != nil {
		return nil, nil, err
	}
	bridgeInsts, err := extractInstsFromShardBlock(bridgeBlock, beaconBlocks, bc)
	if err != nil {
		return nil, nil, err
	}
	bridgeInst, bridgeInstId := findCommSwapInst(bridgeInsts)
	if bridgeInstId < 0 {
		return nil, nil, nil
	}

	// Build merkle proof for instruction in bridge block
	bridgeProof := buildInstProof(bridgeInsts, bridgeInstId)

	// Get committee pubkey and signature
	pubkeys, signerIdxs, err := getBridgeSignerPubkeys(bridgeBlock, db)
	if err != nil {
		return nil, nil, err
	}
	bridgeSignerPubkeys := make([]string, len(pubkeys))
	for i, pk := range pubkeys {
		bridgeSignerPubkeys[i] = hex.EncodeToString(pk)
	}

	// Build merkle proof for signer pubkeys
	signerProof := buildSignersProof(pubkeys, signerIdxs)
	bridgeSignerPaths := make([][]string, len(pubkeys))
	bridgeSignerPathIsLeft := make([][]bool, len(pubkeys))
	for i, p := range signerProof {
		bridgeSignerPaths[i] = p.getPath()
		bridgeSignerPathIsLeft[i] = p.left
	}

	// Get meta hash and block hash
	bridgeInstRoot := hex.EncodeToString(bridgeBlock.Header.InstructionMerkleRoot[:])
	bridgeMetaHash := bridgeBlock.Header.MetaHash()
	bridgeBlkHash := bridgeBlock.Header.Hash()

	return &instructionProof{
		inst:             bridgeInst,
		instPath:         bridgeProof.getPath(),
		instPathIsLeft:   bridgeProof.left,
		instRoot:         bridgeInstRoot,
		blkData:          hex.EncodeToString(bridgeMetaHash[:]),
		blkHash:          hex.EncodeToString(bridgeBlkHash[:]),
		signerPubkeys:    bridgeSignerPubkeys,
		signerSig:        bridgeBlock.AggregatedSig,
		signerPaths:      bridgeSignerPaths,
		signerPathIsLeft: bridgeSignerPathIsLeft,
	}, beaconBlocks, nil
}

func getBeaconSwapProofOnBeacon(
	inst []string,
	beaconBlocks []*blockchain.BeaconBlock,
	db database.DatabaseInterface,
) (*instructionProof, error) {
	// Get beacon block and check if it contains beacon swap instruction
	beaconBlock := findBeaconBlockWithInst(beaconBlocks, inst)
	if beaconBlock == nil {
		return nil, fmt.Errorf("cannot find corresponding beacon block that includes swap instruction")
	}

	beaconInsts := beaconBlock.Body.Instructions
	_, beaconInstId := findCommSwapInst(beaconInsts)
	if beaconInstId < 0 {
		return nil, fmt.Errorf("cannot find swap instruction in beacon block")
	}

	// Build merkle proof for instruction in beacon block
	beaconProof := buildInstProof(beaconInsts, beaconInstId)

	// Get committee pubkey and signature
	pubkeys, signerIdxs, err := getBeaconSignerPubkeys(beaconBlock, db)
	if err != nil {
		return nil, err
	}
	beaconSignerPubkeys := make([]string, len(pubkeys))
	for i, pk := range pubkeys {
		beaconSignerPubkeys[i] = hex.EncodeToString(pk)
	}

	// Build merkle proof for signer pubkeys
	signerProof := buildSignersProof(pubkeys, signerIdxs)
	beaconSignerPaths := make([][]string, len(pubkeys))
	beaconSignerPathIsLeft := make([][]bool, len(pubkeys))
	for i, p := range signerProof {
		beaconSignerPaths[i] = p.getPath()
		beaconSignerPathIsLeft[i] = p.left
	}

	// Get meta hash and block hash
	beaconInstRoot := hex.EncodeToString(beaconBlock.Header.InstructionMerkleRoot[:])
	beaconMetaHash := beaconBlock.Header.MetaHash()
	beaconBlkHash := beaconBlock.Header.Hash()

	return &instructionProof{
		instPath:         beaconProof.getPath(),
		instPathIsLeft:   beaconProof.left,
		instRoot:         beaconInstRoot,
		blkData:          hex.EncodeToString(beaconMetaHash[:]),
		blkHash:          hex.EncodeToString(beaconBlkHash[:]),
		signerPubkeys:    beaconSignerPubkeys,
		signerSig:        beaconBlock.AggregatedSig,
		signerPaths:      beaconSignerPaths,
		signerPathIsLeft: beaconSignerPathIsLeft,
	}, nil
}

// getIncludedBeaconBlocks retrieves all beacon blocks included in a shard block
func getIncludedBeaconBlocks(
	bc *blockchain.BlockChain,
	db database.DatabaseInterface,
	shardHeight uint64,
	beaconHeight uint64,
	shardID byte,
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

// findCommSwapInst finds a beacon swap instruction in a list, returns it with its index
func findCommSwapInst(insts [][]string) ([]string, int) {
	for i, inst := range insts {
		if strconv.Itoa(metadata.BeaconPubkeyRootMeta) == inst[0] {
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
	flattenInsts := common.FlattenAndConvertStringInst(insts)
	return buildProof(flattenInsts, id)
}

// getBeaconSignerPubkeys finds the pubkeys of all signers of a beacon block
func getBeaconSignerPubkeys(shardBlock *blockchain.BeaconBlock, db database.DatabaseInterface) ([][]byte, []int, error) {
	return nil, nil, nil
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
