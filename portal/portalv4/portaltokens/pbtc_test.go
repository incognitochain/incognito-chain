package portaltokens

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/incognitochain/incognito-chain/wallet"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func insertUnshieldIDIntoStateDB(waitingUnshieldState map[string]*statedb.WaitingUnshieldRequest,
	tokenID string, remoteAddress string, unshieldID string, amount uint64, beaconHeight uint64) {
	key := statedb.GenerateWaitingUnshieldRequestObjectKey(tokenID, unshieldID).String()
	waitingUnshieldState[key] = statedb.NewWaitingUnshieldRequestStateWithValue(remoteAddress, amount, unshieldID, beaconHeight)
}

func insertUTXOIntoStateDB(utxos map[string]*statedb.UTXO, key string, amount uint64) {
	curUTXO := &statedb.UTXO{}
	curUTXO.SetOutputAmount(amount)
	utxos[key] = curUTXO
}

func printBroadcastTxs(t *testing.T, broadcastTxs []*BroadcastTx) {
	t.Logf("Len of broadcast txs: %v\n", len(broadcastTxs))
	for i, tx := range broadcastTxs {
		t.Logf("+ Broadcast Tx %v\n", i)
		for idx, utxo := range tx.UTXOs {
			t.Logf("++ UTXO %v: %v\n", idx, utxo.GetOutputAmount())
		}
		t.Logf("+ Unshield IDs: %v \n", tx.UnshieldIDs)
	}
}

func TestChooseUnshieldIDsFromCandidates(t *testing.T) {
	p := &PortalBTCTokenProcessor{}

	tokenID := "btc"
	waitingUnshieldState := map[string]*statedb.WaitingUnshieldRequest{}
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_1", "unshield_1", 10000, 1) // pBTC
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_2", "unshield_2", 5000, 2)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_3", "unshield_3", 20000, 3)

	// Not enough UTXO
	utxos := map[string]*statedb.UTXO{}
	insertUTXOIntoStateDB(utxos, "utxo_1", 900) // BTC

	tinyAmount := uint64(100000)

	broadcastTxs := p.ChooseUnshieldIDsFromCandidates(utxos, waitingUnshieldState, tinyAmount)
	printBroadcastTxs(t, broadcastTxs)

	// Broadcast a part of unshield requests
	utxos = map[string]*statedb.UTXO{}
	insertUTXOIntoStateDB(utxos, "utxo_2", 1500)

	broadcastTxs = p.ChooseUnshieldIDsFromCandidates(utxos, waitingUnshieldState, tinyAmount)
	printBroadcastTxs(t, broadcastTxs)

	// Broadcast all unshield requests
	utxos = map[string]*statedb.UTXO{}
	insertUTXOIntoStateDB(utxos, "utxo_3", 5000)

	broadcastTxs = p.ChooseUnshieldIDsFromCandidates(utxos, waitingUnshieldState, tinyAmount)
	printBroadcastTxs(t, broadcastTxs)

	// First unshield request need multiple UTXOs
	waitingUnshieldState = map[string]*statedb.WaitingUnshieldRequest{}
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_4", "unshield_4", 20000, 4)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_5", "unshield_5", 10000, 5)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_6", "unshield_6", 15000, 6)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_7", "unshield_7", 100000, 7)

	utxos = map[string]*statedb.UTXO{}
	insertUTXOIntoStateDB(utxos, "utxo_4", 500)
	insertUTXOIntoStateDB(utxos, "utxo_5", 1600)
	insertUTXOIntoStateDB(utxos, "utxo_6", 1000)

	broadcastTxs = p.ChooseUnshieldIDsFromCandidates(utxos, waitingUnshieldState, tinyAmount)
	printBroadcastTxs(t, broadcastTxs)

	// Broadcast multiple txs
	waitingUnshieldState = map[string]*statedb.WaitingUnshieldRequest{}
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_8", "unshield_8", 20000, 8)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_9", "unshield_9", 10000, 9)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_10", "unshield_10", 2000, 10)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_11", "unshield_11", 1000, 11)

	utxos = map[string]*statedb.UTXO{}
	insertUTXOIntoStateDB(utxos, "utxo_7", 150)
	insertUTXOIntoStateDB(utxos, "utxo_8", 150)
	insertUTXOIntoStateDB(utxos, "utxo_9", 1000)
	insertUTXOIntoStateDB(utxos, "utxo_10", 1600)
	insertUTXOIntoStateDB(utxos, "utxo_11", 1000)

	broadcastTxs = p.ChooseUnshieldIDsFromCandidates(utxos, waitingUnshieldState, tinyAmount)
	printBroadcastTxs(t, broadcastTxs)

	// Broadcast multiple txs
	waitingUnshieldState = map[string]*statedb.WaitingUnshieldRequest{}
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_8", "unshield_12", 10000, 8)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_9", "unshield_13", 10005, 9)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_10", "unshield_14", 10000, 10)

	utxos = map[string]*statedb.UTXO{}
	insertUTXOIntoStateDB(utxos, "utxo_12", 10)
	insertUTXOIntoStateDB(utxos, "utxo_13", 15)
	insertUTXOIntoStateDB(utxos, "utxo_14", 1000)
	insertUTXOIntoStateDB(utxos, "utxo_15", 20)
	insertUTXOIntoStateDB(utxos, "utxo_16", 1000)
	insertUTXOIntoStateDB(utxos, "utxo_17", 1000)
	insertUTXOIntoStateDB(utxos, "utxo_18", 500)
	insertUTXOIntoStateDB(utxos, "utxo_19", 500)

	broadcastTxs = p.ChooseUnshieldIDsFromCandidates(utxos, waitingUnshieldState, tinyAmount)
	printBroadcastTxs(t, broadcastTxs)
}

// func getBlockCypherAPI(networkName string) gobcy.API {
// 	//explicitly
// 	bc := gobcy.API{}
// 	bc.Token = "a8ed119b4edf4f609a83bd3fbe9a3831"
// 	bc.Coin = "btc"        //options: "btc","bcy","ltc","doge"
// 	bc.Chain = networkName //depending on coin: "main","test3","test"
// 	return bc
// }

