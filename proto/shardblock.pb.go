// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.10.0
// source: proto/shardblock.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ShardHeaderBytes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Producer              int32             `protobuf:"varint,1,opt,name=Producer,proto3" json:"Producer,omitempty"`
	ShardID               int32             `protobuf:"varint,2,opt,name=ShardID,proto3" json:"ShardID,omitempty"`
	Version               int32             `protobuf:"varint,3,opt,name=Version,proto3" json:"Version,omitempty"`
	PreviousBlockHash     []byte            `protobuf:"bytes,4,opt,name=PreviousBlockHash,proto3" json:"PreviousBlockHash,omitempty"`
	Height                uint64            `protobuf:"varint,5,opt,name=Height,proto3" json:"Height,omitempty"`
	Round                 int32             `protobuf:"varint,6,opt,name=Round,proto3" json:"Round,omitempty"`
	Epoch                 uint64            `protobuf:"varint,7,opt,name=Epoch,proto3" json:"Epoch,omitempty"`
	CrossShardBitMap      []byte            `protobuf:"bytes,8,opt,name=CrossShardBitMap,proto3" json:"CrossShardBitMap,omitempty"`
	BeaconHeight          uint64            `protobuf:"varint,9,opt,name=BeaconHeight,proto3" json:"BeaconHeight,omitempty"`
	BeaconHash            []byte            `protobuf:"bytes,10,opt,name=BeaconHash,proto3" json:"BeaconHash,omitempty"`
	TotalTxsFee           map[string]uint64 `protobuf:"bytes,11,rep,name=TotalTxsFee,proto3" json:"TotalTxsFee,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	ConsensusType         string            `protobuf:"bytes,12,opt,name=ConsensusType,proto3" json:"ConsensusType,omitempty"`
	Timestamp             int64             `protobuf:"zigzag64,13,opt,name=Timestamp,proto3" json:"Timestamp,omitempty"`
	TxRoot                []byte            `protobuf:"bytes,14,opt,name=TxRoot,proto3" json:"TxRoot,omitempty"`
	ShardTxRoot           []byte            `protobuf:"bytes,15,opt,name=ShardTxRoot,proto3" json:"ShardTxRoot,omitempty"`
	CrossTransactionRoot  []byte            `protobuf:"bytes,16,opt,name=CrossTransactionRoot,proto3" json:"CrossTransactionRoot,omitempty"`
	InstructionsRoot      []byte            `protobuf:"bytes,17,opt,name=InstructionsRoot,proto3" json:"InstructionsRoot,omitempty"`
	CommitteeRoot         []byte            `protobuf:"bytes,18,opt,name=CommitteeRoot,proto3" json:"CommitteeRoot,omitempty"`
	PendingValidatorRoot  []byte            `protobuf:"bytes,19,opt,name=PendingValidatorRoot,proto3" json:"PendingValidatorRoot,omitempty"`
	StakingTxRoot         []byte            `protobuf:"bytes,20,opt,name=StakingTxRoot,proto3" json:"StakingTxRoot,omitempty"`
	InstructionMerkleRoot []byte            `protobuf:"bytes,21,opt,name=InstructionMerkleRoot,proto3" json:"InstructionMerkleRoot,omitempty"`
	Proposer              int32             `protobuf:"varint,22,opt,name=Proposer,proto3" json:"Proposer,omitempty"`
	ProposeTime           int64             `protobuf:"zigzag64,23,opt,name=ProposeTime,proto3" json:"ProposeTime,omitempty"`
	CommitteeFromBlock    []byte            `protobuf:"bytes,24,opt,name=CommitteeFromBlock,proto3" json:"CommitteeFromBlock,omitempty"`
	FinalityHeight        uint64            `protobuf:"varint,25,opt,name=FinalityHeight,proto3" json:"FinalityHeight,omitempty"`
}

func (x *ShardHeaderBytes) Reset() {
	*x = ShardHeaderBytes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_shardblock_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShardHeaderBytes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShardHeaderBytes) ProtoMessage() {}

func (x *ShardHeaderBytes) ProtoReflect() protoreflect.Message {
	mi := &file_proto_shardblock_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShardHeaderBytes.ProtoReflect.Descriptor instead.
func (*ShardHeaderBytes) Descriptor() ([]byte, []int) {
	return file_proto_shardblock_proto_rawDescGZIP(), []int{0}
}

func (x *ShardHeaderBytes) GetProducer() int32 {
	if x != nil {
		return x.Producer
	}
	return 0
}

func (x *ShardHeaderBytes) GetShardID() int32 {
	if x != nil {
		return x.ShardID
	}
	return 0
}

func (x *ShardHeaderBytes) GetVersion() int32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *ShardHeaderBytes) GetPreviousBlockHash() []byte {
	if x != nil {
		return x.PreviousBlockHash
	}
	return nil
}

func (x *ShardHeaderBytes) GetHeight() uint64 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *ShardHeaderBytes) GetRound() int32 {
	if x != nil {
		return x.Round
	}
	return 0
}

func (x *ShardHeaderBytes) GetEpoch() uint64 {
	if x != nil {
		return x.Epoch
	}
	return 0
}

func (x *ShardHeaderBytes) GetCrossShardBitMap() []byte {
	if x != nil {
		return x.CrossShardBitMap
	}
	return nil
}

func (x *ShardHeaderBytes) GetBeaconHeight() uint64 {
	if x != nil {
		return x.BeaconHeight
	}
	return 0
}

func (x *ShardHeaderBytes) GetBeaconHash() []byte {
	if x != nil {
		return x.BeaconHash
	}
	return nil
}

func (x *ShardHeaderBytes) GetTotalTxsFee() map[string]uint64 {
	if x != nil {
		return x.TotalTxsFee
	}
	return nil
}

func (x *ShardHeaderBytes) GetConsensusType() string {
	if x != nil {
		return x.ConsensusType
	}
	return ""
}

func (x *ShardHeaderBytes) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *ShardHeaderBytes) GetTxRoot() []byte {
	if x != nil {
		return x.TxRoot
	}
	return nil
}

func (x *ShardHeaderBytes) GetShardTxRoot() []byte {
	if x != nil {
		return x.ShardTxRoot
	}
	return nil
}

func (x *ShardHeaderBytes) GetCrossTransactionRoot() []byte {
	if x != nil {
		return x.CrossTransactionRoot
	}
	return nil
}

func (x *ShardHeaderBytes) GetInstructionsRoot() []byte {
	if x != nil {
		return x.InstructionsRoot
	}
	return nil
}

func (x *ShardHeaderBytes) GetCommitteeRoot() []byte {
	if x != nil {
		return x.CommitteeRoot
	}
	return nil
}

func (x *ShardHeaderBytes) GetPendingValidatorRoot() []byte {
	if x != nil {
		return x.PendingValidatorRoot
	}
	return nil
}

func (x *ShardHeaderBytes) GetStakingTxRoot() []byte {
	if x != nil {
		return x.StakingTxRoot
	}
	return nil
}

func (x *ShardHeaderBytes) GetInstructionMerkleRoot() []byte {
	if x != nil {
		return x.InstructionMerkleRoot
	}
	return nil
}

func (x *ShardHeaderBytes) GetProposer() int32 {
	if x != nil {
		return x.Proposer
	}
	return 0
}

func (x *ShardHeaderBytes) GetProposeTime() int64 {
	if x != nil {
		return x.ProposeTime
	}
	return 0
}

func (x *ShardHeaderBytes) GetCommitteeFromBlock() []byte {
	if x != nil {
		return x.CommitteeFromBlock
	}
	return nil
}

func (x *ShardHeaderBytes) GetFinalityHeight() uint64 {
	if x != nil {
		return x.FinalityHeight
	}
	return 0
}

type InstrucstionTmp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []string `protobuf:"bytes,1,rep,name=Data,proto3" json:"Data,omitempty"`
}

func (x *InstrucstionTmp) Reset() {
	*x = InstrucstionTmp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_shardblock_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InstrucstionTmp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InstrucstionTmp) ProtoMessage() {}

func (x *InstrucstionTmp) ProtoReflect() protoreflect.Message {
	mi := &file_proto_shardblock_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InstrucstionTmp.ProtoReflect.Descriptor instead.
func (*InstrucstionTmp) Descriptor() ([]byte, []int) {
	return file_proto_shardblock_proto_rawDescGZIP(), []int{1}
}

func (x *InstrucstionTmp) GetData() []string {
	if x != nil {
		return x.Data
	}
	return nil
}

type CrossTransactionTmp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data [][]byte `protobuf:"bytes,1,rep,name=Data,proto3" json:"Data,omitempty"`
}

func (x *CrossTransactionTmp) Reset() {
	*x = CrossTransactionTmp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_shardblock_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CrossTransactionTmp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CrossTransactionTmp) ProtoMessage() {}

func (x *CrossTransactionTmp) ProtoReflect() protoreflect.Message {
	mi := &file_proto_shardblock_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CrossTransactionTmp.ProtoReflect.Descriptor instead.
func (*CrossTransactionTmp) Descriptor() ([]byte, []int) {
	return file_proto_shardblock_proto_rawDescGZIP(), []int{2}
}

func (x *CrossTransactionTmp) GetData() [][]byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type ShardBodyBytes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Instrucstions     []*InstrucstionTmp             `protobuf:"bytes,1,rep,name=Instrucstions,proto3" json:"Instrucstions,omitempty"`
	CrossTransactions map[int32]*CrossTransactionTmp `protobuf:"bytes,2,rep,name=CrossTransactions,proto3" json:"CrossTransactions,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Transactions      [][]byte                       `protobuf:"bytes,3,rep,name=Transactions,proto3" json:"Transactions,omitempty"`
}

