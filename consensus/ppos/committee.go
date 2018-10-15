package ppos

import (
	"errors"
	"sync"

	"github.com/ninjadotorg/cash-prototype/cashec"

	"github.com/ninjadotorg/cash-prototype/common"
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
	pkey := base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00))
	for idx := byte(0); idx < byte(common.TotalValidators); idx++ {
		validator := self.currentCommittee[int((1+int(idx))%common.TotalValidators)]
		if pkey == validator {
			return idx
		}
	}
	return common.TotalValidators // nope, you're not in the committee
}

type ValidatorList struct {
	sync.Mutex
	Committee []cashec.KeySetSealer
	Candidate []cashec.KeySetSealer
}
