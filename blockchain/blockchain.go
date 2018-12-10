package blockchain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/privacy-protocol/client"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/voting"
	"github.com/ninjadotorg/constant/wallet"
)

const (
	ChainCount = 20
)

/*
blockChain is a view presents for data in blockchain network
because we use 20 chain data to contain all block in system, so
this struct has a array best state with len = 20,
every beststate present for a best block in every chain
*/
type BlockChain struct {
	BestState []*BestState //BestState of 20 chain.

	config    Config
	chainLock sync.RWMutex
}

// config is a descriptor which specifies the blockchain instance configuration.
type Config struct {
	// dataBase defines the database which houses the blocks and will be used to
	// store all metadata created by this package.
	//
	// This field is required.
	DataBase database.DatabaseInterface

	// Interrupt specifies a channel the caller can close to signal that
	// long running operations, such as catching up indexes or performing
	// database migrations, should be interrupted.
	//
	// This field can be nil if the caller does not desire the behavior.
	Interrupt <-chan struct{}

	// chainParams identifies which chain parameters the chain is associated
	// with.
	//
	// This field is required.
	ChainParams *Params

	//Light mode flag
	Light bool
	//Wallet for light mode
	Wallet *wallet.Wallet
	//snapshot reward
	customTokenRewardSnapshot map[string]uint64
}

func (self *BlockChain) GetHeight() int32 {
	return self.BestState[0].BestBlock.Header.Height
}

func (self *BlockChain) GetDCBBoardPubKeys() [][]byte {
	return self.BestState[0].BestBlock.Header.DCBGovernor.DCBBoardPubKeys
}

func (self *BlockChain) GetDCBParams() params.DCBParams {
	return self.BestState[0].BestBlock.Header.DCBConstitution.DCBParams
}

func (self *BlockChain) GetGOVParams() params.GOVParams {
	return self.BestState[0].BestBlock.Header.GOVConstitution.GOVParams
}

func (self *BlockChain) GetLoanTxs(loanID []byte) ([][]byte, error) {
	return self.config.DataBase.GetLoanTxs(loanID)
}

func (self *BlockChain) GetLoanPayment(loanID []byte) (uint64, uint64, uint32, error) {
	return self.config.DataBase.GetLoanPayment(loanID)
}

func (self *BlockChain) GetCrowdsaleData(saleID []byte) (*voting.SaleData, error) {
	return self.config.DataBase.LoadCrowdsaleData(saleID)
}

/*
Init - init a blockchain view from config
*/
func (self *BlockChain) Init(config *Config) error {
	// Enforce required config fields.
	if config.DataBase == nil {
		return NewBlockChainError(UnExpectedError, errors.New("Database is not config"))
	}
	if config.ChainParams == nil {
		return NewBlockChainError(UnExpectedError, errors.New("Chain parameters is not config"))
	}

	self.config = *config

	// Initialize the chain state from the passed database.  When the db
	// does not yet contain any chain state, both it and the chain state
	// will be initialized to contain only the genesis block.
	if err := self.initChainState(); err != nil {
		return err
	}

	for chainIndex, bestState := range self.BestState {
		Logger.log.Infof("BlockChain state for chain #%d (Height %d, Best block hash %+v, Total tx %d, Salary fund %d, Gov Param %+v)",
			chainIndex, bestState.Height, bestState.BestBlockHash.String(), bestState.TotalTxns, bestState.BestBlock.Header.SalaryFund, bestState.BestBlock.Header.GOVConstitution)
	}

	return nil
}

// -------------- Blockchain retriever's implementation --------------
func (self *BlockChain) GetNulltifiersList(chainId byte) ([][]byte, error) {
	txViewPoint, err := self.FetchTxViewPoint(chainId)
	if err != nil {
		return nil, err
	}
	nullifierDb := txViewPoint.ListNullifiers()
	return nullifierDb, nil
}

// GetCustomTokenTxsHash - return list of tx which relate to custom token
func (self *BlockChain) GetCustomTokenTxs(tokenID *common.Hash) (map[common.Hash]metadata.Transaction, error) {
	txHashesInByte, err := self.config.DataBase.CustomTokenTxs(tokenID)
	if err != nil {
		return nil, err
	}
	result := make(map[common.Hash]metadata.Transaction)
	for _, temp := range txHashesInByte {
		_, _, _, tx, err := self.GetTransactionByHash(temp)
		if err != nil {
			return nil, err
		}
		result[*tx.Hash()] = tx
	}
	return result, nil
}

// -------------- End of Blockchain retriever's implementation --------------

