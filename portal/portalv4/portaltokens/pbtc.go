package portaltokens

import (
	"bytes"
	"crypto/ed25519"
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
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
)

type PortalBTCTokenProcessor struct {
	*PortalToken
}

func genBTCPrivateKey(IncKeyBytes []byte) []byte {
	BTCKeyBytes := ed25519.NewKeyFromSeed(IncKeyBytes)[32:]
	return BTCKeyBytes
}

func (p PortalBTCTokenProcessor) GetExpectedMemoForShielding(incAddress string) string {
	return "PS1-" + incAddress
}

func (p PortalBTCTokenProcessor) ConvertExternalToIncAmount(externalAmt uint64) uint64 {
	return externalAmt * 10
}

func (p PortalBTCTokenProcessor) ConvertIncToExternalAmount(incAmt uint64) uint64 {
	return incAmt / 10 // incAmt / 10^9 * 10^8
}

func (p PortalBTCTokenProcessor) parseAndVerifyProofBTCChain(
	proof string, btcChain *btcrelaying.BlockChain, expectedMemo string, expectedMultisigAddress string) (bool, []*statedb.UTXO, error) {
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

	// extract attached message from txOut's OP_RETURN
	btcAttachedMsg, err := btcrelaying.ExtractAttachedMsgFromTx(btcTxProof.BTCTx)
	if err != nil {
		Logger.log.Errorf("Could not extract attached message from BTC tx proof with err: %v", err)
		return false, nil, fmt.Errorf("Could not extract attached message from BTC tx proof with err: %v", err)
	}
	if btcAttachedMsg != expectedMemo {
		Logger.log.Errorf("ShieldingIncAddress in the btc attached message is not matched with ShieldingIncAddress in metadata")
		return false, nil, fmt.Errorf("ShieldingIncAddress in the btc attached message %v is not matched with ShieldingIncAddress in metadata %v", btcAttachedMsg, expectedMemo)
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
		))
	}

	if len(listUTXO) == 0 || p.ConvertExternalToIncAmount(totalValue) < p.GetMinTokenAmount() {
		Logger.log.Errorf("Shielding amount: %v is less than the minimum threshold: %v\n", totalValue, p.GetMinTokenAmount())
		return false, nil, fmt.Errorf("Shielding amount: %v is less than the minimum threshold: %v", totalValue, p.GetMinTokenAmount())
	}

	return true, listUTXO, nil
}

func (p PortalBTCTokenProcessor) ParseAndVerifyProof(
	proof string, bc metadata.ChainRetriever, expectedMemo string, expectedMultisigAddress string) (bool, []*statedb.UTXO, error) {
	btcChain := bc.GetBTCHeaderChain()
	return p.parseAndVerifyProofBTCChain(proof, btcChain, expectedMemo, expectedMultisigAddress)
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

// generate multisig wallet address from seeds (seed is mining key of beacon validator in byte array)
func GenerateMultiSigWalletFromSeeds(chainParams *chaincfg.Params, seeds [][]byte, numSigsRequired int) ([]byte, []string, string, error) {
	if len(seeds) < numSigsRequired || numSigsRequired < 0 {
		return nil, nil, "", errors.New("Invalid signature requirment")
	}
	bitcoinPrvKeys := make([]*btcec.PrivateKey, 0)
	bitcoinPrvKeyStrs := make([]string, 0) // btc private key hex encoded
	// create redeem script for 2 of 3 multi-sig
	builder := txscript.NewScriptBuilder()
	// add the minimum number of needed signatures
	builder.AddOp(byte(txscript.OP_1 - 1 + numSigsRequired))
	for _, seed := range seeds {
		BTCKeyBytes := genBTCPrivateKey(seed)
		privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), BTCKeyBytes)
		// add the public key to redeem script
		builder.AddData(privKey.PubKey().SerializeCompressed())
		bitcoinPrvKeys = append(bitcoinPrvKeys, privKey)
		bitcoinPrvKeyStrs = append(bitcoinPrvKeyStrs, hex.EncodeToString(privKey.Serialize()))
	}
	// add the total number of public keys in the multi-sig script
	builder.AddOp(byte(txscript.OP_1 - 1 + len(seeds)))
	// add the check-multi-sig op-code
	builder.AddOp(txscript.OP_CHECKMULTISIG)
	// redeem script is the script program in the format of []byte
	redeemScript, err := builder.Script()
	if err != nil {
		return nil, nil, "", err
	}
	// generate multisig address
	multiAddress := btcutil.Hash160(redeemScript)
	addr, err := btcutil.NewAddressScriptHashFromHash(multiAddress, chainParams)

	return redeemScript, bitcoinPrvKeyStrs, addr.String(), nil
}

