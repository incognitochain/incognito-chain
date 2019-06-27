package netsync

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wire"
	"testing"
)
var _= func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()
func TestNetSyncHandleTxWithRole(t *testing.T) {
	netSync := NetSync{}.New(&NetSyncConfig{
		RoleInCommittees: 0,
		PubSubManager: pubsub.NewPubSubManager(),
	})
	base58CheckData := "13AHwQTSToFvjZFGjxKZbVKL6Mq2sLaGyevhwFUGMvdNd7745xSXhzJ2q2ra6d8pmb8CgieHM4zgbMJSVpsxJFhprUzUu9Ba7FWYcdmHCkqviF1LhysmBugQP7mEGQVo6PJrFp7Xim72ndfwPovTLhWrqYa8475MAM8XMKMEtYRvNXjPghMJuYvN9yxPPWzXUFQGnw3temoR3heaXMCwH4ZGMVWVaom2prZZwQqdKQkC3Yi2E7hjntDny2oCVsvyHpCv7nAZganw9ZqXc2H4V15GNnvTKWUBRxZo8tRK79tnyoRH23SNtPrc7hG3tJc8gtLoZ2FLQP2UrGPi88XpnzNagTix3ydZ3HUMu29bBRJALKdREn7jsMPECN8EjgZayx2XuZNZGJ1Tqzdq7dosUB1TfKfdXk9rVQXuLqxSM7sszST2kdYQ4sLB5YnpcePtFpX3SsR2fbBB7NWNzdq8JroWiLcAcxmH5jVMtkXTkrdLYtEdHDDSm1RQgPEzTkKT3479Hoge5ocDeVBdEqp8v21vbTLvBYWW7cjy3UH9KucqCRWWN7Tiztbnujs3nkazAAHBLFgSSbmVHD9MGKScqnB249ECGqeoHwCnuTTC2GPehyFb3N67mR9bHNFfJz1MG7qwAt3CDVACR68FqCrBpEkkk1ko2jhedYe2ZAmDAWgL8HWvXK112Ji3ucwyaNm3CvjGDWpyyYAs7TjkF58RtMGojwGCutHNoNHwkEtdGoekRoHfVrTQgmgfsxEsAPbL4JcCNKzjMReDAwGVNXu4kWF6KwFtyD5z6jQyYZZX9aEm1fcMmTvFcmX1DBUjsW4BnELPZR6iYeJkikL7bYZZPLmfzKenD17jFLe4F5xhwFcBPN2kxjTFsLKG6VT2EKeShJGzuWgNAbBVBjgjECk6jrZC44HaXgUTdRcgr3MEfp1JAgegCaKJTaXn3eacHqPQrDovMoJ1wijwok4DNFWhA2Ano75udeZHZjHEvytsdEtL1VV74Rs6dFjz591EHvGqvCVFx9qgTRHmSxRDkGox2rYD5LFXSDsL9AZ2aLaoisysT21Dc81WNuG36QhgRjpZ9LZeaCVL8ffjDV5Wbe8uqedN8uZy9P6ZXothXgCjBiAxsDgyHLAbMqND742Sq3wh2N1mTgTRg75S8ijN1smca3U7Snj2iELXRhLrwr6zR8aUxWe4Tdy19knpbEwBoUQcd5i9pwii3q67nYSKLYQZyPnJC5K1tmhskFp7Sxi1Em4wxm3WrQ9aRj2oGLSaKimQ8ECyD8nKSDRynotXC1mFJB2MaMujDUKkEQ9SiVVg2sDvVy6EHGXvQbJtxFu9LiNYa9s23ro4qMAmNL7GaRVew7ss6Rh2f5D6mhBpEG58DewYwNfQnS43xmhDPQ4x55GGDCbxBApRcQd5pqCxF6HmLvipCy3DKm2F59y9QyKNVr1eJzgyRiXUNZFnaSvXdMKcAwBakNjoDSmFzLTQ711WFN731eCr8KtzVR3hpu9U3mptYwRRZXD8oL7FwZt4LaytcvB6x2Rz7ZkQcRFh6wMcEzVugwZvLPAoFyA4ZhcTsFCW8RKgJ1aEUHg79Rjti8p8QffqYaMapr2zKiqqcd4CxA8ssEy1DCNtXDmyMVBEeyQfr9ELeMk8XMTFYGKYMfF6tSDLQWn4gq3beVXC7vD3H2qGk2WKM8sY6jZC4rmJL1arRrN8NARMLwnkqTMWrtWb5Vmj7YDVR92nxdKFGK7Fabhyz6cQP3SxEAHbMfo4B6C1hTv1XebCp4pcBBLsurndpoDgSkTUFdHBFygg5t3jNByF7eJHoRoDQd892wg53zjY13rQ8QbrMJav"
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)
	if err != nil {
		t.Fatal("Error parse tx", err)
	}
	var tx transaction.Tx
	err = json.Unmarshal(rawTxBytes, &tx)
	msg := &wire.MessageTx{ Transaction: &tx, }
	if !netSync.HandleTxWithRole(msg.Transaction) {
		t.Fatal("NetSync should acception this transaction")
	}
}
func TestNetSyncHandleCacheTx(t *testing.T) {
	netSync := NetSync{}.New(&NetSyncConfig{
		PubSubManager: pubsub.NewPubSubManager(),
	})
	hash := common.HashH([]byte{0})
	res := netSync.HandleCacheTx(hash)
	if res {
		t.Fatal("Hash should not be in cache")
	}
	res = netSync.HandleCacheTx(hash)
	if !res {
		t.Fatal("Hash should be in cache")
	}
}
func TestNetSyncHandleCacheTxHash(t *testing.T) {
	netSync := NetSync{}.New(&NetSyncConfig{
		PubSubManager: pubsub.NewPubSubManager(),
	})
	hash := common.HashH([]byte{0})
	netSync.HandleCacheTxHash(hash)
	res := netSync.HandleCacheTx(hash)
	if !res {
		t.Fatal("Hash should be in cache")
	}
}

