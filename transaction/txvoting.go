package transaction

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy/client"
)

// TxVoting ...
type TxVoting struct {
	Tx
	PublicKey string
}

// CreateEmptyVotingTx - return an init tv voting
func CreateEmptyVotingTx(pubkey string) (*TxVoting, error) {
	emptyTx, err := CreateEmptyTx(common.TxVotingType)
	if err != nil {
		return nil, err
	}
	txVoting := &TxVoting{
		Tx:        *emptyTx,
		PublicKey: pubkey,
	}
	return txVoting, nil
}

func (tx *TxVoting) GetValue() uint64 {
	val := uint64(0)
	for _, desc := range tx.Descs {
		for _, note := range desc.Note {
			val += note.Value
		}
	}
	return val
}

func (tx *TxVoting) SetTxID(txId *common.Hash) {
	tx.Tx.txId = txId
}

func (tx *TxVoting) GetTxID() *common.Hash {
	return tx.Tx.txId
}

// Hash returns the hash of all fields of the transaction
func (tx TxVoting) Hash() *common.Hash {
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
	record += tx.PublicKey
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

// ValidateTransaction returns true if transaction is valid:
// - Signature matches the signing public key
// - JSDescriptions are valid (zk-snark proof satisfied)
// Note: This method doesn't check for double spending
func (tx *TxVoting) ValidateTransaction() bool {
	if tx.Tx.ValidateTransaction() {
		return true
	}

	// TODO: check the burnt money is sufficient or not
	var receviedAddr client.SpendingAddress
	for _, desc := range tx.Descs {
		for _, note := range desc.Note {
			if note.Apk == receviedAddr {
				return true
			}
		}
	}

	return false
}

// GetType returns the type of the transaction
func (tx *TxVoting) GetType() string {
	return tx.Tx.Type
}

// GetTxVirtualSize computes the virtual size of a given transaction
func (tx *TxVoting) GetTxVirtualSize() uint64 {
	var sizeVersion uint64 = 1  // int8
	var sizeType uint64 = 8     // string
	var sizeLockTime uint64 = 8 // int64
	var sizeFee uint64 = 8      // uint64
	var sizeDescs = uint64(max(1, len(tx.Tx.Descs))) * EstimateJSDescSize()
	var sizejSPubKey uint64 = 64 // [64]byte
	var sizejSSig uint64 = 64    // [64]byte
	var sizeNodeAddr uint64 = 8  // string
	estimateTxSizeInByte := sizeVersion + sizeType + sizeLockTime + sizeFee + sizeDescs + sizejSPubKey + sizejSSig + sizeNodeAddr
	return uint64(math.Ceil(float64(estimateTxSizeInByte) / 1024))
}

func (tx *TxVoting) GetSenderAddrLastByte() byte {
	return tx.Tx.AddressLastByte
}

func (tx *TxVoting) GetTxFee() uint64 {
	return tx.Fee
}

// CreateVotingTx creates transaction with appropriate proof for a private payment
// rts: mapping from the chainID to the root of the commitment merkle tree at current block
// 		(the latest block of the node creating this tx)
func CreateVotingTx(
	senderKey *client.SpendingKey,
	paymentInfo []*client.PaymentInfo,
	rts map[byte]*common.Hash,
	usableTx map[byte][]*Tx,
	nullifiers map[byte]([][]byte),
	commitments map[byte]([][]byte),
	fee uint64,
	senderChainID byte,
	nodeAddr string,
) (*TxVoting, error) {
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
	if sumInputValue < value+fee && false { // TODO remove false
		return nil, fmt.Errorf("Input value less than output value")
	}

	senderFullKey := cashec.KeySet{}
	senderFullKey.ImportFromPrivateKeyByte(senderKey[:])

	// Create tx before adding js descs
	tx, err := CreateEmptyVotingTx(nodeAddr)
	if err != nil {
		return nil, err
	}
	tempKeySet := cashec.KeySet{}
	tempKeySet.ImportFromPrivateKey(senderKey)
	lastByte := tempKeySet.PaymentAddress.Apk[len(tempKeySet.PaymentAddress.Apk)-1]
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

				found := true
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
		for len(paymentInfo) > 0 && len(outputs) < NumDescOutputs-1 && inputValue >= 0 { // Leave out 1 output note for change // TODO remove equal 0
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

		if inputValue >= 0 { // TODO remove equal 0
			// Still has some room left, check if one more output note is possible to add
			var p *client.PaymentInfo
			if len(paymentInfo) > 0 {
				p = paymentInfo[len(paymentInfo)-1]
			}

			if p != nil && p.Amount == inputValue { // TODO remove equal 0
				// Exactly equal, add this output note to js desc
				outNote := &client.Note{Value: p.Amount, Apk: p.PaymentAddress.Apk}
				output := &client.JSOutput{EncKey: p.PaymentAddress.Pkenc, OutputNote: outNote}
				outputs = append(outputs, output)
				paymentInfo = paymentInfo[:len(paymentInfo)-1]
				fmt.Printf("Exactly enough, include 1 more output %+v, %x\n", outNote.Value, outNote.Apk)
			} else {
				// Cannot put the output note into this js desc, create a change note instead
				outNote := &client.Note{Value: inputValue, Apk: senderFullKey.PaymentAddress.Apk}
				output := &client.JSOutput{EncKey: senderFullKey.PaymentAddress.Pkenc, OutputNote: outNote}
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

		// Generate proof and sign tx
		var reward uint64 // Zero reward for non-salary transaction
		err = tx.Tx.BuildNewJSDesc(inputs, outputs, latestAnchor, reward, feeApply, true)
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
	tx, err = SignTxVoting(tx)
	if err != nil {
		return nil, err
	}

	fmt.Printf("jspubkey: %x\n", tx.JSPubKey)
	fmt.Printf("jssig: %x\n", tx.JSSig)
	return tx, nil
}

// SignTxVoting ...
func SignTxVoting(tx *TxVoting) (*TxVoting, error) {
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

func VerifySignVotingTx(tx *TxVoting) (bool, error) {
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
