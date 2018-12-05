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
	"github.com/ninjadotorg/constant/wallet"
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

	mtx            sync.RWMutex
	config         Config
	pool           map[common.Hash]*TxDesc
	poolNullifiers map[common.Hash][][]byte
}

/*
Init Txpool from config
*/
func (tp *TxPool) Init(cfg *Config) {
	tp.config = *cfg
	tp.pool = make(map[common.Hash]*TxDesc)
	tp.poolNullifiers = make(map[common.Hash][][]byte)
}

// ----------- transaction.MempoolRetriever's implementation -----------------
func (tp *TxPool) GetPoolNullifiers() map[common.Hash][][]byte {
	return tp.poolNullifiers
}

func (tp *TxPool) GetTxsInMem() map[common.Hash]transaction.TxDesc {
	txsInMem := make(map[common.Hash]transaction.TxDesc)
	for hash, txDesc := range tp.pool {
		txsInMem[hash] = txDesc.Desc
	}
	return txsInMem
}

// ----------- end of transaction.MempoolRetriever's implementation -----------------

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
	tp.poolNullifiers[*tx.Hash()] = txD.Desc.Tx.ListNullifiers()
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
	ok := tx.CheckTxVersion(tp.config.Policy.MaxTxVersion)
	if !ok {
		err := MempoolTxError{}
		err.Init(RejectVersion, errors.New(fmt.Sprintf("%+v's version is invalid", txHash.String())))
		return nil, nil, err
	}

	// check fee of tx
	minFee := tp.config.Policy.BlockChain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams.TxFee
	txFee := tx.GetTxFee()
	ok = tx.CheckTransactionFee(minFee)
	if !ok {
		err := MempoolTxError{}
		err.Init(RejectVersion, errors.New(fmt.Sprintf("transaction %+v has %d fees which is under the required amount of %d", tx.Hash().String(), txFee, minFee)))
		return nil, nil, err
	}
	// end check with policy

	// check tx with all txs in current mempool
	err = tx.ValidateTxWithCurrentMempool(tp)
	if err != nil {
		return nil, nil, err
	}

	// validate tx with data of blockchain
	err = tx.ValidateTxWithBlockChain(tp.config.BlockChain, chainID)
	// err = tp.ValidateTxWithBlockChain(tx, chainID)
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

	// ValidateTransaction tx by it self
	validate := tp.ValidateTxByItSelf(tx)
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

// ValidateTransaction sanity for normal tx data
func (tp *TxPool) validateSanityNormalTxData(tx *transaction.Tx, allowReward bool) (bool, error) {
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
		if desc.Proof == nil && len(desc.Note) == 0 {
			return false, errors.New("Wrong tx desc's proof")
		}
		if desc.Proof != nil {
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
		}
		// check nulltifier is existed in DB
		if desc.Reward != 0 && !allowReward {
			return false, errors.New("Wrong tx desc's reward")
		}
	}
	return true, nil
}

func (tp *TxPool) validateSanityCustomTokenTxData(txCustomToken *transaction.TxCustomToken, allowToUseDCBFund bool) (bool, error) {
	ok, err := tp.validateSanityNormalTxData(&txCustomToken.Tx, allowToUseDCBFund)
	if err != nil || !ok {
		return ok, err
	}
	vins := txCustomToken.TxTokenData.Vins
	zeroHash := common.Hash{}
	for _, vin := range vins {
		if len(vin.PaymentAddress.Pk) == 0 {
			return false, errors.New("Wrong input transaction")
		}
		dbcAccount, _ := wallet.Base58CheckDeserialize(blockchain.DCBAddress)
		if bytes.Equal(vin.PaymentAddress.Pk, dbcAccount.KeySet.PaymentAddress.Pk) {
			if !allowToUseDCBFund {
				return false, errors.New("Cannot use DCB's fund here")
			}
		} else if vin.Signature == "" {
			return false, errors.New("Wrong signature")
		}
		if vin.TxCustomTokenID.String() == zeroHash.String() {
			return false, errors.New("Wrong input transaction")
		}
	}
	vouts := txCustomToken.TxTokenData.Vouts
	for _, vout := range vouts {
		if len(vout.PaymentAddress.Pk) == 0 {
			return false, errors.New("Wrong input transaction")
		}
		if vout.Value == 0 {
			return false, errors.New("Wrong input transaction")
		}
	}
	return true, nil
}