func TestNetSyncHandleMessageTx(t *testing.T) {
	//mempool.Logger.Init(mempoolLogger)
	//Logger.Init(netsyncLogger)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		} else {
			t.Fatal("Blockchain not called")
		}
	}()
	pubSub := pubsub.NewPubSubManager()
	txPool := mempool.TxPool{}
	txPool.Init(&mempool.Config{
		PubSubManager: pubSub,
	})
	netSync := NetSync{}.New(&NetSyncConfig{
		RoleInCommittees: 0,
		PubSubManager: pubSub,
		TxMemPool: &txPool,
	})
	base58CheckData := "13AHwQTSToFvjZFGjxKZbVKL6Mq2sLaGyevhwFUGMvdNd7745xSXhzJ2q2ra6d8pmb8CgieHM4zgbMJSVpsxJFhprUzUu9Ba7FWYcdmHCkqviF1LhysmBugQP7mEGQVo6PJrFp7Xim72ndfwPovTLhWrqYa8475MAM8XMKMEtYRvNXjPghMJuYvN9yxPPWzXUFQGnw3temoR3heaXMCwH4ZGMVWVaom2prZZwQqdKQkC3Yi2E7hjntDny2oCVsvyHpCv7nAZganw9ZqXc2H4V15GNnvTKWUBRxZo8tRK79tnyoRH23SNtPrc7hG3tJc8gtLoZ2FLQP2UrGPi88XpnzNagTix3ydZ3HUMu29bBRJALKdREn7jsMPECN8EjgZayx2XuZNZGJ1Tqzdq7dosUB1TfKfdXk9rVQXuLqxSM7sszST2kdYQ4sLB5YnpcePtFpX3SsR2fbBB7NWNzdq8JroWiLcAcxmH5jVMtkXTkrdLYtEdHDDSm1RQgPEzTkKT3479Hoge5ocDeVBdEqp8v21vbTLvBYWW7cjy3UH9KucqCRWWN7Tiztbnujs3nkazAAHBLFgSSbmVHD9MGKScqnB249ECGqeoHwCnuTTC2GPehyFb3N67mR9bHNFfJz1MG7qwAt3CDVACR68FqCrBpEkkk1ko2jhedYe2ZAmDAWgL8HWvXK112Ji3ucwyaNm3CvjGDWpyyYAs7TjkF58RtMGojwGCutHNoNHwkEtdGoekRoHfVrTQgmgfsxEsAPbL4JcCNKzjMReDAwGVNXu4kWF6KwFtyD5z6jQyYZZX9aEm1fcMmTvFcmX1DBUjsW4BnELPZR6iYeJkikL7bYZZPLmfzKenD17jFLe4F5xhwFcBPN2kxjTFsLKG6VT2EKeShJGzuWgNAbBVBjgjECk6jrZC44HaXgUTdRcgr3MEfp1JAgegCaKJTaXn3eacHqPQrDovMoJ1wijwok4DNFWhA2Ano75udeZHZjHEvytsdEtL1VV74Rs6dFjz591EHvGqvCVFx9qgTRHmSxRDkGox2rYD5LFXSDsL9AZ2aLaoisysT21Dc81WNuG36QhgRjpZ9LZeaCVL8ffjDV5Wbe8uqedN8uZy9P6ZXothXgCjBiAxsDgyHLAbMqND742Sq3wh2N1mTgTRg75S8ijN1smca3U7Snj2iELXRhLrwr6zR8aUxWe4Tdy19knpbEwBoUQcd5i9pwii3q67nYSKLYQZyPnJC5K1tmhskFp7Sxi1Em4wxm3WrQ9aRj2oGLSaKimQ8ECyD8nKSDRynotXC1mFJB2MaMujDUKkEQ9SiVVg2sDvVy6EHGXvQbJtxFu9LiNYa9s23ro4qMAmNL7GaRVew7ss6Rh2f5D6mhBpEG58DewYwNfQnS43xmhDPQ4x55GGDCbxBApRcQd5pqCxF6HmLvipCy3DKm2F59y9QyKNVr1eJzgyRiXUNZFnaSvXdMKcAwBakNjoDSmFzLTQ711WFN731eCr8KtzVR3hpu9U3mptYwRRZXD8oL7FwZt4LaytcvB6x2Rz7ZkQcRFh6wMcEzVugwZvLPAoFyA4ZhcTsFCW8RKgJ1aEUHg79Rjti8p8QffqYaMapr2zKiqqcd4CxA8ssEy1DCNtXDmyMVBEeyQfr9ELeMk8XMTFYGKYMfF6tSDLQWn4gq3beVXC7vD3H2qGk2WKM8sY6jZC4rmJL1arRrN8NARMLwnkqTMWrtWb5Vmj7YDVR92nxdKFGK7Fabhyz6cQP3SxEAHbMfo4B6C1hTv1XebCp4pcBBLsurndpoDgSkTUFdHBFygg5t3jNByF7eJHoRoDQd892wg53zjY13rQ8QbrMJav"
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)
	if err != nil {
		t.Fatal("Error parse tx", err)
	}
	var tx transaction.Tx
	err = json.Unmarshal(rawTxBytes, &tx)
	msg := &wire.MessageTx{ Transaction: &tx, }
	netSync.HandleMessageTx(msg)
	res := netSync.HandleCacheTx(*msg.Transaction.Hash())
	if !res {
		t.Fatal("Transaction should be in cache")
	}
}