// func buildBTCBlockFromCypher(networkName string, blkHeight int) (*btcutil.Block, error) {
// 	bc := getBlockCypherAPI(networkName)
// 	cypherBlock, err := bc.GetBlock(blkHeight, "", nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	prevBlkHash, _ := chainhash.NewHashFromStr(cypherBlock.PrevBlock)
// 	merkleRoot, _ := chainhash.NewHashFromStr(cypherBlock.MerkleRoot)
// 	msgBlk := wire.MsgBlock{
// 		Header: wire.BlockHeader{
// 			Version:    int32(cypherBlock.Ver),
// 			PrevBlock:  *prevBlkHash,
// 			MerkleRoot: *merkleRoot,
// 			Timestamp:  cypherBlock.Time,
// 			Bits:       uint32(cypherBlock.Bits),
// 			Nonce:      uint32(cypherBlock.Nonce),
// 		},
// 		Transactions: []*wire.MsgTx{},
// 	}
// 	blk := btcutil.NewBlock(&msgBlk)
// 	blk.SetHeight(int32(blkHeight))
// 	return blk, nil
// }

// func setGenesisBlockToChainParams(networkName string, genesisBlkHeight int) (*chaincfg.Params, error) {
// 	blk, err := buildBTCBlockFromCypher(networkName, genesisBlkHeight)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// chainParams := chaincfg.MainNetParams
// 	chainParams := chaincfg.TestNet3Params
// 	chainParams.GenesisBlock = blk.MsgBlock()
// 	chainParams.GenesisHash = blk.Hash()
// 	return &chainParams, nil
// }

// func tearDownRelayBTCHeadersTest(dbName string, t *testing.T) {
// 	t.Logf("Tearing down RelayBTCHeadersTest...")
// 	os.RemoveAll(dbName)
// }

// func printExtractedUTXOs(isValid bool, utxos []*statedb.UTXO, t *testing.T) {
// 	t.Logf("Is Valid Proof: %v\n", isValid)
// 	for idx, utxo := range utxos {
// 		t.Logf("++ UTXO %v: %v\n", idx, utxo.GetOutputAmount())
// 	}
// }

// func testProofFromHeight(proof string, height int, t *testing.T) {
// 	expectedMultisigAddress := "2MvpFqydTR43TT4emMD84Mzhgd8F6dCow1X"
// 	networkName := "test3"
// 	p := &PortalBTCTokenProcessor{}
// 	p.PortalToken = &PortalToken{}
// 	shieldingIncAddress := "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci"
// 	expectedMemo := p.GetExpectedMemoForShielding(shieldingIncAddress)

// 	genesisBlockHeight := height
// 	chainParams, err := setGenesisBlockToChainParams(networkName, genesisBlockHeight)
// 	if err != nil {
// 		t.Errorf("Could not set genesis block to chain params with err: %v", err)
// 		return
// 	}

// 	dbName := "btc-blocks-test"
// 	btcChain, err := btcrelaying.GetChainV2(dbName, chainParams, int32(genesisBlockHeight))
// 	defer tearDownRelayBTCHeadersTest(dbName, t)
// 	if err != nil {
// 		t.Errorf("Could not get chain instance with err: %v", err)
// 		return
// 	}

// 	for i := genesisBlockHeight + 1; i <= genesisBlockHeight+10; i++ {
// 		blk, err := buildBTCBlockFromCypher(networkName, i)
// 		if err != nil {
// 			t.Errorf("buildBTCBlockFromCypher fail on block %v: %v\n", i, err)
// 			return
// 		}
// 		isMainChain, isOrphan, err := btcChain.ProcessBlockV2(blk, 0)
// 		if err != nil {
// 			t.Errorf("ProcessBlock fail on block %v: %v\n", i, err)
// 			return
// 		}
// 		if isOrphan {
// 			t.Errorf("ProcessBlock incorrectly returned block %v "+
// 				"is an orphan\n", i)
// 			return
// 		}
// 		t.Logf("Block %s (%d) is on main chain: %t\n", blk.Hash(), blk.Height(), isMainChain)
// 		time.Sleep(1 * time.Second)
// 	}

// 	t.Logf("Session: best block hash %s and block height %d\n", btcChain.BestSnapshot().Hash.String(), btcChain.BestSnapshot().Height)

// 	isValid, utxos, err := p.parseAndVerifyProofBTCChain(proof, btcChain, expectedMemo, expectedMultisigAddress)

// 	if err != nil {
// 		t.Errorf("Parse proof error %v", err)
// 		return
// 	}

// 	printExtractedUTXOs(isValid, utxos, t)

// 	btcChain.GetDB().Close()
// }

// func TestParseAndVerifyProof(t *testing.T) {
// 	// Comment all lines of code that write logs
// 	type TestCaseProof struct {
// 		height int
// 		proof  string
// 	}

