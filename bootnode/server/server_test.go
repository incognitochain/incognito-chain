package server

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/wallet"
	"testing"
)

func TestRpcServer_AddOrUpdatePeer(t *testing.T) {
	rpcServer := RpcServer{}
	rpcServer.Init(&RpcServerConfig{
		Port: 9333,
	})

	keyWallet, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	if err != nil {
		t.Error(err)
	}
	err = keyWallet.KeySet.ImportFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	if err != nil {
		t.Error(err)
	}

	args := &PingArgs{
		RawAddress: "localhost:9333",
		PublicKey:  base58.Base58Check{}.Encode(keyWallet.KeySet.PaymentAddress.Pk, common.ZeroByte),
	}
	signDataB58, err := keyWallet.KeySet.SignDataB58([]byte(args.RawAddress))
	if err != nil {
		t.Error(err)
	}
	args.SignData = signDataB58
	rpcServer.AddOrUpdatePeer(args.RawAddress, args.PublicKey, args.SignData)
	if len(rpcServer.Peers) == 0 {
		t.Error("AddOrUpdatePeer fail")
	}
}

func TestRpcServer_RemovePeerByPbk(t *testing.T) {
	rpcServer := RpcServer{}
	rpcServer.Init(&RpcServerConfig{
		Port: 9333,
	})

	keyWallet, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	if err != nil {
		t.Error(err)
	}
	err = keyWallet.KeySet.ImportFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	if err != nil {
		t.Error(err)
	}

	args := &PingArgs{
		RawAddress: "localhost:9333",
		PublicKey:  base58.Base58Check{}.Encode(keyWallet.KeySet.PaymentAddress.Pk, common.ZeroByte),
	}
	signDataB58, err := keyWallet.KeySet.SignDataB58([]byte(args.RawAddress))
	if err != nil {
		t.Error(err)
	}
	args.SignData = signDataB58
	rpcServer.AddOrUpdatePeer(args.RawAddress, args.PublicKey, args.SignData)
	if len(rpcServer.Peers) == 0 {
		t.Error("AddOrUpdatePeer fail")
	}

	rpcServer.RemovePeerByPbk(args.PublicKey)
	if len(rpcServer.Peers) > 0 {
		t.Error("RemovePeerByPbk fail")
	}
}

func TestRpcServer_PeerHeartBeat(t *testing.T) {
	rpcServer := RpcServer{}
	rpcServer.Init(&RpcServerConfig{
		Port: 9333,
	})

	keyWallet, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	if err != nil {
		t.Error(err)
	}
	err = keyWallet.KeySet.ImportFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	if err != nil {
		t.Error(err)
	}

	args := &PingArgs{
		RawAddress: "localhost:9333",
		PublicKey:  base58.Base58Check{}.Encode(keyWallet.KeySet.PaymentAddress.Pk, common.ZeroByte),
	}
	signDataB58, err := keyWallet.KeySet.SignDataB58([]byte(args.RawAddress))
	if err != nil {
		t.Error(err)
	}
	args.SignData = signDataB58
	rpcServer.AddOrUpdatePeer(args.RawAddress, args.PublicKey, args.SignData)
	if len(rpcServer.Peers) == 0 {
		t.Error("AddOrUpdatePeer fail")
	}

	go rpcServer.PeerHeartBeat(6)
	for {
		if len(rpcServer.Peers) == 0 {
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
