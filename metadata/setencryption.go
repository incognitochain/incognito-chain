package metadata

import (
	"github.com/big0t/constant-chain/common"
	"github.com/big0t/constant-chain/database"
)

type SetEncryptionLastBlockMetadata struct {
	boardType   common.BoardType
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
	db.SetEncryptionLastBlockHeight(boardType, height)
	return nil
}

type SetEncryptionFlagMetadata struct {
	boardType common.BoardType
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
	db.SetEncryptFlag(boardType, flag)
	return nil
}
