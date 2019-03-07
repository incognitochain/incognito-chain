package frombeaconins

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
)

type SetEncryptionLastBlockIns struct {
	boardType   common.BoardType
	blockHeight uint64
}

func NewSetEncryptionLastBlockIns(boardType common.BoardType, blockHeight uint64) *SetEncryptionLastBlockIns {
	return &SetEncryptionLastBlockIns{boardType: boardType, blockHeight: blockHeight}
}

func (setEncryptionLastBlock *SetEncryptionLastBlockIns) GetStringFormat() ([]string, error) {
	panic("implement me")
}

func (setEncryptionLastBlock *SetEncryptionLastBlockIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) (metadata.Transaction, error) {
	panic("implement me")
}

type SetEncryptionFlagIns struct {
	boardType common.BoardType
	flag      byte
}

func NewSetEncryptionFlagIns(boardType common.BoardType, flag byte) *SetEncryptionFlagIns {
	return &SetEncryptionFlagIns{boardType: boardType, flag: flag}
}

func (setEncryptionFlag *SetEncryptionFlagIns) GetStringFormat() ([]string, error) {
	panic("implement me")
}

func (setEncryptionFlag *SetEncryptionFlagIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) (metadata.Transaction, error) {
	panic("implement me")
}
