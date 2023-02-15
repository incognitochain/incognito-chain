package portaltokens

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	btcec "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	btcwire "github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
)

type PortalBTCTokenProcessor struct {
	*PortalToken
	ChainParam    *chaincfg.Params
	PortalTokenID string
}

func (p PortalBTCTokenProcessor) GetPortalTokenID() string {
	return p.PortalTokenID
}

func (p PortalBTCTokenProcessor) ConvertExternalToIncAmount(externalAmt uint64) uint64 {
	return externalAmt * 10
}

func (p PortalBTCTokenProcessor) ConvertIncToExternalAmount(incAmt uint64) uint64 {
	return incAmt / 10 // incAmt / 10^9 * 10^8
}

func (p PortalBTCTokenProcessor) GetTxHashFromProof(proof string) (string, error) {
	btcTxProof, err := btcrelaying.ParseAndValidateSanityBTCProofFromB64EncodeStr(proof)
	if err != nil {
		Logger.log.Errorf("Can not decode proof string %v\n", err)
		return "", fmt.Errorf("Can not decode proof string %v\n", err)
	}
	if btcTxProof.BTCTx == nil {
		Logger.log.Errorf("BTCTx is invalid %v\n", err)
		return "", fmt.Errorf("BTCTx is invalid  %v\n", err)
	}
	return btcTxProof.BTCTx.TxHash().String(), nil
}

func (p PortalBTCTokenProcessor) parseAndVerifyProofBTCChain(
	proof string, btcChain *btcrelaying.BlockChain, expectedMultisigAddress string, chainCodeSeed string, minShieldAmt uint64) (bool, []*statedb.UTXO, error) {
	if btcChain == nil {
		Logger.log.Error("BTC relaying chain should not be null")
		return false, nil, errors.New("BTC relaying chain should not be null")
	}
	// parse BTCProof in meta
	btcTxProof, err := btcrelaying.ParseAndValidateSanityBTCProofFromB64EncodeStr(proof)
	if err != nil {
		Logger.log.Errorf("ShieldingProof is invalid %v\n", err)
		return false, nil, fmt.Errorf("ShieldingProof is invalid %v\n", err)
	}

	// verify tx with merkle proofs
	isValid, err := btcChain.VerifyTxWithMerkleProofs(btcTxProof)
	if !isValid || err != nil {
		Logger.log.Errorf("Verify btcTxProof failed %v", err)
		return false, nil, fmt.Errorf("Verify btcTxProof failed %v", err)
	}

	// check whether amount transfer in txBNB is equal porting amount or not
	// check receiver and amount in tx
	outputs := btcTxProof.BTCTx.TxOut
	totalValue := uint64(0)

	listUTXO := []*statedb.UTXO{}

	for idx, out := range outputs {
		addrStr, err := btcChain.ExtractPaymentAddrStrFromPkScript(out.PkScript)
		if err != nil {
			Logger.log.Errorf("[portal] ExtractPaymentAddrStrFromPkScript: could not extract payment address string from pkscript with err: %v\n", err)
			continue
		}
		if strings.ToLower(addrStr) != strings.ToLower(expectedMultisigAddress) {
			continue
		}

		if p.ConvertExternalToIncAmount(uint64(out.Value)) < minShieldAmt {
			Logger.log.Errorf("Shielding UTXO amount: %v is less than the minimum threshold: %v\n", out.Value, minShieldAmt)
			continue
		}

		totalValue += uint64(out.Value)

		listUTXO = append(listUTXO, statedb.NewUTXOWithValue(
			addrStr,
			btcTxProof.BTCTx.TxHash().String(),
			uint32(idx),
			uint64(out.Value),
			chainCodeSeed,
		))
	}

	if len(listUTXO) == 0 {
		Logger.log.Errorf("Could not find any valid UTXO")
		return false, nil, fmt.Errorf("Could not find any valid UTXO")
	}

	return true, listUTXO, nil
}

func (p PortalBTCTokenProcessor) ParseAndVerifyShieldProof(
	proof string, bc metadata.ChainRetriever, expectedReceivedMultisigAddress string, chainCodeSeed string, minShieldAmt uint64,
) (bool, []*statedb.UTXO, error) {
	btcChain := bc.GetBTCHeaderChain()
	return p.parseAndVerifyProofBTCChain(proof, btcChain, expectedReceivedMultisigAddress, chainCodeSeed, minShieldAmt)
}

