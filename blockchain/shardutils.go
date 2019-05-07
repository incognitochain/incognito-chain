package blockchain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
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

func CreateCrossShardByteArray(txList []metadata.Transaction, fromShardID byte) []byte {
	crossIDs := []byte{}
	byteMap := make([]byte, common.MAX_SHARD_NUMBER)
	for _, tx := range txList {
		if tx.GetProof() != nil {
			for _, outCoin := range tx.GetProof().OutputCoins {
				lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
				shardID := common.GetShardIDFromLastByte(lastByte)
				byteMap[common.GetShardIDFromLastByte(shardID)] = 1
			}
		}
		switch tx.GetType() {
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
				if customTokenTx.TxTokenPrivacyData.TxNormal.GetProof() != nil {
					for _, outCoin := range customTokenTx.TxTokenPrivacyData.TxNormal.GetProof().OutputCoins {
						lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
						shardID := common.GetShardIDFromLastByte(lastByte)
						byteMap[common.GetShardIDFromLastByte(shardID)] = 1
					}
				}
			}
		}
	}

	for k := range byteMap {
		if byteMap[k] == 1 && k != int(fromShardID) {
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
func CreateShardInstructionsFromTransactionAndIns(
	transactions []metadata.Transaction,
	bc *BlockChain,
	shardID byte,
	producerAddress *privacy.PaymentAddress,
	shardBlockHeight uint64,
	beaconBlocks []*BeaconBlock,
	beaconHeight uint64,
) (instructions [][]string, err error) {
	// Generate stake action
	stakeShardPubKey := []string{}
	stakeBeaconPubKey := []string{}
	instructions, err = buildStabilityActions(transactions, bc, shardID, producerAddress, shardBlockHeight, beaconBlocks, beaconHeight)
	if err != nil {
		fmt.Println("[ndh] - wtf err???", err)
		return nil, err
	}

	for _, tx := range transactions {
		if tx.GetMetadataType() != 38 {
			//fmt.Println("[voting] - CreateShardInstructionsFromTransactionAndIns: ", tx.GetMetadataType())
		}
		switch tx.GetMetadataType() {
		case metadata.ShardStakingMeta:
			pk := tx.GetProof().InputCoins[0].CoinDetails.PublicKey.Compress()
			pkb58 := base58.Base58Check{}.Encode(pk, common.ZeroByte)
			stakeShardPubKey = append(stakeShardPubKey, pkb58)
		case metadata.BeaconStakingMeta:
			pk := tx.GetProof().InputCoins[0].CoinDetails.PublicKey.Compress()
			pkb58 := base58.Base58Check{}.Encode(pk, common.ZeroByte)
			stakeBeaconPubKey = append(stakeBeaconPubKey, pkb58)
		}
	}

	if !reflect.DeepEqual(stakeShardPubKey, []string{}) {
		instruction := []string{StakeAction, strings.Join(stakeShardPubKey, ","), "shard"}
		instructions = append(instructions, instruction)
	}
	if !reflect.DeepEqual(stakeBeaconPubKey, []string{}) {
		instruction := []string{StakeAction, strings.Join(stakeBeaconPubKey, ","), "beacon"}
		instructions = append(instructions, instruction)
	}

	return instructions, nil
}

//=======================================END SHARD BLOCK UTIL
//====================OLD Merkle Tree============
/*
	Return value #1: outputcoin hash
	Return value #2: merkle data created from outputcoin hash
*/
func CreateShardTxRoot(txList []metadata.Transaction) ([]common.Hash, []common.Hash) {
	//calculate output coin hash for each shard
	crossShardDataHash := getCrossShardDataHash(txList)
	// calculate merkel path for a shardID
	// step 1: calculate merkle data : [1, 2, 3, 4, 12, 34, 1234]
	/*
			   	1234=hash(12,34)
			   /			  \
		  12=hash(1,2)	 34=hash(3,4)
			 / \	 		 / \
			1	2			3	4
	*/
	merkleData := crossShardDataHash
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
	return crossShardDataHash, merkleData
}

//Receive tx list from shard block body, produce merkle path of UTXO CrossShard List from specific shardID
func GetMerklePathCrossShard(txList []metadata.Transaction, shardID byte) (merklePathShard []common.Hash, merkleShardRoot common.Hash) {
	crossShardDataHash, merkleData := CreateShardTxRoot2(txList)
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
			cursor += len(crossShardDataHash)
		} else {
			tmp := cursor
			cursor += (cursor - lastCursor) / 2
			lastCursor = tmp
		}
		time++
	}
	merkleShardRoot = merkleData[len(merkleData)-1]
	return merklePathShard, merkleShardRoot
}

