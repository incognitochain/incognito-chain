package peer

import "fmt"

const (
	PeerGenerateKeyPairErr   = "PeerGenerateKeyPairErr"
	CreateP2PNodeErr         = "CreateP2PNodeErr"
	CreateP2PAddressErr      = "CreateP2PAddressErr"
	GetPeerIdFromProtocolErr = "GetPeerIdFromProtocolErr"
	OpeningStreamP2PErr      = "OpeningStreamP2PErr"
)

var ErrCodeMessage = map[string]struct {
	code    int
	message string
}{
	// -1xxx for peer
	PeerGenerateKeyPairErr:   {-1000, "Can not generate key pair with reader"},
	CreateP2PNodeErr:         {-1001, "Can not create libp2p node"},
	CreateP2PAddressErr:      {-1002, "Can not create libp2p address for node"},
	GetPeerIdFromProtocolErr: {-1003, "Can not get peer id from protocol"},
	OpeningStreamP2PErr:      {-1004, "Fail in opening stream "},
}

type PeerError struct {
	err     error
	code    int
	message string

	peer *Peer
}

func (e PeerError) Error() string {
	return fmt.Sprintf("%v: %v", e.code, e.message)
}

func NewPeerError(key string, err error, peer *Peer) *PeerError {
	return &PeerError{
		err:     err,
		code:    ErrCodeMessage[key].code,
		message: ErrCodeMessage[key].message,
	}
}
