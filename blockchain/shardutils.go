package blockchain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

//=======================================BEGIN SHARD BLOCK UTIL
func GetAssignInstructionFromBeaconBlock(beaconBlocks []*BeaconBlock, shardID byte) [][]string {
	assignInstruction := [][]string{}
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			if l[0] == "assign" && l[2] == "shard" {
				if strings.Compare(l[3], strconv.Itoa(int(shardID))) == 0 {
					assignInstruction = append(assignInstruction, l)
				}
			}
		}
	}
	return assignInstruction
}

func FetchBeaconBlockFromHeight(db database.DatabaseInterface, from uint64, to uint64) ([]*BeaconBlock, error) {
	beaconBlocks := []*BeaconBlock{}
	for i := from; i <= to; i++ {
		hash, err := db.GetBeaconBlockHashByIndex(i)
		if err != nil {
			return beaconBlocks, err
		}
		beaconBlockByte, err := db.FetchBeaconBlock(hash)
		if err != nil {
			return beaconBlocks, err
		}
		beaconBlock := BeaconBlock{}
		err = json.Unmarshal(beaconBlockByte, &beaconBlock)
		if err != nil {
			return beaconBlocks, NewBlockChainError(UnmashallJsonBlockError, err)
		}
		beaconBlocks = append(beaconBlocks, &beaconBlock)
	}
	return beaconBlocks, nil
}

func CreateCrossShardByteArray(txList []metadata.Transaction) (crossIDs []byte) {
	byteMap := make([]byte, common.MAX_SHARD_NUMBER)
	for _, tx := range txList {
		switch tx.GetType() {
		case common.TxNormalType, common.TxSalaryType:
			{
				for _, outCoin := range tx.GetProof().OutputCoins {
					lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
					shardID := common.GetShardIDFromLastByte(lastByte)
					byteMap[common.GetShardIDFromLastByte(shardID)] = 1
				}
			}
		case common.TxCustomTokenType:
			{
				customTokenTx := tx.(*transaction.TxCustomToken)
				for _, out := range customTokenTx.TxTokenData.Vouts {
					lastByte := out.PaymentAddress.Pk[len(out.PaymentAddress.Pk)-1]
					shardID := common.GetShardIDFromLastByte(lastByte)
					byteMap[common.GetShardIDFromLastByte(shardID)] = 1
				}
			}
		case common.TxCustomTokenPrivacyType:
			{
				customTokenTx := tx.(*transaction.TxCustomTokenPrivacy)
				for _, outCoin := range customTokenTx.TxTokenPrivacyData.TxNormal.GetProof().OutputCoins {
					lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
					shardID := common.GetShardIDFromLastByte(lastByte)
					byteMap[common.GetShardIDFromLastByte(shardID)] = 1
				}
			}
		}
	}

	for k := range byteMap {
		if byteMap[k] == 1 {
			crossIDs = append(crossIDs, byte(k))
		}
	}

	return crossIDs
}

/*
	Create Swap Action
	Return param:
	#1: swap instruction
	#2: new pending validator list after swapped
	#3: new committees after swapped
	#4: error
*/
func CreateSwapAction(pendingValidator []string, commitees []string, committeeSize int, shardID byte) ([]string, []string, []string, error) {
	fmt.Println("Shard Producer/Create Swap Action: pendingValidator", pendingValidator)
	fmt.Println("Shard Producer/Create Swap Action: commitees", commitees)
	newPendingValidator, newShardCommittees, shardSwapedCommittees, shardNewCommittees, err := SwapValidator(pendingValidator, commitees, committeeSize, common.OFFSET)
	if err != nil {
		return nil, nil, nil, err
	}
	swapInstruction := []string{"swap", strings.Join(shardNewCommittees, ","), strings.Join(shardSwapedCommittees, ","), "shard", strconv.Itoa(int(shardID))}
	return swapInstruction, newPendingValidator, newShardCommittees, nil
}

