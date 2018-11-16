package mempool

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy-protocol/client"
	"github.com/ninjadotorg/constant/transaction"
	"golang.org/x/crypto/sha3"
)

// config is a descriptor containing the memory pool configuration.
type Config struct {
	// Policy defines the various mempool configuration options related
	// to policy.
	Policy Policy

	// Block chain of node
	BlockChain *blockchain.BlockChain

	DataBase database.DatabaseInterface

	ChainParams *blockchain.Params

	// FeeEstimatator provides a feeEstimator. If it is not nil, the mempool
	// records all new transactions it observes into the feeEstimator.
	FeeEstimator map[byte]*FeeEstimator
}

// TxDesc is transaction description in mempool
type TxDesc struct {
	// transaction details
	Desc transaction.TxDesc

	StartingPriority int
}

// TxPool is transaction pool
type TxPool struct {
	// The following variables must only be used atomically.
	lastUpdated int64 // last time pool was updated

	mtx    sync.RWMutex
	config Config
	pool   map[common.Hash]*TxDesc
}

/*
Init Txpool from config
*/
func (tp *TxPool) Init(cfg *Config) {
	tp.config = *cfg
	tp.pool = make(map[common.Hash]*TxDesc)
}

// check transaction in pool
func (tp *TxPool) isTxInPool(hash *common.Hash) bool {
	if _, exists := tp.pool[*hash]; exists {
		return true
	}
	return false
}

/*
// add transaction into pool
*/
func (tp *TxPool) addTx(tx transaction.Transaction, height int32, fee uint64) *TxDesc {
	txD := &TxDesc{
		Desc: transaction.TxDesc{
			Tx:     tx,
			Added:  time.Now(),
			Height: height,
			Fee:    fee,
		},
		StartingPriority: 1, //@todo we will apply calc function for it.
	}
	log.Printf(tx.Hash().String())
	tp.pool[*tx.Hash()] = txD
	atomic.StoreInt64(&tp.lastUpdated, time.Now().Unix())

	// Record this tx for fee estimation if enabled. only apply for normal tx
	if tx.GetType() == common.TxNormalType {
		if tp.config.FeeEstimator != nil {
			chainId, err := common.GetTxSenderChain(tx.(*transaction.Tx).AddressLastByte)
			if err == nil {
				tp.config.FeeEstimator[chainId].ObserveTransaction(txD)
			} else {
				Logger.log.Error(err)
			}
		}
	} else if tx.GetType() == common.TxRegisterCandidateType {
		if tp.config.FeeEstimator != nil {
			chainId, err := common.GetTxSenderChain(tx.(*transaction.TxRegisterCandidate).AddressLastByte)
			if err == nil {
				tp.config.FeeEstimator[chainId].ObserveTransaction(txD)
			} else {
				Logger.log.Error(err)
			}
		}
	}

	return txD
}

