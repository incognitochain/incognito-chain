package mubft

import (
	"bytes"
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
)

func (protocol *BFTProtocol) phasePropose() error {
	go protocol.CreateBlockMsg()
	phaseDuration := getTimeout(protocol.phase, len(protocol.RoundData.Committee))
	timeout := time.AfterFunc(phaseDuration, func() {
		fmt.Println("BFT: Propose phase timeout", time.Since(protocol.startTime).Seconds())
		protocol.closeTimeoutCh()
	})
	timeout2 := time.AfterFunc(phaseDuration/2, func() {
		fmt.Println("BFT: Request ready msg", time.Since(protocol.startTime).Seconds())
		if protocol.RoundData.Layer == common.BEACON_ROLE {
			msgReq, _ := MakeMsgBFTReq(protocol.RoundData.BestStateHash, protocol.RoundData.Round, protocol.EngineCfg.UserKeySet)
			if err := protocol.EngineCfg.Server.PushMessageToBeacon(msgReq, map[libp2p.ID]bool{}); err != nil {
				fmt.Println("BFT: no beacon", err)
			}
		} else {
			msgReq, _ := MakeMsgBFTReq(protocol.RoundData.BestStateHash, protocol.RoundData.Round, protocol.EngineCfg.UserKeySet)
			if err := protocol.EngineCfg.Server.PushMessageToShard(msgReq, protocol.RoundData.ShardID, map[libp2p.ID]bool{}); err != nil {
				fmt.Println("BFT: no shard", err)
			}
		}
	})

	readyMsgs := make(map[string]*wire.MessageBFTReady)

	fmt.Println("BFT: Listen for ready msg", time.Since(protocol.startTime).Seconds())
phase:
	for {
		select {
		case msgReady := <-protocol.cBFTMsg:
			if msgReady.MessageType() == wire.CmdBFTReady {

				isMatchBestState := msgReady.(*wire.MessageBFTReady).BestStateHash == protocol.RoundData.BestStateHash
				isMatchRound := msgReady.(*wire.MessageBFTReady).Round == protocol.RoundData.Round
				isCommittee := common.IndexOfStr(msgReady.(*wire.MessageBFTReady).Pubkey, protocol.RoundData.Committee) != -1

				fmt.Println("BFT: pro ", isMatchBestState, isMatchRound, isCommittee, time.Now().Unix(), protocol.RoundData.BestStateHash, msgReady.(*wire.MessageBFTReady).BestStateHash, protocol.RoundData.Round, msgReady.(*wire.MessageBFTReady).Round)

				if isMatchBestState && isMatchRound && isCommittee {
					readyMsgs[msgReady.(*wire.MessageBFTReady).Pubkey] = msgReady.(*wire.MessageBFTReady)
					if len(readyMsgs) >= (2*len(protocol.RoundData.Committee)/3)-1 {
						timeout.Stop()
						timeout2.Stop()
						fmt.Println("BFT: Collected enough ready", time.Since(protocol.startTime).Seconds())
						protocol.closeTimeoutCh()
					}
				}
			}
		case <-protocol.cTimeout:
			if len(readyMsgs) >= (2*len(protocol.RoundData.Committee)/3)-1 {
				fmt.Println("BFT: Propose block", time.Since(protocol.startTime).Seconds())

				msg := <-protocol.proposeCh
				if msg == nil {
					return errors.New("Failed to propose block")
				}
				protocol.forwardMsg(msg)
				protocol.phase = BFT_AGREE
				protocol.closeProposeCh()
			} else {
				protocol.closeProposeCh()
				fmt.Println("BFT: Didn't received enough ready msg", time.Since(protocol.startTime).Seconds())
				return errors.New("Didn't received enough ready msg")
			}
			break phase
		}
	}

	return nil
}

