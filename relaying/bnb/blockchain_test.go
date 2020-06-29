package bnb

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/types"
	"testing"
)

func getBNBHeaderStrFromBinanceNetwork(blockHeight int64, url string) (string, error) {
	blockHeader, err := getBNBHeaderFromBinanceNetwork(blockHeight, url)
	if err != nil {
		return "", err
	}
	bnbHeaderBytes, err2 := json.Marshal(blockHeader)
	if err2 != nil {
		return "", err2
	}

	bnbHeaderStr := base64.StdEncoding.EncodeToString(bnbHeaderBytes)
	return bnbHeaderStr, nil
}

func getBNBHeaderFromBinanceNetwork(blockHeight int64, url string) (*types.Block, error) {
	block, err := GetBlock(blockHeight, url)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, errors.New("Can not get block from bnb chain")
	}

	blockHeader := &types.Block{
		Header:     block.Header,
		LastCommit: block.LastCommit,
	}
	return blockHeader, nil
}

func TestProcessNewBlock(t *testing.T) {
	// set up bnb chain state with genesis block
	state := new(BNBChainState)
	state.LatestBlock = getGenesisBNBBlockTestnet()
	nextBlockHeight := int64(TestnetGenesisBlockHeight + 1)
	oldBlockHeight := nextBlockHeight

	fmt.Printf("Before state.LatestBlock.Height : %v\n", state.LatestBlock.Height)

	for i := 0; i < 10; i++ {
		nextBlock, err := getBNBHeaderFromBinanceNetwork(nextBlockHeight, TestnetURLRemote)
		assert.Nil(t, err)

		err = state.ProcessNewBlock(nextBlock, TestnetBNBChainID)
		assert.Nil(t, err)
		fmt.Printf("bnb chain state after processing block height: %v\n", nextBlockHeight)
		fmt.Printf("state.FinalBlocks: %v\n", state.FinalBlocks)
		fmt.Printf("state.LatestBlock: %v\n", state.LatestBlock)
		fmt.Printf("state.CandidateNextBlocks: %v\n", state.CandidateNextBlocks)
		fmt.Printf("state.OrphanBlocks: %v\n", state.OrphanBlocks)
		fmt.Printf("====================================================================\n")

		nextBlockHeight += 2
	}

	for i := 0; i < 10; i++ {
		nextBlock, err := getBNBHeaderFromBinanceNetwork(oldBlockHeight, TestnetURLRemote)
		assert.Nil(t, err)

		err = state.ProcessNewBlock(nextBlock, TestnetBNBChainID)
		assert.Nil(t, err)
		fmt.Printf("bnb chain state after processing block height: %v\n", oldBlockHeight)
		fmt.Printf("state.FinalBlocks: %v\n", state.FinalBlocks)
		fmt.Printf("state.LatestBlock: %v\n", state.LatestBlock)
		fmt.Printf("state.CandidateNextBlocks: %v\n", state.CandidateNextBlocks)
		fmt.Printf("state.OrphanBlocks: %v\n", state.OrphanBlocks)
		fmt.Printf("====================================================================\n")

		oldBlockHeight++
	}
}