/*
	Action Generate From Transaction:
	- Stake
	- Stable param: set, del,...
*/
func CreateShardInstructionsFromTransaction(transactions []metadata.Transaction, bcr metadata.BlockchainRetriever, shardID byte) (instructions [][]string) {
	// Generate stake action
	stakeShardPubKey := []string{}
	stakeBeaconPubKey := []string{}
	instructions = buildStabilityActions(transactions, bcr, shardID)

	for _, tx := range transactions {
		switch tx.GetMetadataType() {
		case metadata.ShardStakingMeta:
			pk := tx.GetProof().InputCoins[0].CoinDetails.PublicKey.Compress()
			pkb58 := base58.Base58Check{}.Encode(pk, common.ZeroByte)
			stakeShardPubKey = append(stakeShardPubKey, pkb58)
		case metadata.BeaconStakingMeta:
			pk := tx.GetProof().InputCoins[0].CoinDetails.PublicKey.Compress()
			pkb58 := base58.Base58Check{}.Encode(pk, common.ZeroByte)
			stakeBeaconPubKey = append(stakeBeaconPubKey, pkb58)
			//TODO: stable param 0xsancurasolus
			// case metadata.BuyFromGOVRequestMeta:
		}
	}

	if !reflect.DeepEqual(stakeShardPubKey, []string{}) {
		instruction := []string{"stake", strings.Join(stakeShardPubKey, ","), "shard"}
		instructions = append(instructions, instruction)
	}
	if !reflect.DeepEqual(stakeBeaconPubKey, []string{}) {
		instruction := []string{"stake", strings.Join(stakeBeaconPubKey, ","), "beacon"}
		instructions = append(instructions, instruction)
	}

	return instructions
}

//=======================================END SHARD BLOCK UTIL
//=======================================BEGIN CROSS SHARD UTIL
/*
	Return value #1: outputcoin hash
	Return value #2: merkle data created from outputcoin hash
*/
func CreateShardTxRoot(txList []metadata.Transaction) ([]common.Hash, []common.Hash) {
	//calculate output coin hash for each shard
	outputCoinHash := getOutCoinHashEachShard(txList)
	// calculate merkel path for a shardID
	// step 1: calculate merkle data : [1, 2, 3, 4, 12, 34, 1234]
	/*
			   	1234=hash(12,34)
			   /			  \
		  12=hash(1,2)	 34=hash(3,4)
			 / \	 		 / \
			1	2			3	4
	*/
	merkleData := outputCoinHash
	// fmt.Println("merkleData 1 ", merkleData)
	// fmt.Println("outputCoinHash 1 ", outputCoinHash)
	cursor := 0
	for {
		v1 := merkleData[cursor]
		v2 := merkleData[cursor+1]
		merkleData = append(merkleData, common.HashH(append(v1.GetBytes(), v2.GetBytes()...)))
		cursor += 2
		if cursor >= len(merkleData)-1 {
			break
		}
	}
	return outputCoinHash, merkleData
}

//Receive tx list from shard block body, produce merkle path of UTXO CrossShard List from specific shardID
func GetMerklePathCrossShard(txList []metadata.Transaction, shardID byte) (merklePathShard []common.Hash, merkleShardRoot common.Hash) {
	outputCoinHash, merkleData := CreateShardTxRoot(txList)
	//TODO: @kumi check again (fix infinity loop ->fix by @merman)
	// step 2: get merkle path
	cursor := 0
	lastCursor := 0
	sid := int(shardID)
	i := sid
	time := 0
	for {
		if cursor >= len(merkleData)-2 {
			break
		}
		if i%2 == 0 {
			merklePathShard = append(merklePathShard, merkleData[cursor+i+1])
		} else {
			merklePathShard = append(merklePathShard, merkleData[cursor+i-1])
		}
		i = i / 2

		if time == 0 {
			cursor += len(outputCoinHash)
		} else {
			tmp := cursor
			cursor += (cursor - lastCursor) / 2
			lastCursor = tmp
		}
		time++
	}
	// fmt.Println("merklePathShard", merklePathShard)
	// fmt.Println("merkleShardRoot", merkleShardRoot)
	merkleShardRoot = merkleData[len(merkleData)-1]
	return merklePathShard, merkleShardRoot
}

//Receive a cross shard block and merkle path, verify whether the UTXO list is valid or not
func VerifyCrossShardBlockUTXO(block *CrossShardBlock, merklePathShard []common.Hash) bool {
	outCoins := block.CrossOutputCoin
	tmpByte := []byte{}
	for _, coin := range outCoins {
		tmpByte = append(tmpByte, coin.Bytes()...)
	}
	finalHash := common.HashH(tmpByte)
	for _, hash := range merklePathShard {
		finalHash = common.HashH(append(finalHash.GetBytes(), hash.GetBytes()...))
	}
	return VerifyMerkleTree(finalHash, merklePathShard, block.Header.ShardTxRoot)
}

