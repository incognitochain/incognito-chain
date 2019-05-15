package blockchain

import (
	"errors"
	"math/big"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/privacy"
)

func ValidateAggSignature(validatorIdx [][]int, committees []string, aggSig string, R string, blockHash *common.Hash) error {
	// return nil //single-node
	//multi-node
	pubKeysR := []*privacy.PublicKey{}
	for _, index := range validatorIdx[0] {
		pubkeyBytes, _, err := base58.Base58Check{}.Decode(committees[index])
		if err != nil {
			return errors.New("Error in convert Public key from string to byte")
		}
		pubKey := privacy.PublicKey{}
		pubKey = pubkeyBytes
		pubKeysR = append(pubKeysR, &pubKey)
	}
	pubKeysAggSig := []*privacy.PublicKey{}
	for _, index := range validatorIdx[1] {
		pubkeyBytes, _, err := base58.Base58Check{}.Decode(committees[index])
		if err != nil {
			return errors.New("Error in convert Public key from string to byte")
		}
		pubKey := privacy.PublicKey{}
		pubKey = pubkeyBytes
		pubKeysAggSig = append(pubKeysAggSig, &pubKey)
	}
	RCombined := new(privacy.EllipticPoint)
	RCombined.Set(big.NewInt(0), big.NewInt(0))
	Rbytesarr, byteVersion, err := base58.Base58Check{}.Decode(R)
	if (err != nil) || (byteVersion != common.ZeroByte) {
		return err
	}
	err = RCombined.Decompress(Rbytesarr)
	if err != nil {
		return err
	}

	tempAggSig, _, err := base58.Base58Check{}.Decode(aggSig)
	if err != nil {
		return errors.New("Error in convert aggregated signature from string to byte")
	}
	schnMultiSig := &privacy.SchnMultiSig{}
	schnMultiSig.SetBytes(tempAggSig)
	if schnMultiSig.VerifyMultiSig(blockHash.GetBytes(), pubKeysR, pubKeysAggSig, RCombined) == false {
		return errors.New("Invalid Agg signature")
	}
	return nil
}
