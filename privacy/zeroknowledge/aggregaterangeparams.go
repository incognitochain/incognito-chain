package zkp

import (
	"math/big"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

/***** Bullet proof component *****/

// bulletproofParams includes all generator for aggregated range proof
type bulletproofParams struct {
	G []*privacy.EllipticPoint
	H []*privacy.EllipticPoint
	U *privacy.EllipticPoint
}

func newBulletproofParams(m int) *bulletproofParams {
	gen := new(bulletproofParams)
	capacity := 64 * m // fixed value
	gen.G = make([]*privacy.EllipticPoint, capacity)
	gen.H = make([]*privacy.EllipticPoint, capacity)

	var wg sync.WaitGroup
	wg.Add(capacity)
	for i := 0; i < capacity; i++ {
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			gen.G[i] = privacy.PedCom.G[0].Hash(int64(5 + i))
			gen.H[i] = privacy.PedCom.G[0].Hash(int64(5 + i + capacity))
		}(i, &wg)
	}
	wg.Wait()
	gen.U = new(privacy.EllipticPoint)
	gen.U = gen.H[0].Hash(int64(5 + 2*capacity))

	return gen
}

func generateChallengeForAggRange(AggParam *bulletproofParams, values [][]byte) *big.Int {
	bytes := AggParam.G[0].Compress()
	for i := 1; i < len(AggParam.G); i++ {
		bytes = append(bytes, AggParam.G[i].Compress()...)
	}

	for i := 0; i < len(AggParam.H); i++ {
		bytes = append(bytes, AggParam.H[i].Compress()...)
	}

	bytes = append(bytes, AggParam.U.Compress()...)

	for i := 0; i < len(values); i++ {
		bytes = append(bytes, values[i]...)
	}

	hash := common.HashB(bytes)

	res := new(big.Int).SetBytes(hash)
	res.Mod(res, privacy.Curve.Params().N)
	return res
}
