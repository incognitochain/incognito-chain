package mock

import "github.com/incognitochain/incognito-chain/common"

type Server struct{}

func (s *Server) PushBlockToAll(block common.BlockInterface, isBeacon bool) error {
	return nil
}
