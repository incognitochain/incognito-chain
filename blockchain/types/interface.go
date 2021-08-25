package types

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type BlockPoolInterface interface {
	GetPrevHash() common.Hash
	Hash() *common.Hash
	GetHeight() uint64
	GetShardID() int
	GetRound() int
}

type ChainInterface interface {
	GetShardID() int
	GetBestView() View
}

type View interface {
	GetHash() *common.Hash
	GetPreviousHash() *common.Hash
	GetHeight() uint64
	GetCommittee() []incognitokey.CommitteePublicKey
	GetPreviousBlockCommittee(db incdb.Database) ([]incognitokey.CommitteePublicKey, error)
	CommitteeStateVersion() int
	GetBlock() BlockInterface
	GetBeaconHeight() uint64
	GetProposerByTimeSlot(ts int64, version int) (incognitokey.CommitteePublicKey, int)
}

type BlockInterface interface {
	GetVersion() int
	GetHeight() uint64
	Hash() *common.Hash
	// AddValidationField(validateData string) error
	GetProducer() string
	GetValidationField() string
	GetRound() int
	GetRoundKey() string
	GetInstructions() [][]string
	GetConsensusType() string
	GetCurrentEpoch() uint64
	GetProduceTime() int64
	GetProposeTime() int64
	GetPrevHash() common.Hash
	GetProposer() string
	Type() string
	CommitteeFromBlock() common.Hash
	BodyHash() common.Hash
	AddValidationField(validationData string)
}
