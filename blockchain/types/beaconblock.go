package types

import (
	"encoding/json"
	"fmt"
	"strconv"

	ggproto "github.com/golang/protobuf/proto"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/proto"
	"github.com/pkg/errors"
)

type BlockConsensusData struct {
	BlockHash      common.Hash
	BlockHeight    uint64
	FinalityHeight uint64
	Proposer       string
	ProposerTime   int64
	ValidationData string
}

type BeaconBlock struct {
	ValidationData string `json:"ValidationData"`
	Body           BeaconBody
	Header         BeaconHeader
}

type BeaconBody struct {
	// Shard State extract from shard to beacon block
	// Store all shard state == store content of all shard to beacon block
	ShardState   map[byte][]ShardState
	Instructions [][]string
}

func NewBeaconBody(shardState map[byte][]ShardState, instructions [][]string) BeaconBody {
	return BeaconBody{ShardState: shardState, Instructions: instructions}
}

func (b *BeaconBody) SetInstructions(inst [][]string) {
	b.Instructions = inst
}

type BeaconHeader struct {
	Version           int         `json:"Version"`
	Height            uint64      `json:"Height"`
	Epoch             uint64      `json:"Epoch"`
	Round             int         `json:"Round"`
	Timestamp         int64       `json:"Timestamp"`
	PreviousBlockHash common.Hash `json:"PreviousBlockHash"`
	InstructionHash   common.Hash `json:"InstructionHash"` // hash of all parameters == hash of instruction
	ShardStateHash    common.Hash `json:"ShardStateHash"`  // each shard will have a list of blockHash, shardRoot is hash of all list
	// Merkle root of all instructions (using Keccak256 hash func) to relay to Ethreum
	// This obsoletes InstructionHash but for simplicity, we keep it for now
	InstructionMerkleRoot           common.Hash `json:"InstructionMerkleRoot"`
	BeaconCommitteeAndValidatorRoot common.Hash `json:"BeaconCommitteeAndValidatorRoot"` //Build from two list: BeaconCommittee + BeaconPendingValidator
	BeaconCandidateRoot             common.Hash `json:"BeaconCandidateRoot"`             // CandidateBeaconWaitingForCurrentRandom + CandidateBeaconWaitingForNextRandom
	ShardCandidateRoot              common.Hash `json:"ShardCandidateRoot"`              // CandidateShardWaitingForCurrentRandom + CandidateShardWaitingForNextRandom
	ShardCommitteeAndValidatorRoot  common.Hash `json:"ShardCommitteeAndValidatorRoot"`
	AutoStakingRoot                 common.Hash `json:"AutoStakingRoot"`
	ShardSyncValidatorRoot          common.Hash `json:"ShardSyncValidatorRoot"`
	ConsensusType                   string      `json:"ConsensusType"`
	Producer                        string      `json:"Producer"`
	ProducerPubKeyStr               string      `json:"ProducerPubKeyStr"`

	//for version 2
	Proposer    string `json:"Proposer"`
	ProposeTime int64  `json:"ProposeTime"`

	//for version 6
	FinalityHeight uint64 `json:"FinalityHeight"`

	//for version 8, instant finality
	ProcessBridgeFromBlock *uint64 `json:"integer,omitempty"`
}

func NewBeaconHeader(version int, height uint64, epoch uint64, round int, timestamp int64, previousBlockHash common.Hash, consensusType string, producer string, producerPubKeyStr string) BeaconHeader {
	return BeaconHeader{Version: version, Height: height, Epoch: epoch, Round: round, Timestamp: timestamp, PreviousBlockHash: previousBlockHash, ConsensusType: consensusType, Producer: producer, ProducerPubKeyStr: producerPubKeyStr}
}

