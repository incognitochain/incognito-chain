package blockchain

import (
	"encoding/json"
	"errors"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/transaction"
)

type BlockBodyShard struct {
	RefBlocks    []BlockRef
	Transactions []metadata.Transaction
}

type BlockRef struct {
	ShardID byte
	Block   common.Hash
}

func (self *BlockBodyShard) Hash() common.Hash {
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
func (self *BlockBodyShard) UnmarshalJSON(data []byte) error {
	Logger.log.Info("UnmarshalJSON of block")
	type Alias BlockBodyShard
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

func (self *BlockBodyShard) CalcMerkleRootShard() {
	//TODO

}
func (self *BlockBodyShard) CalcMerkleRootTx() {
	//TODO

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
