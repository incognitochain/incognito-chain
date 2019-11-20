package blockchain

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/core/rawdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

func (blockchain *BlockChain) ValidateBlockWithPrevShardBestState(block *ShardBlock) error {
	prevBST, err := rawdb.FetchPrevBestState(blockchain.config.DataBase, false, block.Header.ShardID)
	if err != nil {
		return err
	}
	shardBestState := ShardBestState{}
	if err := json.Unmarshal(prevBST, &shardBestState); err != nil {
		return err
	}

	// blkHash := block.Header.Hash()
	// producerPk := base58.Base58Check{}.Encode(block.Header.ProducerAddress.Pk, common.ZeroByte)
	// err = incognitokey.ValidateDataB58(producerPk, block.ProducerSig, blkHash.GetBytes())
	// if err != nil {
	// 	return NewBlockChainError(ProducerError, errors.New("Producer's sig not match"))
	// }
	//verify producer
	// block.GetValidationField()
	producerPk := block.Header.Producer
	producerPosition := (shardBestState.ShardProposerIdx + block.Header.Round) % len(shardBestState.ShardCommittee)
	tempProducer := shardBestState.ShardCommittee[producerPosition].GetMiningKeyBase58(shardBestState.ConsensusAlgorithm)
	if strings.Compare(tempProducer, producerPk) != 0 {
		return NewBlockChainError(ProducerError, errors.New("Producer should be should be :"+tempProducer))
	}
	// if block.Header.Version != SHARD_BLOCK_VERSION {
	// 	return NewBlockChainError(, errors.New("Version should be :"+strconv.Itoa(VERSION)))
	// }
	// Verify parent hash exist or not
	prevBlockHash := block.Header.PreviousBlockHash
	parentBlockData, err := rawdb.FetchBlock(blockchain.config.DataBase, prevBlockHash)
	if err != nil {
		return NewBlockChainError(DatabaseError, err)
	}
	parentBlock := ShardBlock{}
	json.Unmarshal(parentBlockData, &parentBlock)
	// Verify block height with parent block
	if parentBlock.Header.Height+1 != block.Header.Height {
		return NewBlockChainError(ShardStateError, errors.New("block height of new block should be :"+strconv.Itoa(int(block.Header.Height+1))))
	}
	return nil
}

//This only happen if user is a shard committee member.
func (blockchain *BlockChain) RevertShardState(shardID byte) error {
	//Steps:
	// 1. Restore current beststate to previous beststate
	// 2. Set pool shardstate
	// 3. Delete newly inserted block
	// 4. Remove incoming crossShardBlks
	// 5. Delete txs and its related stuff (ex: txview) belong to block

	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	return blockchain.revertShardState(shardID)
}

func (blockchain *BlockChain) revertShardBestState(shardID byte) error {
	prevBST, err := rawdb.FetchPrevBestState(blockchain.GetDatabase(), false, shardID)
	if err != nil {
		return err
	}
	shardBestState := ShardBestState{}
	if err := json.Unmarshal(prevBST, &shardBestState); err != nil {
		return err
	}

	SetBestStateShard(shardID, &shardBestState)

	blockchain.config.ShardPool[shardID].RevertShardPool(shardBestState.ShardHeight)
	for sid, height := range shardBestState.BestCrossShard {
		blockchain.config.CrossShardPool[sid].RevertCrossShardPool(height)
	}

	return nil
}

