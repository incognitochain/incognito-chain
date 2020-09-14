package slashing

import "github.com/incognitochain/incognito-chain/incognitokey"

type SlashMissingSignature interface {
	AddMissingSignature(validationData string, committees []*incognitokey.CommitteePublicKey) error
	GetAllSlashingPenalty() map[string]uint64
	GetSlashingPenalty(key *incognitokey.CommitteePublicKey)
	Reset()
}
