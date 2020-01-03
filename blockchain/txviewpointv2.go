package blockchain

import (
	"errors"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"github.com/incognitochain/incognito-chain/transaction"
)

// fetch from desc of tx to get serialNumber and commitments
// (note: still storage full data of commitments, serialnumbers, snderivator to check double spend)
func (view *TxViewPoint) processFetchTxViewPointV2(
	stateDB *statedb.StateDB,
	shardID byte,
	proof *zkp.PaymentProof,
	tokenID *common.Hash,
) ([][]byte, map[string][][]byte, map[string][]privacy.OutputCoin, map[string][]privacy.Scalar, error) {
	acceptedSerialNumbers := make([][]byte, 0)
	acceptedCommitments := make(map[string][][]byte)
	acceptedOutputcoins := make(map[string][]privacy.OutputCoin)
	acceptedSnD := make(map[string][]privacy.Scalar)
	if proof == nil {
		return acceptedSerialNumbers, acceptedCommitments, acceptedOutputcoins, acceptedSnD, nil
	}
	// Get data for serialnumbers
	// Process input of transaction
	// Get Serial numbers of input
	// Append into accepttedSerialNumbers if this serial number haven't exist yet
	for _, item := range proof.GetInputCoins() {
		serialNum := item.CoinDetails.GetSerialNumber().ToBytesS()
		ok, err := statedb.HasSerialNumber(stateDB, *tokenID, serialNum, shardID)
		if err != nil {
			return acceptedSerialNumbers, acceptedCommitments, acceptedOutputcoins, acceptedSnD, err
		}
		if !ok {
			acceptedSerialNumbers = append(acceptedSerialNumbers, serialNum)
		}
	}

	// Process Output Coins (just created UTXO of this transaction)
	// Proccessed variable: commitment, snd, outputcoins
	// Commitment and SND must not exist before in db
	// Outputcoins will be stored as new utxo for next transaction
	for _, item := range proof.GetOutputCoins() {
		commitment := item.CoinDetails.GetCoinCommitment().ToBytesS()
		pubkey := item.CoinDetails.GetPublicKey().ToBytesS()
		pubkeyStr := base58.Base58Check{}.Encode(pubkey, common.ZeroByte)
		ok, err := statedb.HasCommitment(stateDB, *tokenID, commitment, shardID)
		if err != nil {
			return acceptedSerialNumbers, acceptedCommitments, acceptedOutputcoins, acceptedSnD, err
		}
		if !ok {
			publicKeyStr := base58.Base58Check{}.Encode(pubkey, common.ZeroByte)
			if acceptedCommitments[publicKeyStr] == nil {
				acceptedCommitments[publicKeyStr] = make([][]byte, 0)
			}
			// get data for commitments
			acceptedCommitments[publicKeyStr] = append(acceptedCommitments[publicKeyStr], item.CoinDetails.GetCoinCommitment().ToBytesS())

			// get data for output coin
			if acceptedOutputcoins[publicKeyStr] == nil {
				acceptedOutputcoins[publicKeyStr] = make([]privacy.OutputCoin, 0)
			}
			acceptedOutputcoins[publicKeyStr] = append(acceptedOutputcoins[publicKeyStr], *item)
		}

		// get data for Snderivators
		snD := item.CoinDetails.GetSNDerivator()
		ok, err = statedb.HasSNDerivator(stateDB, *tokenID, snD.ToBytesS())
		if !ok && err == nil {
			acceptedSnD[pubkeyStr] = append(acceptedSnD[pubkeyStr], *snD)
		}
	}
	return acceptedSerialNumbers, acceptedCommitments, acceptedOutputcoins, acceptedSnD, nil
}

/*
fetchTxViewPointFromBlock get list serialnumber and commitments, output coins from txs in block and check if they are not in Main chain db
return a tx view point which contains list new serialNumbers and new commitments from block
// (note: still storage full data of commitments, serialnumbers, snderivator to check double spend)
*/

