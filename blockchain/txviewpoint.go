package blockchain

import (
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/database"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

type TxViewPoint struct {
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

// BestHash returns the hash of the best block in the chain the view currently
// represents.
func (view *TxViewPoint) BestHash() *common.Hash {
	return &view.currentBestHash
}

// SetBestHash sets the hash of the best block in the chain the view currently
// represents.
func (view *TxViewPoint) SetBestHash(hash *common.Hash) {
	view.currentBestHash = *hash
}

/**
// fetchUsedTx get list nullifiers from txs in block and check if they are not in Main chain db
return list new nullifiers
*/
func (view *TxViewPoint) fetchUsedTx(db database.DB, block *Block) (error) {
	transactions := block.Transactions
	// Loop through all of the transaction descs (except for the coinbase tx)
	acceptedNullifiers := make(map[string][][]byte)
	acceptedCommitments := make(map[string][][]byte)
	for _, tx := range transactions[1:] {
		for _, desc := range tx.(*transaction.Tx).Desc {
			for _, item := range desc.Nullifiers {
				temp, err := db.HasNullifier(item, desc.Type)
				if err != nil {
					return err
				}
				if !temp {
					acceptedNullifiers[desc.Type] = append(acceptedNullifiers[desc.Type], item)
				}
			}
			for _, item := range desc.Commitments {
				temp, err := db.HasCommitment(item, desc.Type)
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

func NewTxViewPoint() *TxViewPoint {
	return &TxViewPoint{
		listNullifiers:  make(map[string]([][]byte)),
		listCommitments: make(map[string]([][]byte)),
	}
}
