package mlsag

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
)


func (this *Mlsag) SignConfidentialAsset(message []byte) (*MlsagSig, error) {
	if len(message) != common.HashSize {
		return nil, errors.New("Cannot mlsag sign the message because its length is not 32, maybe it has not been hashed")
	}
	message32byte := [32]byte{}
	copy(message32byte[:], message)

	alpha, r := this.createRandomChallenges()          // step 2 in paper
	c, err := this.calculateC(message32byte, alpha, r) // step 3 and 4 in paper

	if err != nil {
		return nil, err
	}
	return &MlsagSig{
		c[0], this.keyImages, r,
	}, nil
}