func (view *TxViewPoint) fetchTxViewPointFromBlockV2(stateDB *statedb.StateDB, block *ShardBlock) error {
	transactions := block.Body.Transactions
	// Loop through all of the transaction descs (except for the salary tx)
	acceptedSerialNumbers := make([][]byte, 0)
	acceptedCommitments := make(map[string][][]byte)
	acceptedOutputcoins := make(map[string][]privacy.OutputCoin)
	acceptedSnD := make(map[string][][]byte)
	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	for indexTx, tx := range transactions {
		switch tx.GetType() {
		case common.TxNormalType, common.TxRewardType, common.TxReturnStakingType:
			{
				normalTx := tx.(*transaction.Tx)
				serialNumbers, commitments, outCoins, snDs, err := view.processFetchTxViewPointV2(stateDB, block.Header.ShardID, normalTx.Proof, prvCoinID)
				if err != nil {
					return NewBlockChainError(UnExpectedError, err)
				}
				acceptedSerialNumbers = append(acceptedSerialNumbers, serialNumbers...)
				for pubkey, data := range commitments {
					if acceptedCommitments[pubkey] == nil {
						acceptedCommitments[pubkey] = make([][]byte, 0)
					}
					acceptedCommitments[pubkey] = append(acceptedCommitments[pubkey], data...)
					view.txByPubKey[pubkey+"_"+base58.Base58Check{}.Encode(tx.Hash().GetBytes(), 0x0)+"_"+strconv.Itoa(int(block.Header.ShardID))] = true
				}
				for pubkey, data := range outCoins {
					if acceptedOutputcoins[pubkey] == nil {
						acceptedOutputcoins[pubkey] = make([]privacy.OutputCoin, 0)
					}
					acceptedOutputcoins[pubkey] = append(acceptedOutputcoins[pubkey], data...)
				}
				for pubkey, data := range snDs {
					if acceptedSnD[pubkey] == nil {
						acceptedSnD[pubkey] = make([][]byte, 0)
					}
					for _, snd := range data {
						acceptedSnD[pubkey] = append(acceptedSnD[pubkey], snd.ToBytesS())
					}
				}
			}
		case common.TxCustomTokenPrivacyType:
			{
				tx := tx.(*transaction.TxCustomTokenPrivacy)
				serialNumbers, commitments, outCoins, snDs, err := view.processFetchTxViewPointV2(stateDB, block.Header.ShardID, tx.Proof, prvCoinID)
				if err != nil {
					return NewBlockChainError(UnExpectedError, err)
				}
				acceptedSerialNumbers = append(acceptedSerialNumbers, serialNumbers...)
				for pubkey, data := range commitments {
					if acceptedCommitments[pubkey] == nil {
						acceptedCommitments[pubkey] = make([][]byte, 0)
					}
					acceptedCommitments[pubkey] = append(acceptedCommitments[pubkey], data...)
					view.txByPubKey[pubkey+"_"+base58.Base58Check{}.Encode(tx.Hash().GetBytes(), 0x0)+"_"+strconv.Itoa(int(block.Header.ShardID))] = true
				}
				for pubkey, data := range outCoins {
					if acceptedOutputcoins[pubkey] == nil {
						acceptedOutputcoins[pubkey] = make([]privacy.OutputCoin, 0)
					}
					acceptedOutputcoins[pubkey] = append(acceptedOutputcoins[pubkey], data...)
				}
				for pubkey, data := range snDs {
					if snDs[pubkey] == nil {
						snDs[pubkey] = make([]privacy.Scalar, 0)
					}
					snDs[pubkey] = append(snDs[pubkey], data...)
				}
				// acceptedSnD = append(acceptedSnD, snDs...)
				/*if err != nil {
					return NewBlockChainError(UnExpectedError, err)
				}*/

				// sub view for privacy custom token
				subView := NewTxViewPoint(block.Header.ShardID)
				subView.tokenID = &tx.TxPrivacyTokenData.PropertyID
				serialNumbersP, commitmentsP, outCoinsP, snDsP, errP := subView.processFetchTxViewPointV2(stateDB, subView.shardID, tx.TxPrivacyTokenData.TxNormal.Proof, subView.tokenID)
				if errP != nil {
					return NewBlockChainError(UnExpectedError, errP)
				}
				subView.listSerialNumbers = serialNumbersP
				for pubkey, data := range commitmentsP {
					if subView.mapCommitments[pubkey] == nil {
						subView.mapCommitments[pubkey] = make([][]byte, 0)
					}
					subView.mapCommitments[pubkey] = append(subView.mapCommitments[pubkey], data...)
					view.txByPubKey[pubkey+"_"+base58.Base58Check{}.Encode(tx.Hash().GetBytes(), 0x0)+"_"+strconv.Itoa(int(block.Header.ShardID))] = true
				}
				for pubkey, data := range outCoinsP {
					if subView.mapOutputCoins[pubkey] == nil {
						subView.mapOutputCoins[pubkey] = make([]privacy.OutputCoin, 0)
					}
					subView.mapOutputCoins[pubkey] = append(subView.mapOutputCoins[pubkey], data...)
				}
				for pubkey, data := range snDsP {
					if subView.mapSnD[pubkey] == nil {
						subView.mapSnD[pubkey] = make([][]byte, 0)
					}
					for _, b := range data {
						temp := b.ToBytesS()
						subView.mapSnD[pubkey] = append(subView.mapSnD[pubkey], temp)
					}
				}
				// subView.listSnD = append(subView.listSnD, snDsP...)
				/*if err != nil {
					return NewBlockChainError(UnExpectedError, err)
				}*/

				view.privacyCustomTokenViewPoint[int32(indexTx)] = subView
				view.privacyCustomTokenTxs[int32(indexTx)] = tx
			}
		default:
			{
				return NewBlockChainError(UnExpectedError, errors.New("TxNormal type is invalid"))
			}
		}
	}

	if len(acceptedSerialNumbers) > 0 {
		view.listSerialNumbers = acceptedSerialNumbers
	}
	if len(acceptedCommitments) > 0 {
		view.mapCommitments = acceptedCommitments
	}
	if len(acceptedOutputcoins) > 0 {
		view.mapOutputCoins = acceptedOutputcoins
	}
	if len(acceptedSnD) > 0 {
		view.mapSnD = acceptedSnD
	}
	return nil
}

