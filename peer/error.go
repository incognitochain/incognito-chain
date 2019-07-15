package peer

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	// RemotePeer err
	UnexpectedErr = iota
	PeerGenerateKeyPairErr
	CreateP2PNodeErr
	CreateP2PAddressErr
	GetPeerIdFromProtocolErr
	OpeningStreamP2PErr

	// PeerConn err
)

var ErrCodeMessage = map[int]struct {
	code    int
	message string
}{
	// -1xxx for peer
	UnexpectedErr:            {-1001, "Unexpected"},
	PeerGenerateKeyPairErr:   {-1001, "Can not generate key pair with reader"},
	CreateP2PNodeErr:         {-1002, "Can not create libp2p node"},
	CreateP2PAddressErr:      {-1003, "Can not create libp2p address for node"},
	GetPeerIdFromProtocolErr: {-1004, "Can not get peer id from protocol"},
	OpeningStreamP2PErr:      {-1005, "Fail in opening stream "},

	// -2xxx for peer connection
}

type PeerError struct {
	err     error
	code    int
	message string

	peer *Peer
}

func (e PeerError) Error() string {
	return fmt.Sprintf("%+v: %+v \n %+v", e.code, e.message, e.err)
}

func NewPeerError(key int, err error, peer *Peer) *PeerError {
	return &PeerError{
		err:     errors.Wrap(err, ErrCodeMessage[key].message),
		code:    ErrCodeMessage[key].code,
		message: ErrCodeMessage[key].message,
		peer:    peer,
	}
}