func (p PortalBTCTokenProcessor) ParseAndVerifyUnshieldProof(
	proof string,
	bc metadata.ChainRetriever,
	expectedReceivedMultisigAddress string,
	chainCodeSeed string,
	expectPaymentInfo []*OutputTx,
	utxos []*statedb.UTXO,
) (bool, []*statedb.UTXO, string, uint64, error) {
	btcChain := bc.GetBTCHeaderChain()
	if btcChain == nil {
		Logger.log.Error("BTC relaying chain should not be null")
		return false, nil, "", 0, errors.New("BTC relaying chain should not be null")
	}
	// parse BTCProof in meta
	btcTxProof, err := btcrelaying.ParseAndValidateSanityBTCProofFromB64EncodeStr(proof)
	if err != nil {
		Logger.log.Errorf("UnShieldingProof is invalid %v\n", err)
		return false, nil, "", 0, fmt.Errorf("UnShieldingProof is invalid %v\n", err)
	}

	// verify tx with merkle proofs
	isValid, err := btcChain.VerifyTxWithMerkleProofs(btcTxProof)
	if !isValid || err != nil {
		Logger.log.Errorf("Verify btcTxProof failed %v", err)
		return false, nil, "", 0, fmt.Errorf("Verify btcTxProof failed %v", err)
	}

	// verify spent outputs
	if len(btcTxProof.BTCTx.TxIn) < 1 {
		Logger.log.Errorf("Can not find the tx inputs in proof")
		return false, nil, "", 0, fmt.Errorf("Submit confirmed tx: no tx inputs in proof")
	}

	if len(utxos) != len(btcTxProof.BTCTx.TxIn) {
		Logger.log.Errorf("Length of transaction input coins is not match")
		return false, nil, "", 0, fmt.Errorf("Submit confirmed tx: Length of transaction input coins is not match")
	}
	for i, input := range btcTxProof.BTCTx.TxIn {
		if utxos[i].GetTxHash() != input.PreviousOutPoint.Hash.String() ||
			utxos[i].GetOutputIndex() != input.PreviousOutPoint.Index {
			Logger.log.Errorf("Submit confirmed: tx inputs from proof is diff utxos from unshield batch")
			return false, nil, "", 0, fmt.Errorf("Submit confirmed tx: tx inputs from proof is diff utxos from unshield batch")
		}
	}

	// verify outputs of the tx
	externalFee := uint64(0)
	outputs := btcTxProof.BTCTx.TxOut
	for idx, value := range expectPaymentInfo {
		receiverAddress := value.ReceiverAddress
		unshieldAmt := value.Amount
		if idx >= len(outputs) {
			Logger.log.Error("BTC-TxProof is invalid")
			return false, nil, "", 0, errors.New("BTC-TxProof is invalid")
		}
		addrStr, err := btcChain.ExtractPaymentAddrStrFromPkScript(outputs[idx].PkScript)
		if err != nil {
			Logger.log.Errorf("[portal] ExtractPaymentAddrStrFromPkScript: could not extract payment address string from pkscript with err: %v\n", err)
			return false, nil, "", 0, errors.New("Could not extract address from proof")
		}
		if strings.ToLower(addrStr) != strings.ToLower(receiverAddress) {
			Logger.log.Error("BTC-TxProof is invalid")
			return false, nil, "", 0, errors.New("BTC-TxProof is invalid")
		}
		if externalFee == 0 {
			valueInExternal := uint64(outputs[idx].Value)
			unshieldAmtInExternal := p.ConvertIncToExternalAmount(unshieldAmt)
			if unshieldAmtInExternal <= valueInExternal {
				Logger.log.Errorf("[portal] Calculate external fee error")
				return false, nil, "", 0, fmt.Errorf("[portal] Calculate external fee error")
			}
			externalFee = p.ConvertExternalToIncAmount(unshieldAmtInExternal - valueInExternal)
		}
	}

	// check the change output coin
	listUTXO := []*statedb.UTXO{}
	for idx, out := range outputs {
		addrStr, err := btcChain.ExtractPaymentAddrStrFromPkScript(out.PkScript)
		if err != nil {
			Logger.log.Errorf("[portal] ExtractPaymentAddrStrFromPkScript: could not extract payment address string from pkscript with err: %v\n", err)
			continue
		}
		if addrStr != expectedReceivedMultisigAddress {
			continue
		}
		listUTXO = append(listUTXO, statedb.NewUTXOWithValue(
			addrStr,
			btcTxProof.BTCTx.TxHash().String(),
			uint32(idx),
			uint64(out.Value),
			chainCodeSeed,
		))
	}

	return true, listUTXO, btcTxProof.BTCTx.TxHash().String(), externalFee, nil
}

