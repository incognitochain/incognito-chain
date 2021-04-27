package devframework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/transaction"
)

//this function used to update several field that need recalculated in block header (in case there is change of body content)
func UpdateShardHeaderOnBodyChange(block *types.ShardBlock, bc *blockchain.BlockChain) error {
	totalTxsFee := make(map[common.Hash]uint64)
	for _, tx := range block.Body.Transactions {
		totalTxsFee[*tx.GetTokenID()] += tx.GetTxFee()
		txType := tx.GetType()
		if txType == common.TxCustomTokenPrivacyType {
			txCustomPrivacy := tx.(*transaction.TxCustomTokenPrivacy)
			totalTxsFee[*txCustomPrivacy.GetTokenID()] = txCustomPrivacy.GetTxFeeToken()
		}
	}
	merkleRoots := blockchain.Merkle{}.BuildMerkleTreeStore(block.Body.Transactions)
	merkleRoot := &common.Hash{}
	if len(merkleRoots) > 0 {
		merkleRoot = merkleRoots[len(merkleRoots)-1]
	}
	crossTransactionRoot, err := blockchain.CreateMerkleCrossTransaction(block.Body.CrossTransactions)
	if err != nil {
		fmt.Println(err)
	}
	txInstructions, err := blockchain.CreateShardInstructionsFromTransactionAndInstruction(block.Body.Transactions, bc, block.Header.ShardID, block.Header.Height)
	if err != nil {
		return err
	}
	totalInstructions := []string{}
	for _, value := range txInstructions {
		totalInstructions = append(totalInstructions, value...)
	}
	for _, value := range block.Body.Instructions {
		totalInstructions = append(totalInstructions, value...)
	}
	instructionsHash, err := generateHashFromStringArray(totalInstructions)
	if err != nil {
		return err
	}
	// Instruction merkle root
	flattenTxInsts, err := blockchain.FlattenAndConvertStringInst(txInstructions)
	if err != nil {
		return fmt.Errorf("Instruction from Tx: %+v", err)
	}
	flattenInsts, err := blockchain.FlattenAndConvertStringInst(block.Body.Instructions)
	if err != nil {
		return fmt.Errorf("Instruction from block body: %+v", err)
	}
	insts := append(flattenTxInsts, flattenInsts...) // Order of instructions must be preserved in shardprocess
	instMerkleRoot := blockchain.GetKeccak256MerkleRoot(insts)

	_, shardTxMerkleData := blockchain.CreateShardTxRoot(block.Body.Transactions)
	block.Header.TotalTxsFee = totalTxsFee
	block.Header.TxRoot = *merkleRoot
	block.Header.ShardTxRoot = shardTxMerkleData[len(shardTxMerkleData)-1]
	block.Header.CrossTransactionRoot = *crossTransactionRoot
	block.Header.CrossShardBitMap = blockchain.CreateCrossShardByteArray(block.Body.Transactions, block.Header.ShardID)
	block.Header.InstructionsRoot = instructionsHash
	copy(block.Header.InstructionMerkleRoot[:], instMerkleRoot)
	return nil
}

//this function used to update several field that need recalculated in block header (in case there is change of body content)
func UpdateBeaconHeaderOnBodyChange(block *types.BeaconBlock) error {

	// Shard state hash
	tempShardStateHash, err := generateHashFromShardState(block.Body.ShardState)
	if err != nil {
		return err
	}
	// Instruction Hash
	tempInstructionArr := []string{}
	for _, strs := range block.Body.Instructions {
		tempInstructionArr = append(tempInstructionArr, strs...)
	}
	tempInstructionHash, err := generateHashFromStringArray(tempInstructionArr)
	if err != nil {
		return err
	}
	// Instruction merkle root
	flattenInsts, err := blockchain.FlattenAndConvertStringInst(block.Body.Instructions)
	if err != nil {
		return err
	}

	block.Header.ShardStateHash = tempShardStateHash
	block.Header.InstructionHash = tempInstructionHash
	copy(block.Header.InstructionMerkleRoot[:], blockchain.GetKeccak256MerkleRoot(flattenInsts))
	return nil
}
func generateZeroValueHash() (common.Hash, error) {
	hash := common.Hash{}
	hash.SetBytes(make([]byte, 32))
	return hash, nil
}

func generateHashFromStringArray(strs []string) (common.Hash, error) {
	// if input is empty list
	// return hash value of bytes zero
	if len(strs) == 0 {
		return generateZeroValueHash()
	}
	var (
		hash common.Hash
		buf  bytes.Buffer
	)
	for _, value := range strs {
		buf.WriteString(value)
	}
	temp := common.HashB(buf.Bytes())
	if err := hash.SetBytes(temp[:]); err != nil {
		return common.Hash{}, err
	}
	return hash, nil
}

func generateHashFromShardState(allShardState map[byte][]types.ShardState) (common.Hash, error) {
	allShardStateStr := []string{}
	var keys []int
	for k := range allShardState {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		res := ""
		for _, shardState := range allShardState[byte(shardID)] {
			res += strconv.Itoa(int(shardState.Height))
			res += shardState.Hash.String()
			crossShard, _ := json.Marshal(shardState.CrossShard)
			res += string(crossShard)
		}
		allShardStateStr = append(allShardStateStr, res)
	}
	return generateHashFromStringArray(allShardStateStr)
}
