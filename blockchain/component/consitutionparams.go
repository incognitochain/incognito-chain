package component

import (
	"bytes"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type SubmitProposalInfo struct {
	ExecuteDuration   uint64
	Explanation       string
	PaymentAddress    privacy.PaymentAddress
	ConstitutionIndex uint32
}

type Voter struct {
	PaymentAddress privacy.PaymentAddress
	AmountOfVote   int32
}

func (voter *Voter) Greater(voter2 Voter) bool {
	return voter.AmountOfVote > voter2.AmountOfVote ||
		(voter.AmountOfVote == voter2.AmountOfVote && bytes.Compare(voter.PaymentAddress.Bytes(), voter2.PaymentAddress.Bytes()) > 0)
}

func (voter *Voter) ToBytes() []byte {
	record := voter.PaymentAddress.String()
	record += string(voter.AmountOfVote)
	return []byte(record)
}

func (submitProposalInfo SubmitProposalInfo) ToBytes() []byte {
	record := string(common.Uint64ToBytes(submitProposalInfo.ExecuteDuration))
	record += submitProposalInfo.Explanation
	record += string(submitProposalInfo.PaymentAddress.Bytes())
	record += string(common.Uint32ToBytes(submitProposalInfo.ConstitutionIndex))
	return []byte(record)
}

func (submitProposalInfo SubmitProposalInfo) ValidateSanityData() bool {
	if submitProposalInfo.ExecuteDuration < common.MinimumBlockOfProposalDuration ||
		submitProposalInfo.ExecuteDuration > common.MaximumBlockOfProposalDuration {
		return false
	}
	if len(submitProposalInfo.Explanation) > common.MaximumProposalExplainationLength {
		return false
	}
	return true
}

func (submitProposalInfo SubmitProposalInfo) ValidateTxWithBlockChain(
	boardType common.BoardType,
	chainID byte,
	db database.DatabaseInterface,
) bool {
	//if br.GetConstitutionEndHeight(DCBBoard, chainID)+submitProposalInfo.ExecuteDuration+common.MinimumBlockOfProposalDuration >
	//	br.GetBoardEndHeight(boardType, chainID) {
	//	return false
	//}
	return true
}
