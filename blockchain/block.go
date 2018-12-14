package blockchain

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
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
	Transactions     []metadata.Transaction
	BlockProducer    string // in base58check.encode
	BlockProducerSig string

	blockHash *common.Hash
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

	self.Header = temp.Alias.Header
	return nil
}

/*
AddTransaction adds a new transaction into block
*/
// #1 - tx
func (self *Block) AddTransaction(tx metadata.Transaction) error {
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

func (block *Block) updateDCBConstitution(tx metadata.Transaction, blockgen *BlkTmplGenerator) error {
	metadataAcceptDCBProposal := tx.GetMetadata().(*metadata.AcceptDCBProposalMetadata)
	_, _, _, getTx, err := blockgen.chain.GetTransactionByHash(&metadataAcceptDCBProposal.DCBProposalTXID)
	DCBProposal := getTx.GetMetadata().(*metadata.SubmitDCBProposalMetadata)
	if err != nil {
		return err
	}
	block.Header.DCBConstitution.StartedBlockHeight = block.Header.Height
	block.Header.DCBConstitution.ExecuteDuration = DCBProposal.ExecuteDuration
	block.Header.DCBConstitution.ProposalTXID = metadataAcceptDCBProposal.DCBProposalTXID
	block.Header.DCBConstitution.CurrentDCBNationalWelfare = GetOracleDCBNationalWelfare()

	block.Header.DCBConstitution.DCBParams = DCBProposal.DCBParams
	return nil
}

func (block *Block) updateGOVConstitution(tx metadata.Transaction, blockgen *BlkTmplGenerator) error {
	metadataAcceptGOVProposal := tx.GetMetadata().(*metadata.AcceptGOVProposalMetadata)
	_, _, _, getTx, err := blockgen.chain.GetTransactionByHash(&metadataAcceptGOVProposal.GOVProposalTXID)
	GOVProposal := getTx.GetMetadata().(*metadata.SubmitGOVProposalMetadata)
	if err != nil {
		return err
	}
	block.Header.GOVConstitution.StartedBlockHeight = block.Header.Height
	block.Header.GOVConstitution.ExecuteDuration = GOVProposal.ExecuteDuration
	block.Header.GOVConstitution.ProposalTXID = metadataAcceptGOVProposal.GOVProposalTXID
	block.Header.GOVConstitution.CurrentGOVNationalWelfare = GetOracleGOVNationalWelfare()

	block.Header.GOVConstitution.GOVParams = GOVProposal.GOVParams
	return nil
}
