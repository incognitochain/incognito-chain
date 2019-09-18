package rpcservice

import "github.com/incognitochain/incognito-chain/connmanager"

type NetworkService struct{
	ConnMgr *connmanager.ConnManager
}

func (networkService  NetworkService) GetConnectionCount() int {
	if networkService.ConnMgr == nil || networkService.ConnMgr.GetListeningPeer() == nil {
		return 0
	}

	listeningPeer := networkService.ConnMgr.GetListeningPeer()
	return len(listeningPeer.GetPeerConns())
}
