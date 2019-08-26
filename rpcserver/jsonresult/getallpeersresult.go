package jsonresult

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/addrmanager"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/connmanager"
)

type GetAllPeersResult struct {
	Peers []string `json:"Peers"`
}

func NewGetAllPeersResult(addrMgr addrmanager.AddrManager) *GetAllPeersResult {
	result := &GetAllPeersResult{}
	peersMap := []string{}
	peers := addrMgr.AddressCache()
	for _, peer := range peers {
		for _, peerConn := range peer.GetPeerConns() {
			peersMap = append(peersMap, peerConn.GetRemoteRawAddress())
		}
	}
	result.Peers = peersMap
	return result
}

type GetAllConnectedPeersResult struct {
	Peers []map[string]string `json:"Peers"`
}

func NewGetAllConnectedPeersResult(connMgr connmanager.ConnManager) *GetAllConnectedPeersResult {
	result := &GetAllConnectedPeersResult{}
	peersMap := []map[string]string{}
	listeningPeer := connMgr.GetListeningPeer()
	bestState := blockchain.GetBeaconBestState()
	beaconCommitteeList := bestState.BeaconCommittee
	shardCommitteeList := bestState.GetShardCommittee()

	for _, peerConn := range listeningPeer.GetPeerConns() {
		peerItem := map[string]string{
			"RawAddress": peerConn.GetRemoteRawAddress(),
			"PublicKey":  peerConn.GetRemotePeer().GetPublicKey(),
			"NodeType":   "",
		}
		isInBeaconCommittee := common.IndexOfStr(peerConn.GetRemotePeer().GetPublicKey(), beaconCommitteeList) != -1
		if isInBeaconCommittee {
			peerItem["NodeType"] = "Beacon"
		}
		for shardID, committees := range shardCommitteeList {
			isInShardCommitee := common.IndexOfStr(peerConn.GetRemotePeer().GetPublicKey(), committees) != -1
			if isInShardCommitee {
				peerItem["NodeType"] = fmt.Sprintf("Shard-%d", shardID)
				break
			}
		}
		peersMap = append(peersMap, peerItem)
	}
	result.Peers = peersMap
	return result
}
