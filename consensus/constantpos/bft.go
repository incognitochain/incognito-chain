package constantpos

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ninjadotorg/constant/cashec"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/wire"
)

type BFTProtocol struct {
	sync.Mutex
	Phase      string
	cQuit      chan struct{}
	cTimeout   chan struct{}
	cBFTMsg    chan wire.Message
	BlockGen   *blockchain.BlkTmplGenerator
	Chain      *blockchain.BlockChain
	Server     serverInterface
	UserKeySet *cashec.KeySet
	Committee  []string
	started    bool

	pendingBlock blockchain.BFTBlockInterface
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
	self.Phase = "listen"
	if isProposer {
		self.Phase = "propose"
	}
	go func() {
		for {
			self.cTimeout = make(chan struct{})
			select {
			case <-self.cQuit:
				return
			default:
				switch self.Phase {
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
					self.Phase = "prepare"
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
							// Create random Ri and broadcast
							myRi := []byte{0}
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
							self.pendingBlock = phaseData.Block
							self.Phase = "prepare"
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

							//Todo Sig block with R Here

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

							self.Phase = "commit"
							break
						}
					}
				case "commit":
					time.AfterFunc(CommitTimeout*time.Second, func() {
						close(self.cTimeout)
					})
					type validatorSig struct {
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
									ValidatorsIdx: msgCommit.(*wire.MessageBFTCommit).ValidatorsIdx,
									Sig:           msgCommit.(*wire.MessageBFTCommit).CommitSig,
								}
								R := msgCommit.(*wire.MessageBFTCommit).R
								phaseData.Sigs[R] = append(phaseData.Sigs[R], newSig)
							}
						case <-self.cTimeout:
							//Combine collected Sigs with the same R that has the longest list must has size > 1/2size(committee)

							//Todo combine Sigs

							var phaseData struct {
								ValidatorsIdx []int
								AggregatedSig string
							}
							msg, err := MakeMsgBFTReply(phaseData.AggregatedSig, phaseData.ValidatorsIdx)
							if err != nil {
								Logger.log.Error(err)
								return
							}
							if layer == "beacon" {
								self.Server.PushMessageToBeacon(msg)
							} else {
								self.Server.PushMessageToShard(msg, shardID)
							}

							self.Phase = "reply"
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
