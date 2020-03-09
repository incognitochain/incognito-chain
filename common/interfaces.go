package common

type BlockPoolInterface interface {
	GetPrevHash() string
	Hash() *Hash
	GetHeight() uint64
	GetShardID() int
}

type CrossShardBlkPoolInterface interface {
	Hash() *Hash
	GetHeight() uint64
}

type BlockInterface interface {
	GetHeight() uint64
	Hash() *Hash
	// AddValidationField(validateData string) error
	GetProducer() string
	GetValidationField() string
	GetRound() int
	GetRoundKey() string
	GetInstructions() [][]string
	GetConsensusType() string
	GetCurrentEpoch() uint64
}