/*
// maybeAcceptTransaction is the internal function which implements the public
// MaybeAcceptTransaction.  See the comment for MaybeAcceptTransaction for
// more details.
//
// This function MUST be called with the mempool lock held (for writes).
*/
func (tp *TxPool) maybeAcceptTransaction(tx transaction.Transaction) (*common.Hash, *TxDesc, error) {
	txHash := tx.Hash()

	// that make sure transaction is accepted when passed any rules
	var chainID byte
	var err error

	// get chainID of tx
	chainID, err = common.GetTxSenderChain(tx.GetSenderAddrLastByte())

	if err != nil {
		return nil, nil, err
	}
	bestHeight := tp.config.BlockChain.BestState[chainID].BestBlock.Header.Height
	// nextBlockHeight := bestHeight + 1
	// Check tx with policy
	// check version
	ok := tp.config.Policy.CheckTxVersion(&tx)
	if !ok {
		err := MempoolTxError{}
		err.Init(RejectVersion, errors.New(fmt.Sprintf("%+v's version is invalid", txHash.String())))
		return nil, nil, err
	}

	// check fee of tx
	txFee, err := tp.CheckTransactionFee(tx)
	if err != nil {
		return nil, nil, err
	}
	// end check with policy

	// validate tx with data of blockchain
	err = tp.ValidateTxWithBlockChain(tx, chainID)
	if err != nil {
		return nil, nil, err
	}

	// Don't accept the transaction if it already exists in the pool.
	if tp.isTxInPool(txHash) {
		str := fmt.Sprintf("already have transaction %+v", txHash.String())
		err := MempoolTxError{}
		err.Init(RejectDuplicateTx, errors.New(str))
		return nil, nil, err
	}

	// A standalone transaction must not be a salary transaction.
	if tp.config.BlockChain.IsSalaryTx(tx) {
		err := MempoolTxError{}
		err.Init(RejectSalaryTx, errors.New(fmt.Sprintf("%+v is salary tx", txHash.String())))
		return nil, nil, err
	}

	// sanity data
	if validate, errS := tp.ValidateSanityData(tx); !validate {
		err := MempoolTxError{}
		err.Init(RejectSansityTx, errors.New(fmt.Sprintf("transaction's sansity %v is error %v", txHash.String(), errS.Error())))
		return nil, nil, err
	}

	// Validate tx by it self
	validate := tx.ValidateTransaction()
	if !validate {
		err := MempoolTxError{}
		err.Init(RejectInvalidTx, errors.New("Invalid tx"))
		return nil, nil, err
	}

	txD := tp.addTx(tx, bestHeight, txFee)
	return tx.Hash(), txD, nil
}

// remove transaction for pool
func (tp *TxPool) removeTx(tx *transaction.Transaction) error {
	Logger.log.Infof((*tx).Hash().String())
	if _, exists := tp.pool[*(*tx).Hash()]; exists {
		delete(tp.pool, *(*tx).Hash())
		atomic.StoreInt64(&tp.lastUpdated, time.Now().Unix())
		return nil
	} else {
		return errors.New("Not exist tx in pool")
	}
	return nil
}

// Validate sanity for normal tx data
func (tp *TxPool) validateSanityNormalTxData(tx *transaction.Tx) (bool, error) {
	txN := tx
	//check version
	if txN.Version > transaction.TxVersion {
		return false, errors.New("Wrong tx version")
	}
	// check LockTime before now
	if int64(txN.LockTime) > time.Now().Unix() {
		return false, errors.New("Wrong tx locktime")
	}
	// check Type is normal or salary tx
	if len(txN.Type) != 1 || (txN.Type != common.TxNormalType && txN.Type != common.TxSalaryType) { // only 1 byte
		return false, errors.New("Wrong tx type")
	}
	// check length of JSPubKey
	if len(txN.JSPubKey) != 64 {
		return false, errors.New("Wrong tx jspubkey")
	}
	// check length of JSSig
	if len(txN.JSSig) != 64 {
		return false, errors.New("Wrong tx jssig")
	}
	//check Descs

	for _, desc := range txN.Descs {
		// check length of Anchor
		if len(desc.Anchor) != 2 {
			return false, errors.New("Wrong tx desc's anchor")
		}
		// check length of EphemeralPubKey
		if len(desc.EphemeralPubKey) != client.EphemeralKeyLength {
			return false, errors.New("Wrong tx desc's ephemeralpubkey")
		}
		// check length of HSigSeed
		if len(desc.HSigSeed) != 32 {
			return false, errors.New("Wrong tx desc's hsigseed")
		}
		// check length of Nullifiers
		if len(desc.Nullifiers) != 2 {
			return false, errors.New("Wrong tx desc's nullifiers")
		}
		if len(desc.Nullifiers[0]) != 32 {
			return false, errors.New("Wrong tx desc's nullifiers")
		}
		if len(desc.Nullifiers[1]) != 32 {
			return false, errors.New("Wrong tx desc's nullifiers")
		}
		// check length of Commitments
		if len(desc.Commitments) != 2 {
			return false, errors.New("Wrong tx desc's commitments")
		}
		if len(desc.Commitments[0]) != 32 {
			return false, errors.New("Wrong tx desc's commitments")
		}
		if len(desc.Commitments[1]) != 32 {
			return false, errors.New("Wrong tx desc's commitments")
		}
		// check length of Vmacs
		if len(desc.Vmacs) != 2 {
			return false, errors.New("Wrong tx desc's vmacs")
		}
		if len(desc.Vmacs[0]) != 32 {
			return false, errors.New("Wrong tx desc's vmacs")
		}
		if len(desc.Vmacs[1]) != 32 {
			return false, errors.New("Wrong tx desc's vmacs")
		}
		//
		if desc.Proof == nil {
			return false, errors.New("Wrong tx desc's proof")
		}
		// check length of Proof
		if len(desc.Proof.G_A) != 33 ||
			len(desc.Proof.G_APrime) != 33 ||
			len(desc.Proof.G_B) != 65 ||
			len(desc.Proof.G_BPrime) != 33 ||
			len(desc.Proof.G_C) != 33 ||
			len(desc.Proof.G_CPrime) != 33 ||
			len(desc.Proof.G_K) != 33 ||
			len(desc.Proof.G_H) != 33 {
			return false, errors.New("Wrong tx desc's proof")
		}
		//
		if len(desc.EncryptedData) != 2 {
			return false, errors.New("Wrong tx desc's encryptedData")
		}
		// check nulltifier is existed in DB
		if desc.Reward != 0 {
			return false, errors.New("Wrong tx desc's reward")
		}
	}
	return true, nil
}

