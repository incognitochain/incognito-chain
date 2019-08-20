package common

type BlockInterface interface {
	GetHeight() uint64
	Hash() *Hash
	// AddValidationField(validateData string) error
	GetValidationField() string
	GetRound() int
	GetRoundKey() string
	GetInstructions() [][]string
}