func (p PortalBTCTokenProcessor) IsValidRemoteAddress(address string, bcr metadata.ChainRetriever) (bool, error) {
	btcHeaderChain := bcr.GetBTCHeaderChain()
	if btcHeaderChain == nil {
		return false, nil
	}
	return btcHeaderChain.IsBTCAddressValid(address), nil
}

func (p PortalBTCTokenProcessor) GetChainID() string {
	return p.ChainID
}

func (p PortalBTCTokenProcessor) GetMinTokenAmount() uint64 {
	return p.MinTokenAmount
}

func (p PortalBTCTokenProcessor) GetMultipleTokenAmount() uint64 {
	return p.MultipleTokenAmount
}

func (p PortalBTCTokenProcessor) generatePublicKeyFromPrivateKey(privateKey []byte) []byte {
	pkx, pky := btcec.S256().ScalarBaseMult(privateKey)
	pubKey := btcec.PublicKey{Curve: btcec.S256(), X: pkx, Y: pky}
	return pubKey.SerializeCompressed()
}

func (p PortalBTCTokenProcessor) generatePublicKeyFromSeed(seed []byte) []byte {
	// generate BTC master account
	BTCPrivateKeyMaster := chainhash.HashB(seed) // private mining key => private key btc
	return p.generatePublicKeyFromPrivateKey(BTCPrivateKeyMaster)
}

func (p PortalBTCTokenProcessor) generateOTPrivateKey(seed []byte, chainCodeSeed string) ([]byte, error) {
	BTCPrivateKeyMaster := chainhash.HashB(seed) // private mining key => private key btc

	// this Incognito address is marked for the address that received change UTXOs
	if chainCodeSeed == "" {
		return BTCPrivateKeyMaster, nil
	} else {
		chainCode := chainhash.HashB([]byte(chainCodeSeed))
		extendedBTCPrivateKey := hdkeychain.NewExtendedKey(p.ChainParam.HDPrivateKeyID[:], BTCPrivateKeyMaster, chainCode, []byte{}, 0, 0, true)
		extendedBTCChildPrivateKey, err := extendedBTCPrivateKey.Child(0)
		if err != nil {
			return []byte{}, fmt.Errorf("Could not generate child private key for incognito address: %v", chainCodeSeed)
		}
		btcChildPrivateKey, err := extendedBTCChildPrivateKey.ECPrivKey()
		if err != nil {
			return []byte{}, fmt.Errorf("Could not get private key from extended private key")
		}
		btcChildPrivateKeyBytes := btcChildPrivateKey.Serialize()
		return btcChildPrivateKeyBytes, nil
	}
}

