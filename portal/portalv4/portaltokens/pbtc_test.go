package portaltokens

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"

	btcec "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
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
		ChainParam:    &chaincfg.TestNet3Params,
		PortalTokenID: "4584d5e9b2fc0337dfb17f4b5bb025e5b82c38cfa4f54e8a3d4fcdd03954ff82",
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
		ChainParam:    &chaincfg.TestNet3Params,
		PortalTokenID: "4584d5e9b2fc0337dfb17f4b5bb025e5b82c38cfa4f54e8a3d4fcdd03954ff82",
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
		ChainParam:    &chaincfg.TestNet3Params,
		PortalTokenID: "4584d5e9b2fc0337dfb17f4b5bb025e5b82c38cfa4f54e8a3d4fcdd03954ff82",
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

func TestGenerateOTMultisigAddressFromMasterPubKeys(t *testing.T) {
	p := &PortalBTCTokenProcessor{
		ChainParam:    &chaincfg.TestNet3Params,
		PortalTokenID: "b832e5d3b1f01a4f0623f7fe91d6673461e1f5d37d91fe78c5c2e6183ff39696",
	}

	masterPubKeys := [][]byte{
		// []byte{0x2, 0x39, 0x42, 0x3d, 0xad, 0x93, 0x8f, 0xcb, 0xe5, 0xb5, 0xef, 0x7b, 0x7b, 0x9a, 0xf, 0x28,
		// 	0x4, 0x19, 0x53, 0x66, 0x7f, 0xee, 0x72, 0xe4, 0x81, 0xf9, 0xe6, 0xb, 0x81, 0x41, 0xd7, 0x3a, 0x36},
		// []byte{0x2, 0x8d, 0xc, 0xd7, 0x83, 0x9d, 0x5e, 0xc5, 0x7b, 0x77, 0x1a, 0xf1, 0x2, 0xb8, 0x72, 0xd0,
		// 	0x4f, 0x34, 0xb4, 0xeb, 0x17, 0xac, 0xa1, 0x9f, 0xdf, 0xa, 0x64, 0xbf, 0xd, 0x36, 0x76, 0x66, 0x87},
		// []byte{0x3, 0x78, 0x52, 0x33, 0xe3, 0x8, 0x3a, 0xd8, 0x58, 0x77, 0x76, 0x29, 0xa0, 0x17, 0xb6, 0xdd,
		// 	0x16, 0x43, 0x18, 0x8b, 0xb4, 0xa3, 0xaf, 0x45, 0xf0, 0xb5, 0x91, 0x8c, 0x84, 0xf2, 0x73, 0x56, 0x44},
		// []byte{0x3, 0x61, 0x9d, 0xc9, 0xfb, 0x6d, 0x8, 0x2a, 0x5c, 0x98, 0x45, 0xbc, 0xbf, 0x86, 0xfb, 0x47,
		// 	0x4, 0xbe, 0x67, 0x46, 0xa, 0x59, 0xc4, 0xbc, 0x1d, 0xec, 0xc0, 0xe8, 0xe4, 0x3e, 0x1d, 0x6d, 0x0},
		// []byte{0x2, 0xe4, 0x1d, 0x40, 0xe6, 0xf3, 0x80, 0xad, 0x51, 0xca, 0x17, 0x87, 0xfe, 0xc8, 0x23, 0x8d,
		// 	0xa4, 0xc2, 0x88, 0xfc, 0xfb, 0x6f, 0x2b, 0xcc, 0xd9, 0xa6, 0x1c, 0x2, 0xe5, 0x4a, 0x31, 0x34, 0x39},
		// []byte{0x2, 0xf0, 0xc, 0xe3, 0xec, 0x4, 0xdb, 0x75, 0x59, 0x99, 0x70, 0xc6, 0xfd, 0xc5, 0x2, 0x2f,
		// 	0xad, 0x6b, 0x8d, 0x18, 0x86, 0x71, 0x44, 0xcf, 0xe6, 0x93, 0x92, 0xbb, 0xd1, 0x60, 0xc1, 0x1b, 0x5c},
		// []byte{0x2, 0x65, 0x96, 0x49, 0xab, 0xd4, 0xe5, 0x97, 0x7d, 0x5b, 0x67, 0x4c, 0x6d, 0xa1, 0xf, 0x9,
		// 	0x28, 0xa0, 0x8c, 0x67, 0x8d, 0x7f, 0x50, 0xcc, 0x10, 0xf0, 0xfe, 0xe5, 0x68, 0xa8, 0x57, 0x63, 0xd8},

		[]byte{0x2, 0x30, 0x34, 0xcb, 0x1a, 0x50, 0xf6, 0x7f, 0x5e, 0xb2, 0x53, 0x9e, 0x68, 0x3b, 0xd4,
			0x80, 0x73, 0x71, 0x2a, 0xdf, 0xf3, 0x25, 0x94, 0x34, 0x72, 0x6d, 0x62, 0x80, 0x83, 0xd2, 0x6f, 0x4c, 0xdd},
		[]byte{0x2, 0x74, 0x61, 0x32, 0x93, 0xe7, 0x93, 0x85, 0x94, 0xd2, 0x58, 0xfb, 0xcf, 0xc5, 0x33,
			0x78, 0xdc, 0x82, 0xcd, 0x64, 0xd1, 0xc0, 0x33, 0x1, 0x71, 0x2f, 0x90, 0x85, 0x72, 0xb9, 0x17, 0xab, 0xc7},
		[]byte{0x3, 0x67, 0x7a, 0x81, 0xfc, 0x9c, 0x4c, 0x9c, 0x6, 0x28, 0xd2, 0xf6, 0xd0, 0x1e, 0x27,
			0x15, 0xbb, 0x54, 0x11, 0x75, 0xe9, 0x62, 0xae, 0x78, 0x8f, 0xff, 0x26, 0x75, 0x1e, 0xb5, 0x24, 0xe0, 0xeb},
		[]byte{0x3, 0x2, 0xdb, 0xd4, 0xd4, 0x6b, 0x4e, 0xef, 0xe9, 0xa6, 0xe8, 0x64, 0xce, 0xeb, 0xb5,
			0x11, 0x25, 0x71, 0x28, 0x8a, 0xc4, 0xce, 0xca, 0xf4, 0x10, 0xd4, 0x16, 0x5f, 0x4c, 0x4c, 0xeb, 0x27, 0xe3},
	}
	chainCode := "12si2KgWLGuhXACeqHGquGpyQy7JZiA5qRTCWW7YTYrEzZBuZC2eGBfckc2NRXkQXiw7XwK2WVfKxC8AcwKGCsyRVr9SR8bN9vTcnk2PPbymztCWadgr9JMP1UY6oSk9XZb56EAKvK1fJd5S8ptY"
	// chainCode := common.PortalConvertVaultChainCode
	script, address, err := p.GenerateOTMultisigAddress(masterPubKeys, 3, chainCode)
	if err != nil {
		t.Logf("Error: %v\n", err)
		t.FailNow()
	}
	t.Logf("P2WSH Bech32 address: %v\n", address)
	t.Logf("P2WSH Bech32 hex encode: %v\n", hex.EncodeToString(script))
}

