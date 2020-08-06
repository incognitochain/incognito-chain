package instructionsprocessor

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/pkg/errors"
)

func GetTxDataByHash(
	sDB incdb.Database,
	txID common.Hash,
) (
	metadata.Transaction,
	error,
) {
	blockHash, index, err := rawdbv2.GetTransactionByHash(sDB, txID)
	if err != nil {
		return nil, err
	}
	shardBlockBytes, err := rawdbv2.GetShardBlockByHash(sDB, blockHash)
	if err != nil {
		return nil, err
	}
	shardBlock := NewShardBlock()
	err = json.Unmarshal(shardBlockBytes, shardBlock)
	if err != nil {
		return nil, err
	}
	if err != nil || shardBlock == nil {
		err = errors.Errorf("ERROR", err, "NO Transaction in block with hash", blockHash, "and index", index, "contains", shardBlock.Body.Transactions[index])
		return nil, err
	}
	return shardBlock.Body.Transactions[index], nil
}