// 	cases := []*TestCaseProof{
// 		//valid
// 		{
// 			height: 1939008,
// 			proof:  "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE1MSwxNCw4NSwyMTUsMTI3LDIzOCwxMDcsMTE3LDE5NCw5MSwxNCwyMDUsMzEsMTM3LDE2MCw0MSwxOTQsMTgyLDg2LDIsMjQ3LDc0LDcyLDIzNywxMjQsMTE0LDE3MiwxNDQsMTQsNzMsMjIxLDEzMF0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMjM4LDczLDY5LDQ3LDQ4LDE3NSwxNzQsMjA0LDcxLDI1NSwyNTAsMTgsODcsOTcsNDUsNzMsMjAxLDE5NywxMTUsMTM0LDIyOSw1OCwyNDEsMTcyLDI1MSwxMDUsOTYsMTg2LDMyLDIzMiwxNjQsMjRdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjAwLDEwOCwxNiwyNTUsMjUzLDExMSwxNzQsMTM1LDIwLDEzNSwxODYsMTg2LDc4LDE2OSwxMzIsMjAsMTI3LDI1NSwxODksMTkxLDIxNywxNDgsMTkwLDE4LDYyLDE4OSwxOSwyMiw5MCwzLDI1NSwxNV0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOls1OSw2MiwyNywwLDE4MywxNTYsNTgsMTcwLDE4MSwxMTEsMjYsMTc1LDU0LDI1LDIyNSwyMzksMTQwLDEzOSwxMjUsMTE0LDUwLDgxLDk3LDcyLDYxLDE5MCwxNDYsMjMwLDE2NCw0NywyMzYsMTA5XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls0OSwyNSw3MywyMDUsNjUsMTQ4LDEwNCwxMjgsMTkxLDkyLDMsNjIsMTY4LDMyLDE2NiwxNTYsMTYxLDIxNiwxMDksMTMyLDI0NSw1MCwxNTEsMjA5LDM2LDE3OSwyNCw0NiwxNjMsMTkwLDIzMywyMTBdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMiwxNDMsNTEsMjMzLDM4LDIyMSwxNDMsNDMsMTM0LDExNSwxNTgsMTI3LDkxLDE1MCw4Myw5NywzMiwxODksMjU0LDEzNiwxOTQsMjEwLDE1NywyMjksNzUsNDQsMTAyLDE1NiwyMjcsMTQ0LDEwOCwyM10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTU4LDIzOCwxMDIsMyw2NywxODEsMjMsMTYwLDE1MCw0MywxOTYsMjMsMTkzLDk2LDM0LDEwNCwxNjEsMzQsMTc0LDE1MCw1MSwxNzgsNjEsMzIsMTcsNTgsNCw2NCw5MywyMDksMjAsMjI5XSwiSXNMZWZ0IjpmYWxzZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoxLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOlsxNzIsMjI0LDEwLDIwNCwyMTgsMTUyLDE3Miw5NCwyMTUsMjE4LDcxLDE3MSwxLDEyNyw4OSwyNywxMzQsNjQsMTQ4LDEsMjAyLDIxNSwyMTIsMjA4LDE2OCwxMDUsMTI5LDI1NSwxMzQsMTg0LDIyMywxNThdLCJJbmRleCI6Mn0sIlNpZ25hdHVyZVNjcmlwdCI6IlNEQkZBaUVBalJDS0p1RzMvNlErdVFDaTVPL1MwRVA2VkZOT3hrU25yQTN3RnZ3ZVhKb0NJSFpKVklQVmRsbE16L0JFRTVBU3BadGQ5NHQ4SWhKY0dJS1FnQUNuQjg2MEFTRUR6eUFUVDFaSTR2YnhkNlpWS3lXNmwrUmJGSVZPVHhNcVN2RCtmYWsveFB3PSIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3Mjk1fV0sIlR4T3V0IjpbeyJWYWx1ZSI6MCwiUGtTY3JpcHQiOiJhbXRRVXpFdE1USlROVXh5Y3pGWVpWRk1ZbkZPTkhsVGVVdDBha0ZxWkRKa04zTkNVREowYWtacGFucHRjRFpoZG5KeWExRkRUa1pOY0d0WWJUTkdVSHBxTWxkamRUSmFUbkZLUlcxb09VcHlhVloxVWtWeVZuZG9kVkZ1VEcxWFUyRm5aMjlpUlZkelFrVmphUT09In0seyJWYWx1ZSI6NDAwLCJQa1NjcmlwdCI6InFSUW5KNmR2OHZvNVhjVWxaS2pyS3F1TS9sSElkb2M9In0seyJWYWx1ZSI6MTk1NzcyLCJQa1NjcmlwdCI6ImRxa1Vndnk2bFFpK0VpUXk5N3VRMmw5MEFVQ21XNGlJckE9PSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOlsxNzQsMTI5LDE0OCwxOTIsNzMsMjMwLDIzMSwxOTIsMTY0LDE4MSwxOTAsMjQ5LDI1MywyNTUsMjU1LDIyNCwzNywxNTQsMzQsMjUxLDkyLDU2LDEzOCw1NCwxMywwLDAsMCwwLDAsMCwwXX0=",
// 		},

// 		// merkle root is not valid
// 		{
// 			height: 1937579,
// 			proof:  "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE1MSwxNCw4NSwyMTUsMTI3LDIzOCwxMDcsMTE3LDE5NCw5MSwxNCwyMDUsMzEsMTM3LDE2MCw0MSwxOTQsMTgyLDg2LDIsMjQ3LDc0LDcyLDIzNywxMjQsMTE0LDE3MiwxNDQsMTQsNzMsMjIxLDEzMF0sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMjM4LDczLDY5LDQ3LDQ4LDE3NSwxNzQsMjA0LDcxLDI1NSwyNTAsMTgsODcsOTcsNDUsNzMsMjAxLDE5NywxMTUsMTM0LDIyOSw1OCwyNDEsMTcyLDI1MSwxMDUsOTYsMTg2LDMyLDIzMiwxNjQsMjRdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMjAwLDEwOCwxNiwyNTUsMjUzLDExMSwxNzQsMTM1LDIwLDEzNSwxODYsMTg2LDc4LDE2OSwxMzIsMjAsMTI3LDI1NSwxODksMTkxLDIxNywxNDgsMTkwLDE4LDYyLDE4OSwxOSwyMiw5MCwzLDI1NSwxNV0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOls1OSw2MiwyNywwLDE4MywxNTYsNTgsMTcwLDE4MSwxMTEsMjYsMTc1LDU0LDI1LDIyNSwyMzksMTQwLDEzOSwxMjUsMTE0LDUwLDgxLDk3LDcyLDYxLDE5MCwxNDYsMjMwLDE2NCw0NywyMzYsMTA5XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls0OSwyNSw3MywyMDUsNjUsMTQ4LDEwNCwxMjgsMTkxLDkyLDMsNjIsMTY4LDMyLDE2NiwxNTYsMTYxLDIxNiwxMDksMTMyLDI0NSw1MCwxNTEsMjA5LDM2LDE3OSwyNCw0NiwxNjMsMTkwLDIzMywyMTBdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMiwxNDMsNTEsMjMzLDM4LDIyMSwxNDMsNDMsMTM0LDExNSwxNTgsMTI3LDkxLDE1MCw4Myw5NywzMiwxODksMjU0LDEzNiwxOTQsMjEwLDE1NywyMjksNzUsNDQsMTAyLDE1NiwyMjcsMTQ0LDEwOCwyM10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTU4LDIzOCwxMDIsMyw2NywxODEsMjMsMTYwLDE1MCw0MywxOTYsMjMsMTkzLDk2LDM0LDEwNCwxNjEsMzQsMTc0LDE1MCw1MSwxNzgsNjEsMzIsMTcsNTgsNCw2NCw5MywyMDksMjAsMjI5XSwiSXNMZWZ0IjpmYWxzZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoxLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOlsxNzIsMjI0LDEwLDIwNCwyMTgsMTUyLDE3Miw5NCwyMTUsMjE4LDcxLDE3MSwxLDEyNyw4OSwyNywxMzQsNjQsMTQ4LDEsMjAyLDIxNSwyMTIsMjA4LDE2OCwxMDUsMTI5LDI1NSwxMzQsMTg0LDIyMywxNThdLCJJbmRleCI6Mn0sIlNpZ25hdHVyZVNjcmlwdCI6IlNEQkZBaUVBalJDS0p1RzMvNlErdVFDaTVPL1MwRVA2VkZOT3hrU25yQTN3RnZ3ZVhKb0NJSFpKVklQVmRsbE16L0JFRTVBU3BadGQ5NHQ4SWhKY0dJS1FnQUNuQjg2MEFTRUR6eUFUVDFaSTR2YnhkNlpWS3lXNmwrUmJGSVZPVHhNcVN2RCtmYWsveFB3PSIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3Mjk1fV0sIlR4T3V0IjpbeyJWYWx1ZSI6MCwiUGtTY3JpcHQiOiJhbXRRVXpFdE1USlROVXh5Y3pGWVpWRk1ZbkZPTkhsVGVVdDBha0ZxWkRKa04zTkNVREowYWtacGFucHRjRFpoZG5KeWExRkRUa1pOY0d0WWJUTkdVSHBxTWxkamRUSmFUbkZLUlcxb09VcHlhVloxVWtWeVZuZG9kVkZ1VEcxWFUyRm5aMjlpUlZkelFrVmphUT09In0seyJWYWx1ZSI6NDAwLCJQa1NjcmlwdCI6InFSUW5KNmR2OHZvNVhjVWxaS2pyS3F1TS9sSElkb2M9In0seyJWYWx1ZSI6MTk1NzcyLCJQa1NjcmlwdCI6ImRxa1Vndnk2bFFpK0VpUXk5N3VRMmw5MEFVQ21XNGlJckE9PSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOlsxNzQsMTI5LDE0OCwxOTIsNzMsMjMwLDIzMSwxOTIsMTY0LDE4MSwxOTAsMjQ5LDI1MywyNTUsMjU1LDIyNCwzNywxNTQsMzQsMjUxLDkyLDU2LDEzOCw1NCwxMywwLDAsMCwwLDAsMCwwXX0=",
// 		},

