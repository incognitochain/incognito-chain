package aggregaterange

import (
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

/***** Bullet proof component *****/

// bulletproofParams includes all generator for aggregated range proof
type bulletproofParams struct {
	g []*privacy.Point
	h []*privacy.Point
	u *privacy.Point
}

func newBulletproofParams(m int) *bulletproofParams {
	gen := new(bulletproofParams)
	capacity := maxExp * m // fixed value
	gen.g = make([]*privacy.Point, capacity)
	gen.h = make([]*privacy.Point, capacity)

	var wg sync.WaitGroup
	wg.Add(capacity)
	for i := 0; i < capacity; i++ {
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			gen.g[i] = privacy.HashToPoint(int64(5 + i))
			gen.h[i] = privacy.HashToPoint(int64(5 + i + maxOutputNumber*maxExp))
		}(i, &wg)
	}
	wg.Wait()
	gen.u = new(privacy.Point)
	gen.u = privacy.HashToPoint(int64(5 + 2*maxOutputNumber*maxExp))

	return gen
}

func addBulletproofParams(extraNumber int) *bulletproofParams {
	currentCapacity := len(AggParam.g)
	newCapacity := currentCapacity + maxExp * extraNumber

	for i := 0; i < newCapacity - currentCapacity; i++ {
		AggParam.g = append(AggParam.g, privacy.HashToPoint(int64(5 + i + currentCapacity)))
		AggParam.h = append(AggParam.h, privacy.HashToPoint(int64(5 + i + currentCapacity + maxOutputNumber*maxExp)))
	}

	return AggParam
}

var AggParam = newBulletproofParams(numOutputParam)

func generateChallengeForAggRange(AggParam *bulletproofParams, values [][]byte) *privacy.Scalar {
	bytes := []byte{}
	for i := 0; i < len(AggParam.g); i++ {
		bytes = append(bytes, privacy.ArrayToSlice(AggParam.g[i].ToBytes())...)
	}

	for i := 0; i < len(AggParam.h); i++ {
		bytes = append(bytes, privacy.ArrayToSlice(AggParam.h[i].ToBytes())...)
	}

	bytes = append(bytes, privacy.ArrayToSlice(AggParam.u.ToBytes())...)

	for i := 0; i < len(values); i++ {
		bytes = append(bytes, values[i]...)
	}

	hash := common.HashB(bytes)

	res := new(privacy.Scalar).FromBytes(privacy.SliceToArray(hash))
	return res
}
