package relaying

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/binance-chain/go-sdk/client/rpc"
	bnbtx "github.com/binance-chain/go-sdk/types/tx"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"
)

// there are 2 ways to build proof
// 1. Call API get tx by hash with param prove = true,
// It returns the transaction and the proof for that transaction
// 2. Call API get all txs in a block height and build proof from those txs by Tendermint's code

func BuildProof1(txHash string) (*types.TxProof, *BNBRelayingError) {
	txProof, err := getProofFromTxHash(txHash)
	if err != nil {
		return nil, err
	}

	return txProof, nil
}

// buildProof creates a proof for tx at indexTx in block height
func BuildProof2(indexTx int, blockHeight int64) (*types.TxProof, *BNBRelayingError) {
	txs, err := getTxsInBlockHeight(blockHeight)
	if err != nil {
		return nil, err
	}

	proof := txs.Proof(indexTx)
	return &proof, nil
}

func VerifyProof(txProof *types.TxProof, blockHeight int64, dataHash []byte) (bool, *BNBRelayingError) {
	// todo: get dataHash in blockHeight from db
	// if there is no blockheight in db, return error
	//dataHash := []byte{}
	err := txProof.Validate(dataHash)
	if err != nil {
		return false, NewBNBRelayingError(InvalidTxProofErr, err)
	}
	return true, nil
}

func getProofFromTxHash(txHashStr string) (*types.TxProof, *BNBRelayingError) {
	txHash, err := hex.DecodeString(txHashStr)
	if err != nil {
		return nil, NewBNBRelayingError(UnexpectedErr, err)
	}

	client := client.NewHTTP(URLRemote, "/websocket")
	err = client.Start()
	if err != nil {
		// handle error
	}
	defer client.Stop()
	tx, err := client.Tx(txHash, true)
	//fmt.Printf("tx: %+v\n", tx)

	return &tx.Proof, nil
}

func getTxsInBlockHeight(blockHeight int64) (*types.Txs, *BNBRelayingError) {
	block, err := getBlock(blockHeight)
	if err != nil {
		return nil, err
	}

	return &block.Txs, nil
}

func getBlock(blockHeight int64) (*types.Block, *BNBRelayingError) {
	client := client.NewHTTP(URLRemote, "/websocket")
	err := client.Start()
	if err != nil {
		// handle error
	}
	defer client.Stop()
	block, err := client.Block(&blockHeight)
	fmt.Printf("block: %+v\n", block)

	return block.Block, nil
}

func ParseProofFromB64EncodeJsonStr(b64EncodedJsonStr string) (*types.TxProof, *BNBRelayingError) {
	jsonBytes, err := base64.StdEncoding.DecodeString(b64EncodedJsonStr)
	if err != nil {
		return nil, NewBNBRelayingError(UnexpectedErr, err)
	}

	proof := types.TxProof{}
	err = json.Unmarshal(jsonBytes, &proof)
	if err != nil {
		return nil, NewBNBRelayingError(UnexpectedErr, err)
	}

	return &proof, nil
}

func ParseTxFromData(data []byte) (*bnbtx.StdTx, *BNBRelayingError) {
	tx, err := rpc.ParseTx(bnbtx.Cdc, data)
	if err != nil {
		return nil, NewBNBRelayingError(UnexpectedErr, err)
	}
	stdTx := tx.(bnbtx.StdTx)
	return &stdTx, nil
}


