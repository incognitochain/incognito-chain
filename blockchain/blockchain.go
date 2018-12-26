package blockchain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/database/lvdb"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
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

	//LightMode mode flag
	LightMode bool

	//Wallet for light mode
	Wallet *wallet.Wallet

	//snapshot reward
	customTokenRewardSnapshot map[string]uint64
}

func (self *BlockChain) GetDatabase() database.DatabaseInterface {
	return self.config.DataBase
}

func (self *BlockChain) GetHeight() int32 {
	return self.BestState[0].BestBlock.Header.Height
}

func (self *BlockChain) GetDCBBoardPubKeys() [][]byte {
	return self.BestState[0].BestBlock.Header.DCBGovernor.DCBBoardPubKeys
}

func (self *BlockChain) GetGOVBoardPubKeys() [][]byte {
	return self.BestState[0].BestBlock.Header.GOVGovernor.GOVBoardPubKeys
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

func (self *BlockChain) GetCrowdsaleTxs(requestTxHash []byte) ([][]byte, error) {
	return self.config.DataBase.GetCrowdsaleTxs(requestTxHash)
}

func (self *BlockChain) GetCrowdsaleData(saleID []byte) (*voting.SaleData, error) {
	endBlock, buyingAsset, buyingAmount, sellingAsset, sellingAmount, err := self.config.DataBase.LoadCrowdsaleData(saleID)
	var saleData *voting.SaleData
	if err != nil {
		saleData = &voting.SaleData{
			SaleID:        saleID,
			EndBlock:      endBlock,
			BuyingAsset:   buyingAsset,
			BuyingAmount:  buyingAmount,
			SellingAsset:  sellingAsset,
			SellingAmount: sellingAmount,
		}
	}
	return saleData, err
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

// GetOracleParams returns oracle params
func (self *BlockChain) GetOracleParams() *params.Oracle {
	return self.BestState[0].BestBlock.Header.Oracle
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

	self.BestState[chainId] = &BestState{}
	self.BestState[chainId].Init(initBlock /*, tree*/)

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
	if self.config.LightMode {
		err = json.Unmarshal(blockBytes, &block)
		if err != nil {
			// with light node, block data can only contain block header
			err = json.Unmarshal(blockBytes, &blockHeader)
			if err != nil {
				return nil, err
			}
			block.Header = blockHeader
		}
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
	if self.config.LightMode {
		err = json.Unmarshal(blockBytes, &block)
		if err != nil {
			// with light node, block data can only contain block header
			err = json.Unmarshal(blockBytes, &blockHeader)
			if err != nil {
				return nil, err
			}
			block.Header = blockHeader
		}
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
Uses an existing database to update the set of used tx by saving list nullifier of privacy,
this is a list tx-out which are used by a new tx
*/
func (self *BlockChain) StoreSerialNumbersFromTxViewPoint(view TxViewPoint) error {
	for _, item1 := range view.listSerialNumbers {
		err := self.config.DataBase.StoreSerialNumbers(view.tokenID, item1, view.chainID)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
Uses an existing database to update the set of used tx by saving list SNDerivator of privacy,
this is a list tx-out which are used by a new tx
*/
func (self *BlockChain) StoreSNDerivatorsFromTxViewPoint(view TxViewPoint) error {
	for _, item1 := range view.listSnD {
		err := self.config.DataBase.StoreSNDerivators(view.tokenID, item1, view.chainID)

		if err != nil {
			return err
		}
	}
	return nil
}

/*
Uses an existing database to update the set of not used tx by saving list commitments of privacy,
this is a list tx-in which are used by a new tx
*/
func (self *BlockChain) StoreCommitmentsFromTxViewPoint(view TxViewPoint) error {
	for pubkey, item1 := range view.mapCommitments {
		pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
		if err != nil {
			return err
		}
		for _, com := range item1 {
			err = self.config.DataBase.StoreCommitments(view.tokenID, pubkeyBytes, com, view.chainID)
			if err != nil {
				return err
			}
		}
	}
	for pubkey, item1 := range view.mapOutputCoins {
		pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
		if err != nil {
			return err
		}
		for _, com := range item1 {
			lastByte := pubkeyBytes[len(pubkeyBytes)-1]
			chainID, err := common.GetTxSenderChain(lastByte)
			err = self.config.DataBase.StoreOutputCoins(view.tokenID, pubkeyBytes, com.Bytes(), chainID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/*
Uses an existing database to update the set of used tx by saving list nullifier of privacy,
this is a list tx-out which are used by a new tx
*/
/*func (self *BlockChain) StoreNullifiersFromListNullifier(nullifiers [][]byte, chainId byte) error {
	for _, nullifier := range nullifiers {
		err := self.config.DataBase.StoreSerialNumbers(nullifier, chainId)
		if err != nil {
			return err
		}
	}
	return nil
}*/

/*
Uses an existing database to update the set of used tx by saving list nullifier of privacy,
this is a list tx-out which are used by a new tx
*/
/*func (self *BlockChain) StoreNullifiersFromTx(tx *transaction.Tx) error {
	for _, desc := range tx.Proof.InputCoins {
		chainId, err := common.GetTxSenderChain(tx.PubKeyLastByteSender)
		if err != nil {
			return err
		}
		err = self.config.DataBase.StoreSerialNumbers(desc.CoinDetails.SerialNumber.Compress(), chainId)
		if err != nil {
			return err
		}
	}
	return nil
}*/

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
			if self.config.LightMode {
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
		if self.config.LightMode {
			err = json.Unmarshal(blockBytes, &block)
			if err != nil {
				// with light node, block data can only contain block header
				err = json.Unmarshal(blockBytes, &blockHeader)
				if err != nil {
					return nil, err
				}
				block.Header = blockHeader
			}
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
	//TODO: need to update to new txprivacy's fields
	// txNormal := tx.(*transaction.Tx)
	// for _, desc := range txNormal.Descs {
	// 	for _, note := range desc.Note {
	// 		accountDCB, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	// 		dcbPk := accountDCB.KeySet.PaymentAddress.Pk
	// 		if bytes.Equal(note.Apk[:], dcbPk) {
	// 			value += note.Value
	// 		}
	// 	}
	// }
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
				// TODO(@0xbunyip): store multiple responses with different suffixes
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

// parseCustomTokenUTXO helper method for parsing UTXO data for updating dividend payout
/*func (self *BlockChain) parseCustomTokenUTXO(tokenID *common.Hash, pubkey []byte) ([]transaction.TxTokenVout, error) {
	utxoData, err := self.config.DataBase.GetCustomTokenPaymentAddressUTXO(tokenID, pubkey)
	if err != nil {
		return nil, err
	}
	var finalErr error
	vouts := []transaction.TxTokenVout{}
	for key, value := range utxoData {
		keys := strings.Split(key, string(lvdb.Splitter))
		values := strings.Split(value, string(lvdb.Splitter))
		// get unspent and unreward transaction output
		if strings.Compare(values[1], string(lvdb.Unspent)) == 0 {
			vout := transaction.TxTokenVout{}
			vout.PaymentAddress = privacy.PaymentAddress{Pk: pubkey}
			txHash, err := common.Hash{}.NewHash([]byte(keys[3]))
			if err != nil {
				finalErr = err
				continue
			}
			vout.SetTxCustomTokenID(*txHash)
			voutIndexByte := []byte(keys[4])[0]
			voutIndex := int(voutIndexByte)
			vout.SetIndex(voutIndex)
			value, err := strconv.Atoi(values[0])
			if err != nil {
				finalErr = err
				continue
			}
			vout.Value = uint64(value)
			vouts = append(vouts, vout)
		}
	}
	return vouts, finalErr
}*/

func (self *BlockChain) UpdateDividendPayout(block *Block) error {
	for _, tx := range block.Transactions {
		switch tx.GetMetadataType() {
		case metadata.DividendMeta:
			{
				tx := tx.(*transaction.Tx)
				meta := tx.Metadata.(*metadata.Dividend)
				for _, _ = range tx.Proof.OutputCoins {
					keySet := cashec.KeySet{
						PaymentAddress: meta.PaymentAddress,
					}
					vouts, err := self.GetUnspentTxCustomTokenVout(keySet, meta.TokenID)
					if err != nil {
						return err
					}
					for _, vout := range vouts {
						txHash := vout.GetTxCustomTokenID()
						err := self.config.DataBase.UpdateRewardAccountUTXO(meta.TokenID, keySet.PaymentAddress.Pk, &txHash, vout.GetIndex())
						if err != nil {
							return err
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

func (self *BlockChain) UpdateVoteTokenHolder(block *Block) error {
	for _, tx := range block.Transactions {
		switch tx.GetMetadataType() {
		case metadata.SendInitDCBVoteTokenMeta:
			{
				meta := tx.GetMetadata().(*metadata.SendInitDCBVoteTokenMetadata)
				err := self.config.DataBase.SendInitDCBVoteToken(uint32(block.Header.Height), meta.ReceiverPubKey, meta.Amount)
				if err != nil {
					return err
				}
			}
		case metadata.SendInitGOVVoteTokenMeta:
			{
				meta := tx.GetMetadata().(*metadata.SendInitGOVVoteTokenMetadata)
				err := self.config.DataBase.SendInitGOVVoteToken(uint32(block.Header.Height), meta.ReceiverPubKey, meta.Amount)
				if err != nil {
					return err
				}
			}

		}
	}
	return nil
}

func (self *BlockChain) ProcessVoteProposal(block *Block) error {
	nextDCBConstitutionBlockHeight := uint32(block.Header.DCBConstitution.GetEndedBlockHeight())
	nextGOVConstitutionBlockHeight := uint32(block.Header.GOVConstitution.GetEndedBlockHeight())
	for _, tx := range block.Transactions {
		meta := tx.GetMetadata()
		switch tx.GetMetadataType() {
		case metadata.SealedLv3DCBBallotMeta:
			underlieMetadata := meta.(*metadata.SealedLv3DCBBallotMetadata)
			self.config.DataBase.AddVoteLv3Proposal("dcb", nextDCBConstitutionBlockHeight, underlieMetadata.Hash())
		case metadata.SealedLv2DCBBallotMeta:
			underlieMetadata := meta.(*metadata.SealedLv2DCBBallotMetadata)
			self.config.DataBase.AddVoteLv1or2Proposal("dcb", nextDCBConstitutionBlockHeight, &underlieMetadata.PointerToLv3Ballot)
		case metadata.SealedLv1DCBBallotMeta:
			underlieMetadata := meta.(*metadata.SealedLv1DCBBallotMetadata)
			self.config.DataBase.AddVoteLv1or2Proposal("dcb", nextDCBConstitutionBlockHeight, &underlieMetadata.PointerToLv3Ballot)
		case metadata.NormalDCBBallotMetaFromOwnerMeta:
			underlieMetadata := meta.(*metadata.NormalDCBBallotFromOwnerMetadata)
			self.config.DataBase.AddVoteNormalProposalFromOwner("dcb", nextDCBConstitutionBlockHeight, &underlieMetadata.PointerToLv3Ballot, underlieMetadata.Ballot)
		case metadata.NormalDCBBallotMetaFromSealerMeta:
			underlieMetadata := meta.(*metadata.NormalDCBBallotFromSealerMetadata)
			self.config.DataBase.AddVoteNormalProposalFromSealer("dcb", nextDCBConstitutionBlockHeight, &underlieMetadata.PointerToLv3Ballot, underlieMetadata.Ballot)
		case metadata.AcceptDCBProposalMeta:
			underlieMetadata := meta.(*metadata.AcceptDCBProposalMetadata)
			self.config.DataBase.TakeVoteTokenFromWinner("dcb", nextDCBConstitutionBlockHeight, underlieMetadata.Voter)
			self.config.DataBase.SetNewWinningVoter("dcb", nextDCBConstitutionBlockHeight, underlieMetadata.Voter.PubKey)
		case metadata.SealedLv3GOVBallotMeta:
			underlieMetadata := meta.(*metadata.SealedLv3GOVBallotMetadata)
			self.config.DataBase.AddVoteLv3Proposal("gov", nextGOVConstitutionBlockHeight, underlieMetadata.Hash())
		case metadata.SealedLv2GOVBallotMeta:
			underlieMetadata := meta.(*metadata.SealedLv2GOVBallotMetadata)
			self.config.DataBase.AddVoteLv1or2Proposal("gov", nextGOVConstitutionBlockHeight, &underlieMetadata.PointerToLv3Ballot)
		case metadata.SealedLv1GOVBallotMeta:
			underlieMetadata := meta.(*metadata.SealedLv1GOVBallotMetadata)
			self.config.DataBase.AddVoteLv1or2Proposal("gov", nextGOVConstitutionBlockHeight, &underlieMetadata.PointerToLv3Ballot)
		case metadata.NormalGOVBallotMetaFromOwnerMeta:
			underlieMetadata := meta.(*metadata.NormalGOVBallotFromOwnerMetadata)
			self.config.DataBase.AddVoteNormalProposalFromOwner("gov", nextGOVConstitutionBlockHeight, &underlieMetadata.PointerToLv3Ballot, underlieMetadata.Ballot)
		case metadata.NormalGOVBallotMetaFromSealerMeta:
			underlieMetadata := meta.(*metadata.NormalGOVBallotFromSealerMetadata)
			self.config.DataBase.AddVoteNormalProposalFromSealer("gov", nextGOVConstitutionBlockHeight, &underlieMetadata.PointerToLv3Ballot, underlieMetadata.Ballot)
		case metadata.AcceptGOVProposalMeta:
			underlieMetadata := meta.(*metadata.AcceptGOVProposalMetadata)
			self.config.DataBase.TakeVoteTokenFromWinner("gov", nextGOVConstitutionBlockHeight, underlieMetadata.Voter)
			self.config.DataBase.SetNewWinningVoter("gov", nextGOVConstitutionBlockHeight, underlieMetadata.Voter.PubKey)
		}
	}
	return nil
}

func (self *BlockChain) ProcessCrowdsaleTxs(block *Block) error {
	for _, tx := range block.Transactions {
		switch tx.GetMetadataType() {
		case metadata.AcceptDCBProposalMeta:
			{
				meta := tx.GetMetadata().(*metadata.AcceptDCBProposalMetadata)
				_, _, _, getTx, err := self.GetTransactionByHash(&meta.DCBProposalTXID)
				proposal := getTx.GetMetadata().(*metadata.SubmitDCBProposalMetadata)
				if err != nil {
					return err
				}

				// Store saledata in db if needed
				if proposal.DCBParams.SaleData != nil {
					saleData := proposal.DCBParams.SaleData
					if _, _, _, _, _, err := self.config.DataBase.LoadCrowdsaleData(saleData.SaleID); err == nil {
						return fmt.Errorf("SaleID not unique")
					}
					if err := self.config.DataBase.SaveCrowdsaleData(
						saleData.SaleID,
						saleData.EndBlock,
						saleData.BuyingAsset,
						saleData.BuyingAmount,
						saleData.SellingAsset,
						saleData.SellingAmount,
					); err != nil {
						return err
					}
				}
			}
		case metadata.CrowdsaleRequestMeta:
			{
				meta := tx.GetMetadata().(*metadata.CrowdsaleRequest)
				hash := tx.Hash()
				if err := self.config.DataBase.StoreCrowdsaleRequest(hash[:], meta.SaleID, meta.PaymentAddress.Pk[:], meta.PaymentAddress.Tk[:], meta.Info); err != nil {
					return err
				}
			}
		case metadata.CrowdsaleResponseMeta:
			{
				meta := tx.GetMetadata().(*metadata.CrowdsaleResponse)
				_, _, _, txRequest, err := self.GetTransactionByHash(meta.RequestedTxID)
				if err != nil {
					return err
				}
				requestHash := txRequest.Hash()

				hash := tx.Hash()
				if err := self.config.DataBase.StoreCrowdsaleResponse(requestHash[:], hash[:]); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// CreateAndSaveTxViewPointFromBlock - fetch data from block, put into txviewpoint variable and save into db
// need to check light or not light mode
// with light mode - node only fetch outputcoins of account in local wallet -> smaller data
// with not light mode - node fetch all outputcoins of all accounts in network -> big data
// (note: still storage full data of commitments, serialnumbersm snderivator to check double spend)
func (self *BlockChain) CreateAndSaveTxViewPointFromBlock(block *Block) error {
	// Fetch data from block into tx View point
	view := NewTxViewPoint(block.Header.ChainID)
	var err error
	if !self.config.LightMode {
		// skip local wallet -> store full data
		err = view.fetchTxViewPointFromBlock(self.config.DataBase, block, nil)
	} else {
		err = view.fetchTxViewPointFromBlock(self.config.DataBase, block, self.config.Wallet)
	}
	if err != nil {
		return err
	}

	// check normal custom token
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
		err = self.StoreCustomTokenPaymentAddresstHistory(customTokenTx)
		if err != nil {
			// Skip double spend
			return err
		}
		err = self.config.DataBase.StoreCustomTokenTx(&customTokenTx.TxTokenData.PropertyID, block.Header.ChainID, block.Header.Height, indexTx, customTokenTx.Hash()[:])
		if err != nil {
			return err
		}

		// replace 1000 with proper value for snapshot
		if block.Header.Height%1000 == 0 {
			// list of unreward-utxo
			self.config.customTokenRewardSnapshot, err = self.config.DataBase.GetCustomTokenPaymentAddressesBalance(&customTokenTx.TxTokenData.PropertyID)
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
	}

	// check privacy custom token
	for indexTx, privacyCustomTokenSubView := range view.privacyCustomTokenViewPoint {
		privacyCustomTokenTx := view.privacyCustomTokenTxs[indexTx]
		switch privacyCustomTokenTx.TxTokenPrivacyData.Type {
		case transaction.CustomTokenInit:
			{
				Logger.log.Info("Store custom token when it is issued", privacyCustomTokenTx.TxTokenPrivacyData.PropertyID, privacyCustomTokenTx.TxTokenPrivacyData.PropertySymbol, privacyCustomTokenTx.TxTokenPrivacyData.PropertyName)
				err = self.config.DataBase.StorePrivacyCustomToken(&privacyCustomTokenTx.TxTokenPrivacyData.PropertyID, privacyCustomTokenTx.Hash()[:])
				if err != nil {
					return err
				}
			}
		case transaction.CustomTokenTransfer:
			{
				Logger.log.Info("Transfer custom token %+v", privacyCustomTokenTx)
			}
		}
		err = self.config.DataBase.StorePrivacyCustomTokenTx(&privacyCustomTokenTx.TxTokenPrivacyData.PropertyID, block.Header.ChainID, block.Header.Height, indexTx, privacyCustomTokenTx.Hash()[:])
		if err != nil {
			return err
		}

		err = self.StoreSerialNumbersFromTxViewPoint(*privacyCustomTokenSubView)
		if err != nil {
			return err
		}

		err = self.StoreCommitmentsFromTxViewPoint(*privacyCustomTokenSubView)
		if err != nil {
			return err
		}

		err = self.StoreSNDerivatorsFromTxViewPoint(*privacyCustomTokenSubView)
		if err != nil {
			return err
		}
	}

	// Update the list nullifiers and commitment, snd set using the state of the used tx view point. This
	// entails adding the new
	// ones created by the block.
	err = self.StoreSerialNumbersFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	err = self.StoreCommitmentsFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	err = self.StoreSNDerivatorsFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	return nil
}

// /*
// 	Key: token-paymentAddress  -[-]-  {tokenId}  -[-]-  {paymentAddress}  -[-]-  {txHash}  -[-]-  {voutIndex}
//   H: value-spent/unspent-rewarded/unreward
// */
func (self *BlockChain) StoreCustomTokenPaymentAddresstHistory(customTokenTx *transaction.TxCustomToken) error {
	Splitter := lvdb.Splitter
	TokenPaymentAddressPrefix := lvdb.TokenPaymentAddressPrefix
	unspent := lvdb.Unspent
	spent := lvdb.Spent
	unreward := lvdb.Unreward

	tokenKey := TokenPaymentAddressPrefix
	tokenKey = append(tokenKey, Splitter...)
	tokenKey = append(tokenKey, []byte((customTokenTx.TxTokenData.PropertyID).String())...)
	for _, vin := range customTokenTx.TxTokenData.Vins {
		paymentAddressPubkey := base58.Base58Check{}.Encode(vin.PaymentAddress.Pk, 0x00)
		utxoHash := []byte(vin.TxCustomTokenID.String())
		voutIndex := vin.VoutIndex
		paymentAddressKey := tokenKey
		paymentAddressKey = append(paymentAddressKey, Splitter...)
		paymentAddressKey = append(paymentAddressKey, paymentAddressPubkey...)
		paymentAddressKey = append(paymentAddressKey, Splitter...)
		paymentAddressKey = append(paymentAddressKey, utxoHash[:]...)
		paymentAddressKey = append(paymentAddressKey, Splitter...)
		paymentAddressKey = append(paymentAddressKey, byte(voutIndex))
		_, err := self.config.DataBase.HasValue(paymentAddressKey)
		if err != nil {
			return err
		}
		value, err := self.config.DataBase.Get(paymentAddressKey)
		if err != nil {
			return err
		}
		// old value: {value}-unspent-unreward/reward
		values := strings.Split(string(value), string(Splitter))
		if strings.Compare(values[1], string(unspent)) != 0 {
			return errors.New("Double Spend Detected")
		}
		// new value: {value}-spent-unreward/reward
		newValues := values[0] + string(Splitter) + string(spent) + string(Splitter) + values[2]
		if err := self.config.DataBase.Put(paymentAddressKey, []byte(newValues)); err != nil {
			return err
		}
	}
	for index, vout := range customTokenTx.TxTokenData.Vouts {
		paymentAddressPubkey := base58.Base58Check{}.Encode(vout.PaymentAddress.Pk, 0x00)
		utxoHash := []byte(customTokenTx.Hash().String())
		voutIndex := index
		value := vout.Value
		paymentAddressKey := tokenKey
		paymentAddressKey = append(paymentAddressKey, Splitter...)
		paymentAddressKey = append(paymentAddressKey, paymentAddressPubkey...)
		paymentAddressKey = append(paymentAddressKey, Splitter...)
		paymentAddressKey = append(paymentAddressKey, utxoHash[:]...)
		paymentAddressKey = append(paymentAddressKey, Splitter...)
		paymentAddressKey = append(paymentAddressKey, byte(voutIndex))
		ok, err := self.config.DataBase.HasValue(paymentAddressKey)
		// Vout already exist
		if ok {
			return errors.New("UTXO already exist")
		}
		if err != nil {
			return err
		}
		// init value: {value}-unspent-unreward
		paymentAddressValue := strconv.Itoa(int(value)) + string(Splitter) + string(unspent) + string(Splitter) + string(unreward)
		if err := self.config.DataBase.Put(paymentAddressKey, []byte(paymentAddressValue)); err != nil {
			return err
		}
	}
	return nil
}

// DecryptTxByKey - process outputcoin to get outputcoin data which relate to keyset
func (self *BlockChain) DecryptOutputCoinByKey(outCoinTemp *privacy.OutputCoin, keySet *cashec.KeySet, chainID byte, tokenID *common.Hash) *privacy.OutputCoin {
	/*
		- Param keyset - (priv-key, payment-address, readonlykey)
		in case priv-key: return unspent outputcoin tx
		in case readonly-key: return all outputcoin tx with amount value
		in case payment-address: return all outputcoin tx with no amount value
	*/
	pubkeyCompress := outCoinTemp.CoinDetails.PublicKey.Compress()
	if bytes.Equal(pubkeyCompress, keySet.PaymentAddress.Pk[:]) {
		result := &privacy.OutputCoin{
			CoinDetails:          outCoinTemp.CoinDetails,
			CoinDetailsEncrypted: outCoinTemp.CoinDetailsEncrypted,
		}
		if result.CoinDetailsEncrypted != nil {
			if len(keySet.PrivateKey) > 0 || len(keySet.ReadonlyKey.Rk) > 0 {
				// try to decrypt to get more data
				err := result.Decrypt(keySet.ReadonlyKey)
				if err == nil {
					result.CoinDetails = outCoinTemp.CoinDetails
				}
			}
		}
		if len(keySet.PrivateKey) > 0 {
			// check spent with private-key
			result.CoinDetails.SerialNumber = privacy.Eval(new(big.Int).SetBytes(keySet.PrivateKey),
				result.CoinDetails.SNDerivator, privacy.PedCom.G[privacy.SK])
			ok, err := self.config.DataBase.HasSerialNumber(tokenID, result.CoinDetails.SerialNumber.Compress(), chainID)
			if ok || err != nil {
				return nil
			}
		}
		return result
	}
	return nil
}

/*
GetListOutputCoinsByKeyset - Read all blocks to get txs(not action tx) which can be decrypt by readonly secret key.
With private-key, we can check unspent tx by check nullifiers from database
- Param #1: keyset - (priv-key, payment-address, readonlykey)
in case priv-key: return unspent outputcoin tx
in case readonly-key: return all outputcoin tx with amount value
in case payment-address: return all outputcoin tx with no amount value
- Param #2: coinType - which type of joinsplitdesc(COIN or BOND)
*/
func (self *BlockChain) GetListOutputCoinsByKeyset(keyset *cashec.KeySet, chainID byte, tokenID *common.Hash) ([]*privacy.OutputCoin, error) {
	// lock chain
	self.chainLock.Lock()
	defer self.chainLock.Unlock()

	// get list outputcoin of pubkey from db
	outCointsInBytes, err := self.config.DataBase.GetOutcoinsByPubkey(tokenID, keyset.PaymentAddress.Pk[:], chainID)
	if err != nil {
		return nil, err
	}
	// convert from []byte to object
	outCoints := make([]*privacy.OutputCoin, 0)
	for _, item := range outCointsInBytes {
		outcoin := &privacy.OutputCoin{}
		outcoin.Init()
		outcoin.SetBytes(item)
		outCoints = append(outCoints, outcoin)
	}

	// loop on all outputcoin to decrypt data
	results := make([]*privacy.OutputCoin, 0)
	for _, out := range outCoints {
		pubkeyCompress := out.CoinDetails.PublicKey.Compress()
		if bytes.Equal(pubkeyCompress, keyset.PaymentAddress.Pk[:]) {
			out = self.DecryptOutputCoinByKey(out, keyset, chainID, tokenID)
			if out == nil {
				continue
			} else {
				results = append(results, out)
			}
		}
	}
	if err != nil {
		return nil, err
	}

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
	data, err := self.config.DataBase.GetCustomTokenPaymentAddressUTXO(tokenID, receiverKeyset.PaymentAddress.Pk)
	fmt.Println(data)
	if err != nil {
		return nil, err
	}
	splitter := []byte("-[-]-")
	unspent := []byte("unspent")
	voutList := []transaction.TxTokenVout{}
	for key, value := range data {
		keys := strings.Split(key, string(splitter))
		values := strings.Split(value, string(splitter))
		// get unspent and unreward transaction output
		if strings.Compare(values[1], string(unspent)) == 0 {

			vout := transaction.TxTokenVout{}
			vout.PaymentAddress = receiverKeyset.PaymentAddress
			txHash, err := common.Hash{}.NewHashFromStr(string(keys[3]))
			if err != nil {
				return nil, err
			}
			vout.SetTxCustomTokenID(*txHash)
			voutIndexByte := []byte(keys[4])[0]
			voutIndex := int(voutIndexByte)
			vout.SetIndex(voutIndex)
			value, err := strconv.Atoi(values[0])
			if err != nil {
				return nil, err
			}
			vout.Value = uint64(value)
			fmt.Println("GetCustomTokenPaymentAddressUTXO VOUT", vout)
			voutList = append(voutList, vout)
		}
	}
	return voutList, nil
}

// GetTransactionByHash - retrieve tx from txId(txHash)
func (self *BlockChain) GetTransactionByHash(txHash *common.Hash) (byte, *common.Hash, int, metadata.Transaction, error) {
	blockHash, index, err := self.config.DataBase.GetTransactionIndexById(txHash)
	if err != nil {
		return byte(255), nil, -1, nil, err
	}
	block, err := self.GetBlockByBlockHash(blockHash)
	if err != nil {
		Logger.log.Errorf("ERROR", err, "NO Transaction in block with hash &+v", blockHash, "and index", index, "contains", block.Transactions[index])
		return byte(255), nil, -1, nil, NewBlockChainError(UnExpectedError, err)
	}
	Logger.log.Infof("Transaction in block with hash &+v", blockHash, "and index", index, "contains", block.Transactions[index])
	return block.Header.ChainID, blockHash, index, block.Transactions[index], nil
}

func (self *BlockChain) GetTransactionSenderByHash(txHash *common.Hash) ([]byte, error) {
	_, _, _, tx, err := self.GetTransactionByHash(txHash)
	if err != nil {
		return nil, err
	}
	key := tx.GetJSPubKey()

	return key, nil
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
			return nil, NewBlockChainError(UnExpectedError, err)
		}
		txCustomToken := tx.(*transaction.TxCustomToken)
		result[txCustomToken.TxTokenData.PropertyID] = *txCustomToken
	}
	return result, nil
}

// ListCustomToken - return all custom token which existed in network
func (self *BlockChain) ListPrivacyCustomToken() (map[common.Hash]transaction.TxCustomTokenPrivacy, error) {
	data, err := self.config.DataBase.ListPrivacyCustomToken()
	if err != nil {
		return nil, err
	}
	result := make(map[common.Hash]transaction.TxCustomTokenPrivacy)
	for _, txData := range data {
		hash := common.Hash{}
		hash.SetBytes(txData)
		_, blockHash, index, tx, err := self.GetTransactionByHash(&hash)
		_ = blockHash
		_ = index
		if err != nil {
			return nil, err
		}
		txPrivacyCustomToken := tx.(*transaction.TxCustomTokenPrivacy)
		result[txPrivacyCustomToken.TxTokenPrivacyData.PropertyID] = *txPrivacyCustomToken
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

// GetPrivacyCustomTokenTxsHash - return list hash of tx which relate to custom token
func (self *BlockChain) GetPrivacyCustomTokenTxsHash(tokenID *common.Hash) ([]common.Hash, error) {
	txHashesInByte, err := self.config.DataBase.PrivacyCustomTokenTxs(tokenID)
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
	result, err := self.config.DataBase.GetCustomTokenPaymentAddressesBalance(tokenID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (self *BlockChain) GetCustomTokenRewardSnapshot() map[string]uint64 {
	return self.config.customTokenRewardSnapshot
}

func (self *BlockChain) GetNumberOfDCBGovernors() int {
	return common.NumberOfDCBGovernors
}

func (self *BlockChain) GetNumberOfGOVGovernors() int {
	return common.NumberOfGOVGovernors
}

func (self BlockChain) RandomCommitmentsProcess(usableInputCoins []*privacy.InputCoin, randNum int, chainID byte, tokenID *common.Hash) (commitmentIndexs []uint64, myCommitmentIndexs []uint64) {
	return transaction.RandomCommitmentsProcess(usableInputCoins, randNum, self.config.DataBase, chainID, tokenID)
}

func (self BlockChain) CheckSNDerivatorExistence(tokenID *common.Hash, snd *big.Int, chainID byte) (bool, error) {
	return transaction.CheckSNDerivatorExistence(tokenID, snd, chainID, self.config.DataBase)
}

// GetFeePerKbTx - return fee (per kb of tx) from GOV params data
func (self BlockChain) GetFeePerKbTx() uint64 {
	return self.BestState[0].BestBlock.Header.GOVConstitution.GOVParams.FeePerKbTx
}