// GeneratePartPrivateKeyFromSeed generate private key from seed
// return the private key serialized in bytes array
func (p PortalBTCTokenProcessor) GeneratePrivateKeyFromSeed(seed []byte) ([]byte, error) {
	if len(seed) == 0 {
		return nil, errors.New("Invalid seed")
	}

	btcKeyBytes := genBTCPrivateKey(seed)
	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), btcKeyBytes)

	return privKey.Serialize(), nil
}

// CreateRawExternalTx creates raw btc transaction (not include signatures of beacon validator)
// inputs: UTXO state of beacon, unit of amount in btc
// outputs: unit of amount in pbtc ~ unshielding amount
// feePerOutput: unit in pbtc
func (p PortalBTCTokenProcessor) CreateRawExternalTx(inputs []*statedb.UTXO, outputs []*OutputTx, feePerOutput uint64, memo string, bc metadata.ChainRetriever) (string, string, error) {
	msgTx := wire.NewMsgTx(wire.TxVersion)

	// add TxIns into raw tx
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
	totalOutputAmount := uint64(0)
	for _, out := range outputs {
		// adding the output to tx
		decodedAddr, err := btcutil.DecodeAddress(out.ReceiverAddress, bc.GetBTCChainParams())
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
		if out.Amount <= feePerOutput {
			Logger.log.Errorf("[CreateRawExternalTx-BTC] Output amount %v must greater than fee %v", out.Amount, feePerOutput)
			return "", "", fmt.Errorf("[CreateRawExternalTx-BTC] Output amount %v must greater than fee %v", out.Amount, feePerOutput)
		}
		redeemTxOut := wire.NewTxOut(int64(p.ConvertIncToExternalAmount(out.Amount-feePerOutput)), destinationAddrByte)
		msgTx.AddTxOut(redeemTxOut)
		totalOutputAmount += out.Amount
	}
	totalOutputAmount = p.ConvertIncToExternalAmount(totalOutputAmount)

	// calculate the change output
	if totalInputAmount-totalOutputAmount > 0 {
		// adding the output to tx
		multiSigAddress := bc.GetPortalV4MultiSigAddress(common.PortalBTCIDStr, 0)
		decodedAddr, err := btcutil.DecodeAddress(multiSigAddress, bc.GetBTCChainParams())
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

	// add memo into raw tx
	script := append([]byte{txscript.OP_RETURN}, byte(len([]byte(memo))))
	script = append(script, []byte(memo)...)
	msgTx.AddTxOut(wire.NewTxOut(0, script))

	var rawTxBytes bytes.Buffer
	err := msgTx.Serialize(&rawTxBytes)
	if err != nil {
		Logger.log.Errorf("[CreateRawExternalTx-BTC] Error when serializing raw tx: %v", err)
		return "", "", err
	}

	hexRawTx := hex.EncodeToString(rawTxBytes.Bytes())
	return hexRawTx, msgTx.TxHash().String(), nil
}

func (p PortalBTCTokenProcessor) PartSignOnRawExternalTx(seedKey []byte, multiSigScript []byte, rawTxBytes []byte) ([][]byte, string, error) {
	// new MsgTx from rawTxBytes
	msgTx := new(btcwire.MsgTx)
	rawTxBuffer := bytes.NewBuffer(rawTxBytes)
	err := msgTx.Deserialize(rawTxBuffer)
	if err != nil {
		return nil, "", fmt.Errorf("[PartSignOnRawExternalTx] Error when deserializing raw tx bytes: %v", err)
	}

	// sign on each TxIn
	sigs := [][]byte{}
	for i := range msgTx.TxIn {
		signature := txscript.NewScriptBuilder()
		signature.AddOp(txscript.OP_FALSE)

		// generate btc private key from seed: private key of bridge consensus
		btcPrivateKeyBytes, err := p.GeneratePrivateKeyFromSeed(seedKey)
		if err != nil {
			return nil, "", fmt.Errorf("[PartSignOnRawExternalTx] Error when generate btc private key from seed: %v", err)
		}
		btcPrivateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), btcPrivateKeyBytes)
		sig, err := txscript.RawTxInSignature(msgTx, i, multiSigScript, txscript.SigHashAll, btcPrivateKey)
		if err != nil {
			return nil, "", fmt.Errorf("[PartSignOnRawExternalTx] Error when signing on raw btc tx: %v", err)
		}
		sigs = append(sigs, sig)
	}

	return sigs, msgTx.TxHash().String(), nil
}

