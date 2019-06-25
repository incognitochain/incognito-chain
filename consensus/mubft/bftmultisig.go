package mubft

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/incognitochain/incognito-chain/cashec"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

type bftCommittedSig struct {
	ValidatorsIdxR []int
	Sig            string
}

type multiSigScheme struct {
	//user data use for sign
	dataToSig common.Hash
	personal  struct {
		Ri []byte
		r  []byte
	}
	//user data user for combine sig
	combine struct {
		CommitSig           string
		R                   string
		ValidatorsIdxR      []int
		ValidatorsIdxAggSig []int
		SigningCommittee    []string
	}
	scheme *privacy.MultiSigScheme
}

func (multiSig *multiSigScheme) Init(userKeySet *cashec.KeySet, committee []string) {
	multiSig.combine.SigningCommittee = make([]string, len(committee))
	copy(multiSig.combine.SigningCommittee, committee)
	multiSig.scheme = new(privacy.MultiSigScheme)
	multiSig.scheme.Init()
	multiSig.scheme.Keyset.Set(&userKeySet.PrivateKey, &userKeySet.PaymentAddress.Pk)
}

func (multiSig *multiSigScheme) Prepare() error {
	myRiECCPoint, myrBigInt := multiSig.scheme.GenerateRandom()
	myRi := myRiECCPoint.Compress()
	myr := myrBigInt.Bytes()
	for len(myr) < privacy.BigIntSize {
		myr = append([]byte{0}, myr...)
	}

	multiSig.personal.Ri = myRi
	multiSig.personal.r = myr
	return nil
}

func (multiSig *multiSigScheme) SignData(RiList map[string][]byte) error {
	numbOfSigners := len(RiList)
	listPubkeyOfSigners := make([]*privacy.PublicKey, numbOfSigners)
	listROfSigners := make([]*privacy.EllipticPoint, numbOfSigners)
	RCombined := new(privacy.EllipticPoint)
	RCombined.Set(big.NewInt(0), big.NewInt(0))
	counter := 0

	for szPubKey, bytesR := range RiList {
		pubKeyTemp, byteVersion, err := base58.Base58Check{}.Decode(szPubKey)
		listPubkeyOfSigners[counter] = new(privacy.PublicKey)
		*listPubkeyOfSigners[counter] = pubKeyTemp
		if (err != nil) || (byteVersion != common.ZeroByte) {
			return err
		}
		listROfSigners[counter] = new(privacy.EllipticPoint)
		err = listROfSigners[counter].Decompress(bytesR)
		if err != nil {
			return err
		}
		RCombined = RCombined.Add(listROfSigners[counter])
		multiSig.combine.ValidatorsIdxR = append(multiSig.combine.ValidatorsIdxR, common.IndexOfStr(szPubKey, multiSig.combine.SigningCommittee))
		counter++
	}
	sort.Ints(multiSig.combine.ValidatorsIdxR)

	commitSig := multiSig.scheme.Keyset.SignMultiSig(multiSig.dataToSig.GetBytes(), listPubkeyOfSigners, listROfSigners, new(big.Int).SetBytes(multiSig.personal.r))

	multiSig.combine.R = base58.Base58Check{}.Encode(RCombined.Compress(), common.ZeroByte)
	multiSig.combine.CommitSig = base58.Base58Check{}.Encode(commitSig.Bytes(), common.ZeroByte)

	return nil
}

func (multiSig *multiSigScheme) VerifyCommitSig(validatorPk string, commitSig string, R string, validatorsIdx []int) error {
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
	listPubkeyOfSigners := GetPubKeysFromIdx(multiSig.combine.SigningCommittee, validatorsIdx)
	validatorPubkey := new(privacy.PublicKey)
	pubKeyTemp, byteVersion, err := base58.Base58Check{}.Decode(validatorPk)
	if err != nil {
		return err
	}
	if byteVersion != common.ZeroByte {
		return fmt.Errorf("VerifyCommitSig byte version %+v", byteVersion)
	}
	*validatorPubkey = pubKeyTemp
	var valSigbytesarr []byte
	valSigbytesarr, byteVersion, err = base58.Base58Check{}.Decode(commitSig)
	if err != nil {
		return err
	}
	if byteVersion != common.ZeroByte {
		return fmt.Errorf("VerifyCommitSig byte version %+v", byteVersion)
	}
	valSig := new(privacy.SchnMultiSig)
	err = valSig.SetBytes(valSigbytesarr)
	if err != nil {
		return err
	}
	resValidateEachSigOfSigners := valSig.VerifyMultiSig(multiSig.dataToSig.GetBytes(), listPubkeyOfSigners, []*privacy.PublicKey{validatorPubkey}, RCombined)
	if !resValidateEachSigOfSigners {
		return errors.New("Validator's sig is invalid " + validatorPk)
	}
	return nil
}

func (multiSig *multiSigScheme) CombineSigs(R string, commitSigs map[string]bftCommittedSig) (string, error) {
	var listSigOfSigners []*privacy.SchnMultiSig
	var validatorsIdxR []int
	for pubkey, valSig := range commitSigs {
		sig := new(privacy.SchnMultiSig)
		bytesSig, byteVersion, err := base58.Base58Check{}.Decode(valSig.Sig)
		if (err != nil) || (byteVersion != common.ZeroByte) {
			return "", err
		}
		sig.SetBytes(bytesSig)
		if err != nil {
			return "", err
		}
		listSigOfSigners = append(listSigOfSigners, sig)
		multiSig.combine.ValidatorsIdxAggSig = append(multiSig.combine.ValidatorsIdxAggSig, common.IndexOfStr(pubkey, multiSig.combine.SigningCommittee))
		validatorsIdxR = valSig.ValidatorsIdxR
	}
	sort.Ints(multiSig.combine.ValidatorsIdxAggSig)
	multiSig.combine.R = R
	multiSig.combine.ValidatorsIdxR = make([]int, len(validatorsIdxR))
	copy(multiSig.combine.ValidatorsIdxR, validatorsIdxR)
	aggregatedSig := multiSig.scheme.CombineMultiSig(listSigOfSigners)
	return base58.Base58Check{}.Encode(aggregatedSig.Bytes(), common.ZeroByte), nil
}
