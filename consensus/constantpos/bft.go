package constantpos

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ninjadotorg/constant/common/base58"

	"github.com/ninjadotorg/constant/common"

	privacy "github.com/ninjadotorg/constant/privacy"

	"github.com/ninjadotorg/constant/cashec"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/wire"
)

type BFTProtocol struct {
	sync.Mutex

	cBFTMsg    chan wire.Message
	BlockGen   *blockchain.BlkTmplGenerator
	Chain      *blockchain.BlockChain
	Server     serverInterface
	UserKeySet *cashec.KeySet
	Committee  []string

	phase    string
	cQuit    chan struct{}
	cTimeout chan struct{}
	started  bool

	pendingBlock interface{}
	dataForSig   struct {
		Ri []byte
		r  []byte
	}
	dataForCombine struct {
		mySig         string
		R             string
		ValidatorsIdx []int
	}
}

type blockFinalSig struct {
	Count         int
	ValidatorsIdx []int
}

func (self *BFTProtocol) Start(isProposer bool, layer string, shardID byte, prevAggregatedSig string, prevValidatorsIdx []int) {
	self.Lock()
	defer self.Unlock()
	if self.started {
		// return errors.New("Protocol is already started")
		return
	}
	self.started = true
	self.cQuit = make(chan struct{})
	self.phase = "listen"
	if isProposer {
		self.phase = "propose"
	}
	fmt.Println("Starting PBFT protocol for " + layer)

	// go func() {
	multiSigScheme := new(privacy.MultiSigScheme)
	multiSigScheme.Init()
	multiSigScheme.Keyset.Set(&self.UserKeySet.PrivateKey, &self.UserKeySet.PaymentAddress.Pk)
	for {
		self.cTimeout = make(chan struct{})
		select {
		case <-self.cQuit:
			return
		default:
			switch self.phase {
			case "propose":
				if layer == "beacon" {
					newBlock, err := self.BlockGen.NewBlockBeacon(&self.UserKeySet.PaymentAddress, &self.UserKeySet.PrivateKey)
					if err != nil {
						Logger.log.Error(err)
						return
					}
					fmt.Println("Propose block")
					jsonBlock, _ := json.Marshal(newBlock)
					msg, err := MakeMsgBFTPropose(prevAggregatedSig, prevValidatorsIdx, jsonBlock)
					if err != nil {
						Logger.log.Error(err)
						return
					}
					go self.Server.PushMessageToBeacon(msg)
					self.pendingBlock = newBlock
				} else {
					newBlock, err := self.BlockGen.NewBlockShard(&self.UserKeySet.PaymentAddress, &self.UserKeySet.PrivateKey, shardID)
					if err != nil {
						return
					}
					jsonBlock, _ := json.Marshal(newBlock)
					msg, err := MakeMsgBFTPropose(prevAggregatedSig, prevValidatorsIdx, jsonBlock)
					if err != nil {
						Logger.log.Error(err)
						return
					}
					go self.Server.PushMessageToShard(msg, shardID)
					self.pendingBlock = newBlock
				}

				myRiECCPoint, myrBigInt := multiSigScheme.GenerateRandom()
				myRi := myRiECCPoint.Compress()
				myr := myrBigInt.Bytes()
				for len(myr) < privacy.BigIntSize {
					myr = append([]byte{0}, myr...)
				}

				self.dataForSig.Ri = myRi
				self.dataForSig.r = myr

				msg, err := MakeMsgBFTPrepare(myRi, self.UserKeySet.GetPublicKeyB58())
				if err != nil {
					Logger.log.Error(err)
					return
				}
				time.AfterFunc(2000*time.Millisecond, func() {
					fmt.Println("Sending out prepare msg")
					if layer == "beacon" {
						self.Server.PushMessageToBeacon(msg)
					} else {
						self.Server.PushMessageToShard(msg, shardID)
					}
				})
				self.phase = "prepare"
			case "listen":
				fmt.Println("Listen phase")
				timeout := time.AfterFunc(ListenTimeout*time.Second, func() {
					close(self.cTimeout)
				})
				select {
				case msgPropose := <-self.cBFTMsg:
					var phaseData struct {
						PrevAggregatedSig string
						PrevValidatorsIdx []int
						Block             interface{}
					}
					if msgPropose.MessageType() == wire.CmdBFTPropose {
						fmt.Println("Propose block received")
						phaseData.PrevAggregatedSig = msgPropose.(*wire.MessageBFTPropose).AggregatedSig
						phaseData.PrevValidatorsIdx = msgPropose.(*wire.MessageBFTPropose).ValidatorsIdx
						if layer == "beacon" {
							pendingBlk := blockchain.BeaconBlock{}
							pendingBlk.UnmarshalJSON(msgPropose.(*wire.MessageBFTPropose).Block)
							self.Chain.VerifyPreProcessingBeaconBlock(&pendingBlk)
							phaseData.Block = &pendingBlk
						} else {
							pendingBlk := blockchain.ShardBlock{}
							pendingBlk.UnmarshalJSON(msgPropose.(*wire.MessageBFTPropose).Block)
							self.Chain.VerifyPreProcessingShardBlock(&pendingBlk)
							phaseData.Block = &pendingBlk
						}
						// Todo create random Ri and broadcast

						myRiECCPoint, myrBigInt := multiSigScheme.GenerateRandom()
						myRi := myRiECCPoint.Compress()
						myr := myrBigInt.Bytes()
						for len(myr) < privacy.BigIntSize {
							myr = append([]byte{0}, myr...)
						}
						msg, err := MakeMsgBFTPrepare(myRi, self.UserKeySet.GetPublicKeyB58())
						if err != nil {
							Logger.log.Error(err)
							return
						}
						time.AfterFunc(1500*time.Millisecond, func() {
							fmt.Println("Sending out prepare msg")
							if layer == "beacon" {
								self.Server.PushMessageToBeacon(msg)
							} else {
								self.Server.PushMessageToShard(msg, shardID)
							}
						})

						self.dataForSig.Ri = myRi
						self.dataForSig.r = myr
						self.pendingBlock = phaseData.Block

						self.phase = "prepare"
						timeout.Stop()
						continue
					}
				case <-self.cTimeout:
				}
			case "prepare":
				fmt.Println("Prepare phase")
				time.AfterFunc(PrepareTimeout*time.Second, func() {
					fmt.Println("Prepare phase timeout")
					close(self.cTimeout)
				})
				var phaseData struct {
					RiList map[string][]byte

					R             string
					ValidatorsIdx []int
					CommitBlkSig  string
				}
				phaseData.RiList = make(map[string][]byte)
				phaseData.RiList[self.UserKeySet.GetPublicKeyB58()] = self.dataForSig.Ri
			preparephase:
				for {
					select {
					case msgPrepare := <-self.cBFTMsg:
						if msgPrepare.MessageType() == wire.CmdBFTPrepare {
							fmt.Println("Prepare msg received")
							if common.IndexOfStr(msgPrepare.(*wire.MessageBFTPrepare).Pubkey, self.Committee) >= 0 {
								phaseData.RiList[msgPrepare.(*wire.MessageBFTPrepare).Pubkey] = msgPrepare.(*wire.MessageBFTPrepare).Ri
							}
						}
					case <-self.cTimeout:
						//Use collected Ri to calc R & get ValidatorsIdx if len(Ri) > 1/2size(committee)
						// then sig block with this R
						// phaseData.R = base58.Base58Check{}.Encode(Rbytes, byte(0x00))
						// base58.Base58Check{}.
						numbOfSigners := len(phaseData.RiList)
						listPubkeyOfSigners := make([]*privacy.PublicKey, numbOfSigners)
						listROfSigners := make([]*privacy.EllipticPoint, numbOfSigners)
						RCombined := new(privacy.EllipticPoint)
						RCombined.Set(big.NewInt(0), big.NewInt(0))
						counter := 0
						// var byteVersion byte
						// var err error

						for szPubKey, bytesR := range phaseData.RiList {
							pubKeyTemp, byteVersion, err := base58.Base58Check{}.Decode(szPubKey)
							listPubkeyOfSigners[counter] = new(privacy.PublicKey)
							*listPubkeyOfSigners[counter] = pubKeyTemp
							if (err != nil) || (byteVersion != byte(0x00)) {
								//Todo
								fmt.Println("BLah err", err)
								return
							}
							listROfSigners[counter] = new(privacy.EllipticPoint)
							err = listROfSigners[counter].Decompress(bytesR)
							if err != nil {
								//Todo
								fmt.Println("BLah err", err)
								return
							}
							RCombined = RCombined.Add(listROfSigners[counter])
							// phaseData.ValidatorsIdx[counter] = sort.SearchStrings(self.Committee, szPubKey)
							phaseData.ValidatorsIdx = append(phaseData.ValidatorsIdx, common.IndexOfStr(szPubKey, self.Committee))
							counter++
						}

						//Todo Sig block with R Here
						var blockData common.Hash
						if layer == "beacon" {
							blockData = self.pendingBlock.(*blockchain.BeaconBlock).Header.Hash()
						} else {
							blockData = self.pendingBlock.(*blockchain.ShardBlock).Header.Hash()
						}

						multiSigScheme.Signature = multiSigScheme.Keyset.SignMultiSig(blockData.GetBytes(), listPubkeyOfSigners, listROfSigners, new(big.Int).SetBytes(self.dataForSig.r))
						phaseData.CommitBlkSig = base58.Base58Check{}.Encode(multiSigScheme.Signature.Bytes(), byte(0x00))
						phaseData.R = base58.Base58Check{}.Encode(RCombined.Compress(), byte(0x00))

						msg, err := MakeMsgBFTCommit(phaseData.CommitBlkSig, phaseData.R, phaseData.ValidatorsIdx, self.UserKeySet.GetPublicKeyB58())
						if err != nil {
							fmt.Println("BLah err", err)
							return
						}
						if layer == "beacon" {
							self.Server.PushMessageToBeacon(msg)
						} else {
							self.Server.PushMessageToShard(msg, shardID)
						}

						self.dataForCombine.R = phaseData.R
						self.dataForCombine.ValidatorsIdx = phaseData.ValidatorsIdx
						self.dataForCombine.mySig = phaseData.CommitBlkSig

						self.phase = "commit"
						break preparephase
					}
				}
			case "commit":
				fmt.Println("Commit phase")
				time.AfterFunc(CommitTimeout*time.Second, func() {
					fmt.Println("Commit phase timeout")
					close(self.cTimeout)
				})
				type validatorSig struct {
					Pubkey        string
					ValidatorsIdx []int
					Sig           string
				}
				var phaseData struct {
					Sigs map[string][]validatorSig
				}

				phaseData.Sigs = make(map[string][]validatorSig)
				phaseData.Sigs[self.dataForCombine.R] = append(phaseData.Sigs[self.dataForCombine.R], validatorSig{
					Pubkey:        self.UserKeySet.GetPublicKeyB58(),
					Sig:           self.dataForCombine.mySig,
					ValidatorsIdx: self.dataForCombine.ValidatorsIdx,
				})
			commitphase:
				for {
					select {
					case msgCommit := <-self.cBFTMsg:
						if msgCommit.MessageType() == wire.CmdBFTCommit {
							fmt.Println("Commit msg received")
							newSig := validatorSig{
								Pubkey:        msgCommit.(*wire.MessageBFTCommit).Pubkey,
								ValidatorsIdx: msgCommit.(*wire.MessageBFTCommit).ValidatorsIdx,
								Sig:           msgCommit.(*wire.MessageBFTCommit).CommitSig,
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

							//RCombined decode from base58
							RCombined := new(privacy.EllipticPoint)
							RCombined.Set(big.NewInt(0), big.NewInt(0))
							Rbytesarr, byteVersion, err := base58.Base58Check{}.Decode(R)
							if (err != nil) || (byteVersion != byte(0x00)) {
								fmt.Println("BLah err", err)
								return
							}
							err = RCombined.Decompress(Rbytesarr)
							if err != nil {
								fmt.Println("BLah err", err)
								return
							}
							listPubkeyOfSigners := make([]*privacy.PublicKey, len(newSig.ValidatorsIdx))
							var pubKeyTemp []byte
							for i := 0; i < len(newSig.ValidatorsIdx); i++ {
								listPubkeyOfSigners[i] = new(privacy.PublicKey)
								pubKeyTemp, byteVersion, err = base58.Base58Check{}.Decode(self.Committee[newSig.ValidatorsIdx[i]])
								if (err != nil) || (byteVersion != byte(0x00)) {
									fmt.Println("BLah err", err)
									return
								}
								*listPubkeyOfSigners[i] = pubKeyTemp
							}
							selfPubkey := new(privacy.PublicKey)
							pubKeyTemp, byteVersion, err = base58.Base58Check{}.Decode(newSig.Pubkey)
							if (err != nil) || (byteVersion != byte(0x00)) {
								//Todo
								fmt.Println("BLah err", err)
								return
							}
							*selfPubkey = pubKeyTemp
							var valSigbytesarr []byte
							valSigbytesarr, byteVersion, err = base58.Base58Check{}.Decode(newSig.Sig)
							valSig := new(privacy.SchnMultiSig)
							valSig.SetBytes(valSigbytesarr)
							var blockData common.Hash
							if layer == "beacon" {
								blockData = self.pendingBlock.(*blockchain.BeaconBlock).Header.Hash()
							} else {
								blockData = self.pendingBlock.(*blockchain.ShardBlock).Header.Hash()
							}

							resValidateEachSigOfSigners := valSig.VerifyMultiSig(blockData.GetBytes(), listPubkeyOfSigners, selfPubkey, RCombined)
							if !resValidateEachSigOfSigners {
								fmt.Println("BLah err22222")
								return
							}
							phaseData.Sigs[R] = append(phaseData.Sigs[R], newSig)
						}
					case <-self.cTimeout:
						//Combine collected Sigs with the same R that has the longest list must has size > 1/2size(committee)
						var szRCombined string
						szRCombined = "1"
						for szR := range phaseData.Sigs {
							if len(phaseData.Sigs[szR]) > (len(self.Committee) >> 1) {
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
							return
						}
						listSigOfSigners := make([]*privacy.SchnMultiSig, len(phaseData.Sigs[szRCombined]))
						for i, valSig := range phaseData.Sigs[szRCombined] {
							listSigOfSigners[i] = new(privacy.SchnMultiSig)
							bytesSig, byteVersion, err := base58.Base58Check{}.Decode(valSig.Sig)
							if (err != nil) || (byteVersion != byte(0x00)) {
								fmt.Println("BLah err", err)
								return
							}
							listSigOfSigners[i].SetBytes(bytesSig)
						}

						AggregatedSig := multiSigScheme.CombineMultiSig(listSigOfSigners)

						// listPubkeyOfSigners := make([]*privacy.PublicKey, len(phaseData.Sigs[szRCombined][0].ValidatorsIdx))
						// for i := 0; i < len(phaseData.Sigs[szRCombined][0].ValidatorsIdx); i++ {
						// 	listPubkeyOfSigners[i] = new(privacy.PublicKey)
						// 	pubKeyTemp, byteVersion, err := base58.Base58Check{}.Decode(self.Committee[phaseData.Sigs[szRCombined][0].ValidatorsIdx[i]])
						// 	if (err != nil) || (byteVersion != byte(0x00)) {

						// 		return
						// 	}
						// 	*listPubkeyOfSigners[i] = pubKeyTemp
						// }

						// var blockData common.Hash
						// if layer == "beacon" {
						// 	blockData = self.pendingBlock.(*blockchain.BeaconBlock).Header.Hash()
						// } else {
						// 	blockData = self.pendingBlock.(*blockchain.ShardBlock).Header.Hash()
						// }
						// fmt.Println(AggregatedSig.VerifyMultiSig(blockData.GetBytes(), listPubkeyOfSigners, nil, nil))

						var replyData struct {
							ValidatorsIdx []int
							AggregatedSig string
						}
						replyData.ValidatorsIdx = make([]int, len(phaseData.Sigs[szRCombined][0].ValidatorsIdx))
						copy(replyData.ValidatorsIdx, phaseData.Sigs[szRCombined][0].ValidatorsIdx)
						replyData.AggregatedSig = base58.Base58Check{}.Encode(AggregatedSig.Bytes(), byte(0x00))
						msg, err := MakeMsgBFTReply(replyData.AggregatedSig, replyData.ValidatorsIdx)
						if err != nil {
							fmt.Println("BLah err", err)
							return
						}
						fmt.Println("Sending out reply msg")
						if layer == "beacon" {
							self.Server.PushMessageToBeacon(msg)
						} else {
							self.Server.PushMessageToShard(msg, shardID)
						}

						self.phase = "reply"
						break commitphase
					}
				}
			case "reply":
				fmt.Println("Reply phase")
				time.AfterFunc(ReplyTimeout*time.Second, func() {
					close(self.cTimeout)
				})
				for {
					select {
					case msgReply := <-self.cBFTMsg:
						fmt.Println(msgReply)
					case <-self.cTimeout:

					}
				}
			}
		}

	}
	// }()
	return
}

func (self *BFTProtocol) Stop() error {
	self.Lock()
	defer self.Unlock()
	if !self.started {
		return errors.New("Protocol is already stopped")
	}
	self.started = false
	close(self.cTimeout)
	close(self.cQuit)
	return nil
}
