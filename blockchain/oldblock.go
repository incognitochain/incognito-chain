package blockchain

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/transaction"
)

/*
Block is struct present every block in blockchain
block contains many types of transaction
- normal tx:
- action tx:
*/
type Block struct {
	Header           BlockHeader
	Transactions     []transaction.Transaction
	BlockProducer    string // in base58check.encode
	BlockProducerSig string

	blockHash *common.Hash
}

/*
Customize UnmarshalJSON to parse list TxNormal
because we have many types of block, so we can need to customize data from marshal from json string to build a block
*/
func (self *Block) UnmarshalJSON(data []byte) error {
	Logger.log.Info("UnmarshalJSON of block")
	type Alias Block
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

	self.Header = temp.Alias.Header
	return nil
}

/*
AddTransaction adds a new transaction into block
*/
// #1 - tx
func (self *Block) AddTransaction(tx transaction.Transaction) error {
	if self.Transactions == nil {
		return NewBlockChainError(UnExpectedError, errors.New("Not init tx arrays"))
	}
	self.Transactions = append(self.Transactions, tx)
	return nil
}

/*
Hash creates a hash from block data
*/

func (self Block) Hash() *common.Hash {
	if self.blockHash != nil {
		return self.blockHash
	}

	record := common.EmptyString

	// add data from header
	record += strconv.FormatInt(self.Header.Timestamp, 10) +
		string(self.Header.ChainID) +
		self.Header.MerkleRoot.String() +
		//self.Header.MerkleRootCommitments.String() +
		self.Header.PrevBlockHash.String() +
		strconv.Itoa(int(self.Header.SalaryFund)) +
		strconv.Itoa(int(self.Header.GOVConstitution.GOVParams.SalaryPerTx)) +
		strconv.Itoa(int(self.Header.GOVConstitution.GOVParams.BasicSalary)) +
		strings.Join(self.Header.Committee, ",")

	// add data from body
	record += strconv.Itoa(self.Header.Version) +
		self.BlockProducer +
		self.BlockProducerSig +
		strconv.Itoa(len(self.Transactions)) +
		strconv.Itoa(int(self.Header.Height))

	// add data from tx
	for _, tx := range self.Transactions {
		record += tx.Hash().String()
	}

	hash := common.DoubleHashH([]byte(record))
	self.blockHash = &hash
	return self.blockHash
}

/*func (block *Block) updateDCBConstitution(tx transaction.Transaction, blockgen *BlkTmplGenerator) error {
	txAcceptDCBProposal := tx.(transaction.TxAcceptDCBProposal)
	_, _, _, getTx, err := blockgen.chain.GetTransactionByHash(txAcceptDCBProposal.DCBProposalTXID)
	DCBProposal := getTx.(*transaction.TxSubmitDCBProposal)
	if err != nil {
		return err
	}
	block.Header.DCBConstitution.StartedBlockHeight = block.Header.Height
	block.Header.DCBConstitution.ExecuteDuration = DCBProposal.DCBProposalData.ExecuteDuration
	block.Header.DCBConstitution.ProposalTXID = txAcceptDCBProposal.DCBProposalTXID
	block.Header.DCBConstitution.CurrentDCBNationalWelfare = GetOracleDCBNationalWelfare()

	//	proposalParams := DCBProposal.DCBProposalData.DCBParams // not use yet
	block.Header.DCBConstitution.DCBParams = DCBParams{}
	return nil
}

func (block *Block) updateGOVConstitution(tx transaction.Transaction, blockgen *BlkTmplGenerator) error {
	txAcceptGOVProposal := tx.(transaction.TxAcceptGOVProposal)
	_, _, _, getTx, err := blockgen.chain.GetTransactionByHash(txAcceptGOVProposal.GOVProposalTXID)
	GOVProposal := getTx.(*transaction.TxSubmitGOVProposal)
	if err != nil {
		return err
	}
	block.Header.GOVConstitution.StartedBlockHeight = block.Header.Height
	block.Header.GOVConstitution.ExecuteDuration = GOVProposal.GOVProposalData.ExecuteDuration
	block.Header.GOVConstitution.ProposalTXID = txAcceptGOVProposal.GOVProposalTXID
	block.Header.GOVConstitution.CurrentGOVNationalWelfare = GetOracleGOVNationalWelfare()

	proposalParams := GOVProposal.GOVProposalData.GOVParams
	block.Header.GOVConstitution.GOVParams = GOVParams{
		proposalParams.SalaryPerTx,
		proposalParams.BasicSalary,
		proposalParams.TxFee,
		&SellingBonds{
			proposalParams.SellingBonds.BondsToSell,
			proposalParams.SellingBonds.BondPrice,
			proposalParams.SellingBonds.Maturity,
			proposalParams.SellingBonds.BuyBackPrice,
			proposalParams.SellingBonds.StartSellingAt,
			proposalParams.SellingBonds.SellingWithin,
		},
		&RefundInfo{
			proposalParams.RefundInfo.ThresholdToLargeTx,
			proposalParams.RefundInfo.RefundAmount,
		},
	}
	return nil
}*/
