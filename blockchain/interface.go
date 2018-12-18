package blockchain

type BFTBlock interface {
	Verify()
	GetType()
}
