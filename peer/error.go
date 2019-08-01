package peer

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	// RemotePeer err
	UnexpectedError = iota
	PeerGenerateKeyPairError
	CreateP2PNodeError
	CreateP2PAddressError
	GetPeerIdFromProtocolError
	OpeningStreamP2PError
	HandleNewStreamError

	// PeerConn err
)

var ErrCodeMessage = map[int]struct {
	Code    int
	Message string
}{
	// -1xxx for peer
	UnexpectedError:            {-1001, "Unexpected"},
	PeerGenerateKeyPairError:   {-1001, "Can not generate key pair with reader"},
	CreateP2PNodeError:         {-1002, "Can not create libp2p node"},
	CreateP2PAddressError:      {-1003, "Can not create libp2p address for node"},
	GetPeerIdFromProtocolError: {-1004, "Can not get peer id from protocol"},
	OpeningStreamP2PError:      {-1005, "Fail in opening stream "},
	HandleNewStreamError:       {-1006, "Handle new stream error"},

	// -2xxx for peer connection
}

type PeerError struct {
	err     error
	Code    int
	Message string

	peer *Peer
}

func (e PeerError) Error() string {
	return fmt.Sprintf("%+v: %+v \n %+v", e.Code, e.Message, e.err)
}

func NewPeerError(key int, err error, peer *Peer) *PeerError {
	return &PeerError{
		err:     errors.Wrap(err, ErrCodeMessage[key].Message),
		Code:    ErrCodeMessage[key].Code,
		Message: ErrCodeMessage[key].Message,
		peer:    peer,
	}
}