//Receive a cross shard block and merkle path, verify whether the UTXO list is valid or not
/*
	Calculate Final Hash as Hash of:
		1. CrossTransactionFinalHash
		2. TxTokenDataVoutFinalHash
	These hashes will be calculated as comment in getCrossShardDataHash function
*/
func VerifyCrossShardBlockUTXO(block *CrossShardBlock, merklePathShard []common.Hash) bool {
	var outputCoinHash common.Hash
	var txTokenDataHash common.Hash
	var txTokenPrivacyDataHash common.Hash

	outCoins := block.CrossOutputCoin
	outputCoinHash = calHashOutCoinCrossShard(outCoins)

	txTokenDataList := block.CrossTxTokenData
	txTokenDataHash = calHashTxTokenDataHashList(txTokenDataList)

	txTokenPrivacyDataList := block.CrossTxTokenPrivacyData
	txTokenPrivacyDataHash = calHashTxTokenPrivacyDataHashList(txTokenPrivacyDataList)

	tmpByte := append(append(outputCoinHash.GetBytes(), txTokenDataHash.GetBytes()...), txTokenPrivacyDataHash.GetBytes()...)
	finalHash := common.HashH(tmpByte)
	// return Merkle{}.VerifyMerkleRootFromMerklePath(finalHash, merklePathShard, block.Header.ShardTxRoot, block.ToShardID)
	return VerifyMerkleTree(finalHash, merklePathShard, block.Header.ShardTxRoot, block.ToShardID)
}

func VerifyMerkleTree(finalHash common.Hash, merklePath []common.Hash, merkleRoot common.Hash, receiverShardID byte) bool {
	i := int(receiverShardID)
	for _, hashPath := range merklePath {
		if i%2 == 0 {
			finalHash = common.HashH(append(finalHash.GetBytes(), hashPath.GetBytes()...))
		} else {
			finalHash = common.HashH(append(hashPath.GetBytes(), finalHash.GetBytes()...))
		}
		i = i / 2
	}
	merkleRootString := merkleRoot.String()
	if strings.Compare(finalHash.String(), merkleRootString) != 0 {
		return false
	} else {
		return true
	}
}

//====================END OLD Merkle Tree============

//====================New Merkle Tree================
func CreateShardTxRoot2(txList []metadata.Transaction) ([]common.Hash, []common.Hash) {
	//calculate output coin hash for each shard
	crossShardDataHash := getCrossShardDataHash(txList)
	// calculate merkel path for a shardID
	// step 1: calculate merkle data : [1, 2, 3, 4, 12, 34, 1234]
	/*
			   	1234=hash(12,34)
			   /			  \
		  12=hash(1,2)	 34=hash(3,4)
			 / \	 		 / \
			1	2			3	4
	*/
	merkleTree := Merkle{}
	merkleData := merkleTree.BuildMerkleTreeOfHashes2(crossShardDataHash, common.MAX_SHARD_NUMBER)
	return crossShardDataHash, merkleData
}
func GetMerklePathCrossShard2(txList []metadata.Transaction, shardID byte) (merklePathShard []common.Hash, merkleShardRoot common.Hash) {
	_, merkleTree := CreateShardTxRoot2(txList)
	merklePathShard, merkleShardRoot = Merkle{}.GetMerklePathForCrossShard(common.MAX_SHARD_NUMBER, merkleTree, shardID)
	return merklePathShard, merkleShardRoot
}

