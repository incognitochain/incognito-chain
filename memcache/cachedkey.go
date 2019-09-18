package memcache

import "github.com/incognitochain/incognito-chain/common"

// GetListOutputcoinCachedKey - build key on memcache for list output coin of publickey
func GetListOutputcoinCachedKey(publicKey []byte, tokenID *common.Hash, shardID byte) []byte {
	key := make([]byte, 0)
	key = append(key, []byte(splitChar)...)
	key = append(key, []byte(outputCoinCacheKey)...)
	key = append(key, []byte(splitChar)...)
	key = append(key, publicKey...)
	key = append(key, []byte(splitChar)...)
	key = append(key, tokenID.GetBytes()...)
	key = append(key, shardID)
	return key
}

func GetShardBestStateCachedKey() []byte {
	key := make([]byte, 0)
	key = append(key, []byte(shardBestStateCacheKey)...)
	return key
}
