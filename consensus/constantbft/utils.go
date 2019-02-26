package constantbft

import (
	"github.com/ninjadotorg/constant/common/base58"
	privacy "github.com/ninjadotorg/constant/privacy"
)

func GetPubKeysFromIdx(pubkeyList []string, idxs []int) []*privacy.PublicKey {
	listPubkeyOfSigners := make([]*privacy.PublicKey, len(idxs))
	for i := 0; i < len(idxs); i++ {
		listPubkeyOfSigners[i] = new(privacy.PublicKey)
		pubKeyTemp, byteVersion, err := base58.Base58Check{}.Decode(pubkeyList[idxs[i]])
		if (err != nil) || (byteVersion != byte(0x00)) {
			Logger.log.Info(err)
			continue
		}
		*listPubkeyOfSigners[i] = pubKeyTemp
	}
	return listPubkeyOfSigners
}