// 		// valid
// 		{
// 			height: 1939010,
// 			proof:  "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzIzMiwyOSw4Nyw0Nyw1NSwyMDUsMTk1LDYyLDM1LDgwLDMwLDE0MCwxMCwyMTIsODAsMTc0LDEyMCwxNzEsNjgsMjAsMTQ2LDIwMiwyMTMsMjUsMTIxLDIxNiwyMzUsMjM2LDksMTM0LDIzNywyMTVdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzkyLDMwLDYsOTcsMTk1LDYzLDk5LDEsMzMsMjI3LDQyLDIxOCwxNzksMTM1LDE3OSwxNTcsOTAsMjM3LDIwNiwyMTYsMjI0LDI4LDI5LDEzMyw2OCwyMzksMzEsMjAsMTYxLDIwNSwyMzksMTcyXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOlsyMjAsMzUsOTksMjI0LDE2MCw3MiwxMTEsMjEzLDE0OCwxOTksMTk5LDM4LDIzNSwxOTIsNzksNTksMTQwLDEsMzgsMzQsMTg0LDExOCwxODEsNTQsODYsOTcsNDAsODgsMjQ4LDI0MSwxOTAsMTU2XSwiSXNMZWZ0Ijp0cnVlfSx7IlByb29mSGFzaCI6WzU3LDE2OSwxNjAsMTcsMTMyLDE4OSw0MCw5Miw1MSwyMzksNjgsMTk2LDI1MywyMCwxNTgsODIsMTgxLDc3LDIwLDEwNiwxNzUsMjIyLDEzLDExOCw5MiwxNTMsMjM5LDE5MCw1NywyMTMsMTY3LDk3XSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOlsyNDUsMTIyLDE4NSwxMTQsNDcsNzUsMjI2LDExNCwxNTcsMTcxLDIwOSwyNDQsOTMsMTUzLDIzNiw5Myw2MSwwLDE5NSwxMjYsMTY0LDc0LDEzNywxODEsNjYsMTI1LDEyLDI0MywyNDEsMjMsMjA5LDEyXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls2MywxMzgsOTEsMTIzLDY0LDksMjI5LDIxLDE3NCwxOTksMSwyMTEsNiw2OCwxNzYsMjI3LDQ1LDIzMyw3MCwxMDMsNDgsOTQsNDEsMTQ5LDE1NSw0NCwyNDAsMTc1LDExNiw1OCw5NiwxNDRdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMTU4LDIzOCwxMDIsMyw2NywxODEsMjMsMTYwLDE1MCw0MywxOTYsMjMsMTkzLDk2LDM0LDEwNCwxNjEsMzQsMTc0LDE1MCw1MSwxNzgsNjEsMzIsMTcsNTgsNCw2NCw5MywyMDksMjAsMjI5XSwiSXNMZWZ0IjpmYWxzZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoxLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOls4NSwyMjcsMjMsMTUsMjA5LDEzNiwyMTksMTEsOTEsMjIyLDI0LDE3NiwxMzIsMjMxLDE4OCwyNDAsMTgzLDI4LDI0OSwxODUsNTcsMTk5LDE5OCw4NCw4Miw3Miw1NiwxNTMsNjQsMzQsMzEsMzddLCJJbmRleCI6Mn0sIlNpZ25hdHVyZVNjcmlwdCI6IlNEQkZBaUVBMVd5OHVRellVOUpqV3hDZXVUb0lGNERvT2RoRUFjdmNiVmwrblQ1ck9qWUNJQk9aU2VaYnVvT0FoS3dzNE1yMDMrMDNBaEhXZlhTMkVNZ2w5VWhaOHZHQUFTRUR6eUFUVDFaSTR2YnhkNlpWS3lXNmwrUmJGSVZPVHhNcVN2RCtmYWsveFB3PSIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3Mjk1fV0sIlR4T3V0IjpbeyJWYWx1ZSI6MCwiUGtTY3JpcHQiOiJhbXRRVXpFdE1USlROVXh5Y3pGWVpWRk1ZbkZPTkhsVGVVdDBha0ZxWkRKa04zTkNVREowYWtacGFucHRjRFpoZG5KeWExRkRUa1pOY0d0WWJUTkdVSHBxTWxkamRUSmFUbkZLUlcxb09VcHlhVloxVWtWeVZuZG9kVkZ1VEcxWFUyRm5aMjlpUlZkelFrVmphUT09In0seyJWYWx1ZSI6ODAwLCJQa1NjcmlwdCI6InFSUW5KNmR2OHZvNVhjVWxaS2pyS3F1TS9sSElkb2M9In0seyJWYWx1ZSI6MTk0NDcyLCJQa1NjcmlwdCI6ImRxa1Vndnk2bFFpK0VpUXk5N3VRMmw5MEFVQ21XNGlJckE9PSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOlsxNzQsMTI5LDE0OCwxOTIsNzMsMjMwLDIzMSwxOTIsMTY0LDE4MSwxOTAsMjQ5LDI1MywyNTUsMjU1LDIyNCwzNywxNTQsMzQsMjUxLDkyLDU2LDEzOCw1NCwxMywwLDAsMCwwLDAsMCwwXX0=",
// 		},

