package types

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

type ShardBlock struct {
	ValidationData string `json:"ValidationData"`
	Body           ShardBody
	Header         ShardHeader
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

func (shardBlock *ShardBlock) CommitteeFromBlock() common.Hash {
	return shardBlock.Header.CommitteeFromBlock
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

func (shardBlock *ShardBlock) BuildShardBlockBody(instructions [][]string, crossTransaction map[byte][]CrossTransaction, transactions []metadata.Transaction) {
	shardBlock.Body.Instructions = append(shardBlock.Body.Instructions, instructions...)
	shardBlock.Body.CrossTransactions = crossTransaction
	shardBlock.Body.Transactions = append(shardBlock.Body.Transactions, transactions...)
}

func (shardBlock ShardBlock) Hash() *common.Hash {
	hash := shardBlock.Header.Hash()
	return &hash
}
func (shardBlock *ShardBlock) validateSanityData() (bool, error) {
	//Check Header
	if shardBlock.Header.Height == 1 && len(shardBlock.ValidationData) != 0 {
		return false, errors.New("Expect Shard Block with Height 1 to not have validationData")
	}
	// producer address must have 66 bytes: 33-byte public key, 33-byte transmission key
	if shardBlock.Header.Height > 1 && len(shardBlock.ValidationData) == 0 {
		return false, errors.New("Expect Shard Block to have validationData")
	}
	if int(shardBlock.Header.ShardID) < 0 || int(shardBlock.Header.ShardID) > 256 {
		return false, fmt.Errorf("Expect Shard Block ShardID in range 0 - 255 but get %+v ", shardBlock.Header.ShardID)
	}
	if shardBlock.Header.Version < BFT_VERSION {
		return false, fmt.Errorf("Expect Shard Block Version greater or equal than %+v but get %+v ", BFT_VERSION, shardBlock.Header.Version)
	}
	if len(shardBlock.Header.PreviousBlockHash[:]) != common.HashSize {
		return false, fmt.Errorf("Expect Shard Block Previous Hash in the right format")
	}
	if shardBlock.Header.Height < 1 {
		return false, fmt.Errorf("Expect Shard Block Height to be greater than 0")
	}
	if shardBlock.Header.Height == 1 && !shardBlock.Header.PreviousBlockHash.IsEqual(&common.Hash{}) {
		return false, fmt.Errorf("Expect Shard Block with Height 1 (first block) have Zero Hash Value")
	}
	if shardBlock.Header.Height > 1 && shardBlock.Header.PreviousBlockHash.IsEqual(&common.Hash{}) {
		return false, fmt.Errorf("Expect Shard Block with Height greater than 1 have Non-Zero Hash Value")
	}
	if shardBlock.Header.Round < 1 {
		return false, fmt.Errorf("Expect Shard Block Round greater or equal than 1")
	}
	if shardBlock.Header.Epoch < 1 {
		return false, fmt.Errorf("Expect Shard Block Epoch greater or equal than 1")
	}
	if shardBlock.Header.Timestamp <= 0 {
		return false, fmt.Errorf("Expect Shard Block Time greater than 0")
	}
	if len(shardBlock.Header.TxRoot[:]) != common.HashSize {
		return false, fmt.Errorf("Expect Shard Block Tx Root in the right format")
	}
	if len(shardBlock.Header.ShardTxRoot[:]) != common.HashSize {
		return false, fmt.Errorf("Expect Shard Block Shard Tx Root in the right format")
	}
	if len(shardBlock.Header.CrossTransactionRoot[:]) != common.HashSize {
		return false, fmt.Errorf("Expect Shard Block Cross Transaction Root in the right format")
	}
	if len(shardBlock.Header.InstructionsRoot[:]) != common.HashSize {
		return false, fmt.Errorf("Expect Shard Block Instructions Root in the right format")
	}
	if len(shardBlock.Header.CommitteeRoot[:]) != common.HashSize {
		return false, fmt.Errorf("Expect Shard Block Committee Root in the right format")
	}
	if shardBlock.Header.Height == 1 && !shardBlock.Header.CommitteeRoot.IsEqual(&common.Hash{}) {
		return false, fmt.Errorf("Expect Shard Block with Height 1 have Zero Hash Value")
	}
	if shardBlock.Header.Height > 1 && shardBlock.Header.CommitteeRoot.IsEqual(&common.Hash{}) {
		return false, fmt.Errorf("Expect Shard Block with Height greater than 1 have Non-Zero Hash Value")
	}
	if len(shardBlock.Header.PendingValidatorRoot[:]) != common.HashSize {
		return false, fmt.Errorf("Expect Shard Block Committee Root in the right format")
	}
	if len(shardBlock.Header.StakingTxRoot[:]) != common.HashSize {
		return false, fmt.Errorf("Expect Shard Block Staking Tx Root in the right format")
	}
	if len(shardBlock.Header.CrossShardBitMap) > 254 {
		return false, fmt.Errorf("Expect Shard Block Cross Shard Length Less Than 255")
	}
	if shardBlock.Header.BeaconHeight < 1 {
		return false, fmt.Errorf("Expect Shard Block has Beacon Height greater or equal than 1")
	}
	//if shardBlock.Header.BeaconHeight == 1 && !shardBlock.Header.BeaconHash.IsPointEqual(&common.Hash{}) {
	//	return false fmt.Errorf("Expect Shard Block with Beacon Height 1 have Zero Hash Value")
	//}
	if shardBlock.Header.BeaconHeight > 1 && shardBlock.Header.BeaconHash.IsEqual(&common.Hash{}) {
		return false, fmt.Errorf("Expect Shard Block with Beacon Height greater or equal than 1 have Non-Zero Hash Value")
	}
	if shardBlock.Header.TotalTxsFee == nil {
		return false, fmt.Errorf("Expect Shard Block Total Txs Fee have nil value")
	}
	if len(shardBlock.Header.InstructionMerkleRoot[:]) != common.HashSize {
		return false, fmt.Errorf("Expect Shard Block Instruction Merkle Root in the right format")
	}
	// body
	if shardBlock.Body.Instructions == nil {
		return false, fmt.Errorf("Expect Shard Block Instruction is not nil")
	}
	if len(shardBlock.Body.Instructions) != 0 && shardBlock.Header.InstructionMerkleRoot.IsEqual(&common.Hash{}) {
		return false, fmt.Errorf("Expect Shard Block Instruction Merkle Root have Non-Zero Hash Value because Instrucstion List is not empty")
	}
	if shardBlock.Body.CrossTransactions == nil {
		return false, fmt.Errorf("Expect Shard Block Cross Transactions Map is not nil")
	}
	if len(shardBlock.Body.CrossTransactions) != 0 && shardBlock.Header.CrossTransactionRoot.IsEqual(&common.Hash{}) {
		return false, fmt.Errorf("Expect Shard Block Cross Transaction Root have Non-Zero Hash Value because Cross Transaction List is not empty")
	}
	if shardBlock.Body.Transactions == nil {
		return false, fmt.Errorf("Expect Shard Block Transactions is not nil")
	}
	if len(shardBlock.Body.Transactions) != 0 && shardBlock.Header.TxRoot.IsEqual(&common.Hash{}) {
		return false, fmt.Errorf("Expect Shard Block Tx Root have Non-Zero Hash Value because Transactions List is not empty")
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
		return err
	}
	shardBlock.ValidationData = tempShardBlock.ValidationData

	blkBody := ShardBody{}
	err = blkBody.UnmarshalJSON(*tempShardBlock.Body)
	if err != nil {
		return err
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
		return err
	}
	shardBlock.Body = blkBody
	return nil
}

func (shardBlock *ShardBlock) AddTransaction(tx metadata.Transaction) error {
	if shardBlock.Body.Transactions == nil {
		return errors.New("not init tx arrays")
	}
	shardBlock.Body.Transactions = append(shardBlock.Body.Transactions, tx)
	return nil
}

func (block *ShardBlock) AddValidationField(validationData string) {
	block.ValidationData = validationData
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

func (block ShardBlock) Type() string {
	return common.ShardChainKey
}

func (block ShardBlock) BodyHash() common.Hash {
	return block.Body.Hash()
}
