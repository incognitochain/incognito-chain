package internal

import(
	"encoding/base64"
	// "encoding/json"
	"math/big"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
)

const MaxSizeByte = (1 << 8) - 1

type SigPubKey struct {
	Indexes [][]*big.Int
}
func (sigPub SigPubKey) Bytes() ([]byte, error) {
	n := len(sigPub.Indexes)
	if n == 0 {
		return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is empty")
	}
	if n > MaxSizeByte {
		return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is too large, too many rows")
	}
	m := len(sigPub.Indexes[0])
	if m > MaxSizeByte {
		return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is too large, too many columns")
	}
	for i := 1; i < n; i += 1 {
		if len(sigPub.Indexes[i]) != m {
			return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is not a rectangle array")
		}
	}

	b := make([]byte, 0)
	b = append(b, byte(n))
	b = append(b, byte(m))
	for i := 0; i < n; i += 1 {
		for j := 0; j < m; j += 1 {
			currentByte := sigPub.Indexes[i][j].Bytes()
			lengthByte := len(currentByte)
			if lengthByte > MaxSizeByte {
				return nil, errors.New("TxSigPublicKeyVer2.ToBytes: IndexesByte is too large")
			}
			b = append(b, byte(lengthByte))
			b = append(b, currentByte...)
		}
	}
	return b, nil
}

func (sigPub *SigPubKey) SetBytes(b []byte) error {
	if len(b) < 2 {
		return errors.New("txSigPubKeyFromBytes: cannot parse length of Indexes, length of input byte is too small")
	}
	n := int(b[0])
	m := int(b[1])
	offset := 2
	indexes := make([][]*big.Int, n)
	for i := 0; i < n; i += 1 {
		row := make([]*big.Int, m)
		for j := 0; j < m; j += 1 {
			if offset >= len(b) {
				return errors.New("txSigPubKeyFromBytes: cannot parse byte length of index[i][j], length of input byte is too small")
			}
			byteLength := int(b[offset])
			offset += 1
			if offset+byteLength > len(b) {
				return errors.New("txSigPubKeyFromBytes: cannot parse big int index[i][j], length of input byte is too small")
			}
			currentByte := b[offset : offset+byteLength]
			offset += byteLength
			row[j] = new(big.Int).SetBytes(currentByte)
		}
		indexes[i] = row
	}
	if sigPub == nil {
		sigPub = new(SigPubKey)
	}
	sigPub.Indexes = indexes
	return nil
}

type TxBase struct {
	// Basic data, required
	Version  int8   `json:"Version"`
	Type     string `json:"Type"` // Transaction type
	LockTime int64  `json:"LockTime"`
	Fee      uint64 `json:"Fee"` // Fee applies: always consant
	Info     []byte // 512 bytes
	// Sign and Privacy proof, required
	SigPubKey            []byte `json:"SigPubKey, omitempty"` // 33 bytes
	Sig                  []byte `json:"Sig, omitempty"`       //
	Proof                privacy.Proof
	PubKeyLastByteSender byte
	// private field, not use for json parser, only use as temp variable
	sigPrivKey       []byte       // is ALWAYS private property of struct, if privacy: 64 bytes, and otherwise, 32 bytes
	cachedHash       *Hash // cached hash data of tx
	cachedActualSize *uint64      // cached actualsize data for tx
}

type Tx struct {
	TxBase
}

type CoinCache struct{
	PublicKeys 		[][]byte 			`json: "public_keys"`
	Commitments 	[][]byte 			`json: "commitments"`
	AssetTags		[][]byte 			`json: "asset_tags, omitempty"`
	PkToIndexMap 	map[string]uint64 	`json: "pk_to_index"`
}

var b64 = base64.StdEncoding
var genericError = errors.New("Generic error for ASM")

// []byte equivalents are by default encoded with base64 when handled by JSON
type InitParamsAsm struct{
	SenderSK    privacy.PrivateKey		`json: "sender_sk"`
	PaymentInfo []privacy.PaymentInfo	`json: "payment_infos"`
	// TODO: implement serializer for coin ver2 in JS
	InputCoins  []CoinInter 	 		`json: "input_coins"`
	Fee         uint64 					`json: "fee"`
	HasPrivacy  bool 					`json: "has_privacy, omitempty"`
	TokenID     []byte 					`json: "token_id, omitempty"`
	MetaData    map[string]interface{}	`json: "metadata, omitempty"`
	Info        []byte 					`json: "info, omitempty"`
	Kvargs		map[string]interface{} 	`json: "kvargs, omitempty"`

	Cache 		CoinCache 				`json: "coin_cache"`
}

