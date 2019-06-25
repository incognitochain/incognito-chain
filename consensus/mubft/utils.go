package mubft

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy"
)

func GetPubKeysFromIdx(pubkeyList []string, idxs []int) []*privacy.PublicKey {
	listPubkeyOfSigners := make([]*privacy.PublicKey, len(idxs))
	for i := 0; i < len(idxs); i++ {
		listPubkeyOfSigners[i] = new(privacy.PublicKey)
		pubKeyTemp, byteVersion, err := base58.Base58Check{}.Decode(pubkeyList[idxs[i]])
		if (err != nil) || (byteVersion != common.ZeroByte) {
			Logger.log.Info(err)
			continue
		}
		*listPubkeyOfSigners[i] = pubKeyTemp
	}
	return listPubkeyOfSigners
}

func GetClosestPoolState(poolStates []map[byte]uint64) map[byte]uint64 {
	closestPoolState := make(map[byte]uint64)

	for _, poolState := range poolStates {
		for shardID, blkHeight := range poolState {
			if closestBlkHeight, ok := closestPoolState[shardID]; !ok {
				closestPoolState[shardID] = blkHeight
			} else {
				if closestBlkHeight < blkHeight {
					closestPoolState[shardID] = blkHeight
				}
			}
		}
	}

	return closestPoolState
}
