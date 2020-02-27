package common

type BlockPoolInterface interface {
	GetPrevHash() string
	GetHash() string
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