func (blockchain *BlockChain) revertShardState(shardID byte) error {
	//Steps:
	// 1. Restore current beststate to previous beststate
	// 2. Set pool shardstate
	// 3. Delete newly inserted block
	// 4. Remove incoming crossShardBlks
	// 5. Delete txs and its related stuff (ex: txview) belong to block
	var currentBestState ShardBestState
	currentBestState.cloneShardBestStateFrom(blockchain.BestState.Shard[shardID])
	currentBestStateBlk := currentBestState.BestBlock

	if currentBestState.ShardHeight == blockchain.BestState.Shard[shardID].ShardHeight {
		return NewBlockChainError(RevertStateError, errors.New("can't revert same beststate"))
	}

	err := blockchain.revertShardBestState(shardID)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}

	err = blockchain.DeleteIncomingCrossShard(currentBestStateBlk)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}

	for _, tx := range currentBestState.BestBlock.Body.Transactions {
		if err := rawdb.DeleteTransactionIndex(blockchain.GetDatabase(), *tx.Hash()); err != nil {
			return NewBlockChainError(RevertStateError, err)
		}
	}

	if err := blockchain.restoreFromTxViewPoint(currentBestStateBlk); err != nil {
		return NewBlockChainError(RevertStateError, err)
	}

	if err := blockchain.restoreFromCrossTxViewPoint(currentBestStateBlk); err != nil {
		return NewBlockChainError(RevertStateError, err)
	}

	prevBeaconHeight := currentBestState.BeaconHeight
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain.GetDatabase(), prevBeaconHeight+1, currentBestStateBlk.Header.BeaconHeight)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}

	if err := blockchain.restoreDatabaseFromBeaconInstruction(beaconBlocks, currentBestStateBlk.Header.ShardID); err != nil {
		return NewBlockChainError(RevertStateError, err)
	}

	// DeleteIncomingCrossShard
	rawdb.DeleteBlock(blockchain.GetDatabase(), currentBestStateBlk.Header.Hash(), currentBestStateBlk.Header.Height, shardID)

	if err := blockchain.StoreShardBestState(shardID, nil); err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	Logger.log.Critical("REVERT SHARD SUCCESS")
	return nil
}

func (blockchain *BlockChain) BackupCurrentShardState(block *ShardBlock, beaconblks []*BeaconBlock) error {

	//Steps:
	// 1. Backup beststate
	// 2.	Backup data that will be modify by new block data

	tempMarshal, err := json.Marshal(blockchain.BestState.Shard[block.Header.ShardID])
	if err != nil {
		return NewBlockChainError(UnmashallJsonShardBlockError, err)
	}

	if err := rawdb.StorePrevBestState(blockchain.GetDatabase(), tempMarshal, false, block.Header.ShardID); err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	if err := blockchain.createBackupFromTxViewPoint(block); err != nil {
		return NewBlockChainError(BackupFromTxViewPointError, err)
	}

	if err := blockchain.createBackupFromCrossTxViewPoint(block); err != nil {
		return NewBlockChainError(BackupFromCrossTxViewPointError, err)
	}

	if err := blockchain.backupDatabaseFromBeaconInstruction(beaconblks, block.Header.ShardID); err != nil {
		return NewBlockChainError(BackupDatabaseFromBeaconInstructionError, err)
	}

	return nil
}

