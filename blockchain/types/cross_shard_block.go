package types

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/transaction"

	"github.com/incognitochain/incognito-chain/common"
)

type CrossShardBlock struct {
	ValidationData  string `json:"ValidationData"`
	Header          ShardHeader
	ToShardID       byte
	MerklePathShard []common.Hash
	// Cross Shard data for PRV
	CrossOutputCoin []privacy.Coin
	// Cross Shard For Custom token privacy
	CrossTxTokenPrivacyData []ContentCrossShardTokenPrivacyData
}

type CrossOutputCoin struct {
	BlockHeight uint64
	BlockHash   common.Hash
	OutputCoin  []coin.Coin
}

type CrossTokenPrivacyData struct {
	BlockHeight      uint64
	BlockHash        common.Hash
	TokenPrivacyData []ContentCrossShardTokenPrivacyData
}

type CrossTransaction struct {
	BlockHeight      uint64
	BlockHash        common.Hash
	TokenPrivacyData []ContentCrossShardTokenPrivacyData
	OutputCoin       []coin.Coin
}

type ContentCrossShardTokenPrivacyData struct {
	OutputCoin     []coin.Coin
	PropertyID     common.Hash // = hash of TxCustomTokenprivacy data
	PropertyName   string
	PropertySymbol string
	Type           int    // action type
	Mintable       bool   // default false
	Amount         uint64 // init amount
}

type CrossShardTokenPrivacyMetaData struct {
	TokenID        common.Hash
	PropertyName   string
	PropertySymbol string
	Type           int    // action type
	Mintable       bool   // default false
	Amount         uint64 // init amount
}

func (block CrossShardBlock) GetProducer() string {
	return block.Header.Producer
}

func (block CrossShardBlock) GetVersion() int {
	return block.Header.Version
}

func (block CrossShardBlock) GetHeight() uint64 {
	return block.Header.Height
}

func (block CrossShardBlock) GetBeaconHeight() uint64 {
	return block.Header.BeaconHeight
}

// consensus interface
func (block CrossShardBlock) ProposeHash() *common.Hash {
	panic("Not implement")
}

func (block *CrossShardBlock) AddValidationField(validationData string) {
	panic("Not implement")
}

func (block CrossShardBlock) FullHashString() string {
	//TODO implement me
	panic("implement me")
}

//end consensus interface

func (block CrossShardBlock) GetShardID() int {
	return int(block.Header.ShardID)
}

func (block CrossShardBlock) GetValidationField() string {
	return block.ValidationData
}
func (block CrossShardBlock) SetValidationField(string) {
	panic("should not come here!")
}

func (block CrossShardBlock) GetRound() int {
	return block.Header.Round
}

func (block CrossShardBlock) GetRoundKey() string {
	return fmt.Sprint(block.Header.Height, "_", block.Header.Round)
}

func (block CrossShardBlock) GetInstructions() [][]string {
	return [][]string{}
}

func (block ShardBlock) GetConsensusType() string {
	return block.Header.ConsensusType
}

func (block CrossShardBlock) GetConsensusType() string {
	return block.Header.ConsensusType
}

func (crossShardBlock CrossShardBlock) CommitteeFromBlock() common.Hash {
	return common.Hash{}
}

func (crossShardBlock CrossShardBlock) GetProposer() string {
	return crossShardBlock.Header.Proposer
}

func (crossShardBlock CrossShardBlock) GetProposeTime() int64 {
	return crossShardBlock.Header.ProposeTime
}

func (crossShardBlock CrossShardBlock) GetProduceTime() int64 {
	return crossShardBlock.Header.Timestamp
}

func (crossShardBlock CrossShardBlock) GetCurrentEpoch() uint64 {
	return crossShardBlock.Header.Epoch
}

func (crossShardBlock CrossShardBlock) GetPrevHash() common.Hash {
	return crossShardBlock.Header.PreviousBlockHash
}

func (crossShardBlock *CrossShardBlock) Hash() *common.Hash {
	hash := crossShardBlock.Header.Hash()
	return &hash
}

func (crossShardBlock *CrossShardBlock) GetAggregateRootHash() common.Hash {
	panic("do not call this function")
}

func (crossShardBlock *CrossShardBlock) GetFinalityHeight() uint64 {
	panic("do not call this function")
}

