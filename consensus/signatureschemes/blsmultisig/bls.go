package blsmultisig

import (
	"errors"
	"math/big"
	"reflect"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

// Sign return BLS signature
func Sign(data, skBytes []byte, selfIdx int, committee []PublicKey) ([]byte, error) {
	if len(skBytes) != CSKSz {
		return []byte{0}, NewBLSSignatureError(InvalidPrivateKeyErr, errors.New(ErrCodeMessage[InvalidPrivateKeyErr].Message))
	}
	sk := B2I(skBytes)
	if selfIdx >= len(committee) || (selfIdx < 0) || (len(committee) < 1) {
		return []byte{0}, NewBLSSignatureError(InvalidCommitteeInfoErr, errors.New(ErrCodeMessage[InvalidCommitteeInfoErr].Message))
	}
	dataPn := B2G1P(data)
	aiSk := AiGen(committee, selfIdx)
	aiSk.Mul(aiSk, sk)
	aiSk.Mod(aiSk, bn256.Order)
	sig := dataPn.ScalarMult(dataPn, aiSk)
	return CmprG1(sig), nil
}

// Verify verify BLS sig on given data and list public key
func Verify(sig, data []byte, signersIdx []int, committee []PublicKey) (bool, error) {
	// if len(skBytes) != CSKSz {
	// 	return []byte{0}, NewBLSSignatureError(InvalidPrivateKeyErr, errors.New(ErrCodeMessage[InvalidPrivateKeyErr].Message))
	// }
	// sk := B2I(skBytes)
	for _, idx := range signersIdx {
		if (idx < 0) || (idx >= len(committee)) {
			return false, NewBLSSignatureError(InvalidCommitteeInfoErr, errors.New(ErrCodeMessage[InvalidCommitteeInfoErr].Message))
		}
	}
	if len(signersIdx) > len(committee) || (len(committee) < 1) {
		return false, NewBLSSignatureError(InvalidCommitteeInfoErr, errors.New(ErrCodeMessage[InvalidCommitteeInfoErr].Message))
	}
	gG2Pn := new(bn256.G2)
	gG2Pn.ScalarBaseMult(big.NewInt(1))
	sigPn, err := DecmprG1(sig)
	if err != nil {
		return false, NewBLSSignatureError(DecompressFromByteErr, err)
	}
	lPair := bn256.Pair(sigPn, gG2Pn)
	apk := APKGen(committee, signersIdx)
	dataPn := B2G1P(data)
	rPair := bn256.Pair(dataPn, apk)
	if !reflect.DeepEqual(lPair.Marshal(), rPair.Marshal()) {
		return false, nil
	}
	return true, nil
}

// Combine combine list of bls signature
func Combine(sigs [][]byte) ([]byte, error) {
	cSigPn, err := DecmprG1(sigs[0])
	if err != nil {
		return []byte{0}, NewBLSSignatureError(DecompressFromByteErr, err)
	}
	for i := 1; i < len(sigs); i++ {
		tmp, err := DecmprG1(sigs[i])
		if err != nil {
			return []byte{0}, NewBLSSignatureError(DecompressFromByteErr, err)
		}
		cSigPn.Add(cSigPn, tmp)
	}
	return CmprG1(cSigPn), nil
}
