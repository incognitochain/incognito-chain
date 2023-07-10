package common

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
)

type MetadataBase struct {
	Type int
}

func (mb *MetadataBase) Sign(privateKey *privacy.PrivateKey, tx Transaction) error {
	return nil
}
func (mb *MetadataBase) VerifyMetadataSignature(publicKey []byte, tx Transaction) (bool, error) {
	return true, nil
}

func (mb *MetadataBase) SetSharedRandom(_ []byte) {
	return
}

type MetadataBaseWithSignature struct {
	MetadataBase
	Sig []byte `json:"Sig,omitempty"`
}

func NewMetadataBaseWithSignature(thisType int) *MetadataBaseWithSignature {
	return &MetadataBaseWithSignature{MetadataBase: MetadataBase{Type: thisType}, Sig: []byte{}}
}

func (mbs *MetadataBaseWithSignature) Sign(privateKey *privacy.PrivateKey, tx Transaction) error {
	hashForMd := tx.HashWithoutMetadataSig()
	if hashForMd == nil {
		// the metadata type does not need signing
		return nil
	}
	if len(mbs.Sig) > 0 {
		return errors.New("Cannot overwrite metadata signature")
	}

	/****** using Schnorr signature *******/
	sk := new(privacy.Scalar).FromBytesS(*privateKey)
	r := new(privacy.Scalar).FromUint64(0)
	sigKey := new(privacy.SchnorrPrivateKey)
	sigKey.Set(sk, r)

	// signing
	signature, err := sigKey.Sign(hashForMd[:])
	if err != nil {
		return err
	}

	// convert signature to byte array
	mbs.Sig = signature.Bytes()
	return nil
}

func (mbs *MetadataBaseWithSignature) VerifyMetadataSignature(publicKey []byte, tx Transaction) (bool, error) {
	hashForMd := tx.HashWithoutMetadataSig()
	if len(mbs.Sig) == 0 {
		if hashForMd == nil {
			// the metadata type does not need signing
			sigPubKey := tx.GetSigPubKey()
			return bytes.Equal(sigPubKey, publicKey), nil
		} else {
			Logger.Log.Error("CheckAuthorizedSender: should have sig for metadata to verify")
			return false, errors.New("CheckAuthorizedSender should have sig for metadata to verify")
		}
	}
	verifyKey := new(privacy.SchnorrPublicKey)
	metaSigPublicKey, err := new(privacy.Point).FromBytesS(publicKey)
	if err != nil {
		Logger.Log.Error(err)
		return false, err
	}
	verifyKey.Set(metaSigPublicKey)
	signature := new(privacy.SchnSignature)
	if err := signature.SetBytes(mbs.Sig); err != nil {
		Logger.Log.Errorf("Invalid signature %v", err)
		return false, err
	}
	return verifyKey.Verify(signature, hashForMd[:]), nil
}

func NewMetadataBase(thisType int) *MetadataBase {
	return &MetadataBase{Type: thisType}
}

func (mb MetadataBase) IsMinerCreatedMetaType() bool {
	metaType := mb.GetType()
	for _, mType := range minerCreatedMetaTypes {
		if metaType == mType {
			return true
		}
	}
	return false
}

func (mb *MetadataBase) CalculateSize() uint64 {
	return 0
}

func (mb *MetadataBase) Validate() error {
	return nil
}

func (mb *MetadataBase) Process() error {
	return nil
}

func (mb MetadataBase) GetType() int {
	return mb.Type
}

func (mb MetadataBase) Hash() *common.Hash {
	record := strconv.Itoa(mb.Type)
	data := []byte(record)
	hash := common.HashH(data)
	return &hash
}

func (mb MetadataBase) HashWithoutSig() *common.Hash {
	return mb.Hash()
}

func (mb MetadataBase) CheckTransactionFee(tx Transaction, minFeePerKbTx uint64, minFeePerTx uint64, beaconHeight int64, stateDB *statedb.StateDB) bool {

	// stateDB.
	// mempool.NewFeeEstimator()

	// limitFee, minFeePerTx := blockchain.GetFeeInfo(tx.GetMetadata(), tp.config.FeeEstimator[shardID])

	// minFeePerTx := config.Config().MinFeePerTx
	// specifiedFeeTx := config.Config().SpecifiedFeePerTx

	// // set min fee for specified tx metadata types
	// if tx.GetMetadata() != nil && IsSpecifiedFeeMetaType(tx.GetMetadataType()) && minFeePerTx < specifiedFeeTx {
	// 	minFeePerTx = specifiedFeeTx
	// }

	if tx.GetType() == common.TxCustomTokenPrivacyType || tx.GetType() == common.TxTokenConversionType {
		feeNativeToken := tx.GetTxFee()
		feePToken := tx.GetTxFeeToken()
		if feePToken > 0 {
			tokenID := tx.GetTokenID()
			feePTokenToNativeTokenTmp, err := ConvertPrivacyTokenToNativeToken(feePToken, tx.GetTokenID(), beaconHeight, stateDB)
			if err != nil {
				fmt.Printf("transaction %+v: %+v %v can not convert to native token",
					tx.Hash().String(), feePToken, tokenID)
				return false
			}
			feePTokenToNativeToken := uint64(math.Ceil(feePTokenToNativeTokenTmp))
			feeNativeToken += feePTokenToNativeToken
		}

		// check minFeePerTx
		if feeNativeToken < minFeePerTx {
			fmt.Printf("transaction %+v has %d fees PRV which is under the required min fee per tx %d",
				tx.Hash().String(), feeNativeToken, minFeePerTx)
			return false
		}
		// get limit fee in native token
		actualTxSize := tx.GetTxActualSize()
		// check fee in native token
		minFee := actualTxSize * minFeePerKbTx
		if feeNativeToken < minFee {
			fmt.Printf("transaction %+v has %d fees PRV which is under the required amount of %d, tx size %d",
				tx.Hash().String(), feeNativeToken, minFee, actualTxSize)
			return false
		}
		return true
	}
	// normal privacy tx
	txFee := tx.GetTxFee()

	// check minFeePerTx
	if txFee < minFeePerTx {
		fmt.Printf("transaction %+v has %d fees PRV which is under the required min fee per tx %d",
			tx.Hash().String(), txFee, minFeePerTx)
		return false
	}

	fullFee := minFeePerKbTx * tx.GetTxActualSize()
	return !(txFee < fullFee)
}

func (mb *MetadataBase) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	return [][]string{}, nil
}

func (mb MetadataBase) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *MintData, shardID byte, tx Transaction, chainRetriever ChainRetriever, ac *AccumulatedValues, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever) (bool, error) {
	return true, nil
}
