package metadata

import (
	"github.com/ninjadotorg/constant/database"
)

type BoardType byte

func (boardType BoardType) BoardTypeDB() database.BoardTypeDB {
	return database.BoardTypeDB(boardType)
}

func NewBoardTypeFromString(s string) BoardType {
	if s == "dcb" {
		return DCBBoard
	} else {
		return GOVBoard
	}
}
