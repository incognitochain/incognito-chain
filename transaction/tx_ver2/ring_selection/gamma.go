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
	prvGammaShape = 0.62
	prvGammaScale = 16.0

	tokenGammaShape = 1.74
	tokenGammaScale = 13.31

	unitTime = 3600 * 24 // 1 day

	MaxGammaTries = 1000
)

// GammaPicker implements a Gamma distribution picker for choosing random decoys.
type GammaPicker struct {
	distuv.Gamma
}

// NewGammaPicker returns a new GammaPicker.
// It is used for the MAIN-NET only.
func NewGammaPicker(shape, scale float64) *GammaPicker {
	randSrc := rand.NewSource(common.RandUint64())
	gamma := distuv.Gamma{
		Alpha: shape,
		Beta:  1 / scale,
		Src:   randSrc,
	}

	return &GammaPicker{gamma}
}

// Pick returns a random CoinV2 from the pre-defined Gamma distribution.
func Pick(db *statedb.StateDB, shardID byte, tokenID common.Hash, latestHeight uint64) (*big.Int, *coin.CoinV2, error) {
	gp := NewGammaPicker(prvGammaShape, prvGammaScale)
	if tokenID.String() != common.PRVIDStr {
		gp = NewGammaPicker(tokenGammaShape, tokenGammaScale)
	}

	x := gp.Rand()
	passedBlock := uint64(math.Ceil(x * unitTime / config.Param().BlockTime.MaxShardBlockCreation.Seconds()))
	if passedBlock > latestHeight {
		utils.Logger.Log.Errorf("bad pick: passedBlock %v is greater than the current block %v, shardID %v\n", passedBlock, latestHeight, shardID)
		return nil, nil, fmt.Errorf("bad pick: passedBlock %v is greater than the current block %v, shardID %v", passedBlock, latestHeight, shardID)
	}

	blkHeight := latestHeight - passedBlock
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
		return nil, nil, fmt.Errorf(msg)
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