/*
	Calculate Final Hash as Hash of:
		1. CrossTransactionFinalHash
		2. TxTokenDataVoutFinalHash
		3. CrossTxTokenPrivacyData
	These hashes will be calculated as comment in getCrossShardDataHash function
*/
func VerifyCrossShardBlockUTXO2(block *CrossShardBlock, merklePathShard []common.Hash) bool {
	var outputCoinHash common.Hash
	var txTokenDataHash common.Hash
	var txTokenPrivacyDataHash common.Hash

	outCoins := block.CrossOutputCoin
	outputCoinHash = calHashOutCoinCrossShard(outCoins)

	txTokenDataList := block.CrossTxTokenData
	txTokenDataHash = calHashTxTokenDataHashList(txTokenDataList)

	txTokenPrivacyDataList := block.CrossTxTokenPrivacyData
	txTokenPrivacyDataHash = calHashTxTokenPrivacyDataHashList(txTokenPrivacyDataList)

	tmpByte := append(append(outputCoinHash.GetBytes(), txTokenDataHash.GetBytes()...), txTokenPrivacyDataHash.GetBytes()...)
	finalHash := common.HashH(tmpByte)
	return Merkle{}.VerifyMerkleRootFromMerklePath(finalHash, merklePathShard, block.Header.ShardTxRoot, block.ToShardID)
}

//====================End New Merkle Tree================
/*
	Helper function: group OutputCoin into shard and get the hash of each group
	Return value
		- Array of hash created from 256 group cross shard data hash
		- Length array is 256
		- Value is sorted as shardID from low to high
		- ShardID which have no outputcoin received hash of emptystring value

	Hash Procedure:
		- For each shard:
			CROSS OUTPUT COIN
			+ Get outputcoin and append to a list of that shard
			+ Calculate value for Hash:
				* if receiver shard has no outcoin then received hash value of empty string
				* if receiver shard has >= 1 outcoin then concatenate all outcoin bytes value then hash
				* At last, we compress all cross out put coin into a CrossOutputCoinFinalHash
			TXTOKENDATA
			+ Do the same as above

			=> Then Final Hash of each shard is Hash of value in this order:
				1. CrossOutputCoinFinalHash
				2. TxTokenDataVoutFinalHash
	TxTokenOut DataStructure
		- Use Only One TxTokenData for one TokenID
		- Vouts of one tokenID from many transaction will be compress into One Vouts List
		- Using Key-Value structure for accessing one token ID data:
			key: token ID
			value: TokenData of that token
*/
func getCrossShardDataHash(txList []metadata.Transaction) []common.Hash {
	// group transaction by shardID
	outCoinEachShard := make([][]privacy.OutputCoin, common.MAX_SHARD_NUMBER)
	txTokenDataEachShard := make([]map[common.Hash]*transaction.TxTokenData, common.MAX_SHARD_NUMBER)
	txTokenPrivacyDataMap := make([]map[common.Hash]*ContentCrossTokenPrivacyData, common.MAX_SHARD_NUMBER)
	for _, tx := range txList {
		switch tx.GetType() {
		//==================For Constant Transfer Only
		case common.TxNormalType, common.TxSalaryType:
			{
				//==================Proof Process
				if tx.GetProof() != nil {
					for _, outCoin := range tx.GetProof().OutputCoins {
						lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
						shardID := common.GetShardIDFromLastByte(lastByte)
						outCoinEachShard[shardID] = append(outCoinEachShard[shardID], *outCoin)
					}
				}
			}
		//==================For Constant & TxCustomToken Transfer
		case common.TxCustomTokenType:
			{
				customTokenTx := tx.(*transaction.TxCustomToken)
				//==================Proof Process
				if customTokenTx.GetProof() != nil {
					for _, outCoin := range customTokenTx.GetProof().OutputCoins {
						lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
						shardID := common.GetShardIDFromLastByte(lastByte)
						outCoinEachShard[shardID] = append(outCoinEachShard[shardID], *outCoin)
					}
				}
				//==================Tx Token Data Process
				for _, vout := range customTokenTx.TxTokenData.Vouts {
					lastByte := vout.PaymentAddress.Pk[len(vout.PaymentAddress.Pk)-1]
					shardID := common.GetShardIDFromLastByte(lastByte)
					if txTokenDataEachShard[shardID] == nil {
						txTokenDataEachShard[shardID] = make(map[common.Hash]*transaction.TxTokenData)
					}
					if _, ok := txTokenDataEachShard[shardID][customTokenTx.TxTokenData.PropertyID]; !ok {
						newTxTokenData := cloneTxTokenDataForCrossShard(customTokenTx.TxTokenData)
						txTokenDataEachShard[shardID][customTokenTx.TxTokenData.PropertyID] = &newTxTokenData
					}
					vouts := txTokenDataEachShard[shardID][customTokenTx.TxTokenData.PropertyID].Vouts
					vouts = append(vouts, vout)
					txTokenDataEachShard[shardID][customTokenTx.TxTokenData.PropertyID].Vouts = vouts
				}
			}
		case common.TxCustomTokenPrivacyType:
			{
				customTokenPrivacyTx := tx.(*transaction.TxCustomTokenPrivacy)
				//==================Proof Process
				if customTokenPrivacyTx.GetProof() != nil {
					for _, outCoin := range customTokenPrivacyTx.GetProof().OutputCoins {
						lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
						shardID := common.GetShardIDFromLastByte(lastByte)
						outCoinEachShard[shardID] = append(outCoinEachShard[shardID], *outCoin)
					}
				}
				//==================Tx Token Privacy Data Process
				if customTokenPrivacyTx.TxTokenPrivacyData.TxNormal.GetProof() != nil {
					for _, outCoin := range customTokenPrivacyTx.TxTokenPrivacyData.TxNormal.GetProof().OutputCoins {
						lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
						shardID := common.GetShardIDFromLastByte(lastByte)
						if txTokenPrivacyDataMap[shardID] == nil {
							txTokenPrivacyDataMap[shardID] = make(map[common.Hash]*ContentCrossTokenPrivacyData)
						}
						if _, ok := txTokenPrivacyDataMap[shardID][customTokenPrivacyTx.TxTokenPrivacyData.PropertyID]; !ok {
							contentCrossTokenPrivacyData := cloneTxTokenPrivacyDataForCrossShard(customTokenPrivacyTx.TxTokenPrivacyData)
							txTokenPrivacyDataMap[shardID][customTokenPrivacyTx.TxTokenPrivacyData.PropertyID] = &contentCrossTokenPrivacyData
						}
						txTokenPrivacyDataMap[shardID][customTokenPrivacyTx.TxTokenPrivacyData.PropertyID].OutputCoin = append(txTokenPrivacyDataMap[shardID][customTokenPrivacyTx.TxTokenPrivacyData.PropertyID].OutputCoin, *outCoin)
					}
				}
			}
		}
	}
	//calcualte hash for each shard
	outputCoinHash := make([]common.Hash, common.MAX_SHARD_NUMBER)
	txTokenOutHash := make([]common.Hash, common.MAX_SHARD_NUMBER)
	txTokenPrivacyOutHash := make([]common.Hash, common.MAX_SHARD_NUMBER)
	combinedHash := make([]common.Hash, common.MAX_SHARD_NUMBER)
	for i := 0; i < common.MAX_SHARD_NUMBER; i++ {
		outputCoinHash[i] = calHashOutCoinCrossShard(outCoinEachShard[i])
		txTokenOutHash[i] = calHashTxTokenDataHashFromMap(txTokenDataEachShard[i])
		txTokenPrivacyOutHash[i] = calHashTxTokenPrivacyDataHashFromMap(txTokenPrivacyDataMap[i])

		tmpByte := append(append(outputCoinHash[i].GetBytes(), txTokenOutHash[i].GetBytes()...), txTokenPrivacyOutHash[i].GetBytes()...)
		combinedHash[i] = common.HashH(tmpByte)
	}
	return combinedHash
}

