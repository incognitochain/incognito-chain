package txpool

import (
	"github.com/incognitochain/incognito-chain/metadata"
)

// type BlockTxsVerifier interface {
// 	ValidateBlockTransactions(
// 		txP TxPool,
// 		sView interface{},
// 		bcView interface{},
// 		txs []metadata.Transaction,
// 	) bool
// 	ValidateBatchRangeProof([]metadata.Transaction) (bool, error)
// }

// type TxVerifier interface {
// 	ValidateAuthentications(metadata.Transaction) (bool, error)
// 	ValidateDataCorrectness(metadata.Transaction) (bool, error)
// 	ValidateTxZKProof(metadata.Transaction) (bool, error)

// 	ValidateWithBlockChain(
// 		tx metadata.Transaction,
// 		sView interface{},
// 		bcView interface{},
// 	) (bool, error)

// 	ValidateDoubleSpend(
// 		txs []metadata.Transaction,
// 		sView interface{},
// 		bcView interface{},
// 	) (bool, error)

// 	ValidateTxAndAddToListTxs(
// 		txNew metadata.Transaction,
// 		txs []metadata.Transaction,
// 		sView interface{},
// 		bcView interface{},
// 		better func(txA, txB metadata.Transaction) bool, // return true if we want txA when txA and txB is double spending
// 	) (bool, error)

// 	FilterDoubleSpend(
// 		txs []metadata.Transaction,
// 		sView interface{},
// 		bcView interface{},
// 		better func(txA, txB metadata.Transaction) bool, // return true if we want txA when txA and txB is double spending
// 	) ([]metadata.Transaction, error)
// }

type TxsVerifier struct {
}

func (v *TxsVerifier) ValidateAuthentications(tx metadata.Transaction) (bool, error) {
	panic("Implement me")
	// return false, nil
}

func (v *TxsVerifier) ValidateDataCorrectness(metadata.Transaction) (bool, error) {
	panic("Implement me")
	// return false, nil
}

func (v *TxsVerifier) ValidateTxZKProof(metadata.Transaction) (bool, error) {
	panic("Implement me")
	// return false, nil
}

func (v *TxsVerifier) ValidateBlockTransactions(
	sView interface{},
	bcView interface{},
	txs []metadata.Transaction,
) bool {
	panic("Implement me")
	// return false
}

func (v *TxsVerifier) ValidateWithBlockChain(
	tx metadata.Transaction,
	sView interface{},
	bcView interface{},
) (
	bool,
	error,
) {
	panic("Implement me")
	// return false, nil
}

func (v *TxsVerifier) ValidateDoubleSpend(
	txs []metadata.Transaction,
	sView interface{},
	bcView interface{},
) (
	bool,
	error,
) {
	panic("Implement me")
	// return false, nil
}

func (v *TxsVerifier) ValidateTxAndAddToListTxs(
	txNew metadata.Transaction,
	txs []metadata.Transaction,
	sView interface{},
	bcView interface{},
	better func(txA, txB metadata.Transaction) bool, // return true if we want txA when txA and txB is double spending
) (
	bool,
	error,
) {
	panic("Implement me")
	// return false, nil
}

func (v *TxsVerifier) FilterDoubleSpend(
	txs []metadata.Transaction,
	sView interface{},
	bcView interface{},
	better func(txA, txB metadata.Transaction) bool, // return true if we want txA when txA and txB is double spending
) (
	[]metadata.Transaction,
	error,
) {
	panic("Implement me")
	// return false, nil
}
