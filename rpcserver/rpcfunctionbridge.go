package rpcserver

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

// handleGetBestBlock implements the getbestblock command.
func (rpcServer RpcServer) handleGetBeaconSwapProof(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBeaconSwapProof params: %+v", params)
	listParams := params.([]interface{})
	height := uint64(listParams[0].(float64))
	bridgeID := byte(1) // TODO(@0xbunyip): replace with bridge shardID
	bc := rpcServer.config.BlockChain
	db := *rpcServer.config.Database

	// Get bridge block and check if it contains beacon swap instruction
	shardBlock, err := bc.GetShardBlockByHeight(height-1, bridgeID)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	shardInsts, err := extractInstsFromShardBlock(shardBlock, bc, db)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	bridgeInst, bridgeInstId := findCommSwapInst(shardInsts)
	if bridgeInstId < 0 {
		return nil, nil
	}

	// Build instruction merkle proof for bridge block
	bridgeProof := buildKeccak256MerkleProof(shardInsts, bridgeInstId)

	flattenbridgeInst := []byte{}
	for _, part := range bridgeInst {
		flattenbridgeInst = append(flattenbridgeInst, []byte(part)...)
	}
	bridgeInstRoot := hex.EncodeToString(shardBlock.Header.InstructionMerkleRoot[:])
	bridgeMetaHash := shardBlock.Header.MetaHash()
	bridgeBlkHash := shardBlock.Header.Hash()

	return jsonresult.GetBeaconSwapProof{
		Instruction:            hex.EncodeToString(flattenbridgeInst),
		BridgeInstPath:         bridgeProof.getPath(),
		BridgeInstPathIsLeft:   bridgeProof.left,
		BridgeInstRoot:         bridgeInstRoot,
		BridgeBlkData:          hex.EncodeToString(bridgeMetaHash[:]),
		BridgeBlkHash:          hex.EncodeToString(bridgeBlkHash[:]),
		BridgeSignerPubkeys:    nil,
		BridgeSignerSig:        shardBlock.Header.AggregatedSig,
		BridgeSignerPaths:      nil,
		BridgeSignerPathIsLeft: nil,
	}, nil

	// Get the corresponding beacon block with the swap instruction

	// Build instruction merkle proof for bridge block

	// Get committee signature and meta hash
}

func extractInstsFromShardBlock(
	shardBlock *blockchain.ShardBlock,
	bc *blockchain.BlockChain,
	db database.DatabaseInterface,
) ([][]string, error) {
	prevShardBlock, err := bc.GetShardBlockByHeight(shardBlock.Header.Height-1, shardBlock.Header.ShardID)
	if err != nil {
		return nil, err
	}
	beaconBlocks, err := blockchain.FetchBeaconBlockFromHeight(
		db,
		prevShardBlock.Header.BeaconHeight+1,
		shardBlock.Header.BeaconHeight,
	)
	if err != nil {
		return nil, err
	}
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

func (p *keccak256MerkleProof) getPath() []string {
	path := make([]string, len(p.path))
	for i, h := range p.path {
		path[i] = hex.EncodeToString(h)
	}
	return path
}

func buildKeccak256MerkleProof(insts [][]string, id int) *keccak256MerkleProof {
	flattenInsts := common.FlattenAndConvertStringInst(insts)
	fmt.Println("[db] flattenInsts", flattenInsts)
	path, left := blockchain.GetKeccak256MerkleProof(flattenInsts, id)
	return &keccak256MerkleProof{path: path, left: left}
}
