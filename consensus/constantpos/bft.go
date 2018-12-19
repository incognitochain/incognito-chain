package constantpos

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/wire"
)

type BlockShardGenFn func() *blockchain.ShardBlock
type BlockBeaconGenFn func() *blockchain.BeaconBlock

type BFTProtocol struct {
	sync.Mutex
	Phase    string
	cQuit    chan struct{}
	cTimeout chan struct{}
	cBFTMsg  chan wire.Message
	BlockGen struct {
		Layer  BlockShardGenFn
		Beacon BlockBeaconGenFn
	}
	Server serverInterface

	started bool
}

type blockFinalSig struct {
	Count         int
	ValidatorsIdx []int
}

func (self *BFTProtocol) Start() error {
	self.Lock()
	defer self.Unlock()
	if self.started {
		return errors.New("Protocol is already started")
	}
	self.started = true
	self.cQuit = make(chan struct{})
	go func() {
		for {
			self.cTimeout = make(chan struct{})
			select {
			case <-self.cQuit:
				return
			default:
				switch self.Phase {
				case "propose":
					newBlock := self.BlockGen.Layer()
					fmt.Println(newBlock)
				case "listen":
					time.AfterFunc(ListenTimeout*time.Second, func() {
						close(self.cTimeout)
					})
					select {
					case msgPropose := <-self.cBFTMsg:
						if msgPropose.MessageType() == wire.CmdBFTPropose {
							fmt.Println(msgPropose)
						}
						self.Phase = "prepare"
					case <-self.cTimeout:
					}
				case "prepare":
					time.AfterFunc(PrepareTimeout*time.Second, func() {
						close(self.cTimeout)
					})

					for {
						select {
						case msgPrepare := <-self.cBFTMsg:
							fmt.Println(msgPrepare)
						case <-self.cTimeout:
							break
						}
					}
					self.Phase = "commit"
				case "commit":
					time.AfterFunc(CommitTimeout*time.Second, func() {
						close(self.cTimeout)
					})
					for {
						select {
						case msgCommit := <-self.cBFTMsg:
							fmt.Println(msgCommit)
						case <-self.cTimeout:
							break
						}
					}

					self.Phase = "reply"
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

// func (self *BFTProtocol)
