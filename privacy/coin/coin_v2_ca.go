package coin

import(
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

const MAX_TRIES_OTA int = 50000

func (coin *CoinV2) ComputeCommitmentCA() (*operation.Point,error){
	if coin==nil || coin.GetRandomness()==nil || coin.GetAmount()==nil{
		return nil, errors.New("missing arguments for committing")
	}
	gValue_immutable := operation.PedCom.G[operation.PedersenValueIndex]
	commitment := new(operation.Point).ScalarMult(coin.GetAssetTag(),coin.GetRandomness())
	commitment.Add(commitment,new(operation.Point).ScalarMult(gValue_immutable,coin.GetAmount()))
	return commitment,nil
}

func ComputeCommitmentCA(assetTag *operation.Point, r, v *operation.Scalar) (*operation.Point,error){
	if assetTag==nil || r==nil || v==nil{
		return nil, errors.New("missing arguments for committing to CA coin")
	}
	gValue_immutable := operation.PedCom.G[operation.PedersenValueIndex]
	commitment := new(operation.Point).ScalarMult(assetTag,r)
	commitment.Add(commitment,new(operation.Point).ScalarMult(gValue_immutable,v))
	return commitment,nil
}

func ComputeAssetTagBlinder(sharedSecret *operation.Point, indexForShard uint32) (*operation.Scalar,error){
	if sharedSecret==nil{
		return nil, errors.New("missing arguments for asset tag blinder")
	}
	blinder := operation.HashToScalar(append(sharedSecret.ToBytesS(), append([]byte("assettag"), common.Uint32ToBytes(indexForShard)...)...))
	return blinder, nil
}

// this should be an input coin
func (coin *CoinV2) RecomputeSharedSecret(privateKey []byte) (*operation.Point,error){
	pk := new(operation.Scalar).FromBytesS(privateKey)
	// this is g^SharedRandom, previously created by sender of the coin
	sharedRanPoint, _, err := coin.GetTxRandomDetail()
	if err != nil {
		return nil, errors.New("cannot retrieve tx random detail")
	}
	sharedSecret := new(operation.Point).ScalarMult(sharedRanPoint, pk)
	return sharedSecret, nil
}

// for confidential asset only
func GenerateOTACoinAndSharedSecret(info *key.PaymentInfo, tokenID *common.Hash) (*CoinV2, *operation.Point, error) {
	receiverPublicKey, err := new(operation.Point).FromBytesS(info.PaymentAddress.Pk)
	if err != nil {
		errStr := fmt.Sprintf("Cannot parse outputCoinV2 from PaymentInfo when parseByte PublicKey, error %v ", err)
		return nil, nil, errors.New(errStr)
	}
	receiverPublicKeyBytes := receiverPublicKey.ToBytesS()
	targetShardID := common.GetShardIDFromLastByte(receiverPublicKeyBytes[len(receiverPublicKeyBytes)-1])

	c := new(CoinV2).Init()
	// Amount, Randomness, SharedRandom is transparency until we call concealData
	c.SetAmount(new(operation.Scalar).FromUint64(info.Amount))
	c.SetRandomness(operation.RandomScalar())
	c.SetSharedRandom(operation.RandomScalar()) // r
	c.SetInfo(info.Message)
	c.SetCommitment(operation.PedCom.CommitAtIndex(c.GetAmount(), c.GetRandomness(), operation.PedersenValueIndex))

	// If this is going to burning address then dont need to create ota
	if wallet.IsPublicKeyBurningAddress(info.PaymentAddress.Pk) {
		publicKey, err := new(operation.Point).FromBytesS(info.PaymentAddress.Pk)
		if err != nil {
			panic("Something is wrong with info.paymentAddress.Pk, burning address should be a valid point")
		}
		c.SetPublicKey(publicKey)
		return c, nil, nil
	}

	// Increase index until have the right shardID
	index := uint32(0)
	publicView := info.PaymentAddress.GetPublicView()
	publicSpend := info.PaymentAddress.GetPublicSpend()
	rK := new(operation.Point).ScalarMult(publicView, c.GetSharedRandom())
	for i:=MAX_TRIES_OTA;i>0;i--{
		index += 1

		// Get publickey
		hash := operation.HashToScalar(append(rK.ToBytesS(), common.Uint32ToBytes(index)...))
		HrKG := new(operation.Point).ScalarMultBase(hash)
		publicKey := new(operation.Point).Add(HrKG, publicSpend)
		c.SetPublicKey(publicKey)

		currentShardID, err := c.GetShardID()
		if err != nil {
			return nil, nil, err
		}
		if currentShardID == targetShardID {
			c.SetTxRandomDetail(new(operation.Point).ScalarMultBase(c.GetSharedRandom()), index)
			blinder,_ := ComputeAssetTagBlinder(rK,index)
			if tokenID == nil {
				return nil, nil, errors.New("Cannot create coin without tokenID")
			}
			assetTag := operation.HashToPoint(tokenID[:])
			assetTag.Add(assetTag,new(operation.Point).ScalarMultBase(blinder))
			c.SetAssetTag(assetTag)
			com, err := c.ComputeCommitmentCA()
			if err != nil{
				return nil, nil, errors.New("Cannot compute commitment for confidential asset")
			}
			c.SetCommitment(com)
			
			return c, rK, nil
		}
	}
	// MAX_TRIES_OTA could be exceeded if the OS's RNG or the statedb is corrupted
	return nil, nil, errors.New(fmt.Sprintf("Cannot create OTA after %d attempts", MAX_TRIES_OTA))
}