// Validate sanity for registration candidate tx data
func (tp *TxPool) validateSanityVotingTxData(txVoting *transaction.TxRegisterCandidate) (bool, error) {
	if !common.ValidateNodeAddress(txVoting.PublicKey) {
		return false, errors.New("Wrong voting node data")
	}
	tx := txVoting.Tx
	txN := tx
	//check version
	if txN.Version > transaction.TxVersion {
		return false, errors.New("Wrong tx version")
	}
	// check LockTime before now
	if int64(txN.LockTime) > time.Now().Unix() {
		return false, errors.New("Wrong tx locktime")
	}
	// check Type equal "n"
	if txN.Type != common.TxRegisterCandidateType {
		return false, errors.New("Wrong tx type")
	}
	// check length of JSPubKey
	if len(txN.JSPubKey) != 64 {
		return false, errors.New("Wrong tx jspubkey")
	}
	// check length of JSSig
	if len(txN.JSSig) != 64 {
		return false, errors.New("Wrong tx jssig")
	}
	//check Descs

	for _, desc := range txN.Descs {
		// check length of Anchor
		if len(desc.Anchor) != 2 {
			return false, errors.New("Wrong tx desc's anchor")
		}
		// check length of EphemeralPubKey
		if len(desc.EphemeralPubKey) != client.EphemeralKeyLength {
			return false, errors.New("Wrong tx desc's ephemeralpubkey")
		}
		// check length of HSigSeed
		if len(desc.HSigSeed) != 32 {
			return false, errors.New("Wrong tx desc's hsigseed")
		}
		// check length of Nullifiers
		if len(desc.Nullifiers) != 2 {
			return false, errors.New("Wrong tx desc's nullifiers")
		}
		if len(desc.Nullifiers[0]) != 32 {
			return false, errors.New("Wrong tx desc's nullifiers")
		}
		if len(desc.Nullifiers[1]) != 32 {
			return false, errors.New("Wrong tx desc's nullifiers")
		}
		// check length of Commitments
		if len(desc.Commitments) != 2 {
			return false, errors.New("Wrong tx desc's commitments")
		}
		if len(desc.Commitments[0]) != 32 {
			return false, errors.New("Wrong tx desc's commitments")
		}
		if len(desc.Commitments[1]) != 32 {
			return false, errors.New("Wrong tx desc's commitments")
		}
		// check length of Vmacs
		if len(desc.Vmacs) != 2 {
			return false, errors.New("Wrong tx desc's vmacs")
		}
		if len(desc.Vmacs[0]) != 32 {
			return false, errors.New("Wrong tx desc's vmacs")
		}
		if len(desc.Vmacs[1]) != 32 {
			return false, errors.New("Wrong tx desc's vmacs")
		}
		//
		if desc.Proof != nil { // no privacy-protocol
			return false, errors.New("Wrong tx desc's proof")
		}
		if len(desc.EncryptedData) != 0 {
			return false, errors.New("Wrong tx desc's encryptedData")
		}
		// check nulltifier is existed in DB
		if desc.Reward != 0 {
			return false, errors.New("Wrong tx desc's reward")
		}
	}
	return true, nil
}

