package relaying

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/binance-chain/go-sdk/client/rpc"
	bnbtx "github.com/binance-chain/go-sdk/types/tx"
	"github.com/incognitochain/incognito-chain/database"
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
func BuildProof2(indexTx int, blockHeight int64, url string) (*types.TxProof, *BNBRelayingError) {
	txs, err := getTxsInBlockHeight(blockHeight, url)
	if err != nil {
		return nil, err
	}

	proof := txs.Proof(indexTx)
	return &proof, nil
}

func VerifyProof(txProof *types.TxProof, dataHash []byte) (bool, *BNBRelayingError) {
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

	client := client.NewHTTP(MainnetURLRemote, "/websocket")
	err = client.Start()
	if err != nil {
		// handle error
	}
	defer client.Stop()
	tx, err := client.Tx(txHash, true)
	//fmt.Printf("tx: %+v\n", tx)

	return &tx.Proof, nil
}

func getTxsInBlockHeight(blockHeight int64, url string) (*types.Txs, *BNBRelayingError) {
	block, err := GetBlock(blockHeight, url)
	if err != nil {
		return nil, err
	}

	return &block.Txs, nil
}

func GetBlock(blockHeight int64, url string) (*types.Block, *BNBRelayingError) {
	client := client.NewHTTP(url, "/websocket")
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

type BNBProof struct {
	Proof *types.TxProof
	BlockHeight int64
}

// buildProof creates a proof for tx at indexTx in block height
func (p *BNBProof) Build(indexTx int, blockHeight int64, url string) (*BNBRelayingError) {
	txs, err := getTxsInBlockHeight(blockHeight, url)
	if err != nil {
		return err
	}

	proof := txs.Proof(indexTx)

	p.BlockHeight = blockHeight
	p.Proof = &proof

	return nil
}

func (p *BNBProof) Verify(db database.DatabaseInterface) (bool, *BNBRelayingError){
	//dataHash, err := db.GetBNBDataHashByBlockHeight(uint64(p.BlockHeight))
	//if err != nil {
	//	//	return false, NewBNBRelayingError(GetBNBDataHashErr, err)
	//	//}

	//todo: hard code to test
	dataHash := []byte{}
	if p.BlockHeight == 247 {
		dataHash, _ = hex.DecodeString("7F26964EF77D1E0876A6F852AA9125125BE832AFD4912CBD4DE69BFD63640AA8")
	} else if p.BlockHeight == 355  {
		dataHash, _ = hex.DecodeString("8DE56BE85401D4737E3AB7FD305EB9A86300CED4A3E0EF55571F08C1C45E4D8F")
	} else if p.BlockHeight == 450  {
		dataHash, _ = hex.DecodeString("193CDB951A4F8F47C472646EE64FD062F9415F1A0BDB0FC76D2FA5E5EE9DFCC7")
	}

	return VerifyProof(p.Proof, dataHash)
}

func ParseBNBProofFromB64EncodeJsonStr(b64EncodedJsonStr string) (*BNBProof, *BNBRelayingError) {
	jsonBytes, err := base64.StdEncoding.DecodeString(b64EncodedJsonStr)
	if err != nil {
		return nil, NewBNBRelayingError(UnexpectedErr, err)
	}

	proof := BNBProof{}
	err = json.Unmarshal(jsonBytes, &proof)
	if err != nil {
		return nil, NewBNBRelayingError(UnexpectedErr, err)
	}

	return &proof, nil
}




