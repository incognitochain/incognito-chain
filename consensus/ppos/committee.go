package ppos

import (
	"errors"
	"fmt"
	"time"

	"encoding/binary"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
)

func (self *Engine) GetCommittee() []string {
	self.committee.Lock()
	defer self.committee.Unlock()
	committee := make([]string, common.TotalValidators)
	copy(committee, self.committee.CurrentCommittee)
	return committee
}

func (self *Engine) CheckCandidate(candidate string) error {
	return nil
}

func (self *Engine) CheckCommittee(committee []string, blockHeight int, shardID byte) bool {

	return true
}

func (self *Engine) signData(data []byte) (string, error) {
	signatureByte, err := self.config.ProducerKeySet.Sign(data)
	if err != nil {
		return "", errors.New("Can't sign data. " + err.Error())
	}
	return base58.Base58Check{}.Encode(signatureByte, byte(0x00)), nil
}

// getMyChain validator shardID and committee of that shardID
func (self *Engine) getMyChain() byte {
	pbk := base58.Base58Check{}.Encode(self.config.ProducerKeySet.PaymentAddress.Pk, byte(0x00))
	return self.getshardIDByPbk(pbk)
}

func (self *Engine) getshardIDByPbk(pbk string) byte {
	committee := self.GetCommittee()
	return byte(common.IndexOfStr(pbk, committee))
}

func (committee *committeeStruct) UpdateCommitteePoint(chainLeader string, validatorSig []string) {
	committee.Lock()
	defer committee.Unlock()
	committee.ValidatorBlkNum[chainLeader]++
	committee.ValidatorReliablePts[chainLeader] += BlkPointAdd
	for idx, sig := range validatorSig {
		if sig != "" {
			committee.ValidatorReliablePts[committee.CurrentCommittee[idx]] += SigPointAdd
		}
	}
	for validator := range committee.ValidatorReliablePts {
		committee.ValidatorReliablePts[validator] += SigPointMin
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
		case <-self.cQuitCommitteeWatcher:
			Logger.log.Info("Committee watcher stopped")
			return
		case _ = <-self.cNewBlock:

		case <-time.After(common.MaxBlockTime * time.Second):
			self.committee.Lock()
			myPubKey := base58.Base58Check{}.Encode(self.config.ProducerKeySet.PaymentAddress.Pk, byte(0x00))
			fmt.Println(myPubKey, common.IndexOfStr(myPubKey, self.committee.CurrentCommittee))
			if common.IndexOfStr(myPubKey, self.committee.CurrentCommittee) != -1 {
				for idx := 0; idx < common.TotalValidators && self.committee.CurrentCommittee[idx] != myPubKey; idx++ {
					blkTime := time.Since(time.Unix(self.config.BlockChain.BestState[idx].BestBlock.Header.Timestamp, 0))
					fmt.Println(blkTime)
					if blkTime > common.MaxBlockTime*time.Second {

					}
				}
			}

			self.committee.Unlock()
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
	for shardID, bestState := range self.config.BlockChain.BestState {
		bestState.RemoveCandidate(producerPbk)
		self.config.BlockChain.StoreBestState(byte(shardID))
	}

	return nil
}

func (self *Engine) getRawBytesForSwap(lockTime int64, requesterPbk string, shardID byte, producerPbk string) []byte {
	rawBytes := []byte{}
	bTime := make([]byte, 8)
	binary.LittleEndian.PutUint64(bTime, uint64(lockTime))
	rawBytes = append(rawBytes, bTime...)
	rawBytes = append(rawBytes, []byte(requesterPbk)...)
	rawBytes = append(rawBytes, shardID)
	rawBytes = append(rawBytes, []byte(producerPbk)...)
	return rawBytes
}
