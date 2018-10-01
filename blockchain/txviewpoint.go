package blockchain

import (
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/database"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

type TxViewPoint struct {
	chainId         byte
	listNullifiers  map[string]([][]byte)
	listCommitments map[string]([][]byte)

	// hash of best block in current
	currentBestHash common.Hash
}

func (view *TxViewPoint) ListNullifiers(joinSplitDescType string) [][]byte {
	return view.listNullifiers[joinSplitDescType]
}

func (view *TxViewPoint) ListCommitments(joinSplitDescType string) [][]byte {
	return view.listCommitments[joinSplitDescType]
}

/*
BestHash returns the hash of the best block in the chain the view currently
represents.
*/
func (view *TxViewPoint) BestHash() *common.Hash {
	return &view.currentBestHash
}

/*
SetBestHash sets the hash of the best block in the chain the view currently
represents.
*/
func (view *TxViewPoint) SetBestHash(hash *common.Hash) {
	view.currentBestHash = *hash
}

/*
fetchTxViewPoint get list nullifiers and commitments from txs in block and check if they are not in Main chain db
return a tx view point which contains list new nullifiers and new commitments from block
*/
func (view *TxViewPoint) fetchTxViewPoint(db database.DB, block *Block) error {
	transactions := block.Transactions
	// Loop through all of the transaction descs (except for the coinbase tx)
	acceptedNullifiers := make(map[string][][]byte)
	acceptedCommitments := make(map[string][][]byte)
	for _, tx := range transactions {
		for _, desc := range tx.(*transaction.Tx).Descs {
			for _, item := range desc.Nullifiers {
				temp, err := db.HasNullifier(item, desc.Type, block.Header.ChainID)
				if err != nil {
					return err
				}
				if !temp {
					acceptedNullifiers[desc.Type] = append(acceptedNullifiers[desc.Type], item)
				}
			}
			for _, item := range desc.Commitments {
				temp, err := db.HasCommitment(item, desc.Type, block.Header.ChainID)
				if err != nil {
					return err
				}
				if !temp {
					acceptedCommitments[desc.Type] = append(acceptedCommitments[desc.Type], item)
				}
			}
		}
	}

	if len(acceptedNullifiers) > 0 {
		for key, item := range acceptedNullifiers {
			view.listNullifiers[key] = append(view.listNullifiers[key], item...)
		}
		for key, item := range acceptedCommitments {
			view.listCommitments[key] = append(view.listCommitments[key], item...)
		}
	}
	return nil
}

/*
Create a Tx view point, which contains data about nullifiers and commitments
 */
func NewTxViewPoint(chainId byte) *TxViewPoint {
	return &TxViewPoint{
		chainId:         chainId,
		listNullifiers:  make(map[string]([][]byte)),
		listCommitments: make(map[string]([][]byte)),
	}
}