func TestNetSyncHandleMessageTxToken(t *testing.T) {
	//mempool.Logger.Init(mempoolLogger)
	//Logger.Init(netsyncLogger)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		} else {
			t.Fatal("Blockchain not called")
		}
	}()
	pubSub := pubsub.NewPubSubManager()
	txPool := mempool.TxPool{}
	txPool.Init(&mempool.Config{
		PubSubManager: pubSub,
	})
	netSync := NetSync{}.New(&NetSyncConfig{
		RoleInCommittees: 1,
		PubSubManager: pubSub,
		TxMemPool: &txPool,
	})
	base58CheckData := "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"
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)
	if err != nil {
		t.Fatal("Error parse tx", err)
	}
	tx := transaction.TxCustomToken{}
	err = json.Unmarshal(rawTxBytes, &tx)
	msg := &wire.MessageTxToken{ Transaction: &tx, }
	netSync.HandleMessageTxToken(msg)
	res := netSync.HandleCacheTx(*msg.Transaction.Hash())
	if !res {
		t.Fatal("Transaction should be in cache")
	}
}

func TestNetSyncHandleMessageTxPrivacyToken(t *testing.T) {
	//mempool.Logger.Init(mempoolLogger)
	//Logger.Init(netsyncLogger)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		} else {
			t.Fatal("Blockchain not called")
		}
	}()
	pubSub := pubsub.NewPubSubManager()
	txPool := mempool.TxPool{}
	txPool.Init(&mempool.Config{
		PubSubManager: pubSub,
	})
	netSync := NetSync{}.New(&NetSyncConfig{
		RoleInCommittees: 1,
		PubSubManager: pubSub,
		TxMemPool: &txPool,
	})
	base58CheckData := "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"
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)
	if err != nil {
		t.Fatal("Error parse tx", err)
	}
	tx := transaction.TxCustomTokenPrivacy{}
	err = json.Unmarshal(rawTxBytes, &tx)
	msg := &wire.MessageTxPrivacyToken{ Transaction: &tx, }
	netSync.HandleMessageTxPrivacyToken(msg)
	res := netSync.HandleCacheTx(*msg.Transaction.Hash())
	if !res {
		t.Fatal("Transaction should be in cache")
	}
}
/*
	case *wire.MessageBFTPropose:
		{
			netSync.HandleMessageBFTMsg(msg)
		}
	case *wire.MessageBFTAgree:
		{
			netSync.HandleMessageBFTMsg(msg)
		}
	case *wire.MessageBFTCommit:
		{
			netSync.HandleMessageBFTMsg(msg)
		}
	case *wire.MessageBFTReady:
		{
			netSync.HandleMessageBFTMsg(msg)
		}
	case *wire.MessageBFTReq:
		{
			netSync.HandleMessageBFTMsg(msg)
		}
*/
func TestHandleCacheBlock(t *testing.T) {
	netSync := NetSync{}.New(&NetSyncConfig{
		PubSubManager: pubsub.NewPubSubManager(),
	})
	hash := common.HashH([]byte{0})
	res := netSync.HandleCacheBlock(hash.String())
	if res {
		t.Fatal("Hash should not be in cache")
	}
	res = netSync.HandleCacheBlock(hash.String())
	if !res {
		t.Fatal("Hash should be in cache")
	}
}
func TestNetSyncHandleMessageBeaconBlock(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		} else {
			t.Fatal("Blockchain not called")
		}
	}()
	block := blockchain.BeaconBlock{}
	block.Header.Height = 2
	netSync := NetSync{}.New(&NetSyncConfig{
		PubSubManager: pubsub.NewPubSubManager(),
	})
	netSync.HandleMessageBeaconBlock(&wire.MessageBeaconBlock{Block: block})
}
func TestNetSyncHandleMessageShardBlock(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		} else {
			t.Fatal("Blockchain not called")
		}
	}()
	block := blockchain.ShardBlock{}
	block.Header.Height = 2
	netSync := NetSync{}.New(&NetSyncConfig{
		PubSubManager: pubsub.NewPubSubManager(),
	})
	netSync.HandleMessageShardBlock(&wire.MessageShardBlock{Block: block})
}
func TestNetSyncHandleMessageShardToBeacon(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		} else {
			t.Fatal("Blockchain not called")
		}
	}()
	block := blockchain.ShardToBeaconBlock{}
	block.Header.Height = 2
	netSync := NetSync{}.New(&NetSyncConfig{
		PubSubManager: pubsub.NewPubSubManager(),
	})
	netSync.HandleMessageShardToBeacon(&wire.MessageShardToBeacon{Block: block})
}
func TestNetSyncHandleMessageCrossShard(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		} else {
			t.Fatal("Blockchain not called")
		}
	}()
	block := blockchain.CrossShardBlock{}
	block.Header.Height = 2
	netSync := NetSync{}.New(&NetSyncConfig{
		PubSubManager: pubsub.NewPubSubManager(),
	})
	netSync.HandleMessageCrossShard(&wire.MessageCrossShard{Block: block})
}