// Generate Bech32 P2WSH multisig address for each Incognito address
// Return redeem script, OTMultisigAddress
func (p PortalBTCTokenProcessor) GenerateOTMultisigAddress(masterPubKeys [][]byte, numSigsRequired int, chainCodeSeed string) ([]byte, string, error) {
	if len(masterPubKeys) < numSigsRequired || numSigsRequired < 0 {
		return []byte{}, "", fmt.Errorf("Invalid signature requirement")
	}

	pubKeys := [][]byte{}
	// this Incognito address is marked for the address that received change UTXOs
	if chainCodeSeed == "" {
		pubKeys = masterPubKeys[:]
	} else {
		chainCode := chainhash.HashB([]byte(chainCodeSeed))
		for idx, masterPubKey := range masterPubKeys {
			// generate BTC child public key for this Incognito address
			extendedBTCPublicKey := hdkeychain.NewExtendedKey(p.ChainParam.HDPublicKeyID[:], masterPubKey, chainCode, []byte{}, 0, 0, false)
			extendedBTCChildPubKey, _ := extendedBTCPublicKey.Child(0)
			childPubKey, err := extendedBTCChildPubKey.ECPubKey()
			if err != nil {
				return []byte{}, "", fmt.Errorf("Master BTC Public Key (#%v) %v is invalid - Error %v", idx, masterPubKey, err)
			}
			pubKeys = append(pubKeys, childPubKey.SerializeCompressed())
		}
	}

	// create redeem script for m of n multi-sig
	builder := txscript.NewScriptBuilder()
	// add the minimum number of needed signatures
	builder.AddOp(byte(txscript.OP_1 - 1 + numSigsRequired))
	// add the public key to redeem script
	for _, pubKey := range pubKeys {
		builder.AddData(pubKey)
	}
	// add the total number of public keys in the multi-sig script
	builder.AddOp(byte(txscript.OP_1 - 1 + len(pubKeys)))
	// add the check-multi-sig op-code
	builder.AddOp(txscript.OP_CHECKMULTISIG)

	redeemScript, err := builder.Script()
	if err != nil {
		return []byte{}, "", fmt.Errorf("Could not build script - Error %v", err)
	}

	// generate P2WSH address
	scriptHash := sha256.Sum256(redeemScript)
	addr, err := btcutil.NewAddressWitnessScriptHash(scriptHash[:], p.ChainParam)
	if err != nil {
		return []byte{}, "", fmt.Errorf("Could not generate address from script - Error %v", err)
	}
	addrStr := addr.EncodeAddress()

	return redeemScript, addrStr, nil
}

// CreateRawExternalTx creates raw btc transaction (not include signatures of beacon validator)
// inputs: UTXO state of beacon, unit of amount in btc
// outputs: unit of amount in pbtc ~ unshielding amount
// feePerOutput: unit in pbtc
func (p PortalBTCTokenProcessor) CreateRawExternalTx(inputs []*statedb.UTXO, outputs []*OutputTx, feePerOutput uint64,
	bc metadata.ChainRetriever, beaconHeight uint64) (string, string, error) {
	msgTx := wire.NewMsgTx(wire.TxVersion)

	// convert feePerOutput from inc unit to external unit
	feePerOutputInExternal := p.ConvertIncToExternalAmount(feePerOutput)

	// add TxIns into raw tx
	// totalInputAmount in external unit
	totalInputAmount := uint64(0)
	for _, in := range inputs {
		utxoHash, err := chainhash.NewHashFromStr(in.GetTxHash())
		if err != nil {
			Logger.log.Errorf("[CreateRawExternalTx-BTC] Error when new TxIn for tx: %v", err)
			return "", "", err
		}
		outPoint := wire.NewOutPoint(utxoHash, in.GetOutputIndex())
		txIn := wire.NewTxIn(outPoint, nil, nil)
		txIn.Sequence = uint32(feePerOutputInExternal)
		msgTx.AddTxIn(txIn)
		totalInputAmount += in.GetOutputAmount()
	}

	// add TxOuts into raw tx
	// totalOutputAmount in external unit
	totalOutputAmount := uint64(0)
	for _, out := range outputs {
		// adding the output to tx
		decodedAddr, err := btcutil.DecodeAddress(out.ReceiverAddress, p.ChainParam)
		if err != nil {
			Logger.log.Errorf("[CreateRawExternalTx-BTC] Error when decoding receiver address: %v", err)
			return "", "", err
		}
		destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
		if err != nil {
			Logger.log.Errorf("[CreateRawExternalTx-BTC] Error when new Address Script: %v", err)
			return "", "", err
		}

		// adding the destination address and the amount to the transaction
		outAmountInExternal := p.ConvertIncToExternalAmount(out.Amount)
		if outAmountInExternal <= feePerOutputInExternal {
			Logger.log.Errorf("[CreateRawExternalTx-BTC] Output amount %v must greater than fee %v", out.Amount, feePerOutputInExternal)
			return "", "", fmt.Errorf("[CreateRawExternalTx-BTC] Output amount %v must greater than fee %v", out.Amount, feePerOutputInExternal)
		}
		redeemTxOut := wire.NewTxOut(int64(outAmountInExternal-feePerOutputInExternal), destinationAddrByte)
		msgTx.AddTxOut(redeemTxOut)
		totalOutputAmount += outAmountInExternal
	}

	// check amount of input coins and output coins
	if totalInputAmount < totalOutputAmount {
		Logger.log.Errorf("[CreateRawExternalTx-BTC] Total input amount %v is less than total output amount %v", totalInputAmount, totalOutputAmount)
		return "", "", fmt.Errorf("[CreateRawExternalTx-BTC] Total input amount %v is less than total output amount %v", totalInputAmount, totalOutputAmount)
	}

	// calculate the change output
	if totalInputAmount > totalOutputAmount {
		// adding the output to tx
		multiSigAddress := bc.GetPortalV4GeneralMultiSigAddress(p.GetPortalTokenID(), beaconHeight)
		decodedAddr, err := btcutil.DecodeAddress(multiSigAddress, p.ChainParam)
		if err != nil {
			Logger.log.Errorf("[CreateRawExternalTx-BTC] Error when decoding multisig address: %v", err)
			return "", "", err
		}
		destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
		if err != nil {
			Logger.log.Errorf("[CreateRawExternalTx-BTC] Error when new multisig Address Script: %v", err)
			return "", "", err
		}

		// adding the destination address and the amount to the transaction
		redeemTxOut := wire.NewTxOut(int64(totalInputAmount-totalOutputAmount), destinationAddrByte)
		msgTx.AddTxOut(redeemTxOut)
	}

	var rawTxBytes bytes.Buffer
	err := msgTx.Serialize(&rawTxBytes)
	if err != nil {
		Logger.log.Errorf("[CreateRawExternalTx-BTC] Error when serializing raw tx: %v", err)
		return "", "", err
	}

	hexRawTx := hex.EncodeToString(rawTxBytes.Bytes())
	return hexRawTx, msgTx.TxHash().String(), nil
}

