package blockchain

import "github.com/internet-cash/prototype/database"

type BlockChain struct {
	ChainParams *Params
	DB          database.DB

	Blocks []Block
}
