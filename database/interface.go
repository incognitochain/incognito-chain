package database

type DB interface {

	GetBlock(key []byte) []byte
	SaveBlock(key []byte, value []byte) (bool, error)
}
