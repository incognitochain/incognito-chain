package blockchain

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"log"
)

var CommitteeFromBlockBootStrapCache = map[byte]*lru.Cache{}

func (s *BlockChain) RestoreCommitteeFromBlockCache() {
	allCacheCommitteeFromBlock, err := rawdbv2.GetAllCacheCommitteeFromBlock(s.BeaconChain.BlockStorage.blockStorageDB)
	if err != nil {
		panic(err)
	}
	for sid, data := range allCacheCommitteeFromBlock {
		if cache, ok := CommitteeFromBlockBootStrapCache[byte(sid)]; !ok || cache == nil {
			CommitteeFromBlockBootStrapCache[byte(sid)], _ = lru.New(10)
		}
		for hash, cpks := range data {
			log.Println("restore", sid, hash.String(), len(cpks))
			CommitteeFromBlockBootStrapCache[byte(sid)].Add(hash.String(), cpks)
		}
	}
}
