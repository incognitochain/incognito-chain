package constantpos

import (
	"errors"
	"sync"
)

type BFTProtocol struct {
	sync.Mutex
	Phase   string
	cQuit   chan struct{}
	cTimeout chan struct{}
	started bool
}

func (self *BFTProtocol) Start() {
	self.Lock()
	defer.Unlock()
	if self.started {
		return errors.New("Consensus engine is already started")
	}
	self.started = true
	self.cQuit = make(chan struct{})
	go func ()  {
		for {
			self.cTimeout = make(chan struct{})
			select {
			case <-self.cQuit:
				return
			default:
				switch self.Phase {
				case "propose":
					
					<-self.cTimeout 
				case "listen":
	
					<-self.cTimeout 
				case "prepare":

					<-self.cTimeout 
				case "commit":

					<-self.cTimeout 
				case "reply":

					<-self.cTimeout 
				}
			}
	
		}
	}

}

func (self *BFTProtocol) Stop() {
	self.Lock()
	defer.Unlock()
	if !self.started {
		return errors.New("Consensus engine is already started")
	}
	self.started = false
	close(self.cTimeout)
	close(self.cQuit)
}