// 		// invalid memo
// 		{
// 			height: 1939010,
// 			proof:  "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6Wzk4LDE3LDI1LDQsMTQ0LDIwNSwyMjMsMjI4LDE3OSwyOSwyMzUsMTIwLDkxLDAsNDgsMTA2LDE2MCwxMTcsMTEyLDI0Myw4NywyNDUsMjgsNzgsMjksMTQzLDIzMywyMjksMTQxLDgwLDIxOSw1MV0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsxNTQsMjE0LDIxNCwxMjMsMTU1LDE4MSwxNDMsMTg5LDEsMTM4LDEzOSwyNTAsNDUsMjM3LDE3MSwxOTQsNTgsMjAsNTMsMCwxOSwxNzAsOTYsMTYwLDI0OCwxNTAsMTYsOTgsMjQ5LDAsOTYsNDVdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMTA0LDEwMyw1OSwyMzIsMTM2LDEzOSwyMTEsMjM2LDksMTkzLDE5OCwxNjUsMTM4LDE1NiwyMTEsNDAsNDcsMzEsODMsNjgsMjEyLDc2LDEyOCwxMTcsMTA1LDYxLDUxLDIxNSwxMTEsMTYxLDgwLDE0OF0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOlsyMjksMTI0LDYxLDE1NCw3OCwxNTYsMTcxLDE4NywxOTgsMTkzLDU1LDIzLDE1MywxMjcsMTEzLDE0NywzLDE4LDE4MSwyMTgsMjMwLDI1LDU5LDExMywyMDIsMTQ5LDE2Miw2MywyNDEsMTU5LDE3OSwyNF0sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOls1LDM0LDk5LDExNSw0NSwyNTEsMjI5LDI0NSw2NCwzNCw5MCw0MywxMTAsMiwxOSw5Miw0MCwyNywyMjAsMTE0LDEzOSw4LDEwMSw4NywyMTYsMjI0LDE4Niw0MywxNDcsMjQwLDEwMiw1M10sIklzTGVmdCI6dHJ1ZX0seyJQcm9vZkhhc2giOls0NSwxMzIsMTI4LDIzNSwxMTMsNDksMTM3LDc5LDY2LDYyLDE2MCw2OSwxMjIsNjAsMjUwLDI0NSwxMSwyMDMsMTMsNDksMjQ4LDksMTA2LDEyMCw3NCw0OCw2OSwxODQsMTYzLDE1NCwyMTUsMTQxXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOlsxNjEsMjQ2LDQzLDE4MCwxMTAsMTY3LDIzMSwxNDcsODcsOTEsNzcsNjAsMjE1LDEzMiwxMDksMjE0LDYzLDksMTcsNDYsMywyMzAsMTA1LDE0MCwxOTEsMTAwLDIxMiwxMTUsMTAyLDE3NiwyMzQsMTcwXSwiSXNMZWZ0IjpmYWxzZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoxLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOlsxODMsMTYwLDk0LDE4NiwxMzksMTg0LDI0MywxMTQsMjMyLDEzNSwxMDQsNTMsMTIyLDI0NSwyMDgsMjQxLDQ4LDQxLDYwLDgsMTcyLDIxNCwxNzUsMjMxLDE3MSw4NywxMDMsMTM3LDEyNCwxMDgsNzksMTgwXSwiSW5kZXgiOjJ9LCJTaWduYXR1cmVTY3JpcHQiOiJTREJGQWlFQTl4V0MzdUQyL1FSbWNtSmJDdmpiYlNMZUljWDFnWG9PSmJTeC9LL0dJWEFDSUVVN05MMjAvSmFZRmJtcjlWRG0wTVFkcGNDWVVQanZvTXM1WVRENU9DVE5BU0VEenlBVFQxWkk0dmJ4ZDZaVkt5VzZsK1JiRklWT1R4TXFTdkQrZmFrL3hQdz0iLCJXaXRuZXNzIjpudWxsLCJTZXF1ZW5jZSI6NDI5NDk2NzI5NX1dLCJUeE91dCI6W3siVmFsdWUiOjAsIlBrU2NyaXB0IjoiYW10UVV6SXRNVEpUTlV4eWN6RllaVkZNWW5GT05IbFRlVXQwYWtGcVpESmtOM05DVURKMGFrWnBhbnB0Y0RaaGRuSnlhMUZEVGtaTmNHdFliVE5HVUhwcU1sZGpkVEphVG5GS1JXMW9PVXB5YVZaMVVrVnlWbmRvZFZGdVRHMVhVMkZuWjI5aVJWZHpRa1ZqYVE9PSJ9LHsiVmFsdWUiOjUwMCwiUGtTY3JpcHQiOiJxUlFuSjZkdjh2bzVYY1VsWktqcktxdU0vbEhJZG9jPSJ9LHsiVmFsdWUiOjE5MzQ3MiwiUGtTY3JpcHQiOiJkcWtVZ3Z5NmxRaStFaVF5OTd1UTJsOTBBVUNtVzRpSXJBPT0ifV0sIkxvY2tUaW1lIjowfSwiQmxvY2tIYXNoIjpbMTE4LDIzMCw5MywxODcsMjUwLDgwLDIxMCwxMDgsODQsMTgsMTIsMjE1LDE2NSw4OSwyMjIsMzYsMTgzLDcxLDMsMjQ1LDY3LDEwMCwyNDksOCw2LDAsMCwwLDAsMCwwLDBdfQ==",
// 		},

