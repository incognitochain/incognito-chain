package blockchain

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/transaction"
	"errors"
)

type TxViewPoint struct {
	chainID         byte
	listNullifiers  [][]byte
	listCommitments [][]byte
	customTokenTxs  map[int32]*transaction.TxCustomToken

	// hash of best block in current
	currentBestBlockHash common.Hash
}

/*
ListNullifiers returns list nullifers which is contained in TxViewPoint
*/
// #1: joinSplitDescType is "Coin" Or "Bond" or other token
func (view *TxViewPoint) ListNullifiers() [][]byte {
	return view.listNullifiers
}

/*
ListNullifiers returns list nullifers which is contained in TxViewPoint
*/
// #1: joinSplitDescType is "Coin" Or "Bond"
func (view *TxViewPoint) ListCommitments() [][]byte {
	return view.listCommitments
}

/*
CurrentBestBlockHash returns the hash of the best block in the chain the view currently
represents.
*/
func (view *TxViewPoint) CurrentBestBlockHash() *common.Hash {
	return &view.currentBestBlockHash
}

// fetch from desc of tx to get nullifiers and commitments
func (view *TxViewPoint) processFetchTxViewPoint(block *Block, db database.DatabaseInterface, desc *transaction.JoinSplitDesc) ([][]byte, [][]byte, error) {
	acceptedNullifiers := make([][]byte, 0)
	acceptedCommitments := make([][]byte, 0)
	for _, item := range desc.Nullifiers {
		temp, err := db.HasNullifier(item, block.Header.ChainID)
		if err != nil {
			return nil, nil, err
		}
		if !temp {
			acceptedNullifiers = append(acceptedNullifiers, item)
		}
	}
	for _, item := range desc.Commitments {
		temp, err := db.HasCommitment(item, block.Header.ChainID)
		if err != nil {
			return nil, nil, err
		}
		if !temp {
			acceptedCommitments = append(acceptedCommitments, item)
		}
	}
	return acceptedNullifiers, acceptedCommitments, nil
}

/*
fetchTxViewPoint get list nullifiers and commitments from txs in block and check if they are not in Main chain db
return a tx view point which contains list new nullifiers and new commitments from block
*/
func (view *TxViewPoint) fetchTxViewPoint(db database.DatabaseInterface, block *Block) error {
	transactions := block.Transactions
	// Loop through all of the transaction descs (except for the salary tx)
	acceptedNullifiers := make([][]byte, 0)
	acceptedCommitments := make([][]byte, 0)
	for indexTx, tx := range transactions {
		switch tx.GetType() {
		case common.TxNormalType:
			{
				normalTx := tx.(*transaction.Tx)
				for _, desc := range normalTx.Descs {
					temp1, temp2, err := view.processFetchTxViewPoint(block, db, desc)
					acceptedNullifiers = append(acceptedNullifiers, temp1...)
					acceptedCommitments = append(acceptedCommitments, temp2...)
					if err != nil {
						return NewBlockChainError(UnExpectedError, err)
					}
				}
			}
		case common.TxSalaryType:
			{
				normalTx := tx.(*transaction.Tx)
				for _, desc := range normalTx.Descs {
					temp1, temp2, err := view.processFetchTxViewPoint(block, db, desc)
					acceptedNullifiers = append(acceptedNullifiers, temp1...)
					acceptedCommitments = append(acceptedCommitments, temp2...)
					if err != nil {
						return NewBlockChainError(UnExpectedError, err)
					}
				}
			}
		case common.TxRegisterCandidateType:
			{
				votingTx := tx.(*transaction.TxRegisterCandidate)
				for _, desc := range votingTx.Descs {
					temp1, temp2, err := view.processFetchTxViewPoint(block, db, desc)
					acceptedNullifiers = append(acceptedNullifiers, temp1...)
					acceptedCommitments = append(acceptedCommitments, temp2...)
					if err != nil {
						return NewBlockChainError(UnExpectedError, err)
					}
				}
			}
		case common.TxCustomTokenType:
			{
				tx := tx.(*transaction.TxCustomToken)
				for _, desc := range tx.Descs {
					temp1, temp2, err := view.processFetchTxViewPoint(block, db, desc)
					acceptedNullifiers = append(acceptedNullifiers, temp1...)
					acceptedCommitments = append(acceptedCommitments, temp2...)
					if err != nil {
						return NewBlockChainError(UnExpectedError, err)
					}
				}
				view.customTokenTxs[int32(indexTx)] = tx
			}
		default:
			{
				return NewBlockChainError(UnExpectedError, errors.New("Tx type is invalid"))
			}
		}
	}

	if len(acceptedNullifiers) > 0 {
		for _, item := range acceptedNullifiers {
			view.listNullifiers = append(view.listNullifiers, item)
		}
		for _, item := range acceptedCommitments {
			view.listCommitments = append(view.listCommitments, item)
		}
	}
	return nil
}

/*
Create a Tx view point, which contains data about nullifiers and commitments
*/
func NewTxViewPoint(chainId byte) *TxViewPoint {
	return &TxViewPoint{
		chainID:         chainId,
		listNullifiers:  make([][]byte, 0),
		listCommitments: make([][]byte, 0),
		customTokenTxs:  make(map[int32]*transaction.TxCustomToken, 0),
	}
}
