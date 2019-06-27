package server

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/incognitochain/incognito-chain/wire"
	"testing"
)

func TestHandler_Ping(t *testing.T) {
	rpcServer := RpcServer{}
	rpcServer.Init(&RpcServerConfig{
		Port: 9333,
	})
	handler := Handler{rpcServer: &rpcServer}
	keyWallet, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	if err != nil {
		t.Error(err)
	}
	keyWallet.KeySet.ImportFromPrivateKey(&keyWallet.KeySet.PrivateKey)

	args := &PingArgs{
		RawAddress: "localhost:9333",
		PublicKey:  base58.Base58Check{}.Encode(keyWallet.KeySet.PaymentAddress.Pk, common.ZeroByte),
	}
	signDataB58, err := keyWallet.KeySet.SignDataB58([]byte(args.RawAddress))
	if err != nil {
		t.Error(err)
	}
	args.SignData = signDataB58

	var response = make([]wire.RawPeer, 0)
	err = handler.Ping(args, &response)
	if err != nil {
		t.Error(err)
	}

}
