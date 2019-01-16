package constantpos

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/ninjadotorg/constant/common"

	"github.com/ninjadotorg/constant/cashec"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/wire"
)

type BFTProtocol struct {
	cBFTMsg   chan wire.Message
	EngineCfg *EngineConfig

	BlockGen   *blockchain.BlkTmplGenerator
	Chain      *blockchain.BlockChain
	Server     serverInterface
	UserKeySet *cashec.KeySet

	cQuit    chan struct{}
	cTimeout chan struct{}

	phase string

	pendingBlock interface{}

	RoleData struct {
		IsProposer bool
		Layer      string
		ShardID    byte
		Committee  []string
	}

	multiSigScheme *multiSigScheme
}

func (self *BFTProtocol) Start(isProposer bool, layer string, shardID byte) (interface{}, error) {

	self.phase = "listen"
	if isProposer {
		self.phase = "propose"
	}
	fmt.Println("Starting PBFT protocol for " + layer)
	self.multiSigScheme = new(multiSigScheme)
	self.multiSigScheme.Init(self.UserKeySet, self.RoleData.Committee)
	err := self.multiSigScheme.Prepare()
	if err != nil {
		return nil, err
	}
	for {
		fmt.Println("New Phase")
		self.cTimeout = make(chan struct{})
		select {
		case <-self.cQuit:
			return nil, errors.New("Consensus quit")
		default:
			switch self.phase {
			case "propose":
				timeout := time.AfterFunc(ListenTimeout*time.Second, func() {
					fmt.Println("Propose phase timeout")
					close(self.cTimeout)
				})
				var (
					msg           wire.Message
					shardID       byte
					readyMsgCount int
				)
				if layer == "beacon" {
					newBlock, err := self.BlockGen.NewBlockBeacon(&self.UserKeySet.PaymentAddress, &self.UserKeySet.PrivateKey)
					if err != nil {
						return nil, err
					}
					fmt.Println("Propose block")
					jsonBlock, _ := json.Marshal(newBlock)
					msg, err = MakeMsgBFTPropose(jsonBlock)
					if err != nil {
						return nil, err
					}
					go self.Server.PushMessageToBeacon(msg)
					self.pendingBlock = newBlock
					self.multiSigScheme.dataToSig = newBlock.Header.Hash()
				} else {
					newBlock, err := self.BlockGen.NewBlockShard(&self.UserKeySet.PaymentAddress, &self.UserKeySet.PrivateKey, shardID)
					if err != nil {
						return nil, err
					}
					jsonBlock, _ := json.Marshal(newBlock)
					msg, err = MakeMsgBFTPropose(jsonBlock)
					if err != nil {
						return nil, err
					}
					go self.Server.PushMessageToShard(msg, shardID)
					self.pendingBlock = newBlock
					fmt.Println("\n", newBlock.Header)
					self.multiSigScheme.dataToSig = newBlock.Header.Hash()
				}

				for {
					select {
					case msgReady := <-self.cBFTMsg:
						if msgReady.MessageType() == wire.CmdBFTReady {
							readyMsgCount++
							if readyMsgCount > (2 * len(self.RoleData.Committee) / 3) {
								timeout.Stop()
								fmt.Println("Collected enough ready")
								select {
								case <-self.cTimeout:
									continue
								default:
									close(self.cTimeout)
								}
							}
						}
					case <-self.cTimeout:

					}
				}
				self.phase = "prepare"
			case "listen":
				fmt.Println("Listen phase")
				timeout := time.AfterFunc(ListenTimeout*time.Second, func() {
					fmt.Println("Listen phase timeout")
					close(self.cTimeout)
				})
			listenphase:
				for {
					select {
					case msgPropose := <-self.cBFTMsg:
						if msgPropose.MessageType() == wire.CmdBFTPropose {
							fmt.Println("Propose block received")
							if layer == "beacon" {
								pendingBlk := blockchain.BeaconBlock{}
								pendingBlk.UnmarshalJSON(msgPropose.(*wire.MessageBFTPropose).Block)
								blkHash := pendingBlk.Header.Hash()
								err := cashec.ValidateDataB58(pendingBlk.Header.Producer, pendingBlk.ProducerSig, blkHash.GetBytes())
								if err != nil {
									Logger.log.Error(err)
									continue
								}
								self.Chain.VerifyPreSignBeaconBlock(&pendingBlk)
								self.pendingBlock = &pendingBlk
								self.multiSigScheme.dataToSig = pendingBlk.Header.Hash()
							} else {
								pendingBlk := blockchain.ShardBlock{}
								pendingBlk.UnmarshalJSON(msgPropose.(*wire.MessageBFTPropose).Block)
								blkHash := pendingBlk.Header.Hash()
								err := cashec.ValidateDataB58(pendingBlk.Header.Producer, pendingBlk.ProducerSig, blkHash.GetBytes())
								if err != nil {
									Logger.log.Error(err)
									continue
								}
								self.Chain.VerifyPreSignShardBlock(&pendingBlk, self.RoleData.ShardID)
								self.pendingBlock = &pendingBlk
								self.multiSigScheme.dataToSig = pendingBlk.Header.Hash()
							}

							self.phase = "prepare"
							timeout.Stop()
							break listenphase
						}
					case <-self.cTimeout:
						return nil, errors.New("Listen phase timeout")
					}
				}
			case "prepare":
				fmt.Println("Prepare phase")
				time.AfterFunc(PrepareTimeout*time.Second, func() {
					fmt.Println("Prepare phase timeout")
					close(self.cTimeout)
				})
				time.AfterFunc(1500*time.Millisecond, func() {
					fmt.Println("Sending out prepare msg")
					msg, err := MakeMsgBFTPrepare(self.multiSigScheme.personal.Ri, self.UserKeySet.GetPublicKeyB58(), self.multiSigScheme.dataToSig.String())
					if err != nil {
						Logger.log.Error(err)
						return
					}
					if layer == "beacon" {
						self.Server.PushMessageToBeacon(msg)
					} else {
						self.Server.PushMessageToShard(msg, shardID)
					}
				})

				var collectedRiList map[string][]byte //map of members and their Ri
				collectedRiList = make(map[string][]byte)
				collectedRiList[self.UserKeySet.GetPublicKeyB58()] = self.multiSigScheme.personal.Ri
			preparephase:
				for {
					select {
					case msgPrepare := <-self.cBFTMsg:
						if msgPrepare.MessageType() == wire.CmdBFTPrepare {
							fmt.Println("Prepare msg received")
							if common.IndexOfStr(msgPrepare.(*wire.MessageBFTPrepare).Pubkey, self.RoleData.Committee) >= 0 && (self.multiSigScheme.dataToSig.String() == msgPrepare.(*wire.MessageBFTPrepare).BlkHash) {
								collectedRiList[msgPrepare.(*wire.MessageBFTPrepare).Pubkey] = msgPrepare.(*wire.MessageBFTPrepare).Ri
							}
						}
					case <-self.cTimeout:
						//Use collected Ri to calc R & get ValidatorsIdx if len(Ri) > 1/2size(committee)
						// then sig block with this R
						if len(collectedRiList) < (len(self.RoleData.Committee) >> 1) {
							return nil, errors.New("Didn't receive enough Ri to continue")
						}
						err := self.multiSigScheme.SignData(collectedRiList)
						if err != nil {
							return nil, err
						}

						self.phase = "commit"
						break preparephase
					}
				}
			case "commit":
				fmt.Println("Commit phase")
				cmTimeout := time.AfterFunc(CommitTimeout*time.Second, func() {
					fmt.Println("Commit phase timeout")
					close(self.cTimeout)
				})

				time.AfterFunc(1500*time.Millisecond, func() {
					msg, err := MakeMsgBFTCommit(self.multiSigScheme.combine.CommitSig, self.multiSigScheme.combine.R, self.multiSigScheme.combine.ValidatorsIdxR, self.UserKeySet.GetPublicKeyB58())
					if err != nil {
						Logger.log.Error(err)
						return
					}
					fmt.Println("Sending out commit msg")
					if layer == "beacon" {
						self.Server.PushMessageToBeacon(msg)
					} else {
						self.Server.PushMessageToShard(msg, shardID)
					}
				})
				var phaseData struct {
					Sigs map[string][]bftCommittedSig
				}

				phaseData.Sigs = make(map[string][]bftCommittedSig)
				phaseData.Sigs[self.multiSigScheme.combine.R] = append(phaseData.Sigs[self.multiSigScheme.combine.R], bftCommittedSig{
					Pubkey:         self.UserKeySet.GetPublicKeyB58(),
					Sig:            self.multiSigScheme.combine.CommitSig,
					ValidatorsIdxR: self.multiSigScheme.combine.ValidatorsIdxR,
				})
				// commitphase:
				for {
					select {
					case <-self.cTimeout:
						//Combine collected Sigs with the same R that has the longest list must has size > 1/2size(committee)
						var szRCombined string
						szRCombined = "1"
						for szR := range phaseData.Sigs {
							if len(phaseData.Sigs[szR]) > (len(self.RoleData.Committee) >> 1) {
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

						AggregatedSig, err := self.multiSigScheme.CombineSigs(szRCombined, phaseData.Sigs[szRCombined])
						if err != nil {
							return nil, err
						}
						ValidatorsIdxAggSig := make([]int, len(self.multiSigScheme.combine.ValidatorsIdxAggSig))
						ValidatorsIdxR := make([]int, len(self.multiSigScheme.combine.ValidatorsIdxR))
						copy(ValidatorsIdxAggSig, self.multiSigScheme.combine.ValidatorsIdxAggSig)
						copy(ValidatorsIdxR, self.multiSigScheme.combine.ValidatorsIdxR)

						fmt.Println("\n \n Block consensus reach", ValidatorsIdxR, ValidatorsIdxAggSig, AggregatedSig, "\n")

						if layer == "beacon" {
							self.pendingBlock.(*blockchain.BeaconBlock).R = self.multiSigScheme.combine.R
							self.pendingBlock.(*blockchain.BeaconBlock).AggregatedSig = AggregatedSig
							self.pendingBlock.(*blockchain.BeaconBlock).ValidatorsIdx = make([][]int, 2)
							self.pendingBlock.(*blockchain.BeaconBlock).ValidatorsIdx[0] = make([]int, len(ValidatorsIdxR))
							self.pendingBlock.(*blockchain.BeaconBlock).ValidatorsIdx[1] = make([]int, len(ValidatorsIdxAggSig))
							copy(self.pendingBlock.(*blockchain.BeaconBlock).ValidatorsIdx[0], ValidatorsIdxR)
							copy(self.pendingBlock.(*blockchain.BeaconBlock).ValidatorsIdx[1], ValidatorsIdxAggSig)
						} else {
							self.pendingBlock.(*blockchain.ShardBlock).R = self.multiSigScheme.combine.R
							self.pendingBlock.(*blockchain.ShardBlock).AggregatedSig = AggregatedSig
							self.pendingBlock.(*blockchain.ShardBlock).ValidatorsIdx = make([][]int, 2)
							self.pendingBlock.(*blockchain.ShardBlock).ValidatorsIdx[0] = make([]int, len(ValidatorsIdxR))
							self.pendingBlock.(*blockchain.ShardBlock).ValidatorsIdx[1] = make([]int, len(ValidatorsIdxAggSig))
							copy(self.pendingBlock.(*blockchain.ShardBlock).ValidatorsIdx[0], ValidatorsIdxR)
							copy(self.pendingBlock.(*blockchain.ShardBlock).ValidatorsIdx[1], ValidatorsIdxAggSig)
						}

						return self.pendingBlock, nil
					case msgCommit := <-self.cBFTMsg:
						if msgCommit.MessageType() == wire.CmdBFTCommit {
							fmt.Println("Commit msg received")
							newSig := bftCommittedSig{
								Pubkey:         msgCommit.(*wire.MessageBFTCommit).Pubkey,
								ValidatorsIdxR: msgCommit.(*wire.MessageBFTCommit).ValidatorsIdx,
								Sig:            msgCommit.(*wire.MessageBFTCommit).CommitSig,
							}
							R := msgCommit.(*wire.MessageBFTCommit).R

							//Check that Validators Index array in newSig and Validators Index array in each of sig have the same R are equality
							// for _, valSig := range phaseData.Sigs[R] {
							// 	for i, value := range valSig.ValidatorsIdx {
							// 		if value != newSig.ValidatorsIdx[i] {
							// 			return
							// 		}
							// 	}
							// }

							err := self.multiSigScheme.VerifyCommitSig(newSig.Pubkey, newSig.Sig, R, newSig.ValidatorsIdxR)
							if err != nil {
								return nil, err
							}
							phaseData.Sigs[R] = append(phaseData.Sigs[R], newSig)
							if len(phaseData.Sigs[R]) > (2 * len(self.RoleData.Committee) / 3) {
								cmTimeout.Stop()
								fmt.Println("Collected enough R")
								select {
								case <-self.cTimeout:
									continue
								default:
									close(self.cTimeout)
								}
							}
						}
					}
				}
			case "reply":

			}
		}
	}
}