func (p PortalBTCTokenProcessor) PartSignOnRawExternalTx(seedKey []byte, masterPubKeys [][]byte, numSigsRequired int, rawTxBytes []byte, inputs []*statedb.UTXO) ([][]byte, string, error) {
	// new MsgTx from rawTxBytes
	msgTx := new(btcwire.MsgTx)
	rawTxBuffer := bytes.NewBuffer(rawTxBytes)
	err := msgTx.Deserialize(rawTxBuffer)
	if err != nil {
		return nil, "", fmt.Errorf("[PartSignOnRawExternalTx] Error when deserializing raw tx bytes: %v", err)
	}
	// sign on each TxIn
	if len(inputs) != len(msgTx.TxIn) {
		return nil, "", fmt.Errorf("[PartSignOnRawExternalTx] Len of Public seeds %v and len of TxIn %v are not correct", len(inputs), len(msgTx.TxIn))
	}
	sigs := [][]byte{}
	for i := range msgTx.TxIn {
		// generate btc private key from seed: private key of bridge consensus
		btcPrivateKeyBytes, err := p.generateOTPrivateKey(seedKey, inputs[i].GetChainCodeSeed())
		if err != nil {
			return nil, "", fmt.Errorf("[PartSignOnRawExternalTx] Error when generate btc private key from seed: %v", err)
		}
		btcPrivateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), btcPrivateKeyBytes)
		multiSigScript, _, err := p.GenerateOTMultisigAddress(masterPubKeys, numSigsRequired, inputs[i].GetChainCodeSeed())
		if err != nil {
			return nil, "", fmt.Errorf("[PartSignOnRawExternalTx] Error when generate multi sig address: %v", err)
		}
		sig, err := txscript.RawTxInWitnessSignature(msgTx, txscript.NewTxSigHashes(msgTx), i, int64(inputs[i].GetOutputAmount()), multiSigScript, txscript.SigHashAll, btcPrivateKey)
		if err != nil {
			return nil, "", fmt.Errorf("[PartSignOnRawExternalTx] Error when signing on raw btc tx: %v", err)
		}
		sigs = append(sigs, sig)
	}

	return sigs, msgTx.TxHash().String(), nil
}

