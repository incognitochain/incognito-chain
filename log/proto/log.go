package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

type FeatureID int32

const (
	FID_CONSENSUS_BEACON FeatureID = 0

	FID_CONSENSUS_SHARD FeatureID = 1
)

// Enum value maps for FeatureID.
var (
	FeatureID_name = map[int32]string{
		0: "FEATURE_CONSENSUS_BEACON",
		1: "FEATURE_CONSENSUS_SHARD",
	}
	FeatureID_value = map[string]int32{
		"FEATURE_CONSENSUS_BEACON": 0,
		"FEATURE_CONSENSUS_SHARD":  1,
	}
)

func (x FeatureID) Enum() *FeatureID {
	p := new(FeatureID)
	*p = x
	return p
}

func (x FeatureID) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (FeatureID) Descriptor() protoreflect.EnumDescriptor {
	return file_log_proto_log_proto_enumTypes[0].Descriptor()
}

func (FeatureID) Type() protoreflect.EnumType {
	return &file_log_proto_log_proto_enumTypes[0]
}

func (x FeatureID) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use FeatureID.Descriptor instead.
func (FeatureID) EnumDescriptor() ([]byte, []int) {
	return file_log_proto_log_proto_rawDescGZIP(), []int{0}
}