func (crossShardBlock *CrossShardBlock) UnmarshalJSON(data []byte) error {
	type Alias CrossShardBlock
	temp := &struct {
		CrossOutputCoin []json.RawMessage
		*Alias
	}{
		Alias: (*Alias)(crossShardBlock),
	}

	if err := json.Unmarshal(data, temp); err != nil {
		return fmt.Errorf("UnmarshalJSON crossShardBlock. Error %v", err)
	}

	outputCoinList, err := coin.ParseCoinsFromBytes(temp.CrossOutputCoin)
	if err != nil {
		return fmt.Errorf("UnmarshalJSON Cannot parse crossOutputCoins. Error %v", err)
	}

	crossShardBlock.CrossOutputCoin = outputCoinList
	return nil
}

func (contentCrossShardTokenPrivacyData ContentCrossShardTokenPrivacyData) Bytes() []byte {
	res := []byte{}
	for _, item := range contentCrossShardTokenPrivacyData.OutputCoin {
		res = append(res, item.Bytes()...)
	}
	res = append(res, contentCrossShardTokenPrivacyData.PropertyID.GetBytes()...)
	res = append(res, []byte(contentCrossShardTokenPrivacyData.PropertyName)...)
	res = append(res, []byte(contentCrossShardTokenPrivacyData.PropertySymbol)...)
	typeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(typeBytes, uint32(contentCrossShardTokenPrivacyData.Type))
	res = append(res, typeBytes...)
	amountBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(amountBytes, uint32(contentCrossShardTokenPrivacyData.Amount))
	res = append(res, amountBytes...)
	if contentCrossShardTokenPrivacyData.Mintable {
		res = append(res, []byte("true")...)
	} else {
		res = append(res, []byte("false")...)
	}
	return res
}

func (contentCrossShardTokenPrivacyData ContentCrossShardTokenPrivacyData) Hash() common.Hash {
	return common.HashH(contentCrossShardTokenPrivacyData.Bytes())
}

func (contentCrossShardTokenPrivacyData *ContentCrossShardTokenPrivacyData) UnmarshalJSON(data []byte) error {
	type Alias ContentCrossShardTokenPrivacyData
	temp := &struct {
		OutputCoin []json.RawMessage
		*Alias
	}{
		Alias: (*Alias)(contentCrossShardTokenPrivacyData),
	}
	if err := json.Unmarshal(data, temp); err != nil {
		fmt.Errorf("UnmarshalJSON ContentCrossShardTokenPrivacyData", err)
		return err
	}
	outputCoinList, err := coin.ParseCoinsFromBytes(temp.OutputCoin)
	if err != nil {
		fmt.Errorf("UnmarshalJSON Cannot parse crossOutputCoins", err)
		return err
	}
	contentCrossShardTokenPrivacyData.OutputCoin = outputCoinList
	return nil
}

func (crossOutputCoin CrossOutputCoin) Hash() common.Hash {
	res := []byte{}
	res = append(res, crossOutputCoin.BlockHash.GetBytes()...)
	for _, coins := range crossOutputCoin.OutputCoin {
		res = append(res, coins.Bytes()...)
	}
	return common.HashH(res)
}

func (crossOutputCoin *CrossOutputCoin) UnmarshalJSON(data []byte) error {
	type Alias CrossOutputCoin
	temp := &struct {
		OutputCoin []json.RawMessage
		*Alias
	}{
		Alias: (*Alias)(crossOutputCoin),
	}
	if err := json.Unmarshal(data, temp); err != nil {
		fmt.Errorf("UnmarshalJSON CrossOutputCoin", err)
		return err
	}
	outputCoinList, err := coin.ParseCoinsFromBytes(temp.OutputCoin)
	if err != nil {
		fmt.Errorf("UnmarshalJSON Cannot parse CrossOutputCoin", err)
		return err
	}
	crossOutputCoin.OutputCoin = outputCoinList
	return nil
}

func (crossTransaction CrossTransaction) Bytes() []byte {
	res := []byte{}
	res = append(res, crossTransaction.BlockHash.GetBytes()...)
	for _, coins := range crossTransaction.OutputCoin {
		res = append(res, coins.Bytes()...)
	}
	for _, coins := range crossTransaction.TokenPrivacyData {
		res = append(res, coins.Bytes()...)
	}
	return res
}