func (p PortalBTCTokenProcessor) IsAcceptableTxSize(numInputs int, numOutputs int) bool {
	return p.ExternalInputSize*uint(numInputs)+p.ExternalOutputSize*uint(numOutputs) <= p.ExternalTxMaxSize
}

func (p PortalBTCTokenProcessor) MatchUTXOsAndUnshieldIDsNew(
	utxos map[string]*statedb.UTXO,
	waitingUnshieldReqs map[string]*statedb.WaitingUnshieldRequest,
	tinyValueThreshold uint64,
	minUTXOs uint64,
) ([]*BroadcastTx, error) {
	batchTxs := []*BroadcastTx{}
	if len(utxos) == 0 || len(waitingUnshieldReqs) == 0 {
		return batchTxs, nil
	}

	// sort UTXOs descending by amount
	sortedUTXOs := p.sortUTXOsByAmountDescending(utxos)
	// sort UTXOs ascending by beaconHeight
	sortedUnshieldReqs := p.sortUnshieldReqsByBeaconHeightAscending(waitingUnshieldReqs)

	remainUnshieldReqs := []unshieldItem{}
	// choose UTXO(s) for each waiting unshielding requests
	// create one batch for one unshielding request
	for _, unshieldReq := range sortedUnshieldReqs {
		chosenUtxos, chosenIndices, err := p.ChooseUTXOsForUnshieldReq(sortedUTXOs, unshieldReq.value.GetAmount())
		if err != nil {
			Logger.log.Errorf("Error when choose utxo for unshield ID: %v - Error %v\n",
				unshieldReq.value.GetUnshieldID(), err)
			remainUnshieldReqs = append(remainUnshieldReqs, unshieldReq)
			continue
		}
		chosenUtxoObjs := []*statedb.UTXO{}
		for _, u := range chosenUtxos {
			chosenUtxoObjs = append(chosenUtxoObjs, u.value)
		}
		batchTxs = append(batchTxs, &BroadcastTx{
			UTXOs:       chosenUtxoObjs,
			UnshieldIDs: []string{unshieldReq.value.GetUnshieldID()},
		})
		// remove chosen utxos in sortedUTXOs (for next unshielding requests)
		offsetIndex := 0
		for _, index := range chosenIndices {
			if index == 0 {
				sortedUTXOs = sortedUTXOs[1:]
			} else {
				sortedUTXOs = append(sortedUTXOs[:index-offsetIndex], sortedUTXOs[index-offsetIndex+1:]...)
			}
			offsetIndex++
		}
	}

	// fmt.Printf("batchTxs: %+v\n", batchTxs)
	// merge small batches
	batchTxs = p.MergeBatches(batchTxs)

	// appending tiny into batch txs
	batchTxs = p.AppendTinyUTXOs(batchTxs, sortedUTXOs, tinyValueThreshold, minUTXOs)

	// appending remain unshielding request (if any)
	batchTxs = p.AppendWaitingUnshieldRequests(batchTxs, remainUnshieldReqs, waitingUnshieldReqs)

	return batchTxs, nil
}