func (x *ShardBodyBytes) Reset() {
	*x = ShardBodyBytes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_shardblock_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShardBodyBytes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShardBodyBytes) ProtoMessage() {}

func (x *ShardBodyBytes) ProtoReflect() protoreflect.Message {
	mi := &file_proto_shardblock_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShardBodyBytes.ProtoReflect.Descriptor instead.
func (*ShardBodyBytes) Descriptor() ([]byte, []int) {
	return file_proto_shardblock_proto_rawDescGZIP(), []int{3}
}

func (x *ShardBodyBytes) GetInstrucstions() []*InstrucstionTmp {
	if x != nil {
		return x.Instrucstions
	}
	return nil
}

func (x *ShardBodyBytes) GetCrossTransactions() map[int32]*CrossTransactionTmp {
	if x != nil {
		return x.CrossTransactions
	}
	return nil
}

func (x *ShardBodyBytes) GetTransactions() [][]byte {
	if x != nil {
		return x.Transactions
	}
	return nil
}

type ShardBlockBytes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ValidationData []byte            `protobuf:"bytes,1,opt,name=ValidationData,proto3" json:"ValidationData,omitempty"`
	Body           *ShardBodyBytes   `protobuf:"bytes,2,opt,name=Body,proto3" json:"Body,omitempty"`
	Header         *ShardHeaderBytes `protobuf:"bytes,3,opt,name=Header,proto3" json:"Header,omitempty"`
}

