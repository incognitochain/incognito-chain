package database

import (
	"math/big"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/voting"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// DatabaseInterface provides the interface that is used to store blocks.
type DatabaseInterface interface {
	Put(key, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
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
	StoreTransactionIndex(txId *common.Hash, blockHash *common.Hash, indexInBlock int) error
	GetTransactionIndexById(txId *common.Hash) (*common.Hash, int, error)

	// Best state of chain
	StoreBestState(interface{}, byte) error
	FetchBestState(byte) ([]byte, error)
	CleanBestState() error

	// SerialNumber
	StoreSerialNumbers(tokenID *common.Hash, data []byte, chainID byte) error
	FetchSerialNumbers(tokenID *common.Hash, chainID byte) ([][]byte, error)
	HasSerialNumber(tokenID *common.Hash, data []byte, chainID byte) (bool, error)
	CleanSerialNumbers() error

	// PedersenCommitment
	StoreCommitments(tokenID *common.Hash, pubkey []byte, commitment []byte, chainID byte) error
	StoreOutputCoins(tokenID *common.Hash, pubkey []byte, outputcoin []byte, chainID byte) error
	FetchCommitments(tokenID *common.Hash, chainID byte) ([][]byte, error)
	HasCommitment(tokenID *common.Hash, commitment []byte, chainID byte) (bool, error)
	HasCommitmentIndex(tokenID *common.Hash, commitmentIndex uint64, chainID byte) (bool, error)
	GetCommitmentByIndex(tokenID *common.Hash, commitmentIndex uint64, chainID byte) ([]byte, error)
	GetCommitmentIndex(tokenID *common.Hash, commitment []byte, chainId byte) (*big.Int, error)
	GetCommitmentLength(tokenID *common.Hash, chainId byte) (*big.Int, error)
	GetCommitmentIndexsByPubkey(tokenID *common.Hash, pubkey []byte, chainID byte) ([][]byte, error)
	GetOutcoinsByPubkey(tokenID *common.Hash, pubkey []byte, chainID byte) ([][]byte, error)
	CleanCommitments() error

	// SNDerivator
	StoreSNDerivators(tokenID *common.Hash, data big.Int, shardID byte) error
	FetchSNDerivator(tokenID *common.Hash, chainID byte) ([]big.Int, error)
	HasSNDerivator(tokenID *common.Hash, data big.Int, chainID byte) (bool, error)
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
	GetCustomTokenPaymentAddressUTXO(tokenID *common.Hash, pubkey []byte) (map[string]string, error)            // get list of utxo of an paymentaddress.pubkey of a token
	GetCustomTokenPaymentAddressesBalance(tokenID *common.Hash) (map[string]uint64, error)                      // get balance of all paymentaddress of a token (only return payment address with balance > 0)
	UpdateRewardAccountUTXO(*common.Hash, []byte, *common.Hash, int) error
	GetCustomTokenListPaymentAddress(*common.Hash) ([][]byte, error) // get all paymentaddress owner that have balance > 0 of a custom token
	GetCustomTokenPaymentAddressesBalanceUnreward(tokenID *common.Hash) (map[string]uint64, error)

	// privacy Custom token
	StorePrivacyCustomToken(tokenID *common.Hash, data []byte) error // store custom token. Param: tokenID, txInitToken-id, data tx
	StorePrivacyCustomTokenTx(tokenID *common.Hash, chainID byte, blockHeight int32, txIndex int32, txHash []byte) error
	ListPrivacyCustomToken() ([][]byte, error)                          // get list all custom token which issued in network
	PrivacyCustomTokenTxs(tokenID *common.Hash) ([]*common.Hash, error) // from token id get all custom txs

	// Loans
	StoreLoanRequest([]byte, []byte) error                 // param: loanID, tx hash
	StoreLoanResponse([]byte, []byte) error                // param: loanID, tx hash
	GetLoanTxs([]byte) ([][]byte, error)                   // param: loanID
	StoreLoanPayment([]byte, uint64, uint64, uint32) error // param: loanID, principle, interest, deadline
	GetLoanPayment([]byte) (uint64, uint64, uint32, error) // param: loanID; return: principle, interest, deadline

	// Crowdsale
	SaveCrowdsaleData([]byte, int32, []byte, uint64, []byte, uint64) error // param: saleID, end block, buying asset, buying amount, selling asset, selling amount
	LoadCrowdsaleData([]byte) (int32, []byte, uint64, []byte, uint64, error)
	StoreCrowdsaleRequest([]byte, []byte, []byte, []byte, []byte) error
	StoreCrowdsaleResponse([]byte, []byte) error
	GetCrowdsaleTxs([]byte) ([][]byte, error)

	//Vote
	AddVoteDCBBoard(uint32, []byte, []byte, uint64) error
	AddVoteGOVBoard(uint32, []byte, []byte, uint64) error
	GetTopMostVoteDCBGovernor(uint32) (CandidateList, error)
	GetTopMostVoteGOVGovernor(uint32) (CandidateList, error)
	NewIterator(*util.Range, *opt.ReadOptions) iterator.Iterator
	GetKey(string, interface{}) []byte
	GetVoteDCBBoardListPrefix() []byte
	GetVoteGOVBoardListPrefix() []byte
	GetThreePhraseSealerPrefix() []byte
	GetThreePhraseOwnerPrefix() []byte
	GetThreePhraseVoteValuePrefix() []byte
	SendInitDCBVoteToken(uint32, []byte, uint32) error
	SendInitGOVVoteToken(uint32, []byte, uint32) error
	AddVoteLv3Proposal(string, uint32, *common.Hash) error
	AddVoteLv1or2Proposal(string, uint32, *common.Hash) error
	AddVoteNormalProposalFromOwner(string, uint32, *common.Hash, []byte) error
	AddVoteNormalProposalFromSealer(string, uint32, *common.Hash, []byte) error
	GetAmountVoteToken(string, uint32, []byte) (uint32, error)
	TakeVoteTokenFromWinner(string, uint32, voting.Voter) error
	SetNewWinningVoter(string, uint32, []byte) error
	GetDCBVoteTokenAmount(startedBlock uint32, pubKey []byte) (uint32, error)
	GetGOVVoteTokenAmount(startedBlock uint32, pubKey []byte) (uint32, error)

	Close() error
}
