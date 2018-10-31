package ppos

import (
	"errors"
	"time"

	"github.com/ninjadotorg/cash/common"
	"github.com/ninjadotorg/cash/common/base58"
	"encoding/binary"
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

func (self *Engine) ProposeCandidateToCommittee() {

}

func (self *Engine) CheckCommittee(committee []string, blockHeight int, chainID byte) bool {

	return true
}

func (self *Engine) signData(data []byte) (string, error) {
	signatureByte, err := self.config.ValidatorKeySet.Sign(data)
	if err != nil {
		return "", errors.New("Can't sign data. " + err.Error())
	}
	return base58.Base58Check{}.Encode(signatureByte, byte(0x00)), nil
}

// getMyChain validator chainID and committee of that chainID
func (self *Engine) getMyChain() byte {
	pbk := base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00))
	return self.getChainIdByPbk(pbk)
}

func (self *Engine) getChainIdByPbk(pbk string) byte {
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

func (self *Engine) CommitteeWatcher() {
	self.cQuitCommitteeWatcher = make(chan struct{})
	for {
		select {
		case <-self.cQuitCommitteeWatcher:
			Logger.log.Info("Committee watcher stoppeds")
			return
		case _ = <-self.cNewBlock:

		case <-time.After(common.MaxBlockTime * time.Second):
			self.committee.Lock()
			myPubKey := base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00))
			if common.IndexOfStr(myPubKey, self.committee.CurrentCommittee) != -1 {
				for idx := 0; idx < common.TotalValidators; idx++ {
					if self.committee.CurrentCommittee[idx] != myPubKey {
						go func(validator string) {
							peerIDs := self.config.Server.GetPeerIDsFromPublicKey(validator)
							if len(peerIDs) != 0 {
								// Peer exist
							} else {
								// Peer not exist
							}
						}(self.committee.CurrentCommittee[idx])
					}
				}
			}

			self.committee.Unlock()
		}
	}
}

func (self *Engine) updateCommittee(sealerPbk string, chanId byte) error {
	self.committee.Lock()
	defer self.committee.Unlock()

	committee := make([]string, common.TotalValidators)
	copy(committee, self.committee.CurrentCommittee)

	idx := common.IndexOfStr(sealerPbk, committee)
	if idx >= 0 {
		return errors.New("pbk is existed on committee list")
	}
	currentCommittee := make([]string, common.TotalValidators)
	currentCommittee = append(committee[:chanId], sealerPbk)
	currentCommittee = append(currentCommittee, committee[chanId+1:]...)
	self.committee.CurrentCommittee = currentCommittee
	//remove sealerPbk from candidate list
	for chainId, bestState := range self.config.BlockChain.BestState {
		bestState.RemoveCandidate(sealerPbk)
		self.config.BlockChain.StoreBestState(byte(chainId))
	}

	return nil
}

func (self *Engine) getRawBytesForSwap(lockTime int64, requesterPbk string, chainId byte, sealerPbk string) ([]byte) {
	rawBytes := []byte{}
	bTime := make([]byte, 8)
	binary.LittleEndian.PutUint64(bTime, uint64(lockTime))
	rawBytes = append(rawBytes, bTime...)
	rawBytes = append(rawBytes, []byte(requesterPbk)...)
	rawBytes = append(rawBytes, chainId)
	rawBytes = append(rawBytes, []byte(sealerPbk)...)
	return rawBytes
}
