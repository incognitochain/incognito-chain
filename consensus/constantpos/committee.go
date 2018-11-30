package constantpos

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"encoding/binary"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
)

type committeeStruct struct {
	ValidatorBlkNum      map[string]int //track the number of block created by each validator
	ValidatorReliablePts map[string]int //track how reliable is the validator node
	currentCommittee     []string
	cmWatcherStarted     bool

	CQuitCommitteeWatcher chan struct{}

	sync.Mutex
	LastUpdate int64
}

func (self *committeeStruct) GetCommittee() []string {
	self.Lock()
	defer self.Unlock()
	committee := make([]string, common.TotalValidators)
	copy(committee, self.currentCommittee)
	return committee
}

func (self *committeeStruct) CheckCandidate(candidate string) error {
	return nil
}

func (self *committeeStruct) CheckCommittee(committee []string, blockHeight int, chainID byte) bool {

	return true
}

func (self *committeeStruct) getChainIdByPbk(pbk string) byte {
	committee := self.GetCommittee()
	return byte(common.IndexOfStr(pbk, committee))
}

func (self *committeeStruct) UpdateCommitteePoint(chainLeader string, validatorSig []string) {
	self.Lock()
	defer self.Unlock()
	self.ValidatorBlkNum[chainLeader]++
	self.ValidatorReliablePts[chainLeader] += BlkPointAdd
	for idx, sig := range validatorSig {
		if sig != common.EmptyString {
			self.ValidatorReliablePts[self.currentCommittee[idx]] += SigPointAdd
		}
	}
	for validator := range self.ValidatorReliablePts {
		self.ValidatorReliablePts[validator] += SigPointMin
	}
}

func (self *Engine) StartCommitteeWatcher() {
	if self.committee.cmWatcherStarted {
		Logger.log.Error("Producer already started")
		return
	}
	self.committee.cmWatcherStarted = true
	Logger.log.Info("Committee watcher started")
	for {
		select {
		case <-self.CQuitCommitteeWatcher:
			Logger.log.Info("Committee watcher stopped")
			return
		case _ = <-self.cNewBlock:

		case <-time.After(common.MaxBlockTime * time.Second):
			self.Lock()
			myPubKey := base58.Base58Check{}.Encode(self.config.ProducerKeySet.PaymentAddress.Pk, byte(0x00))
			fmt.Println(myPubKey, common.IndexOfStr(myPubKey, self.currentCommittee))
			if common.IndexOfStr(myPubKey, self.committee.CurrentCommittee) != -1 {
				for idx := 0; idx < common.TotalValidators && self.committee.CurrentCommittee[idx] != myPubKey; idx++ {
					blkTime := time.Since(time.Unix(self.config.BlockChain.BestState[idx].BestBlock.Header.Timestamp, 0))
					fmt.Println(blkTime)
					if blkTime > common.MaxBlockTime*time.Second {

					}
				}
			}

			self.Unlock()
		}
	}
}

func (self *Engine) StopCommitteeWatcher() {
	if self.committee.cmWatcherStarted {
		Logger.log.Info("Stopping Committee watcher...")
		close(self.cQuitCommitteeWatcher)
		self.committee.cmWatcherStarted = false
	}
}

func (self *Engine) updateCommittee(producerPbk string, chanId byte) error {
	self.committee.Lock()
	defer self.committee.Unlock()

	committee := make([]string, common.TotalValidators)
	copy(committee, self.committee.CurrentCommittee)

	idx := common.IndexOfStr(producerPbk, committee)
	if idx >= 0 {
		return errors.New("pbk is existed on committee list")
	}
	currentCommittee := make([]string, common.TotalValidators)
	currentCommittee = append(committee[:chanId], producerPbk)
	currentCommittee = append(currentCommittee, committee[chanId+1:]...)
	self.committee.CurrentCommittee = currentCommittee
	//remove producerPbk from candidate list
	for chainId, bestState := range self.config.BlockChain.BestState {
		bestState.RemoveCandidate(producerPbk)
		self.config.BlockChain.StoreBestState(byte(chainId))
	}

	return nil
}

func (self *Engine) getRawBytesForSwap(lockTime int64, requesterPbk string, chainId byte, producerPbk string) []byte {
	rawBytes := []byte{}
	bTime := make([]byte, 8)
	binary.LittleEndian.PutUint64(bTime, uint64(lockTime))
	rawBytes = append(rawBytes, bTime...)
	rawBytes = append(rawBytes, []byte(requesterPbk)...)
	rawBytes = append(rawBytes, chainId)
	rawBytes = append(rawBytes, []byte(producerPbk)...)
	return rawBytes
}
