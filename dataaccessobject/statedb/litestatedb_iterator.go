package statedb

import (
	"bytes"
	"github.com/incognitochain/incognito-chain/incdb"
	"sort"
	"strings"
)

type LiteStateDBIterator struct {
	dbIterator   incdb.Iterator
	dbPrefix     []byte
	memIndex     int
	memKeySort   [][]byte //assume no duplicate key
	memValueSort [][]byte
	firstCall    bool
	firstDBCall  bool
	firstMemCall bool
	nextMemIndex int
}

func NewLiteStateDBIterator(db incdb.Database, dbPrefix, prefix []byte, kvMap map[string][]byte) *LiteStateDBIterator {

	dbIterator := db.NewIteratorWithPrefix(append(dbPrefix, prefix...))
	memKeySort := [][]byte{}
	memValueSort := [][]byte{}
	for k, _ := range kvMap {
		if strings.Index(k, string(prefix)) == 0 {
			memKeySort = append(memKeySort, []byte(k))
		}
	}
	sort.Slice(memKeySort, func(i, j int) bool {
		for index := range memKeySort[i] {
			if memKeySort[i][index] < memKeySort[j][index] {
				return true
			}
			if memKeySort[i][index] > memKeySort[j][index] {
				return false
			}
		}
		return false
	})
	for _, v := range memKeySort {
		memValueSort = append(memValueSort, kvMap[string(v)])
	}

	iter := &LiteStateDBIterator{
		dbIterator,
		dbPrefix,
		0,
		memKeySort,
		memValueSort,
		false,
		false,
		false,
		0,
	}
	iter.dbIterator.Next()
	return iter

}

func (l *LiteStateDBIterator) whichZoneToSelect() int {
	if l.memIndex < len(l.memKeySort) {
		memKey := l.memKeySort[l.memIndex]
		dbkey := l.dbIterator.Key()
		if len(dbkey) == 0 {
			return 1
		}
		dbkey = dbkey[len(l.dbPrefix):]
		if bytes.Compare(dbkey, memKey) < 0 {
			return 0
		}

		if bytes.Compare(dbkey, memKey) > 0 {
			return 1
		}

		//duplicate key, which mean there is update of the key in the mem, select mem and bypass db
		if bytes.Compare(memKey, dbkey) == 0 {
			l.dbIterator.Next()
			return 1
		}
	} else {
		return 0
	}
	return 0
}

func (l *LiteStateDBIterator) Next() bool {
	selectedZone := l.whichZoneToSelect()
	if selectedZone == 0 {
		if !l.firstCall { //the first call is init call in Next() iterator, db already init, so need to bypass
			l.firstCall = true
			if l.dbIterator.Key() == nil {
				return false
			}
			return true
		}
		if ok := l.dbIterator.Next(); !ok { //if db dont have next, check mem, else increase db index
			return l.memIndex < len(l.memKeySort)
		}
		return true
	} else {
		if !l.firstCall { //the first call is init call in Next() iterator, mem already init at 0, so need to bypass
			l.firstCall = true
			return true
		}
		l.memIndex++                         //increase mem index
		if l.memIndex >= len(l.memKeySort) { //if mem dont have next, check db
			return l.dbIterator.Key() != nil
		}

		return true
	}

}

func (l *LiteStateDBIterator) Key() []byte {
	selectedZone := l.whichZoneToSelect()
	if selectedZone == 0 {
		if len(l.dbIterator.Key()) == 0 {
			return nil
		}
		return l.dbIterator.Key()[len(l.dbPrefix):]
	} else {
		if l.memIndex >= len(l.memKeySort) {
			return []byte{}
		}
		return l.memKeySort[l.memIndex][:]
	}
}

func (l *LiteStateDBIterator) Value() []byte {
	selectedZone := l.whichZoneToSelect()
	if selectedZone == 0 {
		return l.dbIterator.Value()[1:]
	} else {
		if l.memIndex >= len(l.memValueSort) {
			return []byte{}
		}
		return l.memValueSort[l.memIndex]
	}
}

func (l LiteStateDBIterator) Error() error {
	panic("implement me")
}

func (l LiteStateDBIterator) Last() bool {
	panic("implement me")
}

func (l LiteStateDBIterator) Release() {
	panic("implement me")
}