func (blockchain *BlockChain) backupDatabaseFromBeaconInstruction(
	beaconBlocks []*BeaconBlock,
	shardID byte,
) error {
	rewardReceivers := make(map[string]string)
	committee := make(map[byte][]incognitokey.CommitteePublicKey)
	isInit := false
	epoch := uint64(0)
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			if l[0] == StakeAction || l[0] == RandomAction || l[0] == SwapAction || l[0] == AssignAction || l[0] == StopAutoStake {
				continue
			}
			if len(l) <= 2 {
				continue
			}
			shardToProcess, err := strconv.Atoi(l[1])
			if err != nil {
				continue
			}
			metaType, err := strconv.Atoi(l[0])
			if err != nil {
				return err
			}
			if shardToProcess == int(shardID) {
				switch metaType {
				case metadata.BeaconRewardRequestMeta:
					beaconBlkRewardInfo, err := metadata.NewBeaconBlockRewardInfoFromStr(l[3])
					if err != nil {
						return err
					}
					publicKeyCommittee, _, err := base58.Base58Check{}.Decode(beaconBlkRewardInfo.PayToPublicKey)
					if err != nil {
						return err
					}
					for key := range beaconBlkRewardInfo.BeaconReward {
						err = rawdb.BackupCommitteeReward(blockchain.GetDatabase(), publicKeyCommittee, key)
						if err != nil {
							return err
						}
					}
					continue

				case metadata.IncDAORewardRequestMeta:
					incDAORewardInfo, err := metadata.NewIncDAORewardInfoFromStr(l[3])
					if err != nil {
						return err
					}
					keyWalletDevAccount, err := wallet.Base58CheckDeserialize(blockchain.config.ChainParams.IncognitoDAOAddress)
					if err != nil {
						return err
					}
					for key := range incDAORewardInfo.IncDAOReward {
						err = rawdb.BackupCommitteeReward(blockchain.GetDatabase(), keyWalletDevAccount.KeySet.PaymentAddress.Pk, key)
						if err != nil {
							return err
						}
					}
					continue
				}
			}
			switch metaType {
			case metadata.ShardBlockRewardRequestMeta:
				shardRewardInfo, err := metadata.NewShardBlockRewardInfoFromString(l[3])
				if err != nil {
					return err
				}
				if (!isInit) || (epoch != shardRewardInfo.Epoch) {
					isInit = true
					epoch = shardRewardInfo.Epoch
					rewardReceiverBytes, err := rawdb.FetchRewardReceiverByHeight(blockchain.GetDatabase(), epoch*blockchain.config.ChainParams.Epoch)
					if err != nil {
						return err
					}
					err = json.Unmarshal(rewardReceiverBytes, &rewardReceivers)
					if err != nil {
						return err
					}
					committeeBytes, err := rawdb.FetchShardCommitteeByHeight(blockchain.GetDatabase(), epoch*blockchain.config.ChainParams.Epoch)
					if err != nil {
						return err
					}
					err = json.Unmarshal(committeeBytes, &committee)
					if err != nil {
						return err
					}
				}
				//TODO: check later
				err = blockchain.getRewardAmountForUserOfShard(shardID, shardRewardInfo, committee[byte(shardToProcess)], &rewardReceivers, true)
				if err != nil {
					return err
				}
				continue
			}

		}
	}
	return nil
}

func (blockchain *BlockChain) createBackupFromTxViewPoint(block *ShardBlock) error {
	// Fetch data from block into tx View point
	view := NewTxViewPoint(block.Header.ShardID)
	err := view.fetchTxViewPointFromBlock(blockchain.GetDatabase(), block)
	if err != nil {
		return err
	}

	// check privacy custom token
	backupedView := make(map[string]bool)
	for _, privacyCustomTokenSubView := range view.privacyCustomTokenViewPoint {
		if ok := backupedView[privacyCustomTokenSubView.tokenID.String()]; !ok {
			err = blockchain.backupSerialNumbersFromTxViewPoint(*privacyCustomTokenSubView)
			if err != nil {
				return err
			}

			err = blockchain.backupCommitmentsFromTxViewPoint(*privacyCustomTokenSubView)
			if err != nil {
				return err
			}
			backupedView[privacyCustomTokenSubView.tokenID.String()] = true
		}

	}
	err = blockchain.backupSerialNumbersFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	err = blockchain.backupCommitmentsFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	return nil
}

func (blockchain *BlockChain) createBackupFromCrossTxViewPoint(block *ShardBlock) error {
	view := NewTxViewPoint(block.Header.ShardID)
	err := view.fetchCrossTransactionViewPointFromBlock(blockchain.GetDatabase(), block)

	for _, privacyCustomTokenSubView := range view.privacyCustomTokenViewPoint {
		err = blockchain.backupCommitmentsFromTxViewPoint(*privacyCustomTokenSubView)
		if err != nil {
			return err
		}
	}
	err = blockchain.backupCommitmentsFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	return nil
}

func (blockchain *BlockChain) backupSerialNumbersFromTxViewPoint(view TxViewPoint) error {
	err := rawdb.BackupSerialNumbersLen(blockchain.GetDatabase(), *view.tokenID, view.shardID)
	if err != nil {
		return err
	}
	return nil
}

