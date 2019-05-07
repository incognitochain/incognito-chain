package metadata

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
)

type UpdatingOracleBoard struct {
	Action        int8
	OraclePubKeys []string          // array of hex string of pub keys
	Signs         map[string][]byte // key: pub key string, value: signature
	MetadataBase
}

func NewUpdatingOracleBoard(
	action int8,
	oraclePubKeys []string,
	signs map[string][]byte,
	metaType int,
) *UpdatingOracleBoard {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &UpdatingOracleBoard{
		Action:        action,
		OraclePubKeys: oraclePubKeys,
		Signs:         signs,
		MetadataBase:  metadataBase,
	}
}

func (uob UpdatingOracleBoard) GetType() int {
	return uob.Type
}

func (uob UpdatingOracleBoard) CheckTransactionFee(
	tr Transaction,
	minFeePerKbTx uint64,
) bool {
	txFee := tr.GetTxFee()
	fullFee := minFeePerKbTx * tr.GetTxActualSize()
	return !(txFee < fullFee)
}

func (uob UpdatingOracleBoard) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	// signatures validation will be done in beacon chain so dont need to do it here anymore
	return true, nil
}

func (uob UpdatingOracleBoard) ValidateSanityData(
	bcr BlockchainRetriever,
	txr Transaction,
) (bool, bool, error) {
	if uob.Action == 0 {
		return false, false, errors.New("Wrong request info's action")
	}
	if len(uob.OraclePubKeys) == 0 {
		return false, false, errors.New("Wrong request info's OraclePubKeys")
	}
	for _, pk := range uob.OraclePubKeys {
		if len(pk) == 0 {
			return false, false, errors.New("Wrong request info's OraclePubKey")
		}
	}
	if len(uob.Signs) == 0 {
		return false, false, errors.New("Wrong request info's Signs")
	}
	for pkStr, sign := range uob.Signs {
		if len(pkStr) == 0 || len(sign) == 0 {
			return false, false, errors.New("Wrong request info's Signs")
		}
	}
	return true, true, nil
}

func (uob UpdatingOracleBoard) ValidateMetadataByItself() bool {
	if uob.Type != UpdatingOracleBoardMeta {
		return false
	}
	if uob.Action != Add && uob.Action != Remove {
		return false
	}
	return true
}

func (uob UpdatingOracleBoard) Hash() *common.Hash {
	record := string(uob.Action)
	for _, pk := range uob.OraclePubKeys {
		record += pk
	}
	// record += string(common.ToBytes(uob.Signs))
	record += uob.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (uob UpdatingOracleBoard) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := map[string]interface{}{
		"meta": uob,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(UpdatingOracleBoardMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (uob UpdatingOracleBoard) CalculateSize() uint64 {
	return calculateSize(uob)
}

func (uob UpdatingOracleBoard) VerifyMultiSigs(
	tx Transaction,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (uob UpdatingOracleBoard) ProcessWhenInsertBlockShard(tx Transaction, retriever BlockchainRetriever) error {
	return nil
}

func (uob UpdatingOracleBoard) IsMinerCreatedMetaType() bool {
	metaType := uob.GetType()
	for _, mType := range minerCreatedMetaTypes {
		if metaType == mType {
			return true
		}
	}
	return false
}

func (uob UpdatingOracleBoard) VerifyMinerCreatedTxBeforeGettingInBlock(
	insts [][]string,
	instsUsed []int,
	shardID byte,
	txr Transaction,
	bcr BlockchainRetriever,
	accumulatedData *component.UsedInstData,
) (bool, error) {
	return true, nil
}