func (protocol *BFTProtocol) phaseListen() error {
	if protocol.RoundData.Layer == common.BEACON_ROLE {
		msgReady, _ := MakeMsgBFTReady(protocol.RoundData.BestStateHash, protocol.RoundData.Round, protocol.EngineCfg.ShardToBeaconPool.GetLatestValidPendingBlockHeight(), protocol.EngineCfg.UserKeySet)
		protocol.EngineCfg.Server.PushMessageToBeacon(msgReady, map[libp2p.ID]bool{})
	} else {
		msgReady, _ := MakeMsgBFTReady(protocol.RoundData.BestStateHash, protocol.RoundData.Round, protocol.EngineCfg.CrossShardPool[protocol.RoundData.ShardID].GetLatestValidBlockHeight(), protocol.EngineCfg.UserKeySet)
		protocol.EngineCfg.Server.PushMessageToShard(msgReady, protocol.RoundData.ShardID, map[libp2p.ID]bool{})
	}

	var timeSinceLastBlk time.Duration
	additionalWaitTime := timeSinceLastBlk
	if protocol.RoundData.Layer == common.BEACON_ROLE {
		timeSinceLastBlk = time.Since(time.Unix(protocol.EngineCfg.BlockChain.BestState.Beacon.BestBlock.Header.Timestamp, 0))
		additionalWaitTime = common.MinBeaconBlkInterval - timeSinceLastBlk
	} else {
		timeSinceLastBlk = time.Since(time.Unix(protocol.EngineCfg.BlockChain.BestState.Shard[protocol.RoundData.ShardID].BestBlock.Header.Timestamp, 0))
		additionalWaitTime = common.MinShardBlkInterval - timeSinceLastBlk
	}
	if additionalWaitTime < 0 {
		additionalWaitTime = 0
	}
	if protocol.RoundData.Layer == common.BEACON_ROLE {
		additionalWaitTime += common.MinBeaconBlkInterval
	} else {
		additionalWaitTime += common.MinShardBlkInterval
	}
	fmt.Println("BFT: Listen phase", time.Since(protocol.startTime).Seconds())

	phaseDuration := getTimeout(protocol.phase, len(protocol.RoundData.Committee))
	timeout := time.AfterFunc(phaseDuration+additionalWaitTime, func() {
		fmt.Println("BFT: Listen phase timeout", time.Since(protocol.startTime).Seconds())
		protocol.closeTimeoutCh()
	})

phase:
	for {
		select {
		case <-protocol.cTimeout:
			return errors.New("Listen phase timeout")
		case msg := <-protocol.cBFTMsg:
			if msg.MessageType() == wire.CmdBFTPropose {
				fmt.Println("BFT: Propose block received", time.Since(protocol.startTime).Seconds())
				protocol.forwardMsg(msg)
				if protocol.RoundData.Layer == common.BEACON_ROLE {
					pendingBlk := blockchain.BeaconBlock{}
					err := pendingBlk.UnmarshalJSON(msg.(*wire.MessageBFTPropose).Block)
					if err != nil {
						Logger.log.Error(err)
						continue
					}
					verifyTime := time.Now()
					err = protocol.EngineCfg.BlockChain.VerifyPreSignBeaconBlock(&pendingBlk, true)
					if err != nil {
						fmt.Println("BFT: verify beaconblk err:", err)
						Logger.log.Error(err)
						continue
					}
					protocol.blockCreateTime = time.Since(verifyTime)
					protocol.pendingBlock = &pendingBlk
					protocol.multiSigScheme.dataToSig = pendingBlk.Header.Hash()
				} else {
					pendingBlk := blockchain.ShardBlock{}
					err := pendingBlk.UnmarshalJSON(msg.(*wire.MessageBFTPropose).Block)
					if err != nil {
						Logger.log.Error(err)
						continue
					}
					verifyTime := time.Now()
					err = protocol.EngineCfg.BlockChain.VerifyPreSignShardBlock(&pendingBlk, protocol.RoundData.ShardID)
					if err != nil {
						Logger.log.Error(err)
						continue
					}
					protocol.blockCreateTime = time.Since(verifyTime)

					protocol.pendingBlock = &pendingBlk
					protocol.multiSigScheme.dataToSig = pendingBlk.Header.Hash()
				}
				fmt.Println("BFT: Forward propose message", time.Since(protocol.startTime).Seconds())
				protocol.phase = BFT_AGREE
				timeout.Stop()
				break phase
			} else {
				if msg.MessageType() == wire.CmdBFTReq {
					go func() {
						isMatchBeststate := msg.(*wire.MessageBFTReq).BestStateHash == protocol.RoundData.BestStateHash
						isMatchRound := msg.(*wire.MessageBFTReq).Round == protocol.RoundData.Round
						isCommitee := common.IndexOfStr(msg.(*wire.MessageBFTReq).Pubkey, protocol.RoundData.Committee) != -1
						fmt.Println("BFT: val ", isMatchBeststate, isMatchRound, isCommitee, time.Now().Unix(), protocol.RoundData.BestStateHash, msg.(*wire.MessageBFTReq).BestStateHash, blockchain.GetBeaconBestState().BeaconHeight)
						if isMatchBeststate && isMatchRound && isCommitee {
							if protocol.RoundData.Layer == common.BEACON_ROLE {
								if userRole, _ := protocol.EngineCfg.BlockChain.BestState.Beacon.GetPubkeyRole(msg.(*wire.MessageBFTReq).Pubkey, protocol.RoundData.Round); userRole == common.PROPOSER_ROLE {
									msgReady, _ := MakeMsgBFTReady(protocol.RoundData.BestStateHash, protocol.RoundData.Round, protocol.EngineCfg.ShardToBeaconPool.GetLatestValidPendingBlockHeight(), protocol.EngineCfg.UserKeySet)
									protocol.EngineCfg.Server.PushMessageToBeacon(msgReady, map[libp2p.ID]bool{})
								}
							} else {
								if userRole := protocol.EngineCfg.BlockChain.BestState.Shard[protocol.RoundData.ShardID].GetPubkeyRole(msg.(*wire.MessageBFTReq).Pubkey, protocol.RoundData.Round); userRole == common.PROPOSER_ROLE {
									msgReady, _ := MakeMsgBFTReady(protocol.RoundData.BestStateHash, protocol.RoundData.Round, protocol.EngineCfg.CrossShardPool[protocol.RoundData.ShardID].GetLatestValidBlockHeight(), protocol.EngineCfg.UserKeySet)
									protocol.EngineCfg.Server.PushMessageToShard(msgReady, protocol.RoundData.ShardID, map[libp2p.ID]bool{})
								}
							}
						}
					}()
				} else {
					go func() {
						protocol.earlyMsgCh <- msg
					}()
				}
			}
		}
	}

	return nil
}

