package jsonresult

type GetAmountVoteTokenResult struct {
	DCBVoteTokenAmount uint32 `json:dcbVoteTokenAmount`
	GOVVoteTokenAmount uint32 `json:govVoteTokenAmount`
}
