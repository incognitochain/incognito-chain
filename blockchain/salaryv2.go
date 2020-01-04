package blockchain

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"strconv"
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

func (blockchain *BlockChain) updateDatabaseFromBeaconInstructionsV2(beaconBlocks []*BeaconBlock, shardID byte) error {
	rewardReceivers := make(map[string]string)
	committee := make(map[byte][]incognitokey.CommitteePublicKey)
	isInit := false
	epoch := uint64(0)
	for _, beaconBlock := range beaconBlocks {
		//fmt.Printf("RewardLog Process BeaconBlock %v\n", beaconBlock.GetHeight())
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
					// fmt.Printf("RewardLog Process Beacon %v\n", l)
					beaconBlkRewardInfo, err := metadata.NewBeaconBlockRewardInfoFromStr(l[3])
					if err != nil {
						return err
					}
					publicKeyCommittee, _, err := base58.Base58Check{}.Decode(beaconBlkRewardInfo.PayToPublicKey)
					if err != nil {
						return err
					}
					for key := range beaconBlkRewardInfo.BeaconReward {
						err = rawdb.AddCommitteeReward(blockchain.GetDatabase(), publicKeyCommittee, beaconBlkRewardInfo.BeaconReward[key], key)
						if err != nil {
							return err
						}
					}
					continue

				case metadata.IncDAORewardRequestMeta:
					fmt.Printf("RewardLog Process Dev %v\n", l)
					incDAORewardInfo, err := metadata.NewIncDAORewardInfoFromStr(l[3])
					if err != nil {
						return err
					}
					keyWalletDevAccount, err := wallet.Base58CheckDeserialize(blockchain.config.ChainParams.IncognitoDAOAddress)
					if err != nil {
						return err
					}
					for key := range incDAORewardInfo.IncDAOReward {
						err = rawdb.AddCommitteeReward(blockchain.GetDatabase(), keyWalletDevAccount.KeySet.PaymentAddress.Pk, incDAORewardInfo.IncDAOReward[key], key)
						if err != nil {
							return err
						}
					}
					continue
				}
			}
			switch metaType {
			case metadata.ShardBlockRewardRequestMeta:
				//fmt.Printf("RewardLog Process Shard %v\n", l)
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
					json.Unmarshal(rewardReceiverBytes, &rewardReceivers)
					committeeBytes, err := rawdb.FetchShardCommitteeByHeight(blockchain.GetDatabase(), epoch*blockchain.config.ChainParams.Epoch)
					if err != nil {
						return err
					}
					json.Unmarshal(committeeBytes, &committee)
				}
				err = blockchain.getRewardAmountForUserOfShard(shardID, shardRewardInfo, committee[byte(shardToProcess)], &rewardReceivers, false)
				if err != nil {
					return err
				}
				continue
			}

		}
	}
	return nil
}