func (tp *TxPool) validateBuySellReqTxSanity(buySellReqTx *transaction.TxBuySellRequest, allowToUseDCBFund bool) (bool, error) {
	ok, err := tp.validateSanityNormalTxData(&buySellReqTx.Tx, allowToUseDCBFund)
	if err != nil || !ok {
		return ok, err
	}
	if len(buySellReqTx.PaymentAddress.Pk) == 0 {
		return false, errors.New("Wrong request info's payment address")
	}
	if buySellReqTx.BuyPrice == 0 {
		return false, errors.New("Wrong request info's buy price")
	}
	if buySellReqTx.Amount == 0 {
		return false, errors.New("Wrong request info's amount")
	}
	if len(buySellReqTx.AssetType) != common.HashSize {
		return false, errors.New("Wrong request info's asset type")
	}
	return true, nil
}

func (tp *TxPool) validateBuyBackReqTxSanity(buyBackRequestTx *transaction.TxBuyBackRequest, allowToUseDCBFund bool) (bool, error) {
	ok, err := tp.validateSanityNormalTxData(buyBackRequestTx.Tx, allowToUseDCBFund)
	if err != nil || !ok {
		return ok, err
	}
	if buyBackRequestTx.VoutIndex < 0 {
		return false, errors.New("Wrong request info's vout index")
	}
	if len(buyBackRequestTx.BuyBackFromTxID) == 0 {
		return false, errors.New("Wrong request info's BuyBackFromTxID")
	}
	return true, nil
}

func (tp *TxPool) validateSanityVoteDCBBoardTx(voteDCBBoard *transaction.TxVoteDCBBoard) (bool, error) {
	ok, err := tp.validateSanityCustomTokenTxData(&voteDCBBoard.TxCustomToken, false)
	if err != nil || !ok {
		return ok, err
	}
	if len(voteDCBBoard.VoteDCBBoardData.CandidatePubKey) != 33 {
		return false, nil
	}
	return true, nil
}

