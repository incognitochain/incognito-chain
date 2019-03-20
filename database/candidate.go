package database

import (
	"bytes"
	"github.com/constant-money/constant-chain/privacy"
)

type CandidateElement struct {
	PaymentAddress privacy.PaymentAddress
	VoteAmount     uint64
	NumberOfVote   uint32
}

type CandidateList []CandidateElement

func (A CandidateList) Len() int {
	return len(A)
}
func (A CandidateList) Swap(i, j int) {
	A[i], A[j] = A[j], A[i]
}
func (A CandidateList) Less(i, j int) bool {
	return A[i].VoteAmount < A[j].VoteAmount ||
		(A[i].VoteAmount == A[j].VoteAmount && A[i].NumberOfVote < A[j].NumberOfVote) ||
		(A[i].VoteAmount == A[j].VoteAmount && A[i].NumberOfVote == A[j].NumberOfVote &&
			bytes.Compare(A[i].PaymentAddress.Bytes(), A[j].PaymentAddress.Bytes()) == -1)
}
