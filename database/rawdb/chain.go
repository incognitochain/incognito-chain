package rawdb

import "github.com/ninjadotorg/cash-prototype/common"


func ReadHeadBlockHash(db DatabaseReader) common.Hash {
	data, _ := db.Get(headBlockKey)
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.Hash{}.BytesToHash(data)
}
