package ppos

import (
	"errors"
	"time"

	"encoding/binary"

	"github.com/ninjadotorg/cash/blockchain"
	"github.com/ninjadotorg/cash/common"
	"github.com/ninjadotorg/cash/common/base58"
	"github.com/ninjadotorg/cash/transaction"
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
	signatureByte, err := self.config.ProducerKeySet.Sign(data)
	if err != nil {
		return common.EmptyString, errors.New("Can't sign data. " + err.Error())
	}
	return base58.Base58Check{}.Encode(signatureByte, byte(0x00)), nil
}

// getMyChain validator chainID and committee of that chainID
func (self *Engine) getMyChain() byte {
	pbk := base58.Base58Check{}.Encode(self.config.ProducerKeySet.SpublicKey, byte(0x00))
	return self.getChainIdByPbk(pbk)
}

func (self *Engine) getChainIdByPbk(pbk string) byte {
	committee := self.GetCommittee()
	return byte(common.IndexOfStr(pbk, committee))
}

func (self *Engine) GetCandidateCommitteeList(block *blockchain.Block) map[string]blockchain.CommitteeCandidateInfo {
	bestState := self.config.BlockChain.BestState[block.Header.ChainID]
	candidates := bestState.Candidates
	if candidates == nil {
		candidates = make(map[string]blockchain.CommitteeCandidateInfo)
	}
	for _, tx := range block.Transactions {
		if tx.GetType() == common.TxVotingType {
			txV, ok := tx.(*transaction.TxVoting)
			nodeAddr := txV.PublicKey
			cndVal, ok := candidates[nodeAddr]
			_ = cndVal
			if !ok {
				candidates[nodeAddr] = blockchain.CommitteeCandidateInfo{
					Value:     txV.GetValue(),
					Timestamp: block.Header.Timestamp,
					ChainID:   block.Header.ChainID,
				}
			} else {
				candidates[nodeAddr] = blockchain.CommitteeCandidateInfo{
					Value:     cndVal.Value + txV.GetValue(),
					Timestamp: block.Header.Timestamp,
					ChainID:   block.Header.ChainID,
				}
			}
		}
	}
	return candidates
}

func (committee *committeeStruct) UpdateCommitteePoint(chainLeader string, validatorSig []string) {
	committee.Lock()
	defer committee.Unlock()
	committee.ValidatorBlkNum[chainLeader]++
	committee.ValidatorReliablePts[chainLeader] += BlkPointAdd
	for idx, sig := range validatorSig {
		if sig != common.EmptyString {
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
			myPubKey := base58.Base58Check{}.Encode(self.config.ProducerKeySet.SpublicKey, byte(0x00))
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