func (p PortalBTCTokenProcessor) IsAcceptableTxSize(num_utxos int, num_unshield_id int) bool {
	// TODO: do experiments depend on external chain miner's habit
	A := 1 // input size (include sig size) in byte
	B := 1 // output size in byte
	C := 6 // max transaction size in byte ~ 10 KB
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

func (p PortalBTCTokenProcessor) ParseAndVerifyUnshieldProof(
	proof string, bc metadata.ChainRetriever, expectedMemo string, expectedMultisigAddress string, expectPaymentInfo map[string]uint64, utxos []*statedb.UTXO) (bool, []*statedb.UTXO, error) {
	btcChain := bc.GetBTCHeaderChain()
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

	// extract attached message from txOut's OP_RETURN
	btcAttachedMsg, err := btcrelaying.ExtractAttachedMsgFromTx(btcTxProof.BTCTx)
	if err != nil {
		Logger.log.Errorf("Could not extract attached message from BTC tx proof with err: %v", err)
		return false, nil, fmt.Errorf("Could not extract attached message from BTC tx proof with err: %v", err)
	}
	if btcAttachedMsg != expectedMemo {
		Logger.log.Errorf("ShieldingId in the btc attached message is not matched with portingID in metadata")
		return false, nil, fmt.Errorf("ShieldingId in the btc attached message %v is not matched with portingID in metadata %v", btcAttachedMsg, expectedMemo)
	}

	// verify spent outputs
	if len(btcTxProof.BTCTx.TxIn) < 1 {
		Logger.log.Errorf("Can not find the tx inputs in proof")
		return false, nil, fmt.Errorf("Submit confirmed tx: no tx inputs in proof")
	}
	isMatched := false
	for _, input := range btcTxProof.BTCTx.TxIn {
		for _, v := range utxos {
			if v.GetTxHash() == input.PreviousOutPoint.Hash.String() && v.GetOutputIndex() == input.PreviousOutPoint.Index {
				isMatched = true
				break
			}
		}
		if !isMatched {
			Logger.log.Errorf("Submit confirmed: tx inputs from proof is diff utxos from unshield batch")
			return false, nil, fmt.Errorf("Submit confirmed tx: tx inputs from proof is diff utxos from unshield batch")
		}
		isMatched = false
	}

	// check whether amount transfer in txBNB is equal porting amount or not
	// check receiver and amount in tx
	outputs := btcTxProof.BTCTx.TxOut
	for receiverAddress := range expectPaymentInfo {
		isMatched = false
		for _, out := range outputs {
			addrStr, err := btcChain.ExtractPaymentAddrStrFromPkScript(out.PkScript)
			if err != nil {
				Logger.log.Errorf("[portal] ExtractPaymentAddrStrFromPkScript: could not extract payment address string from pkscript with err: %v\n", err)
				continue
			}
			if addrStr != receiverAddress {
				continue
			}
			isMatched = true
			break
		}
		if !isMatched {
			Logger.log.Error("BTC-TxProof is invalid")
			return false, nil, errors.New("BTC-TxProof is invalid")
		}
	}

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
		listUTXO = append(listUTXO, statedb.NewUTXOWithValue(
			addrStr,
			btcTxProof.BTCTx.TxHash().String(),
			uint32(idx),
			uint64(out.Value),
		))
	}

	return true, listUTXO, nil
}
