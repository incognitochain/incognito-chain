package transaction

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"math"
	"math/big"
	"sort"
	"strconv"

	"github.com/ninjadotorg/cash/cashec"
	"github.com/ninjadotorg/cash/common"
	"github.com/ninjadotorg/cash/privacy/client"
)

// TxTokenVin ...
type TxTokenVin struct {
	Vout      int
	Signature string
	PubKey    string
}

// TxTokenVout ...
type TxTokenVout struct {
	Value       uint64
	ScripPubKey string
}

// TxToken ...
type TxToken struct {
	Version         byte
	Type            byte
	PropertyName    string
	PropertySymbol  string
	Vin             []TxTokenVin
	Vout            []TxTokenVout
	TxCustomTokenID hash.Hash
	Amount          float64
}

// TxCustomToken ...
type TxCustomToken struct {
	Tx

	TxToken TxToken
}

type CustomTokenReceiver struct {
	PubKey string  `json:"PubKey"`
	Amount float64 `json:"Amount"`
}

// CustomTokenParamTx ...
type CustomTokenParamTx struct {
	PropertyName    string                `json:"TokenName"`
	PropertySymbol  string                `json:"TokenSymbol"`
	TxCustomTokenID string                `json:"TokenHash"`
	Amount          float64               `json:"TokenAmount"`
	TokenTxType     float64               `json:"TokenTxType"`
	Receivers       []CustomTokenReceiver `json:"TokenReceivers"`
}

func CreateCustomTokenReceiverArray(data interface{}) []CustomTokenReceiver {
	result := make([]CustomTokenReceiver, 0)
	receivers := data.([]map[string]interface{})
	for _, item := range receivers {
		resultItem := CustomTokenReceiver{
			PubKey: item["PubKey"].(string),
			Amount: item["Amount"].(float64),
		}
		result = append(result, resultItem)
	}
	return result
}

// CreateEmptyCustomTokenTx - return an init custom token transaction
func CreateEmptyCustomTokenTx() (*TxCustomToken, error) {
	emptyTx, err := CreateEmptyTx(common.TxCustomTokenType)

	if err != nil {
		return nil, err
	}

	txToken := TxToken{}

	txCustomToken := &TxCustomToken{
		Tx:      *emptyTx,
		TxToken: txToken,
	}
	return txCustomToken, nil
}

// SetTxID ...
func (tx *TxCustomToken) SetTxID(txId *common.Hash) {
	tx.Tx.txId = txId
}

// GetTxID ...
func (tx *TxCustomToken) GetTxID() *common.Hash {
	return tx.Tx.txId
}

