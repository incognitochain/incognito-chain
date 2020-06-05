package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/privacy/coin"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

type ShardBlock struct {
	ValidationData string `json:"ValidationData"`
	Body           ShardBody
	Header         ShardHeader
}

func (shardBlock *ShardBlock) GetProposer() string {
	return shardBlock.Header.Proposer
}

func (shardBlock *ShardBlock) GetProposeTime() int64 {
	return shardBlock.Header.ProposeTime
}

func (shardBlock *ShardBlock) GetProduceTime() int64 {
	return shardBlock.Header.Timestamp
}

func (shardBlock *ShardBlock) GetShardID() int {
	return int(shardBlock.Header.ShardID)
}
func (shardBlock *ShardBlock) GetPrevHash() common.Hash {
	return shardBlock.Header.PreviousBlockHash
}

type ShardToBeaconBlock struct {
	ValidationData string `json:"ValidationData"`
	Instructions   [][]string
	Header         ShardHeader
}

func (shardToBeaconBlock *ShardToBeaconBlock) GetPrevHash() common.Hash {
	return shardToBeaconBlock.Header.PreviousBlockHash
}

func (shardToBeaconBlock *ShardToBeaconBlock) GetShardID() int {
	return int(shardToBeaconBlock.Header.ShardID)
}

type CrossShardBlock struct {
	// AggregatedSig   string  `json:"AggregatedSig"`
	// R               string  `json:"R"`
	// ValidatorsIdx   [][]int `json:"ValidatorsIdx"` //[0]: R | [1]:AggregatedSig
	// ProducerSig     string  `json:"ProducerSig"`
	ValidationData  string `json:"ValidationData"`
	Header          ShardHeader
	ToShardID       byte
	MerklePathShard []common.Hash
	// Cross Shard data for PRV
	CrossOutputCoin []coin.Coin
	// Cross Shard For Custom token privacy
	CrossTxTokenPrivacyData []ContentCrossShardTokenPrivacyData
}

func NewShardBlock() *ShardBlock {
	return &ShardBlock{
		Header: ShardHeader{},
		Body: ShardBody{
			Instructions:      [][]string{},
			CrossTransactions: make(map[byte][]CrossTransaction),
			Transactions:      make([]metadata.Transaction, 0),
		},
	}
}
func NewShardBlockWithHeader(header ShardHeader) *ShardBlock {
	return &ShardBlock{
		Header: header,
		Body: ShardBody{
			Instructions:      [][]string{},
			CrossTransactions: make(map[byte][]CrossTransaction),
			Transactions:      make([]metadata.Transaction, 0),
		},
	}
}
func NewShardBlockWithBody(body ShardBody) *ShardBlock {
	return &ShardBlock{
		Header: ShardHeader{},
		Body:   body,
	}
}
func NewShardBlockFull(header ShardHeader, body ShardBody) *ShardBlock {
	return &ShardBlock{
		Header: header,
		Body:   body,
	}
}

func (shardBlock *ShardBlock) BuildShardBlockBody(instructions [][]string, crossTransaction map[byte][]CrossTransaction, transactions []metadata.Transaction) {
	shardBlock.Body.Instructions = append(shardBlock.Body.Instructions, instructions...)
	shardBlock.Body.CrossTransactions = crossTransaction
	shardBlock.Body.Transactions = append(shardBlock.Body.Transactions, transactions...)
}

