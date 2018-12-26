package voting

import "github.com/ninjadotorg/constant/common"

type Voter struct {
	PubKey       []byte
	AmountOfVote int32
}

func (voter *Voter) Greater(voter2 Voter) bool {
	return voter.AmountOfVote > voter2.AmountOfVote ||
		(voter.AmountOfVote == voter2.AmountOfVote && string(voter.PubKey) > string(voter2.PubKey))
}

func (voter *Voter) Hash() *common.Hash {
	record := string(voter.PubKey)
	record += string(voter.AmountOfVote)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

type ProposalVote struct {
	TxId         common.Hash
	AmountOfVote int64
	NumberOfVote uint32
}

func (proposalVote ProposalVote) Greater(proposalVote2 ProposalVote) bool {
	return proposalVote.AmountOfVote > proposalVote2.AmountOfVote ||
		(proposalVote.AmountOfVote == proposalVote2.AmountOfVote || proposalVote.NumberOfVote > proposalVote2.NumberOfVote) ||
		(proposalVote.AmountOfVote == proposalVote2.AmountOfVote || proposalVote.NumberOfVote == proposalVote2.NumberOfVote || string(proposalVote.TxId.GetBytes()) > string(proposalVote2.TxId.GetBytes()))
}