// MaybeAcceptTransaction is the main workhorse for handling insertion of new
// free-standing transactions into a memory pool.  It includes functionality
// such as rejecting duplicate transactions, ensuring transactions follow all
// rules, detecting orphan transactions, and insertion into the memory pool.
//
// If the transaction is an orphan (missing parent transactions), the
// transaction is NOT added to the orphan pool, but each unknown referenced
// parent is returned.  Use ProcessTransaction instead if new orphans should
// be added to the orphan pool.
//
// This function is safe for concurrent access.
func (tp *TxPool) MaybeAcceptTransaction(tx transaction.Transaction) (*common.Hash, *TxDesc, error) {
	tp.mtx.Lock()
	hash, txDesc, err := tp.maybeAcceptTransaction(tx)
	tp.mtx.Unlock()
	return hash, txDesc, err
}

// ValidateTxWithBlockChain - process validation of tx with old data in blockchain
// - check double spend
func (tp *TxPool) ValidateTxWithBlockChain(tx transaction.Transaction, chainID byte) error {
	blockChain := tp.config.BlockChain
	switch tx.GetType() {
	// TODO check tx with current mempool
	case common.TxNormalType:
		{
			// check double spend
			return blockChain.ValidateDoubleSpend(tx, chainID)
		}
	case common.TxRegisterCandidateType:
		{
			// check double spend
			return blockChain.ValidateDoubleSpend(tx, chainID)
		}
	case common.TxSalaryType:
		{
			// TODO
			return nil
		}
	case common.TxCustomTokenType:
		{
			// verify custom token signs
			if !blockChain.VerifyCustomTokenSigns(tx) {
				return errors.New("Custom token signs validation is not passed.")
			}

			// check double spend for constant coin
			return blockChain.ValidateDoubleSpend(tx, chainID)
			// TODO check double spend custom token
		}
	case common.TxActionParamsType:
		{
			// TODO
			return nil
		}
	case common.TxLoanRequest:
		{
			txLoan, ok := tx.(*transaction.TxLoanRequest)
			if !ok {
				return fmt.Errorf("Fail parsing LoanRequest transaction")
			}

			// Check if loan's params are correct
			currentParams := blockChain.BestState[chainID].BestBlock.Header.LoanParams
			if txLoan.Params != currentParams {
				return fmt.Errorf("LoanRequest transaction has incorrect params")
			}

			// Check if loan id is unique across all chains
			// TODO(@0xbunyip): should we check in db/chain or only in best state?
			for chainID, bestState := range blockChain.BestState {
				for _, id := range bestState.LoanIDs {
					if bytes.Equal(txLoan.LoanID, id) {
						return fmt.Errorf("LoanID already existed on chain %d", chainID)
					}
				}
			}
		}
	case common.TxLoanResponse:
		{
			txResponse, ok := tx.(*transaction.TxLoanResponse)
			if !ok {
				return fmt.Errorf("Fail parsing LoanResponse transaction")
			}

			// Check if a loan request with the same id exists on any chain
			txHashes, err := tp.config.DataBase.GetLoanTxs(txResponse.LoanID)
			if err != nil {
				return err
			}
			found := false
			for _, txHash := range txHashes {
				hash := &common.Hash{}
				copy(hash[:], txHash)
				_, _, _, txOld, err := blockChain.GetTransactionByHash(hash)
				if txOld == nil || err != nil {
					return fmt.Errorf("Error finding corresponding loan request")
				}
				switch txOld.GetType() {
				case common.TxLoanResponse:
					{
						return fmt.Errorf("Loan already had response")
					}
				case common.TxLoanRequest:
					{
						_, ok := txOld.(*transaction.TxLoanRequest)
						if !ok {
							return fmt.Errorf("Error parsing loan request tx")
						}
						found = true
					}
				}
			}

			if found == false {
				return fmt.Errorf("Corresponding loan request not found")
			} else {
				return nil
			}
		}
	case common.TxLoanPayment:
		{
			txPayment, ok := tx.(*transaction.TxLoanPayment)
			if !ok {
				return fmt.Errorf("Fail parsing LoanPayment transaction")
			}

			// Check if a loan request with the same id exists on any chain
			txHashes, err := tp.config.DataBase.GetLoanTxs(txPayment.LoanID)
			if err != nil {
				return err
			}
			found := false
			for _, txHash := range txHashes {
				hash := &common.Hash{}
				copy(hash[:], txHash)
				_, _, _, txOld, err := blockChain.GetTransactionByHash(hash)
				if txOld == nil || err != nil {
					return fmt.Errorf("Error finding corresponding loan request")
				}
				switch txOld.GetType() {
				case common.TxLoanResponse:
					{
						found = true
					}
				}
			}

			if found == false {
				return fmt.Errorf("Corresponding loan response not found")
			} else {
				return nil
			}
		}
	case common.TxLoanWithdraw:
		{
			txWithdraw, ok := tx.(*transaction.TxLoanWithdraw)
			if !ok {
				return fmt.Errorf("Fail parsing LoanResponse transaction")
			}

			// Check if a loan response with the same id exists on any chain
			txHashes, err := tp.config.DataBase.GetLoanTxs(txWithdraw.LoanID)
			if err != nil {
				return err
			}
			foundResponse := false
			keyCorrect := false
			for _, txHash := range txHashes {
				hash := &common.Hash{}
				copy(hash[:], txHash)
				_, _, _, txOld, err := blockChain.GetTransactionByHash(hash)
				if txOld == nil || err != nil {
					return fmt.Errorf("Error finding corresponding loan request")
				}
				switch txOld.GetType() {
				case common.TxLoanRequest:
					{
						// Check if key is correct
						txRequest, ok := tx.(*transaction.TxLoanRequest)
						if !ok {
							return fmt.Errorf("Error parsing corresponding loan request")
						}
						h := make([]byte, 32)
						sha3.ShakeSum256(h, txWithdraw.Key)
						if bytes.Equal(h, txRequest.KeyDigest) {
							keyCorrect = true
						}
					}
				case common.TxLoanResponse:
					{
						// Check if loan is accepted
						txResponse, ok := tx.(*transaction.TxLoanResponse)
						if !ok {
							return fmt.Errorf("Error parsing corresponding loan response")
						}
						if txResponse.Response != transaction.Accept {
							foundResponse = true
						}
					}

				}
			}

			if !foundResponse {
				return fmt.Errorf("Corresponding loan response not found")
			} else if !keyCorrect {
				return fmt.Errorf("Provided key is incorrect")
			}
			return nil
		}
	}
	return nil
}