/*
// initChainState attempts to load and initialize the chain state from the
// database.  When the db does not yet contain any chain state, both it and the
// chain state are initialized to the genesis block.
*/
func (self *BlockChain) initChainState() error {
	// Determine the state of the chain database. We may need to initialize
	// everything from scratch or upgrade certain buckets.
	var initialized bool
	self.BestState = make([]*BestState, ChainCount)
	for chainId := byte(0); chainId < ChainCount; chainId++ {
		bestStateBytes, err := self.config.DataBase.FetchBestState(chainId)
		if err == nil {
			err = json.Unmarshal(bestStateBytes, &self.BestState[chainId])
			if err != nil {
				initialized = false
			} else {
				initialized = true
			}
		} else {
			initialized = false
		}

		if !initialized {
			// At this point the database has not already been initialized, so
			// initialize both it and the chain state to the genesis block.
			err := self.createChainState(chainId)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

/*
// createChainState initializes both the database and the chain state to the
// genesis block.  This includes creating the necessary buckets and inserting
// the genesis block, so it must only be called on an uninitialized database.
*/
func (self *BlockChain) createChainState(chainId byte) error {
	// Create a new block from genesis block and set it as best block of chain
	var initBlock *Block
	if chainId == 0 {
		initBlock = self.config.ChainParams.GenesisBlock
	} else {
		initBlock = &Block{}
		initBlock.Header = self.config.ChainParams.GenesisBlock.Header
		initBlock.Header.ChainID = chainId
		initBlock.Header.PrevBlockHash = common.Hash{}
	}
	initBlock.Header.Height = 1

	tree := new(client.IncMerkleTree) // Build genesis block commitment merkle tree
	if err := UpdateMerkleTreeForBlock(tree, initBlock); err != nil {
		return NewBlockChainError(UpdateMerkleTreeForBlockError, err)
	}

	self.BestState[chainId] = &BestState{}
	self.BestState[chainId].Init(initBlock, tree)

	err := self.ConnectBlock(initBlock)
	if err != nil {
		Logger.log.Error(err)
		return err
	}

	// store best state
	err = self.StoreBestState(chainId)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	return nil
}

/*
Get block index(height) of block
*/
func (self *BlockChain) GetBlockHeightByBlockHash(hash *common.Hash) (int32, byte, error) {
	return self.config.DataBase.GetIndexOfBlock(hash)
}

/*
Get block hash by block index(height)
*/
func (self *BlockChain) GetBlockHashByBlockHeight(height int32, chainId byte) (*common.Hash, error) {
	return self.config.DataBase.GetBlockByIndex(height, chainId)
}

/*
Fetch DatabaseInterface and get block by index(height) of block
*/
func (self *BlockChain) GetBlockByBlockHeight(height int32, chainId byte) (*Block, error) {
	hashBlock, err := self.config.DataBase.GetBlockByIndex(height, chainId)
	if err != nil {
		return nil, err
	}
	blockBytes, err := self.config.DataBase.FetchBlock(hashBlock)
	if err != nil {
		return nil, err
	}

	block := Block{}
	blockHeader := BlockHeader{}
	if self.config.Light {
		// with light node, we can only get data of header of block
		err = json.Unmarshal(blockBytes, &blockHeader)
		if err != nil {
			return nil, err
		}
		block.Header = blockHeader
	} else {
		err = json.Unmarshal(blockBytes, &block)
		if err != nil {
			return nil, err
		}
	}
	return &block, nil
}

/*
Fetch DatabaseInterface and get block data by block hash
*/
func (self *BlockChain) GetBlockByBlockHash(hash *common.Hash) (*Block, error) {
	blockBytes, err := self.config.DataBase.FetchBlock(hash)
	if err != nil {
		return nil, err
	}
	block := Block{}
	blockHeader := BlockHeader{}
	if self.config.Light {
		// with light node, we can only get data of header of block
		err = json.Unmarshal(blockBytes, &blockHeader)
		if err != nil {
			return nil, err
		}
		block.Header = blockHeader
	} else {
		err = json.Unmarshal(blockBytes, &block)
		if err != nil {
			return nil, err
		}
	}
	return &block, nil
}

/*
Store best state of block(best block, num of tx, ...) into Database
*/
func (self *BlockChain) StoreBestState(chainId byte) error {
	return self.config.DataBase.StoreBestState(self.BestState[chainId], chainId)
}

/*
GetBestState - return a best state from a chain
*/
// #1 - chainID - index of chain
func (self *BlockChain) GetBestState(chainId byte) (*BestState, error) {
	bestState := BestState{}
	bestStateBytes, err := self.config.DataBase.FetchBestState(chainId)
	if err == nil {
		err = json.Unmarshal(bestStateBytes, &bestState)
	}
	return &bestState, err
}

/*
Store block into Database
*/
func (self *BlockChain) StoreBlock(block *Block) error {
	return self.config.DataBase.StoreBlock(block, block.Header.ChainID)
}

/*
	Store Only Block Header into database
*/
func (self *BlockChain) StoreBlockHeader(block *Block) error {
	//Logger.log.Infof("Store Block Header, block header %+v, block hash %+v, chain id %+v",block.Header, block.blockHash, block.Header.ChainID)
	return self.config.DataBase.StoreBlockHeader(block.Header, block.Hash(), block.Header.ChainID)
}

/*
	Store Transaction in Light mode
*/
func (self *BlockChain) StoreUnspentTransactionLightMode(privatKey *privacy.SpendingKey, chainId byte, blockHeight int32, txIndex int, tx *transaction.Tx) error {
	return self.config.DataBase.StoreTransactionLightMode(privatKey, chainId, blockHeight, txIndex, tx)
}

/*
Save index(height) of block by block hash
and
Save block hash by index(height) of block
*/
func (self *BlockChain) StoreBlockIndex(block *Block) error {
	return self.config.DataBase.StoreBlockIndex(block.Hash(), block.Header.Height, block.Header.ChainID)
}

func (self *BlockChain) StoreTransactionIndex(txHash *common.Hash, blockHash *common.Hash, index int) error {
	return self.config.DataBase.StoreTransactionIndex(txHash, blockHash, index)
}

/*
Uses an existing database to update the set of used tx by saving list nullifier of privacy-protocol,
this is a list tx-out which are used by a new tx
*/
func (self *BlockChain) StoreNullifiersFromTxViewPoint(view TxViewPoint) error {
	for _, item1 := range view.listNullifiers {
		err := self.config.DataBase.StoreNullifiers(item1, view.chainID)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
Uses an existing database to update the set of not used tx by saving list commitments of privacy-protocol,
this is a list tx-in which are used by a new tx
*/
func (self *BlockChain) StoreCommitmentsFromTxViewPoint(view TxViewPoint) error {
	for _, item1 := range view.listCommitments {
		err := self.config.DataBase.StoreCommitments(item1, view.chainID)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
Uses an existing database to update the set of used tx by saving list nullifier of privacy-protocol,
this is a list tx-out which are used by a new tx
*/
func (self *BlockChain) StoreNullifiersFromListNullifier(nullifiers [][]byte, chainId byte) error {
	for _, nullifier := range nullifiers {
		err := self.config.DataBase.StoreNullifiers(nullifier, chainId)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
Uses an existing database to update the set of not used tx by saving list commitments of privacy-protocol,
this is a list tx-in which are used by a new tx
*/
func (self *BlockChain) StoreCommitmentsFromListCommitment(commitments [][]byte, chainId byte) error {
	for _, item := range commitments {
		err := self.config.DataBase.StoreCommitments(item, chainId)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
Uses an existing database to update the set of used tx by saving list nullifier of privacy-protocol,
this is a list tx-out which are used by a new tx
*/
func (self *BlockChain) StoreNullifiersFromTx(tx *transaction.Tx) error {
	for _, desc := range tx.Descs {
		for _, nullifier := range desc.Nullifiers {
			chainId, err := common.GetTxSenderChain(tx.AddressLastByte)
			if err != nil {
				return err
			}
			err = self.config.DataBase.StoreNullifiers(nullifier, chainId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/*
Uses an existing database to update the set of not used tx by saving list commitments of privacy-protocol,
this is a list tx-in which are used by a new tx
*/
func (self *BlockChain) StoreCommitmentsFromTx(tx *transaction.Tx) error {
	for _, desc := range tx.Descs {
		for _, item := range desc.Commitments {
			chainId, err := common.GetTxSenderChain(tx.AddressLastByte)
			if err != nil {
				return err
			}
			err = self.config.DataBase.StoreCommitments(item, chainId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/*
Get all blocks in chain
Return block array
*/
func (self *BlockChain) GetAllBlocks() ([][]*Block, error) {
	result := make([][]*Block, 0)
	data, err := self.config.DataBase.FetchAllBlocks()
	if err != nil {
		return nil, err
	}

	for chainID, chain := range data {
		for _, item := range chain {
			blockBytes, err := self.config.DataBase.FetchBlock(item)
			if err != nil {
				return nil, err
			}
			block := Block{}
			blockHeader := BlockHeader{}
			if self.config.Light {
				// with light node, we can only get data of header of block
				err = json.Unmarshal(blockBytes, &blockHeader)
				if err != nil {
					return nil, err
				}
				block.Header = blockHeader
			} else {
				err = json.Unmarshal(blockBytes, &block)
				if err != nil {
					return nil, err
				}
			}
			result[chainID] = append(result[chainID], &block)
		}
	}

	return result, nil
}

func (self *BlockChain) GetChainBlocks(chainID byte) ([]*Block, error) {
	result := make([]*Block, 0)
	data, err := self.config.DataBase.FetchChainBlocks(chainID)
	if err != nil {
		return nil, err
	}

	for _, item := range data {
		blockBytes, err := self.config.DataBase.FetchBlock(item)
		if err != nil {
			return nil, err
		}
		block := Block{}
		blockHeader := BlockHeader{}
		if self.config.Light {
			// with light node, we can only get data of header of block
			err = json.Unmarshal(blockBytes, &blockHeader)
			if err != nil {
				return nil, err
			}
			block.Header = blockHeader
		} else {
			err = json.Unmarshal(blockBytes, &block)
			if err != nil {
				return nil, err
			}
		}
		result = append(result, &block)
	}

	return result, nil
}

/*
Get all hash of blocks in chain
Return hashes array
*/
func (self *BlockChain) GetAllHashBlocks() (map[byte][]*common.Hash, error) {
	data, err := self.config.DataBase.FetchAllBlocks()
	if err != nil {
		return nil, err
	}
	return data, err
}

func (self *BlockChain) GetLoanRequestMeta(loanID []byte) (*metadata.LoanRequest, error) {
	txs, err := self.config.DataBase.GetLoanTxs(loanID)
	if err != nil {
		return nil, err
	}

	for _, txHash := range txs {
		hash := &common.Hash{}
		copy(hash[:], txHash)
		_, _, _, tx, err := self.GetTransactionByHash(hash)
		if err != nil {
			return nil, err
		}
		if tx.GetMetadataType() == metadata.LoanRequestMeta {
			meta := tx.GetMetadata()
			if meta == nil {
				continue
			}
			requestMeta, ok := meta.(*metadata.LoanRequest)
			if !ok {
				continue
			}
			if bytes.Equal(requestMeta.LoanID, loanID) {
				return requestMeta, nil
			}
		}
	}
	return nil, nil
}

func (self *BlockChain) ProcessLoanPayment(tx metadata.Transaction) error {
	value := uint64(0)
	txNormal := tx.(*transaction.Tx)
	for _, desc := range txNormal.Descs {
		for _, note := range desc.Note {
			accountDCB, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
			dcbPk := accountDCB.KeySet.PaymentAddress.Pk
			if bytes.Equal(note.Apk[:], dcbPk) {
				value += note.Value
			}
		}
	}
	meta := tx.GetMetadata().(*metadata.LoanPayment)
	principle, interest, deadline, err := self.config.DataBase.GetLoanPayment(meta.LoanID)
	if meta.PayPrinciple {
		if err != nil {
			return err
		}
		if principle < value {
			value = principle
		}
		principle -= value
	} else {
		requestMeta, err := self.GetLoanRequestMeta(meta.LoanID)
		if err != nil {
			return err
		}
		interestPerPeriod := GetInterestAmount(principle, requestMeta.Params.InterestRate)
		periodInc := uint32(0)
		if value < interest {
			interest -= value
		} else {
			periodInc = 1 + uint32((value-interest)/interestPerPeriod)
			interest = interestPerPeriod - (value-interest)%interestPerPeriod
		}
		deadline = deadline + periodInc*requestMeta.Params.Maturity
	}
	return self.config.DataBase.StoreLoanPayment(meta.LoanID, principle, interest, deadline)
}

func (self *BlockChain) ProcessLoanForBlock(block *Block) error {
	for _, tx := range block.Transactions {
		switch tx.GetMetadataType() {
		case metadata.LoanRequestMeta:
			{
				tx := tx.(*transaction.Tx)
				meta := tx.Metadata.(*metadata.LoanRequest)
				self.config.DataBase.StoreLoanRequest(meta.LoanID, tx.Hash()[:])
			}
		case metadata.LoanResponseMeta:
			{
				tx := tx.(*transaction.Tx)
				meta := tx.Metadata.(*metadata.LoanResponse)
				self.config.DataBase.StoreLoanResponse(meta.LoanID, tx.Hash()[:])
			}
		case metadata.LoanUnlockMeta:
			{
				// Update loan payment info after withdrawing Constant
				tx := tx.(*transaction.Tx)
				meta := tx.GetMetadata().(*metadata.LoanUnlock)
				requestMeta, _ := self.GetLoanRequestMeta(meta.LoanID)
				principle := requestMeta.LoanAmount
				interest := GetInterestAmount(principle, requestMeta.Params.InterestRate)
				self.config.DataBase.StoreLoanPayment(meta.LoanID, principle, interest, uint32(block.Header.Height))
			}
		case metadata.LoanPaymentMeta:
			{
				self.ProcessLoanPayment(tx)
			}
		}
	}
	return nil
}

func (self *BlockChain) UpdateDividendPayout(block *Block) error {
	for _, tx := range block.Transactions {
		switch tx.GetMetadataType() {
		case metadata.DividendMeta:
			{
				tx := tx.(*transaction.Tx)
				meta := tx.Metadata.(*metadata.Dividend)
				tokenID := meta.TokenID
				for _, desc := range tx.Descs {
					for _, note := range desc.Note {
						// TODO(@0xbunyip): replace note.Apk with bytes of PaymentAddress, not just Pk
						paymentAddress := (&privacy.PaymentAddress{}).FromBytes(note.Apk[:])
						utxos, err := self.config.DataBase.GetCustomTokenPaymentAddressUTXO(tokenID, *paymentAddress)
						if err != nil {
							return err
						}
						for _, utxo := range utxos {
							txHash := utxo.GetTxCustomTokenID()
							err := self.config.DataBase.UpdateRewardAccountUTXO(tokenID, *paymentAddress, &txHash, utxo.GetIndex())
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func (self *BlockChain) UpdateVoteCountBoard(block *Block) error {
	DCBEndedBlock := block.Header.DCBGovernor.EndBlock
	GOVEndedBlock := block.Header.GOVGovernor.EndBlock
	for _, tx := range block.Transactions {
		switch tx.GetMetadataType() {
		case metadata.VoteDCBBoardMeta:
			{
				tx := tx.(*transaction.TxCustomToken)
				voteAmount := tx.GetAmountOfVote()
				voteDCBBoardMetadata := tx.Metadata.(*metadata.VoteDCBBoardMetadata)
				err := self.config.DataBase.AddVoteDCBBoard(DCBEndedBlock, tx.TxTokenData.Vins[0].PaymentAddress.Pk, voteDCBBoardMetadata.CandidatePubKey, voteAmount)
				if err != nil {
					return err
				}
			}
		case metadata.VoteGOVBoardMeta:
			{
				tx := tx.(*transaction.TxCustomToken)
				voteAmount := tx.GetAmountOfVote()
				voteGOVBoardMetadata := tx.Metadata.(*metadata.VoteGOVBoardMetadata)
				err := self.config.DataBase.AddVoteGOVBoard(GOVEndedBlock, tx.TxTokenData.Vins[0].PaymentAddress.Pk, voteGOVBoardMetadata.CandidatePubKey, voteAmount)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (self *BlockChain) ProcessCrowdsaleTxs(block *Block) error {
	for _, tx := range block.Transactions {
		switch tx.GetMetadataType() {
		case metadata.VoteDCBBoardMeta:
			{
				txAccepted := tx.(*transaction.TxAcceptDCBProposal)
				_, _, _, getTx, err := self.GetTransactionByHash(txAccepted.DCBProposalTXID)
				proposal := getTx.(*transaction.TxSubmitDCBProposal)
				if err != nil {
					return err
				}

				// Store saledata in db if needed
				if proposal.DCBProposalData.DCBParams.SaleData != nil {
					err := self.config.DataBase.SaveCrowdsaleData(proposal.DCBProposalData.DCBParams.SaleData)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

/*
FetchTxViewPoint -  return a tx view point, which contain list commitments and nullifiers
Param coinType - COIN or BOND
*/
func (self *BlockChain) FetchTxViewPoint(chainId byte) (*TxViewPoint, error) {
	view := NewTxViewPoint(chainId)
	commitments, err := self.config.DataBase.FetchCommitments(chainId)
	if err != nil {
		return nil, err
	}
	view.listCommitments = commitments
	nullifiers, err := self.config.DataBase.FetchNullifiers(chainId)
	if err != nil {
		return nil, err
	}
	view.listNullifiers = nullifiers
	// view.SetBestHash(self.BestState.BestBlockHash)
	return view, nil
}

func (self *BlockChain) CreateAndSaveTxViewPoint(block *Block) error {
	view := NewTxViewPoint(block.Header.ChainID)

	err := view.fetchTxViewPoint(self.config.DataBase, block)
	if err != nil {
		return err
	}

	// check custom token and save
	for indexTx, customTokenTx := range view.customTokenTxs {
		switch customTokenTx.TxTokenData.Type {
		case transaction.CustomTokenInit:
			{
				Logger.log.Info("Store custom token when it is issued", customTokenTx.TxTokenData.PropertyID, customTokenTx.TxTokenData.PropertySymbol, customTokenTx.TxTokenData.PropertyName)
				err = self.config.DataBase.StoreCustomToken(&customTokenTx.TxTokenData.PropertyID, customTokenTx.Hash()[:])
				if err != nil {
					return err
				}
			}
		case transaction.CustomTokenTransfer:
			{
				Logger.log.Info("Transfer custom token %+v", customTokenTx)
			}
		}
		// save tx which relate to custom token
		// Reject Double spend UTXO before enter this state
		err = self.config.DataBase.StoreCustomTokenPaymentAddresstHistory(&customTokenTx.TxTokenData.PropertyID, customTokenTx)
		// TODO: detect/cactch/revert/skip double spend tx
		if err != nil {
			// Skip double spend
			continue
		}
		err = self.config.DataBase.StoreCustomTokenTx(&customTokenTx.TxTokenData.PropertyID, block.Header.ChainID, block.Header.Height, indexTx, customTokenTx.Hash()[:])

		// replace 1000 with proper value for snapshot
		if block.Header.Height%1000 == 0 {
			// list of unreward-utxo
			self.config.customTokenRewardSnapshot, err = self.config.DataBase.GetCustomTokenListPaymentAddressesBalance(&customTokenTx.TxTokenData.PropertyID)
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
	}

	// Update the list nullifiers and commitment set using the state of the used tx view point. This
	// entails adding the new
	// ones created by the block.
	err = self.StoreNullifiersFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	err = self.StoreCommitmentsFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	return nil
}

/*
GetListTxByReadonlyKey - Read all blocks to get txs(not action tx) which can be decrypt by readonly secret key
- Param #1: key - key set which contain readonly-key and pub-key
- Param #2: coinType - which type of joinsplitdesc(COIN or BOND)
*/
func (self *BlockChain) GetListTxByReadonlyKey(keySet *cashec.KeySet) (map[byte][]transaction.Tx, error) {
	results := make(map[byte][]transaction.Tx, 0)

	// lock chain
	self.chainLock.Lock()

	for _, bestState := range self.BestState {
		// get best block
		bestBlock := bestState.BestBlock
		chainId := bestState.BestBlock.Header.ChainID
		blockHeight := bestBlock.Header.Height

		for blockHeight > 0 {
			txsInBlock := bestBlock.Transactions
			txsInBlockAccepted := make([]transaction.Tx, 0)
			for _, txInBlock := range txsInBlock {
				if txInBlock.GetType() == common.TxNormalType || txInBlock.GetType() == common.TxSalaryType {
					tx := txInBlock.(*transaction.Tx)
					copyTx := transaction.Tx{
						Version:  tx.Version,
						JSSig:    tx.JSSig,
						JSPubKey: tx.JSPubKey,
						Fee:      tx.Fee,
						Type:     tx.Type,
						LockTime: tx.LockTime,
						Descs:    make([]*transaction.JoinSplitDesc, 0),
					}
					// try to decrypt each of desc in tx with readonly PubKey and add to txsInBlockAccepted
					listDesc := make([]*transaction.JoinSplitDesc, 0)
					for _, desc := range tx.Descs {
						copyDesc := &transaction.JoinSplitDesc{
							Anchor:        desc.Anchor,
							Commitments:   make([][]byte, 0),
							EncryptedData: make([][]byte, 0),
						}
						if desc.Proof != nil && len(desc.EncryptedData) > 0 {
							// privacy-protocol
							for i, encData := range desc.EncryptedData {
								var epk client.EphemeralPubKey
								copy(epk[:], desc.EphemeralPubKey)
								// var hSig []byte
								// copy(hSig, desc.HSigSeed)
								hSig := client.HSigCRH(desc.HSigSeed, desc.Nullifiers[0], desc.Nullifiers[1], copyTx.JSPubKey)
								note := new(client.Note)
								note, err := client.DecryptNote(encData, keySet.ReadonlyKey.Rk, keySet.PaymentAddress.Tk, epk, hSig)
								spew.Dump(note)
								if err == nil && note != nil {
									copyDesc.EncryptedData = append(copyDesc.EncryptedData, encData)
									copyDesc.AppendNote(note)
									copyDesc.Commitments = append(copyDesc.Commitments, desc.Commitments[i])
								} else {
									continue
								}
							}
						} else {
							// no privacy-protocol
							for i, note := range desc.Note {
								if bytes.Equal(note.Apk[:], keySet.PaymentAddress.Pk[:]) {
									copyDesc.AppendNote(note)
									copyDesc.Commitments = append(copyDesc.Commitments, desc.Commitments[i])
								}
							}
						}
						if len(copyDesc.Note) > 0 {
							listDesc = append(listDesc, copyDesc)
						}
					}
					if len(listDesc) > 0 {
						copyTx.Descs = listDesc
					}
					txsInBlockAccepted = append(txsInBlockAccepted, copyTx)
				}
			}
			// detected some tx can be accepted
			if len(txsInBlockAccepted) > 0 {
				// add to result
				results[chainId] = append(results[chainId], txsInBlockAccepted...)
			}

			// continue with previous block
			blockHeight--
			if blockHeight > 0 {
				// not is genesis block
				preBlockHash := bestBlock.Header.PrevBlockHash
				bestBlock, err := self.GetBlockByBlockHash(&preBlockHash)
				if blockHeight != bestBlock.Header.Height || err != nil {
					// pre-block is not the same block-height with calculation -> invalid blockchain
					return nil, errors.New("Invalid blockchain")
				}
			}
		}
	}

	// unlock chain
	self.chainLock.Unlock()
	return results, nil
}

func (self *BlockChain) DecryptTxByKey(txInBlock metadata.Transaction, nullifiersInDb [][]byte, keys *cashec.KeySet) transaction.Tx {
	tx := txInBlock.(*transaction.Tx)
	copyTx := transaction.Tx{
		Version:         tx.Version,
		JSSig:           tx.JSSig,
		JSPubKey:        tx.JSPubKey,
		Fee:             tx.Fee,
		Type:            tx.Type,
		LockTime:        tx.LockTime,
		Descs:           make([]*transaction.JoinSplitDesc, 0),
		AddressLastByte: tx.AddressLastByte,
	}
	// try to decrypt each of desc in tx with readonly PubKey and add to txsInBlockAccepted
	listDesc := make([]*transaction.JoinSplitDesc, 0)
	for _, desc := range tx.Descs {
		copyDesc := &transaction.JoinSplitDesc{
			Anchor:        desc.Anchor,
			Reward:        desc.Reward,
			Commitments:   make([][]byte, 0),
			EncryptedData: make([][]byte, 0),
		}
		if desc.Proof != nil && len(desc.EncryptedData) > 0 {
			if len(keys.PrivateKey) == 0 || len(keys.ReadonlyKey.Rk) == 0 {
				continue
			}
			// have privacy-protocol
			for i, encData := range desc.EncryptedData {
				var epk client.EphemeralPubKey
				copy(epk[:], desc.EphemeralPubKey)
				hSig := client.HSigCRH(desc.HSigSeed, desc.Nullifiers[0], desc.Nullifiers[1], copyTx.JSPubKey)
				note := new(client.Note)
				note, err := client.DecryptNote(encData, keys.ReadonlyKey.Rk, keys.PaymentAddress.Tk, epk, hSig)
				if err == nil && note != nil && note.Value > 0 {
					// can decrypt data -> got candidate commitment
					candidateCommitment := desc.Commitments[i]
					if len(nullifiersInDb) > 0 {
						// -> check commitment with db nullifiers
						var rho [32]byte
						copy(rho[:], note.Rho)
						candidateNullifier := client.GetNullifier(keys.PrivateKey, rho)
						if len(candidateNullifier) == 0 {
							continue
						}
						checkCandiateNullifier, err := common.SliceBytesExists(nullifiersInDb, candidateNullifier)
						if err != nil || checkCandiateNullifier == true {
							// candidate nullifier is not existed in db
							continue
						}
					}
					copyDesc.EncryptedData = append(copyDesc.EncryptedData, encData)
					copyDesc.AppendNote(note)
					note.Cm = candidateCommitment
					note.Apk = privacy.GeneratePublicKey(keys.PrivateKey)
					copyDesc.Commitments = append(copyDesc.Commitments, candidateCommitment)
				} else {
					continue
				}
			}
		} else {
			for i, note := range desc.Note {
				if bytes.Equal(note.Apk[:], keys.PaymentAddress.Pk[:]) && note.Value > 0 {
					// no privacy-protocol
					candidateCommitment := desc.Commitments[i]
					candidateNullifier := desc.Nullifiers[i]
					if len(nullifiersInDb) > 0 {
						// -> check commitment with db nullifiers
						if len(keys.PrivateKey) > 0 {
							var rho [32]byte
							copy(rho[:], note.Rho)
							candidateNullifier = client.GetNullifier(keys.PrivateKey, rho)
							if len(candidateNullifier) == 0 {
								continue
							}
						}
						checkCandiateNullifier, err := common.SliceBytesExists(nullifiersInDb, candidateNullifier)
						if err != nil || checkCandiateNullifier == true {
							// candidate nullifier is not existed in db
							continue
						}
					}
					copyDesc.AppendNote(note)
					note.Cm = candidateCommitment
					note.Apk = privacy.GeneratePublicKey(keys.PrivateKey)
					copyDesc.Commitments = append(copyDesc.Commitments, candidateCommitment)
				} else {
					continue
				}
			}
		}
		if len(copyDesc.Note) > 0 {
			listDesc = append(listDesc, copyDesc)
		}
	}
	if len(listDesc) > 0 {
		copyTx.Descs = listDesc
	}
	return copyTx
}

// GetListUnspentTxByKeysetInBlock - fetch block to get unspent tx commitment which privatekey can use it
// return a list tx which contain commitment which can be used
func (self *BlockChain) GetListUnspentTxByKeysetInBlock(keys *cashec.KeySet, block *Block, nullifiersInDb [][]byte, returnFullTx bool) (map[byte][]transaction.Tx, error) {
	results := make(map[byte][]transaction.Tx)

	chainId := block.Header.ChainID
	results[chainId] = make([]transaction.Tx, 0)

	txsInBlock := block.Transactions
	txsInBlockAccepted := make([]transaction.Tx, 0)
	for _, txInBlock := range txsInBlock {
		if txInBlock.GetType() == common.TxNormalType || txInBlock.GetType() == common.TxSalaryType {
			// copyTx ONLY contains commitment which relate to keys
			copyTx := self.DecryptTxByKey(txInBlock, nullifiersInDb, keys)
			if len(copyTx.Descs) > 0 {
				if !returnFullTx {
					// only return copy tx which contain unspent commitment which relate with private key
					txsInBlockAccepted = append(txsInBlockAccepted, copyTx)
				} else {
					// only return full tx which contain unspent commitment which relate with private key and other commitments
					txsInBlockAccepted = append(txsInBlockAccepted, *txInBlock.(*transaction.Tx))
				}
			}
		}
	}
	// detected some tx can be accepted
	if len(txsInBlockAccepted) > 0 {
		// add to result
		results[chainId] = append(results[chainId], txsInBlockAccepted...)
	}
	return results, nil
}

/*
GetListTxByPrivateKey - Read all blocks to get txs(not action tx) which can be decrypt by readonly secret key.
With private-key, we can check unspent tx by check nullifiers from database
- Param #1: privateKey - byte[] of privatekey
- Param #2: coinType - which type of joinsplitdesc(COIN or BOND)
*/
func (self *BlockChain) GetListUnspentTxByKeyset(keyset *cashec.KeySet, sortType int, sortAsc bool) (map[byte][]transaction.Tx, error) {
	results := make(map[byte][]transaction.Tx)

	// lock chain
	self.chainLock.Lock()
	defer self.chainLock.Unlock()

	// get list nullifiers from db to check spending
	nullifiersInDb := make([][]byte, 0)
	for _, bestState := range self.BestState {
		bestBlock := bestState.BestBlock
		chainId := bestBlock.Header.ChainID
		txViewPoint, err := self.FetchTxViewPoint(chainId)
		if err != nil {
			return nil, err
		}
		nullifiersInDb = append(nullifiersInDb, txViewPoint.listNullifiers...)
	}
	if self.config.Light {
		// Get unspent tx with light mode
		fullTxs, err := self.config.DataBase.GetTransactionLightModeByPrivateKey(&keyset.PrivateKey)
		Logger.log.Infof("UTXO lightmode %+v", fullTxs)
		if err != nil {
			return nil, err
		}
		// decrypt to get utxo with commitments with relate to private key
		for chainID, txArrays := range fullTxs {
			for _, tx := range txArrays {
				keys := cashec.KeySet{}
				keys.ImportFromPrivateKey(&keyset.PrivateKey)
				copyTx := self.DecryptTxByKey(&tx, nullifiersInDb, &keys)
				results[chainID] = append(results[chainID], copyTx)
			}
		}
		if len(results) > 0 {
			return results, nil
		}
	}
	// loop on all chains
	for _, bestState := range self.BestState {
		// get best block
		block := bestState.BestBlock
		chainId := block.Header.ChainID
		blockHeight := bestState.BestBlock.Header.Height
		// loop on all blocks in chain
		for blockHeight > 0 {
			var err1 error
			// fetch block to get tx
			resultsInChain, err1 := self.GetListUnspentTxByKeysetInBlock(keyset, block, nullifiersInDb, false)
			if err1 != nil {
				// unlock chain
				//self.chainLock.Unlock()
				return nil, err1
			}
			results[chainId] = append(results[chainId], resultsInChain[chainId]...)
			// continue with previous block
			blockHeight--
			if chainId != 0 && blockHeight == 1 {
				break
			}
			if blockHeight > 0 {
				// not is genesis block
				preBlockHash := block.Header.PrevBlockHash
				preBlock, err := self.GetBlockByBlockHash(&preBlockHash)
				if err != nil || blockHeight != preBlock.Header.Height {
					// pre-block is not the same block-height with calculation -> invalid blockchain
					//self.chainLock.Unlock()
					return nil, errors.New("Invalid blockchain")
				}
				block = preBlock
			}
		}
		// sort txs
		transaction.SortArrayTxs(results[chainId], sortType, sortAsc)
	}

	// unlock chain
	//self.chainLock.Unlock()

	return results, nil
}

func (self *BlockChain) GetCommitteCandidate(pubkeyParam string) *CommitteeCandidateInfo {
	for _, bestState := range self.BestState {
		for pubkey, candidateInfo := range bestState.Candidates {
			if pubkey == pubkeyParam {
				return &candidateInfo
			}
		}
	}
	return nil
}

/*
Get Candidate List from all chain and merge all to one - return pubkey of them
*/
func (self *BlockChain) GetCommitteeCandidateList() []string {
	candidatePubkeyList := []string{}
	for _, bestState := range self.BestState {
		for pubkey, _ := range bestState.Candidates {
			if common.IndexOfStr(pubkey, candidatePubkeyList) < 0 {
				candidatePubkeyList = append(candidatePubkeyList, pubkey)
			}
		}
	}
	sort.Slice(candidatePubkeyList, func(i, j int) bool {
		cndInfoi := self.GetCommitteeCandidateInfo(candidatePubkeyList[i])
		cndInfoj := self.GetCommitteeCandidateInfo(candidatePubkeyList[j])
		if cndInfoi.Value == cndInfoj.Value {
			if cndInfoi.Timestamp < cndInfoj.Timestamp {
				return true
			} else if cndInfoi.Timestamp > cndInfoj.Timestamp {
				return false
			} else {
				if cndInfoi.ChainID <= cndInfoj.ChainID {
					return true
				} else if cndInfoi.ChainID < cndInfoj.ChainID {
					return false
				}
			}
		} else if cndInfoi.Value > cndInfoj.Value {
			return true
		} else {
			return false
		}
		return false
	})
	return candidatePubkeyList
}

func (self *BlockChain) GetCommitteeCandidateInfo(nodeAddr string) CommitteeCandidateInfo {
	var cndVal CommitteeCandidateInfo
	for _, bestState := range self.BestState {
		cndValTmp, ok := bestState.Candidates[nodeAddr]
		if ok {
			cndVal.Value += cndValTmp.Value
			if cndValTmp.Timestamp > cndVal.Timestamp {
				cndVal.Timestamp = cndValTmp.Timestamp
				cndVal.ChainID = cndValTmp.ChainID
			}
		}
	}
	return cndVal
}

// GetUnspentTxCustomTokenVout - return all unspent tx custom token out of sender
func (self *BlockChain) GetUnspentTxCustomTokenVout(receiverKeyset cashec.KeySet, tokenID *common.Hash) ([]transaction.TxTokenVout, error) {
	voutList, err := self.config.DataBase.GetCustomTokenPaymentAddressUTXO(tokenID, receiverKeyset.PaymentAddress)
	if err != nil {
		return nil, err
	}
	return voutList, nil
}

func (self *BlockChain) GetTransactionByHashInLightMode(txHash *common.Hash) (byte, *common.Hash, int, metadata.Transaction, error) {
	const (
		bigNumber   = 999999999
		bigNumberTx = 999999999
	)
	var (
		blockHeight uint32
		txIndex     uint32
		chainId     []byte
	)
	// Get transaction
	tx := transaction.Tx{}
	locationByte, txByte, err := self.config.DataBase.GetTransactionLightModeByHash(txHash)
	Logger.log.Info("GetTransactionByHash - 1", locationByte, txByte, err)
	if err != nil {
		return byte(255), nil, -1, nil, err
	}
	err = json.Unmarshal(txByte, &tx)
	if err != nil {
		return byte(255), nil, -1, nil, err
	}
	// Handle string to get chainId, blockheight, txindex information
	locations := strings.Split(string(locationByte), string("-"))
	// Get Chain Id
	chainId = []byte(locations[2])
	// Get Block Height
	tempBlockHeight := []byte(locations[3])
	bufBlockHeight := bytes.NewBuffer(tempBlockHeight)
	err = binary.Read(bufBlockHeight, binary.LittleEndian, &blockHeight)
	if err != nil {
		return byte(255), nil, -1, nil, err
	}
	blockHeight = uint32(bigNumber - blockHeight)
	Logger.log.Info("Testing in GetTransactionByHash, blockHeight", blockHeight)
	block, err := self.GetBlockByBlockHeight(int32(blockHeight), chainId[0])
	if err != nil {
		Logger.log.Error("ERROR in GetTransactionByHash, get Block by height", err)
		return byte(255), nil, -1, nil, err
	}
	//Get txIndex
	tempTxIndex := []byte(locations[4])
	bufTxIndex := bytes.NewBuffer(tempTxIndex)
	err = binary.Read(bufTxIndex, binary.LittleEndian, &txIndex)
	if err != nil {
		return byte(255), nil, -1, nil, err
	}
	txIndex = uint32(bigNumberTx - txIndex)
	Logger.log.Info("Testing in GetTransactionByHash, blockHash, index, tx", block.Hash(), txIndex, tx)
	return chainId[0], block.Hash(), int(txIndex), &tx, nil
}

// GetTransactionByHash - retrieve tx from txId(txHash)
func (self *BlockChain) GetTransactionByHash(txHash *common.Hash) (byte, *common.Hash, int, metadata.Transaction, error) {
	blockHash, index, err := self.config.DataBase.GetTransactionIndexById(txHash)
	if err != nil {
		// check lightweight
		if self.config.Light {
			// with light node, we can try get data in light mode
			Logger.log.Info("ERROR in GetTransactionByHash, change to get in light mode", err)
			return self.GetTransactionByHashInLightMode(txHash)
		}

		return byte(255), nil, -1, nil, err
	}
	block, err := self.GetBlockByBlockHash(blockHash)
	if err != nil {
		Logger.log.Errorf("ERROR", err, "NO Transaction in block with hash &+v", blockHash, "and index", index, "contains", block.Transactions[index])
		return byte(255), nil, -1, nil, err
	}
	Logger.log.Infof("Transaction in block with hash &+v", blockHash, "and index", index, "contains", block.Transactions[index])
	return block.Header.ChainID, blockHash, index, block.Transactions[index], nil
}

// ListCustomToken - return all custom token which existed in network
func (self *BlockChain) ListCustomToken() (map[common.Hash]transaction.TxCustomToken, error) {
	data, err := self.config.DataBase.ListCustomToken()
	if err != nil {
		return nil, err
	}
	result := make(map[common.Hash]transaction.TxCustomToken)
	for _, txData := range data {
		hash := common.Hash{}
		hash.SetBytes(txData)
		_, blockHash, index, tx, err := self.GetTransactionByHash(&hash)
		_ = blockHash
		_ = index
		if err != nil {
			return nil, err
		}
		txCustomToken := tx.(*transaction.TxCustomToken)
		result[txCustomToken.TxTokenData.PropertyID] = *txCustomToken
	}
	return result, nil
}

// GetCustomTokenTxsHash - return list hash of tx which relate to custom token
func (self *BlockChain) GetCustomTokenTxsHash(tokenID *common.Hash) ([]common.Hash, error) {
	txHashesInByte, err := self.config.DataBase.CustomTokenTxs(tokenID)
	if err != nil {
		return nil, err
	}
	result := []common.Hash{}
	for _, temp := range txHashesInByte {
		result = append(result, *temp)
	}
	return result, nil
}

// GetListTokenHolders - return list paymentaddress (in hexstring) of someone who hold custom token in network
func (self *BlockChain) GetListTokenHolders(tokenID *common.Hash) (map[string]uint64, error) {
	result, err := self.config.DataBase.GetCustomTokenListPaymentAddressesBalance(tokenID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (self *BlockChain) GetCustomTokenRewardSnapshot() map[string]uint64 {
	return self.config.customTokenRewardSnapshot
}

func (self *BlockChain) GetNumberOfDCBGovernors() int {
	return NumberOfDCBGovernors
}
func (self *BlockChain) GetNumberOfGOVGovernors() int {
	return NumberOfGOVGovernors
}
