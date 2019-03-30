package lvdb

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func ViewDBByPrefix(db database.DatabaseInterface, prefix []byte)  map[string]string{
	begin := prefix
	// +1 to search in that range
	end := common.BytesPlusOne(prefix)
	
	searchRange := util.Range{
		Start: begin,
		Limit: end,
	}
	iter := db.NewIterator(&searchRange, nil)
	res := make(map[string]string)
	for iter.Next() {
		res[string(iter.Key())] = string(iter.Value())
	}
	return res
}