func (tp *TxPool) validateSanityVoteGOVBoardTx(voteGOVBoard *transaction.TxVoteGOVBoard) (bool, error) {
	ok, err := tp.validateSanityCustomTokenTxData(&voteGOVBoard.TxCustomToken, false)
	if err != nil || !ok {
		return ok, err
	}
	if len(voteGOVBoard.VoteGOVBoardData.CandidatePubKey) != 33 {
		return false, nil
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

// ValidateDoubleSpendTxWithCurrentMempool - check double spend for new tx with all txs in mempool
func (tp *TxPool) ValidateDoubleSpendTxWithCurrentMempool(txNormal transaction.Tx) error {
	for _, temp1 := range tp.poolNullifiers {
		for _, desc := range txNormal.Descs {
			for _, nullifier := range desc.Nullifiers {
				if ok, err := common.SliceBytesExists(temp1, nullifier); !ok || err != nil {
					return errors.New("Double spend")
				}
			}
		}
	}
	return nil
}

// ValidateTxWithCurrentMempool - check new tx with all txs in mempool
func (tp *TxPool) ValidateTxWithCurrentMempool(tx transaction.Transaction) error {
	switch tx.GetType() {
	case common.TxNormalType:
		{
			txNormal := tx.(*transaction.Tx)
			err := tp.ValidateDoubleSpendTxWithCurrentMempool(*txNormal)
			if err != nil {
				return err
			}
			return nil
		}
	case common.TxSalaryType:
		{
			return errors.New("Can not receive a salary tx from other node, this is a violation")
		}
	case common.TxCustomTokenType:
		{
			err := tp.validateTxCustomTokenInPool(tx)
			return err
		}
	case common.TxVoteDCBBoard:
		{
			txVoteDCBBoard := tx.(*transaction.TxVoteDCBBoard)
			txCustomToKen := txVoteDCBBoard.TxCustomToken
			err := tp.validateTxCustomTokenInPool(&txCustomToKen)
			if err != nil {
				return err
			}
			return nil
		}
	case common.TxVoteGOVBoard:
		{
			txVoteGOVBoard := tx.(*transaction.TxVoteGOVBoard)
			txCustomToKen := txVoteGOVBoard.TxCustomToken
			err := tp.validateTxCustomTokenInPool(&txCustomToKen)
			if err != nil {
				return err
			}
			return nil
		}
	case common.TxBuyFromGOVRequest:
		{
			reqTx := tx.(*transaction.TxBuySellRequest)
			normalTx := reqTx.Tx
			err := tp.ValidateDoubleSpendTxWithCurrentMempool(normalTx)
			if err != nil {
				return err
			}
			return nil
		}
	case common.TxBuyBackRequest:
		{
			reqTx := tx.(*transaction.TxBuyBackRequest)
			normalTx := reqTx.Tx
			err := tp.ValidateDoubleSpendTxWithCurrentMempool(*normalTx)
			if err != nil {
				return err
			}
			return nil
		}
	case common.TxBuySellDCBRequest:
		{
			txRequest := tx.(*transaction.TxBuySellRequest)
			return tp.validateTxCustomTokenInPool(txRequest.TxCustomToken)
		}
	case common.TxBuySellDCBResponse:
		{
			txResponse := tx.(*transaction.TxBuySellDCBResponse)
			return tp.validateTxCustomTokenInPool(txResponse.TxCustomToken)
		}
	default:
		{
			return errors.New("Wrong tx type")
		}
	}
	return errors.New("No check tx")
}

func (tp *TxPool) validateTxCustomTokenInPool(tx transaction.Transaction) error {
	txCustomToken := tx.(*transaction.TxCustomToken)
	txNormal := txCustomToken.Tx
	err := tp.ValidateDoubleSpendTxWithCurrentMempool(txNormal)
	if err != nil {
		return err
	}
	txsInMem := tp.pool
	for _, txInMem := range txsInMem {
		err := tp.config.BlockChain.ValidateDoubleSpendCustomTokenOnTx(txCustomToken, txInMem.Desc.Tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tp *TxPool) ValidateTxCustomTokenBlockChain(tx transaction.Transaction, chainID byte) error {
	blockChain := tp.config.BlockChain
	if !blockChain.VerifyCustomTokenSigns(tx) {
		return errors.New("Custom token signs validation is not passed.")
	}

	// check double spend for constant coin with blockchain
	err := blockChain.ValidateDoubleSpend(tx, chainID)
	if err != nil {
		return err
	}
	// check double spend for custom token with blockchain data
	err = blockChain.ValidateDoubleSpendCustomToken(tx.(*transaction.TxCustomToken))
	if err != nil {
		return err
	}
	return nil
}

// ValidateTxWithBlockChain - process validation of tx with old data in blockchain
// - check double spend
func (tp *TxPool) ValidateTxWithBlockChain(tx transaction.Transaction, chainID byte) error {
	blockChain := tp.config.BlockChain
	switch tx.GetType() {
	case common.TxNormalType:
		{
			// check double spend
			err := blockChain.ValidateDoubleSpend(tx, chainID)
			if err != nil {
				return err
			}
			return nil
		}
	case common.TxSalaryType:
		{
			//return errors.New("Can not receive a salary tx from other node, this is a violation")
			return nil
		}
	case common.TxCustomTokenType:
		{
			return tp.ValidateTxCustomTokenBlockChain(tx, chainID)
			// verify custom token signs
		}
	case common.TxVoteDCBBoard:
		{
			txVoteDCBBoard := tx.(*transaction.TxVoteDCBBoard)
			txCustomToken := txVoteDCBBoard.TxCustomToken
			return tp.ValidateTxCustomTokenBlockChain(&txCustomToken, chainID)
		}
	case common.TxVoteGOVBoard:
		{
			txVoteGOVBoard := tx.(*transaction.TxVoteGOVBoard)
			txCustomToken := txVoteGOVBoard.TxCustomToken
			return tp.ValidateTxCustomTokenBlockChain(&txCustomToken, chainID)
		}
	case common.TxLoanRequest:
		{
			return blockChain.ValidateTxLoanRequest(tx, chainID)
		}
	case common.TxLoanResponse:
		{
			return blockChain.ValidateTxLoanResponse(tx, chainID)
		}
	case common.TxLoanPayment:
		{
			return blockChain.ValidateTxLoanPayment(tx, chainID)
		}
	case common.TxLoanWithdraw:
		{
			return blockChain.ValidateTxLoanWithdraw(tx, chainID)
		}
	case common.TxDividendPayout:
		{
			return blockChain.ValidateTxDividendPayout(tx, chainID)
		}
	case common.TxBuySellDCBRequest:
		{
			return blockChain.ValidateTxBuySellDCBRequest(tx, chainID)
		}
	case common.TxBuySellDCBResponse:
		{
			return blockChain.ValidateTxBuySellDCBResponse(tx, chainID)
		}
	case common.TxSubmitDCBProposal:
		{
			return blockChain.ValidateTxSubmitDCBProposal(tx, chainID)
		}
	case common.TxAcceptDCBProposal:
		{
			return blockChain.ValidateTxAcceptDCBProposal(tx, chainID)
		}
	case common.TxVoteDCBProposal:
		{
			return blockChain.ValidateTxVoteDCBProposal(tx, chainID)
		}
	case common.TxSubmitGOVProposal:
		{
			return blockChain.ValidateTxSubmitGOVProposal(tx, chainID)
		}
	case common.TxAcceptGOVProposal:
		{
			return blockChain.ValidateTxAcceptGOVProposal(tx, chainID)
		}
	case common.TxVoteGOVProposal:
		{
			return blockChain.ValidateTxVoteGOVProposal(tx, chainID)
		}
	case common.TxBuyFromGOVRequest:
		{
			return blockChain.ValidateBuyFromGOVRequestTx(tx, chainID)
		}
	case common.TxBuyBackRequest:
		{
			return blockChain.ValidateBuyBackRequestTx(tx, chainID)
		}
	default:
		{
			return errors.New("Wrong tx type")
		}
	}
	return errors.New("No check Tx")
}

// GetListUTXOFromTxCustomToken - get list utxo with related to vins of a TxCustomToken
// from vins list of custom token
// loop and get all older tx custom token which are contained in vins list
// return map[hash]{instant of custom token tx} for vin of given TxCustomToken
func (tp *TxPool) GetListUTXOFromTxCustomToken(txCustomToken *transaction.TxCustomToken) bool {
	data := make(map[common.Hash]transaction.TxCustomToken)
	for _, vin := range txCustomToken.TxTokenData.Vins {
		_, _, _, utxo, err := tp.config.BlockChain.GetTransactionByHash(&vin.TxCustomTokenID)
		if err != nil {
			Logger.log.Error(err)
			return false
		}
		data[vin.TxCustomTokenID] = *(utxo.(*transaction.TxCustomToken))
	}
	if len(data) == 0 {
		Logger.log.Error(errors.New("Can not find any utxo for TxCustomToken"))
		return false
	}
	txCustomToken.SetListUtxo(data)
	return true
}

// ValidateTxByItSelf - Each of Tx instance should be validated by it self

func (tp *TxPool) ValidateTxByItSelf(tx transaction.Transaction) bool {
	switch tx.GetType() {
	case common.TxCustomTokenType:
		{
			// with custom token tx, we need to get utxo for custom token and for validation
			txCustomToken := tx.(*transaction.TxCustomToken)
			ok := tp.GetListUTXOFromTxCustomToken(txCustomToken)
			if ok == false {
				return false
			}
			return txCustomToken.ValidateTransaction()
		}
	case common.TxVoteDCBBoard:
		{
			txVoteDCBBoard := tx.(*transaction.TxVoteDCBBoard)
			txCustomToken := txVoteDCBBoard.TxCustomToken
			ok := tp.GetListUTXOFromTxCustomToken(&txCustomToken)
			if ok == false {
				return false
			}
			return txCustomToken.ValidateTransaction() && txVoteDCBBoard.Validate()
		}
	case common.TxVoteGOVBoard:
		{
			txVoteGOVBoard := tx.(*transaction.TxVoteGOVBoard)
			txCustomToken := txVoteGOVBoard.TxCustomToken
			ok := tp.GetListUTXOFromTxCustomToken(&txCustomToken)
			if ok == false {
				return false
			}
			return txCustomToken.ValidateTransaction() && txVoteGOVBoard.Validate()
		}

	default:
		return tx.ValidateTransaction()
	}
	return false
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
	case common.TxBuySellDCBRequest:
		{
			tx := tx.(*transaction.TxBuySellRequest)
			err := tp.config.Policy.CheckCustomTokenTransactionFee(tx.TxCustomToken)
			return tx.TxCustomToken.Fee, err
		}
	case common.TxBuySellDCBResponse:
		{
			return 0, nil // Response tx for sale has no fee
		}
	default:
		{
			return 0, errors.New("Wrong tx type")
		}
	}
	return 0, errors.New("No check tx")
}

/*
ValidateSanityData - validate sansity data of tx
*/
func (tp *TxPool) ValidateSanityData(tx transaction.Transaction) (bool, error) {
	switch tx.GetType() {
	case common.TxNormalType, common.TxSalaryType:
		{
			txA := tx.(*transaction.Tx)
			ok, err := tp.validateSanityNormalTxData(txA, false)
			return ok, err
		}
	case common.TxCustomTokenType:
		{
			txCustomToken := tx.(*transaction.TxCustomToken)
			ok, err := tp.validateSanityCustomTokenTxData(txCustomToken, false)
			return ok, err
		}
	case common.TxVoteDCBBoard:
		{
			txVoteDCBBoard := tx.(*transaction.TxVoteDCBBoard)
			ok, err := tp.validateSanityVoteDCBBoardTx(txVoteDCBBoard)
			return ok, err
		}
	case common.TxVoteGOVBoard:
		{
			txVoteGOVBoard := tx.(*transaction.TxVoteGOVBoard)
			ok, err := tp.validateSanityVoteGOVBoardTx(txVoteGOVBoard)
			return ok, err
		}
	case common.TxBuyFromGOVRequest:
		{
			buySellReqTx := tx.(*transaction.TxBuySellRequest)
			ok, err := tp.validateBuySellReqTxSanity(buySellReqTx, false)
			return ok, err
		}
	case common.TxBuyBackRequest:
		{
			buyBackReqTx := tx.(*transaction.TxBuyBackRequest)
			ok, err := tp.validateBuyBackReqTxSanity(buyBackReqTx, false)
			return ok, err
		}
	case common.TxBuySellDCBRequest:
		{
			txRequest := tx.(*transaction.TxBuySellRequest)
			return tp.validateSanityCustomTokenTxData(txRequest.TxCustomToken, false)
		}
	case common.TxBuySellDCBResponse:
		{
			txResponse := tx.(*transaction.TxBuySellDCBResponse)
			return tp.validateSanityCustomTokenTxData(txResponse.TxCustomToken, true)
		}
	default:
		{
			return false, errors.New("Wrong tx type")
		}
	}
	return false, errors.New("No check tx")
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