// helper function to get cross data (send to a shard) from list of transaction:
// 1. (Privacy) Constant: Output coin
// 2. Tx Custom Token: Tx Token Data
// 3. Privacy Custom Token: Token Data + Output coin
func getCrossShardData(txList []metadata.Transaction, shardID byte) ([]privacy.OutputCoin,
	[]transaction.TxTokenData,
	[]ContentCrossTokenPrivacyData,
) {
	coinList := []privacy.OutputCoin{}
	txTokenDataMap := make(map[common.Hash]*transaction.TxTokenData)
	txTokenPrivacyDataMap := make(map[common.Hash]*ContentCrossTokenPrivacyData)
	var txTokenDataList []transaction.TxTokenData
	var txTokenPrivacyDataList []ContentCrossTokenPrivacyData
	for _, tx := range txList {
		if tx.GetProof() != nil {
			for _, outCoin := range tx.GetProof().OutputCoins {
				lastByte := common.GetShardIDFromLastByte(outCoin.CoinDetails.GetPubKeyLastByte())
				if lastByte == shardID {
					coinList = append(coinList, *outCoin)
				}
			}
		}
		if tx.GetType() == common.TxCustomTokenType {
			customTokenTx := tx.(*transaction.TxCustomToken)
			for _, vout := range customTokenTx.TxTokenData.Vouts {
				lastByte := common.GetShardIDFromLastByte(vout.PaymentAddress.Pk[len(vout.PaymentAddress.Pk)-1])
				if lastByte == shardID {
					if _, ok := txTokenDataMap[customTokenTx.TxTokenData.PropertyID]; !ok {
						newTxTokenData := cloneTxTokenDataForCrossShard(customTokenTx.TxTokenData)
						txTokenDataMap[customTokenTx.TxTokenData.PropertyID] = &newTxTokenData
					}
					vouts := txTokenDataMap[customTokenTx.TxTokenData.PropertyID].Vouts
					vouts = append(vouts, vout)
					txTokenDataMap[customTokenTx.TxTokenData.PropertyID].Vouts = vouts
				}
			}
		}
		if tx.GetType() == common.TxCustomTokenPrivacyType {
			customTokenPrivacyTx := tx.(*transaction.TxCustomTokenPrivacy)
			if customTokenPrivacyTx.TxTokenPrivacyData.TxNormal.GetProof() != nil {
				for _, outCoin := range customTokenPrivacyTx.TxTokenPrivacyData.TxNormal.GetProof().OutputCoins {
					lastByte := common.GetShardIDFromLastByte(outCoin.CoinDetails.GetPubKeyLastByte())
					if lastByte == shardID {
						if _, ok := txTokenPrivacyDataMap[customTokenPrivacyTx.TxTokenPrivacyData.PropertyID]; !ok {
							contentCrossTokenPrivacyData := cloneTxTokenPrivacyDataForCrossShard(customTokenPrivacyTx.TxTokenPrivacyData)
							txTokenPrivacyDataMap[customTokenPrivacyTx.TxTokenPrivacyData.PropertyID] = &contentCrossTokenPrivacyData
						}
						txTokenPrivacyDataMap[customTokenPrivacyTx.TxTokenPrivacyData.PropertyID].OutputCoin = append(txTokenPrivacyDataMap[customTokenPrivacyTx.TxTokenPrivacyData.PropertyID].OutputCoin, *outCoin)
					}
				}
			}
		}
	}
	if len(txTokenDataMap) != 0 {
		for _, value := range txTokenDataMap {
			txTokenDataList = append(txTokenDataList, *value)
		}
		sort.SliceStable(txTokenDataList[:], func(i, j int) bool {
			return txTokenDataList[i].PropertyID.String() < txTokenDataList[j].PropertyID.String()
		})
	}
	if len(txTokenPrivacyDataMap) != 0 {
		for _, value := range txTokenPrivacyDataMap {
			txTokenPrivacyDataList = append(txTokenPrivacyDataList, *value)
		}
		sort.SliceStable(txTokenPrivacyDataList[:], func(i, j int) bool {
			return txTokenPrivacyDataList[i].PropertyID.String() < txTokenPrivacyDataList[j].PropertyID.String()
		})
	}
	return coinList, txTokenDataList, txTokenPrivacyDataList
}