func (protocol *BFTProtocol) phaseAgree() error {
	fmt.Println("BFT: Agree phase", time.Since(protocol.startTime).Seconds())
	phaseDuration := getTimeout(protocol.phase, len(protocol.RoundData.Committee))
	timeout := time.AfterFunc(phaseDuration+(protocol.blockCreateTime*4/5), func() {
		fmt.Println("BFT: Agree phase timeout", time.Since(protocol.startTime).Seconds())
		protocol.closeTimeoutCh()
	})

	fmt.Println("BFT: Sending out Agree msg", time.Since(protocol.startTime).Seconds())
	msg, err := MakeMsgBFTAgree(protocol.multiSigScheme.personal.Ri, protocol.EngineCfg.UserKeySet, protocol.multiSigScheme.dataToSig)
	if err != nil {
		Logger.log.Error(err)
		return err
	}
	protocol.forwardMsg(msg)

	//map of members and their Ri
	collectedRiList := make(map[string][]byte)
	collectedRiList[protocol.EngineCfg.UserKeySet.GetPublicKeyInBase58CheckEncode()] = protocol.multiSigScheme.personal.Ri
phase:
	for {
		select {
		case <-protocol.cTimeout:
			//Use collected Ri to calc r & get ValidatorsIndex if len(Ri) > 1/2size(committee)
			// then sig block with this r
			if len(collectedRiList) < (2 * len(protocol.RoundData.Committee) / 3) {
				fmt.Println("BFT: Didn't receive enough Ri to continue", time.Since(protocol.startTime).Seconds())
				return errors.New("Didn't receive enough Ri to continue")
			}
			err := protocol.multiSigScheme.SignData(collectedRiList)
			if err != nil {
				return err
			}

			protocol.phase = BFT_COMMIT
			break phase
		case msg := <-protocol.cBFTMsg:
			if msg.MessageType() == wire.CmdBFTAgree {
				fmt.Println("BFT: Agree msg received", time.Since(protocol.startTime).Seconds())
				if common.IndexOfStr(msg.(*wire.MessageBFTAgree).Pubkey, protocol.RoundData.Committee) >= 0 && bytes.Equal(protocol.multiSigScheme.dataToSig[:], msg.(*wire.MessageBFTAgree).BlkHash[:]) {
					if _, ok := collectedRiList[msg.(*wire.MessageBFTAgree).Pubkey]; !ok {
						collectedRiList[msg.(*wire.MessageBFTAgree).Pubkey] = msg.(*wire.MessageBFTAgree).Ri
						protocol.forwardMsg(msg)
						if len(collectedRiList) == len(protocol.RoundData.Committee) {
							fmt.Println("BFT: Collected enough Ri", time.Since(protocol.startTime).Seconds())
							timeout.Stop()
							protocol.closeTimeoutCh()
						}
					}
				}
			} else {
				go func() {
					protocol.earlyMsgCh <- msg
				}()
			}
		}
	}

	return nil
}