// 		// invalid proof
// 		{
// 			height: 1939010,
// 			proof:  "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE4Miw3MCwzOSwxOTksNjksMTEzLDEyNiwyMDYsMjQxLDI1LDE1NCwyMzEsNzgsMTY4LDE3MSwxOTgsMjUyLDE5MCwyNTMsMTA3LDE3MywyMjcsOTAsMjMyLDMwLDE3MCwxOTAsMjIzLDE1LDE2MSwyMjMsNjddLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzE3Myw5OCwxOTIsMTcxLDEwMSwxNDIsMTIxLDE5MiwyMDYsOTUsMTA1LDE3MSwxODUsMTQ0LDI0LDExLDIxOSwyMywxODYsMTgwLDI0NiwyMCwxMzMsNzgsMTUsNzIsNzgsMjksOSw5NSwxNTUsMjUyXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls5MywyMTgsMTA3LDE1MywxMCwyMzQsMjUzLDMwLDMzLDExMywxODIsMTE1LDEzLDE3OSwyMjQsMTc3LDYsMTMwLDQ2LDEzNSw3OCwxMzUsMjQyLDI1NCwxOTUsOTAsOTUsMTQxLDk1LDIwNywxOSwyNTRdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEzMiw5OCwxNjMsMjA1LDY4LDExOSwxMzQsMTQsMTkwLDE5Miw4MSw2Niw3MiwxMTIsMzQsMTEyLDIzLDE3MywyMDYsMTM5LDE1MSwyNDUsNTIsNTksNTQsODIsNjcsMTg3LDEzNywyNDksMTA5LDE2M10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTM4LDQwLDE1MCwzMiwxOTAsMTc1LDIxNCwzOCw1MSw3Nyw0MCwyMDAsODksMTEyLDIwNywxNywyMjcsMiw0Niw5OSwyNTAsNzksMzYsMTg2LDEwMCw4OCw0NywxMTgsMTA5LDU1LDEzNCwxNjZdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzcyLDMsMTg1LDIyNCw0OCwxMDgsMjIyLDE5LDYzLDYyLDY0LDgsMTIzLDE5MywxNjksNTAsMSwyMyw4MywxODgsNzMsMTM5LDE1MSwxMzMsMTY0LDE4NywyMTksMzYsMjMwLDgwLDE2MiwxNjRdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMTYxLDI0Niw0MywxODAsMTEwLDE2NywyMzEsMTQ3LDg3LDkxLDc3LDYwLDIxNSwxMzIsMTA5LDIxNCw2Myw5LDE3LDQ2LDMsMjMwLDEwNSwxNDAsMTkxLDEwMCwyMTIsMTE1LDEwMiwxNzYsMjM0LDE3MF0sIklzTGVmdCI6ZmFsc2V9XSwiQlRDVHgiOnsiVmVyc2lvbiI6MSwiVHhJbiI6W3siUHJldmlvdXNPdXRQb2ludCI6eyJIYXNoIjpbMjA1LDIyOSwyNDUsMTc5LDExNCwzMSwxNzIsMTkwLDkzLDEwNiwyMiwxNzMsNDIsMTU5LDE0OSwyMjAsNjEsMyw0NSwxMTUsODQsNDksMzksNTEsMjA0LDIxLDE1MiwxNiwxNzksMTY0LDE3MCwxNDddLCJJbmRleCI6Mn0sIlNpZ25hdHVyZVNjcmlwdCI6IlJ6QkVBaUE0OSs0QUx0dnR3VXViMGgxNmE4STQybW9jaW9teGpXaXBTUzdCdGZpMTVRSWdNTmxGaEMzTlVuUzNsRWFMMm45TmtEbEFIa25DMVhDMmRoT3Y4bThiWjZvQklRUFBJQk5QVmtqaTl2RjNwbFVySmJxWDVGc1VoVTVQRXlwSzhQNTlxVC9FL0E9PSIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3Mjk1fV0sIlR4T3V0IjpbeyJWYWx1ZSI6MCwiUGtTY3JpcHQiOiJhbXRRVXpFdE1USlROVXh5Y3pGWVpWRk1ZbkZOTkhsVGVVdDBha0ZxWkRKa04zTkNVREowYWtacGFucHRjRFpoZG5KeWExRkRUa1pOY0d0WWJUTkdVSHBxTWxkamRUSmFUbkZLUlcxb09VcHlhVloxVWtWeVZuZG9kVkZ1VEcxWFUyRm5aMjlpUlZkelFrVmphUT09In0seyJWYWx1ZSI6NTAwLCJQa1NjcmlwdCI6InFSUW5KNmR2OHZvNVhjVWxaS2pyS3F1TS9sSElkb2M9In0seyJWYWx1ZSI6MTkyNDcyLCJQa1NjcmlwdCI6ImRxa1Vndnk2bFFpK0VpUXk5N3VRMmw5MEFVQ21XNGlJckE9PSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOlsxMTgsMjMwLDkzLDE4NywyNTAsODAsMjEwLDEwOCw4NCwxOCwxMiwyMTUsMTY1LDg5LDIyMiwzNiwxODMsNzEsMywyNDUsNjcsMTAwLDI0OSw4LDYsMCwwLDAsMCwwLDAsMF19==",
// 		},

