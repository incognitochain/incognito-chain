package metadata

import (
	"github.com/ninjadotorg/constant/database"
)

type BoardType byte

func (boardType BoardType) BoardTypeDB() database.BoardTypeDB {
	return database.BoardTypeDB(boardType)
}
