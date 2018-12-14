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
		Logger.log.Debugf("TxNormal json data: ", string(txTempJson))
		switch txTemp["Type"].(string) {
		case common.TxNormalType:
			{
				txNormal := &transaction.Tx{}
				_ = json.Unmarshal(txTempJson, &txNormal)
				self.Transactions = append(self.Transactions, txNormal)
			}
		case common.TxSalaryType:
			{
				txNormal := &transaction.Tx{}
				err := json.Unmarshal(txTempJson, &txNormal)
				if err != nil {
					Logger.log.Error(err)
					panic(err)
				}
				self.Transactions = append(self.Transactions, txNormal)
			}
		case common.TxCustomTokenType:
			{
				txCustomToken := &transaction.TxCustomToken{}
				_ = json.Unmarshal(txTempJson, &txCustomToken)
				self.Transactions = append(self.Transactions, txCustomToken)
			}
			/*case common.TxBuyRequest, common.TxSellRequest:
			  {
				  buySellReqTx := &transaction.TxBuySellRequest{}
				  _ = json.Unmarshal(txTempJson, &buySellReqTx)
				  self.Transactions = append(self.Transactions, buySellReqTx)
			  }*/
			/*case common.TxBuyFromGOVResponse:
				  {
					  buyFromGOVResTx := &transaction.TxCustomToken{}
					  _ = json.Unmarshal(txTempJson, &buyFromGOVResTx)
					  self.Transactions = append(self.Transactions, buyFromGOVResTx)
				  }
			  case common.TxSubmitDCBProposal:
				  {
					  submitDCBProposalTx := &transaction.TxSubmitDCBProposal{}
					  _ = json.Unmarshal(txTempJson, &submitDCBProposalTx)
					  self.Transactions = append(self.Transactions, submitDCBProposalTx)
				  }
			  case common.TxSubmitGOVProposal:
				  {
					  submitGOVProposalTx := &transaction.TxSubmitGOVProposal{}
					  _ = json.Unmarshal(txTempJson, &submitGOVProposalTx)
					  self.Transactions = append(self.Transactions, submitGOVProposalTx)
				  }
			  case common.TxVoteDCBProposal:
				  {
					  VoteDCBProposalTx := &transaction.TxVoteDCBProposal{}
					  _ = json.Unmarshal(txTempJson, &VoteDCBProposalTx)
					  self.Transactions = append(self.Transactions, VoteDCBProposalTx)
				  }
			  case common.TxVoteGOVProposal:
				  {
					  VoteDcbProposalTx := &transaction.TxVoteGOVProposal{}
					  _ = json.Unmarshal(txTempJson, &VoteDcbProposalTx)
					  self.Transactions = append(self.Transactions, VoteDcbProposalTx)
				  }
			  case common.TxAcceptDCBProposal:
				  {
					  AcceptDCBProposal := &transaction.TxAcceptDCBProposal{}
					  _ = json.Unmarshal(txTempJson, &AcceptDCBProposal)
					  self.Transactions = append(self.Transactions, AcceptDCBProposal)
				  }
			  case common.TxAcceptGOVProposal:
				  {
					  AcceptGovProposal := &transaction.TxAcceptGOVProposal{}
					  _ = json.Unmarshal(txTempJson, &AcceptGovProposal)
					  self.Transactions = append(self.Transactions, AcceptGovProposal)
				  }
			  case common.TxLoanRequest:
				  {
					  tx := &transaction.TxLoanRequest{}
					  _ = json.Unmarshal(txTempJson, &tx)
					  self.Transactions = append(self.Transactions, tx)
				  }
			  case common.TxLoanResponse:
				  {
					  tx := &transaction.TxLoanResponse{}
					  _ = json.Unmarshal(txTempJson, &tx)
					  self.Transactions = append(self.Transactions, tx)
				  }
			  case common.TxLoanPayment:
				  {
					  tx := &transaction.TxLoanPayment{}
					  _ = json.Unmarshal(txTempJson, &tx)
					  self.Transactions = append(self.Transactions, tx)
				  }
			  case common.TxLoanWithdraw:
				  {
					  tx := &transaction.TxLoanWithdraw{}
					  _ = json.Unmarshal(txTempJson, &tx)
					  self.Transactions = append(self.Transactions, tx)
				  }*/
			/*case common.TxBuySellDCBRequest:
				  {
					  tx := &transaction.TxBuySellRequest{}
					  _ = json.Unmarshal(txTempJson, &tx)
					  self.Transactions = append(self.Transactions, tx)
				  }
			  case common.TxBuySellDCBResponse:
				  {
					  tx := &transaction.TxBuySellDCBResponse{}
					  _ = json.Unmarshal(txTempJson, &tx)
					  self.Transactions = append(self.Transactions, tx)
				  }*/

		default:
			{
				return NewBlockChainError(UnmashallJsonBlockError, errors.New("Can not parse a wrong tx"))
			}
		}
	}
	return nil
}

func (self *BlockBodyShard) CalcMerkleRootShard() {

}
func (self *BlockBodyShard) CalcMerkleRootTx() {

}
