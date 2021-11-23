package consensustypes

import "time"

type BlackListValidator struct {
	StartTime  time.Time
	ErrorName  error
	FirstVote  *BFTVote
	SecondVote *BFTVote
	TTL        time.Duration
}

func NewBlackListValidator() *BlackListValidator {
	return &BlackListValidator{}
}
