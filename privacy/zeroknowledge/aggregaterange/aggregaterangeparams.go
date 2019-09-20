package aggregaterange

import (
	"math/big"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

/***** Bullet proof component *****/

// bulletproofParams includes all generator for aggregated range proof
type bulletproofParams struct {
	g []*privacy.EllipticPoint
	h []*privacy.EllipticPoint
	u *privacy.EllipticPoint
}

func newBulletproofParams(m int) *bulletproofParams {
	gen := new(bulletproofParams)
	capacity := maxExp * m // fixed value
	gen.g = make([]*privacy.EllipticPoint, capacity)
	gen.h = make([]*privacy.EllipticPoint, capacity)

	var wg sync.WaitGroup
	wg.Add(capacity)
	for i := 0; i < capacity; i++ {
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			gen.g[i] = privacy.PedCom.G[0].Hash(int64(5 + i))
			gen.h[i] = privacy.PedCom.G[0].Hash(int64(5 + i + maxOutputNumber*maxExp))
		}(i, &wg)
	}
	wg.Wait()
	gen.u = new(privacy.EllipticPoint)
	gen.u = gen.h[0].Hash(int64(5 + 2*maxOutputNumber*64))

	return gen
}

func addBulletproofParams(extraNumber int) *bulletproofParams {
	currentCapacity := len(AggParam.g)
	newCapacity := currentCapacity + maxExp * extraNumber

	for i := 0; i < newCapacity - currentCapacity; i++ {
		AggParam.g = append(AggParam.g, privacy.PedCom.G[0].Hash(int64(5 + i + currentCapacity)))
		AggParam.h = append(AggParam.h, privacy.PedCom.G[0].Hash(int64(5 + i + currentCapacity + maxOutputNumber*maxExp)))
	}

	return AggParam
}

var AggParam = newBulletproofParams(numOutputParam)

func generateChallengeForAggRange(AggParam *bulletproofParams, values [][]byte) *big.Int {
	bytes := AggParam.g[0].Compress()
	for i := 1; i < len(AggParam.g); i++ {
		bytes = append(bytes, AggParam.g[i].Compress()...)
	}

	for i := 0; i < len(AggParam.h); i++ {
		bytes = append(bytes, AggParam.h[i].Compress()...)
	}

	bytes = append(bytes, AggParam.u.Compress()...)

	for i := 0; i < len(values); i++ {
		bytes = append(bytes, values[i]...)
	}

	hash := common.HashB(bytes)

	res := new(big.Int).SetBytes(hash)
	res.Mod(res, privacy.Curve.Params().N)
	return res
}