func (blockchain *BlockChain) backupCommitmentsFromTxViewPoint(view TxViewPoint) error {

	// commitment
	keys := make([]string, 0, len(view.mapCommitments))
	for k := range view.mapCommitments {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		pubkey := k
		pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
		if err != nil {
			return err
		}
		lastByte := pubkeyBytes[len(pubkeyBytes)-1]
		pubkeyShardID := common.GetShardIDFromLastByte(lastByte)
		if pubkeyShardID == view.shardID {
			err = rawdb.BackupCommitmentsOfPublicKey(blockchain.GetDatabase(), *view.tokenID, view.shardID, pubkeyBytes)
			if err != nil {
				return err
			}
		}
	}

	// outputs
	keys = make([]string, 0, len(view.mapOutputCoins))
	for k := range view.mapOutputCoins {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// for _, k := range keys {
	// 	pubkey := k

	// 	pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	lastByte := pubkeyBytes[len(pubkeyBytes)-1]
	// 	pubkeyShardID := common.GetShardIDFromLastByte(lastByte)
	// 	if pubkeyShardID == view.shardID {
	// 		err = rawdb.BackupOutputCoin(*view.tokenID, pubkeyBytes, pubkeyShardID)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	return nil
}

func (blockchain *BlockChain) restoreDatabaseFromBeaconInstruction(beaconBlocks []*BeaconBlock,
	shardID byte) error {

	shardCommittee := make(map[byte][]string)
	isInit := false
	epoch := uint64(0)
	// listShardCommittee := rawdb.FetchCommitteeByEpoch
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			if l[0] == StakeAction || l[0] == RandomAction {
				continue
			}
			if len(l) <= 2 {
				continue
			}
			shardToProcess, err := strconv.Atoi(l[1])
			if err != nil {
				continue
			}
			if shardToProcess == int(shardID) {
				metaType, err := strconv.Atoi(l[0])
				if err != nil {
					return err
				}
				switch metaType {
				case metadata.BeaconRewardRequestMeta:
					beaconBlkRewardInfo, err := metadata.NewBeaconBlockRewardInfoFromStr(l[3])
					if err != nil {
						return err
					}
					publicKeyCommittee, _, err := base58.Base58Check{}.Decode(beaconBlkRewardInfo.PayToPublicKey)
					if err != nil {
						return err
					}
					for key := range beaconBlkRewardInfo.BeaconReward {
						err = rawdb.RestoreCommitteeReward(blockchain.GetDatabase(), publicKeyCommittee, key)
						if err != nil {
							return err
						}
					}
					continue

				case metadata.IncDAORewardRequestMeta:
					incDAORewardInfo, err := metadata.NewIncDAORewardInfoFromStr(l[3])
					if err != nil {
						return err
					}
					keyWalletDevAccount, err := wallet.Base58CheckDeserialize(blockchain.config.ChainParams.IncognitoDAOAddress)
					if err != nil {
						return err
					}
					for key := range incDAORewardInfo.IncDAOReward {
						err = rawdb.RestoreCommitteeReward(blockchain.GetDatabase(), keyWalletDevAccount.KeySet.PaymentAddress.Pk, key)
						if err != nil {
							return err
						}
					}
					continue

				case metadata.ShardBlockRewardRequestMeta:
					shardRewardInfo, err := metadata.NewShardBlockRewardInfoFromString(l[3])
					if err != nil {
						return err
					}
					if (!isInit) || (epoch != shardRewardInfo.Epoch) {
						isInit = true
						epoch = shardRewardInfo.Epoch
						temp, err := rawdb.FetchShardCommitteeByHeight(blockchain.GetDatabase(), epoch*blockchain.config.ChainParams.Epoch)
						if err != nil {
							return err
						}
						json.Unmarshal(temp, &shardCommittee)
					}
					err = blockchain.restoreShareRewardForShardCommittee(shardRewardInfo.Epoch, shardRewardInfo.ShardReward, shardCommittee[shardID])
					if err != nil {
						return err
					}
					continue
				}
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) restoreShareRewardForShardCommittee(epoch uint64, totalReward map[common.Hash]uint64, listCommitee []string) error {
	// reward := totalReward / uint64(len(listCommitee))
	reward := map[common.Hash]uint64{}
	for key, value := range totalReward {
		reward[key] = value / uint64(len(listCommitee))
	}
	for key := range totalReward {
		for _, committee := range listCommitee {
			committeeBytes, _, err := base58.Base58Check{}.Decode(committee)
			if err != nil {
				return err
			}
			err = rawdb.RestoreCommitteeReward(blockchain.GetDatabase(), committeeBytes, key)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) restoreFromTxViewPoint(block *ShardBlock) error {
	// Fetch data from block into tx View point
	view := NewTxViewPoint(block.Header.ShardID)
	err := view.fetchTxViewPointFromBlock(blockchain.GetDatabase(), block)
	if err != nil {
		return err
	}

	// check privacy custom token
	for indexTx, privacyCustomTokenSubView := range view.privacyCustomTokenViewPoint {
		privacyCustomTokenTx := view.privacyCustomTokenTxs[indexTx]
		switch privacyCustomTokenTx.TxPrivacyTokenData.Type {
		case transaction.TokenInit:
			{
				err = rawdb.DeletePrivacyToken(blockchain.GetDatabase(), privacyCustomTokenTx.TxPrivacyTokenData.PropertyID)
				if err != nil {
					return err
				}
			}
		}
		err = rawdb.DeletePrivacyTokenTx(blockchain.GetDatabase(), privacyCustomTokenTx.TxPrivacyTokenData.PropertyID, indexTx, block.Header.ShardID, block.Header.Height)
		if err != nil {
			return err
		}

		err = blockchain.restoreSerialNumbersFromTxViewPoint(*privacyCustomTokenSubView)
		if err != nil {
			return err
		}

		err = blockchain.restoreCommitmentsFromTxViewPoint(*privacyCustomTokenSubView, block.Header.ShardID)
		if err != nil {
			return err
		}
	}

	err = blockchain.restoreSerialNumbersFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	err = blockchain.restoreCommitmentsFromTxViewPoint(*view, block.Header.ShardID)
	if err != nil {
		return err
	}

	return nil
}

func (blockchain *BlockChain) restoreFromCrossTxViewPoint(block *ShardBlock) error {
	view := NewTxViewPoint(block.Header.ShardID)
	err := view.fetchCrossTransactionViewPointFromBlock(blockchain.GetDatabase(), block)

	for _, privacyCustomTokenSubView := range view.privacyCustomTokenViewPoint {
		tokenID := privacyCustomTokenSubView.tokenID
		if err := rawdb.DeletePrivacyTokenCrossShard(blockchain.GetDatabase(), *tokenID); err != nil {
			return err
		}
		err = blockchain.restoreCommitmentsFromTxViewPoint(*privacyCustomTokenSubView, block.Header.ShardID)
		if err != nil {
			return err
		}
	}

	err = blockchain.restoreCommitmentsFromTxViewPoint(*view, block.Header.ShardID)
	if err != nil {
		return err
	}
	return nil
}

func (blockchain *BlockChain) restoreSerialNumbersFromTxViewPoint(view TxViewPoint) error {
	err := rawdb.RestoreSerialNumber(blockchain.GetDatabase(), *view.tokenID, view.shardID, view.listSerialNumbers)
	if err != nil {
		return err
	}
	return nil
}

func (blockchain *BlockChain) restoreCommitmentsFromTxViewPoint(view TxViewPoint, shardID byte) error {

	// commitment
	keys := make([]string, 0, len(view.mapCommitments))
	for k := range view.mapCommitments {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		pubkey := k
		item1 := view.mapCommitments[k]
		pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
		if err != nil {
			return err
		}
		lastByte := pubkeyBytes[len(pubkeyBytes)-1]
		pubkeyShardID := common.GetShardIDFromLastByte(lastByte)
		if pubkeyShardID == view.shardID {
			err = rawdb.RestoreCommitmentsOfPubkey(blockchain.GetDatabase(), *view.tokenID, view.shardID, pubkeyBytes, item1)
			if err != nil {
				return err
			}
		}
	}

	// outputs
	for _, k := range keys {
		publicKey := k
		publicKeyBytes, _, err := base58.Base58Check{}.Decode(publicKey)
		if err != nil {
			return err
		}
		lastByte := publicKeyBytes[len(publicKeyBytes)-1]
		publicKeyShardID := common.GetShardIDFromLastByte(lastByte)
		if publicKeyShardID == shardID {
			outputCoinArray := view.mapOutputCoins[k]
			outputCoinBytesArray := make([][]byte, 0)
			for _, outputCoin := range outputCoinArray {
				outputCoinBytesArray = append(outputCoinBytesArray, outputCoin.Bytes())
			}
			err = rawdb.DeleteOutputCoin(blockchain.GetDatabase(), *view.tokenID, publicKeyBytes, outputCoinBytesArray, publicKeyShardID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) ValidateBlockWithPrevBeaconBestState(block *BeaconBlock) error {
	prevBST, err := rawdb.FetchPrevBestState(blockchain.GetDatabase(), true, 0)
	if err != nil {
		return err
	}
	beaconBestState := BeaconBestState{}
	if err := json.Unmarshal(prevBST, &beaconBestState); err != nil {
		return err
	}
	producerPk := block.Header.Producer
	// err = incognitokey.ValidateDataB58(producerPk, block.ProducerSig, blkHash.GetBytes())
	// if err != nil {
	// 	return NewBlockChainError(ProducerError, errors.New("Producer's sig not match"))
	// }
	//verify producer
	producerPosition := (beaconBestState.BeaconProposerIndex + block.Header.Round) % len(beaconBestState.BeaconCommittee)
	tempProducer := beaconBestState.BeaconCommittee[producerPosition].GetMiningKeyBase58(beaconBestState.ConsensusAlgorithm)
	if strings.Compare(tempProducer, producerPk) != 0 {
		return NewBlockChainError(ProducerError, errors.New("Producer should be should be :"+tempProducer))
	}
	//verify version
	if block.Header.Version != BEACON_BLOCK_VERSION {
		return NewBlockChainError(WrongVersionError, errors.New("Version should be :"+strconv.Itoa(BEACON_BLOCK_VERSION)))
	}
	prevBlockHash := block.Header.PreviousBlockHash
	// Verify parent hash exist or not
	parentBlockBytes, err := rawdb.FetchBeaconBlock(blockchain.GetDatabase(), prevBlockHash)
	if err != nil {
		return NewBlockChainError(DatabaseError, err)
	}
	parentBlock := NewBeaconBlock()
	err = json.Unmarshal(parentBlockBytes, parentBlock)
	if err != nil {

	}
	// Verify block height with parent block
	if parentBlock.Header.Height+1 != block.Header.Height {
		return NewBlockChainError(WrongBlockHeightError, errors.New("block height of new block should be :"+strconv.Itoa(int(block.Header.Height+1))))
	}
	return nil
}

//This only happen if user is a beacon committee member.
func (blockchain *BlockChain) RevertBeaconState() error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	return blockchain.revertBeaconState()
}

func (blockchain *BlockChain) revertBeaconBestState() error {
	prevBST, err := rawdb.FetchPrevBestState(blockchain.GetDatabase(), true, 0)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	beaconBestState := BeaconBestState{}
	if err := json.Unmarshal(prevBST, &beaconBestState); err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	SetBeaconBestState(&beaconBestState)

	blockchain.config.BeaconPool.RevertBeconPool(beaconBestState.BeaconHeight)
	for sid, height := range blockchain.BestState.Beacon.GetBestShardHeight() {
		blockchain.config.ShardToBeaconPool.RevertShardToBeaconPool(sid, height)
	}

	return nil
}

func (blockchain *BlockChain) revertBeaconState() error {
	//Steps:
	// 1. Restore current beststate to previous beststate
	// 2. Set beacon/shardtobeacon pool state
	// 3. Delete newly inserted block
	// 4. Delete data store by block
	var currentBestState BeaconBestState
	currentBestState.CloneBeaconBestStateFrom(blockchain.BestState.Beacon)
	currentBestStateBlk := currentBestState.BestBlock

	err := blockchain.revertBeaconBestState()
	if err != nil {
		return err
	}

	if err := rawdb.DeleteCommitteeByHeight(blockchain.GetDatabase(), currentBestStateBlk.Header.Height); err != nil {
		return err
	}

	for shardID, shardStates := range currentBestStateBlk.Body.ShardState {
		for _, shardState := range shardStates {
			rawdb.DeleteAcceptedShardToBeacon(blockchain.GetDatabase(), shardID, shardState.Hash)
		}
	}

	lastCrossShardState := beaconBestState.LastCrossShardState
	for fromShard, toShards := range lastCrossShardState {
		for toShard, height := range toShards {
			rawdb.RestoreCrossShardNextHeights(blockchain.GetDatabase(), fromShard, toShard, height)
		}
		blockchain.config.CrossShardPool[fromShard].UpdatePool()
	}
	for _, inst := range currentBestStateBlk.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not bridge instruction
		}
		if inst[0] == SetAction || inst[0] == StakeAction || inst[0] == RandomAction || inst[0] == SwapAction || inst[0] == AssignAction {
			continue
		}
		var err error
		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			continue
		}
		switch metaType {
		case metadata.AcceptedBlockRewardInfoMeta:
			acceptedBlkRewardInfo, err := metadata.NewAcceptedBlockRewardInfoFromStr(inst[2])
			if err != nil {
				return err
			}
			if val, ok := acceptedBlkRewardInfo.TxsFee[common.PRVCoinID]; ok {
				acceptedBlkRewardInfo.TxsFee[common.PRVCoinID] = val + blockchain.getRewardAmount(acceptedBlkRewardInfo.ShardBlockHeight)
			} else {
				if acceptedBlkRewardInfo.TxsFee == nil {
					acceptedBlkRewardInfo.TxsFee = map[common.Hash]uint64{}
				}
				acceptedBlkRewardInfo.TxsFee[common.PRVCoinID] = blockchain.getRewardAmount(acceptedBlkRewardInfo.ShardBlockHeight)
			}
			Logger.log.Infof("TxsFee in Epoch: %+v of shardID: %+v:\n", currentBestStateBlk.Header.Epoch, acceptedBlkRewardInfo.ShardID)
			for key, value := range acceptedBlkRewardInfo.TxsFee {
				Logger.log.Infof("===> TokenID:%+v: Amount: %+v\n", key, value)
				err = rawdb.RestoreShardRewardRequest(blockchain.GetDatabase(), currentBestStateBlk.Header.Epoch, acceptedBlkRewardInfo.ShardID, key)
				if err != nil {
					return err
				}

			}
		}
	}
	err = rawdb.DeleteBeaconBlock(blockchain.GetDatabase(), currentBestStateBlk.Header.Hash(), currentBestStateBlk.Header.Height)
	if err != nil {
		return err
	}

	if err := blockchain.StoreBeaconBestState(nil); err != nil {
		return err
	}
	Logger.log.Critical("REVERT BEACON SUCCESS")
	return nil
}

func (blockchain *BlockChain) BackupCurrentBeaconState(block *BeaconBlock) error {
	//Steps:
	// 1. Backup beststate
	tempMarshal, err := json.Marshal(blockchain.BestState.Beacon)
	if err != nil {
		return NewBlockChainError(UnmashallJsonShardBlockError, err)
	}
	if err := rawdb.StorePrevBestState(blockchain.GetDatabase(), tempMarshal, true, 0); err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	for _, inst := range block.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not bridge instruction
		}
		if inst[0] == SetAction || inst[0] == StakeAction || inst[0] == RandomAction || inst[0] == SwapAction || inst[0] == AssignAction {
			continue
		}
		var err error
		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			continue
		}

		switch metaType {
		case metadata.AcceptedBlockRewardInfoMeta:
			acceptedBlkRewardInfo, err := metadata.NewAcceptedBlockRewardInfoFromStr(inst[2])
			if err != nil {
				return err
			}
			if val, ok := acceptedBlkRewardInfo.TxsFee[common.PRVCoinID]; ok {
				acceptedBlkRewardInfo.TxsFee[common.PRVCoinID] = val + blockchain.getRewardAmount(acceptedBlkRewardInfo.ShardBlockHeight)
			} else {
				if acceptedBlkRewardInfo.TxsFee == nil {
					acceptedBlkRewardInfo.TxsFee = map[common.Hash]uint64{}
				}
				acceptedBlkRewardInfo.TxsFee[common.PRVCoinID] = blockchain.getRewardAmount(acceptedBlkRewardInfo.ShardBlockHeight)
			}
			for key := range acceptedBlkRewardInfo.TxsFee {
				err = rawdb.BackupShardRewardRequest(blockchain.GetDatabase(), block.Header.Epoch, acceptedBlkRewardInfo.ShardID, key)
				if err != nil {
					return err
				}

			}
		}
	}
	return nil
}
