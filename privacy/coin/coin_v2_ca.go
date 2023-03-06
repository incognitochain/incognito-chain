package coin

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

const MaxAttempts int = 50000

func (coin *CoinV2) ComputeCommitmentCA() (*operation.Point, error) {
	if coin == nil || coin.GetRandomness() == nil || coin.GetAmount() == nil {
		return nil, fmt.Errorf("missing arguments for committing")
	}
	// must not change gRan
	gRan := operation.PedCom.G[operation.PedersenRandomnessIndex]
	commitment := new(operation.Point).ScalarMult(coin.GetAssetTag(), coin.GetAmount())
	commitment.Add(commitment, new(operation.Point).ScalarMult(gRan, coin.GetRandomness()))
	return commitment, nil
}

func ComputeCommitmentCA(assetTag *operation.Point, r, v *operation.Scalar) (*operation.Point, error) {
	if assetTag == nil || r == nil || v == nil {
		return nil, fmt.Errorf("missing arguments for committing to CA coin")
	}
	// must not change gRan
	gRan := operation.PedCom.G[operation.PedersenRandomnessIndex]
	commitment := new(operation.Point).ScalarMult(assetTag, v)
	commitment.Add(commitment, new(operation.Point).ScalarMult(gRan, r))
	return commitment, nil
}

func ComputeAssetTagBlinder(sharedSecret *operation.Point) (*operation.Scalar, error) {
	if sharedSecret == nil {
		return nil, fmt.Errorf("missing arguments for asset tag blinder")
	}
	blinder := operation.HashToScalar(append(sharedSecret.ToBytesS(), []byte("assettag")...))
	return blinder, nil
}

// this should be an input coin
func (coin *CoinV2) RecomputeSharedSecret(privateKey []byte) (*operation.Point, error) {
	// sk := new(operation.Scalar).FromBytesS(privateKey)
	var privOTA []byte = key.GeneratePrivateOTAKey(privateKey)[:]
	sk := new(operation.Scalar).FromBytesS(privOTA)
	// this is g^SharedRandom, previously created by sender of the coin
	sharedOTARandomPoint, err := coin.GetTxRandom().GetTxOTARandomPoint()
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve tx random detail")
	}
	sharedSecret := new(operation.Point).ScalarMult(sharedOTARandomPoint, sk)
	return sharedSecret, nil
}

func (coin *CoinV2) ValidateAssetTag(sharedSecret *operation.Point, tokenID *common.Hash) (bool, error) {
	if coin.GetAssetTag() == nil {
		if tokenID == nil || *tokenID == common.PRVCoinID {
			// a valid PRV coin
			return true, nil
		}
		return false, fmt.Errorf("CA coin must have asset tag")
	}
	if tokenID == nil || *tokenID == common.PRVCoinID {
		// invalid
		return false, fmt.Errorf("PRV coin cannot have asset tag")
	}
	recomputedAssetTag := operation.HashToPoint(tokenID[:])
	if operation.IsPointEqual(recomputedAssetTag, coin.GetAssetTag()) {
		return true, nil
	}

	blinder, err := ComputeAssetTagBlinder(sharedSecret)
	if err != nil {
		return false, err
	}

	recomputedAssetTag.Add(recomputedAssetTag, new(operation.Point).ScalarMult(operation.PedCom.G[PedersenRandomnessIndex], blinder))
	if operation.IsPointEqual(recomputedAssetTag, coin.GetAssetTag()) {
		return true, nil
	}
	return false, nil
}

func (coin *CoinV2) SetPlainTokenID(tokenID *common.Hash) error {
	assetTag := operation.HashToPoint(tokenID[:])
	coin.SetAssetTag(assetTag)
	com, err := coin.ComputeCommitmentCA()
	if err != nil {
		return err
	}
	coin.SetCommitment(com)
	return nil
}

// GetTokenId attempts to retrieve the asset a CoinV2.
// Parameters:
//   - keySet: the key set of the user, must contain an OTAKey
//   - rawAssetTags: a pre-computed mapping from a raw assetTag to the tokenId (e.g, HashToPoint(PRV) => PRV).
func (c *CoinV2) GetTokenId(keySet *incognitokey.KeySet, rawAssetTags map[string]*common.Hash) (*common.Hash, error) {
	if c.GetAssetTag() == nil {
		return &common.PRVCoinID, nil
	}

	if asset, ok := rawAssetTags[c.GetAssetTag().String()]; ok {
		return asset, nil
	}

	belong, sharedSecret := c.DoesCoinBelongToKeySet(keySet)
	if !belong {
		return nil, fmt.Errorf("coin does not belong to the keyset")
	}

	blinder := operation.HashToScalar(append(sharedSecret.ToBytesS(), []byte("assettag")...))
	rawAssetTag := new(operation.Point).Sub(
		c.GetAssetTag(),
		new(operation.Point).ScalarMult(operation.PedCom.G[PedersenRandomnessIndex], blinder),
	)

	if asset, ok := rawAssetTags[rawAssetTag.String()]; ok {
		return asset, nil
	}

	return nil, fmt.Errorf("cannot find the tokenId")
}

