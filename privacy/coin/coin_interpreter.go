package coin

import "github.com/incognitochain/incognito-chain/privacy/operation"

// Commit coin only with g^v * h^r
func ParseCommitmentToV2WithCoin(c *PlainCoinV1) *operation.Point {
	return operation.PedCom.CommitAtIndex(
		new(operation.Scalar).FromUint64(c.GetValue()),
		c.GetRandomness(),
		operation.PedersenValueIndex,
	)
}

func ParseCommitmentToV2(commitment *operation.Point, publicKey *operation.Point, snd *operation.Scalar, shardID byte) *operation.Point {
	// This is already v2 coin
	if snd == nil {
		return commitment
	}
	G2SND := new(operation.Point).ScalarMult(
		operation.PedCom.G[operation.PedersenSndIndex],
		snd,
	)
	G3Shard := new(operation.Point).ScalarMult(
		operation.PedCom.G[operation.PedersenShardIDIndex],
		new(operation.Scalar).FromUint64(uint64(shardID)),
	)
	commitment.Sub(commitment, publicKey)
	commitment.Sub(commitment, G2SND)
	commitment.Sub(commitment, G3Shard)
	return commitment
}

func ParseCommitmentToV2ByBytes(commitmentBytes []byte, publicKeyBytes []byte, sndBytes []byte, shardID byte) ([]byte, error) {
	// Check if commitmentBytes is bug
	commitment, err := new(operation.Point).FromBytesS(commitmentBytes)
	if err != nil {
		return nil, err
	}
	// This is already a ver2 coin
	if len(sndBytes) == 0 {
		return commitmentBytes, nil
	}
	publicKey, err := new(operation.Point).FromBytesS(publicKeyBytes)
	if err != nil {
		return nil, err
	}
	snd := new(operation.Scalar).FromBytesS(sndBytes)
	commitment = ParseCommitmentToV2(commitment, publicKey, snd, shardID)
	return commitment.ToBytesS(), nil
}