//TODO: update
// Choose list of pairs (UTXOs and unshield IDs) for broadcast external transactions
func (p PortalBTCTokenProcessor) MatchUTXOsAndUnshieldIDs(
	utxos map[string]*statedb.UTXO,
	waitingUnshieldReqs map[string]*statedb.WaitingUnshieldRequest,
	dustValueThreshold uint64) []*BroadcastTx {
	if len(utxos) == 0 || len(waitingUnshieldReqs) == 0 {
		return []*BroadcastTx{}
	}

	// descending sort utxo by value
	type utxoItem struct {
		key   string
		value *statedb.UTXO
	}
	utxosArr := []utxoItem{}
	for k, req := range utxos {
		utxosArr = append(
			utxosArr,
			utxoItem{
				key:   k,
				value: req,
			})
	}
	sort.SliceStable(utxosArr, func(i, j int) bool {
		if utxosArr[i].value.GetOutputAmount() > utxosArr[j].value.GetOutputAmount() {
			return true
		} else if utxosArr[i].value.GetOutputAmount() == utxosArr[j].value.GetOutputAmount() {
			return utxosArr[i].key < utxosArr[j].key
		}
		return false
	})

	// ascending sort waitingUnshieldReqs by beaconHeight
	type unshieldItem struct {
		key   string
		value *statedb.WaitingUnshieldRequest
	}

	// convert unshield amount to external token amount
	wReqsArr := []unshieldItem{}
	for k, req := range waitingUnshieldReqs {
		wReqsArr = append(
			wReqsArr,
			unshieldItem{
				key: k,
				value: statedb.NewWaitingUnshieldRequestStateWithValue(
					req.GetRemoteAddress(), p.ConvertIncToExternalAmount(req.GetAmount()), req.GetUnshieldID(), req.GetBeaconHeight()),
			})
	}

	sort.SliceStable(wReqsArr, func(i, j int) bool {
		if wReqsArr[i].value.GetBeaconHeight() < wReqsArr[j].value.GetBeaconHeight() {
			return true
		} else if wReqsArr[i].value.GetBeaconHeight() == wReqsArr[j].value.GetBeaconHeight() {
			return wReqsArr[i].key < wReqsArr[j].key
		}
		return false
	})

	broadcastTxs := []*BroadcastTx{}
	utxoIdx := 0
	unshieldIdx := 0
	for utxoIdx < len(utxos) && unshieldIdx < len(wReqsArr) {
		// utxoIdx always increases at least 1 in this scope
		lastUTXOIdx, lastUnshieldIdx := utxoIdx, unshieldIdx

		chosenUTXOs := []*statedb.UTXO{}
		chosenUnshieldIDs := []string{}

		curSumAmount := uint64(0)
		cnt := 0
		if utxosArr[utxoIdx].value.GetOutputAmount() >= wReqsArr[unshieldIdx].value.GetAmount() {
			// find the last unshield idx that the cumulative sum of unshield amount <= current utxo amount
			for unshieldIdx < len(wReqsArr) && curSumAmount+wReqsArr[unshieldIdx].value.GetAmount() <= utxosArr[utxoIdx].value.GetOutputAmount() && p.IsAcceptableTxSize(1, cnt+1) {
				curSumAmount += wReqsArr[unshieldIdx].value.GetAmount()
				chosenUnshieldIDs = append(chosenUnshieldIDs, wReqsArr[unshieldIdx].value.GetUnshieldID())
				unshieldIdx += 1
				cnt += 1
			}
			chosenUTXOs = append(chosenUTXOs, utxosArr[utxoIdx].value)
			utxoIdx += 1 // utxoIdx increases
		} else {
			// find the first utxo idx that the cumulative sum of utxo amount >= current unshield amount
			for utxoIdx < len(utxos) && curSumAmount+utxosArr[utxoIdx].value.GetOutputAmount() < wReqsArr[unshieldIdx].value.GetAmount() {
				curSumAmount += utxosArr[utxoIdx].value.GetOutputAmount()
				chosenUTXOs = append(chosenUTXOs, utxosArr[utxoIdx].value)
				utxoIdx += 1 // utxoIdx increases
				cnt += 1
			}
			if utxoIdx < len(utxos) && p.IsAcceptableTxSize(cnt+1, 1) {
				curSumAmount += utxosArr[utxoIdx].value.GetOutputAmount()
				chosenUTXOs = append(chosenUTXOs, utxosArr[utxoIdx].value)
				utxoIdx += 1
				cnt += 1

				newCnt := 0
				target := curSumAmount
				curSumAmount = 0

				// insert new unshield IDs if the current utxos still has enough amount
				for unshieldIdx < len(wReqsArr) && curSumAmount+wReqsArr[unshieldIdx].value.GetAmount() <= target && p.IsAcceptableTxSize(cnt, newCnt+1) {
					curSumAmount += wReqsArr[unshieldIdx].value.GetAmount()
					chosenUnshieldIDs = append(chosenUnshieldIDs, wReqsArr[unshieldIdx].value.GetUnshieldID())
					unshieldIdx += 1
					newCnt += 1
				}

			} else {
				// not enough utxo for last unshield IDs
				utxoIdx, unshieldIdx = lastUTXOIdx, lastUnshieldIdx
				break
			}
		}

		broadcastTxs = append(broadcastTxs, &BroadcastTx{
			UTXOs:       chosenUTXOs,
			UnshieldIDs: chosenUnshieldIDs,
		})
	}

	// merged small batches into bigger batches
	mergedBatches := []*BroadcastTx{}
	if len(broadcastTxs) > 0 {
		mergedBatches = append(mergedBatches, broadcastTxs[0])
		for idx := 1; idx < len(broadcastTxs); idx++ {
			mergedIdx := len(mergedBatches) - 1
			prevUTXOs, prevRequests := mergedBatches[mergedIdx].UTXOs, mergedBatches[mergedIdx].UnshieldIDs
			curUTXOs, curRequests := broadcastTxs[idx].UTXOs, broadcastTxs[idx].UnshieldIDs

			lenUTXOs := len(prevUTXOs) + len(curUTXOs)
			lenRequests := len(prevRequests) + len(curRequests)
			if p.IsAcceptableTxSize(lenUTXOs, lenRequests) {
				mergedBatches[mergedIdx] = &BroadcastTx{
					UTXOs:       append(prevUTXOs, curUTXOs...),
					UnshieldIDs: append(prevRequests, curRequests...),
				}
			} else {
				mergedBatches = append(mergedBatches, &BroadcastTx{
					UTXOs:       curUTXOs,
					UnshieldIDs: curRequests,
				})
			}
		}
	}

	// add a dust UTXO
	dustUTXOUsed := 0
	for idx := 0; idx < len(mergedBatches); idx++ {
		if utxoIdx < len(utxos)-dustUTXOUsed &&
			utxosArr[len(utxos)-dustUTXOUsed-1].value.GetOutputAmount() <= dustValueThreshold &&
			p.IsAcceptableTxSize(len(mergedBatches[idx].UTXOs)+1, len(mergedBatches[idx].UnshieldIDs)) {
			dustUTXOUsed += 1
			mergedBatches[idx].UTXOs = append(mergedBatches[idx].UTXOs, utxosArr[len(utxos)-dustUTXOUsed].value)
		}
	}

	// add small unshield requests while they still can fit in merged batches
	unshieldAmountMap := map[string]uint64{}
	for _, req := range waitingUnshieldReqs {
		unshieldAmountMap[req.GetUnshieldID()] = p.ConvertIncToExternalAmount(req.GetAmount())
	}
	for idx := 0; idx < len(mergedBatches); idx++ {
		sumUTXOAmount, sumRequestAmount := uint64(0), uint64(0)
		for _, val := range mergedBatches[idx].UTXOs {
			sumUTXOAmount += val.GetOutputAmount()
		}
		for _, val := range mergedBatches[idx].UnshieldIDs {
			sumRequestAmount += unshieldAmountMap[val]
		}
		for unshieldIdx < len(wReqsArr) && sumRequestAmount+wReqsArr[unshieldIdx].value.GetAmount() <= sumUTXOAmount &&
			p.IsAcceptableTxSize(len(mergedBatches[idx].UTXOs), len(mergedBatches[idx].UnshieldIDs)+1) {
			sumRequestAmount += wReqsArr[unshieldIdx].value.GetAmount()
			mergedBatches[idx].UnshieldIDs = append(mergedBatches[idx].UnshieldIDs, wReqsArr[unshieldIdx].value.GetUnshieldID())
			unshieldIdx += 1
		}
	}

	return mergedBatches
}

func (p PortalBTCTokenProcessor) GetTxHashFromRawTx(rawTx string) (string, error) {
	// parse rawTx to get externalTxHash
	hexRawTx, err := hex.DecodeString(rawTx)
	if err != nil {
		return "", err
	}
	buffer := bytes.NewReader(hexRawTx)
	externalTx := wire.NewMsgTx(wire.TxVersion)
	err = externalTx.Deserialize(buffer)
	if err != nil {
		return "", err
	}

	return externalTx.TxHash().String(), nil
}