func (beaconHeader *BeaconHeader) AddBeaconHeaderHash(
	instructionHash common.Hash,
	shardStateHash common.Hash,
	instructionMerkleRoot []byte,
	beaconCommitteeAndValidatorRoot common.Hash,
	beaconCandidateRoot common.Hash,
	shardCandidateRoot common.Hash,
	shardCommitteeAndValidatorRoot common.Hash,
	autoStakingRoot common.Hash,
	shardSyncValidatorRoot common.Hash) {
	beaconHeader.InstructionHash = instructionHash
	beaconHeader.ShardStateHash = shardStateHash
	copy(beaconHeader.InstructionMerkleRoot[:], instructionMerkleRoot)
	beaconHeader.BeaconCommitteeAndValidatorRoot = beaconCommitteeAndValidatorRoot
	beaconHeader.BeaconCandidateRoot = beaconCandidateRoot
	beaconHeader.ShardCandidateRoot = shardCandidateRoot
	beaconHeader.ShardCommitteeAndValidatorRoot = shardCommitteeAndValidatorRoot
	beaconHeader.AutoStakingRoot = autoStakingRoot
	beaconHeader.ShardSyncValidatorRoot = shardSyncValidatorRoot

}

type ShardState struct {
	ValidationData         string
	PreviousValidationData string
	CommitteeFromBlock     common.Hash
	Height                 uint64
	Hash                   common.Hash
	CrossShard             []byte //In this state, shard i send cross shard tx to which shard
	ProposerTime           int64
	Version                int
}

func NewShardState(validationData string,
	prevValidationData string,
	committeeFromBlock common.Hash,
	height uint64,
	hash common.Hash,
	crossShard []byte,
	proposerTime int64,
	version int,
) ShardState {
	newCrossShard := make([]byte, len(crossShard))
	copy(newCrossShard, crossShard)
	return ShardState{
		ValidationData:         validationData,
		PreviousValidationData: prevValidationData,
		CommitteeFromBlock:     committeeFromBlock,
		Height:                 height,
		Hash:                   hash,
		CrossShard:             newCrossShard,
		ProposerTime:           proposerTime,
		Version:                version,
	}
}

func (sState ShardState) ToProtoShardState() (*proto.ShardStateBytes, error) {
	res := &proto.ShardStateBytes{
		CommitteeFromBlock: make([]byte, common.HashSize),
		Hash:               make([]byte, common.HashSize),
		CrossShard:         make([]byte, len(sState.CrossShard)),
	}
	res.CommitteeFromBlock = sState.CommitteeFromBlock[:]
	copy(res.CrossShard, sState.CrossShard)
	res.Hash = sState.Hash[:]
	res.Height = sState.Height
	if len(sState.ValidationData) != 0 {
		vData, err := consensustypes.DecodeValidationData(sState.ValidationData)
		if err != nil {
			return nil, err
		}
		res.ValidationData, err = vData.ToBytes()
		if err != nil {
			return nil, err
		}
	} else {
		res.ValidationData = []byte{}
	}
	res.Version = int32(sState.Version)
	res.ProposerTime = sState.ProposerTime
	return res, nil
}

func (sState *ShardState) FromProtoShardState(protoData *proto.ShardStateBytes) error {
	sState.CommitteeFromBlock = common.Hash{}
	sState.Hash = common.Hash{}
	sState.CrossShard = make([]byte, len(protoData.CrossShard))
	copy(sState.CommitteeFromBlock[:], protoData.CommitteeFromBlock)
	copy(sState.Hash[:], protoData.Hash)
	copy(sState.CrossShard, protoData.CrossShard)
	sState.Height = protoData.Height
	if len(protoData.ValidationData) != 0 {
		vData := &consensustypes.ValidationData{}
		err := vData.FromBytes(protoData.ValidationData)
		if err != nil {
			return err
		}
		sState.ValidationData, err = consensustypes.EncodeValidationData(*vData)
		if err != nil {
			return err
		}
	} else {
		sState.ValidationData = ""
	}
	sState.Version = int(protoData.Version)
	sState.ProposerTime = protoData.ProposerTime
	return nil
}

