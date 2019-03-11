package blockchain

// import (
// 	"bytes"

// 	"github.com/big0t/constant-chain/common"
// 	"github.com/big0t/constant-chain/metadata"
// 	"github.com/big0t/constant-chain/privacy"
// 	"github.com/big0t/constant-chain/transaction"
// 	"github.com/big0t/constant-chain/wallet"
// 	"github.com/pkg/errors"
// )

// func (blockgen *BlkTmplGenerator) buildCMBRefund(
// 	sourceTxns []*metadata.TxDesc,
// 	shardID byte,
// 	producerPrivateKey *privacy.SpendingKey,
// ) ([]*transaction.Tx, error) {
// 	// Get old block
// 	refunds := []*transaction.Tx{}
// 	header := blockgen.chain.BestState[shardID].BestBlock.Header
// 	lookBackBlockHeight := header.Height - int32(metadata.CMBInitRefundPeriod)
// 	if lookBackBlockHeight < 0 {
// 		return refunds, nil
// 	}
// 	lookBackBlock, err := blockgen.chain.GetBlockByBlockHeight(lookBackBlockHeight, shardID)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return refunds, nil
// 	}

// 	// Build refund tx for each cmb init request that hasn't been approved
// 	for _, tx := range lookBackBlock.Transactions {
// 		metaDataType := tx.GetMetadataType()
// 		Logger.log.Infof("buildCMBRefund: Metadata type of tx: %d", metaDataType)
// 		if metaDataType != metadata.CMBInitRequestMeta {
// 			continue
// 		}

// 		// Check if CMB is still not approved in previous blocks
// 		meta := tx.GetMetadata().(*metadata.CMBInitRequest)
// 		_, _, capital, _, state, _, err := blockgen.chain.GetCMB(meta.MainAccount.Bytes())
// 		if err != nil {
// 			Logger.log.Error(errors.Wrap(err, "buildCMBRefund"))
// 			// Unexpected error, cannot create a block if CMB init request is not refundable
// 			return nil, errors.Errorf("error retrieving cmb for building refund")
// 		}
// 		if state == metadata.CMBRequested {
// 			refund, err := transaction.BuildRefundTx(meta.MainAccount, capital, producerPrivateKey, blockgen.chain.GetDatabase())
// 			if err != nil {
// 				Logger.log.Error(errors.Wrap(err, "buildCMBRefund"))
// 				return nil, errors.Errorf("error building refund tx for cmb init request")
// 			}
// 			refunds = append(refunds, refund)
// 		}
// 	}
// 	return refunds, nil
// }

// func (bc *BlockChain) processCMBInitRequest(tx metadata.Transaction) error {
// 	meta := tx.GetMetadata().(*metadata.CMBInitRequest)

// 	// Members of the CMB
// 	members := [][]byte{}
// 	for _, member := range meta.Members {
// 		members = append(members, member.Bytes())
// 	}

// 	// Capital of the CMB
// 	txPrivacy := tx.(*transaction.Tx)
// 	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
// 	dcbPk := keyWalletDCBAccount.KeySet.PaymentAddress.Pk
// 	capital := uint64(0)
// 	for _, coin := range txPrivacy.Proof.OutputCoins {
// 		if bytes.Equal(coin.CoinDetails.PublicKey.Compress(), dcbPk) {
// 			capital += coin.CoinDetails.Value
// 		}
// 	}

