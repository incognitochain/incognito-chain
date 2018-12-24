package constantpos

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ninjadotorg/constant/common/base58"

	"github.com/ninjadotorg/constant/common"

	privacy "github.com/ninjadotorg/constant/privacy-protocol"

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

	pendingBlock blockchain.BFTBlockInterface
	dataForSig   struct {
		Ri []byte
		r  []byte
	}
}

type blockFinalSig struct {
	Count         int
	ValidatorsIdx []int
}

func (self *BFTProtocol) Start(isProposer bool, layer string, shardID byte, prevAggregatedSig string, prevValidatorsIdx []int) error {
	self.Lock()
	defer self.Unlock()
	if self.started {
		return errors.New("Protocol is already started")
	}
	self.started = true
	self.cQuit = make(chan struct{})
	self.phase = "listen"
	if isProposer {
		self.phase = "propose"
	}

	go func() {
		for {
			self.cTimeout = make(chan struct{})
			multiSigScheme := new(privacy.MultiSigScheme)
			multiSigScheme.Init()
			multiSigScheme.Keyset.Set(&self.UserKeySet.PrivateKey, &self.UserKeySet.PaymentAddress.Pk)
			select {
			case <-self.cQuit:
				return
			default:
				switch self.phase {
				case "propose":
					time.AfterFunc(ProposeTimeout*time.Second, func() {
						close(self.cTimeout)
					})
					if layer == "beacon" {
						newBlock, err := self.BlockGen.NewBlockBeacon(&self.UserKeySet.PaymentAddress, &self.UserKeySet.PrivateKey)
						if err != nil {
							Logger.log.Error(err)
							return
						}
						msg, err := MakeMsgBFTPropose(prevAggregatedSig, prevValidatorsIdx, newBlock)
						if err != nil {
							Logger.log.Error(err)
							return
						}
						self.Server.PushMessageToBeacon(msg)
						self.pendingBlock = newBlock
					} else {
						newBlock, err := self.BlockGen.NewBlockShard(&self.UserKeySet.PaymentAddress, &self.UserKeySet.PrivateKey, shardID)
						if err != nil {
							return
						}
						msg, err := MakeMsgBFTPropose(prevAggregatedSig, prevValidatorsIdx, newBlock)
						if err != nil {
							Logger.log.Error(err)
							return
						}
						self.Server.PushMessageToShard(msg, shardID)
						self.pendingBlock = newBlock
					}
					self.phase = "prepare"
				case "listen":
					time.AfterFunc(ListenTimeout*time.Second, func() {
						close(self.cTimeout)
					})
					select {
					case msgPropose := <-self.cBFTMsg:
						var phaseData struct {
							PrevAggregatedSig string
							PrevValidatorsIdx []int
							Block             blockchain.BFTBlockInterface
						}
						if msgPropose.MessageType() == wire.CmdBFTPropose {
							phaseData.PrevAggregatedSig = msgPropose.(*wire.MessageBFTPropose).AggregatedSig
							phaseData.PrevValidatorsIdx = msgPropose.(*wire.MessageBFTPropose).ValidatorsIdx
							phaseData.Block = msgPropose.(*wire.MessageBFTPropose).Block
							if layer == "beacon" {
								self.Chain.VerifyPreProcessingBlockBeacon(phaseData.Block.(*blockchain.BeaconBlock))
							} else {
								self.Chain.VerifyPreProcessingBlockShard(phaseData.Block.(*blockchain.ShardBlock))
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
							if layer == "beacon" {
								self.Server.PushMessageToBeacon(msg)
							} else {
								self.Server.PushMessageToShard(msg, shardID)
							}
							self.dataForSig.Ri = myRi
							self.dataForSig.r = myr
							self.pendingBlock = phaseData.Block
							self.phase = "prepare"
						}
					case <-self.cTimeout:
					}
				case "prepare":
					time.AfterFunc(PrepareTimeout*time.Second, func() {
						close(self.cTimeout)
					})
					var phaseData struct {
						RiList map[string][]byte

						R             string
						ValidatorsIdx []int
						CommitBlkSig  string
					}
					phaseData.RiList = make(map[string][]byte)
					for {
						select {
						case msgPrepare := <-self.cBFTMsg:
							if msgPrepare.MessageType() == wire.CmdBFTPrepare {
								phaseData.RiList[msgPrepare.(*wire.MessageBFTPrepare).Pubkey] = msgPrepare.(*wire.MessageBFTPrepare).Ri
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
							// RCombined.Set(big.NewInt(0), big.NewInt(0))
							counter := 0
							// var byteVersion byte
							// var err error

							for szPubKey, bytesR := range phaseData.RiList {
								pubKeyTemp, byteVersion, err := base58.Base58Check{}.Decode(szPubKey)
								listPubkeyOfSigners[counter] = new(privacy.PublicKey)
								*listPubkeyOfSigners[counter] = pubKeyTemp
								if (err != nil) || (byteVersion != byte(0x00)) {
									//Todo
									return
								}
								listROfSigners[counter] = new(privacy.EllipticPoint)
								err = listROfSigners[counter].Decompress(bytesR)
								if err != nil {
									//Todo
									return
								}
								RCombined = RCombined.Add(listROfSigners[counter])
								// phaseData.ValidatorsIdx[counter] = sort.SearchStrings(self.Committee, szPubKey)
								phaseData.ValidatorsIdx[counter] = common.IndexOfStr(szPubKey, self.Committee)

								counter++
							}

							//Todo Sig block with R Here
							var blockData common.Hash
							if layer == "beacon" {
								blockData = self.pendingBlock.(*blockchain.BeaconBlock).Header.Hash()
							} else {
								blockData = self.pendingBlock.(*blockchain.ShardBlock).Header.Hash()
							}
							blockData.GetBytes()
							multiSigScheme.Signature = multiSigScheme.Keyset.SignMultiSig(blockData.GetBytes(), listPubkeyOfSigners, listROfSigners, new(big.Int).SetBytes(self.dataForSig.r))
							phaseData.CommitBlkSig = base58.Base58Check{}.Encode(multiSigScheme.Signature.Bytes(), byte(0x00))
							phaseData.R = base58.Base58Check{}.Encode(RCombined.Compress(), byte(0x00))

							msg, err := MakeMsgBFTCommit(phaseData.CommitBlkSig, phaseData.R, phaseData.ValidatorsIdx, self.UserKeySet.GetPublicKeyB58())
							if err != nil {
								Logger.log.Error(err)
								return
							}
							if layer == "beacon" {
								self.Server.PushMessageToBeacon(msg)
							} else {
								self.Server.PushMessageToShard(msg, shardID)
							}

							self.phase = "commit"
							break
						}
					}
				case "commit":
					time.AfterFunc(CommitTimeout*time.Second, func() {
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
					for {
						select {
						case msgCommit := <-self.cBFTMsg:
							if msgCommit.MessageType() == wire.CmdBFTCommit {
								newSig := validatorSig{
									Pubkey:        msgCommit.(*wire.MessageBFTCommit).Pubkey,
									ValidatorsIdx: msgCommit.(*wire.MessageBFTCommit).ValidatorsIdx,
									Sig:           msgCommit.(*wire.MessageBFTCommit).CommitSig,
								}
								R := msgCommit.(*wire.MessageBFTCommit).R
								RCombined := new(privacy.EllipticPoint)
								// RCombined.Set(big.NewInt(0), big.NewInt(0))
								Rbytesarr, byteVersion, err := base58.Base58Check{}.Decode(R)
								if (err != nil) || (byteVersion != byte(0x00)) {
									//Todo
									return
								}
								err = RCombined.Decompress(Rbytesarr)
								if err != nil {
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
							if len(szRCombined) != 1 {
								return
							}
							//Todo combine Sigs
							listSigOfSigners := make([]*privacy.SchnMultiSig, len(phaseData.Sigs[szRCombined]))
							for i, valSig := range phaseData.Sigs[szRCombined] {
								listSigOfSigners[i] = new(privacy.SchnMultiSig)
								bytesSig, byteVersion, err := base58.Base58Check{}.Decode(valSig.Sig)
								if (err != nil) || (byteVersion != byte(0x00)) {
									return
								}
								listSigOfSigners[i].SetBytes(bytesSig)
							}
							AggregatedSig := multiSigScheme.CombineMultiSig(listSigOfSigners)

							var replyData struct {
								ValidatorsIdx []int
								AggregatedSig string
							}
							replyData.ValidatorsIdx = make([]int, len(phaseData.Sigs[szRCombined][0].ValidatorsIdx))
							copy(replyData.ValidatorsIdx, phaseData.Sigs[szRCombined][0].ValidatorsIdx)
							replyData.AggregatedSig = base58.Base58Check{}.Encode(AggregatedSig.Bytes(), byte(0x00))
							msg, err := MakeMsgBFTReply(replyData.AggregatedSig, replyData.ValidatorsIdx)
							if err != nil {
								Logger.log.Error(err)
								return
							}
							if layer == "beacon" {
								self.Server.PushMessageToBeacon(msg)
							} else {
								self.Server.PushMessageToShard(msg, shardID)
							}

							self.phase = "reply"
							break
						}
					}
				case "reply":
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
	}()
	return nil
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

// Moving them soon
// type sortString []string

// func lessString(a, b string) bool {
// 	if len(a) < len(b) {
// 		return true
// 	}
// 	if len(a) > len(b) {
// 		return false
// 	}

// 	for i := 0; i < len(a); i++ {
// 		if a[i] > b[i] {
// 			return false
// 		}
// 		if a[i] < b[i] {
// 			return true
// 		}
// 	}
// 	return false
// }

// func swap(str1, str2 *string) {
// 	*str1, *str2 = *str2, *str1
// }

// func (s sortString) Less(i, j int) bool {
// 	return lessString(s[i], s[j])
// }

// func (s sortString) Swap(i, j int) {
// 	swap(&s[i], &s[j])
// }

// func (s sortString) Len() int {
// 	return len(s)
// }

// func SortString(s []string) []string {
// 	// r := [](s)
// 	sort.Sort(sortString(s))
// 	return s
// }
