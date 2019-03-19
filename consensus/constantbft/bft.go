package constantbft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/constant-money/constant-chain/common"

	"github.com/constant-money/constant-chain/cashec"

	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/wire"
)

type BFTProtocol struct {
	cBFTMsg   chan wire.Message
	EngineCfg *EngineConfig

	ShardToBeaconPool blockchain.ShardToBeaconPool
	CrossShardPool    map[byte]blockchain.CrossShardPool
	BlockGen          *blockchain.BlkTmplGenerator
	BlockChain        *blockchain.BlockChain
	Server            serverInterface
	UserKeySet        *cashec.KeySet

	cQuit    chan struct{}
	cTimeout chan struct{}

	phase string

	pendingBlock interface{}

	RoundData struct {
		BestStateHash    common.Hash
		ProposerOffset   int
		IsProposer       bool
		Layer            string
		ShardID          byte
		Committee        []string
		ClosestPoolState map[byte]uint64
	}
	multiSigScheme *multiSigScheme
}

func (protocol *BFTProtocol) Start() (interface{}, error) {
	protocol.phase = PBFT_LISTEN
	if protocol.RoundData.IsProposer {
		protocol.phase = PBFT_PROPOSE
	}

	Logger.log.Info("Starting PBFT protocol for " + protocol.RoundData.Layer)
	protocol.multiSigScheme = new(multiSigScheme)
	protocol.multiSigScheme.Init(protocol.UserKeySet, protocol.RoundData.Committee)
	err := protocol.multiSigScheme.Prepare()

	if err != nil {
		return nil, err
	}

	for {
		fmt.Println("BFT: New Phase")
		protocol.cTimeout = make(chan struct{})
		select {
		case <-protocol.cQuit:
			return nil, errors.New("Consensus quit")
		default:
			switch protocol.phase {
			case PBFT_PROPOSE:
				//    single-node start    //
				// time.Sleep(2 * time.Second)
				// _, err := protocol.CreateBlockMsg()
				// if err != nil {
				// 	return nil, err
				// }
				// return protocol.pendingBlock, nil
				//    single-node end    //
				timeout := time.AfterFunc(ListenTimeout*time.Second, func() {
					fmt.Println("BFT: Propose phase timeout")
					protocol.closeTimeoutCh()
				})
				timeout2 := time.AfterFunc((ListenTimeout/2)*time.Second, func() {
					fmt.Println("BFT: Request ready msg")
					if protocol.RoundData.Layer == common.BEACON_ROLE {
						msgReq, _ := MakeMsgBFTReq(protocol.RoundData.BestStateHash, protocol.RoundData.ProposerOffset, protocol.UserKeySet)
						if err := protocol.Server.PushMessageToBeacon(msgReq); err != nil {
							fmt.Println("BFT: no beacon", err)
						}
					} else {
						msgReq, _ := MakeMsgBFTReq(protocol.RoundData.BestStateHash, protocol.RoundData.ProposerOffset, protocol.UserKeySet)
						if err := protocol.Server.PushMessageToShard(msgReq, protocol.RoundData.ShardID); err != nil {
							fmt.Println("BFT: no beacon", err)
						}
					}
				})

				var readyMsgs map[string]*wire.MessageBFTReady
				readyMsgs = make(map[string]*wire.MessageBFTReady)

				fmt.Println()
				fmt.Println("BFT: Listen for ready msg")
				fmt.Println()
			proposephase:
				for {
					select {
					case <-protocol.cTimeout:
						if len(readyMsgs) >= (2*len(protocol.RoundData.Committee)/3)-1 {
							if protocol.RoundData.Layer == common.BEACON_ROLE {
								var shToBcPoolStates []map[byte]uint64
								for _, readyMsg := range readyMsgs {
									shToBcPoolStates = append(shToBcPoolStates, readyMsg.PoolState)
								}
								shToBcPoolStates = append(shToBcPoolStates, protocol.ShardToBeaconPool.GetLatestValidPendingBlockHeight())
								protocol.RoundData.ClosestPoolState = GetClosestPoolState(shToBcPoolStates)
							} else {
								var crossShardsPoolStates []map[byte]uint64
								for _, readyMsg := range readyMsgs {
									crossShardsPoolStates = append(crossShardsPoolStates, readyMsg.PoolState)
								}
								crossShardsPoolStates = append(crossShardsPoolStates, protocol.CrossShardPool[protocol.RoundData.ShardID].GetLatestValidBlockHeight())
								protocol.RoundData.ClosestPoolState = GetClosestPoolState(crossShardsPoolStates)
							}

							fmt.Println("BFT: Propose block")
							msg, err := protocol.CreateBlockMsg()
							if err != nil {
								return nil, err
							}
							protocol.forwardMsg(msg)
							protocol.phase = PBFT_PREPARE
						} else {
							return nil, errors.New("Didn't received enough ready msg")
						}
						break proposephase
					case msgReady := <-protocol.cBFTMsg:
						if msgReady.MessageType() == wire.CmdBFTReady {
							fmt.Println("BFT: pro ", protocol.RoundData.BestStateHash, msgReady.(*wire.MessageBFTReady).BestStateHash, blockchain.GetBestStateBeacon().BeaconHeight)
							if msgReady.(*wire.MessageBFTReady).BestStateHash == protocol.RoundData.BestStateHash && msgReady.(*wire.MessageBFTReady).ProposerOffset == protocol.RoundData.ProposerOffset && common.IndexOfStr(msgReady.(*wire.MessageBFTReady).Pubkey, protocol.RoundData.Committee) != -1 {
								readyMsgs[msgReady.(*wire.MessageBFTReady).Pubkey] = msgReady.(*wire.MessageBFTReady)
								if len(readyMsgs) >= (2*len(protocol.RoundData.Committee)/3)-1 {
									timeout.Stop()
									timeout2.Stop()
									fmt.Println("BFT: Collected enough ready")
									protocol.closeTimeoutCh()
								}
							}
						}
					}
				}
			case PBFT_LISTEN:
				if protocol.RoundData.Layer == common.BEACON_ROLE {
					msgReady, _ := MakeMsgBFTReady(protocol.RoundData.BestStateHash, protocol.RoundData.ProposerOffset, protocol.ShardToBeaconPool.GetLatestValidPendingBlockHeight(), protocol.UserKeySet)
					protocol.Server.PushMessageToBeacon(msgReady)
				} else {
					msgReady, _ := MakeMsgBFTReady(protocol.RoundData.BestStateHash, protocol.RoundData.ProposerOffset, protocol.CrossShardPool[protocol.RoundData.ShardID].GetLatestValidBlockHeight(), protocol.UserKeySet)
					protocol.Server.PushMessageToShard(msgReady, protocol.RoundData.ShardID)
				}
				fmt.Println("BFT: Listen phase")
				timeout := time.AfterFunc(ListenTimeout*time.Second, func() {
					fmt.Println("BFT: Listen phase timeout")
					protocol.closeTimeoutCh()
				})
			listenphase:
				for {
					select {
					case msgPropose := <-protocol.cBFTMsg:
						if msgPropose.MessageType() == wire.CmdBFTPropose {
							fmt.Println("BFT: Propose block received")
							if protocol.RoundData.Layer == common.BEACON_ROLE {
								pendingBlk := blockchain.BeaconBlock{}
								pendingBlk.UnmarshalJSON(msgPropose.(*wire.MessageBFTPropose).Block)

								err = protocol.BlockChain.VerifyPreSignBeaconBlock(&pendingBlk, true)
								if err != nil {
									Logger.log.Error(err)
									continue
								}
								protocol.pendingBlock = &pendingBlk
								protocol.multiSigScheme.dataToSig = pendingBlk.Header.Hash()
							} else {
								pendingBlk := blockchain.ShardBlock{}
								pendingBlk.UnmarshalJSON(msgPropose.(*wire.MessageBFTPropose).Block)
								err = protocol.BlockChain.VerifyPreSignShardBlock(&pendingBlk, protocol.RoundData.ShardID)
								if err != nil {
									Logger.log.Error(err)
									continue
								}
								protocol.pendingBlock = &pendingBlk
								protocol.multiSigScheme.dataToSig = pendingBlk.Header.Hash()
							}
							protocol.forwardMsg(msgPropose)
							protocol.phase = PBFT_PREPARE
							timeout.Stop()
							break listenphase
						} else {
							if msgPropose.MessageType() == wire.CmdBFTReq {
								fmt.Println("BFT: val ", protocol.RoundData.BestStateHash, msgPropose.(*wire.MessageBFTReq).BestStateHash, blockchain.GetBestStateBeacon().BeaconHeight)
								go func() {
									if msgPropose.(*wire.MessageBFTReq).BestStateHash == protocol.RoundData.BestStateHash && msgPropose.(*wire.MessageBFTReq).ProposerOffset == protocol.RoundData.ProposerOffset && common.IndexOfStr(msgPropose.(*wire.MessageBFTReq).Pubkey, protocol.RoundData.Committee) != -1 {
										if protocol.RoundData.Layer == common.BEACON_ROLE {
											if userRole, _ := protocol.BlockChain.BestState.Beacon.GetPubkeyRole(msgPropose.(*wire.MessageBFTReq).Pubkey, protocol.RoundData.ProposerOffset); userRole == common.PROPOSER_ROLE {
												msgReady, _ := MakeMsgBFTReady(protocol.RoundData.BestStateHash, protocol.RoundData.ProposerOffset, protocol.ShardToBeaconPool.GetLatestValidPendingBlockHeight(), protocol.UserKeySet)
												protocol.Server.PushMessageToBeacon(msgReady)
											}
										} else {
											if userRole := protocol.BlockChain.BestState.Shard[protocol.RoundData.ShardID].GetPubkeyRole(msgPropose.(*wire.MessageBFTReq).Pubkey, protocol.RoundData.ProposerOffset); userRole == common.PROPOSER_ROLE {
												msgReady, _ := MakeMsgBFTReady(protocol.RoundData.BestStateHash, protocol.RoundData.ProposerOffset, nil, protocol.UserKeySet)
												protocol.Server.PushMessageToShard(msgReady, protocol.RoundData.ShardID)
											}
										}
									}
								}()
							}
						}

					case <-protocol.cTimeout:
						return nil, errors.New("Listen phase timeout")
					}
				}
			case PBFT_PREPARE:
				fmt.Println("BFT: Prepare phase")
				timeout := time.AfterFunc(PrepareTimeout*time.Second, func() {
					fmt.Println("BFT: Prepare phase timeout")
					protocol.closeTimeoutCh()
				})
				time.AfterFunc(DelayTime*time.Millisecond, func() {
					fmt.Println("BFT: Sending out prepare msg")
					msg, err := MakeMsgBFTPrepare(protocol.multiSigScheme.personal.Ri, protocol.UserKeySet, protocol.multiSigScheme.dataToSig)
					if err != nil {
						Logger.log.Error(err)
						return
					}
					protocol.forwardMsg(msg)
				})

				var collectedRiList map[string][]byte //map of members and their Ri
				collectedRiList = make(map[string][]byte)
				collectedRiList[protocol.UserKeySet.GetPublicKeyB58()] = protocol.multiSigScheme.personal.Ri
			preparephase:
				for {
					select {
					case <-protocol.cTimeout:
						//Use collected Ri to calc r & get ValidatorsIdx if len(Ri) > 1/2size(committee)
						// then sig block with this r
						if len(collectedRiList) < (len(protocol.RoundData.Committee) >> 1) {
							return nil, errors.New("Didn't receive enough Ri to continue")
						}
						err := protocol.multiSigScheme.SignData(collectedRiList)
						if err != nil {
							return nil, err
						}

						protocol.phase = PBFT_COMMIT
						break preparephase
					case msgPrepare := <-protocol.cBFTMsg:
						if msgPrepare.MessageType() == wire.CmdBFTPrepare {
							fmt.Println("BFT: Prepare msg received")
							if common.IndexOfStr(msgPrepare.(*wire.MessageBFTPrepare).Pubkey, protocol.RoundData.Committee) >= 0 && bytes.Compare(protocol.multiSigScheme.dataToSig[:], msgPrepare.(*wire.MessageBFTPrepare).BlkHash[:]) == 0 {
								if _, ok := collectedRiList[msgPrepare.(*wire.MessageBFTPrepare).Pubkey]; !ok {
									collectedRiList[msgPrepare.(*wire.MessageBFTPrepare).Pubkey] = msgPrepare.(*wire.MessageBFTPrepare).Ri
									protocol.forwardMsg(msgPrepare)
									if len(collectedRiList) == len(protocol.RoundData.Committee) {
										fmt.Println("BFT: Collected enough Ri")
										timeout.Stop()
										protocol.closeTimeoutCh()
									}
								}
							}
						}
					}
				}
			case PBFT_COMMIT:
				fmt.Println("BFT: Commit phase")
				cmTimeout := time.AfterFunc(CommitTimeout*time.Second, func() {
					fmt.Println("BFT: Commit phase timeout")
					protocol.closeTimeoutCh()
				})

				time.AfterFunc(DelayTime*time.Millisecond, func() {
					msg, err := MakeMsgBFTCommit(protocol.multiSigScheme.combine.CommitSig, protocol.multiSigScheme.combine.R, protocol.multiSigScheme.combine.ValidatorsIdxR, protocol.UserKeySet)
					if err != nil {
						Logger.log.Error(err)
						return
					}
					fmt.Println("BFT: Sending out commit msg")
					protocol.forwardMsg(msg)
				})
				var phaseData struct {
					Sigs map[string]map[string]bftCommittedSig //map[R]map[Pubkey]CommittedSig
				}

				phaseData.Sigs = make(map[string]map[string]bftCommittedSig)
				phaseData.Sigs[protocol.multiSigScheme.combine.R] = make(map[string]bftCommittedSig)
				phaseData.Sigs[protocol.multiSigScheme.combine.R][protocol.UserKeySet.GetPublicKeyB58()] = bftCommittedSig{
					Sig:            protocol.multiSigScheme.combine.CommitSig,
					ValidatorsIdxR: protocol.multiSigScheme.combine.ValidatorsIdxR,
				}
				// commitphase:
				for {
					select {
					case <-protocol.cTimeout:
						//Combine collected Sigs with the same r that has the longest list must has size > 1/2size(committee)
						var szRCombined string
						szRCombined = "1"
						for szR := range phaseData.Sigs {
							if len(phaseData.Sigs[szR]) > (len(protocol.RoundData.Committee) >> 1) {
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
							return nil, errors.New("Not enough sigs to combine")
						}

						AggregatedSig, err := protocol.multiSigScheme.CombineSigs(szRCombined, phaseData.Sigs[szRCombined])
						if err != nil {
							return nil, err
						}
						ValidatorsIdxAggSig := make([]int, len(protocol.multiSigScheme.combine.ValidatorsIdxAggSig))
						ValidatorsIdxR := make([]int, len(protocol.multiSigScheme.combine.ValidatorsIdxR))
						copy(ValidatorsIdxAggSig, protocol.multiSigScheme.combine.ValidatorsIdxAggSig)
						copy(ValidatorsIdxR, protocol.multiSigScheme.combine.ValidatorsIdxR)

						// fmt.Println("BFT: \n \n Block consensus reach", ValidatorsIdxR, ValidatorsIdxAggSig, AggregatedSig)

						if protocol.RoundData.Layer == common.BEACON_ROLE {
							protocol.pendingBlock.(*blockchain.BeaconBlock).R = protocol.multiSigScheme.combine.R
							protocol.pendingBlock.(*blockchain.BeaconBlock).AggregatedSig = AggregatedSig
							protocol.pendingBlock.(*blockchain.BeaconBlock).ValidatorsIdx = make([][]int, 2)
							protocol.pendingBlock.(*blockchain.BeaconBlock).ValidatorsIdx[0] = make([]int, len(ValidatorsIdxR))
							protocol.pendingBlock.(*blockchain.BeaconBlock).ValidatorsIdx[1] = make([]int, len(ValidatorsIdxAggSig))
							copy(protocol.pendingBlock.(*blockchain.BeaconBlock).ValidatorsIdx[0], ValidatorsIdxR)
							copy(protocol.pendingBlock.(*blockchain.BeaconBlock).ValidatorsIdx[1], ValidatorsIdxAggSig)
						} else {
							protocol.pendingBlock.(*blockchain.ShardBlock).R = protocol.multiSigScheme.combine.R
							protocol.pendingBlock.(*blockchain.ShardBlock).AggregatedSig = AggregatedSig
							protocol.pendingBlock.(*blockchain.ShardBlock).ValidatorsIdx = make([][]int, 2)
							protocol.pendingBlock.(*blockchain.ShardBlock).ValidatorsIdx[0] = make([]int, len(ValidatorsIdxR))
							protocol.pendingBlock.(*blockchain.ShardBlock).ValidatorsIdx[1] = make([]int, len(ValidatorsIdxAggSig))
							copy(protocol.pendingBlock.(*blockchain.ShardBlock).ValidatorsIdx[0], ValidatorsIdxR)
							copy(protocol.pendingBlock.(*blockchain.ShardBlock).ValidatorsIdx[1], ValidatorsIdxAggSig)
						}

						return protocol.pendingBlock, nil

					case msgCommit := <-protocol.cBFTMsg:
						if msgCommit.MessageType() == wire.CmdBFTCommit {
							fmt.Println("BFT: Commit msg received")
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
							if _, ok := phaseData.Sigs[R]; !ok {
								phaseData.Sigs[R] = make(map[string]bftCommittedSig)
							}
							if _, ok := phaseData.Sigs[R][msgCommit.(*wire.MessageBFTCommit).Pubkey]; !ok {
								phaseData.Sigs[R][msgCommit.(*wire.MessageBFTCommit).Pubkey] = newSig
								protocol.forwardMsg(msgCommit)
								if len(phaseData.Sigs[R]) >= (2 * len(protocol.RoundData.Committee) / 3) {
									cmTimeout.Stop()
									fmt.Println("BFT: Collected enough Sig")
									protocol.closeTimeoutCh()
								}
							}

						}
					}
				}
			}
		}
	}
}

func (protocol *BFTProtocol) CreateBlockMsg() (wire.Message, error) {
	var msg wire.Message
	if protocol.RoundData.Layer == common.BEACON_ROLE {
		newBlock, err := protocol.BlockGen.NewBlockBeacon(&protocol.UserKeySet.PaymentAddress, &protocol.UserKeySet.PrivateKey, protocol.RoundData.ProposerOffset, protocol.RoundData.ClosestPoolState)
		if err != nil {
			return nil, err
		}
		jsonBlock, _ := json.Marshal(newBlock)
		msg, err = MakeMsgBFTPropose(jsonBlock, protocol.RoundData.Layer, protocol.RoundData.ShardID, protocol.UserKeySet)
		if err != nil {
			return nil, err
		}
		protocol.pendingBlock = newBlock
		protocol.multiSigScheme.dataToSig = newBlock.Header.Hash()
	} else {
		newBlock, err := protocol.BlockGen.NewBlockShard(&protocol.UserKeySet.PaymentAddress, &protocol.UserKeySet.PrivateKey, protocol.RoundData.ShardID, protocol.RoundData.ProposerOffset, protocol.RoundData.ClosestPoolState)
		if err != nil {
			return nil, err
		}
		jsonBlock, _ := json.Marshal(newBlock)
		msg, err = MakeMsgBFTPropose(jsonBlock, protocol.RoundData.Layer, protocol.RoundData.ShardID, protocol.UserKeySet)
		if err != nil {
			return nil, err
		}
		protocol.pendingBlock = newBlock
		protocol.multiSigScheme.dataToSig = newBlock.Header.Hash()
	}
	return msg, nil
}

func (protocol *BFTProtocol) forwardMsg(msg wire.Message) {
	if protocol.RoundData.Layer == common.BEACON_ROLE {
		go protocol.Server.PushMessageToBeacon(msg)
	} else {
		go protocol.Server.PushMessageToShard(msg, protocol.RoundData.ShardID)
	}
}

func (protocol *BFTProtocol) closeTimeoutCh() {
	select {
	case <-protocol.cTimeout:
		return
	default:
		close(protocol.cTimeout)
	}
}
