package blsbft

import (
	"errors"
	"fmt"
	"sort"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
)

type actorV2 struct {
	actorBase
	currentTime     int64
	currentTimeSlot int64
	proposeHistory  *lru.Cache

	receiveBlockByHeight map[uint64][]*ProposeBlockInfo  //blockHeight -> blockInfo
	receiveBlockByHash   map[string]*ProposeBlockInfo    //blockHash -> blockInfo
	voteHistory          map[uint64]types.BlockInterface // bestview height (previsous height )-> block
}

func (actorV2 *actorV2) Destroy() {
	actorV2.actorBase.Destroy()
	close(actorV2.destroyCh)
}

//only run when init process
func (actorV2 *actorV2) run() error {
	go func() {
		//init view maps
		ticker := time.Tick(200 * time.Millisecond)
		cleanMemTicker := time.Tick(5 * time.Minute)
		actorV2.logger.Info("init bls-bftv2 consensus for chain", actorV2.chainKey)

		for { //actor loop
			/*if actorV2.chain.CommitteeEngineVersion() != committeestate.SELF_SWAP_SHARD_VERSION {*/
			//actorV2.logger.Infof("CHAIN ID %+v |Require BFTACTOR V2 FOR Committee Engine V1, current Committee Engine %+v ", actorV2.chain.GetShardID(), actorV2.chain.CommitteeEngineVersion())
			//e.Logger.Info("stop bls-bft2 consensus for chain", e.ChainKey)
			//time.Sleep(time.Second)
			//continue
			/*}*/
			if !actorV2.isStarted { //sleep if this process is not start
				time.Sleep(time.Second)
				continue
			}
			select {
			case <-actorV2.destroyCh:
				actorV2.logger.Info("exit bls-bftv2 consensus for chain", actorV2.chainKey)
				return
			case proposeMsg := <-actorV2.proposeMessageCh:
				//fmt.Println("debug receive propose message", string(proposeMsg.Block))
				blockIntf, err := actorV2.chain.UnmarshalBlock(proposeMsg.Block)
				if err != nil || blockIntf == nil {
					actorV2.logger.Info(err)
					continue
				}
				block := blockIntf.(types.BlockInterface)
				blkHash := block.Hash().String()

				if _, ok := actorV2.receiveBlockByHash[blkHash]; !ok {
					actorV2.receiveBlockByHash[blkHash] = &ProposeBlockInfo{
						block:       block,
						votes:       make(map[string]*BFTVote),
						hasNewVote:  false,
						receiveTime: time.Now(),
					}
					e.Logger.Info(e.ChainKey, "Receive block ", block.Hash().String(), "height", block.GetHeight(), ",block timeslot ", common.CalculateTimeSlot(block.GetProposeTime()))
					e.receiveBlockByHeight[block.GetHeight()] = append(e.receiveBlockByHeight[block.GetHeight()], e.receiveBlockByHash[blkHash])
				} else {
					e.receiveBlockByHash[blkHash].block = block
				}

				if block.GetHeight() <= e.Chain.GetBestView().GetHeight() {
					e.Logger.Infof("%v Receive block create from old view - height %v. Rejected! Expect: %v", e.ChainKey, block.GetHeight(), e.Chain.GetBestView().GetHeight())
					continue
				}

				proposeView := e.Chain.GetViewByHash(block.GetPrevHash())
				if proposeView == nil {
					e.Logger.Infof("%v Request sync block from node %s from %s to %s", e.ChainKey, proposeMsg.PeerID, block.GetPrevHash().String(), block.GetPrevHash().String())
					e.Node.RequestMissingViewViaStream(proposeMsg.PeerID, [][]byte{block.GetPrevHash().Bytes()}, e.Chain.GetShardID(), e.Chain.GetChainName())
				}

			case voteMsg := <-e.VoteMessageCh:
				voteMsg.IsValid = 0
				if b, ok := e.receiveBlockByHash[voteMsg.BlockHash]; ok { //if receiveblock is already initiated
					if _, ok := b.votes[voteMsg.Validator]; !ok { // and not receive validatorA vote
						b.votes[voteMsg.Validator] = &voteMsg // store it
						vid, v := GetValidatorIndex(e.Chain.GetBestView(), voteMsg.Validator)
						if v != nil {
							vbase58, _ := v.ToBase58()
							e.Logger.Infof("%v Receive vote (%d) for block %s from validator %d %v", e.ChainKey, len(e.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, vid, vbase58)
						} else {
							e.Logger.Infof("%v Receive vote (%d) for block from unknown validator", e.ChainKey, len(e.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, voteMsg.Validator)
						}

						b.hasNewVote = true
					}
				} else {
					e.receiveBlockByHash[voteMsg.BlockHash] = &ProposeBlockInfo{
						votes:      make(map[string]*BFTVote),
						hasNewVote: true,
					}
					e.receiveBlockByHash[voteMsg.BlockHash].votes[voteMsg.Validator] = &voteMsg
					vid, v := GetValidatorIndex(e.Chain.GetBestView(), voteMsg.Validator)
					if v != nil {
						vbase58, _ := v.ToBase58()
						e.Logger.Infof("%v Receive vote (%d) for block %s from validator %d %v", e.ChainKey, len(e.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, vid, vbase58)
					} else {
						e.Logger.Infof("%v Receive vote (%d) for block from unknown validator", e.ChainKey, len(e.receiveBlockByHash[voteMsg.BlockHash].votes), voteMsg.BlockHash, voteMsg.Validator)
					}
				}
			case <-cleanMemTicker:
				for h, _ := range e.receiveBlockByHeight {
					if h <= e.Chain.GetFinalView().GetHeight() {
						delete(e.receiveBlockByHeight, h)
					}
				}
				for h, _ := range e.voteHistory {
					if h <= e.Chain.GetFinalView().GetHeight() {
						delete(e.voteHistory, h)
					}
				}
				for h, proposeBlk := range e.receiveBlockByHash {
					if time.Now().Sub(proposeBlk.receiveTime) > time.Minute {
						delete(e.receiveBlockByHash, h)
					}
				}
			case <-ticker:
				if !e.Chain.IsReady() {
					continue
				}
				e.currentTime = time.Now().Unix()

				newTimeSlot := false
				if e.currentTimeSlot != common.CalculateTimeSlot(e.currentTime) {
					newTimeSlot = true
				}

				e.currentTimeSlot = common.CalculateTimeSlot(e.currentTime)
				bestView := e.Chain.GetBestView()

				/*
					Check for whether we should propose block
				*/
				proposerPk, _ := bestView.GetProposerByTimeSlot(e.currentTimeSlot, 2)
				var userProposeKey signatureschemes2.MiningKey
				shouldPropose := false
				shouldListen := true
				for _, userKey := range e.UserKeySet {
					userPk := userKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus)
					if proposerPk.GetMiningKeyBase58(common.BlsConsensus) == userPk {
						shouldListen = false
						if common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()) != e.currentTimeSlot { // current timeslot is not add to view, and this user is proposer of this timeslot
							//using block hash as key of best view -> check if this best view we propose or not
							if _, ok := e.proposeHistory.Get(fmt.Sprintf("%s%d", e.currentTimeSlot)); !ok {
								shouldPropose = true
								userProposeKey = userKey
							}
						}
					}
				}

				if newTimeSlot { //for logging
					e.Logger.Infof("%v", e.ChainKey)
					e.Logger.Infof("%v ======================================================", e.ChainKey)
					e.Logger.Infof("%v", e.ChainKey)
					if shouldListen {
						e.Logger.Infof("%v TS: %v, LISTEN BLOCK %v, Round %v", e.ChainKey, common.CalculateTimeSlot(e.currentTime), bestView.GetHeight()+1, e.currentTimeSlot-common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()))
					}
					if shouldPropose {
						e.Logger.Infof("%v TS: %v, PROPOSE BLOCK %v, Round %v", e.ChainKey, common.CalculateTimeSlot(e.currentTime), bestView.GetHeight()+1, e.currentTimeSlot-common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()))
					}

				}

				if shouldPropose {
					e.proposeHistory.Add(fmt.Sprintf("%s%d", e.currentTimeSlot), 1)
					//Proposer Rule: check propose block connected to bestview(longest chain rule 1) and re-propose valid block with smallest timestamp (including already propose in the past) (rule 2)
					sort.Slice(e.receiveBlockByHeight[bestView.GetHeight()+1], func(i, j int) bool {
						return e.receiveBlockByHeight[bestView.GetHeight()+1][i].block.GetProduceTime() < e.receiveBlockByHeight[bestView.GetHeight()+1][j].block.GetProduceTime()
					})

					var proposeBlock types.BlockInterface = nil
					for _, v := range e.receiveBlockByHeight[bestView.GetHeight()+1] {
						if v.isValid {
							proposeBlock = v.block
							break
						}
					}

					//proposerPk: which include mining pubkey + incokey
					//userKey: only have minigkey
					if createdBlk, err := e.proposeBlock(userProposeKey, proposerPk, proposeBlock); err != nil {
						e.Logger.Critical(UnExpectedError, errors.New("can't propose block"))
						e.Logger.Critical(err)

					} else {
						e.Logger.Infof("%v proposer block %v round %v time slot %v blockTimeSlot %v with hash %v", e.ChainKey, createdBlk.GetHeight(), e.currentTimeSlot-common.CalculateTimeSlot(bestView.GetBlock().GetProposeTime()), e.currentTimeSlot, common.CalculateTimeSlot(createdBlk.GetProduceTime()), createdBlk.Hash().String())
					}
				}

				/*
					Check for valid block to vote
				*/
				validProposeBlock := []*ProposeBlockInfo{}
				//get all block that has height = bestview height  + 1(rule 2 & rule 3) (
				for h, proposeBlockInfo := range e.receiveBlockByHash {
					if proposeBlockInfo.block == nil {
						continue
					}
					// e.Logger.Infof("[Monitor] bestview height %v, finalview height %v, block height %v %v", bestViewHeight, e.Chain.GetFinalView().GetHeight(), proposeBlockInfo.block.GetHeight(), proposeBlockInfo.block.GetProduceTime())
					// check if propose block in current time
					if e.currentTimeSlot == common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) {
						validProposeBlock = append(validProposeBlock, proposeBlockInfo)
					}

					if proposeBlockInfo.block.GetHeight() < e.Chain.GetFinalView().GetHeight() {
						delete(e.receiveBlockByHash, h)
					}
				}
				//rule 1: get history of vote for this height, vote if (round is lower than the vote before) or (round is equal but new proposer) or (there is no vote for this height yet)
				sort.Slice(validProposeBlock, func(i, j int) bool {
					return validProposeBlock[i].block.GetProduceTime() < validProposeBlock[j].block.GetProduceTime()
				})

				for _, v := range validProposeBlock {
					if v.sendVote {
						continue
					}

					blkCreateTimeSlot := common.CalculateTimeSlot(v.block.GetProduceTime())
					bestViewHeight := bestView.GetHeight()

					if lastVotedBlk, ok := e.voteHistory[bestViewHeight+1]; ok {
						if blkCreateTimeSlot < common.CalculateTimeSlot(lastVotedBlk.GetProduceTime()) { //blkCreateTimeSlot is smaller than voted block => vote for this blk
							e.validateAndVote(v)
						} else if blkCreateTimeSlot == common.CalculateTimeSlot(lastVotedBlk.GetProduceTime()) && common.CalculateTimeSlot(v.block.GetProposeTime()) > common.CalculateTimeSlot(lastVotedBlk.GetProposeTime()) { //blk is old block (same round), but new proposer(larger timeslot) => vote again
							e.validateAndVote(v)
						} //blkCreateTimeSlot is larger or equal than voted block => do nothing
					} else { //there is no vote for this height yet
						e.validateAndVote(v)
					}
				}

				/*
					Check for 2/3 vote to commit
				*/
				for k, v := range e.receiveBlockByHash {
					e.processIfBlockGetEnoughVote(k, v)
				}

			}
		}
	}()
	return nil
}
