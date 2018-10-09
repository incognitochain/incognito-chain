package ppos

import (
	"errors"

	"github.com/ninjadotorg/cash-prototype/common/base58"
)

func (self *Engine) SwitchMember() {

}

func (self *Engine) GetNextCommittee() []string {
	return self.currentCommittee
}

func (self *Engine) CheckCandidate() error {
	return nil
}

func (self *Engine) ProposeCandidate() {

}

func (self *Engine) OnCandidateProposal() {

}

func (self *Engine) OnCandidateVote() {

}

func (self *Engine) OnCandidateRequestTx() {

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

// func getChainValidator(chainID byte, committee []string) (string, error) {
// 	return committee[int((1+int(chainID))%TOTAL_VALIDATORS)], nil
// }