type TxPrivacyInitParams struct {
	SenderSK    *privacy.PrivateKey
	PaymentInfo []*privacy.PaymentInfo
	InputCoins  []privacy.PlainCoin
	Fee         uint64
	HasPrivacy  bool
	TokenID     *Hash
	Info        []byte
}

func (params *InitParamsAsm) GetInputCoins() []privacy.PlainCoin{
	var result []privacy.PlainCoin
	for _,ci := range params.InputCoins{
		c := ci.GetCoinV2()
		result = append(result,c)
	}
	return result
}

func (params *InitParamsAsm) GetGenericParams() *TxPrivacyInitParams{
	var pInfos []*privacy.PaymentInfo
	for _, payInf := range params.PaymentInfo{
		pInfos = append(pInfos, &payInf)
	}
	var tid Hash
	tid.SetBytes(params.TokenID)
	// TODO : handle metadata for ASM
	return &TxPrivacyInitParams{SenderSK: &params.SenderSK, PaymentInfo: pInfos, InputCoins: params.GetInputCoins(), Fee: params.Fee, HasPrivacy: params.HasPrivacy, TokenID: &tid, Info: params.Info}

}

type CoinInter struct {
	Version    		uint8
	Info       		[]byte
	PublicKey  		[]byte
	Commitment 		[]byte
	KeyImage   		[]byte

	SharedRandom 	[]byte
	TxRandom     	[]byte
	Mask    		[]byte
	Amount 			[]byte
	// tag is nil unless confidential asset
	AssetTag  		[]byte
}
func (c CoinInter) Bytes() []byte{
	coinBytes := []byte{c.Version}
	info := c.Info
	temp := len(info)
	if privacy.MaxSizeInfoCoin<temp{
		temp = privacy.MaxSizeInfoCoin
	}
	byteLengthInfo := byte(temp)
	coinBytes = append(coinBytes, byteLengthInfo)
	coinBytes = append(coinBytes, info[:byteLengthInfo]...)

	if c.PublicKey != nil {
		coinBytes = append(coinBytes, byte(privacy.Ed25519KeySize))
		coinBytes = append(coinBytes, c.PublicKey...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.Commitment != nil {
		coinBytes = append(coinBytes, byte(privacy.Ed25519KeySize))
		coinBytes = append(coinBytes, c.Commitment...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.KeyImage != nil {
		coinBytes = append(coinBytes, byte(privacy.Ed25519KeySize))
		coinBytes = append(coinBytes, c.KeyImage...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.SharedRandom != nil {
		coinBytes = append(coinBytes, byte(privacy.Ed25519KeySize))
		coinBytes = append(coinBytes, c.SharedRandom...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.TxRandom != nil {
		coinBytes = append(coinBytes, privacy.TxRandomGroupSize)
		coinBytes = append(coinBytes, c.TxRandom...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.Mask != nil {
		coinBytes = append(coinBytes, byte(privacy.Ed25519KeySize))
		coinBytes = append(coinBytes, c.Mask...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.Amount != nil {
		coinBytes = append(coinBytes, byte(privacy.Ed25519KeySize))
		coinBytes = append(coinBytes, c.Amount...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	if c.AssetTag != nil {
		coinBytes = append(coinBytes, byte(privacy.Ed25519KeySize))
		coinBytes = append(coinBytes, c.AssetTag...)
	} else {
		coinBytes = append(coinBytes, byte(0))
	}

	return coinBytes
}
func (ci CoinInter) GetCoinV2() *privacy.CoinV2{
	c := &privacy.CoinV2{}
	err := c.SetBytes(ci.Bytes())
	if err!=nil{
		return nil
	}
	return c
}

func (tx TxBase) String() string {
	record := strconv.Itoa(int(tx.Version))
	record += strconv.FormatInt(tx.LockTime, 10)
	record += strconv.FormatUint(tx.Fee, 10)
	if tx.Proof != nil {
		record += base64.StdEncoding.EncodeToString(tx.Proof.Bytes())
	}

	return record
}

func (tx TxBase) Hash() *Hash {
	if tx.cachedHash != nil {
		return tx.cachedHash
	}
	inBytes := []byte(tx.String())
	hash := HashH(inBytes)
	tx.cachedHash = &hash
	return &hash
}

func generateMlsagRingAsm(inputCoins []privacy.PlainCoin, outputCoins []*privacy.CoinV2, params *InitParamsAsm, pi int, shardID byte, ringSize int) (*mlsag.Ring, [][]*big.Int, error) {
	coinCache := params.Cache
	outputCoinsAsGeneric := make([]privacy.Coin, len(outputCoins))
	for i:=0;i<len(outputCoins);i++{
		outputCoinsAsGeneric[i] = outputCoins[i]
	}
	l := len(coinCache.PublicKeys)
	if len(coinCache.Commitments)!=l{
		err := errors.New("Length mismatch in coin cache")
		return nil, nil, err
	}
	randRange := big.NewInt(0).SetUint64(uint64(l))
	sumOutputsWithFee := calculateSumOutputsWithFee(outputCoinsAsGeneric, params.Fee)

	indexes := make([][]*big.Int, ringSize)
	ring := make([][]*privacy.Point, ringSize)
	for i := 0; i < ringSize; i += 1 {
		sumInputs := new(privacy.Point).Identity()
		sumInputs.Sub(sumInputs, sumOutputsWithFee)

		row := make([]*privacy.Point, len(inputCoins))
		rowIndexes := make([]*big.Int, len(inputCoins))
		if i == pi {
			for j := 0; j < len(inputCoins); j += 1 {
				row[j] = inputCoins[j].GetPublicKey()
				publicKeyBytes := inputCoins[j].GetPublicKey().ToBytesS()
				key := b64.EncodeToString(publicKeyBytes)
				val, exists := coinCache.PkToIndexMap[key]
				if !exists{
					err := errors.New(fmt.Sprintf("Cannot find a coin's index using cached index map - %s", key))
					return nil, nil, err
				}
				rowIndexes[j] = big.NewInt(0).SetUint64(val)
				sumInputs.Add(sumInputs, inputCoins[j].GetCommitment())
			}
		} else {
			for j := 0; j < len(inputCoins); j += 1 {
				temp, _ := RandBigIntMaxRange(randRange)
				pos := int(temp.Uint64())
				pkBytes := coinCache.PublicKeys[pos]
				commitmentBytes := coinCache.Commitments[pos]
				key := b64.EncodeToString(pkBytes)
				val, exists := coinCache.PkToIndexMap[key]
				if !exists{
					err := errors.New(fmt.Sprintf("Cannot find a coin's index using cached index map - %s", key))
					return nil, nil, err
				}
				rowIndexes[j] = big.NewInt(0).SetUint64(val)
				row[j], _ = new(privacy.Point).FromBytesS(pkBytes)

				commitment, _ := new(privacy.Point).FromBytesS(commitmentBytes)
				sumInputs.Add(sumInputs, commitment)
			}
		}
		row = append(row, sumInputs)
		ring[i] = row
		indexes[i] = rowIndexes
	}
	return mlsag.NewRing(ring), indexes, nil
}

func (tx *Tx) proveAsm(params *InitParamsAsm) error {
	var outputCoins []*privacy.CoinV2
	var pInfos []*privacy.PaymentInfo
	for _, payInf := range params.PaymentInfo{
		c, err := privacy.NewCoinFromPaymentInfo(&payInf)
		if err!=nil{
			return err
		}
		outputCoins = append(outputCoins, c)
		pInfos = append(pInfos, &payInf)
	}
	inputCoins := params.GetInputCoins()
	var err error
	tx.Proof, err = privacy.ProveV2(inputCoins, outputCoins, nil, false, pInfos)
	if err != nil {
		return err
	}

	err = tx.signOnMessageAsm(inputCoins, outputCoins, params, tx.Hash()[:])
	return err
}

func (tx *Tx) signOnMessageAsm(inp []privacy.PlainCoin, out []*privacy.CoinV2, params *InitParamsAsm, hashedMessage []byte) error {
	if tx.Sig != nil {
		return genericError
	}
	ringSize := privacy.RingSize
	if !params.HasPrivacy {
		ringSize = 1
	}

	// Generate Ring
	piBig,piErr := RandBigIntMaxRange(big.NewInt(int64(ringSize)))
	if piErr!=nil{
		return piErr
	}
	var pi int = int(piBig.Int64())
	shardID := GetShardIDFromLastByte(tx.PubKeyLastByteSender)
	ring, indexes, err := generateMlsagRingAsm(inp, out, params, pi, shardID, ringSize)
	if err != nil {
		return err
	}

	// Set SigPubKey
	txSigPubKey := new(SigPubKey)
	txSigPubKey.Indexes = indexes
	tx.SigPubKey, err = txSigPubKey.Bytes()
	if err != nil {
		return err
	}

	// Set sigPrivKey
	privKeysMlsag, err := createPrivKeyMlsag(inp, out, &params.SenderSK)
	if err != nil {
		return err
	}
	sag := mlsag.NewMlsag(privKeysMlsag, ring, pi)
	sk, err := privacy.ArrayScalarToBytes(&privKeysMlsag)
	if err != nil {
		return err
	}
	tx.sigPrivKey = sk

	// Set Signature
	mlsagSignature, err := sag.Sign(hashedMessage)
	if err != nil {
		return err
	}
	// inputCoins already hold keyImage so set to nil to reduce size
	mlsagSignature.SetKeyImages(nil)
	tx.Sig, err = mlsagSignature.ToBytes()

	return err
}

func (tx *Tx) InitASM(params *InitParamsAsm) error {
	innerParams := params.GetGenericParams()

	if err := ValidateTxParams(innerParams); err != nil {
		return err
	}

	// Init tx and params (tx and params will be changed)
	if err := tx.initializeTxAndParams(innerParams); err != nil {
		return err
	}
	// if check, err := tx.IsNonPrivacyNonInput(innerParams); check {
	// 	return err
	// }

	if err := tx.proveAsm(params); err != nil {
		return err
	}

	return nil
}

func (tx *TxBase) initializeTxAndParams(params *TxPrivacyInitParams) error {
	var err error
	// Get Keyset from param
	skBytes := *params.SenderSK
	senderPaymentAddress := privacy.GeneratePaymentAddress(skBytes)
	tx.sigPrivKey = skBytes
	// Tx: initialize some values
	tx.LockTime = 0
	tx.Fee = params.Fee
	// normal type indicator
	tx.Type = "n"
	tx.PubKeyLastByteSender = senderPaymentAddress.Pk[len(senderPaymentAddress.Pk)-1]

	// we don't support version 1
	tx.Version = 2
	tx.Info = params.Info
	// Params: update balance if overbalance
	if err = updateParamsWhenOverBalance(params, senderPaymentAddress); err != nil {
		return err
	}
	return nil
}

func ValidateTxParams(params *TxPrivacyInitParams) error {
	if len(params.InputCoins) > 255 {
		return genericError
	}
	if len(params.PaymentInfo) > 254 {
		return genericError
	}

	if params.TokenID == nil {
		// using default PRV
		params.TokenID = &Hash{}
		err := params.TokenID.SetBytes(PRVCoinID[:])
		if err != nil {
			return genericError
		}
	}
	return nil
}

func calculateSumOutputsWithFee(outputCoins []privacy.Coin, fee uint64) *privacy.Point {
	sumOutputsWithFee := new(privacy.Point).Identity()
	for i := 0; i < len(outputCoins); i += 1 {
		sumOutputsWithFee.Add(sumOutputsWithFee, outputCoins[i].GetCommitment())
	}
	feeCommitment := new(privacy.Point).ScalarMult(
		privacy.PedCom.G[privacy.PedersenValueIndex],
		new(privacy.Scalar).FromUint64(fee),
	)
	sumOutputsWithFee.Add(sumOutputsWithFee, feeCommitment)
	return sumOutputsWithFee
}

func createPrivKeyMlsag(inputCoins []privacy.PlainCoin, outputCoins []*privacy.CoinV2, senderSK *privacy.PrivateKey) ([]*privacy.Scalar, error) {
	sumRand := new(privacy.Scalar).FromUint64(0)
	for _, in := range inputCoins {
		sumRand.Add(sumRand, in.GetRandomness())
	}
	for _, out := range outputCoins {
		sumRand.Sub(sumRand, out.GetRandomness())
	}

	privKeyMlsag := make([]*privacy.Scalar, len(inputCoins)+1)
	for i := 0; i < len(inputCoins); i += 1 {
		var err error
		privKeyMlsag[i], err = inputCoins[i].ParsePrivateKeyOfCoin(*senderSK)
		if err != nil {
			return nil, err
		}
	}
	privKeyMlsag[len(inputCoins)] = sumRand
	return privKeyMlsag, nil
}


func updateParamsWhenOverBalance(params *TxPrivacyInitParams, senderPaymentAddree privacy.PaymentAddress) error {
	// Calculate sum of all output coins' value
	sumOutputValue := uint64(0)
	for _, p := range params.PaymentInfo {
		sumOutputValue += p.Amount
	}

	// Calculate sum of all input coins' value
	sumInputValue := uint64(0)
	for _, coin := range params.InputCoins {
		sumInputValue += coin.GetValue()
	}

	overBalance := int64(sumInputValue - sumOutputValue - params.Fee)
	// Check if sum of input coins' value is at least sum of output coins' value and tx fee
	if overBalance < 0 {
		return genericError
	}
	// Create a new payment to sender's pk where amount is overBalance if > 0
	if overBalance > 0 {
		// Should not check error because have checked before
		changePaymentInfo := new(privacy.PaymentInfo)
		changePaymentInfo.Amount = uint64(overBalance)
		changePaymentInfo.PaymentAddress = senderPaymentAddree
		params.PaymentInfo = append(params.PaymentInfo, changePaymentInfo)
	}

	return nil
}