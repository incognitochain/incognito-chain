package blockchain

type BFTBlockInterface interface {
	Verify() error
}
