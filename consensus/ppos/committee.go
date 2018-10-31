package ppos

import (
	"errors"
	"time"

	"github.com/ninjadotorg/cash/common"
	"github.com/ninjadotorg/cash/common/base58"
)

func (self *Engine) GetCommittee() []string {
	if len(self.Committee) <= 0 {
		self.Committee = make([]string, len(self.config.BlockChain.BestState[0].BestBlock.Header.Committee))
		copy(self.Committee, self.config.BlockChain.BestState[0].BestBlock.Header.Committee)
	}
	return self.Committee
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
	committee := self.GetCommittee()
	pkey := base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00))
	for idx := byte(0); idx < byte(common.TotalValidators); idx++ {
		validator := committee[int((1+int(idx))%common.TotalValidators)]
		if pkey == validator {
			return idx
		}
	}
	return common.TotalValidators // nope, you're not in the committee
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
