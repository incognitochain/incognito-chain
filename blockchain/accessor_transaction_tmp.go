package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/transaction/tx_ver2"
	"github.com/incognitochain/incognito-chain/transaction/utils"
)

func (blockchain *BlockChain) StoreTxByDecoy(txList []metadata.Transaction, shardID byte) error {
	var err error
	db := blockchain.GetShardChainDatabase(shardID)

	for _, tx := range txList {
		if tx.GetVersion() < 2 {
			continue
		}
		if tx.GetType() == common.TxRewardType ||
			tx.GetType() == common.TxReturnStakingType ||
			tx.GetType() == common.TxConversionType {
			continue
		}

		txHash := *tx.Hash()
		tokenID := *tx.GetTokenID()
		Logger.log.Infof("Process StoreTxBySerialNumber for tx %v, tokenID %v\n", txHash.String(), tokenID.String())

		if tokenID.String() != common.PRVIDStr {
			txToken, ok := tx.(transaction.TransactionToken)
			if !ok {
				return fmt.Errorf("cannot parse tx %v to transactionToken", txHash.String())
			}
			if txToken.GetTxTokenData().GetType() == utils.CustomTokenInit {
				continue
			}
			txFee := txToken.GetTxBase()
			txNormal := txToken.GetTxNormal()
			mapPRV := make(map[uint64]uint64)
			mapToken := make(map[uint64]uint64)

			prvSig := new(tx_ver2.SigPubKey)
			err = prvSig.SetBytes(txFee.GetSigPubKey())
			if err != nil {
				Logger.log.Errorf("Parse SigPubKey for PRV error with tx %v, %v: %v", tx.Hash().String(), txFee.GetSigPubKey(), err)
				return err
			}
			for i := range prvSig.Indexes {
				for j := range prvSig.Indexes[i] {
					mapPRV[prvSig.Indexes[i][j].Uint64()] += 1
				}
			}
			for idx, count := range mapPRV {
				err = rawdbv2.StoreTxDecoy(db, common.PRVCoinID, idx, *tx.Hash(), count)
				if err != nil {
					Logger.log.Errorf("StoreTxDecoy with idx %v, tokenID %v, shardID %v, txHash %v returns an error: %v\n", idx, common.PRVCoinID.String(), shardID, txHash.String())
					return err
				}
			}

			if tx.GetType() == common.TxTokenConversionType {
				continue
			}
			tokenSig := new(tx_ver2.SigPubKey)
			err = tokenSig.SetBytes(txNormal.GetSigPubKey())
			if err != nil {
				Logger.log.Errorf("Parse SigPubKey for token error with tx %v, %v: %v", tx.Hash().String(), txNormal.GetSigPubKey(), err)
				return err
			}
			for i := range tokenSig.Indexes {
				for j := range tokenSig.Indexes[i] {
					mapToken[tokenSig.Indexes[i][j].Uint64()] += 1
				}
			}
			for idx, count := range mapToken {
				err = rawdbv2.StoreTxDecoy(db, common.ConfidentialAssetID, idx, *tx.Hash(), count)
				if err != nil {
					Logger.log.Errorf("StoreTxDecoy with idx %v, tokenID %v, shardID %v, txHash %v returns an error: %v\n", idx, common.ConfidentialAssetID.String(), shardID, txHash.String())
					return err
				}
			}
		} else {
			mapPRV := make(map[uint64]uint64)
			prvSig := new(tx_ver2.SigPubKey)
			err = prvSig.SetBytes(tx.GetSigPubKey())
			if err != nil {
				Logger.log.Errorf("Parse SigPubKey for PRV error with tx %v, %v: %v", tx.Hash().String(), tx.GetSigPubKey(), err)
				return err
			}
			for i := range prvSig.Indexes {
				for j := range prvSig.Indexes[i] {
					mapPRV[prvSig.Indexes[i][j].Uint64()] += 1
				}
			}
			for idx, count := range mapPRV {
				err = rawdbv2.StoreTxDecoy(db, common.PRVCoinID, idx, *tx.Hash(), count)
				if err != nil {
					Logger.log.Errorf("StoreTxDecoy with idx %v, tokenID %v, shardID %v, txHash %v returns an error: %v\n", idx, common.PRVCoinID.String(), shardID, txHash.String())
					return err
				}
			}
		}
	}

	Logger.log.Infof("Finish StoreTxByDecoys, #txs: %v!!!\n", len(txList))
	return nil
}