// 		// invalid memo
// 		{
// 			height: 1939010,
// 			proof:  "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzE4Miw3MCwzOSwxOTksNjksMTEzLDEyNiwyMDYsMjQxLDI1LDE1NCwyMzEsNzgsMTY4LDE3MSwxOTgsMjUyLDE5MCwyNTMsMTA3LDE3MywyMjcsOTAsMjMyLDMwLDE3MCwxOTAsMjIzLDE1LDE2MSwyMjMsNjddLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzE3Myw5OCwxOTIsMTcxLDEwMSwxNDIsMTIxLDE5MiwyMDYsOTUsMTA1LDE3MSwxODUsMTQ0LDI0LDExLDIxOSwyMywxODYsMTgwLDI0NiwyMCwxMzMsNzgsMTUsNzIsNzgsMjksOSw5NSwxNTUsMjUyXSwiSXNMZWZ0IjpmYWxzZX0seyJQcm9vZkhhc2giOls5MywyMTgsMTA3LDE1MywxMCwyMzQsMjUzLDMwLDMzLDExMywxODIsMTE1LDEzLDE3OSwyMjQsMTc3LDYsMTMwLDQ2LDEzNSw3OCwxMzUsMjQyLDI1NCwxOTUsOTAsOTUsMTQxLDk1LDIwNywxOSwyNTRdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzEzMiw5OCwxNjMsMjA1LDY4LDExOSwxMzQsMTQsMTkwLDE5Miw4MSw2Niw3MiwxMTIsMzQsMTEyLDIzLDE3MywyMDYsMTM5LDE1MSwyNDUsNTIsNTksNTQsODIsNjcsMTg3LDEzNywyNDksMTA5LDE2M10sIklzTGVmdCI6ZmFsc2V9LHsiUHJvb2ZIYXNoIjpbMTM4LDQwLDE1MCwzMiwxOTAsMTc1LDIxNCwzOCw1MSw3Nyw0MCwyMDAsODksMTEyLDIwNywxNywyMjcsMiw0Niw5OSwyNTAsNzksMzYsMTg2LDEwMCw4OCw0NywxMTgsMTA5LDU1LDEzNCwxNjZdLCJJc0xlZnQiOmZhbHNlfSx7IlByb29mSGFzaCI6WzcyLDMsMTg1LDIyNCw0OCwxMDgsMjIyLDE5LDYzLDYyLDY0LDgsMTIzLDE5MywxNjksNTAsMSwyMyw4MywxODgsNzMsMTM5LDE1MSwxMzMsMTY0LDE4NywyMTksMzYsMjMwLDgwLDE2MiwxNjRdLCJJc0xlZnQiOnRydWV9LHsiUHJvb2ZIYXNoIjpbMTYxLDI0Niw0MywxODAsMTEwLDE2NywyMzEsMTQ3LDg3LDkxLDc3LDYwLDIxNSwxMzIsMTA5LDIxNCw2Myw5LDE3LDQ2LDMsMjMwLDEwNSwxNDAsMTkxLDEwMCwyMTIsMTE1LDEwMiwxNzYsMjM0LDE3MF0sIklzTGVmdCI6ZmFsc2V9XSwiQlRDVHgiOnsiVmVyc2lvbiI6MSwiVHhJbiI6W3siUHJldmlvdXNPdXRQb2ludCI6eyJIYXNoIjpbMjA1LDIyOSwyNDUsMTc5LDExNCwzMSwxNzIsMTkwLDkzLDEwNiwyMiwxNzMsNDIsMTU5LDE0OSwyMjAsNjEsMyw0NSwxMTUsODQsNDksMzksNTEsMjA0LDIxLDE1MiwxNiwxNzksMTY0LDE3MCwxNDddLCJJbmRleCI6Mn0sIlNpZ25hdHVyZVNjcmlwdCI6IlJ6QkVBaUE0OSs0QUx0dnR3VXViMGgxNmE4STQybW9jaW9teGpXaXBTUzdCdGZpMTVRSWdNTmxGaEMzTlVuUzNsRWFMMm45TmtEbEFIa25DMVhDMmRoT3Y4bThiWjZvQklRUFBJQk5QVmtqaTl2RjNwbFVySmJxWDVGc1VoVTVQRXlwSzhQNTlxVC9FL0E9PSIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3Mjk1fV0sIlR4T3V0IjpbeyJWYWx1ZSI6MCwiUGtTY3JpcHQiOiJhbXRRVXpFdE1USlROVXh5Y3pGWVpWRk1ZbkZOTkhsVGVVdDBha0ZxWkRKa04zTkNVREowYWtacGFucHRjRFpoZG5KeWExRkRUa1pOY0d0WWJUTkdVSHBxTWxkamRUSmFUbkZLUlcxb09VcHlhVloxVWtWeVZuZG9kVkZ1VEcxWFUyRm5aMjlpUlZkelFrVmphUT09In0seyJWYWx1ZSI6NTAwLCJQa1NjcmlwdCI6InFSUW5KNmR2OHZvNVhjVWxaS2pyS3F1TS9sSElkb2M9In0seyJWYWx1ZSI6MTkyNDcyLCJQa1NjcmlwdCI6ImRxa1Vndnk2bFFpK0VpUXk5N3VRMmw5MEFVQ21XNGlJckE9PSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOlsxMTgsMjMwLDkzLDE4NywyNTAsODAsMjEwLDEwOCw4NCwxOCwxMiwyMTUsMTY1LDg5LDIyMiwzNiwxODMsNzEsMywyNDUsNjcsMTAwLDI0OSw4LDYsMCwwLDAsMCwwLDAsMF19",
// 		},
// 	}

// 	runIdx := rand.Intn(len(cases))
// 	fmt.Printf("Running Case #%v\n", runIdx)
// 	testProofFromHeight(cases[runIdx].proof, cases[runIdx].height, t)
// }

// func TestGenerateMultiSigWalletFromSeeds(t *testing.T) {
// 	required := 3

// 	seeds := [][]byte{
// 		[]byte{0xf1, 0x29, 0xb7, 0xa, 0x46, 0xac, 0x35, 0xc4, 0x17, 0x94, 0x10, 0xf3, 0x52, 0xd7, 0xf5, 0x5c, 0xc5, 0x47, 0xe1, 0xa9, 0x26, 0x1f, 0xe8, 0xed, 0xe7, 0x72, 0x34, 0x4, 0x71, 0xeb, 0xc6, 0x9},
// 		[]byte{0xca, 0xa8, 0xaa, 0xdf, 0x1e, 0xdb, 0xc5, 0x72, 0x80, 0x8f, 0x8, 0x65, 0x1d, 0x41, 0x85, 0xde, 0xd1, 0x21, 0x5a, 0xd4, 0x7, 0xe6, 0x3c, 0xb4, 0x6f, 0x11, 0xc5, 0x5, 0xc6, 0x16, 0x7e, 0xfe},
// 		[]byte{0x64, 0x3b, 0x2d, 0xb2, 0x89, 0x5c, 0x53, 0x11, 0x5a, 0xb1, 0x53, 0xd, 0xfd, 0xb3, 0x32, 0xee, 0x1b, 0xe0, 0x7d, 0xcc, 0xd4, 0x3a, 0xd9, 0xf5, 0x62, 0x9b, 0x4c, 0x50, 0x88, 0xa8, 0xad, 0x1a},
// 		[]byte{0x0, 0xa, 0x43, 0x51, 0xdf, 0x7b, 0x2b, 0x86, 0xc3, 0x40, 0x58, 0xe6, 0x42, 0xa6, 0xc2, 0x5d, 0xb6, 0x6c, 0x30, 0x88, 0x8d, 0xb5, 0x8e, 0xe1, 0x44, 0xce, 0xc0, 0x45, 0xc, 0xf5, 0xa0, 0xeb},
// 	}
// 	multiSigScript, privateKeys, multiSigAddr, err := GenerateMultiSigWalletFromSeeds(&chaincfg.TestNet3Params, seeds, required)
// 	fmt.Printf("multiSigScript: hex encode: %v\n", hex.EncodeToString(multiSigScript))
// 	fmt.Printf("privateKeys: %v\n", privateKeys)
// 	fmt.Printf("multiSigAddr: %v\n", multiSigAddr)
// 	fmt.Printf("err: %v\n", err)
// }