func (beaconBlock *BeaconBlock) GetVersion() int {
	return beaconBlock.Header.Version
}

func (beaconBlock *BeaconBlock) GetPrevHash() common.Hash {
	return beaconBlock.Header.PreviousBlockHash
}

func NewBeaconBlock() *BeaconBlock {
	return &BeaconBlock{}
}

func (beaconBlock BeaconBlock) ToProtoBeaconBlock() (*proto.BeaconBlockBytes, error) {
	res := &proto.BeaconBlockBytes{}
	var err error
	res.Body, err = beaconBlock.Body.ToProtoBeaconBody()
	if err != nil {
		return nil, err
	}
	res.Header, err = beaconBlock.Header.ToProtoBeaconHeader()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (beaconBlock *BeaconBlock) FromProtoBeaconBlock(protoData *proto.BeaconBlockBytes) error {

	if protoData.Header.Height > 1 {
		beaconBlock.AddValidationField("")
	}
	err := beaconBlock.Body.FromProtoBeaconBody(protoData.Body)
	if err != nil {
		return err
	}
	return beaconBlock.Header.FromProtoBeaconHeader(protoData.Header)
}

func (beaconBlock *BeaconBlock) ToBytes() ([]byte, error) {
	protoBlk, err := beaconBlock.ToProtoBeaconBlock()
	if (protoBlk == nil) || (err != nil) {
		return nil, errors.Errorf("Can not convert beacon block %v - %v to protobuf", beaconBlock.Header.Height, beaconBlock.Hash().String())
	}
	protoBytes, err := ggproto.Marshal(protoBlk)
	if err != nil {
		return nil, err
	}
	return protoBytes, nil

}

func (beaconBlock *BeaconBlock) FromBytes(rawBytes []byte) error {
	protoBlk := &proto.BeaconBlockBytes{}
	err := ggproto.Unmarshal(rawBytes, protoBlk)
	if err != nil {
		return err
	}
	return beaconBlock.FromProtoBeaconBlock(protoBlk)

}

func (beaconBlock *BeaconBlock) GetProposer() string {
	return beaconBlock.Header.Proposer
}

func (beaconBlock *BeaconBlock) GetProposeTime() int64 {
	return beaconBlock.Header.ProposeTime
}

func (beaconBlock *BeaconBlock) GetProduceTime() int64 {
	return beaconBlock.Header.Timestamp
}

func (beaconBlock BeaconBlock) Hash() *common.Hash {
	hash := beaconBlock.Header.Hash()
	return &hash
}

//propose hash of beacon contain consensus info
func (beaconBlock BeaconBlock) ProposeHash() *common.Hash {
	hash := beaconBlock.Header.ProposeHash()
	return &hash
}

func (beaconHeader *BeaconHeader) ProposeHash() common.Hash {

	if beaconHeader.Version < INSTANT_FINALITY_VERSION {
		return beaconHeader.Hash()
	}
	res := beaconHeader.toString()
	res += beaconHeader.Proposer
	res += fmt.Sprintf("%v", beaconHeader.ProposeTime)
	res += fmt.Sprintf("%v", beaconHeader.FinalityHeight)
	blkInstHash := beaconHeader.InstructionMerkleRoot
	blkMetaHash := common.Keccak256([]byte(res))
	combined := append(blkMetaHash[:], blkInstHash[:]...)

	return common.Keccak256(combined)
}

func (beaconBlock BeaconBlock) FullHashString() string {
	return beaconBlock.ProposeHash().String() + "~" + beaconBlock.Hash().String()
}

func (beaconBlock BeaconBlock) GetCurrentEpoch() uint64 {
	return beaconBlock.Header.Epoch
}

func (beaconBlock BeaconBlock) GetHeight() uint64 {
	return beaconBlock.Header.Height
}

func (beaconBlock BeaconBlock) GetBeaconHeight() uint64 {
	return beaconBlock.Header.Height
}

func (beaconBlock BeaconBlock) GetShardID() int {
	return -1
}

func (beaconBlock BeaconBlock) CommitteeFromBlock() common.Hash {
	return common.Hash{}
}

func (beaconBlock BeaconBlock) GetFinalityHeight() uint64 {
	return beaconBlock.Header.FinalityHeight
}

func (beaconBlock *BeaconBlock) UnmarshalJSON(data []byte) error {
	tempBeaconBlock := &struct {
		ValidationData string `json:"ValidationData"`
		Header         BeaconHeader
		Body           BeaconBody
	}{}
	err := json.Unmarshal(data, &tempBeaconBlock)
	if err != nil {
		return err
	}
	beaconBlock.ValidationData = tempBeaconBlock.ValidationData
	beaconBlock.Header = tempBeaconBlock.Header
	beaconBlock.Body = tempBeaconBlock.Body
	return nil
}

func (beaconBlock *BeaconBlock) GetAggregateRootHash() common.Hash {
	res := []byte{}
	res = append(res, byte(beaconBlock.Header.Version))
	res = append(res, beaconBlock.Header.InstructionHash.Bytes()...)
	res = append(res, beaconBlock.Header.ShardStateHash.Bytes()...)
	res = append(res, beaconBlock.Header.InstructionMerkleRoot.Bytes()...)
	res = append(res, beaconBlock.Header.BeaconCommitteeAndValidatorRoot.Bytes()...)
	res = append(res, beaconBlock.Header.BeaconCandidateRoot.Bytes()...)
	res = append(res, beaconBlock.Header.ShardCandidateRoot.Bytes()...)
	res = append(res, beaconBlock.Header.ShardCommitteeAndValidatorRoot.Bytes()...)
	res = append(res, beaconBlock.Header.AutoStakingRoot.Bytes()...)
	res = append(res, beaconBlock.Header.ShardSyncValidatorRoot.Bytes()...)

	return common.HashH(res)
}

func (beaconBlock *BeaconBlock) AddValidationField(validationData string) {
	beaconBlock.ValidationData = validationData
	return
}
func (beaconBlock BeaconBlock) GetValidationField() string {
	return beaconBlock.ValidationData
}

func (beaconBlock BeaconBlock) GetRound() int {
	return beaconBlock.Header.Round
}
func (beaconBlock BeaconBlock) GetRoundKey() string {
	return fmt.Sprint(beaconBlock.Header.Height, "_", beaconBlock.Header.Round)
}

func (beaconBlock BeaconBlock) GetInstructions() [][]string {
	return beaconBlock.Body.Instructions
}

func (beaconBlock BeaconBlock) GetProducer() string {
	return beaconBlock.Header.Producer
}

func (beaconBlock BeaconBlock) GetProducerPubKeyStr() string {
	return beaconBlock.Header.ProducerPubKeyStr
}

func (beaconBlock BeaconBlock) GetConsensusType() string {
	return beaconBlock.Header.ConsensusType
}

func (beaconBlock BeaconBlock) Type() string {
	return common.BeaconChainKey
}

func (beaconBlock BeaconBlock) BodyHash() common.Hash {
	return beaconBlock.Body.Hash()
}

func (beaconBody *BeaconBody) toString() string {
	res := ""
	for _, l := range beaconBody.ShardState {
		for _, r := range l {
			res += strconv.Itoa(int(r.Height))
			res += r.Hash.String()
			crossShard, _ := json.Marshal(r.CrossShard)
			res += string(crossShard)

		}
	}
	for _, l := range beaconBody.Instructions {
		for _, r := range l {
			res += r
		}
	}
	return res
}

func (beaconBody BeaconBody) Hash() common.Hash {
	return common.HashH([]byte(beaconBody.toString()))
}

func (beaconBody BeaconBody) ToProtoBeaconBody() (*proto.BeaconBodyBytes, error) {
	res := &proto.BeaconBodyBytes{}
	for _, ins := range beaconBody.Instructions {
		insTmp := &proto.InstrucstionTmp{}
		insTmp.Data = ins
		res.Instrucstions = append(res.Instrucstions, insTmp)
	}
	res.ShardState = map[int32]*proto.ListShardStateBytes{}
	for sID, sStates := range beaconBody.ShardState {
		lProtoSState := &proto.ListShardStateBytes{}
		for _, sState := range sStates {
			protoSState, err := sState.ToProtoShardState()
			if err != nil {
				panic(err)
				return nil, err
			}
			lProtoSState.Data = append(lProtoSState.Data, protoSState)
		}
		res.ShardState[int32(sID)] = lProtoSState
	}
	return res, nil
}

func (beaconBody *BeaconBody) FromProtoBeaconBody(protoData *proto.BeaconBodyBytes) error {
	for _, ins := range protoData.Instrucstions {
		beaconBody.Instructions = append(beaconBody.Instructions, ins.Data)
	}
	beaconBody.ShardState = map[byte][]ShardState{}
	for pSID, pSStates := range protoData.ShardState {
		lSState := []ShardState{}
		for _, pSState := range pSStates.Data {
			sState := &ShardState{}
			err := sState.FromProtoShardState(pSState)
			if err != nil {
				return err
			}
			lSState = append(lSState, *sState)
		}
		beaconBody.ShardState[byte(pSID)] = lSState
	}
	return nil
}

func (header *BeaconHeader) toString() string {
	res := ""
	// res += beaconHeader.ProducerAddress.String()
	res += fmt.Sprintf("%v", header.Version)
	res += fmt.Sprintf("%v", header.Height)
	res += fmt.Sprintf("%v", header.Epoch)
	res += fmt.Sprintf("%v", header.Round)
	res += fmt.Sprintf("%v", header.Timestamp)
	res += header.PreviousBlockHash.String()
	res += header.BeaconCommitteeAndValidatorRoot.String()
	res += header.BeaconCandidateRoot.String()
	res += header.ShardCandidateRoot.String()
	res += header.ShardCommitteeAndValidatorRoot.String()
	res += header.AutoStakingRoot.String()
	res += header.ShardStateHash.String()
	res += header.InstructionHash.String()

	if header.Version >= INSTANT_FINALITY_VERSION {
		//instant finality will move consensus info out of block hash
		if header.ProcessBridgeFromBlock == nil {
			res += "0"
		} else {
			res += fmt.Sprintf("%v", *header.ProcessBridgeFromBlock)
		}

	} else {
		//to compatible with mainnet database, version 3 dont have proposer info
		if header.Version == MULTI_VIEW_VERSION || header.Version >= 4 {
			res += header.Proposer
			res += fmt.Sprintf("%v", header.ProposeTime)
		}

		if header.Version >= LEMMA2_VERSION {
			res += fmt.Sprintf("%v", header.FinalityHeight)
		}
	}

	return res
}

func (header *BeaconHeader) MetaHash() common.Hash {
	return common.Keccak256([]byte(header.toString()))
}

func (header *BeaconHeader) Hash() common.Hash {
	// Block header of beacon uses Keccak256 as a hash func to check on Ethereum when relaying blocks
	blkMetaHash := header.MetaHash()
	blkInstHash := header.InstructionMerkleRoot
	combined := append(blkMetaHash[:], blkInstHash[:]...)
	return common.Keccak256(combined)
}

func (header BeaconHeader) ToProtoBeaconHeader() (*proto.BeaconHeaderBytes, error) {
	var err error
	res := &proto.BeaconHeaderBytes{}
	res.Version = int32(header.Version)
	res.Height = header.Height
	res.Epoch = header.Epoch
	res.Round = int32(header.Round)
	res.Timestamp = header.Timestamp
	res.PreviousBlockHash = header.PreviousBlockHash[:]
	res.InstructionHash = header.InstructionHash[:]
	res.ShardStateHash = header.ShardStateHash[:]
	res.InstructionMerkleRoot = header.InstructionMerkleRoot[:]
	res.BeaconCommitteeAndValidatorRoot = header.BeaconCommitteeAndValidatorRoot[:]
	res.BeaconCandidateRoot = header.BeaconCandidateRoot[:]
	res.ShardCandidateRoot = header.ShardCandidateRoot[:]
	res.ShardCommitteeAndValidatorRoot = header.ShardCommitteeAndValidatorRoot[:]
	res.AutoStakingRoot = header.AutoStakingRoot[:]
	res.ShardSyncValidatorRoot = header.ShardSyncValidatorRoot[:]
	res.ConsensusType = header.ConsensusType
	producerIdx := -1
	proposerIdx := -1
	if header.Producer != "" {
		producerIdx, err = CommitteeProvider.GetValidatorIndex(
			header.Producer,
			common.BeaconChainSyncID,
			common.Hash{},
			header.PreviousBlockHash,
			header.Epoch,
		)
		if err != nil {
			panic(err)
			return nil, err
		}
		proposerIdx = producerIdx
		if (header.Proposer != header.Producer) && (header.Proposer != "") {
			proposerIdx, err = CommitteeProvider.GetValidatorIndex(
				header.Proposer,
				common.BeaconChainSyncID,
				common.Hash{},
				header.PreviousBlockHash,
				header.Epoch,
			)
			if err != nil {
				panic(err)
				return nil, err
			}
		}
	}
	res.Producer = int32(producerIdx)
	res.Proposer = int32(proposerIdx)
	res.ProposeTime = header.ProposeTime
	res.FinalityHeight = header.FinalityHeight
	return res, nil
}

func (header *BeaconHeader) FromProtoBeaconHeader(protoData *proto.BeaconHeaderBytes) error {
	header.Version = int(protoData.Version)
	header.Height = protoData.Height
	header.Epoch = protoData.Epoch
	header.Round = int(protoData.Round)
	header.Timestamp = protoData.Timestamp
	copy(header.PreviousBlockHash[:], protoData.PreviousBlockHash)
	copy(header.InstructionHash[:], protoData.InstructionHash)
	copy(header.ShardStateHash[:], protoData.ShardStateHash)
	copy(header.InstructionMerkleRoot[:], protoData.InstructionMerkleRoot)
	copy(header.BeaconCommitteeAndValidatorRoot[:], protoData.BeaconCommitteeAndValidatorRoot)
	copy(header.BeaconCandidateRoot[:], protoData.BeaconCandidateRoot)
	copy(header.ShardCandidateRoot[:], protoData.ShardCandidateRoot)
	copy(header.ShardCommitteeAndValidatorRoot[:], protoData.ShardCommitteeAndValidatorRoot)
	copy(header.AutoStakingRoot[:], protoData.AutoStakingRoot)
	copy(header.ShardSyncValidatorRoot[:], protoData.ShardSyncValidatorRoot)
	header.ConsensusType = protoData.ConsensusType
	var err error
	producerPk := ""
	proposerPk := ""
	if protoData.Producer != -1 {
		producerPk, err = CommitteeProvider.GetValidatorFromIndex(
			int(protoData.Producer),
			common.BeaconChainSyncID,
			common.Hash{},
			header.PreviousBlockHash,
			protoData.Epoch,
		)
		if err != nil {
			return nil
		}
		proposerPk = producerPk
		if protoData.Producer != protoData.Proposer {
			proposerPk, err = CommitteeProvider.GetValidatorFromIndex(
				int(protoData.Proposer),
				common.BeaconChainSyncID,
				common.Hash{},
				header.PreviousBlockHash,
				protoData.Epoch,
			)
			if err != nil {
				return nil
			}
		}
	}
	header.Producer = producerPk
	header.ProducerPubKeyStr = producerPk
	header.Proposer = proposerPk
	header.ProposeTime = protoData.ProposeTime
	header.FinalityHeight = protoData.FinalityHeight
	return nil
}
