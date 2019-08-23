package jsonresult

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/connmanager"
	"github.com/incognitochain/incognito-chain/wallet"
	"net"
	"os"

	"github.com/incognitochain/incognito-chain/rpcserver"
)

type GetNetworkInfoResult struct {
	Commit          string                   `json:"commit"`
	Version         string                   `json:"version"`
	SubVersion      string                   `json:"SubVersion"`
	ProtocolVersion string                   `json:"ProtocolVersion"`
	NetworkActive   bool                     `json:"NetworkActive"`
	Connections     int                      `json:"Connections"`
	Networks        []map[string]interface{} `json:"Networks"`
	LocalAddresses  []string                 `json:"LocalAddresses"`
	IncrementalFee  uint64                   `json:"IncrementalFee"`
	Warnings        string                   `json:"Warnings"`
}

func NewGetNetworkInfoResult(protocolVerion string, connMgr connmanager.ConnManager, wallet *wallet.Wallet) (*GetNetworkInfoResult, error) {
	result := &GetNetworkInfoResult{
		Commit:          os.Getenv("commit"),
		ProtocolVersion: protocolVerion,
		Version:         rpcserver.RpcServerVersion,
		SubVersion:      common.EmptyString,
		NetworkActive:   connMgr.GetListeningPeer() != nil,
		LocalAddresses:  []string{},
	}
	listener := connMgr.GetListeningPeer()
	result.Connections = len(listener.GetPeerConns())
	result.LocalAddresses = append(result.LocalAddresses, listener.GetRawAddress())
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	networks := []map[string]interface{}{}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			network := map[string]interface{}{}
			network["name"] = "ipv4"
			network["limited"] = false
			network["reachable"] = true
			network["proxy"] = ""
			network["proxy_randomize_credentials"] = false
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To16() != nil {
					network["name"] = "ipv6"
				}
			}
			networks = append(networks, network)
		}
	}
	result.Networks = networks
	if wallet != nil && wallet.GetConfig() != nil {
		result.IncrementalFee = wallet.GetConfig().IncrementalFee
	}
	result.Warnings = common.EmptyString
	return result, nil
}