func TestMatchUTXOsAndUnshieldIDsNew(t *testing.T) {
	p := PortalBTCTokenProcessor{
		PortalToken: &PortalToken{
			ChainID:             "Bitcoin-Mainnet",
			MinTokenAmount:      10,
			MultipleTokenAmount: 10,
			ExternalInputSize:   2,
			ExternalOutputSize:  1,
			ExternalTxMaxSize:   30,
		},
		ChainParam:    &chaincfg.MainNetParams,
		PortalTokenID: "b832e5d3b1f01a4f0623f7fe91d6673461e1f5d37d91fe78c5c2e6183ff39696",
	}
	thresholdTinyValue := uint64(4)
	minUTXOs := uint64(3)

	// utxoAmts := []uint64{1, 8, 4, 5, 2, 7, 12, 15, 3, 2, 3}
	// unshieldAmts := []uint64{30, 60, 250, 50, 90, 40}
	utxoAmts := []uint64{100, 2, 3, 1}
	unshieldAmts := []uint64{70, 90, 80, 40}
	// 3 => 7 => (15, 12) => 5 => (8,1)

	type batch struct {
		utxoAmts     []uint64
		unshieldAmts []uint64
	}
	expectedBatchs := []batch{
		// utxo amount > unshield amount: choose 1 utxo for 1 unshield
		// append tiny utxo(s)
		{
			utxoAmts:     []uint64{3, 7, 2},
			unshieldAmts: []uint64{30, 60},
		},
		// utxo amount < unshield amount: choose utxos for 1 unshield
		// append tiny utxo(s)
		{
			utxoAmts:     []uint64{15, 12, 5},
			unshieldAmts: []uint64{250, 50},
		},
		// num(utxos) < min utxos: don't pick tiny uxto
		{
			utxoAmts:     []uint64{8, 1, 4},
			unshieldAmts: []uint64{90, 40},
		},
	}

	utxos := map[string]*statedb.UTXO{}
	for i, value := range utxoAmts {
		key := "UTXO " + strconv.Itoa(i)
		utxos[key] = statedb.NewUTXOWithValue("", "", 0, value, "")
	}
	waitingUnshieldReqs := map[string]*statedb.WaitingUnshieldRequest{}
	for i, value := range unshieldAmts {
		unshieldID := "Unshield " + strconv.Itoa(i)
		key := statedb.GenerateWaitingUnshieldRequestObjectKey(p.PortalTokenID, unshieldID).String()
		waitingUnshieldReqs[key] = statedb.NewWaitingUnshieldRequestStateWithValue("", value, unshieldID, 100)
	}

	batchTxs, err := p.MatchUTXOsAndUnshieldIDsNew(utxos, waitingUnshieldReqs, thresholdTinyValue, minUTXOs)
	fmt.Printf("err: %v\n", err)

	for i, batchTx := range batchTxs {
		_ = expectedBatchs[i]
		utxoAmts := []uint64{}
		for _, u := range batchTx.UTXOs {
			utxoAmts = append(utxoAmts, u.GetOutputAmount())
		}
		fmt.Printf("utxoAmts: %+v\n", utxoAmts)

		unshieldAmts := []uint64{}
		for _, u := range batchTx.UnshieldIDs {
			key := statedb.GenerateWaitingUnshieldRequestObjectKey(p.PortalTokenID, u).String()
			unshieldAmts = append(unshieldAmts, waitingUnshieldReqs[key].GetAmount())
		}
		fmt.Printf("unshieldAmts: %+v\n", unshieldAmts)
		// assert.Equal(t, true, reflect.DeepEqual(utxoAmts, expectedBatch.utxoAmts))
		// assert.Equal(t, true, reflect.DeepEqual(unshieldAmts, expectedBatch.unshieldAmts))
	}
}
