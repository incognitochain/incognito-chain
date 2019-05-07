package mempool

import (
	"errors"
	"fmt"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/transaction"
)
/*
	Verify Transaction with these condition:
	1. Validate with current mempool:
	2. Validate Init Custom Token
	3. Check tx existed in block
	4. Check duplicate staker public key in block
	5. Check duplicate Init Custom Token in block
*/
func (tp *TxPool) ValidateTxList(txs []metadata.Transaction) error {
	var errCh chan error
	errCh = make(chan error)
	validTxCount := 0
	// salaryTxCount := 0
	//validate individual tx
	go func() {
		for _, tx := range txs {
			go func(tx metadata.Transaction) {
				if tx.GetType() == common.TxCustomTokenType {
					customTokenTx := tx.(*transaction.TxCustomToken)
						if customTokenTx.TxTokenData.Type == transaction.CustomTokenCrossShard {
							errCh <- nil
							return
							}
					}
				err := tp.validateTxIndependProperties(tx)
				errCh <- err
			}(tx)
		}
	}()

	for {
		err := <-errCh
		if err != nil {
			return errors.New("tx in new block error:" + err.Error())
		}
		validTxCount++
		if validTxCount == len(txs) {
			break
		}
	}
	//validate txs list
	for _, tx := range txs {
		txHash := tx.Hash()
		// Don't accept the transaction if it already exists in the pool.
		if tp.isTxInPool(txHash) {
			str := fmt.Sprintf("already have transaction %+v", txHash.String())
			err := MempoolTxError{}
			err.Init(RejectDuplicateTx, errors.New(str))
			return err
		}

		// check tx with all txs in current mempool
		err := tx.ValidateTxWithCurrentMempool(tp)
		if err != nil {
			return err
		}
		if tx.GetType() == common.TxCustomTokenType {
			customTokenTx := tx.(*transaction.TxCustomToken)
			if customTokenTx.TxTokenData.Type == transaction.CustomTokenInit {
				tokenID := customTokenTx.TxTokenData.PropertyID.String()
				tp.tokenIDMtx.Lock()
				found := common.IndexOfStrInHashMap(tokenID, tp.TokenIDPool)
				tp.tokenIDMtx.Unlock()
				if found > 0 {
					str := fmt.Sprintf("Init Transaction of this Token is in pool already %+v", tokenID)
					err := MempoolTxError{}
					err.Init(RejectDuplicateInitTokenTx, errors.New(str))
					return err
				}
			}
		}

		// check duplicate stake public key ONLY with staking transaction
		if tx.GetMetadata() != nil {
			if tx.GetMetadata().GetType() == metadata.ShardStakingMeta || tx.GetMetadata().GetType() == metadata.BeaconStakingMeta {
				pubkey := base58.Base58Check{}.Encode(tx.GetSigPubKey(), common.ZeroByte)
				tp.tokenIDMtx.Lock()
				found := common.IndexOfStrInHashMap(pubkey, tp.CandidatePool)
				tp.tokenIDMtx.Unlock()
				if found > 0 {
					str := fmt.Sprintf("This public key already stake and still in pool %+v", pubkey)
					err := MempoolTxError{}
					err.Init(RejectDuplicateStakeTx, errors.New(str))
					return err
				}
			}
		}

		shardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		bestHeight := tp.config.BlockChain.BestState.Shard[shardID].BestBlock.Header.Height
		txFee := tx.GetTxFee()
		txD := createTxDescMempool(tx, bestHeight, txFee)
		tp.addTx(txD, false)
	}

	return nil
}
/*
SKIP salary transaction
Verify Transaction with these condition:
	1. Validate tx version
	2. Validate fee with tx size
	3. Validate type of tx
	4. Validate sanity data of tx
	5. Validate By it self (data in tx): privacy proof, metadata,...
	6. Validate tx with blockchain: douple spend, ...
*/
func (tp *TxPool) validateTxIndependProperties(tx metadata.Transaction) error {
	var shardID byte
	var err error
	txHash := tx.Hash()

	if tx.IsSalaryTx() {
		return nil
	}
	// check version
	ok := tx.CheckTxVersion(MaxVersion)
	if !ok {
		err := MempoolTxError{}
		err.Init(RejectVersion, fmt.Errorf("transaction %+v's version is invalid", txHash.String()))
		return err
	}

	// check actual size
	actualSize := tx.GetTxActualSize()
	Logger.log.Debugf("Transaction %+v 's size %+v \n", *txHash, actualSize)
	if actualSize >= common.MaxBlockSize || actualSize >= common.MaxTxSize {
		err := MempoolTxError{}
		err.Init(RejectInvalidSize, fmt.Errorf("transaction %+v's size is invalid, more than %+v Kilobyte", txHash.String(), common.MaxBlockSize))
		return err
	}

	// check fee of tx
	minFeePerKbTx := tp.config.BlockChain.GetFeePerKbTx()
	txFee := tx.GetTxFee()
	ok = tx.CheckTransactionFee(minFeePerKbTx)
	if !ok {
		err := MempoolTxError{}
		err.Init(RejectInvalidFee, fmt.Errorf("transaction %+v has %d fees which is under the required amount of %d", tx.Hash().String(), txFee, minFeePerKbTx*tx.GetTxActualSize()))
		return err
	}
	// end check with policy

	ok = tx.ValidateType()
	if !ok {
		return err
	}

	// sanity data
	if validated, errS := tx.ValidateSanityData(tp.config.BlockChain); !validated {
		err := MempoolTxError{}
		err.Init(RejectSansityTx, fmt.Errorf("transaction's sansity %v is error %v", txHash.String(), errS.Error()))
		return err
	}

	// ValidateTransaction tx by it self
	shardID = common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	validated := tx.ValidateTxByItself(tx.IsPrivacy(), tp.config.BlockChain.GetDatabase(), tp.config.BlockChain, shardID)
	if !validated {
		err := MempoolTxError{}
		err.Init(RejectInvalidTx, errors.New("invalid tx"))
		return err
	}
	// validate tx with data of blockchain
	err = tx.ValidateTxWithBlockChain(tp.config.BlockChain, shardID, tp.config.BlockChain.GetDatabase())
	if err != nil {
		return err
	}
	return nil
}