func (x *ShardBlockBytes) Reset() {
	*x = ShardBlockBytes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_shardblock_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShardBlockBytes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShardBlockBytes) ProtoMessage() {}

func (x *ShardBlockBytes) ProtoReflect() protoreflect.Message {
	mi := &file_proto_shardblock_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShardBlockBytes.ProtoReflect.Descriptor instead.
func (*ShardBlockBytes) Descriptor() ([]byte, []int) {
	return file_proto_shardblock_proto_rawDescGZIP(), []int{4}
}

func (x *ShardBlockBytes) GetValidationData() []byte {
	if x != nil {
		return x.ValidationData
	}
	return nil
}

func (x *ShardBlockBytes) GetBody() *ShardBodyBytes {
	if x != nil {
		return x.Body
	}
	return nil
}

func (x *ShardBlockBytes) GetHeader() *ShardHeaderBytes {
	if x != nil {
		return x.Header
	}
	return nil
}

var File_proto_shardblock_proto protoreflect.FileDescriptor

var file_proto_shardblock_proto_rawDesc = []byte{
	0x0a, 0x16, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x68, 0x61, 0x72, 0x64, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf4, 0x07, 0x0a, 0x10, 0x53, 0x68, 0x61,
	0x72, 0x64, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x1a, 0x0a,
	0x08, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x08, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x53, 0x68, 0x61,
	0x72, 0x64, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x53, 0x68, 0x61, 0x72,
	0x64, 0x49, 0x44, 0x12, 0x18, 0x0a, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x2c, 0x0a,
	0x11, 0x50, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61,
	0x73, 0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x11, 0x50, 0x72, 0x65, 0x76, 0x69, 0x6f,
	0x75, 0x73, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68, 0x12, 0x16, 0x0a, 0x06, 0x48,
	0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x48, 0x65, 0x69,
	0x67, 0x68, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x05, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x45, 0x70, 0x6f,
	0x63, 0x68, 0x18, 0x07, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x12,
	0x2a, 0x0a, 0x10, 0x43, 0x72, 0x6f, 0x73, 0x73, 0x53, 0x68, 0x61, 0x72, 0x64, 0x42, 0x69, 0x74,
	0x4d, 0x61, 0x70, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x10, 0x43, 0x72, 0x6f, 0x73, 0x73,
	0x53, 0x68, 0x61, 0x72, 0x64, 0x42, 0x69, 0x74, 0x4d, 0x61, 0x70, 0x12, 0x22, 0x0a, 0x0c, 0x42,
	0x65, 0x61, 0x63, 0x6f, 0x6e, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x0c, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12,
	0x1e, 0x0a, 0x0a, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x48, 0x61, 0x73, 0x68, 0x18, 0x0a, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x0a, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x48, 0x61, 0x73, 0x68, 0x12,
	0x44, 0x0a, 0x0b, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x54, 0x78, 0x73, 0x46, 0x65, 0x65, 0x18, 0x0b,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x53, 0x68, 0x61, 0x72, 0x64, 0x48, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x42, 0x79, 0x74, 0x65, 0x73, 0x2e, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x54, 0x78, 0x73,
	0x46, 0x65, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0b, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x54,
	0x78, 0x73, 0x46, 0x65, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x43, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73,
	0x75, 0x73, 0x54, 0x79, 0x70, 0x65, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x43, 0x6f,
	0x6e, 0x73, 0x65, 0x6e, 0x73, 0x75, 0x73, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x12, 0x52, 0x09,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x16, 0x0a, 0x06, 0x54, 0x78, 0x52,
	0x6f, 0x6f, 0x74, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x54, 0x78, 0x52, 0x6f, 0x6f,
	0x74, 0x12, 0x20, 0x0a, 0x0b, 0x53, 0x68, 0x61, 0x72, 0x64, 0x54, 0x78, 0x52, 0x6f, 0x6f, 0x74,
	0x18, 0x0f, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x53, 0x68, 0x61, 0x72, 0x64, 0x54, 0x78, 0x52,
	0x6f, 0x6f, 0x74, 0x12, 0x32, 0x0a, 0x14, 0x43, 0x72, 0x6f, 0x73, 0x73, 0x54, 0x72, 0x61, 0x6e,
	0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x6f, 0x6f, 0x74, 0x18, 0x10, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x14, 0x43, 0x72, 0x6f, 0x73, 0x73, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x2a, 0x0a, 0x10, 0x49, 0x6e, 0x73, 0x74, 0x72,
	0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x6f, 0x6f, 0x74, 0x18, 0x11, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x10, 0x49, 0x6e, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52,
	0x6f, 0x6f, 0x74, 0x12, 0x24, 0x0a, 0x0d, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65,
	0x52, 0x6f, 0x6f, 0x74, 0x18, 0x12, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0d, 0x43, 0x6f, 0x6d, 0x6d,
	0x69, 0x74, 0x74, 0x65, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x32, 0x0a, 0x14, 0x50, 0x65, 0x6e,
	0x64, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x52, 0x6f, 0x6f,
	0x74, 0x18, 0x13, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x14, 0x50, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67,
	0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x24, 0x0a,
	0x0d, 0x53, 0x74, 0x61, 0x6b, 0x69, 0x6e, 0x67, 0x54, 0x78, 0x52, 0x6f, 0x6f, 0x74, 0x18, 0x14,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x0d, 0x53, 0x74, 0x61, 0x6b, 0x69, 0x6e, 0x67, 0x54, 0x78, 0x52,
	0x6f, 0x6f, 0x74, 0x12, 0x34, 0x0a, 0x15, 0x49, 0x6e, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x4d, 0x65, 0x72, 0x6b, 0x6c, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x18, 0x15, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x15, 0x49, 0x6e, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4d,
	0x65, 0x72, 0x6b, 0x6c, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x50, 0x72, 0x6f,
	0x70, 0x6f, 0x73, 0x65, 0x72, 0x18, 0x16, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x50, 0x72, 0x6f,
	0x70, 0x6f, 0x73, 0x65, 0x72, 0x12, 0x20, 0x0a, 0x0b, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65,
	0x54, 0x69, 0x6d, 0x65, 0x18, 0x17, 0x20, 0x01, 0x28, 0x12, 0x52, 0x0b, 0x50, 0x72, 0x6f, 0x70,
	0x6f, 0x73, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x2e, 0x0a, 0x12, 0x43, 0x6f, 0x6d, 0x6d, 0x69,
	0x74, 0x74, 0x65, 0x65, 0x46, 0x72, 0x6f, 0x6d, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x18, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x12, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x46, 0x72,
	0x6f, 0x6d, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x26, 0x0a, 0x0e, 0x46, 0x69, 0x6e, 0x61, 0x6c,
	0x69, 0x74, 0x79, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x19, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x0e, 0x46, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x74, 0x79, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x1a,
	0x3e, 0x0a, 0x10, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x54, 0x78, 0x73, 0x46, 0x65, 0x65, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22,
	0x25, 0x0a, 0x0f, 0x49, 0x6e, 0x73, 0x74, 0x72, 0x75, 0x63, 0x73, 0x74, 0x69, 0x6f, 0x6e, 0x54,
	0x6d, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x44, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x04, 0x44, 0x61, 0x74, 0x61, 0x22, 0x29, 0x0a, 0x13, 0x43, 0x72, 0x6f, 0x73, 0x73, 0x54,
	0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x6d, 0x70, 0x12, 0x12, 0x0a,
	0x04, 0x44, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x04, 0x44, 0x61, 0x74,
	0x61, 0x22, 0x9e, 0x02, 0x0a, 0x0e, 0x53, 0x68, 0x61, 0x72, 0x64, 0x42, 0x6f, 0x64, 0x79, 0x42,
	0x79, 0x74, 0x65, 0x73, 0x12, 0x36, 0x0a, 0x0d, 0x49, 0x6e, 0x73, 0x74, 0x72, 0x75, 0x63, 0x73,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x49, 0x6e,
	0x73, 0x74, 0x72, 0x75, 0x63, 0x73, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x6d, 0x70, 0x52, 0x0d, 0x49,
	0x6e, 0x73, 0x74, 0x72, 0x75, 0x63, 0x73, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x54, 0x0a, 0x11,
	0x43, 0x72, 0x6f, 0x73, 0x73, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x53, 0x68, 0x61, 0x72, 0x64, 0x42,
	0x6f, 0x64, 0x79, 0x42, 0x79, 0x74, 0x65, 0x73, 0x2e, 0x43, 0x72, 0x6f, 0x73, 0x73, 0x54, 0x72,
	0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x11, 0x43, 0x72, 0x6f, 0x73, 0x73, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x12, 0x22, 0x0a, 0x0c, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x0c, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x1a, 0x5a, 0x0a, 0x16, 0x43, 0x72, 0x6f, 0x73, 0x73, 0x54,
	0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x2a, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x14, 0x2e, 0x43, 0x72, 0x6f, 0x73, 0x73, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x54, 0x6d, 0x70, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x22, 0x89, 0x01, 0x0a, 0x0f, 0x53, 0x68, 0x61, 0x72, 0x64, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x26, 0x0a, 0x0e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e,
	0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61, 0x12, 0x23,
	0x0a, 0x04, 0x42, 0x6f, 0x64, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x53,
	0x68, 0x61, 0x72, 0x64, 0x42, 0x6f, 0x64, 0x79, 0x42, 0x79, 0x74, 0x65, 0x73, 0x52, 0x04, 0x42,
	0x6f, 0x64, 0x79, 0x12, 0x29, 0x0a, 0x06, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x53, 0x68, 0x61, 0x72, 0x64, 0x48, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x42, 0x79, 0x74, 0x65, 0x73, 0x52, 0x06, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x42, 0x09,
	0x5a, 0x07, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_proto_shardblock_proto_rawDescOnce sync.Once
	file_proto_shardblock_proto_rawDescData = file_proto_shardblock_proto_rawDesc
)

func file_proto_shardblock_proto_rawDescGZIP() []byte {
	file_proto_shardblock_proto_rawDescOnce.Do(func() {
		file_proto_shardblock_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_shardblock_proto_rawDescData)
	})
	return file_proto_shardblock_proto_rawDescData
}

