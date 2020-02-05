package relaying

import (
	"fmt"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"
)

// buildProof creates a proof for tx at indexTx in block height
func buildProof(blockHeight int64, indexTx int) (*types.TxProof, *BNBRelayingError) {
	//txs, err := getTxsInBlockHeight(blockHeight)
	//if err != nil {
	//	return nil, err
	//}
	//
	//proof := txs.Proof(indexTx)
	return nil, nil
}

func verifyProof(txProof *types.TxProof, blockHeight int64, dataHash []byte) (bool, *BNBRelayingError){
	// todo: get dataHash in blockHeight from db
	// if there is no blockheight in db, return error
	//dataHash := []byte{}
	err := txProof.Validate(dataHash)
	if err != nil {
		return false, NewBNBRelayingError(InvalidTxProofErr, err)
	}
	return true, nil
}

func getTxByHash(txHash string) (types.Txs, *BNBRelayingError){
	// todo: call API to Binance node to get all txs in block height
	client := client.NewHTTP("https://seed1.longevito.io:443", "/ws")
	err := client.Start()
	if err != nil {
		// handle error
		fmt.Printf("Err: %v\n", err)

	}
	defer client.Stop()
	tx, err := client.Tx([]byte(txHash), true)
	fmt.Printf("tx : %v\n", tx)

	return types.Txs{}, nil
}


