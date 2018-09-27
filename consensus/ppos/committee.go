package ppos

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