func (crossShardBlock *CrossShardBlock) UnmarshalJSON(data []byte) error {
	type Alias CrossShardBlock
	temp := &struct {
		CrossOutputCoin []string
		*Alias
	}{
		Alias: (*Alias)(crossShardBlock),
	}

	if err := json.Unmarshal(data, temp); err != nil {
		Logger.log.Error("UnmarshalJSON crossShardBlock", string(data))
		return err
	}

	outputCoinList, err := coin.ParseCoinsStr(temp.CrossOutputCoin)
	if err != nil {
		Logger.log.Error("UnmarshalJSON Cannot parse crossOutputCoins", err)
		return err
	}
	for i := 0; i < len(temp.CrossOutputCoin); i++ {
		fmt.Println("Detail of CrossOutputCoins", outputCoinList[i])
	}
	crossShardBlock.CrossOutputCoin = outputCoinList
	return nil
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

func (shardToBeaconBlock ShardToBeaconBlock) GetCurrentEpoch() uint64 {
	return shardToBeaconBlock.Header.Epoch
}

func (shardToBeaconBlock *ShardToBeaconBlock) Hash() *common.Hash {
	hash := shardToBeaconBlock.Header.Hash()
	return &hash
}

func (shardBlock ShardBlock) Hash() *common.Hash {
	hash := shardBlock.Header.Hash()
	return &hash
}
func (shardBlock *ShardBlock) validateSanityData() (bool, error) {
	//Check Header
	if shardBlock.Header.Height == 1 && len(shardBlock.ValidationData) != 0 {
		return false, NewBlockChainError(ShardBlockSanityError, errors.New("Expect Shard Block with Height 1 to not have validationData"))
	}
	// producer address must have 66 bytes: 33-byte public key, 33-byte transmission key
	if shardBlock.Header.Height > 1 && len(shardBlock.ValidationData) == 0 {
		return false, NewBlockChainError(ShardBlockSanityError, errors.New("Expect Shard Block to have validationData"))
	}
	if int(shardBlock.Header.ShardID) < 0 || int(shardBlock.Header.ShardID) > 256 {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block ShardID in range 0 - 255 but get %+v ", shardBlock.Header.ShardID))
	}
	if shardBlock.Header.Version < SHARD_BLOCK_VERSION {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Version greater or equal than %+v but get %+v ", SHARD_BLOCK_VERSION, shardBlock.Header.Version))
	}
	if len(shardBlock.Header.PreviousBlockHash[:]) != common.HashSize {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Previous Hash in the right format"))
	}
	if shardBlock.Header.Height < 1 {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Height to be greater than 0"))
	}
	if shardBlock.Header.Height == 1 && !shardBlock.Header.PreviousBlockHash.IsEqual(&common.Hash{}) {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block with Height 1 (first block) have Zero Hash Value"))
	}
	if shardBlock.Header.Height > 1 && shardBlock.Header.PreviousBlockHash.IsEqual(&common.Hash{}) {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block with Height greater than 1 have Non-Zero Hash Value"))
	}
	if shardBlock.Header.Round < 1 {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Round greater or equal than 1"))
	}
	if shardBlock.Header.Epoch < 1 {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Epoch greater or equal than 1"))
	}
	if shardBlock.Header.Timestamp <= 0 {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Time greater than 0"))
	}
	if len(shardBlock.Header.TxRoot[:]) != common.HashSize {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Tx Root in the right format"))
	}
	if len(shardBlock.Header.ShardTxRoot[:]) != common.HashSize {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Shard Tx Root in the right format"))
	}
	if len(shardBlock.Header.CrossTransactionRoot[:]) != common.HashSize {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Cross Transaction Root in the right format"))
	}
	if len(shardBlock.Header.InstructionsRoot[:]) != common.HashSize {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Instructions Root in the right format"))
	}
	if len(shardBlock.Header.CommitteeRoot[:]) != common.HashSize {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Committee Root in the right format"))
	}
	if shardBlock.Header.Height == 1 && !shardBlock.Header.CommitteeRoot.IsEqual(&common.Hash{}) {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block with Height 1 have Zero Hash Value"))
	}
	if shardBlock.Header.Height > 1 && shardBlock.Header.CommitteeRoot.IsEqual(&common.Hash{}) {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block with Height greater than 1 have Non-Zero Hash Value"))
	}
	if len(shardBlock.Header.PendingValidatorRoot[:]) != common.HashSize {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Committee Root in the right format"))
	}
	if len(shardBlock.Header.StakingTxRoot[:]) != common.HashSize {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Staking Tx Root in the right format"))
	}
	if len(shardBlock.Header.CrossShardBitMap) > 254 {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Cross Shard Length Less Than 255"))
	}
	if shardBlock.Header.BeaconHeight < 1 {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block has Beacon Height greater or equal than 1"))
	}
	//if shardBlock.Header.BeaconHeight == 1 && !shardBlock.Header.BeaconHash.IsPointEqual(&common.Hash{}) {
	//	return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block with Beacon Height 1 have Zero Hash Value"))
	//}
	if shardBlock.Header.BeaconHeight > 1 && shardBlock.Header.BeaconHash.IsEqual(&common.Hash{}) {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block with Beacon Height greater or equal than 1 have Non-Zero Hash Value"))
	}
	if shardBlock.Header.TotalTxsFee == nil {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Total Txs Fee have nil value"))
	}
	if len(shardBlock.Header.InstructionMerkleRoot[:]) != common.HashSize {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Instruction Merkle Root in the right format"))
	}
	// body
	if shardBlock.Body.Instructions == nil {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Instruction is not nil"))
	}
	if len(shardBlock.Body.Instructions) != 0 && shardBlock.Header.InstructionMerkleRoot.IsEqual(&common.Hash{}) {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Instruction Merkle Root have Non-Zero Hash Value because Instrucstion List is not empty"))
	}
	if shardBlock.Body.CrossTransactions == nil {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Cross Transactions Map is not nil"))
	}
	if len(shardBlock.Body.CrossTransactions) != 0 && shardBlock.Header.CrossTransactionRoot.IsEqual(&common.Hash{}) {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Cross Transaction Root have Non-Zero Hash Value because Cross Transaction List is not empty"))
	}
	if shardBlock.Body.Transactions == nil {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Transactions is not nil"))
	}
	if len(shardBlock.Body.Transactions) != 0 && shardBlock.Header.TxRoot.IsEqual(&common.Hash{}) {
		return false, NewBlockChainError(ShardBlockSanityError, fmt.Errorf("Expect Shard Block Tx Root have Non-Zero Hash Value because Transactions List is not empty"))
	}
	return true, nil
}

func (shardBlock *ShardBlock) UnmarshalJSON(data []byte) error {
	tempShardBlock := &struct {
		ValidationData string `json:"ValidationData"`
		Header         ShardHeader
		Body           *json.RawMessage
	}{}
	err := json.Unmarshal(data, &tempShardBlock)
	if err != nil {
		return NewBlockChainError(UnmashallJsonShardBlockError, err)
	}
	shardBlock.ValidationData = tempShardBlock.ValidationData

	blkBody := ShardBody{}
	err = blkBody.UnmarshalJSON(*tempShardBlock.Body)
	if err != nil {
		return NewBlockChainError(UnmashallJsonShardBlockError, err)
	}
	shardBlock.Header = tempShardBlock.Header
	if shardBlock.Body.Transactions == nil {
		shardBlock.Body.Transactions = []metadata.Transaction{}
	}
	if shardBlock.Body.Instructions == nil {
		shardBlock.Body.Instructions = [][]string{}
	}
	if shardBlock.Body.CrossTransactions == nil {
		shardBlock.Body.CrossTransactions = make(map[byte][]CrossTransaction)
	}
	if shardBlock.Header.TotalTxsFee == nil {
		shardBlock.Header.TotalTxsFee = make(map[common.Hash]uint64)
	}
	if ok, err := shardBlock.validateSanityData(); !ok || err != nil {
		// panic(string(data) + err.Error())
		return NewBlockChainError(UnmashallJsonShardBlockError, err)
	}
	shardBlock.Body = blkBody
	return nil
}

// /*
// AddTransaction adds a new transaction into block
// */
// // #1 - tx
func (shardBlock *ShardBlock) AddTransaction(tx metadata.Transaction) error {
	if shardBlock.Body.Transactions == nil {
		return NewBlockChainError(UnExpectedError, errors.New("not init tx arrays"))
	}
	shardBlock.Body.Transactions = append(shardBlock.Body.Transactions, tx)
	return nil
}

func (shardBlock *ShardBlock) CreateShardToBeaconBlock(bc *BlockChain) *ShardToBeaconBlock {
	if bc.IsTest {
		return &ShardToBeaconBlock{}
	}
	block := ShardToBeaconBlock{}
	block.ValidationData = shardBlock.ValidationData
	block.Header = shardBlock.Header
	blockInstructions := shardBlock.Body.Instructions
	previousShardBlockByte, err := rawdbv2.GetShardBlockByHash(bc.GetShardChainDatabase(shardBlock.Header.ShardID), shardBlock.Header.PreviousBlockHash)
	if err != nil {
		Logger.log.Errorf("[S2B] CreateShardToBeaconBlock return err:", err)
		return nil
	}
	previousShardBlock := ShardBlock{}
	err = json.Unmarshal(previousShardBlockByte, &previousShardBlock)
	if err != nil {
		Logger.log.Errorf("[S2B] CreateShardToBeaconBlock return err:", err)
		return nil
	}
	instructions, err := CreateShardInstructionsFromTransactionAndInstruction(shardBlock.Body.Transactions, bc, shardBlock.Header.ShardID)
	if err != nil {
		Logger.log.Errorf("[S2B] CreateShardToBeaconBlock return err:", err)
		return nil
	}

	block.Instructions = append(instructions, blockInstructions...)
	return &block
}

func (shardBlock *ShardBlock) CreateAllCrossShardBlock(activeShards int) map[byte]*CrossShardBlock {
	allCrossShard := make(map[byte]*CrossShardBlock)
	if activeShards == 1 {
		return allCrossShard
	}
	for i := 0; i < activeShards; i++ {
		shardID := common.GetShardIDFromLastByte(byte(i))
		if shardID != shardBlock.Header.ShardID {
			crossShard, err := shardBlock.CreateCrossShardBlock(shardID)
			if crossShard != nil {
				Logger.log.Criticalf("Create CrossShardBlock from Shard %+v to Shard %+v: %+v \n", shardBlock.Header.ShardID, shardID, crossShard)
			}
			if crossShard != nil && err == nil {
				allCrossShard[byte(i)] = crossShard
			}
		}
	}
	return allCrossShard
}

func (shardBlock ShardBlock) CreateCrossShardBlock(shardID byte) (*CrossShardBlock, error) {
	crossShard := &CrossShardBlock{}
	crossOutputCoin, crossCustomTokenPrivacyData := getCrossShardData(shardBlock.Body.Transactions, shardID)
	// Return nothing if nothing to cross
	if len(crossOutputCoin) == 0 && len(crossCustomTokenPrivacyData) == 0 {
		return nil, NewBlockChainError(CreateCrossShardBlockError, errors.New("No cross Outputcoin, Cross Custom Token, Cross Custom Token Privacy"))
	}
	merklePathShard, merkleShardRoot := GetMerklePathCrossShard(shardBlock.Body.Transactions, shardID)
	if merkleShardRoot != shardBlock.Header.ShardTxRoot {
		return crossShard, NewBlockChainError(VerifyCrossShardBlockShardTxRootError, fmt.Errorf("Expect Shard Tx Root To be %+v but get %+v", shardBlock.Header.ShardTxRoot, merkleShardRoot))
	}
	crossShard.ValidationData = shardBlock.ValidationData
	crossShard.Header = shardBlock.Header
	crossShard.MerklePathShard = merklePathShard
	crossShard.CrossOutputCoin = crossOutputCoin
	crossShard.CrossTxTokenPrivacyData = crossCustomTokenPrivacyData
	crossShard.ToShardID = shardID
	return crossShard, nil
}

func (block *ShardBlock) AddValidationField(validationData string) error {
	block.ValidationData = validationData
	return nil
}

func (block ShardBlock) GetCurrentEpoch() uint64 {
	return block.Header.Epoch
}

func (block ShardBlock) GetProducer() string {
	return block.Header.Producer
}

func (block ShardBlock) GetProducerPubKeyStr() string {
	return block.Header.ProducerPubKeyStr
}

func (block ShardBlock) GetValidationField() string {
	return block.ValidationData
}

func (block ShardBlock) GetVersion() int {
	return block.Header.Version
}

func (block ShardBlock) GetHeight() uint64 {
	return block.Header.Height
}

func (block ShardBlock) GetRound() int {
	return block.Header.Round
}

func (block ShardBlock) GetRoundKey() string {
	return fmt.Sprint(block.Header.Height, "_", block.Header.Round)
}

func (block ShardBlock) GetInstructions() [][]string {
	return block.Body.Instructions
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

func (block CrossShardBlock) GetShardID() int {
	return int(block.Header.ShardID)
}

func (block CrossShardBlock) GetValidationField() string {
	return block.ValidationData
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

func (block ShardToBeaconBlock) GetValidationField() string {
	return block.ValidationData
}

func (block ShardToBeaconBlock) GetVersion() int {
	return block.Header.Version
}

func (block ShardToBeaconBlock) GetHeight() uint64 {
	return block.Header.Height
}

func (block ShardToBeaconBlock) GetRound() int {
	return block.Header.Round
}

func (block ShardToBeaconBlock) GetRoundKey() string {
	return fmt.Sprint(block.Header.Height, "_", block.Header.Round)
}
func (block ShardToBeaconBlock) GetInstructions() [][]string {
	return block.Instructions
}

func (block ShardToBeaconBlock) GetProposer() string {
	return block.Header.Proposer
}

func (block ShardToBeaconBlock) GetProposeTime() int64 {
	return block.Header.ProposeTime
}

func (block ShardToBeaconBlock) GetProduceTime() int64 {
	return block.Header.Timestamp
}
func (block ShardToBeaconBlock) GetProducer() string {
	return block.Header.Producer
}

func (block ShardBlock) GetConsensusType() string {
	return block.Header.ConsensusType
}

func (block CrossShardBlock) GetConsensusType() string {
	return block.Header.ConsensusType
}

func (block ShardToBeaconBlock) GetConsensusType() string {
	return block.Header.ConsensusType
}