func (view *TxViewPoint) processFetchCrossOutputViewPointV2(
	stateDB *statedb.StateDB,
	shardID byte,
	outputCoins []privacy.OutputCoin,
	tokenID *common.Hash,
) (map[string][][]byte, map[string][]privacy.OutputCoin, map[string][]privacy.Scalar, error) {
	acceptedCommitments := make(map[string][][]byte)
	acceptedOutputcoins := make(map[string][]privacy.OutputCoin)
	acceptedSnD := make(map[string][]privacy.Scalar)
	if len(outputCoins) == 0 {
		return acceptedCommitments, acceptedOutputcoins, acceptedSnD, nil
	}

	// Process Output Coins (just created UTXO of this transaction)
	// Proccessed variable: commitment, snd, outputcoins
	// Commitment and SND must not exist before in db
	// Outputcoins will be stored as new utxo for next transaction
	for _, outputCoin := range outputCoins {
		item := &outputCoin
		commitment := item.CoinDetails.GetCoinCommitment().ToBytesS()
		pubkey := item.CoinDetails.GetPublicKey().ToBytesS()
		pubkeyStr := base58.Base58Check{}.Encode(pubkey, common.ZeroByte)
		ok, err := statedb.HasCommitment(stateDB, *tokenID, commitment, shardID)
		if err != nil {
			return acceptedCommitments, acceptedOutputcoins, acceptedSnD, err
		}
		if !ok {
			pubkeyStr := base58.Base58Check{}.Encode(pubkey, common.ZeroByte)
			if acceptedCommitments[pubkeyStr] == nil {
				acceptedCommitments[pubkeyStr] = make([][]byte, 0)
			}
			// get data for commitments
			acceptedCommitments[pubkeyStr] = append(acceptedCommitments[pubkeyStr], item.CoinDetails.GetCoinCommitment().ToBytesS())

			// get data for output coin
			if acceptedOutputcoins[pubkeyStr] == nil {
				acceptedOutputcoins[pubkeyStr] = make([]privacy.OutputCoin, 0)
			}
			acceptedOutputcoins[pubkeyStr] = append(acceptedOutputcoins[pubkeyStr], *item)
		}

		// get data for Snderivators
		snD := item.CoinDetails.GetSNDerivator()
		ok, err = statedb.HasSNDerivator(stateDB, *tokenID, snD.ToBytesS())
		if !ok && err == nil {
			acceptedSnD[pubkeyStr] = append(acceptedSnD[pubkeyStr], *snD)
		}
	}
	return acceptedCommitments, acceptedOutputcoins, acceptedSnD, nil
}

