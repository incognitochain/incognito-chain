package ppos

import (
	"errors"

	"github.com/ninjadotorg/cash/common"
	"github.com/ninjadotorg/cash/common/base58"
)

func (self *Engine) GetCommittee() []string {



	return self.config.BlockChain.BestState[0].BestBlock.Header.Committee
}

func (self *Engine) CheckCandidate() error {
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

func (committee *committeeStruct) UpdateCommittee(chainLeader string, validatorSig []string) {
	committee.Lock()
	defer committee.Unlock()
	committee.ValidatorBlkNum[chainLeader]++
	for idx, sig := range validatorSig {
		if sig != "" {
			committee.ValidatorReliablePts[committee.CurrentCommittee[idx]]++
		}
	}
	for validator := range committee.ValidatorReliablePts {
		committee.ValidatorReliablePts[validator]--
	}
}
