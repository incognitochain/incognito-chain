package portaltokens

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/portal/portalv4/common"

	"sort"

	"github.com/btcsuite/btcd/btcec"
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
	ChainParam *chaincfg.Params
}

func genBTCPrivateKey(IncKeyBytes []byte) []byte {
	BTCKeyBytes := ed25519.NewKeyFromSeed(IncKeyBytes)[32:]
	return BTCKeyBytes
}

func (p PortalBTCTokenProcessor) ConvertExternalToIncAmount(externalAmt uint64) uint64 {
	return externalAmt * 10
}

func (p PortalBTCTokenProcessor) ConvertIncToExternalAmount(incAmt uint64) uint64 {
	return incAmt / 10 // incAmt / 10^9 * 10^8
}

func (p PortalBTCTokenProcessor) parseAndVerifyProofBTCChain(
	proof string, btcChain *btcrelaying.BlockChain, expectedMultisigAddress string, chainCodeSeed string) (bool, []*statedb.UTXO, error) {
	if btcChain == nil {
		Logger.log.Error("BTC relaying chain should not be null")
		return false, nil, errors.New("BTC relaying chain should not be null")
	}
	// parse BTCProof in meta
	btcTxProof, err := btcrelaying.ParseBTCProofFromB64EncodeStr(proof)
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
		if addrStr != expectedMultisigAddress {
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

	if len(listUTXO) == 0 || p.ConvertExternalToIncAmount(totalValue) < p.GetMinTokenAmount() {
		Logger.log.Errorf("Shielding amount: %v is less than the minimum threshold: %v\n", totalValue, p.GetMinTokenAmount())
		return false, nil, fmt.Errorf("Shielding amount: %v is less than the minimum threshold: %v", totalValue, p.GetMinTokenAmount())
	}

	return true, listUTXO, nil
}

func (p PortalBTCTokenProcessor) ParseAndVerifyShieldProof(
	proof string, bc metadata.ChainRetriever, expectedReceivedMultisigAddress string, chainCodeSeed string) (bool, []*statedb.UTXO, error) {
	btcChain := bc.GetBTCHeaderChain()
	return p.parseAndVerifyProofBTCChain(proof, btcChain, expectedReceivedMultisigAddress, chainCodeSeed)
}

func (p PortalBTCTokenProcessor) ParseAndVerifyUnshieldProof(
	proof string, bc metadata.ChainRetriever, expectedReceivedMultisigAddress string, chainCodeSeed string, expectPaymentInfo []*OutputTx, utxos []*statedb.UTXO) (bool, []*statedb.UTXO, string, uint64, error) {
	btcChain := bc.GetBTCHeaderChain()
	if btcChain == nil {
		Logger.log.Error("BTC relaying chain should not be null")
		return false, nil, "", 0, errors.New("BTC relaying chain should not be null")
	}
	// parse BTCProof in meta
	btcTxProof, err := btcrelaying.ParseBTCProofFromB64EncodeStr(proof)
	if err != nil {
		Logger.log.Errorf("ShieldingProof is invalid %v\n", err)
		return false, nil, "", 0, fmt.Errorf("ShieldingProof is invalid %v\n", err)
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

	for _, input := range btcTxProof.BTCTx.TxIn {
		isMatched := false
		for _, v := range utxos {
			if v.GetTxHash() == input.PreviousOutPoint.Hash.String() && v.GetOutputIndex() == input.PreviousOutPoint.Index {
				isMatched = true
				break
			}
		}
		if !isMatched {
			Logger.log.Errorf("Submit confirmed: tx inputs from proof is diff utxos from unshield batch")
			return false, nil, "", 0, fmt.Errorf("Submit confirmed tx: tx inputs from proof is diff utxos from unshield batch")
		}
	}

	// check whether amount transfer in txBNB is equal porting amount or not
	// check receiver and amount in tx
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
		if addrStr != receiverAddress {
			Logger.log.Error("BTC-TxProof is invalid")
			return false, nil, "", 0, errors.New("BTC-TxProof is invalid")
		}
		if externalFee == 0 {
			tmp := p.ConvertExternalToIncAmount(uint64(outputs[idx].Value))
			if unshieldAmt <= tmp {
				Logger.log.Errorf("[portal] Calculate external fee error")
				return false, nil, "", 0, fmt.Errorf("[portal] Calculate external fee error")
			}
			externalFee = (unshieldAmt - tmp)
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

func (p PortalBTCTokenProcessor) GetExternalTxHashFromProof(proof string) (string, error) {
	// parse BTCProof in meta
	btcTxProof, err := btcrelaying.ParseBTCProofFromB64EncodeStr(proof)
	if err != nil {
		Logger.log.Errorf("ShieldingProof is invalid %v\n", err)
		return "", fmt.Errorf("ShieldingProof is invalid %v\n", err)
	}

	return btcTxProof.BTCTx.TxHash().String(), nil
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
		return []byte{}, "", fmt.Errorf("Invalid signature requirment")
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
func (p PortalBTCTokenProcessor) CreateRawExternalTx(inputs []*statedb.UTXO, outputs []*OutputTx, feePerOutput uint64, bc metadata.ChainRetriever) (string, string, error) {
	msgTx := wire.NewMsgTx(wire.TxVersion)

	// convert feePerOutput from inc unit to external unit
	feePerOutput = p.ConvertIncToExternalAmount(feePerOutput)

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
		txIn.Sequence = uint32(feePerOutput)
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
		if outAmountInExternal <= feePerOutput {
			Logger.log.Errorf("[CreateRawExternalTx-BTC] Output amount %v must greater than fee %v", out.Amount, feePerOutput)
			return "", "", fmt.Errorf("[CreateRawExternalTx-BTC] Output amount %v must greater than fee %v", out.Amount, feePerOutput)
		}
		redeemTxOut := wire.NewTxOut(int64(outAmountInExternal-feePerOutput), destinationAddrByte)
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
		multiSigAddress := bc.GetPortalV4GeneralMultiSigAddress(common.PortalBTCIDStr, 0)
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
		sig, err := txscript.RawTxInWitnessSignature(msgTx, txscript.NewTxSigHashes(msgTx), i, int64(inputs[i].GetOutputAmount()), multiSigScript, txscript.SigHashAll, btcPrivateKey)
		if err != nil {
			return nil, "", fmt.Errorf("[PartSignOnRawExternalTx] Error when signing on raw btc tx: %v", err)
		}
		sigs = append(sigs, sig)
	}

	return sigs, msgTx.TxHash().String(), nil
}

func (p PortalBTCTokenProcessor) IsAcceptableTxSize(num_utxos int, num_unshield_id int) bool {
	// TODO: do experiments depend on external chain miner's habit
	A := 1 // vsize = 192.25 for input-UTXO P2WSH 5-7
	B := 1 // max_vsize = 43 for UTXOs send to P2WSH address
	C := 6 // max_vsize of a transaction in byte ~ 10 KB
	return A*num_utxos+B*num_unshield_id <= C
}

// Choose list of pairs (UTXOs and unshield IDs) for broadcast external transactions
func (p PortalBTCTokenProcessor) ChooseUnshieldIDsFromCandidates(
	utxos map[string]*statedb.UTXO,
	waitingUnshieldReqs map[string]*statedb.WaitingUnshieldRequest,
	tinyAmount uint64) []*BroadcastTx {
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
	utxo_idx := 0
	unshield_idx := 0
	tiny_utxo_used := 0
	for utxo_idx < len(utxos)-tiny_utxo_used && unshield_idx < len(wReqsArr) {
		// utxo_idx always increases at least 1 in this scope

		chosenUTXOs := []*statedb.UTXO{}
		chosenUnshieldIDs := []string{}

		cur_sum_amount := uint64(0)
		cnt := 0
		if utxosArr[utxo_idx].value.GetOutputAmount() >= wReqsArr[unshield_idx].value.GetAmount() {
			// find the last unshield idx that the cummulative sum of unshield amount <= current utxo amount
			for unshield_idx < len(wReqsArr) && cur_sum_amount+wReqsArr[unshield_idx].value.GetAmount() <= utxosArr[utxo_idx].value.GetOutputAmount() && p.IsAcceptableTxSize(1, cnt+1) {
				cur_sum_amount += wReqsArr[unshield_idx].value.GetAmount()
				chosenUnshieldIDs = append(chosenUnshieldIDs, wReqsArr[unshield_idx].value.GetUnshieldID())
				unshield_idx += 1
				cnt += 1
			}
			chosenUTXOs = append(chosenUTXOs, utxosArr[utxo_idx].value)
			utxo_idx += 1 // utxo_idx increases
		} else {
			// find the first utxo idx that the cummulative sum of utxo amount >= current unshield amount
			for utxo_idx < len(utxos)-tiny_utxo_used && cur_sum_amount+utxosArr[utxo_idx].value.GetOutputAmount() < wReqsArr[unshield_idx].value.GetAmount() {
				cur_sum_amount += utxosArr[utxo_idx].value.GetOutputAmount()
				chosenUTXOs = append(chosenUTXOs, utxosArr[utxo_idx].value)
				utxo_idx += 1 // utxo_idx increases
				cnt += 1
			}
			if utxo_idx < len(utxos)-tiny_utxo_used && p.IsAcceptableTxSize(cnt+1, 1) {
				cur_sum_amount += utxosArr[utxo_idx].value.GetOutputAmount()
				chosenUTXOs = append(chosenUTXOs, utxosArr[utxo_idx].value)
				utxo_idx += 1
				cnt += 1

				new_cnt := 0
				target := cur_sum_amount
				cur_sum_amount = 0

				// insert new unshield IDs if the current utxos still has enough amount
				for unshield_idx < len(wReqsArr) && cur_sum_amount+wReqsArr[unshield_idx].value.GetAmount() <= target && p.IsAcceptableTxSize(cnt, new_cnt+1) {
					cur_sum_amount += wReqsArr[unshield_idx].value.GetAmount()
					chosenUnshieldIDs = append(chosenUnshieldIDs, wReqsArr[unshield_idx].value.GetUnshieldID())
					unshield_idx += 1
					new_cnt += 1
				}

			} else {
				// not enough utxo for last unshield IDs
				break
			}
		}

		// use a tiny UTXO
		if utxo_idx < len(utxos)-tiny_utxo_used && utxosArr[len(utxos)-tiny_utxo_used-1].value.GetOutputAmount() <= tinyAmount {
			tiny_utxo_used += 1
			chosenUTXOs = append(chosenUTXOs, utxosArr[len(utxos)-tiny_utxo_used].value)
		}

		// merge small batches
		if len(broadcastTxs) > 0 {
			prevUTXOs := broadcastTxs[len(broadcastTxs)-1].UTXOs
			prevRequests := broadcastTxs[len(broadcastTxs)-1].UnshieldIDs
			lenUTXOs := len(prevUTXOs) + len(chosenUTXOs)
			lenRequests := len(prevRequests) + len(chosenUnshieldIDs)
			if p.IsAcceptableTxSize(lenUTXOs, lenRequests) {
				broadcastTxs[len(broadcastTxs)-1] = &BroadcastTx{
					UTXOs:       append(prevUTXOs, chosenUTXOs...),
					UnshieldIDs: append(prevRequests, chosenUnshieldIDs...),
				}
				continue
			}
		}
		broadcastTxs = append(broadcastTxs, &BroadcastTx{
			UTXOs:       chosenUTXOs,
			UnshieldIDs: chosenUnshieldIDs,
		})
	}
	return broadcastTxs
}
