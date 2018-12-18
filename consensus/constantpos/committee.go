package constantpos

import (
	"sync"

	"encoding/binary"

	"github.com/ninjadotorg/constant/common"
)

type CommitteeStruct struct {
	CurrentCommittee      []string
	cQuitCommitteeWatcher chan struct{}
	cmWatcherStarted      bool

	sync.Mutex
	LastUpdate int64
}

func (self *CommitteeStruct) UpdateCommittee(newCommitee []string) {
	self.Lock()
	defer self.Unlock()
	self.CurrentCommittee = make([]string, common.TotalValidators)
	copy(newCommitee, self.CurrentCommittee)
	// self.LastUpdate
}

func (self *CommitteeStruct) GetCommittee() []string {
	self.Lock()
	defer self.Unlock()
	committee := make([]string, common.TotalValidators)
	copy(committee, self.CurrentCommittee)
	return committee
}

func (self *CommitteeStruct) CheckCandidate(candidate string) error {
	return nil
}

func (self *CommitteeStruct) CheckCommittee(committee []string, blockHeight int, shardID byte) bool {

	return true
}

// func (self *CommitteeStruct) UpdateCommitteePoint(chainLeader string, validatorSig []string) {
// 	self.Lock()
// 	defer self.Unlock()
// 	self.ValidatorBlkNum[chainLeader]++
// 	self.ValidatorReliablePts[chainLeader] += BlkPointAdd
// 	for idx, sig := range validatorSig {
// 		if sig != common.EmptyString {
// 			self.ValidatorReliablePts[self.CurrentCommittee[idx]] += SigPointAdd
// 		}
// 	}
// 	for validator := range self.ValidatorReliablePts {
// 		self.ValidatorReliablePts[validator] += SigPointMin
// 	}
// }

// func (self *CommitteeStruct) UpdateCommittee(producerPbk string, chanId byte) error {
// 	self.Lock()
// 	defer self.Unlock()

// 	committee := make([]string, common.TotalValidators)
// 	copy(committee, self.CurrentCommittee)

// 	idx := common.IndexOfStr(producerPbk, committee)
// 	if idx >= 0 {
// 		return errors.New("pbk is existed on committee list")
// 	}
// 	currentCommittee := make([]string, common.TotalValidators)
// 	currentCommittee = append(committee[:chanId], producerPbk)
// 	currentCommittee = append(currentCommittee, committee[chanId+1:]...)
// 	self.CurrentCommittee = currentCommittee
// 	//remove producerPbk from candidate list
// 	// for shardID, bestState := range self.config.BlockChain.BestState {
// 	// 	bestState.RemoveCandidate(producerPbk)
// 	// 	self.config.BlockChain.StoreBestState(byte(shardID))
// 	// }

// 	return nil
// }

func (self *CommitteeStruct) StartCommitteeWatcher() {
	// if self.cmWatcherStarted {
	// 	Logger.log.Error("Producer already started")
	// 	return
	// }
	// self.cmWatcherStarted = true
	// Logger.log.Info("Committee watcher started")
	// for {
	// 	select {
	// 	case <-self.cQuitCommitteeWatcher:
	// 		Logger.log.Info("Committee watcher stopped")
	// 		return
	// 	case _ = <-self.cNewBlock:

	// 	case <-time.After(common.MaxBlockTime * time.Second):
	// 		self.Lock()
	// 		myPubKey := base58.Base58Check{}.Encode(self.config.ProducerKeySet.PaymentAddress.Pk, byte(0x00))
	// 		fmt.Println(myPubKey, common.IndexOfStr(myPubKey, self.Committee.CurrentCommittee))
	// 		if common.IndexOfStr(myPubKey, self.Committee.CurrentCommittee) != -1 {
	// 			for idx := 0; idx < common.TotalValidators && self.Committee.CurrentCommittee[idx] != myPubKey; idx++ {
	// 				blkTime := time.Since(time.Unix(self.config.BlockChain.BestState[idx].BestBlock.Header.Timestamp, 0))
	// 				fmt.Println(blkTime)
	// 				if blkTime > common.MaxBlockTime*time.Second {

	// 				}
	// 			}
	// 		}

	// 		self.Unlock()
	// 	}
	// }
}

func (self *CommitteeStruct) StopCommitteeWatcher() {
	if self.cmWatcherStarted {
		Logger.log.Info("Stopping Committee watcher...")
		close(self.cQuitCommitteeWatcher)
		self.cmWatcherStarted = false
	}
}

func getRawBytesForSwap(lockTime int64, requesterPbk string, shardID byte, producerPbk string) []byte {
	rawBytes := []byte{}
	bTime := make([]byte, 8)
	binary.LittleEndian.PutUint64(bTime, uint64(lockTime))
	rawBytes = append(rawBytes, bTime...)
	rawBytes = append(rawBytes, []byte(requesterPbk)...)
	rawBytes = append(rawBytes, shardID)
	rawBytes = append(rawBytes, []byte(producerPbk)...)
	return rawBytes
}
