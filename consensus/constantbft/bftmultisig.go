package constantbft

import (
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

func (multiSig *multiSigScheme) Init(userKeySet *cashec.KeySet, committee []string) {
	multiSig.combine.SigningCommittee = make([]string, len(committee))
	copy(multiSig.combine.SigningCommittee, committee)
	multiSig.cryptoScheme = new(privacy.MultiSigScheme)
	multiSig.cryptoScheme.Init()
	multiSig.cryptoScheme.Keyset.Set(&userKeySet.PrivateKey, &userKeySet.PaymentAddress.Pk)
}

func (multiSig *multiSigScheme) Prepare() error {
	myRiECCPoint, myrBigInt := multiSig.cryptoScheme.GenerateRandom()
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
		if (err != nil) || (byteVersion != byte(0x00)) {
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

	commitSig := multiSig.cryptoScheme.Keyset.SignMultiSig(multiSig.dataToSig.GetBytes(), listPubkeyOfSigners, listROfSigners, new(big.Int).SetBytes(multiSig.personal.r))

	multiSig.combine.R = base58.Base58Check{}.Encode(RCombined.Compress(), byte(0x00))
	multiSig.combine.CommitSig = base58.Base58Check{}.Encode(commitSig.Bytes(), byte(0x00))

	return nil
}

func (multiSig *multiSigScheme) VerifyCommitSig(validatorPk string, commitSig string, R string, validatorsIdx []int) error {
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
	listPubkeyOfSigners := GetPubKeysFromIdx(multiSig.combine.SigningCommittee, validatorsIdx)
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

	resValidateEachSigOfSigners := valSig.VerifyMultiSig(multiSig.dataToSig.GetBytes(), listPubkeyOfSigners, []*privacy.PublicKey{validatorPubkey}, RCombined)
	if !resValidateEachSigOfSigners {
		return errors.New("Validator's sig is invalid " + validatorPk)
	}
	return nil
}

func (multiSig *multiSigScheme) CombineSigs(R string, commitSigs []bftCommittedSig) (string, error) {

	listSigOfSigners := make([]*privacy.SchnMultiSig, len(commitSigs))
	for i, valSig := range commitSigs {
		listSigOfSigners[i] = new(privacy.SchnMultiSig)
		bytesSig, byteVersion, err := base58.Base58Check{}.Decode(valSig.Sig)
		if (err != nil) || (byteVersion != byte(0x00)) {
			return "", err
		}
		listSigOfSigners[i].SetBytes(bytesSig)
		multiSig.combine.ValidatorsIdxAggSig = append(multiSig.combine.ValidatorsIdxAggSig, common.IndexOfStr(valSig.Pubkey, multiSig.combine.SigningCommittee))
	}
	sort.Ints(multiSig.combine.ValidatorsIdxAggSig)
	multiSig.combine.R = R
	multiSig.combine.ValidatorsIdxR = make([]int, len(commitSigs[0].ValidatorsIdxR))
	copy(multiSig.combine.ValidatorsIdxR, commitSigs[0].ValidatorsIdxR)
	aggregatedSig := multiSig.cryptoScheme.CombineMultiSig(listSigOfSigners)
	return base58.Base58Check{}.Encode(aggregatedSig.Bytes(), byte(0x00)), nil
}
