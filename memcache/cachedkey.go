package memcache

import "github.com/incognitochain/incognito-chain/common"

func GetListOutputcoinCachedKey(pubKey []byte, tokenID *common.Hash, shardID byte) []byte {
	key := make([]byte, 0)
	key = append(key, []byte("-")...)
	key = append(key, []byte("listoutputcoin")...)
	key = append(key, []byte("-")...)
	key = append(key, pubKey...)
	key = append(key, []byte("-")...)
	key = append(key, tokenID.GetBytes()...)
	key = append(key, shardID)
	return key
}
