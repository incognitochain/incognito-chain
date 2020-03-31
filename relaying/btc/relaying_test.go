package btcrelaying

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func setGenesisBlockToChainParams(networkName string, genesisBlkHeight int) (*chaincfg.Params, error) {
	blk, err := buildBTCBlockFromCypher(networkName, genesisBlkHeight)
	if err != nil {
		return nil, err
	}

	chainParams := chaincfg.MainNetParams
	chainParams.GenesisBlock = blk.MsgBlock()
	chainParams.GenesisHash = blk.Hash()
	return &chainParams, nil
}

func tearDownRelayBTCHeadersTest(dbName string) {
	fmt.Println("Tearing down RelayBTCHeadersTest...")
	dbPath := filepath.Join(testDbRoot, dbName)
	os.RemoveAll(dbPath)
	os.RemoveAll(testDbRoot)
}

func getAllTxsFromCypherBlock(blockHeight int) (string, []string, error) {
	bc := getBlockCypherAPI("main")
	cypherBlock1, err := bc.GetBlock(
		blockHeight,
		"",
		map[string]string{
			"txstart": "0",
			"limit":   "500",
		},
	)
	if err != nil {
		return "", []string{}, err
	}
	cypherBlock2, err := bc.GetBlock(
		blockHeight,
		"",
		map[string]string{
			"txstart": "500",
			"limit":   "1000",
		},
	)
	if err != nil {
		return "", []string{}, err
	}
	txIDs := append(cypherBlock1.TXids, cypherBlock2.TXids...)
	return cypherBlock2.Hash, txIDs, nil
}

func TestRelayBTCHeaders(t *testing.T) {
	networkName := "main"
	genesisBlockHeight := int(308568)

	chainParams, err := setGenesisBlockToChainParams(networkName, genesisBlockHeight)
	if err != nil {
		t.Errorf("Could not set genesis block to chain params with err: %v", err)
		return
	}
	dbName := "btc-blocks-test"
	btcChain1, err := GetChain(dbName, chainParams)
	defer tearDownRelayBTCHeadersTest(dbName)
	if err != nil {
		t.Errorf("Could not get chain instance with err: %v", err)
		return
	}

	for i := genesisBlockHeight + 1; i <= genesisBlockHeight+10; i++ {
		blk, err := buildBTCBlockFromCypher(networkName, i)
		if err != nil {
			t.Errorf("buildBTCBlockFromCypher fail on block %v: %v\n", i, err)
			return
		}
		isMainChain, isOrphan, err := btcChain1.ProcessBlockV2(blk, BFNone)
		if err != nil {
			t.Errorf("ProcessBlock fail on block %v: %v\n", i, err)
			return
		}
		if isOrphan {
			t.Errorf("ProcessBlock incorrectly returned block %v "+
				"is an orphan\n", i)
			return
		}
		fmt.Printf("Block %s (%d) is on main chain: %t\n", blk.Hash(), blk.Height(), isMainChain)
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("Session 1: best block hash %s and block height %d\n", btcChain1.BestSnapshot().Hash.String(), btcChain1.BestSnapshot().Height)
	btcChain1.db.Close()

	// simulate new session
	btcChain2, err := GetChain(dbName, chainParams)
	if err != nil {
		t.Errorf("Could not get chain instance (for session 2) with err: %v", err)
		return
	}
	fmt.Printf("Session 2: best block hash %s and block height %d\n", btcChain2.BestSnapshot().Hash.String(), btcChain2.BestSnapshot().Height)

	if btcChain2.BestSnapshot().Hash != btcChain1.BestSnapshot().Hash ||
		btcChain2.BestSnapshot().Height != btcChain1.BestSnapshot().Height {
		t.Errorf("Best states of session 1 & 2 are different")
		return
	}

	txID := "8bae12b5f4c088d940733dcd1455efc6a3a69cf9340e17a981286d3778615684"
	msgTx := buildMsgTxFromCypher(txID)

	blockHash, txIDs, err := getAllTxsFromCypherBlock(308570)
	if err != nil {
		t.Errorf("Could not get cypher block by height with err: %v", err)
		return
	}
	txHashes := make([]*chainhash.Hash, len(txIDs))
	for i := 0; i < len(txIDs); i++ {
		txHashes[i], _ = chainhash.NewHashFromStr(txIDs[i])
	}

	txHash := msgTx.TxHash()
	blkHash, _ := chainhash.NewHashFromStr(blockHash)
	merkleProofs := buildMerkleProof(txHashes, &txHash)
	btcProof := BTCProof{
		MerkleProofs: merkleProofs,
		BTCTx:        msgTx,
		BlockHash:    blkHash,
	}
	btcProofBytes, _ := json.Marshal(btcProof)
	btcProofStr := base64.StdEncoding.EncodeToString(btcProofBytes)
	decodedProof, err := ParseBTCProofFromB64EncodeStr(btcProofStr)
	if err != nil {
		t.Errorf("Could not parse btc proof from base64 string with err: %v", err)
		return
	}

	isValid, err := btcChain2.VerifyTxWithMerkleProofs(decodedProof)
	if err != nil {
		t.Errorf("Could not verify tx with merkle proofs with err: %v", err)
		return
	}
	if !isValid {
		t.Error("Failed to verify tx with merkle proofs")
		return
	}
	msg, err := ExtractAttachedMsgFromTx(decodedProof.BTCTx)
	if err != nil {
		t.Errorf("Could not extract attached message from tx with err: %v", err)
		return
	}
	if msg != "charley loves heidi" {
		t.Errorf("Expect attached message is %s but got %s", "charley loves heidi", msg)
		return
	}

	addrStr, err := btcChain2.ExtractPaymentAddrStrFromPkScript(decodedProof.BTCTx.TxOut[1].PkScript)
	if err != nil {
		t.Errorf("Could not extract payment address from tx with err: %v", err)
		return
	}
	if addrStr != "1HnhWpkMHMjgt167kvgcPyurMmsCQ2WPgg" {
		t.Errorf("Expect payment address is %s but got %s", "1HnhWpkMHMjgt167kvgcPyurMmsCQ2WPgg", addrStr)
		return
	}
}

func TestBuildBTCBlockFromCypher(t *testing.T) {
	blk, err := buildBTCBlockFromCypher("main", 623600)
	// blk, err := buildBTCBlockFromCypher("test3", 1692037)
	if err != nil {
		t.Errorf("Could not build btc block from cypher - with err: %v", err)
		return
	}
	unixTs := blk.MsgBlock().Header.Timestamp.Unix()
	if unixTs != 1585564707 {
		t.Errorf("Wrong timestamp: expected %d, got %d", 1585564707, unixTs)
		return
	}
	ts := time.Unix(unixTs, 0)
	if ts.UnixNano() != blk.MsgBlock().Header.Timestamp.UnixNano() {
		t.Error("Convertion from unix timestamp to Time is not correct")
	}
}
