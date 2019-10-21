package aggregaterange

import (
	"github.com/incognitochain/incognito-chain/privacy"
)

/***** Bullet proof component *****/

const (
	maxExp          = 64
	numOutputParam  = 32
	maxOutputNumber = 256
)

// bulletproofParams includes all generator for aggregated range proof
type bulletproofParams struct {
	g []*privacy.Point
	h []*privacy.Point
	u *privacy.Point
	//gPrecomputed [][8]C25519.CachedGroupElement
	//hPrecomputed [][8]C25519.CachedGroupElement

	//gPreMultiScalar [][8]C25519.CachedGroupElement
	//hPreMultiScalar [][8]C25519.CachedGroupElement
}

var AggParam = newBulletproofParams(numOutputParam)

func InitParam(noOutPutParam int) *bulletproofParams {
	AggParam = newBulletproofParams(noOutPutParam)
	return AggParam
}

func newBulletproofParams(m int) *bulletproofParams {
	gen := new(bulletproofParams)
	capacity := maxExp * m // fixed value
	gen.g = make([]*privacy.Point, capacity)
	gen.h = make([]*privacy.Point, capacity)

	//gen.gPrecomputed = make([][8]C25519.CachedGroupElement, capacity)
	//gen.hPrecomputed = make([][8]C25519.CachedGroupElement, capacity)
	//
	//gen.gPreMultiScalar = make([][8]C25519.CachedGroupElement, capacity)
	//gen.hPreMultiScalar = make([][8]C25519.CachedGroupElement, capacity)

	for i := 0; i < capacity; i++ {
		gen.g[i] = privacy.HashToPointFromIndex(int64(5 + i))
		//tmpKey := gen.g[i].GetKey()
		//gE := new(C25519.ExtendedGroupElement)
		//gE.FromBytes(&tmpKey)
		//C25519.GePrecompute(&gen.gPrecomputed[i], gE)
		//gen.gPreMultiScalar[i] = C25519.PreComputeForMultiScalar(&tmpKey)

		gen.h[i] = privacy.HashToPointFromIndex(int64(5 + i + maxOutputNumber*maxExp))
		//tmpKey = gen.h[i].GetKey()
		//hE := new(C25519.ExtendedGroupElement)
		//hE.FromBytes(&tmpKey)
		//C25519.GePrecompute(&gen.hPrecomputed[i], hE)
		//gen.hPreMultiScalar[i] = C25519.PreComputeForMultiScalar(&tmpKey)
	}
	gen.u = new(privacy.Point)
	gen.u = privacy.HashToPointFromIndex(int64(5 + 2*maxOutputNumber*maxExp))

	return gen
}

func addBulletproofParams(extraNumber int) *bulletproofParams {
	currentCapacity := len(AggParam.g)
	newCapacity := currentCapacity + maxExp*extraNumber

	for i := 0; i < newCapacity-currentCapacity; i++ {
		AggParam.g = append(AggParam.g, privacy.HashToPointFromIndex(int64(5+i+currentCapacity)))
		//tmpKey := AggParam.g[i].GetKey()
		//gE := new(C25519.ExtendedGroupElement)
		//gE.FromBytes(&tmpKey)
		//var gPre [8]C25519.CachedGroupElement
		//C25519.GePrecompute(&gPre, gE)
		//AggParam.gPrecomputed = append(AggParam.gPrecomputed, gPre)
		//AggParam.gPreMultiScalar = append(AggParam.gPreMultiScalar, C25519.PreComputeForMultiScalar(&tmpKey))

		AggParam.h = append(AggParam.h, privacy.HashToPointFromIndex(int64(5+i+currentCapacity+maxOutputNumber*maxExp)))
		//tmpKey = AggParam.h[i].GetKey()
		//hE := new(C25519.ExtendedGroupElement)
		//hE.FromBytes(&tmpKey)
		//var hPre [8]C25519.CachedGroupElement
		//C25519.GePrecompute(&hPre, hE)
		//AggParam.hPrecomputed = append(AggParam.hPrecomputed, hPre)
		//AggParam.hPreMultiScalar = append(AggParam.hPreMultiScalar, C25519.PreComputeForMultiScalar(&tmpKey))
	}

	return AggParam
}

func generateChallengeForAggRange(AggParam *bulletproofParams, values [][]byte) *privacy.Scalar {
	bytes := []byte{}
	for i := 0; i < len(AggParam.g); i++ {
		bytes = append(bytes, AggParam.g[i].ToBytesS()...)
	}

	for i := 0; i < len(AggParam.h); i++ {
		bytes = append(bytes, AggParam.h[i].ToBytesS()...)
	}

	bytes = append(bytes, AggParam.u.ToBytesS()...)

	for i := 0; i < len(values); i++ {
		bytes = append(bytes, values[i]...)
	}

	hash := privacy.HashToScalar(bytes)
	return hash
}