func (crossTransaction CrossTransaction) Hash() common.Hash {
	return common.HashH(crossTransaction.Bytes())
}

func (crossTransaction *CrossTransaction) UnmarshalJSON(data []byte) error {
	type Alias CrossTransaction
	temp := &struct {
		OutputCoin []json.RawMessage
		*Alias
	}{
		Alias: (*Alias)(crossTransaction),
	}
	if err := json.Unmarshal(data, temp); err != nil {
		fmt.Errorf("UnmarshalJSON CrossTransaction", string(data), err)
		return err
	}
	outputCoinList, err := coin.ParseCoinsFromBytes(temp.OutputCoin)
	if err != nil {
		fmt.Errorf("UnmarshalJSON Cannot parse CrossTransaction", err)
		return err
	}
	crossTransaction.OutputCoin = outputCoinList
	return nil
}

// getCrossShardData get cross data (send to a shard) from list of transaction:
// 1. (Privacy) PRV: Output coin
// 2. Tx Custom Token: Tx Token Data
// 3. Privacy Custom Token: Token Data + Output coin
func GetCrossShardData(txList []metadata.Transaction, shardID byte) ([]privacy.Coin, []ContentCrossShardTokenPrivacyData, error) {
	coinList := []coin.Coin{}
	txTokenPrivacyDataMap := make(map[common.Hash]*ContentCrossShardTokenPrivacyData)
	var txTokenPrivacyDataList []ContentCrossShardTokenPrivacyData
	for _, tx := range txList {
		var prvProof privacy.Proof

		if tx.GetType() == common.TxCustomTokenPrivacyType || tx.GetType() == common.TxTokenConversionType {
			customTokenPrivacyTx, ok := tx.(transaction.TransactionToken)
			if !ok {
				return nil, nil, errors.New("Cannot cast transaction")
			}
			prvProof = customTokenPrivacyTx.GetTxBase().GetProof()
			txTokenData := customTokenPrivacyTx.GetTxTokenData()
			txTokenProof := txTokenData.TxNormal.GetProof()
			if txTokenProof != nil {
				for _, outCoin := range txTokenProof.GetOutputCoins() {
					coinShardID, err := outCoin.GetShardID()
					if err == nil && coinShardID == shardID {
						if _, ok := txTokenPrivacyDataMap[txTokenData.PropertyID]; !ok {
							contentCrossTokenPrivacyData := CloneTxTokenPrivacyDataForCrossShard(txTokenData)
							txTokenPrivacyDataMap[txTokenData.PropertyID] = &contentCrossTokenPrivacyData
						}
						txTokenPrivacyDataMap[txTokenData.PropertyID].OutputCoin = append(txTokenPrivacyDataMap[txTokenData.PropertyID].OutputCoin, outCoin)
					}
				}
			}
		} else {
			prvProof = tx.GetProof()
		}
		if prvProof != nil {
			for _, outCoin := range prvProof.GetOutputCoins() {
				coinShardID, err := outCoin.GetShardID()
				if err == nil && coinShardID == shardID {
					coinList = append(coinList, outCoin)
				}
			}
		}
	}
	if len(txTokenPrivacyDataMap) != 0 {
		for _, value := range txTokenPrivacyDataMap {
			txTokenPrivacyDataList = append(txTokenPrivacyDataList, *value)
		}
		sort.SliceStable(txTokenPrivacyDataList[:], func(i, j int) bool {
			return txTokenPrivacyDataList[i].PropertyID.String() < txTokenPrivacyDataList[j].PropertyID.String()
		})
	}
	return coinList, txTokenPrivacyDataList, nil
}

func CloneTxTokenPrivacyDataForCrossShard(txTokenPrivacyData transaction.TxTokenData) ContentCrossShardTokenPrivacyData {
	newContentCrossTokenPrivacyData := ContentCrossShardTokenPrivacyData{
		PropertyID:     txTokenPrivacyData.PropertyID,
		PropertyName:   txTokenPrivacyData.PropertyName,
		PropertySymbol: txTokenPrivacyData.PropertySymbol,
		Mintable:       txTokenPrivacyData.Mintable,
		Amount:         txTokenPrivacyData.Amount,
		Type:           transaction.CustomTokenCrossShard,
	}
	newContentCrossTokenPrivacyData.OutputCoin = []coin.Coin{}
	return newContentCrossTokenPrivacyData
}

