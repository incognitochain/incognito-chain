package rawdb_consensus

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
)

func StoreShardFinalityProof(
	db incdb.KeyValueWriter,
	shardID byte,
	hash common.Hash,
	previousBlockHash common.Hash,
	finalityProof interface{},
	reProposeSig interface{},
	rootHash common.Hash,
	producer string,
	produceTime int64,
	proposer string,
	proposeTime int64,
) error {

	key := GetShardFinalityProofKey(shardID, hash)

	m := make(map[string]interface{})
	m["FinalityProof"] = finalityProof
	m["ReProposeSignature"] = reProposeSig
	m["PreviousBlockHash"] = previousBlockHash
	m["RootHash"] = rootHash
	m["Producer"] = producer
	m["ProducerTimeSlot"] = produceTime
	m["Proposer"] = proposer
	m["ProposerTimeSlot"] = proposeTime

	b, err := json.Marshal(&m)
	if err != nil {
		return rawdbv2.NewRawdbError(rawdbv2.StoreShardConsensusRootHashError, err)
	}

	if err := db.Put(key, b); err != nil {
		return rawdbv2.NewRawdbError(rawdbv2.StoreShardConsensusRootHashError, err)
	}

	return nil
}

func GetShardFinalityProof(db incdb.KeyValueReader, shardID byte, hash common.Hash) (map[string]interface{}, error) {

	key := GetShardFinalityProofKey(shardID, hash)

	data, err := db.Get(key)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return m, nil
}