// 	// Store in DB
// 	txHash := tx.Hash()
// 	err := bc.config.DataBase.StoreCMB(meta.MainAccount.Bytes(), meta.ReserveAccount.Bytes(), members, capital, txHash[:])
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (bc *BlockChain) processCMBInitResponse(tx metadata.Transaction) error {
// 	// Store board member who approved this cmb init request
// 	meta := tx.GetMetadata().(*metadata.CMBInitResponse)
// 	sender := tx.GetSigPubKey()
// 	err := bc.config.DataBase.StoreCMBResponse(meta.MainAccount.Bytes(), sender)
// 	if err != nil {
// 		return err
// 	}

// 	// Update state of CMB if enough DCB board governors approved
// 	approvers, _ := bc.config.DataBase.GetCMBResponse(meta.MainAccount.Bytes())
// 	minApproval := bc.GetDCBParams().MinCMBApprovalRequire
// 	if len(approvers) == int(minApproval) {
// 		err := bc.config.DataBase.UpdateCMBState(meta.MainAccount.Bytes(), metadata.CMBApproved)
// 		if err != nil {
// 			return err
// 		}
// 		return nil
// 	}
// 	return nil
// }

// func (bc *BlockChain) processCMBInitRefund(tx metadata.Transaction) error {
// 	meta := tx.GetMetadata().(*metadata.CMBInitRefund)
// 	err := bc.config.DataBase.UpdateCMBState(meta.MainAccount.Bytes(), metadata.CMBRefunded)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (bc *BlockChain) processCMBDepositSend(tx metadata.Transaction) error {
// 	meta := tx.GetMetadata().(*metadata.CMBDepositSend)
// 	hash := tx.Hash()
// 	err := bc.config.DataBase.StoreDepositSend(meta.ContractID[:], hash[:])
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (bc *BlockChain) processCMBWithdrawRequest(tx metadata.Transaction) error {
// 	// Store request in db to prevent another request for this contract
// 	meta := tx.GetMetadata().(*metadata.CMBWithdrawRequest)
// 	hash := tx.Hash()
// 	err := bc.config.DataBase.StoreWithdrawRequest(meta.ContractID[:], hash[:])
// 	if err != nil {
// 		return err
// 	}

// 	// Add notice period for later lateness check
// 	_, _, _, txContract, err := bc.GetTransactionByHash(&meta.ContractID)
// 	contractMeta := txContract.GetMetadata().(*metadata.CMBDepositContract)
// 	height, _ := bc.GetTxChainHeight(tx)
// 	endBlock := height + contractMeta.NoticePeriod
// 	err = bc.config.DataBase.StoreNoticePeriod(endBlock, hash[:])
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (bc *BlockChain) processCMBWithdrawResponse(tx metadata.Transaction) error {
// 	// Update state of withdraw request
// 	meta := tx.GetMetadata().(*metadata.CMBWithdrawResponse)
// 	_, _, _, txRequest, err0 := bc.GetTransactionByHash(&meta.RequestTxID)
// 	if err0 != nil {
// 		return err0
// 	}
// 	metaReq := txRequest.GetMetadata().(*metadata.CMBWithdrawRequest)
// 	state := metadata.WithdrawFulfilled
// 	err := bc.config.DataBase.UpdateWithdrawRequestState(metaReq.ContractID[:], state)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (bc *BlockChain) findLateWithdrawResponse(blockHeight uint64) error {
// 	txHashes, err := bc.config.DataBase.GetNoticePeriod(blockHeight)
// 	if err != nil {
// 		return err
// 	}
// 	for _, txHash := range txHashes {
// 		// Get request tx
// 		hash, _ := (&common.Hash{}).NewHash(txHash)
// 		_, _, _, txReq, err0 := bc.GetTransactionByHash(hash)
// 		if err0 != nil {
// 			return err0
// 		}
// 		reqMeta := txReq.GetMetadata().(*metadata.CMBWithdrawRequest)

// 		// Check if request is still not fulfilled
// 		_, state, err := bc.GetWithdrawRequest(reqMeta.ContractID[:])
// 		if err != nil {
// 			return err
// 		}
// 		if state == metadata.WithdrawFulfilled {
// 			continue
// 		}

// 		// Get deposit contract of the request
// 		contIDHash, _ := (&common.Hash{}).NewHash(reqMeta.ContractID[:])
// 		_, _, _, txCont, err := bc.GetTransactionByHash(contIDHash)
// 		if err != nil {
// 			return err
// 		}
// 		contractMeta := txCont.GetMetadata().(*metadata.CMBDepositContract)

// 		// Update fine
// 		_, _, _, _, _, fine, err := bc.config.DataBase.GetCMB(contractMeta.CMBAddress.Bytes())
// 		fine += bc.GetDCBParams().LateWithdrawResponseFine
// 		err = bc.config.DataBase.UpdateCMBFine(contractMeta.CMBAddress.Bytes(), fine)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
