package types

import (
	"encoding/json"
	"fmt"
	"sort"

	ggproto "github.com/golang/protobuf/proto"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/proto"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/pkg/errors"
)

type ShardBlock struct {
	ValidationData string `json:"ValidationData"`
	Body           ShardBody
	Header         ShardHeader
}

type ShardHeader struct {
	Producer              string                 `json:"Producer"`
	ProducerPubKeyStr     string                 `json:"ProducerPubKeyStr"`
	ShardID               byte                   `json:"ShardID"`               // shard ID which block belong to
	Version               int                    `json:"Version"`               // version of block structure
	PreviousBlockHash     common.Hash            `json:"PreviousBlockHash"`     // previous block hash or Parent block hash
	Height                uint64                 `json:"Height"`                // block height
	Round                 int                    `json:"Round"`                 // bpft consensus round
	Epoch                 uint64                 `json:"Epoch"`                 // epoch of block (according to current beacon height)
	CrossShardBitMap      []byte                 `json:"CrossShardBitMap"`      // crossShards bitmap for beacon
	BeaconHeight          uint64                 `json:"BeaconHeight"`          // beacon check point height
	BeaconHash            common.Hash            `json:"BeaconHash"`            // beacon check point hash
	TotalTxsFee           map[common.Hash]uint64 `json:"TotalTxsFee"`           // fee of all txs in block
	ConsensusType         string                 `json:"ConsensusType"`         // consensus type, by which this block is produced
	Timestamp             int64                  `json:"Timestamp"`             // timestamp of block
	TxRoot                common.Hash            `json:"TxRoot"`                // Transaction root created from transaction in shard
	ShardTxRoot           common.Hash            `json:"ShardTxRoot"`           // output root created for other shard
	CrossTransactionRoot  common.Hash            `json:"CrossTransactionRoot"`  // transaction root created from transaction of micro shard to shard block (from other shard)
	InstructionsRoot      common.Hash            `json:"InstructionsRoot"`      // actions root created from Instructions and Metadata of transaction
	CommitteeRoot         common.Hash            `json:"CommitteeRoot"`         // hash from public key list of all committees designated to create this block
	PendingValidatorRoot  common.Hash            `json:"PendingValidatorRoot"`  // hash from public key list of all pending validators designated to this ShardID
	StakingTxRoot         common.Hash            `json:"StakingTxRoot"`         // hash from staking transaction map in shard best state
	InstructionMerkleRoot common.Hash            `json:"InstructionMerkleRoot"` // Merkle root of all instructions (using Keccak256 hash func) to relay to Ethreum
	// This obsoletes InstructionMerkleRoot but for simplicity, we keep it for now

	//for version >= 2
	Proposer    string
	ProposeTime int64

	//for version >= 3
	CommitteeFromBlock common.Hash `json:"CommitteeFromBlock"` // Block Hash Of Swapped Committees Block

	//for version 6
	FinalityHeight uint64 `json:"FinalityHeight"`
}

type ShardBody struct {
	Instructions      [][]string
	CrossTransactions map[byte][]CrossTransaction //CrossOutputCoin from all other shard
	Transactions      []metadata.Transaction
}