func VerifyMerkleTree(finalHash common.Hash, merklePath []common.Hash, merkleRoot common.Hash) bool {
	for _, hashPath := range merklePath {
		finalHash = common.HashH(append(finalHash.GetBytes(), hashPath.GetBytes()...))
	}
	if finalHash != merkleRoot {
		return false
	} else {
		return true
	}
}

/*
	Helper function: group OutputCoin into shard and get the hash of each group
	Return value
		- Array of hash created from 256 group OutputCoin hash
		- Length array is 256
		- Value is sorted as shardID from low to high
		- ShardID which have no outputcoin received hash of emptystring value
*/
func getOutCoinHashEachShard(txList []metadata.Transaction) []common.Hash {
	// group transaction by shardID
	outCoinEachShard := make([][]*privacy.OutputCoin, common.MAX_SHARD_NUMBER)
	for _, tx := range txList {
		for _, outCoin := range tx.GetProof().OutputCoins {
			lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
			shardID := common.GetShardIDFromLastByte(lastByte)
			outCoinEachShard[shardID] = append(outCoinEachShard[shardID], outCoin)
		}
	}
	// fmt.Println("len of outCoinEachShard", len(outCoinEachShard))
	// fmt.Println("value of outCoinEachShard", outCoinEachShard)
	//calcualte hash for each shard
	outputCoinHash := make([]common.Hash, common.MAX_SHARD_NUMBER)
	for i := 0; i < common.MAX_SHARD_NUMBER; i++ {
		if len(outCoinEachShard[i]) == 0 {
			outputCoinHash[i] = common.HashH([]byte(""))
		} else {
			tmpByte := []byte{}
			for _, coin := range outCoinEachShard[i] {
				tmpByte = append(tmpByte, coin.Bytes()...)
			}
			outputCoinHash[i] = common.HashH(tmpByte)
		}
	}
	// fmt.Println("len of outputCoinHash", len(outputCoinHash))
	return outputCoinHash
}

// helper function to get the hash of OutputCoins (send to a shard) from list of transaction
/*
	Get output coin of transaction
	Check receiver last byte
	Append output coin to corresponding shard
*/
func getOutCoinCrossShard(txList []metadata.Transaction, shardID byte) []privacy.OutputCoin {
	coinList := []privacy.OutputCoin{}
	for _, tx := range txList {
		for _, outCoin := range tx.GetProof().OutputCoins {
			lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
			if lastByte == shardID {
				coinList = append(coinList, *outCoin)
			}
		}
	}
	return coinList
}

//=======================================END CROSS SHARD UTIL

//=======================================BEGIN CROSS OUTPUT COIN UTIL
func CreateMerkleCrossOutputCoin(crossOutputCoins map[byte][]CrossOutputCoin) (*common.Hash, error) {
	if len(crossOutputCoins) == 0 {
		res, err := GenerateZeroValueHash()

		return &res, err
	}
	keys := []int{}
	crossOutputCoinHashes := []*common.Hash{}
	for k := range crossOutputCoins {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, value := range crossOutputCoins[byte(shardID)] {
			hash := value.Hash()
			hashByte := hash.GetBytes()
			newHash, err := common.Hash{}.NewHash(hashByte)
			if err != nil {
				return &common.Hash{}, NewBlockChainError(HashError, err)
			}
			crossOutputCoinHashes = append(crossOutputCoinHashes, newHash)
		}
	}
	merkle := Merkle{}
	merkleTree := merkle.BuildMerkleTreeOfHashs(crossOutputCoinHashes)
	return merkleTree[len(merkleTree)-1], nil
}

func VerifyMerkleCrossOutputCoin(crossOutputCoins map[byte][]CrossOutputCoin, rootHash common.Hash) bool {
	res, err := CreateMerkleCrossOutputCoin(crossOutputCoins)
	if err != nil {
		return false
	}
	hashByte := rootHash.GetBytes()
	newHash, err := common.Hash{}.NewHash(hashByte)
	if err != nil {
		return false
	}
	return newHash.IsEqual(res)
}

func (blockchain *BlockChain) StoreIncomingCrossShard(block *ShardBlock) error {
	crossShardMap, _ := block.Body.ExtractIncomingCrossShardMap()
	for crossShard, crossBlks := range crossShardMap {
		for _, crossBlk := range crossBlks {
			blockchain.config.DataBase.StoreIncomingCrossShard(block.Header.ShardID, crossShard, block.Header.Height, &crossBlk)
		}
	}
	return nil
}

//=======================================BEGIN END OUTPUT COIN UTIL