func (protocol *BFTProtocol) phaseCommit() error {
	fmt.Println("BFT: Commit phase", time.Since(protocol.startTime).Seconds())
	phaseDuration := getTimeout(protocol.phase, len(protocol.RoundData.Committee))
	cmTimeout := time.AfterFunc(phaseDuration, func() {
		fmt.Println("BFT: Commit phase timeout", time.Since(protocol.startTime).Seconds())
		protocol.closeTimeoutCh()
	})

	msg, err := MakeMsgBFTCommit(protocol.multiSigScheme.combine.CommitSig, protocol.multiSigScheme.combine.R, protocol.multiSigScheme.combine.ValidatorsIdxR, protocol.EngineCfg.UserKeySet)
	if err != nil {
		Logger.log.Error(err)
		return err
	}
	protocol.forwardMsg(msg)

	var phaseData struct {
		Sigs map[string]map[string]bftCommittedSig //map[R]map[Pubkey]CommittedSig
	}

	phaseData.Sigs = make(map[string]map[string]bftCommittedSig)
	phaseData.Sigs[protocol.multiSigScheme.combine.R] = make(map[string]bftCommittedSig)
	phaseData.Sigs[protocol.multiSigScheme.combine.R][protocol.EngineCfg.UserKeySet.GetPublicKeyInBase58CheckEncode()] = bftCommittedSig{
		Sig:            protocol.multiSigScheme.combine.CommitSig,
		ValidatorsIdxR: protocol.multiSigScheme.combine.ValidatorsIdxR,
	}
phase:
	for {
		select {
		case <-protocol.cTimeout:
			//Combine collected Sigs with the same r that has the longest list must has size > 1/2size(committee)
			var szRCombined string
			szRCombined = "1"
			for szR := range phaseData.Sigs {
				if len(phaseData.Sigs[szR]) > (2 * len(protocol.RoundData.Committee) / 3) {
					if len(szRCombined) == 1 {
						szRCombined = szR
					} else {
						if len(phaseData.Sigs[szR]) > len(phaseData.Sigs[szRCombined]) {
							szRCombined = szR
						}
					}
				}
			}
			if len(szRCombined) == 1 {
				fmt.Println("PhaseData.Sigs: ", phaseData.Sigs)
				fmt.Println("Length of phaseData.Sigs: ", len(phaseData.Sigs))
				fmt.Println("BFT: Not enough sigs to combine", time.Since(protocol.startTime).Seconds())
				return errors.New("Not enough sigs to combine")
			}

			AggregatedSig, err := protocol.multiSigScheme.CombineSigs(szRCombined, phaseData.Sigs[szRCombined])
			if err != nil {
				return err
			}
			ValidatorsIdxAggSig := make([]int, len(protocol.multiSigScheme.combine.ValidatorsIdxAggSig))
			ValidatorsIdxR := make([]int, len(protocol.multiSigScheme.combine.ValidatorsIdxR))
			copy(ValidatorsIdxAggSig, protocol.multiSigScheme.combine.ValidatorsIdxAggSig)
			copy(ValidatorsIdxR, protocol.multiSigScheme.combine.ValidatorsIdxR)

			// fmt.Println("BFT: \n \n Block consensus reach", ValidatorsIdxR, ValidatorsIdxAggSig, AggregatedSig)

			if protocol.RoundData.Layer == common.BEACON_ROLE {
				protocol.pendingBlock.(*blockchain.BeaconBlock).R = protocol.multiSigScheme.combine.R
				protocol.pendingBlock.(*blockchain.BeaconBlock).AggregatedSig = AggregatedSig
				protocol.pendingBlock.(*blockchain.BeaconBlock).ValidatorsIndex = make([][]int, 2)
				protocol.pendingBlock.(*blockchain.BeaconBlock).ValidatorsIndex[0] = make([]int, len(ValidatorsIdxR))
				protocol.pendingBlock.(*blockchain.BeaconBlock).ValidatorsIndex[1] = make([]int, len(ValidatorsIdxAggSig))
				copy(protocol.pendingBlock.(*blockchain.BeaconBlock).ValidatorsIndex[0], ValidatorsIdxR)
				copy(protocol.pendingBlock.(*blockchain.BeaconBlock).ValidatorsIndex[1], ValidatorsIdxAggSig)
			} else {
				protocol.pendingBlock.(*blockchain.ShardBlock).R = protocol.multiSigScheme.combine.R
				protocol.pendingBlock.(*blockchain.ShardBlock).AggregatedSig = AggregatedSig
				protocol.pendingBlock.(*blockchain.ShardBlock).ValidatorsIndex = make([][]int, 2)
				protocol.pendingBlock.(*blockchain.ShardBlock).ValidatorsIndex[0] = make([]int, len(ValidatorsIdxR))
				protocol.pendingBlock.(*blockchain.ShardBlock).ValidatorsIndex[1] = make([]int, len(ValidatorsIdxAggSig))
				copy(protocol.pendingBlock.(*blockchain.ShardBlock).ValidatorsIndex[0], ValidatorsIdxR)
				copy(protocol.pendingBlock.(*blockchain.ShardBlock).ValidatorsIndex[1], ValidatorsIdxAggSig)
			}
			break phase
		case msgCommit := <-protocol.cBFTMsg:
			if msgCommit.MessageType() == wire.CmdBFTCommit {
				newSig := bftCommittedSig{
					ValidatorsIdxR: msgCommit.(*wire.MessageBFTCommit).ValidatorsIdx,
					Sig:            msgCommit.(*wire.MessageBFTCommit).CommitSig,
				}
				R := msgCommit.(*wire.MessageBFTCommit).R
				err := protocol.multiSigScheme.VerifyCommitSig(msgCommit.(*wire.MessageBFTCommit).Pubkey, newSig.Sig, R, newSig.ValidatorsIdxR)
				if err != nil {
					Logger.log.Error(err)
					continue
				}
				fmt.Println("BFT: Commit msg received", time.Since(protocol.startTime).Seconds())
				if _, ok := phaseData.Sigs[R]; !ok {
					phaseData.Sigs[R] = make(map[string]bftCommittedSig)
				}
				if _, ok := phaseData.Sigs[R][msgCommit.(*wire.MessageBFTCommit).Pubkey]; !ok {
					phaseData.Sigs[R][msgCommit.(*wire.MessageBFTCommit).Pubkey] = newSig
					protocol.forwardMsg(msgCommit)
					if len(phaseData.Sigs[R]) > (2 * len(protocol.RoundData.Committee) / 3) {
						cmTimeout.Stop()
						fmt.Println("BFT: Collected enough Sig", time.Since(protocol.startTime).Seconds())
						protocol.closeTimeoutCh()
					}
				}
			}
		}
	}
	return nil
}