// for confidential asset only
func NewCoinCA(p *CoinParams, tokenID *common.Hash) (*CoinV2, *operation.Point, error) {
	receiverPublicKey, err := new(operation.Point).FromBytesS(p.PaymentAddress.Pk)
	if err != nil {
		errStr := fmt.Sprintf("Cannot parse outputCoinV2 from PaymentInfo when parseByte PublicKey, error %v ", err)
		return nil, nil, fmt.Errorf(errStr)
	}
	receiverPublicKeyBytes := receiverPublicKey.ToBytesS()
	targetShardID := common.GetShardIDFromLastByte(receiverPublicKeyBytes[len(receiverPublicKeyBytes)-1])

	c := new(CoinV2).Init()
	// Amount, Randomness, SharedRandom is transparency until we call concealData
	c.SetAmount(new(operation.Scalar).FromUint64(p.Amount))
	c.SetRandomness(operation.RandomScalar())
	c.SetSharedRandom(operation.RandomScalar()) // r
	c.SetSharedConcealRandom(operation.RandomScalar())
	c.SetInfo(p.Message)

	// If this is going to burning address then dont need to create ota
	if common.IsPublicKeyBurningAddress(p.PaymentAddress.Pk) {
		publicKey, err := new(operation.Point).FromBytesS(p.PaymentAddress.Pk)
		if err != nil {
			panic("Something is wrong with info.paymentAddress.pk, burning address should be a valid point")
		}
		c.SetPublicKey(publicKey)
		err = c.SetPlainTokenID(tokenID)
		if err != nil {
			return nil, nil, err
		}
		return c, nil, nil
	}

	// Increase index until have the right shardID
	index := uint32(0)
	publicOTA := p.PaymentAddress.GetOTAPublicKey() // For generating one-time-address
	if publicOTA == nil {
		return nil, nil, fmt.Errorf("public OTA from payment address is nil")
	}
	publicSpend := p.PaymentAddress.GetPublicSpend() // General public key
	rK := new(operation.Point).ScalarMult(publicOTA, c.GetSharedRandom())
	for i := MaxAttempts; i > 0; i-- {
		index++

		// Get publickey
		hash := operation.HashToScalar(append(rK.ToBytesS(), common.Uint32ToBytes(index)...))
		HrKG := new(operation.Point).ScalarMultBase(hash)
		publicKey := new(operation.Point).Add(HrKG, publicSpend)
		c.SetPublicKey(publicKey)

		senderShardID, recvShardID, coinPrivacyType, _ := DeriveShardInfoFromCoin(publicKey.ToBytesS())
		if recvShardID == int(targetShardID) && senderShardID == p.SenderShardID && coinPrivacyType == p.CoinPrivacyType {
			otaSharedRandomPoint := new(operation.Point).ScalarMultBase(c.GetSharedRandom())
			concealSharedRandomPoint := new(operation.Point).ScalarMultBase(c.GetSharedConcealRandom())
			c.SetTxRandomDetail(concealSharedRandomPoint, otaSharedRandomPoint, index)

			b := append(rK.ToBytesS(), common.Uint32ToBytes(index)...)
			b = append(b, []byte("otatag")...)
			hOTATag := operation.HashToScalar(b)
			b = hOTATag.ToBytesS()
			vt := uint8(b[operation.Ed25519KeySize - 1])
			c.otaTag = &vt

			rAsset := new(operation.Point).ScalarMult(publicOTA, c.GetSharedRandom())
			blinder, _ := ComputeAssetTagBlinder(rAsset)
			if tokenID == nil {
				return nil, nil, fmt.Errorf("cannot create coin without tokenID")
			}
			assetTag := operation.HashToPoint(tokenID[:])
			assetTag.Add(assetTag, new(operation.Point).ScalarMult(operation.PedCom.G[PedersenRandomnessIndex], blinder))
			c.SetAssetTag(assetTag)
			com, err := c.ComputeCommitmentCA()
			if err != nil {
				return nil, nil, fmt.Errorf("cannot compute commitment for confidential asset")
			}
			c.SetCommitment(com)

			return c, rAsset, nil
		}
	}
	// MaxAttempts could be exceeded if the OS's RNG or the statedb is corrupted
	return nil, nil, fmt.Errorf("cannot create OTA after %d attempts", MaxAttempts)
}