// RemoveTx safe remove transaction for pool
func (tp *TxPool) RemoveTx(tx transaction.Transaction) error {
	tp.mtx.Lock()
	err := tp.removeTx(&tx)
	tp.mtx.Unlock()
	return err
}

// GetTx get transaction info by hash
func (tp *TxPool) GetTx(txHash *common.Hash) (transaction.Transaction, error) {
	tp.mtx.Lock()
	Logger.log.Info(txHash.String())
	txDesc, exists := tp.pool[*txHash]
	tp.mtx.Unlock()
	if exists {
		return txDesc.Desc.Tx, nil
	}

	return nil, fmt.Errorf("transaction is not in the pool")
}

// // MiningDescs returns a slice of mining descriptors for all the transactions
// // in the pool.
func (tp *TxPool) MiningDescs() []*transaction.TxDesc {
	descs := []*transaction.TxDesc{}
	tp.mtx.Lock()
	for _, desc := range tp.pool {
		descs = append(descs, &desc.Desc)
	}
	tp.mtx.Unlock()

	return descs
}

// Count return len of transaction pool
func (tp *TxPool) Count() int {
	count := len(tp.pool)
	return count
}

/*
Sum of all transactions sizes
*/
func (tp *TxPool) Size() uint64 {
	tp.mtx.RLock()
	size := uint64(0)
	for _, tx := range tp.pool {
		// TODO: need to implement size func in each type of transactions
		// https://stackoverflow.com/questions/31496804/how-to-get-the-size-of-struct-and-its-contents-in-bytes-in-golang?rq=1
		size += tx.Desc.Tx.GetTxVirtualSize()
	}
	tp.mtx.RUnlock()

	return size
}

