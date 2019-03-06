package common

type BoardType byte

func NewBoardTypeFromString(s string) BoardType {
	if s == "dcb" {
		return DCBBoard
	} else {
		return GOVBoard
	}
}

func (boardType *BoardType) Bytes() []byte {
	x := byte(*boardType)
	return []byte{x}
}