func TestGenerateMasterPubKeysFromSeeds(t *testing.T) {
	seeds := [][]byte{
		[]byte{0xf1, 0x29, 0xb7, 0xa, 0x46, 0xac, 0x35, 0xc4, 0x17, 0x94, 0x10, 0xf3, 0x52, 0xd7, 0xf5, 0x5c, 0xc5, 0x47, 0xe1, 0xa9, 0x26, 0x1f, 0xe8, 0xed, 0xe7, 0x72, 0x34, 0x4, 0x71, 0xeb, 0xc6, 0x9},
		[]byte{0xca, 0xa8, 0xaa, 0xdf, 0x1e, 0xdb, 0xc5, 0x72, 0x80, 0x8f, 0x8, 0x65, 0x1d, 0x41, 0x85, 0xde, 0xd1, 0x21, 0x5a, 0xd4, 0x7, 0xe6, 0x3c, 0xb4, 0x6f, 0x11, 0xc5, 0x5, 0xc6, 0x16, 0x7e, 0xfe},
		[]byte{0x64, 0x3b, 0x2d, 0xb2, 0x89, 0x5c, 0x53, 0x11, 0x5a, 0xb1, 0x53, 0xd, 0xfd, 0xb3, 0x32, 0xee, 0x1b, 0xe0, 0x7d, 0xcc, 0xd4, 0x3a, 0xd9, 0xf5, 0x62, 0x9b, 0x4c, 0x50, 0x88, 0xa8, 0xad, 0x1a},
		[]byte{0x0, 0xa, 0x43, 0x51, 0xdf, 0x7b, 0x2b, 0x86, 0xc3, 0x40, 0x58, 0xe6, 0x42, 0xa6, 0xc2, 0x5d, 0xb6, 0x6c, 0x30, 0x88, 0x8d, 0xb5, 0x8e, 0xe1, 0x44, 0xce, 0xc0, 0x45, 0xc, 0xf5, 0xa0, 0xeb},
	}

	btcToken := PortalBTCTokenProcessor{
		PortalToken: &PortalToken{
			ChainID:             "",
			MinTokenAmount:      0,
			MultipleTokenAmount: 10,
		},
		ChainParam: &chaincfg.TestNet3Params,
	}

	masterPubKeys := [][]byte{}
	fmt.Println("======== List master public keys ========")
	for _, s := range seeds {
		pubKey := btcToken.generatePublicKeyFromSeed(s)
		masterPubKeys = append(masterPubKeys, pubKey)
		fmt.Printf("%#v\n", pubKey)
	}
}

func generateBTCPubKeyFromPrivateKey(privateKey []byte) []byte {
	pkx, pky := btcec.S256().ScalarBaseMult(privateKey)
	pubKey := btcec.PublicKey{Curve: btcec.S256(), X: pkx, Y: pky}
	return pubKey.SerializeCompressed()
}
func TestMultiSigAddressDerivation(t *testing.T) {
	tests := []struct {
		name          string
		net           *chaincfg.Params
		incognitoAddr string
	}{
		{
			name:          "test vector 1 master node private",
			incognitoAddr: "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
			net:           &chaincfg.TestNet3Params,
		},
	}
	for i, test := range tests {
		parentFP := []byte{}

		// generate chainCode from shielding Inc address
		incKey, err := wallet.Base58CheckDeserialize(test.incognitoAddr)
		if err != nil {
			t.Errorf("Deserialize incognitoAddr #%d (%s): unexpected error: %v",
				i, test.name, err)
			continue
		}
		chainCode := chainhash.DoubleHashB(incKey.KeySet.PaymentAddress.Pk)

		// generate BTC master account
		BTCPrivateKeyMaster := chainhash.HashB([]byte("PrivateKeyMiningKey")) // private mining key => private key btc
		BTCPublicKeyMaster := generateBTCPubKeyFromPrivateKey(BTCPrivateKeyMaster)

		// extended private key
		extendedBTCPrivateKey := hdkeychain.NewExtendedKey(test.net.HDPrivateKeyID[:], BTCPrivateKeyMaster, chainCode, parentFP, 0, 0, true)
		// extended public key
		extendedBTCPublicKey := hdkeychain.NewExtendedKey(test.net.HDPublicKeyID[:], BTCPublicKeyMaster, chainCode, parentFP, 0, 0, false)

		// generate child account - it is multisig wallet corresponding user inc address
		childPub, _ := extendedBTCPublicKey.Child(0)
		childPubKeyAddr, _ := childPub.Address(test.net)

		// re-generate private key of child account - used to sign on spent uxto
		childPrv, _ := extendedBTCPrivateKey.Child(0)
		childPrvAddrr, _ := childPrv.Address(test.net)

		fmt.Println(childPubKeyAddr.String())
		fmt.Println(childPrvAddrr.String())
		if childPubKeyAddr.String() != childPrvAddrr.String() {
			fmt.Println("something went wrong")
		}
	}
}

func TestGenerateOTMultisigAddress(t *testing.T) {
	p := &PortalBTCTokenProcessor{
		ChainParam: &chaincfg.TestNet3Params,
	}

	seeds := [][]byte{
		[]byte{0xf1, 0x29, 0xb7, 0xa, 0x46, 0xac, 0x35, 0xc4, 0x17, 0x94, 0x10, 0xf3, 0x52, 0xd7, 0xf5, 0x5c, 0xc5, 0x47, 0xe1, 0xa9, 0x26, 0x1f, 0xe8, 0xed, 0xe7, 0x72, 0x34, 0x4, 0x71, 0xeb, 0xc6, 0x9},
		[]byte{0xca, 0xa8, 0xaa, 0xdf, 0x1e, 0xdb, 0xc5, 0x72, 0x80, 0x8f, 0x8, 0x65, 0x1d, 0x41, 0x85, 0xde, 0xd1, 0x21, 0x5a, 0xd4, 0x7, 0xe6, 0x3c, 0xb4, 0x6f, 0x11, 0xc5, 0x5, 0xc6, 0x16, 0x7e, 0xfe},
		[]byte{0x64, 0x3b, 0x2d, 0xb2, 0x89, 0x5c, 0x53, 0x11, 0x5a, 0xb1, 0x53, 0xd, 0xfd, 0xb3, 0x32, 0xee, 0x1b, 0xe0, 0x7d, 0xcc, 0xd4, 0x3a, 0xd9, 0xf5, 0x62, 0x9b, 0x4c, 0x50, 0x88, 0xa8, 0xad, 0x1a},
		[]byte{0x0, 0xa, 0x43, 0x51, 0xdf, 0x7b, 0x2b, 0x86, 0xc3, 0x40, 0x58, 0xe6, 0x42, 0xa6, 0xc2, 0x5d, 0xb6, 0x6c, 0x30, 0x88, 0x8d, 0xb5, 0x8e, 0xe1, 0x44, 0xce, 0xc0, 0x45, 0xc, 0xf5, 0xa0, 0xeb},
	}

	masterPubKeys := [][]byte{}
	for _, seed := range seeds {
		masterPubKeys = append(masterPubKeys, p.generatePublicKeyFromSeed(seed))
	}
	incAddress := "12RuEdPjq4yxivzm8xPxRVHmkL74t4eAdUKPdKKhMEnpxPH3k8GEyULbwq4hjwHWmHQr7MmGBJsMpdCHsYAqNE18jipWQwciBf9yqvQ"
	script, address, err := p.GenerateOTMultisigAddress(masterPubKeys, 3, incAddress)
	if err != nil {
		t.Logf("Error: %v\n", err)
		t.FailNow()
	}
	t.Logf("P2WSH Bech32 address: %v\n", address)
	t.Logf("P2WSH Bech32 hex encode: %v\n", hex.EncodeToString(script))
}
