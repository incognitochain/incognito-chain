package database

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/privacy-protocol"
)

// DatabaseInterface provides the interface that is used to store blocks.
type DatabaseInterface interface {
	// Block
	StoreBlock(interface{}, byte) error
	StoreBlockHeader(interface{}, *common.Hash, byte) error
	FetchBlock(*common.Hash) ([]byte, error)
	HasBlock(*common.Hash) (bool, error)
	FetchAllBlocks() (map[byte][]*common.Hash, error)
	FetchChainBlocks(byte) ([]*common.Hash, error)
	DeleteBlock(*common.Hash, int32, byte) error

	// Block index
	StoreBlockIndex(*common.Hash, int32, byte) error
	GetIndexOfBlock(*common.Hash) (int32, byte, error)
	GetBlockByIndex(int32, byte) (*common.Hash, error)

	// Transaction Index
	StoreTransactionIndex(*common.Hash, *common.Hash, int) error
	StoreTransactionLightMode(*privacy.SpendingKey, byte, int32, int, *transaction.Tx) error
	GetTransactionIndexById(*common.Hash) (*common.Hash, int, error)
	GetTransactionLightModeByPrivateKey(*privacy.SpendingKey) (map[byte][]transaction.Tx, error)
	GetTransactionLightModeByHash(*common.Hash) ([]byte, []byte, error)

	// Best state of chain
	StoreBestState(interface{}, byte) error
	FetchBestState(byte) ([]byte, error)
	CleanBestState() error

	// Nullifier
	StoreNullifiers([]byte, byte) error
	FetchNullifiers(byte) ([][]byte, error)
	HasNullifier([]byte, byte) (bool, error)
	CleanNullifiers() error

	// Commitment
	StoreCommitments([]byte, byte) error
	FetchCommitments(byte) ([][]byte, error)
	HasCommitment([]byte, byte) (bool, error)
	CleanCommitments() error

	// Fee estimator
	StoreFeeEstimator([]byte, byte) error
	GetFeeEstimator(byte) ([]byte, error)
	CleanFeeEstimator() error

	// Custom token
	StoreCustomToken(*common.Hash, []byte) error                       // param: tokenID, txInitToken-id, data tx
	StoreCustomTokenTx(*common.Hash, byte, int32, int32, []byte) error // param: tokenID, chainID, block height, tx-id, data tx
	ListCustomToken() ([][]byte, error)
	CustomTokenTxs(*common.Hash) ([]*common.Hash, error)                                                      // token id
	StoreCustomTokenPaymentAddresstHistory(*common.Hash, *transaction.TxCustomToken) error                    // store account history of custom token
	GetCustomTokenListPaymentAddress(*common.Hash) ([][]byte, error)                                          // get all account that have balance > 0 of a custom token
	GetCustomTokenPaymentAddressUTXO(*common.Hash, privacy.PaymentAddress) ([]transaction.TxTokenVout, error) // get list of utxo of an account of a token
	GetCustomTokenListPaymentAddressesBalance(*common.Hash) (map[string]uint64, error)                        // get balance of all payment address of a token (only return payment address with balance > 0)

	// Loans
	StoreLoanRequest([]byte, []byte) error  // param: loanID, tx hash
	StoreLoanResponse([]byte, []byte) error // param: loanID, tx hash
	GetLoanTxs([]byte) ([][]byte, error)    // param: loanID

	// Crowdsale
	SaveCrowdsaleData([]byte, []byte, string, string, uint64, privacy.PaymentAddress) error // param: saleID, bondID, baseAsset, quoteAsset, price, escrowAccount

	Close() error
}