/*
	Get output coin of transaction
	Check receiver last byte
	Append output coin to corresponding shard
*/
func getOutCoinCrossShard(txList []metadata.Transaction, shardID byte) []privacy.OutputCoin {
	coinList := []privacy.OutputCoin{}
	for _, tx := range txList {
		if tx.GetProof() != nil {
			for _, outCoin := range tx.GetProof().OutputCoins {
				lastByte := common.GetShardIDFromLastByte(outCoin.CoinDetails.GetPubKeyLastByte())
				if lastByte == shardID {
					coinList = append(coinList, *outCoin)
				}
			}
		}
	}
	return coinList
}

// helper function to get the hash of OutputCoins (send to a shard) from list of transaction
/*
	Get tx token data of transaction
	Check receiver (in vout) last byte
	Append tx token data to corresponding shard
*/
func getTxTokenDataCrossShard(txList []metadata.Transaction, shardID byte) []transaction.TxTokenData {
	txTokenDataMap := make(map[common.Hash]*transaction.TxTokenData)
	for _, tx := range txList {
		if tx.GetType() == common.TxCustomTokenType {
			customTokenTx := tx.(*transaction.TxCustomToken)
			for _, vout := range customTokenTx.TxTokenData.Vouts {
				lastByte := common.GetShardIDFromLastByte(vout.PaymentAddress.Pk[len(vout.PaymentAddress.Pk)-1])
				if lastByte == shardID {
					if _, ok := txTokenDataMap[customTokenTx.TxTokenData.PropertyID]; !ok {
						newTxTokenData := cloneTxTokenDataForCrossShard(customTokenTx.TxTokenData)
						txTokenDataMap[customTokenTx.TxTokenData.PropertyID] = &newTxTokenData
					}
					vouts := txTokenDataMap[customTokenTx.TxTokenData.PropertyID].Vouts
					vouts = append(vouts, vout)
					txTokenDataMap[customTokenTx.TxTokenData.PropertyID].Vouts = vouts
				}
			}
		}
	}
	var txTokenDataList []transaction.TxTokenData
	if len(txTokenDataMap) != 0 {
		for _, value := range txTokenDataMap {
			txTokenDataList = append(txTokenDataList, *value)
		}
		sort.SliceStable(txTokenDataList[:], func(i, j int) bool {
			return txTokenDataList[i].PropertyID.String() < txTokenDataList[j].PropertyID.String()
		})
	}
	return txTokenDataList
}
func getTxTokenPrivacyDataCrossShard(txList []metadata.Transaction, shardID byte) []ContentCrossTokenPrivacyData {
	txTokenPrivacyDataMap := make(map[common.Hash]*ContentCrossTokenPrivacyData)
	for _, tx := range txList {
		if tx.GetType() == common.TxCustomTokenPrivacyType {
			customTokenPrivacyTx := tx.(*transaction.TxCustomTokenPrivacy)
			if customTokenPrivacyTx.TxTokenPrivacyData.TxNormal.GetProof() != nil {
				for _, outCoin := range customTokenPrivacyTx.TxTokenPrivacyData.TxNormal.GetProof().OutputCoins {
					lastByte := common.GetShardIDFromLastByte(outCoin.CoinDetails.GetPubKeyLastByte())
					if lastByte == shardID {
						if _, ok := txTokenPrivacyDataMap[customTokenPrivacyTx.TxTokenPrivacyData.PropertyID]; !ok {
							contentCrossTokenPrivacyData := cloneTxTokenPrivacyDataForCrossShard(customTokenPrivacyTx.TxTokenPrivacyData)
							txTokenPrivacyDataMap[customTokenPrivacyTx.TxTokenPrivacyData.PropertyID] = &contentCrossTokenPrivacyData
						}
						txTokenPrivacyDataMap[customTokenPrivacyTx.TxTokenPrivacyData.PropertyID].OutputCoin = append(txTokenPrivacyDataMap[customTokenPrivacyTx.TxTokenPrivacyData.PropertyID].OutputCoin, *outCoin)
					}
				}
			}
		}
	}
	var txTokenPrivacyDataList []ContentCrossTokenPrivacyData
	if len(txTokenPrivacyDataMap) != 0 {
		for _, value := range txTokenPrivacyDataMap {
			txTokenPrivacyDataList = append(txTokenPrivacyDataList, *value)
		}
		sort.SliceStable(txTokenPrivacyDataList[:], func(i, j int) bool {
			return txTokenPrivacyDataList[i].PropertyID.String() < txTokenPrivacyDataList[j].PropertyID.String()
		})
	}
	return txTokenPrivacyDataList
}
func calHashOutCoinCrossShard(outCoins []privacy.OutputCoin) common.Hash {
	tmpByte := []byte{}
	var outputCoinHash common.Hash
	if len(outCoins) != 0 {
		for _, outCoin := range outCoins {
			coin := &outCoin

			tmpByte = append(tmpByte, coin.Bytes()...)
		}
		outputCoinHash = common.HashH(tmpByte)
	} else {
		outputCoinHash = common.HashH([]byte(""))
	}
	return outputCoinHash
}
func calHashTxTokenDataHashFromMap(txTokenDataMap map[common.Hash]*transaction.TxTokenData) common.Hash {
	if len(txTokenDataMap) == 0 {
		return common.HashH([]byte(""))
	}
	var txTokenDataList []transaction.TxTokenData
	for _, value := range txTokenDataMap {
		txTokenDataList = append(txTokenDataList, *value)
	}
	sort.SliceStable(txTokenDataList[:], func(i, j int) bool {
		return txTokenDataList[i].PropertyID.String() < txTokenDataList[j].PropertyID.String()
	})
	return calHashTxTokenDataHashList(txTokenDataList)
}
func calHashTxTokenDataHashList(txTokenDataList []transaction.TxTokenData) common.Hash {
	tmpByte := []byte{}
	if len(txTokenDataList) != 0 {
		for _, txTokenData := range txTokenDataList {
			tempHash, _ := txTokenData.Hash()
			tmpByte = append(tmpByte, tempHash.GetBytes()...)
		}
	} else {
		return common.HashH([]byte(""))
	}

	return common.HashH(tmpByte)
}
func calHashTxTokenPrivacyDataHashFromMap(txTokenPrivacyDataMap map[common.Hash]*ContentCrossTokenPrivacyData) common.Hash {
	if len(txTokenPrivacyDataMap) == 0 {
		return common.HashH([]byte(""))
	}
	var txTokenPrivacyDataList []ContentCrossTokenPrivacyData
	for _, value := range txTokenPrivacyDataMap {
		txTokenPrivacyDataList = append(txTokenPrivacyDataList, *value)
	}
	sort.SliceStable(txTokenPrivacyDataList[:], func(i, j int) bool {
		return txTokenPrivacyDataList[i].PropertyID.String() < txTokenPrivacyDataList[j].PropertyID.String()
	})
	return calHashTxTokenPrivacyDataHashList(txTokenPrivacyDataList)
}
func calHashTxTokenPrivacyDataHashList(txTokenPrivacyDataList []ContentCrossTokenPrivacyData) common.Hash {
	tmpByte := []byte{}
	if len(txTokenPrivacyDataList) != 0 {
		for _, txTokenPrivacyData := range txTokenPrivacyDataList {
			tempHash := txTokenPrivacyData.Hash()
			tmpByte = append(tmpByte, tempHash.GetBytes()...)

		}
	} else {
		return common.HashH([]byte(""))
	}
	return common.HashH(tmpByte)
}

