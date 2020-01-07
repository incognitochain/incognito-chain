package blockchain

import (
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/wallet"
)

func (blockchain *BlockChain) updateDatabaseWithBlockRewardInfoV2(beaconBlock *BeaconBlock) error {
	for _, inst := range beaconBlock.Body.Instructions {
		if len(inst) <= 2 {
			continue
		}
		if inst[0] == SetAction || inst[0] == StakeAction || inst[0] == RandomAction || inst[0] == SwapAction || inst[0] == AssignAction {
			continue
		}
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
			for key, value := range acceptedBlkRewardInfo.TxsFee {
				if value != 0 {
					err = statedb.AddShardRewardRequest(blockchain.BestState.Beacon.rewardStateDB, beaconBlock.Header.Epoch, acceptedBlkRewardInfo.ShardID, key, value)
					if err != nil {
						return err
					}
				}
			}
			continue
		}
	}
	return nil
}

func (blockchain *BlockChain) updateDatabaseFromBeaconInstructionsV2(beaconBlocks []*BeaconBlock, shardID byte, rewardStateDB *statedb.StateDB) error {
	rewardReceivers := make(map[string]string)
	committees := make(map[int][]incognitokey.CommitteePublicKey)
	isInit := false
	epoch := uint64(0)
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
					for key := range beaconBlkRewardInfo.BeaconReward {
						err = statedb.AddCommitteeReward(rewardStateDB, beaconBlkRewardInfo.PayToPublicKey, beaconBlkRewardInfo.BeaconReward[key], key)
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
						tempPublicKey := base58.Base58Check{}.Encode(keyWalletDevAccount.KeySet.PaymentAddress.Pk, common.Base58Version)
						err = statedb.AddCommitteeReward(rewardStateDB, tempPublicKey, incDAORewardInfo.IncDAOReward[key], key)
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
					height := shardRewardInfo.Epoch * blockchain.config.ChainParams.Epoch
					consensusRootHash := blockchain.BestState.Beacon.ConsensusStateRootHash[height]
					consensusStateDB, err := statedb.NewWithPrefixTrie(consensusRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetDatabase()))
					if err != nil {
						return err
					}
					committees, _, _, _, rewardReceivers, _ = statedb.GetAllCommitteeSubstituteCandidateState(consensusStateDB, blockchain.GetShardIDs())

				}
				err = blockchain.getRewardAmountForUserOfShardV2(rewardStateDB, shardID, shardRewardInfo, committees[int(shardID)], rewardReceivers)
				if err != nil {
					return err
				}
				continue
			}

		}
	}
	return nil
}

func (blockchain *BlockChain) getRewardAmountForUserOfShardV2(rewardStateDB *statedb.StateDB, shardID byte, rewardInfoShardToProcess *metadata.ShardBlockRewardInfo, committeeOfShardToProcess []incognitokey.CommitteePublicKey, rewardReceiver map[string]string) (err error) {
	committeeSize := len(committeeOfShardToProcess)
	for _, candidate := range committeeOfShardToProcess {
		wl, err := wallet.Base58CheckDeserialize((rewardReceiver)[candidate.GetIncKeyBase58()])
		if err != nil {
			return err
		}
		if common.GetShardIDFromLastByte(wl.KeySet.PaymentAddress.Pk[common.PublicKeySize-1]) == shardID {
			for key, value := range rewardInfoShardToProcess.ShardReward {
				tempPK := base58.Base58Check{}.Encode(wl.KeySet.PaymentAddress.Pk, common.Base58Version)
				err = statedb.AddCommitteeReward(rewardStateDB, tempPK, value/uint64(committeeSize), key)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
