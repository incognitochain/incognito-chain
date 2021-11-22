package ring_selection

import (
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/transaction/utils"
	"github.com/incognitochain/incognito-chain/wallet"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
	"math"
	"math/big"
)

const (
	prvGammaShape = 0.23
	prvGammaScale = 101579.72

	tokenGammaShape = 0.22
	tokenGammaScale = 43115.41

	unitTime = 40 // 1 block

	// There might be cases where the chosen blocks do not contain any output coins. In this case,
	// we will try to choose a random output coin from one of blocks in the interval [k - 10: k + 10].
	blockDeviation = 10

	MaxGammaTries = 1000
)

// var prvGammaParams = map[byte]gammaParam{
//	0: {shape: 0.21, scale: 62009.44},
//	1: {shape: 0.39, scale: 23779.52},
//	2: {shape: 0.35, scale: 32336.8},
//	3: {shape: 0.25, scale: 45754.25},
//	4: {shape: 0.32, scale: 38802.11},
//	5: {shape: 0.21, scale: 107515.24},
//	6: {shape: 0.26, scale: 35210.31},
//	7: {shape: 0.26, scale: 47902.53},
//}

// var tokenGammaParams = map[byte]gammaParam{
//	0: {shape: 0.21, scale: 69940.95},
//	1: {shape: 0.54, scale: 35918.5},
//	2: {shape: 0.27, scale: 32638.93},
//	3: {shape: 0.4, scale: 26112.0},
//	4: {shape: 0.82, scale: 26394.58},
//	5: {shape: 0.69, scale: 54912.61},
//	6: {shape: 0.27, scale: 45319.47},
//	7: {shape: 0.27, scale: 52005.54},
//}

type gammaParam struct {
	shape float64
	scale float64
}

// gammaPicker implements a Gamma distribution picker for choosing random decoys.
type gammaPicker struct {
	distuv.Gamma
}

// NewGammaPicker returns a new gammaPicker.
// It is used for the MAIN-NET only.
func NewGammaPicker(param gammaParam) *gammaPicker {
	randSrc := rand.NewSource(common.RandUint64())
	gamma := distuv.Gamma{
		Alpha: param.shape,
		Beta:  1 / param.scale,
		Src:   randSrc,
	}

	return &gammaPicker{gamma}
}

// Pick returns a random CoinV2 from the pre-defined Gamma distribution.
func Pick(db *statedb.StateDB, shardID byte, tokenID common.Hash, latestHeight uint64) (*big.Int, *coin.CoinV2, error) {
	param := gammaParam{shape: prvGammaShape, scale: prvGammaScale}
	if tokenID.String() != common.PRVIDStr {
		param = gammaParam{shape: tokenGammaShape, scale: tokenGammaScale}
	}
	gp := NewGammaPicker(param)

	x := gp.Rand()
	passedBlock := uint64(math.Round(x * unitTime / config.Param().BlockTime.MinShardBlockInterval.Seconds()))
	attempt := 0
	tmpPassedBlock := passedBlock
	for attempt < 2*blockDeviation {
		if tmpPassedBlock > latestHeight {
			utils.Logger.Log.Errorf("bad pick: passedBlock %v is greater than the current block %v, shardID %v\n", passedBlock, latestHeight, shardID)
			return nil, nil, fmt.Errorf("bad pick: passedBlock %v is greater than the current block %v, shardID %v", passedBlock, latestHeight, shardID)
		}

		blkHeight := latestHeight - tmpPassedBlock
		currentHeightCoins, err := statedb.GetOTACoinsByHeight(db, tokenID, shardID, blkHeight)
		if err != nil {
			utils.Logger.Log.Errorf("bad pick: GetOTACoinsByHeight(%v, %v, %v) error: %v\n", tokenID.String(), shardID, blkHeight, err)
			return nil, nil, err
		}

		burningPubKey := wallet.GetBurningPublicKey()
		allCoins := make([]*coin.CoinV2, 0)
		for _, coinBytes := range currentHeightCoins {
			tmpCoin := &coin.CoinV2{}
			err = tmpCoin.SetBytes(coinBytes)
			if err != nil {
				continue
			}

			// check if the output coin was sent to the burning address
			if bytes.Equal(tmpCoin.GetPublicKey().ToBytesS(), burningPubKey) {
				continue
			}
			allCoins = append(allCoins, tmpCoin)
		}

		if len(allCoins) == 0 {
			msg := fmt.Sprintf("bad pick: no coin found for shard %v, tokenID %v, blkHeight %v", shardID, tokenID.String(), blkHeight)
			utils.Logger.Log.Errorf("%v\n", msg)

			if common.RandInt()%2 == 0 {
				tmpPassedBlock = passedBlock + 1 + common.RandUint64()%blockDeviation
			} else {
				tmpPassedBlock = passedBlock - 1 - common.RandUint64()%blockDeviation
			}
			attempt++
			continue
		}

		chosenCoin := allCoins[common.RandInt()%len(allCoins)]
		chosenIdx, err := statedb.GetOTACoinIndex(db, tokenID, chosenCoin.GetPublicKey().ToBytesS())
		if err != nil {
			utils.Logger.Log.Errorf("bad pick: GetOTACoinIndex for shard %v, tokenID %v, publicKey %v error: %v\n",
				shardID, tokenID.String(), chosenCoin.GetPublicKey().ToBytesS(), err,
			)
			return nil, nil, err
		}

		return chosenIdx, chosenCoin, nil
	}

	return nil, nil, fmt.Errorf("bad pick: no coin found")
}