// Get Max fee
func (tp *TxPool) MaxFee() uint64 {
	tp.mtx.RLock()
	fee := uint64(0)
	for _, tx := range tp.pool {
		if tx.Desc.Fee > fee {
			fee = tx.Desc.Fee
		}
	}
	tp.mtx.RUnlock()

	return fee
}

/*
// LastUpdated returns the last time a transaction was added to or
	// removed from the source pool.
*/
func (tp *TxPool) LastUpdated() time.Time {
	return time.Unix(tp.lastUpdated, 0)
}

/*
// HaveTransaction returns whether or not the passed transaction hash
	// exists in the source pool.
*/
func (tp *TxPool) HaveTransaction(hash *common.Hash) bool {
	// Protect concurrent access.
	tp.mtx.RLock()
	haveTx := tp.isTxInPool(hash)
	tp.mtx.RUnlock()

	return haveTx
}

/*
CheckTransactionFee - check fee of tx
*/
func (tp *TxPool) CheckTransactionFee(tx transaction.Transaction) (uint64, error) {
	// Salary transactions have no inputs.
	if tp.config.BlockChain.IsSalaryTx(tx) {
		return 0, nil
	}

	txType := tx.GetType()
	switch txType {
	case common.TxCustomTokenType:
		{
			{
				tx := tx.(*transaction.TxCustomToken)
				err := tp.config.Policy.CheckCustomTokenTransactionFee(tx)
				return tx.Fee, err
			}
		}
	case common.TxNormalType:
		{
			normalTx := tx.(*transaction.Tx)
			err := tp.config.Policy.CheckTransactionFee(normalTx)
			return normalTx.Fee, err
		}
	case common.TxActionParamsType:
		{
			return 0, nil
		}
	case common.TxRegisterCandidateType:
		{
			votingTx := tx.(*transaction.TxRegisterCandidate)
			err := tp.config.Policy.CheckVotingTransactionFee(votingTx)
			return votingTx.Fee, err
		}
	default:
		{
			return 0, errors.New("Wrong tx type")
		}
	}
}

/*
ValidateSanityData - validate sansity data of tx
*/
func (tp *TxPool) ValidateSanityData(tx transaction.Transaction) (bool, error) {
	if tx.GetType() == common.TxNormalType || tx.GetType() == common.TxSalaryType {
		txA := tx.(*transaction.Tx)
		ok, err := tp.validateSanityNormalTxData(txA)
		if !ok {
			return false, err
		}
	} else if tx.GetType() == common.TxActionParamsType {
		txA := tx.(*transaction.ActionParamTx)
		// check Version
		if txA.Version > transaction.TxVersion {
			return false, errors.New("Wrong tx version")
		}
		// check LockTime before now
		if int64(txA.LockTime) > time.Now().Unix() {
			return false, errors.New("Wrong tx lockTime")
		}
		// check Type equal "a"
		if txA.Type != common.TxActionParamsType {
			return false, errors.New("Wrong tx type")
		}
	} else if tx.GetType() == common.TxRegisterCandidateType {
		txA := tx.(*transaction.TxRegisterCandidate)
		ok, err := tp.validateSanityVotingTxData(txA)
		if !ok {
			return false, err
		}
	} else if tx.GetType() == common.TxCustomTokenType {
		// TODO check sanity
		return true, nil
	} else {
		return false, errors.New("Wrong tx type")
	}

	return true, nil
}

/*
List all tx ids in mempool
*/
func (tp *TxPool) ListTxs() []string {
	result := make([]string, 0)
	for _, tx := range tp.pool {
		result = append(result, tx.Desc.Tx.Hash().String())
	}
	return result
}
