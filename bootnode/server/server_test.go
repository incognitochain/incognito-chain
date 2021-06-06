package server

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus_v2/blsbft"
)

func TestRpcServer_AddOrUpdatePeer(t *testing.T) {
	rpcServer := RpcServer{}
	rpcServer.Init(&RpcServerConfig{
		Port: 9333,
	})

	blsBft := blsbft.BLSBFT{}
	privateSeed, err := blsBft.LoadUserKeyFromIncPrivateKey("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	if err != nil {
		t.Error(err)
	}
	blsBft.LoadUserKey(privateSeed)
	blsPublicKeyBytes := blsBft.UserKeySet.GetPublicKey().MiningPubKey[common.BridgeConsensus]

	args := &PingArgs{
		RawAddress: "localhost:9333",
		PublicKey:  base58.Base58Check{}.Encode(blsPublicKeyBytes, common.ZeroByte),
	}
	signDataInByte, err := blsBft.UserKeySet.BriSignData([]byte(args.RawAddress)) //, 0, []blsmultisig.PublicKey{blsBft.UserKeySet.GetPublicKey().MiningPubKey[common.BlsConsensus]})
	if err != nil {
		t.Error(err)
	}
	args.SignData = base58.Base58Check{}.Encode(signDataInByte, common.ZeroByte)
	rpcServer.AddOrUpdatePeer(args.RawAddress, common.BlsConsensus, args.PublicKey, args.SignData)
	if len(rpcServer.peers) == 0 {
		t.Error("AddOrUpdatePeer fail")
	}
}

func TestRpcServer_RemovePeerByPbk(t *testing.T) {
	rpcServer := RpcServer{}
	rpcServer.Init(&RpcServerConfig{
		Port: 9333,
	})

	blsBft := blsbft.BLSBFT{}
	privateSeed, err := blsBft.LoadUserKeyFromIncPrivateKey("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	if err != nil {
		t.Error(err)
	}
	blsBft.LoadUserKey(privateSeed)
	briPublicKeyBytes := blsBft.UserKeySet.GetPublicKey().MiningPubKey[common.BridgeConsensus]

	args := &PingArgs{
		RawAddress: "localhost:9333",
		PublicKey:  base58.Base58Check{}.Encode(briPublicKeyBytes, common.ZeroByte),
	}
	signDataInByte, err := blsBft.UserKeySet.BriSignData([]byte(args.RawAddress)) //, 0, []blsmultisig.PublicKey{blsBft.UserKeySet.GetPublicKey().MiningPubKey[common.BlsConsensus]})
	if err != nil {
		t.Error(err)
	}
	args.SignData = base58.Base58Check{}.Encode(signDataInByte, common.ZeroByte)

	rpcServer.AddOrUpdatePeer(args.RawAddress, common.BlsConsensus, args.PublicKey, args.SignData)
	if len(rpcServer.peers) == 0 {
		t.Error("AddOrUpdatePeer fail")
	}

	rpcServer.RemovePeerByPbk(args.PublicKey)
	if len(rpcServer.peers) > 0 {
		t.Error("RemovePeerByPbk fail")
	}
}

func TestRpcServer_PeerHeartBeat(t *testing.T) {
	rpcServer := RpcServer{}
	rpcServer.Init(&RpcServerConfig{
		Port: 9333,
	})

	blsBft := blsbft.BLSBFT{}
	privateSeed, err := blsBft.LoadUserKeyFromIncPrivateKey("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	if err != nil {
		t.Error(err)
	}
	blsBft.LoadUserKey(privateSeed)
	briPublicKeyBytes := blsBft.UserKeySet.GetPublicKey().MiningPubKey[common.BridgeConsensus]

	args := &PingArgs{
		RawAddress: "localhost:9333",
		PublicKey:  base58.Base58Check{}.Encode(briPublicKeyBytes, common.ZeroByte),
	}
	signDataInByte, err := blsBft.UserKeySet.BriSignData([]byte(args.RawAddress)) //, 0, []blsmultisig.PublicKey{blsBft.UserKeySet.GetPublicKey().MiningPubKey[common.BlsConsensus]})
	if err != nil {
		t.Error(err)
	}
	args.SignData = base58.Base58Check{}.Encode(signDataInByte, common.ZeroByte)
	rpcServer.AddOrUpdatePeer(args.RawAddress, common.BlsConsensus, args.PublicKey, args.SignData)
	if len(rpcServer.peers) == 0 {
		t.Error("AddOrUpdatePeer fail")
	}

	go rpcServer.PeerHeartBeat(6)
	for {
		if len(rpcServer.peers) == 0 {
			t.Log("PeerHeartBeat")
			return
		}
	}

}

func TestRpcServer_Start(t *testing.T) {
	rpcServer := RpcServer{}
	rpcServer.Init(&RpcServerConfig{
		Port: 9333,
	})

	go rpcServer.Start()
}