var file_proto_shardblock_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_proto_shardblock_proto_goTypes = []interface{}{
	(*ShardHeaderBytes)(nil),    // 0: ShardHeaderBytes
	(*InstrucstionTmp)(nil),     // 1: InstrucstionTmp
	(*CrossTransactionTmp)(nil), // 2: CrossTransactionTmp
	(*ShardBodyBytes)(nil),      // 3: ShardBodyBytes
	(*ShardBlockBytes)(nil),     // 4: ShardBlockBytes
	nil,                         // 5: ShardHeaderBytes.TotalTxsFeeEntry
	nil,                         // 6: ShardBodyBytes.CrossTransactionsEntry
}
var file_proto_shardblock_proto_depIdxs = []int32{
	5, // 0: ShardHeaderBytes.TotalTxsFee:type_name -> ShardHeaderBytes.TotalTxsFeeEntry
	1, // 1: ShardBodyBytes.Instrucstions:type_name -> InstrucstionTmp
	6, // 2: ShardBodyBytes.CrossTransactions:type_name -> ShardBodyBytes.CrossTransactionsEntry
	3, // 3: ShardBlockBytes.Body:type_name -> ShardBodyBytes
	0, // 4: ShardBlockBytes.Header:type_name -> ShardHeaderBytes
	2, // 5: ShardBodyBytes.CrossTransactionsEntry.value:type_name -> CrossTransactionTmp
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_proto_shardblock_proto_init() }
func file_proto_shardblock_proto_init() {
	if File_proto_shardblock_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_shardblock_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShardHeaderBytes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_shardblock_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InstrucstionTmp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_shardblock_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CrossTransactionTmp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_shardblock_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShardBodyBytes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_shardblock_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShardBlockBytes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_shardblock_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_shardblock_proto_goTypes,
		DependencyIndexes: file_proto_shardblock_proto_depIdxs,
		MessageInfos:      file_proto_shardblock_proto_msgTypes,
	}.Build()
	File_proto_shardblock_proto = out.File
	file_proto_shardblock_proto_rawDesc = nil
	file_proto_shardblock_proto_goTypes = nil
	file_proto_shardblock_proto_depIdxs = nil
}
