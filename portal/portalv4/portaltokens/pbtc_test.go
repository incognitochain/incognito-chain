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
	p := &PortalBTCTokenProcessor{
		PortalToken: &PortalToken{
			ChainID:             "Bitcoin-Testnet",
			MinTokenAmount:      10,
			MultipleTokenAmount: 10,
			ExternalInputSize:   1,
			ExternalOutputSize:  1,
			ExternalTxMaxSize:   6,
		},
		ChainParam: &chaincfg.TestNet3Params,
	}

	tokenID := "btc"
	dustAmount := uint64(500) // satoshi

	waitingUnshieldState := map[string]*statedb.WaitingUnshieldRequest{}
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_1", "unshield_1", 10000, 1) // pBTC
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_2", "unshield_2", 5000, 2)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_3", "unshield_3", 20000, 3)

	// Not enough UTXO
	utxos := map[string]*statedb.UTXO{}
	insertUTXOIntoStateDB(utxos, "utxo_1", 900) // satoshi

	broadcastTxs := p.MatchUTXOsAndUnshieldIDs(utxos, waitingUnshieldState, dustAmount)
	printBroadcastTxs(t, broadcastTxs)

	// Broadcast a part of unshield requests
	utxos = map[string]*statedb.UTXO{}
	insertUTXOIntoStateDB(utxos, "utxo_2", 1500)

	broadcastTxs = p.MatchUTXOsAndUnshieldIDs(utxos, waitingUnshieldState, dustAmount)
	printBroadcastTxs(t, broadcastTxs)

	// Broadcast all unshield requests
	utxos = map[string]*statedb.UTXO{}
	insertUTXOIntoStateDB(utxos, "utxo_3", 5000)

	broadcastTxs = p.MatchUTXOsAndUnshieldIDs(utxos, waitingUnshieldState, dustAmount)
	printBroadcastTxs(t, broadcastTxs)

	// First unshield request need multiple UTXOs + a dust UTXO
	waitingUnshieldState = map[string]*statedb.WaitingUnshieldRequest{}
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_4", "unshield_4", 20000, 4)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_5", "unshield_5", 10000, 5)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_6", "unshield_6", 15000, 6)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_7", "unshield_7", 100000, 7)

	utxos = map[string]*statedb.UTXO{}
	insertUTXOIntoStateDB(utxos, "utxo_4", 500)
	insertUTXOIntoStateDB(utxos, "utxo_5", 1600)
	insertUTXOIntoStateDB(utxos, "utxo_6", 1000)

	broadcastTxs = p.MatchUTXOsAndUnshieldIDs(utxos, waitingUnshieldState, dustAmount)
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

	broadcastTxs = p.MatchUTXOsAndUnshieldIDs(utxos, waitingUnshieldState, dustAmount)
	printBroadcastTxs(t, broadcastTxs)

	// Broadcast multiple txs
	waitingUnshieldState = map[string]*statedb.WaitingUnshieldRequest{}
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_12", "unshield_12", 10000, 8)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_13", "unshield_13", 10010, 9)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_14", "unshield_14", 10000, 10)

	utxos = map[string]*statedb.UTXO{}
	insertUTXOIntoStateDB(utxos, "utxo_12", 10)
	insertUTXOIntoStateDB(utxos, "utxo_13", 15)
	insertUTXOIntoStateDB(utxos, "utxo_14", 1000)
	insertUTXOIntoStateDB(utxos, "utxo_15", 20)
	insertUTXOIntoStateDB(utxos, "utxo_16", 1000)
	insertUTXOIntoStateDB(utxos, "utxo_17", 1000)
	insertUTXOIntoStateDB(utxos, "utxo_18", 500)
	insertUTXOIntoStateDB(utxos, "utxo_19", 500)

	broadcastTxs = p.MatchUTXOsAndUnshieldIDs(utxos, waitingUnshieldState, dustAmount)
	printBroadcastTxs(t, broadcastTxs)

	waitingUnshieldState = map[string]*statedb.WaitingUnshieldRequest{}
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_15", "unshield_15", 10000, 8)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_16", "unshield_16", 10000, 9)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_17", "unshield_17", 100000, 10)
	insertUnshieldIDIntoStateDB(waitingUnshieldState, tokenID, "remoteAddr_18", "unshield_18", 5000, 10)

	utxos = map[string]*statedb.UTXO{}
	insertUTXOIntoStateDB(utxos, "utxo_20", 10)
	insertUTXOIntoStateDB(utxos, "utxo_21", 15)
	insertUTXOIntoStateDB(utxos, "utxo_22", 2000)
	insertUTXOIntoStateDB(utxos, "utxo_23", 20)
	insertUTXOIntoStateDB(utxos, "utxo_24", 1000)
	insertUTXOIntoStateDB(utxos, "utxo_25", 1000)
	insertUTXOIntoStateDB(utxos, "utxo_26", 500)
	insertUTXOIntoStateDB(utxos, "utxo_27", 500)

	broadcastTxs = p.MatchUTXOsAndUnshieldIDs(utxos, waitingUnshieldState, dustAmount)
	printBroadcastTxs(t, broadcastTxs)
}

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
	incAddress := "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci"
	script, address, err := p.GenerateOTMultisigAddress(masterPubKeys, 3, incAddress)
	if err != nil {
		t.Logf("Error: %v\n", err)
		t.FailNow()
	}
	t.Logf("P2WSH Bech32 address: %v\n", address)
	t.Logf("P2WSH Bech32 hex encode: %v\n", hex.EncodeToString(script))
}