// Hash returns the hash of all fields of the transaction
func (tx TxCustomToken) Hash() *common.Hash {
	record := strconv.Itoa(int(tx.Tx.Version))
	record += tx.Tx.Type
	record += strconv.FormatInt(tx.Tx.LockTime, 10)
	record += strconv.FormatUint(tx.Tx.Fee, 10)
	record += strconv.Itoa(len(tx.Tx.Descs))
	for _, desc := range tx.Tx.Descs {
		record += desc.toString()
	}
	record += string(tx.Tx.JSPubKey)
	// record += string(tx.JSSig)
	record += string(tx.Tx.AddressLastByte)
	record += tx.TxToken.PropertyName
	record += tx.TxToken.PropertySymbol
	record += fmt.Sprintf("%f", tx.TxToken.Amount)
	record += base64.URLEncoding.EncodeToString(tx.TxToken.TxCustomTokenID.Sum(nil))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

// ValidateTransaction ...
func (tx *TxCustomToken) ValidateTransaction() bool {
	if tx.Tx.ValidateTransaction() {
		return true
	}
	return false
}

// GetType returns the type of the transaction
func (tx *TxCustomToken) GetType() string {
	return tx.Tx.Type
}

// GetTxVirtualSize computes the virtual size of a given transaction
func (tx *TxCustomToken) GetTxVirtualSize() uint64 {
	var sizeVersion uint64 = 1  // int8
	var sizeType uint64 = 8     // string
	var sizeLockTime uint64 = 8 // int64
	var sizeFee uint64 = 8      // uint64
	var sizeDescs = uint64(max(1, len(tx.Tx.Descs))) * EstimateJSDescSize()
	var sizejSPubKey uint64 = 64      // [64]byte
	var sizejSSig uint64 = 64         // [64]byte
	var sizeTokenName uint64 = 64     // string
	var sizeTokenSymbol uint64 = 64   // string
	var sizeTokenHash uint64 = 64     // string
	var sizeTokenAmount uint64 = 64   // string
	var sizeTokenTxType uint64 = 64   // string
	var sizeTokenReceiver uint64 = 64 // string

	estimateTxSizeInByte := sizeVersion
	estimateTxSizeInByte += sizeType
	estimateTxSizeInByte += sizeLockTime
	estimateTxSizeInByte += sizeFee
	estimateTxSizeInByte += sizeDescs
	estimateTxSizeInByte += sizejSPubKey
	estimateTxSizeInByte += sizejSSig
	estimateTxSizeInByte += sizeTokenName
	estimateTxSizeInByte += sizeTokenSymbol
	estimateTxSizeInByte += sizeTokenHash
	estimateTxSizeInByte += sizeTokenAmount
	estimateTxSizeInByte += sizeTokenTxType
	estimateTxSizeInByte += sizeTokenReceiver
	return uint64(math.Ceil(float64(estimateTxSizeInByte) / 1024))
}

func (tx *TxCustomToken) GetTxFee() uint64 {
	return tx.Fee
}

// GetSenderAddrLastByte ...
func (tx *TxCustomToken) GetSenderAddrLastByte() byte {
	return tx.Tx.AddressLastByte
}

// CreateTxCustomToken ...
func CreateTxCustomToken(senderKey *client.SpendingKey,
	paymentInfo []*client.PaymentInfo,
	rts map[byte]*common.Hash,
	usableTx map[byte][]*Tx,
	nullifiers map[byte]([][]byte),
	commitments map[byte]([][]byte),
	fee uint64,
	assetType string,
	senderChainID byte,
	tokenParams *CustomTokenParamTx) (*TxCustomToken, error) {
	fee = 0 // TODO remove this line
	fmt.Printf("List of all commitments before building tx:\n")
	fmt.Printf("rts: %+v\n", rts)
	for _, cm := range commitments {
		fmt.Printf("%x\n", cm)
	}

	var value uint64
	for _, p := range paymentInfo {
		value += p.Amount
		fmt.Printf("[CreateTx] paymentInfo.Value: %+v, paymentInfo.Apk: %x\n", p.Amount, p.PaymentAddress.Apk)
	}

	type ChainNote struct {
		note    *client.Note
		chainID byte
	}

	// Get list of notes to use
	var inputNotes []*ChainNote
	for chainID, chainTxs := range usableTx {
		for _, tx := range chainTxs {
			for _, desc := range tx.Descs {
				for _, note := range desc.Note {
					chainNote := &ChainNote{note: note, chainID: chainID}
					inputNotes = append(inputNotes, chainNote)
					fmt.Printf("[CreateTx] inputNote.Value: %+v\n", note.Value)
				}
			}
		}
	}

	// Left side value
	var sumInputValue uint64
	for _, chainNote := range inputNotes {
		sumInputValue += chainNote.note.Value
	}
	if sumInputValue < value+fee {
		return nil, fmt.Errorf("Input value less than output value")
	}

	senderFullKey := cashec.KeySet{}
	senderFullKey.ImportFromPrivateKeyByte(senderKey[:])

	// Create tx before adding js descs
	tx, err := CreateEmptyCustomTokenTx()
	if err != nil {
		return nil, err
	}
	tempKeySet := cashec.KeySet{}
	tempKeySet.ImportFromPrivateKey(senderKey)
	lastByte := tempKeySet.PublicKey.Apk[len(tempKeySet.PublicKey.Apk)-1]
	tx.Tx.AddressLastByte = lastByte
	var latestAnchor map[byte][]byte

	for len(inputNotes) > 0 || len(paymentInfo) > 0 {
		// Sort input and output notes ascending by value to start building js descs
		sort.Slice(inputNotes, func(i, j int) bool {
			return inputNotes[i].note.Value < inputNotes[j].note.Value
		})
		sort.Slice(paymentInfo, func(i, j int) bool {
			return paymentInfo[i].Amount < paymentInfo[j].Amount
		})

		// Choose inputs to build js desc
		// var inputsToBuildWitness, inputs []*client.JSInput
		inputsToBuildWitness := make(map[byte][]*client.JSInput)
		inputs := make(map[byte][]*client.JSInput)
		inputValue := uint64(0)
		numInputNotes := 0
		for len(inputNotes) > 0 && len(inputs) < NumDescInputs {
			input := &client.JSInput{}
			chainNote := inputNotes[len(inputNotes)-1] // Get note with largest value
			input.InputNote = chainNote.note
			input.Key = senderKey
			inputs[chainNote.chainID] = append(inputs[chainNote.chainID], input)
			inputsToBuildWitness[chainNote.chainID] = append(inputsToBuildWitness[chainNote.chainID], input)
			inputValue += input.InputNote.Value

			inputNotes = inputNotes[:len(inputNotes)-1]
			numInputNotes++
			fmt.Printf("Choose input note with value %+v and cm %x\n", input.InputNote.Value, input.InputNote.Cm)
		}

		var feeApply uint64 // Zero fee for js descs other than the first one
		if len(tx.Tx.Descs) == 0 {
			// First js desc, applies fee
			feeApply = fee
			tx.Fee = fee
		}
		if len(tx.Tx.Descs) == 0 {
			if inputValue < feeApply {
				return nil, fmt.Errorf("Input note values too small to pay fee")
			}
			inputValue -= feeApply
		}

		// Add dummy input note if necessary
		for numInputNotes < NumDescInputs {
			input := &client.JSInput{}
			input.InputNote = createDummyNote(senderKey)
			input.Key = senderKey
			input.WitnessPath = (&client.MerklePath{}).CreateDummyPath() // No need to build commitment merkle path for dummy note
			dummyNoteChainID := senderChainID                            // Dummy note's chain is the same as sender's
			inputs[dummyNoteChainID] = append(inputs[dummyNoteChainID], input)
			numInputNotes++
			fmt.Printf("Add dummy input note\n")
		}

		// Check if input note's cm is in commitments list
		for chainID, chainInputs := range inputsToBuildWitness {
			for _, input := range chainInputs {
				input.InputNote.Cm = client.GetCommitment(input.InputNote)

				found := false
				for _, c := range commitments[chainID] {
					if bytes.Equal(c, input.InputNote.Cm) {
						found = true
					}
				}
				if found == false {
					return nil, fmt.Errorf("Commitment %x of input note isn't in commitments list of chain %d", input.InputNote.Cm, chainID)
				}
			}
		}

		// Build witness path for the input notes
		newRts, err := client.BuildWitnessPathMultiChain(inputsToBuildWitness, commitments)
		if err != nil {
			return nil, err
		}

		// For first js desc, check if provided Rt is the root of the merkle tree contains all commitments
		if latestAnchor == nil {
			for chainID, rt := range newRts {
				if !bytes.Equal(rt, rts[chainID][:]) {
					return nil, fmt.Errorf("Provided anchor doesn't match commitments list: %d %x %x", chainID, rt, rts[chainID][:])
				}
			}
		}
		latestAnchor = newRts
		// Add dummy anchor to for dummy inputs
		if len(latestAnchor[senderChainID]) == 0 {
			latestAnchor[senderChainID] = make([]byte, 32)
		}

		// Choose output notes for the js desc
		outputs := []*client.JSOutput{}
		for len(paymentInfo) > 0 && len(outputs) < NumDescOutputs-1 && inputValue > 0 { // Leave out 1 output note for change
			p := paymentInfo[len(paymentInfo)-1]
			var outNote *client.Note
			var encKey client.TransmissionKey
			if p.Amount <= inputValue { // Enough for one more output note, include it
				outNote = &client.Note{Value: p.Amount, Apk: p.PaymentAddress.Apk}
				encKey = p.PaymentAddress.Pkenc
				inputValue -= p.Amount
				paymentInfo = paymentInfo[:len(paymentInfo)-1]
				fmt.Printf("Use output value %+v => %x\n", outNote.Value, outNote.Apk)
			} else { // Not enough for this note, send some and save the rest for next js desc
				outNote = &client.Note{Value: inputValue, Apk: p.PaymentAddress.Apk}
				encKey = p.PaymentAddress.Pkenc
				paymentInfo[len(paymentInfo)-1].Amount = p.Amount - inputValue
				inputValue = 0
				fmt.Printf("Partially send %+v to %x\n", outNote.Value, outNote.Apk)
			}

			output := &client.JSOutput{EncKey: encKey, OutputNote: outNote}
			outputs = append(outputs, output)
		}

		if inputValue > 0 {
			// Still has some room left, check if one more output note is possible to add
			var p *client.PaymentInfo
			if len(paymentInfo) > 0 {
				p = paymentInfo[len(paymentInfo)-1]
			}

			if p != nil && p.Amount == inputValue {
				// Exactly equal, add this output note to js desc
				outNote := &client.Note{Value: p.Amount, Apk: p.PaymentAddress.Apk}
				output := &client.JSOutput{EncKey: p.PaymentAddress.Pkenc, OutputNote: outNote}
				outputs = append(outputs, output)
				paymentInfo = paymentInfo[:len(paymentInfo)-1]
				fmt.Printf("Exactly enough, include 1 more output %+v, %x\n", outNote.Value, outNote.Apk)
			} else {
				// Cannot put the output note into this js desc, create a change note instead
				outNote := &client.Note{Value: inputValue, Apk: senderFullKey.PublicKey.Apk}
				output := &client.JSOutput{EncKey: senderFullKey.PublicKey.Pkenc, OutputNote: outNote}
				outputs = append(outputs, output)
				fmt.Printf("Create change outnote %+v, %x\n", outNote.Value, outNote.Apk)

				// Use the change note to continually send to receivers if needed
				if len(paymentInfo) > 0 {
					// outNote data (R and Rho) will be updated when building zk-proof
					chainNote := &ChainNote{note: outNote, chainID: senderChainID}
					inputNotes = append(inputNotes, chainNote)
					fmt.Printf("Reuse change note later\n")
				}
			}
			inputValue = 0
		}

		// Add dummy output note if necessary
		for len(outputs) < NumDescOutputs {
			outputs = append(outputs, CreateRandomJSOutput())
			fmt.Printf("Create dummy output note\n")
		}

		// TODO: Shuffle output notes randomly (if necessary)

		// Generate proof and sign tx
		var reward uint64 // Zero reward for non-salary transaction
		err = tx.BuildNewJSDesc(inputs, outputs, latestAnchor, reward, feeApply, assetType, false)
		if err != nil {
			return nil, err
		}

		// Add new commitments to list to use in next js desc if needed
		for _, output := range outputs {
			fmt.Printf("Add new output cm to list: %x\n", output.OutputNote.Cm)
			commitments[senderChainID] = append(commitments[senderChainID], output.OutputNote.Cm)
		}

		fmt.Printf("Len input and info: %+v %+v\n", len(inputNotes), len(paymentInfo))
	}

	// Sign tx
	tx, err = SignTxCustomToken(tx)
	if err != nil {
		return nil, err
	}

	fmt.Printf("jspubkey: %x\n", tx.JSPubKey)
	fmt.Printf("jssig: %x\n", tx.JSSig)
	return tx, nil
}

// SignTxCustomToken ...
func SignTxCustomToken(tx *TxCustomToken) (*TxCustomToken, error) {
	//Check input transaction
	if tx.Tx.JSSig != nil {
		return nil, errors.New("Inpusut transaction must be an unsigned one")
	}

	// Hash transaction
	tx.SetTxID(tx.Hash())
	hash := tx.GetTxID()
	data := make([]byte, common.HashSize)
	copy(data, hash[:])

	// Sign
	ecdsaSignature := new(client.EcdsaSignature)
	var err error
	ecdsaSignature.R, ecdsaSignature.S, err = client.Sign(rand.Reader, tx.Tx.sigPrivKey, data[:])
	if err != nil {
		return nil, err
	}

	//Signature 64 bytes
	tx.Tx.JSSig = JSSigToByteArray(ecdsaSignature)

	return tx, nil
}

// VerifySignCustomTokenTx ...
func VerifySignCustomTokenTx(tx *TxCustomToken) (bool, error) {
	//Check input transaction
	if tx.Tx.JSSig == nil || tx.Tx.JSPubKey == nil {
		return false, errors.New("Input transaction must be an signed one!")
	}

	// UnParse Public key
	pubKey := new(client.PublicKey)
	pubKey.X = new(big.Int).SetBytes(tx.Tx.JSPubKey[0:32])
	pubKey.Y = new(big.Int).SetBytes(tx.Tx.JSPubKey[32:64])

	// UnParse ECDSA signature
	ecdsaSignature := new(client.EcdsaSignature)
	ecdsaSignature.R = new(big.Int).SetBytes(tx.Tx.JSSig[0:32])
	ecdsaSignature.S = new(big.Int).SetBytes(tx.Tx.JSSig[32:64])

	// Hash origin transaction
	hash := tx.GetTxID()
	data := make([]byte, common.HashSize)
	copy(data, hash[:])

	valid := client.VerifySign(pubKey, data[:], ecdsaSignature.R, ecdsaSignature.S)
	return valid, nil
}
