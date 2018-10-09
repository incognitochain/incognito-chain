package transaction

import "github.com/ninjadotorg/cash-prototype/privacy/client"

// VotingData ...
type VotingData struct {
	NodeAddress     string
	SpublicKey      []byte
	SpendingAddress [client.SpendingAddressLength]byte
	TransmissionKey [client.TransmissionKeyLength]byte
	ReceivingKey    [client.ReceivingKeyLength]byte
}
