package metadata

import (
	"github.com/ninjadotorg/constant/database"
)

type SetEncryptionLastBlockMetadata struct {
	boardType   BoardType
	blockHeight uint64

	MetadataBase
}

func (setEncryptionLastBlock *SetEncryptionLastBlockMetadata) ValidateTxWithBlockChain(tx Transaction, bcr BlockchainRetriever, b byte, db database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (setEncryptionLastBlock *SetEncryptionLastBlockMetadata) ValidateSanityData(
	bcr BlockchainRetriever,
	tx Transaction,
) (bool, bool, error) {
	return true, true, nil
}

func (setEncryptionLastBlock *SetEncryptionLastBlockMetadata) ValidateMetadataByItself() bool {
	return true
}

func (setEncryptionLastBlock *SetEncryptionLastBlockMetadata) ProcessWhenInsertBlockShard(tx Transaction, db database.DatabaseInterface) error {
	boardType := setEncryptionLastBlock.boardType
	height := setEncryptionLastBlock.blockHeight
	db.SetEncryptionLastBlockHeight(boardType.BoardTypeDB(), height)
	return nil
}

type SetEncryptionFlagMetadata struct {
	boardType BoardType
	flag      byte

	MetadataBase
}

func (setEncryptionFlag *SetEncryptionFlagMetadata) ValidateTxWithBlockChain(
	tx Transaction,
	bcr BlockchainRetriever,
	b byte,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (setEncryptionFlag *SetEncryptionFlagMetadata) ValidateSanityData(
	bcr BlockchainRetriever,
	tx Transaction,
) (bool, bool, error) {
	return true, true, nil
}

func (setEncryptionFlag *SetEncryptionFlagMetadata) ValidateMetadataByItself() bool {
	return true
}

func (setEncryptionFlag *SetEncryptionFlagMetadata) ProcessWhenInsertBlockShard(
	tx Transaction,
	db database.DatabaseInterface,
) error {
	boardType := setEncryptionFlag.boardType
	flag := setEncryptionFlag.flag
	db.SetEncryptFlag(boardType.BoardTypeDB(), flag)
	return nil
}
