package constantpos

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	privacy "github.com/ninjadotorg/constant/privacy"
	"github.com/pkg/errors"
)

type bftCommittedSig struct {
	Pubkey         string
	ValidatorsIdxR []int
	Sig            string
}

type multiSigScheme struct {
	userKeySet *cashec.KeySet
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
	cryptoScheme *privacy.MultiSigScheme
}

func (self *multiSigScheme) Init(userKeySet *cashec.KeySet, committee []string) {
	self.combine.SigningCommittee = make([]string, len(committee))
	copy(self.combine.SigningCommittee, committee)
	self.cryptoScheme = new(privacy.MultiSigScheme)
	self.cryptoScheme.Init()
	self.cryptoScheme.Keyset.Set(&userKeySet.PrivateKey, &userKeySet.PaymentAddress.Pk)
}

func (self *multiSigScheme) Prepare() error {
	myRiECCPoint, myrBigInt := self.cryptoScheme.GenerateRandom()
	myRi := myRiECCPoint.Compress()
	myr := myrBigInt.Bytes()
	for len(myr) < privacy.BigIntSize {
		myr = append([]byte{0}, myr...)
	}

	self.personal.Ri = myRi
	self.personal.r = myr
	return nil
}

func (self *multiSigScheme) SignData(RiList map[string][]byte) error {
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
		if (err != nil) || (byteVersion != byte(0x00)) {
			//Todo
			return err
		}
		listROfSigners[counter] = new(privacy.EllipticPoint)
		err = listROfSigners[counter].Decompress(bytesR)
		if err != nil {
			//Todo
			return err
		}
		RCombined = RCombined.Add(listROfSigners[counter])
		// phaseData.ValidatorsIdx[counter] = sort.SearchStrings(self.Committee, szPubKey)
		self.combine.ValidatorsIdxR = append(self.combine.ValidatorsIdxR, common.IndexOfStr(szPubKey, self.combine.SigningCommittee))
		counter++
	}
	sort.Ints(self.combine.ValidatorsIdxAggSig)
	//Todo Sig block with R Here

	commitSig := self.cryptoScheme.Keyset.SignMultiSig(self.dataToSig.GetBytes(), listPubkeyOfSigners, listROfSigners, new(big.Int).SetBytes(self.personal.r))

	self.combine.R = base58.Base58Check{}.Encode(RCombined.Compress(), byte(0x00))
	self.combine.CommitSig = base58.Base58Check{}.Encode(commitSig.Bytes(), byte(0x00))

	return nil
}

func (self *multiSigScheme) VerifyCommitSig(validatorPk string, commitSig string, R string, validatorsIdx []int) error {
	RCombined := new(privacy.EllipticPoint)
	RCombined.Set(big.NewInt(0), big.NewInt(0))
	Rbytesarr, byteVersion, err := base58.Base58Check{}.Decode(R)
	if (err != nil) || (byteVersion != byte(0x00)) {
		return err
	}
	err = RCombined.Decompress(Rbytesarr)
	if err != nil {
		return err
	}
	listPubkeyOfSigners := GetPubKeysFromIdx(self.combine.SigningCommittee, validatorsIdx)
	validatorPubkey := new(privacy.PublicKey)
	pubKeyTemp, byteVersion, err := base58.Base58Check{}.Decode(validatorPk)
	if (err != nil) || (byteVersion != byte(0x00)) {
		return err
	}
	*validatorPubkey = pubKeyTemp
	var valSigbytesarr []byte
	valSigbytesarr, byteVersion, err = base58.Base58Check{}.Decode(commitSig)
	valSig := new(privacy.SchnMultiSig)
	valSig.SetBytes(valSigbytesarr)

	resValidateEachSigOfSigners := valSig.VerifyMultiSig(self.dataToSig.GetBytes(), listPubkeyOfSigners, validatorPubkey, RCombined)
	if !resValidateEachSigOfSigners {
		return errors.New("Validator's sig is invalid " + validatorPk)
	}
	return nil
}

func (self *multiSigScheme) CombineSigs(commitSigs []bftCommittedSig) (string, error) {

	//TODO: Hy include valSig.ValidatorsIdxR in aggregatedSig

	listSigOfSigners := make([]*privacy.SchnMultiSig, len(commitSigs))
	for i, valSig := range commitSigs {
		listSigOfSigners[i] = new(privacy.SchnMultiSig)
		bytesSig, byteVersion, err := base58.Base58Check{}.Decode(valSig.Sig)
		if (err != nil) || (byteVersion != byte(0x00)) {
			return "", err
		}
		listSigOfSigners[i].SetBytes(bytesSig)
		self.combine.ValidatorsIdxAggSig = append(self.combine.ValidatorsIdxAggSig, common.IndexOfStr(valSig.Pubkey, self.combine.SigningCommittee))
	}
	sort.Ints(self.combine.ValidatorsIdxAggSig)
	aggregatedSig := self.cryptoScheme.CombineMultiSig(listSigOfSigners)
	fmt.Println("aaaaaaaaaaaaaaaaaaaa", len(commitSigs))
	// fmt.Println("bbbbbbbbbbbbbbbbbbbb", commitSigs[0], commitSigs[1], commitSigs[2])
	return base58.Base58Check{}.Encode(aggregatedSig.Bytes(), byte(0x00)), nil
}
