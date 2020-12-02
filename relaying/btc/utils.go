package btcrelaying

import (
	"encoding/hex"

	"github.com/blockcypher/gobcy"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/incognitochain/incognito-chain/common/base58"
	"golang.org/x/crypto/sha3"
)

// ConvertExternalBTCAmountToIncAmount converts amount in bTC chain (decimal 8) to amount in inc chain (decimal 9)
func ConvertExternalBTCAmountToIncAmount(externalBTCAmount int64) int64 {
	return externalBTCAmount * 10 // externalBTCAmount / 1^8 * 1^9
}

// ConvertIncPBTCAmountToExternalBTCAmount converts amount in inc chain (decimal 9) to amount in bTC chain (decimal 8)
func ConvertIncPBTCAmountToExternalBTCAmount(incPBTCAmount int64) int64 {
	return incPBTCAmount / 10 // incPBTCAmount / 1^9 * 1^8
}

func HashAndEncodeBase58(msg string) string {
	hash := make([]byte, 16)
	h := sha3.NewShake128()
	h.Write([]byte(msg))
	h.Read(hash)
	b58 := new(base58.Base58)
	return b58.Encode(hash)
	
}

func HexToBytes(s string) ([]byte, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}

func getBlockCypherAPI(networkName string) gobcy.API {
	//explicitly
	bc := gobcy.API{}
	bc.Token = "029727206f7e4c8fb19301e4629c5793"
	bc.Coin = "btc"        //options: "btc","bcy","ltc","doge"
	bc.Chain = networkName //depending on coin: "main","test3","test"
	return bc
}

func buildBTCBlockFromCypher(networkName string, blkHeight int) (*btcutil.Block, error) {
	bc := getBlockCypherAPI(networkName)
	cypherBlock, err := bc.GetBlock(blkHeight, "", nil)
	if err != nil {
		return nil, err
	}
	prevBlkHash, _ := chainhash.NewHashFromStr(cypherBlock.PrevBlock)
	merkleRoot, _ := chainhash.NewHashFromStr(cypherBlock.MerkleRoot)
	msgBlk := wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    int32(cypherBlock.Ver),
			PrevBlock:  *prevBlkHash,
			MerkleRoot: *merkleRoot,
			Timestamp:  cypherBlock.Time,
			Bits:       uint32(cypherBlock.Bits),
			Nonce:      uint32(cypherBlock.Nonce),
		},
		Transactions: []*wire.MsgTx{},
	}
	blk := btcutil.NewBlock(&msgBlk)
	blk.SetHeight(int32(blkHeight))
	return blk, nil
}

func buildMsgTxFromCypher(txID string, networkName string) *wire.MsgTx {
	bc := getBlockCypherAPI(networkName)
	cypherTx, _ := bc.GetTX(txID, nil)

	txIns := []*wire.TxIn{}
	for _, cypherTxIn := range cypherTx.Inputs {
		prevHash, _ := chainhash.NewHashFromStr(cypherTxIn.PrevHash)

		// scriptBytes, _ := HexToBytes("76a914b8268ce4d481413c4e848ff353cd16104291c45b88ac")
		// scriptClass, addrs, _, _ := txscript.ExtractPkScriptAddrs(scriptBytes, &chaincfg.MainNetParams)
		// fmt.Println("scriptClass: ", scriptClass)
		// fmt.Println("addrs: ", addrs[0].String())
		// fmt.Println("addrs: ", addrs[0].EncodeAddress())

		// scriptBytes2, _ := HexToBytes("6a13636861726c6579206c6f766573206865696469")
		// scriptClass2, _, _, _ := txscript.ExtractPkScriptAddrs(scriptBytes2, &chaincfg.MainNetParams)
		// fmt.Println("scriptClass 2: ", scriptClass2)
		// fmt.Println("msg: ", string(scriptBytes2[2:]))

		signatureScript, _ := HexToBytes(cypherTxIn.Script)
		in := &wire.TxIn{
			PreviousOutPoint: wire.OutPoint{
				Hash:  *prevHash,
				Index: uint32(cypherTxIn.OutputIndex),
			},
			SignatureScript: signatureScript,
			Sequence:        uint32(cypherTxIn.Sequence),
		}
		txIns = append(txIns, in)
	}

	txOuts := []*wire.TxOut{}
	for _, cypherTxOut := range cypherTx.Outputs {
		pkScrtip, _ := HexToBytes(cypherTxOut.Script)
		out := &wire.TxOut{
			Value:    int64(cypherTxOut.Value),
			PkScript: pkScrtip,
		}
		txOuts = append(txOuts, out)
	}

	msgTx := wire.MsgTx{
		Version:  int32(cypherTx.Ver),
		TxIn:     txIns,
		TxOut:    txOuts,
		LockTime: uint32(cypherTx.LockTime),
	}
	return &msgTx
}
