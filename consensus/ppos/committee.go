package ppos

import "errors"

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

func getChainValidators(chainID byte, committee []string) ([]string, error) {
	var validators []string
	for index := 1; index <= CHAIN_VALIDATORS_LENGTH; index++ {
		validatorID := (index + int(chainID)) % TOTAL_VALIDATORS
		validators = append(validators, committee[int(validatorID)])
	}
	if len(validators) == CHAIN_VALIDATORS_LENGTH {
		return validators, nil
	}
	return nil, errors.New("can't get chain's validators")
}

func indexOfValidator(validator string, committeeList []string) int {
	for k, v := range committeeList {
		if validator == v {
			return k
		}
	}
	return -1
}