func updateTxEnvWithBlock(sBlk *ShardBlock, tx metadata.Transaction) metadata.ValidationEnviroment {
	valEnv := transaction.WithShardHeight(tx.GetValidationEnv(), sBlk.GetHeight())
	valEnv = transaction.WithBeaconHeight(valEnv, sBlk.Header.BeaconHeight)
	valEnv = transaction.WithConfirmedTime(valEnv, sBlk.GetProduceTime())
	return valEnv
}

func CreateAllCrossShardBlock(shardBlock *ShardBlock, activeShards int) map[byte]*CrossShardBlock {
	allCrossShard := make(map[byte]*CrossShardBlock)
	if activeShards == 1 {
		return allCrossShard
	}
	for i := 0; i < activeShards; i++ {
		shardID := common.GetShardIDFromLastByte(byte(i))
		if shardID != shardBlock.Header.ShardID {
			crossShard, err := CreateCrossShardBlock(shardBlock, shardID)
			if crossShard != nil {
				log.Printf("Create CrossShardBlock from Shard %+v to Shard %+v: %+v \n", shardBlock.Header.ShardID, shardID, crossShard)
			}
			if crossShard != nil && err == nil {
				allCrossShard[byte(i)] = crossShard
			}
		}
	}
	return allCrossShard
}

func CreateCrossShardBlock(shardBlock *ShardBlock, shardID byte) (*CrossShardBlock, error) {
	crossShard := &CrossShardBlock{}
	crossOutputCoin, crossCustomTokenPrivacyData, err := GetCrossShardData(shardBlock.Body.Transactions, shardID)
	if err != nil {
		return nil, err
	}
	// Return nothing if nothing to cross
	if len(crossOutputCoin) == 0 && len(crossCustomTokenPrivacyData) == 0 {
		return nil, errors.New("No cross Outputcoin, Cross Custom Token, Cross Custom Token Privacy")
	}
	merklePathShard, merkleShardRoot := GetMerklePathCrossShard(shardBlock.Body.Transactions, shardID)
	if merkleShardRoot != shardBlock.Header.ShardTxRoot {
		return crossShard, fmt.Errorf("Expect Shard Tx Root To be %+v but get %+v", shardBlock.Header.ShardTxRoot, merkleShardRoot)
	}
	crossShard.ValidationData = shardBlock.ValidationData
	crossShard.Header = shardBlock.Header
	crossShard.MerklePathShard = merklePathShard
	crossShard.CrossOutputCoin = crossOutputCoin
	crossShard.CrossTxTokenPrivacyData = crossCustomTokenPrivacyData
	crossShard.ToShardID = shardID
	return crossShard, nil
}

// VerifyCrossShardBlockUTXO Calculate Final Hash as Hash of:
//  1. CrossTransactionFinalHash
//  2. TxTokenDataVoutFinalHash
//  3. CrossTxTokenPrivacyData
//
// These hashes will be calculated as comment in getCrossShardDataHash function
func VerifyCrossShardBlockUTXO(block *CrossShardBlock) bool {
	var outputCoinHash common.Hash
	var txTokenDataHash common.Hash
	var txTokenPrivacyDataHash common.Hash
	outCoins := block.CrossOutputCoin
	outputCoinHash = calHashOutCoinCrossShard(outCoins)
	txTokenDataHash = calHashTxTokenDataHashList()
	txTokenPrivacyDataList := block.CrossTxTokenPrivacyData
	txTokenPrivacyDataHash = calHashTxTokenPrivacyDataHashList(txTokenPrivacyDataList)
	tmpByte := append(append(outputCoinHash.GetBytes(), txTokenDataHash.GetBytes()...), txTokenPrivacyDataHash.GetBytes()...)
	finalHash := common.HashH(tmpByte)
	return Merkle{}.VerifyMerkleRootFromMerklePath(finalHash, block.MerklePathShard, block.Header.ShardTxRoot, block.ToShardID)
}

func (block CrossShardBlock) Type() string {
	return common.ShardChainKey
}

func (block CrossShardBlock) BodyHash() common.Hash {
	return common.Hash{}
}