func (view *TxViewPoint) fetchCrossTransactionViewPointFromBlockV2(stateDB *statedb.StateDB, block *ShardBlock) error {
	allShardCrossTransactions := block.Body.CrossTransactions
	// Loop through all of the transaction descs (except for the salary tx)
	acceptedOutputcoins := make(map[string][]privacy.OutputCoin)
	acceptedCommitments := make(map[string][][]byte)
	acceptedSnD := make(map[string][][]byte)
	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	//@NOTICE: this function just work for Normal Transaction

	// sort by shard ID
	shardIDs := []int{}
	for k := range allShardCrossTransactions {
		shardIDs = append(shardIDs, int(k))
	}
	sort.Ints(shardIDs)

	indexOut := 0
	for _, shardID := range shardIDs {
		crossTransactions := allShardCrossTransactions[byte(shardID)]
		for _, crossTransaction := range crossTransactions {
			commitments, outCoins, snDs, err := view.processFetchCrossOutputViewPointV2(stateDB, block.Header.ShardID, crossTransaction.OutputCoin, prvCoinID)
			if err != nil {
				return NewBlockChainError(UnExpectedError, err)
			}
			for pubkey, data := range commitments {
				if acceptedCommitments[pubkey] == nil {
					acceptedCommitments[pubkey] = make([][]byte, 0)
				}
				acceptedCommitments[pubkey] = append(acceptedCommitments[pubkey], data...)
			}
			for pubkey, data := range outCoins {
				if acceptedOutputcoins[pubkey] == nil {
					acceptedOutputcoins[pubkey] = make([]privacy.OutputCoin, 0)
				}
				acceptedOutputcoins[pubkey] = append(acceptedOutputcoins[pubkey], data...)
			}
			for pubkey, data := range snDs {
				if acceptedSnD[pubkey] == nil {
					acceptedSnD[pubkey] = make([][]byte, 0)
				}
				for _, snd := range data {
					acceptedSnD[pubkey] = append(acceptedSnD[pubkey], snd.ToBytesS())
				}
			}
			if crossTransaction.TokenPrivacyData != nil && len(crossTransaction.TokenPrivacyData) > 0 {
				for _, tokenPrivacyData := range crossTransaction.TokenPrivacyData {
					subView := NewTxViewPoint(block.Header.ShardID)
					temp, err := common.Hash{}.NewHash(tokenPrivacyData.PropertyID.GetBytes())
					if err != nil {
						return err
					}
					subView.tokenID = temp
					subView.privacyCustomTokenMetadata.TokenID = *temp
					subView.privacyCustomTokenMetadata.PropertyName = tokenPrivacyData.PropertyName
					subView.privacyCustomTokenMetadata.PropertySymbol = tokenPrivacyData.PropertySymbol
					subView.privacyCustomTokenMetadata.Amount = tokenPrivacyData.Amount
					subView.privacyCustomTokenMetadata.Mintable = tokenPrivacyData.Mintable
					commitmentsP, outCoinsP, snDsP, err := subView.processFetchCrossOutputViewPointV2(stateDB, block.Header.ShardID, tokenPrivacyData.OutputCoin, subView.tokenID)
					if err != nil {
						return NewBlockChainError(UnExpectedError, err)
					}
					for pubkey, data := range commitmentsP {
						if subView.mapCommitments[pubkey] == nil {
							subView.mapCommitments[pubkey] = make([][]byte, 0)
						}
						subView.mapCommitments[pubkey] = append(subView.mapCommitments[pubkey], data...)
					}
					for pubkey, data := range outCoinsP {
						if subView.mapOutputCoins[pubkey] == nil {
							subView.mapOutputCoins[pubkey] = make([]privacy.OutputCoin, 0)
						}
						subView.mapOutputCoins[pubkey] = append(subView.mapOutputCoins[pubkey], data...)
					}
					for pubkey, data := range snDsP {
						if subView.mapSnD[pubkey] == nil {
							subView.mapSnD[pubkey] = make([][]byte, 0)
						}
						for _, t := range data {
							temp := t.ToBytesS()
							subView.mapSnD[pubkey] = append(subView.mapSnD[pubkey], temp)
						}
					}
					view.privacyCustomTokenViewPoint[int32(indexOut)] = subView
					indexOut++
				}
			}
		}
	}

	if len(acceptedCommitments) > 0 {
		view.mapCommitments = acceptedCommitments
	}
	if len(acceptedOutputcoins) > 0 {
		view.mapOutputCoins = acceptedOutputcoins
	}
	if len(acceptedSnD) > 0 {
		view.mapSnD = acceptedSnD
	}
	return nil
}
