package database

import (
	"math/big"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/voting"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// DatabaseInterface provides the interface that is used to store blocks.
type DatabaseInterface interface {
	Put(key, value []byte) error
	Get(key []byte) ([]byte, error)
	HasValue(key []byte) (bool, error)

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

	// Transaction index
	StoreTransactionIndex(*common.Hash, *common.Hash, int) error
	StoreTransactionLightMode(*privacy.SpendingKey, byte, int32, int, common.Hash, []byte) error
	GetTransactionIndexById(*common.Hash) (*common.Hash, int, error)
	GetTransactionLightModeByPrivateKey(*privacy.SpendingKey) (map[byte]([]([]byte)), error)
	GetTransactionLightModeByHash(*common.Hash) ([]byte, []byte, error)

	// Best state of chain
	StoreBestState(interface{}, byte) error
	FetchBestState(byte) ([]byte, error)
	CleanBestState() error

	// SerialNumber
	StoreSerialNumbers([]byte, byte) error
	FetchSerialNumbers(byte) ([][]byte, error)
	HasSerialNumber([]byte, byte) (bool, error)
	//HasSerialNumberIndex(serialNumberIndex int64, chainID byte) (bool, error)
	//GetSerialNumberByIndex(serialNumberIndex int64, chainID byte) ([]byte, error)
	CleanSerialNumbers() error

	// PedersenCommitment
	StoreCommitments(commitment []byte, chainID byte) error
	FetchCommitments(chainID byte) ([][]byte, error)
	HasCommitment(commitment []byte, chainID byte) (bool, error)
	HasCommitmentIndex(commitmentIndex uint64, chainID byte) (bool, error)
	GetCommitmentByIndex(commitmentIndex uint64, chainID byte) ([]byte, error)
	GetCommitmentIndex(commitment []byte, chainId byte) (*big.Int, error)
	GetCommitmentLength(chainId byte) (*big.Int, error)
	CleanCommitments() error

	// SNDerivator
	StoreSNDerivators(big.Int, byte) error
	FetchSNDerivator(byte) ([]big.Int, error)
	HasSNDerivator(big.Int, byte) (bool, error)
	CleanSNDerivator() error

	// Fee estimator
	StoreFeeEstimator([]byte, byte) error
	GetFeeEstimator(byte) ([]byte, error)
	CleanFeeEstimator() error

	// Custom token
	StoreCustomToken(tokenID *common.Hash, data []byte) error                                                   // store custom token. Param: tokenID, txInitToken-id, data tx
	StoreCustomTokenTx(tokenID *common.Hash, chainID byte, blockHeight int32, txIndex int32, data []byte) error // store custom token tx. Param: tokenID, chainID, block height, tx-id, data tx
	ListCustomToken() ([][]byte, error)                                                                         // get list all custom token which issued in network
	CustomTokenTxs(tokenID *common.Hash) ([]*common.Hash, error)                                                // from token id get all custom txs
	//StoreCustomTokenPaymentAddresstHistory(tokenID *common.Hash, customTokenTxData *transaction.TxCustomToken) error // store account history of custom token
	GetCustomTokenPaymentAddressUTXO(tokenID *common.Hash, pubkey []byte) (map[string]string, error) // get list of utxo of an paymentaddress.pubkey of a token
	GetCustomTokenPaymentAddressesBalance(tokenID *common.Hash) (map[string]uint64, error)           // get balance of all paymentaddress of a token (only return payment address with balance > 0)
	//UpdateRewardAccountUTXO(*common.Hash, []byte, *common.Hash, int) error
	//GetCustomTokenListPaymentAddress(*common.Hash) ([][]byte, error)                                  // get all paymentaddress owner that have balance > 0 of a custom token

	// Loans
	StoreLoanRequest([]byte, []byte) error                 // param: loanID, tx hash
	StoreLoanResponse([]byte, []byte) error                // param: loanID, tx hash
	GetLoanTxs([]byte) ([][]byte, error)                   // param: loanID
	StoreLoanPayment([]byte, uint64, uint64, uint32) error // param: loanID, principle, interest, deadline
	GetLoanPayment([]byte) (uint64, uint64, uint32, error) // param: loanID; return: principle, interest, deadline

	// Crowdsale
	SaveCrowdsaleData(*voting.SaleData) error
	LoadCrowdsaleData([]byte) (*voting.SaleData, error)

	//Vote
	AddVoteDCBBoard(uint32, []byte, []byte, uint64) error
	AddVoteGOVBoard(uint32, []byte, []byte, uint64) error
	GetTopMostVoteDCBGovernor(uint32) (CandidateList, error)
	GetTopMostVoteGOVGovernor(uint32) (CandidateList, error)
	NewIterator(*util.Range, *opt.ReadOptions) iterator.Iterator
	GetKey(string, interface{}) []byte
	GetVoteDCBBoardListPrefix() []byte
	GetVoteGOVBoardListPrefix() []byte
	SendInitDCBVoteToken(uint32, []byte, uint64) error
	SendInitGOVVoteToken(uint32, []byte, uint64) error
	AddVoteLv3Proposal(string, uint32, *common.Hash) error
	AddVoteLv1or2Proposal(string, uint32, *common.Hash) error
	AddVoteNormalProposalFromOwner(string, uint32, *common.Hash) error
	AddVoteNormalProposalFromSealer(string, uint32, *common.Hash) error

	ReverseGetKey(string, []byte) (interface{}, error)

	Close() error
}