func NewCrossShardBlock() *CrossShardBlock {
	return &CrossShardBlock{}
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

func (shardBlock *ShardBlock) GetAggregateRootHash() common.Hash {

	res := []byte{}
	res = append(res, byte(shardBlock.Header.Version))
	res = append(res, shardBlock.Header.BeaconHash.Bytes()...)
	res = append(res, shardBlock.Header.TxRoot.Bytes()...)
	res = append(res, shardBlock.Header.ShardTxRoot.Bytes()...)
	res = append(res, shardBlock.Header.CrossTransactionRoot.Bytes()...)
	res = append(res, shardBlock.Header.InstructionsRoot.Bytes()...)
	res = append(res, shardBlock.Header.CommitteeRoot.Bytes()...)
	res = append(res, shardBlock.Header.PendingValidatorRoot.Bytes()...)
	res = append(res, shardBlock.Header.StakingTxRoot.Bytes()...)
	res = append(res, shardBlock.Header.InstructionMerkleRoot.Bytes()...)
	res = append(res, shardBlock.Header.CommitteeFromBlock.Bytes()...)

	return common.HashH(res)
}
func (shardBlock ShardBlock) GetFinalityHeight() uint64 {
	return shardBlock.Header.FinalityHeight
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
	if shardBlock.Header.Version < 0 {
		return false, fmt.Errorf("Expect Shard Block Version greater or equal than %+v but get %+v ", SHARD_BLOCK_VERSION, shardBlock.Header.Version)
	}
	if len(shardBlock.Header.PreviousBlockHash[:]) != common.HashSize {
		return false, fmt.Errorf("Expect Shard Block Previous Hash in the right format")
	}
	if shardBlock.Header.Height < 0 {
		return false, fmt.Errorf("Expect Shard Block Height to be greater than 0")
	}
	if shardBlock.Header.Height == 1 && !shardBlock.Header.PreviousBlockHash.IsEqual(&common.Hash{}) {
		return false, fmt.Errorf("Expect Shard Block with Height 1 (first block) have Zero Hash Value")
	}
	if shardBlock.Header.Height > 1 && shardBlock.Header.PreviousBlockHash.IsEqual(&common.Hash{}) {
		return false, fmt.Errorf("Expect Shard Block with Height greater than 1 have Non-Zero Hash Value")
	}
	if shardBlock.Header.Round < 0 {
		return false, fmt.Errorf("Expect Shard Block Round greater or equal than 0")
	}
	if shardBlock.Header.Epoch < 0 {
		return false, fmt.Errorf("Expect Shard Block Epoch greater or equal than 0")
	}
	if shardBlock.Header.Timestamp < 0 {
		return false, fmt.Errorf("Expect Shard Block Time greater or equal than 0")
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
	if shardBlock.Header.BeaconHeight < 0 {
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
	if (shardBlock.Header.Height != 1) && (len(shardBlock.Body.Transactions) != 0) && (shardBlock.Header.TxRoot.IsEqual(&common.Hash{})) {
		return false, fmt.Errorf("Expect Shard Block Tx Root have Non-Zero Hash Value because Transactions List is not empty")
	}
	return true, nil
}

func (sHeader ShardHeader) ToProtoShardHeader() (*proto.ShardHeaderBytes, error) {
	res := &proto.ShardHeaderBytes{}
	var err error
	producerIdx := -1
	proposerIdx := -1
	if sHeader.Producer != "" {
		producerIdx, err = CommitteeProvider.GetValidatorIndex(
			sHeader.Producer,
			sHeader.ShardID,
			sHeader.CommitteeFromBlock,
			sHeader.PreviousBlockHash,
			sHeader.Height,
			sHeader.Epoch,
		)
		if err != nil {
			return nil, err
		}
		proposerIdx = producerIdx
		if (sHeader.Proposer != sHeader.Producer) && (sHeader.Proposer != "") {
			proposerIdx, err = CommitteeProvider.GetValidatorIndex(
				sHeader.Proposer,
				sHeader.ShardID,
				sHeader.CommitteeFromBlock,
				sHeader.PreviousBlockHash,
				sHeader.Height,
				sHeader.Epoch,
			)
			if err != nil {
				return nil, err
			}
		}
	}
	res.Producer = int32(producerIdx)
	res.Proposer = int32(proposerIdx)
	res.ShardID = int32(sHeader.ShardID)
	res.Version = int32(sHeader.Version)
	res.PreviousBlockHash = sHeader.PreviousBlockHash[:]
	res.Height = sHeader.Height
	res.Round = int32(sHeader.Round)
	res.Epoch = sHeader.Epoch
	res.CrossShardBitMap = make([]byte, len(sHeader.CrossShardBitMap))
	copy(res.CrossShardBitMap, sHeader.CrossShardBitMap)
	res.BeaconHeight = sHeader.BeaconHeight
	res.BeaconHash = sHeader.BeaconHash[:]
	res.TotalTxsFee = map[string]uint64{}
	for k, v := range sHeader.TotalTxsFee {
		res.TotalTxsFee[k.String()] = v
	}
	res.ConsensusType = sHeader.ConsensusType
	res.Timestamp = sHeader.Timestamp
	res.TxRoot = sHeader.TxRoot[:]
	res.ShardTxRoot = sHeader.ShardTxRoot[:]
	res.CrossTransactionRoot = sHeader.CrossTransactionRoot[:]
	res.InstructionsRoot = sHeader.InstructionsRoot[:]
	res.CommitteeRoot = sHeader.CommitteeRoot[:]
	res.PendingValidatorRoot = sHeader.PendingValidatorRoot[:]
	res.StakingTxRoot = sHeader.StakingTxRoot[:]
	res.InstructionMerkleRoot = sHeader.InstructionMerkleRoot[:]
	res.ProposeTime = sHeader.ProposeTime
	res.CommitteeFromBlock = sHeader.CommitteeFromBlock[:]
	res.FinalityHeight = sHeader.FinalityHeight
	return res, nil
}

func (sHeader *ShardHeader) FromProtoShardHeader(protoData *proto.ShardHeaderBytes) error {
	copy(sHeader.CommitteeFromBlock[:], protoData.CommitteeFromBlock)
	copy(sHeader.PreviousBlockHash[:], protoData.PreviousBlockHash)
	var err error
	producerPk := ""
	proposerPk := ""
	if protoData.Producer != -1 {
		producerPk, err = CommitteeProvider.GetValidatorFromIndex(
			int(protoData.Producer),
			byte(protoData.ShardID),
			sHeader.CommitteeFromBlock,
			sHeader.PreviousBlockHash,
			protoData.Height,
			protoData.Epoch,
		)
		if err != nil {
			return nil
		}
		proposerPk = producerPk
		if protoData.Producer != protoData.Proposer {
			proposerPk, err = CommitteeProvider.GetValidatorFromIndex(
				int(protoData.Proposer),
				byte(protoData.ShardID),
				sHeader.CommitteeFromBlock,
				sHeader.PreviousBlockHash,
				protoData.Height,
				protoData.Epoch,
			)
			if err != nil {
				return nil
			}
		}
	}
	sHeader.Producer = producerPk
	sHeader.ProducerPubKeyStr = producerPk
	sHeader.Proposer = proposerPk
	sHeader.ShardID = byte(protoData.ShardID)
	sHeader.Version = int(protoData.Version)
	copy(sHeader.PreviousBlockHash[:], protoData.PreviousBlockHash)
	sHeader.Height = protoData.Height
	sHeader.Round = int(protoData.Round)
	sHeader.Epoch = protoData.Epoch
	sHeader.CrossShardBitMap = make([]byte, len(protoData.CrossShardBitMap))
	copy(sHeader.CrossShardBitMap, protoData.CrossShardBitMap)
	sHeader.BeaconHeight = protoData.BeaconHeight
	copy(sHeader.BeaconHash[:], protoData.BeaconHash)
	sHeader.TotalTxsFee = make(map[common.Hash]uint64)
	for k, v := range protoData.TotalTxsFee {
		h := common.Hash{}.NewHashFromStr2(k)
		sHeader.TotalTxsFee[h] = v
	}
	sHeader.ConsensusType = protoData.ConsensusType
	sHeader.Timestamp = protoData.Timestamp
	copy(sHeader.TxRoot[:], protoData.TxRoot)
	copy(sHeader.ShardTxRoot[:], protoData.ShardTxRoot)
	copy(sHeader.CrossTransactionRoot[:], protoData.CrossTransactionRoot)
	copy(sHeader.InstructionsRoot[:], protoData.InstructionsRoot)
	copy(sHeader.CommitteeRoot[:], protoData.CommitteeRoot)
	copy(sHeader.PendingValidatorRoot[:], protoData.PendingValidatorRoot)
	copy(sHeader.StakingTxRoot[:], protoData.StakingTxRoot)
	copy(sHeader.InstructionMerkleRoot[:], protoData.InstructionMerkleRoot)
	sHeader.ProposeTime = protoData.ProposeTime
	sHeader.FinalityHeight = protoData.FinalityHeight
	return nil
}

func (sBody ShardBody) ToProtoShardBody() *proto.ShardBodyBytes {
	res := &proto.ShardBodyBytes{}
	res.CrossTransactions = map[int32]*proto.CrossTransactionTmp{}
	for k, v := range sBody.CrossTransactions {
		crossTxs := &proto.CrossTransactionTmp{}
		for _, tx := range v {
			txBytes, err := json.Marshal(tx)
			if err != nil {
				panic(err)
			}
			crossTxs.Data = append(crossTxs.Data, txBytes)
		}
		res.CrossTransactions[int32(k)] = crossTxs
	}
	for _, v := range sBody.Instructions {
		insTmp := &proto.InstrucstionTmp{
			Data: v,
		}
		res.Instrucstions = append(res.Instrucstions, insTmp)
	}

	for _, tx := range sBody.Transactions {
		txBytes, err := json.Marshal(tx)
		if err != nil {
			panic(err)
		}
		res.Transactions = append(res.Transactions, txBytes)
	}
	return res
}

func (sBody *ShardBody) FromProtoShardBody(protoData *proto.ShardBodyBytes) error {
	sBody.CrossTransactions = map[byte][]CrossTransaction{}
	for k, v := range protoData.CrossTransactions {
		crossTxs := []CrossTransaction{}
		for _, txBytes := range v.Data {
			crossTx := &CrossTransaction{}
			err := json.Unmarshal(txBytes, crossTx)
			if err != nil {
				panic(err)
			}
			crossTxs = append(crossTxs, *crossTx)
		}
		sBody.CrossTransactions[byte(k)] = crossTxs
	}

	for _, ins := range protoData.Instrucstions {
		sBody.Instructions = append(sBody.Instructions, ins.Data)
	}

	for _, txBytes := range protoData.Transactions {
		var tx metadata.Transaction
		var parseErr error
		var txChoice *transaction.TxChoice
		txChoice, parseErr = transaction.DeserializeTransactionJSON(txBytes)
		if parseErr != nil {
			return fmt.Errorf("unmarshall Json Shard Block Is Failed. Error %v", parseErr)
		}
		tx = txChoice.ToTx()
		if tx == nil {
			return fmt.Errorf("unmarshall Json Shard Block Is Failed. Corrupted TX")
		}
		sBody.Transactions = append(sBody.Transactions, tx)
	}
	return nil
}

func (sBlock ShardBlock) ToProtoShardBlock() (*proto.ShardBlockBytes, error) {
	res := &proto.ShardBlockBytes{}
	var err error
	res.Body = sBlock.Body.ToProtoShardBody()
	res.Header, err = sBlock.Header.ToProtoShardHeader()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (sBlock *ShardBlock) FromProtoShardBlock(protoData *proto.ShardBlockBytes) error {
	if protoData.Header.Height > 1 {
		sBlock.AddValidationField("")
	}
	err := sBlock.Body.FromProtoShardBody(protoData.Body)
	if err != nil {
		return err
	}
	return sBlock.Header.FromProtoShardHeader(protoData.Header)
}

func (sBlock *ShardBlock) GetBodyBytes() ([]byte, error) {
	protoBody := sBlock.Body.ToProtoShardBody()
	if protoBody == nil {
		return nil, errors.Errorf("Can not convert shard block %v - %v to protobuf", sBlock.Header.Height, sBlock.Hash().String())
	}
	return ggproto.Marshal(protoBody)
}

func (sBlock *ShardBlock) SetBodyFromBytes(rawBytes []byte) error {
	protoBody := &proto.ShardBodyBytes{}
	err := ggproto.Unmarshal(rawBytes, protoBody)
	if err != nil {
		return err
	}
	return sBlock.Body.FromProtoShardBody(protoBody)
}

func (sBlock *ShardBlock) RemoveBody() {
	sBlock.Body.Transactions = []metadata.Transaction{}
	sBlock.Body.Instructions = [][]string{}
	sBlock.Body.CrossTransactions = map[byte][]CrossTransaction{}
}

func (shardBlock *ShardBlock) FromBytes(data []byte) error {
	protoBlk := &proto.ShardBlockBytes{}
	err := ggproto.Unmarshal(data, protoBlk)
	if err != nil {
		return err
	}
	return shardBlock.FromProtoShardBlock(protoBlk)
}

func (shardBlock *ShardBlock) ToBytes() ([]byte, error) {
	protoBlk, err := shardBlock.ToProtoShardBlock()
	if (protoBlk == nil) || (err != nil) {
		return nil, errors.Errorf("Can not convert shardBlock %v - %v to protobuf, err %v", shardBlock.Header.Height, shardBlock.Hash().String(), err)
	}
	protoBytes, err := ggproto.Marshal(protoBlk)
	if err != nil {
		return nil, err
	}
	return protoBytes, nil
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
	shardBlock.Body = blkBody
	if shardBlock.Body.Transactions == nil {
		shardBlock.Body.Transactions = []metadata.Transaction{}
	}
	for _, tx := range shardBlock.Body.Transactions {
		valEnv := updateTxEnvWithBlock(shardBlock, tx)
		tx.SetValidationEnv(valEnv)
		if tx.GetType() == common.TxCustomTokenPrivacyType {
			txCustom, ok := tx.(transaction.TransactionToken)
			if !ok {
				return errors.Errorf("Can not parse this tx %v to tx custom token privacy", tx.Hash().String())
			}
			valEnvCustom := updateTxEnvWithBlock(shardBlock, txCustom.GetTxNormal())
			txCustom.GetTxNormal().SetValidationEnv(valEnvCustom)
		}
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
	return nil
}

func (shardBlock *ShardBlock) AddTransaction(tx metadata.Transaction) error {
	if shardBlock.Body.Transactions == nil {
		return errors.New("not init tx arrays")
	}
	shardBlock.Body.Transactions = append(shardBlock.Body.Transactions, tx)
	return nil
}

func (shardHeader *ShardHeader) String() string {
	res := utils.EmptyString
	// res += shardHeader.ProducerAddress.String()
	res += string(shardHeader.ShardID)
	res += fmt.Sprintf("%v", shardHeader.Version)
	res += shardHeader.PreviousBlockHash.String()
	res += fmt.Sprintf("%v", shardHeader.Height)
	res += fmt.Sprintf("%v", shardHeader.Round)
	res += fmt.Sprintf("%v", shardHeader.Epoch)
	res += fmt.Sprintf("%v", shardHeader.Timestamp)
	res += shardHeader.TxRoot.String()
	res += shardHeader.ShardTxRoot.String()
	res += shardHeader.CrossTransactionRoot.String()
	res += shardHeader.InstructionsRoot.String()
	res += shardHeader.CommitteeRoot.String()
	res += shardHeader.PendingValidatorRoot.String()
	res += shardHeader.BeaconHash.String()
	res += shardHeader.StakingTxRoot.String()
	res += fmt.Sprintf("%v", shardHeader.BeaconHeight)
	tokenIDs := make([]common.Hash, 0)
	for tokenID, _ := range shardHeader.TotalTxsFee {
		tokenIDs = append(tokenIDs, tokenID)
	}
	sort.Slice(tokenIDs, func(i int, j int) bool {
		res, _ := tokenIDs[i].Cmp(&tokenIDs[j])
		return res == -1
	})

	for _, tokenID := range tokenIDs {
		res += fmt.Sprintf("%v~%v", tokenID.String(), shardHeader.TotalTxsFee[tokenID])
	}
	for _, value := range shardHeader.CrossShardBitMap {
		res += string(value)
	}

	if shardHeader.Version >= MULTI_VIEW_VERSION {
		res += shardHeader.Proposer
		res += fmt.Sprintf("%v", shardHeader.ProposeTime)
	}

	if shardHeader.Version >= SHARD_SFV2_VERSION {
		res += shardHeader.CommitteeFromBlock.String()
	}

	if shardHeader.Version >= LEMMA2_VERSION {
		res += fmt.Sprintf("%v", shardHeader.FinalityHeight)
	}

	return res
}

func (shardHeader *ShardHeader) MetaHash() common.Hash {
	return common.Keccak256([]byte(shardHeader.String()))
}

func (shardHeader *ShardHeader) Hash() common.Hash {
	// Block header of bridge uses Keccak256 as a hash func to check on Ethereum when relaying blocks
	blkMetaHash := shardHeader.MetaHash()
	blkInstHash := shardHeader.InstructionMerkleRoot
	combined := append(blkMetaHash[:], blkInstHash[:]...)
	return common.Keccak256(combined)
}

/*
Customize UnmarshalJSON to parse list TxNormal
because we have many types of block, so we can need to customize data from marshal from json string to build a block
*/
func (shardBody *ShardBody) UnmarshalJSON(data []byte) error {
	type Alias ShardBody
	temp := &struct {
		Transactions []json.RawMessage
		*Alias
	}{
		Alias: (*Alias)(shardBody),
	}

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return fmt.Errorf("unmarshall Json Shard Block Is Failed. Error %v", err)
	}

	// process tx from tx interface of temp
	for _, txTemp := range temp.Transactions {
		// txTempJson, _ := json.MarshalIndent(txTemp, "", "\t")
		var tx metadata.Transaction
		var parseErr error
		var txChoice *transaction.TxChoice
		txChoice, parseErr = transaction.DeserializeTransactionJSON(txTemp)
		if parseErr != nil {
			return fmt.Errorf("unmarshall Json Shard Block Is Failed. Error %v", parseErr)
		}
		tx = txChoice.ToTx()
		if tx == nil {
			return fmt.Errorf("unmarshall Json Shard Block Is Failed. Corrupted TX")
		}

		shardBody.Transactions = append(shardBody.Transactions, tx)
	}
	return nil
}

func (shardBody ShardBody) Hash() common.Hash {
	res := []byte{}

	for _, item := range shardBody.Instructions {
		for _, l := range item {
			res = append(res, []byte(l)...)
		}
	}
	keys := []int{}
	for k := range shardBody.CrossTransactions {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, value := range shardBody.CrossTransactions[byte(shardID)] {
			res = append(res, []byte(fmt.Sprintf("%v", value.BlockHeight))...)
			res = append(res, value.BlockHash.GetBytes()...)
			for _, coins := range value.OutputCoin {
				res = append(res, coins.Bytes()...)
			}
			for _, coins := range value.TokenPrivacyData {
				res = append(res, coins.Bytes()...)
			}
		}
	}
	for _, tx := range shardBody.Transactions {
		res = append(res, tx.Hash().GetBytes()...)
	}
	return common.HashH(res)
}

func (shardBody ShardBody) ExtractIncomingCrossShardMap() (map[byte][]common.Hash, error) {
	crossShardMap := make(map[byte][]common.Hash)
	for shardID, crossblocks := range shardBody.CrossTransactions {
		for _, crossblock := range crossblocks {
			crossShardMap[shardID] = append(crossShardMap[shardID], crossblock.BlockHash)
		}
	}
	return crossShardMap, nil
}

func (shardBody ShardBody) ExtractOutgoingCrossShardMap() (map[byte][]common.Hash, error) {
	crossShardMap := make(map[byte][]common.Hash)
	return crossShardMap, nil
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
