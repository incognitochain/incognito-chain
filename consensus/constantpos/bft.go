package constantpos

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ninjadotorg/constant/blockchain"
)

type BlockGenFn func() *blockchain.BlockV2

type BFTProtocol struct {
	sync.Mutex
	Phase    string
	cQuit    chan struct{}
	cTimeout chan struct{}

	BlockGenFn BlockGenFn
	Server     serverInterface
	started    bool
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
					newBlock := self.BlockGenFn()
					fmt.Println(newBlock)
				case "listen":
					time.AfterFunc(ListenTimeout*time.Second, func() {
						close(self.cTimeout)
					})

					<-self.cTimeout
				case "prepare":
					time.AfterFunc(PrepareTimeout*time.Second, func() {
						close(self.cTimeout)
					})

					<-self.cTimeout
				case "commit":
					time.AfterFunc(CommitTimeout*time.Second, func() {
						close(self.cTimeout)
					})

					<-self.cTimeout
				case "reply":
					time.AfterFunc(ReplyTimeout*time.Second, func() {
						close(self.cTimeout)
					})

					<-self.cTimeout
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
