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
	prvGammaShape = 0.25
	prvGammaScale = 55629.11

	tokenGammaShape = 0.44
	tokenGammaScale = 26088.24

	unitTime = 40 // 1 block

	// There might be cases where the chosen blocks do not contain any output coins. In this case,
	// we will try to choose a random output coin from one of blocks in the interval [k - 10: k + 10].
	blockDeviation = 10

	MaxGammaTries = 1000
)

// gammaPicker implements a Gamma distribution picker for choosing random decoys.
type gammaPicker struct {
	distuv.Gamma
}

// NewGammaPicker returns a new gammaPicker.
// It is used for the MAIN-NET only.
func NewGammaPicker(shape, scale float64) *gammaPicker {
	randSrc := rand.NewSource(common.RandUint64())
	gamma := distuv.Gamma{
		Alpha: shape,
		Beta:  1 / scale,
		Src:   randSrc,
	}

	return &gammaPicker{gamma}
}

// Pick returns a random CoinV2 from the pre-defined Gamma distribution.
func Pick(db *statedb.StateDB, shardID byte, tokenID common.Hash, latestHeight uint64) (*big.Int, *coin.CoinV2, error) {
	gp := NewGammaPicker(prvGammaShape, prvGammaScale)
	if tokenID.String() != common.PRVIDStr {
		gp = NewGammaPicker(tokenGammaShape, tokenGammaScale)
	}

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
