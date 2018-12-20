package blockchain

import (
	"encoding/json"
	"errors"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/transaction"
)

type ShardBody struct {
	RefBlocks    []BlockRef
	Transactions []metadata.Transaction
}
type BlockRef struct {
	ShardID byte
	Block   common.Hash
}

func (self *ShardBody) Hash() common.Hash {
	record := common.EmptyString
	for _, ref := range self.RefBlocks {
		record += string(ref.ShardID) + ref.Block.String()
	}
	for _, tx := range self.Transactions {
		record += tx.Hash().String()
	}
	return common.DoubleHashH([]byte(record))
}

/*
Customize UnmarshalJSON to parse list TxNormal
because we have many types of block, so we can need to customize data from marshal from json string to build a block
*/
func (self *ShardBody) UnmarshalJSON(data []byte) error {
	Logger.log.Info("UnmarshalJSON of block")
	type Alias ShardBody
	temp := &struct {
		Transactions []map[string]interface{}
		*Alias
	}{
		Alias: (*Alias)(self),
	}

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}

	// process tx from tx interface of temp
	for _, txTemp := range temp.Transactions {
		txTempJson, _ := json.MarshalIndent(txTemp, "", "\t")
		Logger.log.Debugf("Tx json data: ", string(txTempJson))

		var tx metadata.Transaction
		var parseErr error
		switch txTemp["Type"].(string) {
		case common.TxNormalType:
			{
				tx = &transaction.Tx{}
				parseErr = json.Unmarshal(txTempJson, &tx)
			}
		case common.TxSalaryType:
			{
				tx = &transaction.Tx{}
				parseErr = json.Unmarshal(txTempJson, &tx)
			}
		case common.TxCustomTokenType:
			{
				tx = &transaction.TxCustomToken{}
				parseErr = json.Unmarshal(txTempJson, &tx)
			}
		default:
			{
				return NewBlockChainError(UnmashallJsonBlockError, errors.New("Can not parse a wrong tx"))
			}
		}

		if parseErr != nil {
			return NewBlockChainError(UnmashallJsonBlockError, parseErr)
		}
		meta, parseErr := parseMetadata(txTemp["Metadata"])
		if parseErr != nil {
			return NewBlockChainError(UnmashallJsonBlockError, parseErr)
		}
		tx.SetMetadata(meta)
		self.Transactions = append(self.Transactions, tx)
	}

	return nil
}

func (self *ShardBody) CalcMerkleRootShard() *common.Hash {
	//TODO
	var shardTxs map[int][]*common.Hash

	for _, tx := range self.Transactions {
		shardID := int(tx.GetSenderAddrLastByte())
		shardTxs[shardID] = append(shardTxs[shardID], tx.Hash())
	}

	shardsHash := make([]*common.Hash, TestNetParams.ShardsNum)
	for idx, _ := range shardsHash {
		h := &common.Hash{}
		shardsHash[idx], _ = h.NewHashFromStr("")
	}

	for idx, shard := range shardTxs {
		txHashStrConcat := ""

		for _, tx := range shard {
			txHashStrConcat += tx.String()
		}

		h := &common.Hash{}
		hash, _ := h.NewHashFromStr(txHashStrConcat)

		shardsHash[idx] = hash
	}

	merkleRoots := Merkle{}.BuildMerkleTreeOfHashs(shardsHash)
	merkleRoot := merkleRoots[len(merkleRoots)-1]
	return merkleRoot
}

func (self *ShardBody) CalcMerkleRootTx() *common.Hash {
	//TODO
	merkleRoots := Merkle{}.BuildMerkleTreeStore(self.Transactions)
	merkleRoot := merkleRoots[len(merkleRoots)-1]
	return merkleRoot
}

func parseMetadata(meta interface{}) (metadata.Metadata, error) {
	if meta == nil {
		return nil, nil
	}

	mtTemp := map[string]interface{}{}
	metaInBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(metaInBytes, &mtTemp)
	if err != nil {
		return nil, err
	}
	var md metadata.Metadata
	switch int(mtTemp["Type"].(float64)) {
	case metadata.BuyFromGOVRequestMeta:
		md = &metadata.BuySellRequest{}

	case metadata.BuyBackRequestMeta:
		md = &metadata.BuyBackRequest{}

	case metadata.BuyFromGOVResponseMeta:
		md = &metadata.BuySellResponse{}

	case metadata.BuyBackResponseMeta:
		md = &metadata.BuyBackResponse{}

	case metadata.LoanRequestMeta:
		md = &metadata.LoanRequest{}

	case metadata.LoanResponseMeta:
		md = &metadata.LoanResponse{}

	default:
		return nil, errors.New("Could not parse metadata with known types.")
	}

	err = json.Unmarshal(metaInBytes, &md)
	if err != nil {
		return nil, err
	}
	return md, nil
}
