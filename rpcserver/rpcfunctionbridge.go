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
	bridgeID := byte(1) // TODO(@0xbunyip): replace with bridge shardID
	bc := rpcServer.config.BlockChain
	db := *rpcServer.config.Database

	// Get bridge block and check if it contains beacon swap instruction
	bridgeBlock, err := bc.GetShardBlockByHeight(height-1, bridgeID)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	bridgeInsts, err := extractInstsFromShardBlock(bridgeBlock, bc, db)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	bridgeInst, bridgeInstId := findCommSwapInst(bridgeInsts)
	if bridgeInstId < 0 {
		return nil, nil
	}

	// Build merkle proof for instruction in bridge block
	bridgeProof := buildInstProof(bridgeInsts, bridgeInstId)

	flattenbridgeInst := []byte{}
	for _, part := range bridgeInst {
		flattenbridgeInst = append(flattenbridgeInst, []byte(part)...)
	}

	// Get committee pubkey and signature
	pubkeys, signerIdxs, err := getBridgeSignerPubkeys(bridgeBlock, db)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
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

	return jsonresult.GetBeaconSwapProof{
		Instruction:            hex.EncodeToString(flattenbridgeInst),
		BridgeInstPath:         bridgeProof.getPath(),
		BridgeInstPathIsLeft:   bridgeProof.left,
		BridgeInstRoot:         bridgeInstRoot,
		BridgeBlkData:          hex.EncodeToString(bridgeMetaHash[:]),
		BridgeBlkHash:          hex.EncodeToString(bridgeBlkHash[:]),
		BridgeSignerPubkeys:    bridgeSignerPubkeys,
		BridgeSignerSig:        bridgeBlock.AggregatedSig,
		BridgeSignerPaths:      bridgeSignerPaths,
		BridgeSignerPathIsLeft: bridgeSignerPathIsLeft,
	}, nil

	// Get the corresponding beacon block with the swap instruction

	// Build instruction merkle proof for bridge block
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

func buildProofFromTree(merkles [][]byte, id int) *keccak256MerkleProof {
	path, left := blockchain.GetKeccak256MerkleProofFromTree(merkles, id)
	return &keccak256MerkleProof{path: path, left: left}
}

func buildProof(data [][]byte, id int) *keccak256MerkleProof {
	merkles := blockchain.BuildKeccak256MerkleTree(data)
	return buildProofFromTree(merkles, id)
}

func buildInstProof(insts [][]string, id int) *keccak256MerkleProof {
	flattenInsts := common.FlattenAndConvertStringInst(insts)
	return buildProof(flattenInsts, id)
}

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