func cloneTxTokenDataForCrossShard(txTokenData transaction.TxTokenData) transaction.TxTokenData {
	newTxTokenData := transaction.TxTokenData{
		PropertyID:     txTokenData.PropertyID,
		PropertyName:   txTokenData.PropertyName,
		PropertySymbol: txTokenData.PropertySymbol,
		Mintable:       txTokenData.Mintable,
		Amount:         txTokenData.Amount,
		Type:           transaction.CustomTokenCrossShard,
	}
	newTxTokenData.Vins = []transaction.TxTokenVin{}
	newTxTokenData.Vouts = []transaction.TxTokenVout{}
	return newTxTokenData
}
func cloneTxTokenPrivacyDataForCrossShard(txTokenPrivacyData transaction.TxTokenPrivacyData) ContentCrossTokenPrivacyData {
	newContentCrossTokenPrivacyData := ContentCrossTokenPrivacyData{
		PropertyID:     txTokenPrivacyData.PropertyID,
		PropertyName:   txTokenPrivacyData.PropertyName,
		PropertySymbol: txTokenPrivacyData.PropertySymbol,
		Mintable:       txTokenPrivacyData.Mintable,
		Amount:         txTokenPrivacyData.Amount,
		Type:           transaction.CustomTokenCrossShard,
	}
	newContentCrossTokenPrivacyData.OutputCoin = []privacy.OutputCoin{}
	return newContentCrossTokenPrivacyData
}
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
	merkleTree := merkle.BuildMerkleTreeOfHashes(crossOutputCoinHashes, len(crossOutputCoinHashes))
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
func CreateMerkleCrossTransaction(crossTransactions map[byte][]CrossTransaction) (*common.Hash, error) {
	if len(crossTransactions) == 0 {
		res, err := GenerateZeroValueHash()
		return &res, err
	}
	keys := []int{}
	crossTransactionHashes := []*common.Hash{}
	for k := range crossTransactions {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, value := range crossTransactions[byte(shardID)] {
			hash := value.Hash()
			crossTransactionHashes = append(crossTransactionHashes, &hash)
		}
	}
	merkle := Merkle{}
	merkleTree := merkle.BuildMerkleTreeOfHashes(crossTransactionHashes, len(crossTransactionHashes))
	return merkleTree[len(merkleTree)-1], nil
}

func VerifyMerkleCrossTransaction(crossTransactions map[byte][]CrossTransaction, rootHash common.Hash) bool {
	res, err := CreateMerkleCrossTransaction(crossTransactions)
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

//=======================================END CROSS SHARD UTIL
