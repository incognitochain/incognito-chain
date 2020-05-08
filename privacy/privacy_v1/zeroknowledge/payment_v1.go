package zkp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/privacy/coin"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
	"github.com/incognitochain/incognito-chain/privacy/proof/agg_interface"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/aggregatedrange"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/oneoutofmany"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/serialnumbernoprivacy"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/serialnumberprivacy"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/utils"
)

// PaymentProof contains all of PoK for spending coin
type PaymentProof struct {
	// for input coins
	oneOfManyProof    []*oneoutofmany.OneOutOfManyProof
	serialNumberProof []*serialnumberprivacy.SNPrivacyProof
	// it is exits when tx has no privacy
	serialNumberNoPrivacyProof []*serialnumbernoprivacy.SNNoPrivacyProof

	// for output coins
	// for proving each value and sum of them are less than a threshold value
	aggregatedRangeProof *aggregatedrange.AggregatedRangeProof

	inputCoins  []coin.PlainCoin
	outputCoins []*coin.CoinV1

	commitmentOutputValue   []*operation.Point
	commitmentOutputSND     []*operation.Point
	commitmentOutputShardID []*operation.Point

	commitmentInputSecretKey *operation.Point
	commitmentInputValue     []*operation.Point
	commitmentInputSND       []*operation.Point
	commitmentInputShardID   *operation.Point

	commitmentIndices []uint64
}

func (proof *PaymentProof) GetVersion() uint8 { return 1 }

// GET/SET function
func (proof PaymentProof) GetOneOfManyProof() []*oneoutofmany.OneOutOfManyProof {
	return proof.oneOfManyProof
}
func (proof PaymentProof) GetSerialNumberProof() []*serialnumberprivacy.SNPrivacyProof {
	return proof.serialNumberProof
}
func (proof PaymentProof) GetSerialNumberNoPrivacyProof() []*serialnumbernoprivacy.SNNoPrivacyProof {
	return proof.serialNumberNoPrivacyProof
}
func (proof PaymentProof) GetAggregatedRangeProof() agg_interface.AggregatedRangeProof {
	return proof.aggregatedRangeProof
}
func (proof PaymentProof) GetCommitmentOutputValue() []*operation.Point {
	return proof.commitmentOutputValue
}
func (proof PaymentProof) GetCommitmentOutputSND() []*operation.Point {
	return proof.commitmentOutputSND
}
func (proof PaymentProof) GetCommitmentOutputShardID() []*operation.Point {
	return proof.commitmentOutputShardID
}
func (proof PaymentProof) GetCommitmentInputSecretKey() *operation.Point {
	return proof.commitmentInputSecretKey
}
func (proof PaymentProof) GetCommitmentInputValue() []*operation.Point {
	return proof.commitmentInputValue
}
func (proof PaymentProof) GetCommitmentInputSND() []*operation.Point { return proof.commitmentInputSND }
func (proof PaymentProof) GetCommitmentInputShardID() *operation.Point {
	return proof.commitmentInputShardID
}
func (proof PaymentProof) GetCommitmentIndices() []uint64 { return proof.commitmentIndices }
func (proof PaymentProof) GetInputCoins() []coin.PlainCoin { return proof.inputCoins }
func (proof PaymentProof) GetOutputCoins() []coin.Coin {
	res := make([]coin.Coin, len(proof.outputCoins))
	for i := 0; i < len(proof.outputCoins); i += 1 {
		res[i] = proof.outputCoins[i]
	}
	return res
}

func (proof *PaymentProof) SetInputCoins(v []coin.PlainCoin) {
	proof.inputCoins = make([]coin.PlainCoin, len(v))
	for i := 0; i < len(v); i += 1 {
		b := v[i].Bytes()
		proof.inputCoins[i], _ = coin.NewPlainCoinFromByte(b)
	}
}

// SetOutputCoins's input should be all coinv1
func (proof *PaymentProof) SetOutputCoins(v []coin.Coin) {
	proof.outputCoins = make([]*coin.CoinV1, len(v))
	for i := 0; i < len(v); i += 1 {
		b := v[i].Bytes()
		proof.outputCoins[i] = new(coin.CoinV1)
		proof.outputCoins[i].SetBytes(b)
	}
}

// End GET/SET function

// Init
func (proof *PaymentProof) Init() {
	aggregatedRangeProof := &aggregatedrange.AggregatedRangeProof{}
	aggregatedRangeProof.Init()
	proof.oneOfManyProof = []*oneoutofmany.OneOutOfManyProof{}
	proof.serialNumberProof = []*serialnumberprivacy.SNPrivacyProof{}
	proof.aggregatedRangeProof = aggregatedRangeProof
	proof.inputCoins = []coin.PlainCoin{}
	proof.outputCoins = []*coin.CoinV1{}

	proof.commitmentOutputValue = []*operation.Point{}
	proof.commitmentOutputSND = []*operation.Point{}
	proof.commitmentOutputShardID = []*operation.Point{}

	proof.commitmentInputSecretKey = new(operation.Point)
	proof.commitmentInputValue = []*operation.Point{}
	proof.commitmentInputSND = []*operation.Point{}
	proof.commitmentInputShardID = new(operation.Point)
}

func (proof PaymentProof) MarshalJSON() ([]byte, error) {
	data := proof.Bytes()
	//temp := base58.Base58Check{}.Encode(data, common.ZeroByte)
	temp := base64.StdEncoding.EncodeToString(data)
	return json.Marshal(temp)
}

func (proof *PaymentProof) UnmarshalJSON(data []byte) error {
	dataStr := common.EmptyString
	errJson := json.Unmarshal(data, &dataStr)
	if errJson != nil {
		return errJson
	}

	//temp, _, err := base58.Base58Check{}.Decode(dataStr)
	temp, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return err
	}

	err = proof.SetBytes(temp)
	if err.(*errhandler.PrivacyError) != nil {
		return err
	}
	return nil
}

func (proof PaymentProof) Bytes() []byte {
	var bytes []byte
	hasPrivacy := len(proof.oneOfManyProof) > 0

	// OneOfManyProofSize
	bytes = append(bytes, byte(len(proof.oneOfManyProof)))
	for i := 0; i < len(proof.oneOfManyProof); i++ {
		oneOfManyProof := proof.oneOfManyProof[i].Bytes()
		bytes = append(bytes, common.IntToBytes(utils.OneOfManyProofSize)...)
		bytes = append(bytes, oneOfManyProof...)
	}

	// SerialNumberProofSize
	bytes = append(bytes, byte(len(proof.serialNumberProof)))
	for i := 0; i < len(proof.serialNumberProof); i++ {
		serialNumberProof := proof.serialNumberProof[i].Bytes()
		bytes = append(bytes, common.IntToBytes(utils.SnPrivacyProofSize)...)
		bytes = append(bytes, serialNumberProof...)
	}

	// SNNoPrivacyProofSize
	bytes = append(bytes, byte(len(proof.serialNumberNoPrivacyProof)))
	for i := 0; i < len(proof.serialNumberNoPrivacyProof); i++ {
		snNoPrivacyProof := proof.serialNumberNoPrivacyProof[i].Bytes()
		bytes = append(bytes, byte(utils.SnNoPrivacyProofSize))
		bytes = append(bytes, snNoPrivacyProof...)
	}

	//ComOutputMultiRangeProofSize
	if hasPrivacy {
		comOutputMultiRangeProof := proof.aggregatedRangeProof.Bytes()
		bytes = append(bytes, common.IntToBytes(len(comOutputMultiRangeProof))...)
		bytes = append(bytes, comOutputMultiRangeProof...)
	} else {
		bytes = append(bytes, []byte{0, 0}...)
	}

	// InputCoins
	bytes = append(bytes, byte(len(proof.inputCoins)))
	for i := 0; i < len(proof.inputCoins); i++ {
		inputCoins := proof.inputCoins[i].Bytes()
		bytes = append(bytes, byte(len(inputCoins)))
		bytes = append(bytes, inputCoins...)
	}

	// OutputCoins
	bytes = append(bytes, byte(len(proof.outputCoins)))
	for i := 0; i < len(proof.outputCoins); i++ {
		outputCoins := proof.outputCoins[i].Bytes()
		lenOutputCoins := len(outputCoins)
		lenOutputCoinsBytes := []byte{}
		if lenOutputCoins < 256 {
			lenOutputCoinsBytes = []byte{byte(lenOutputCoins)}
		} else {
			lenOutputCoinsBytes = common.IntToBytes(lenOutputCoins)
		}

		bytes = append(bytes, lenOutputCoinsBytes...)
		bytes = append(bytes, outputCoins...)
	}

	// ComOutputValue
	bytes = append(bytes, byte(len(proof.commitmentOutputValue)))
	for i := 0; i < len(proof.commitmentOutputValue); i++ {
		comOutputValue := proof.commitmentOutputValue[i].ToBytesS()
		bytes = append(bytes, byte(operation.Ed25519KeySize))
		bytes = append(bytes, comOutputValue...)
	}

	// ComOutputSND
	bytes = append(bytes, byte(len(proof.commitmentOutputSND)))
	for i := 0; i < len(proof.commitmentOutputSND); i++ {
		comOutputSND := proof.commitmentOutputSND[i].ToBytesS()
		bytes = append(bytes, byte(operation.Ed25519KeySize))
		bytes = append(bytes, comOutputSND...)
	}

	// ComOutputShardID
	bytes = append(bytes, byte(len(proof.commitmentOutputShardID)))
	for i := 0; i < len(proof.commitmentOutputShardID); i++ {
		comOutputShardID := proof.commitmentOutputShardID[i].ToBytesS()
		bytes = append(bytes, byte(operation.Ed25519KeySize))
		bytes = append(bytes, comOutputShardID...)
	}

	//ComInputSK 				*operation.Point
	if proof.commitmentInputSecretKey != nil {
		comInputSK := proof.commitmentInputSecretKey.ToBytesS()
		bytes = append(bytes, byte(operation.Ed25519KeySize))
		bytes = append(bytes, comInputSK...)
	} else {
		bytes = append(bytes, byte(0))
	}

	//ComInputValue 		[]*operation.Point
	bytes = append(bytes, byte(len(proof.commitmentInputValue)))
	for i := 0; i < len(proof.commitmentInputValue); i++ {
		comInputValue := proof.commitmentInputValue[i].ToBytesS()
		bytes = append(bytes, byte(operation.Ed25519KeySize))
		bytes = append(bytes, comInputValue...)
	}

	//ComInputSND 			[]*privacy.Point
	bytes = append(bytes, byte(len(proof.commitmentInputSND)))
	for i := 0; i < len(proof.commitmentInputSND); i++ {
		comInputSND := proof.commitmentInputSND[i].ToBytesS()
		bytes = append(bytes, byte(operation.Ed25519KeySize))
		bytes = append(bytes, comInputSND...)
	}

	//ComInputShardID 	*privacy.Point
	if proof.commitmentInputShardID != nil {
		comInputShardID := proof.commitmentInputShardID.ToBytesS()
		bytes = append(bytes, byte(operation.Ed25519KeySize))
		bytes = append(bytes, comInputShardID...)
	} else {
		bytes = append(bytes, byte(0))
	}

	// convert commitment index to bytes array
	for i := 0; i < len(proof.commitmentIndices); i++ {
		bytes = append(bytes, common.AddPaddingBigInt(big.NewInt(int64(proof.commitmentIndices[i])), common.Uint64Size)...)
	}
	//fmt.Printf("BYTES ------------------ %v\n", bytes)
	//fmt.Printf("LEN BYTES ------------------ %v\n", len(bytes))

	return bytes
}

func (proof *PaymentProof) SetBytes(proofbytes []byte) *errhandler.PrivacyError {
	if len(proofbytes) == 0 {
		return errhandler.NewPrivacyErr(errhandler.InvalidInputToSetBytesErr, nil)
	}
	var err error
	offset := 0

	// Set OneOfManyProofSize
	if offset >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range one out of many proof"))
	}
	lenOneOfManyProofArray := int(proofbytes[offset])
	offset += 1
	proof.oneOfManyProof = make([]*oneoutofmany.OneOutOfManyProof, lenOneOfManyProofArray)
	for i := 0; i < lenOneOfManyProofArray; i++ {
		if offset+2 > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range one out of many proof"))
		}
		lenOneOfManyProof := common.BytesToInt(proofbytes[offset : offset+2])
		offset += 2
		proof.oneOfManyProof[i] = new(oneoutofmany.OneOutOfManyProof).Init()

		if offset+lenOneOfManyProof > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range one out of many proof"))
		}
		err := proof.oneOfManyProof[i].SetBytes(proofbytes[offset : offset+lenOneOfManyProof])
		if err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		}
		offset += lenOneOfManyProof
	}

	// Set serialNumberProofSize
	if offset >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range serial number proof"))
	}
	lenSerialNumberProofArray := int(proofbytes[offset])
	offset += 1
	proof.serialNumberProof = make([]*serialnumberprivacy.SNPrivacyProof, lenSerialNumberProofArray)
	for i := 0; i < lenSerialNumberProofArray; i++ {
		if offset+2 > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range serial number proof"))
		}
		lenSerialNumberProof := common.BytesToInt(proofbytes[offset : offset+2])
		offset += 2
		proof.serialNumberProof[i] = new(serialnumberprivacy.SNPrivacyProof).Init()

		if offset+lenSerialNumberProof > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range serial number proof"))
		}
		err := proof.serialNumberProof[i].SetBytes(proofbytes[offset : offset+lenSerialNumberProof])
		if err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		}
		offset += lenSerialNumberProof
	}

	// Set SNNoPrivacyProofSize
	if offset >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range serial number no privacy proof"))
	}
	lenSNNoPrivacyProofArray := int(proofbytes[offset])
	offset += 1
	proof.serialNumberNoPrivacyProof = make([]*serialnumbernoprivacy.SNNoPrivacyProof, lenSNNoPrivacyProofArray)
	for i := 0; i < lenSNNoPrivacyProofArray; i++ {
		if offset >= len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range serial number no privacy proof"))
		}
		lenSNNoPrivacyProof := int(proofbytes[offset])
		offset += 1

		proof.serialNumberNoPrivacyProof[i] = new(serialnumbernoprivacy.SNNoPrivacyProof).Init()
		if offset+lenSNNoPrivacyProof > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range serial number no privacy proof"))
		}
		err := proof.serialNumberNoPrivacyProof[i].SetBytes(proofbytes[offset : offset+lenSNNoPrivacyProof])
		if err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		}
		offset += lenSNNoPrivacyProof
	}

	//ComOutputMultiRangeProofSize *aggregatedRangeProof
	if offset+2 > len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range aggregated range proof"))
	}
	lenComOutputMultiRangeProof := common.BytesToInt(proofbytes[offset : offset+2])
	offset += 2
	if lenComOutputMultiRangeProof > 0 {
		aggregatedRangeProof := &aggregatedrange.AggregatedRangeProof{}
		aggregatedRangeProof.Init()
		proof.aggregatedRangeProof = aggregatedRangeProof
		if offset+lenComOutputMultiRangeProof > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range aggregated range proof"))
		}
		err := proof.aggregatedRangeProof.SetBytes(proofbytes[offset : offset+lenComOutputMultiRangeProof])
		if err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		}
		offset += lenComOutputMultiRangeProof
	}

	//InputCoins  []*coin.PlainCoinV1
	if offset >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
	}
	lenInputCoinsArray := int(proofbytes[offset])
	offset += 1
	proof.inputCoins = make([]coin.PlainCoin, lenInputCoinsArray)
	for i := 0; i < lenInputCoinsArray; i++ {
		if offset >= len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
		}
		lenInputCoin := int(proofbytes[offset])
		offset += 1

		if offset+lenInputCoin > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
		}
		coinBytes := proofbytes[offset : offset+lenInputCoin]
		proof.inputCoins[i], err = coin.NewPlainCoinFromByte(coinBytes)
		if err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Set byte to inputCoin got error"))
		}
		offset += lenInputCoin
	}

	//OutputCoins []*privacy.OutputCoin
	if offset >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
	}
	lenOutputCoinsArray := int(proofbytes[offset])
	offset += 1
	proof.outputCoins = make([]*coin.CoinV1, lenOutputCoinsArray)
	for i := 0; i < lenOutputCoinsArray; i++ {
		proof.outputCoins[i] = new(coin.CoinV1)
		// try get 1-byte for len
		if offset >= len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
		}
		lenOutputCoin := int(proofbytes[offset])
		offset += 1

		if offset+lenOutputCoin > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
		}
		err := proof.outputCoins[i].SetBytes(proofbytes[offset : offset+lenOutputCoin])
		if err != nil {
			// 1-byte is wrong
			// try get 2-byte for len
			if offset+1 > len(proofbytes) {
				return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
			}
			lenOutputCoin = common.BytesToInt(proofbytes[offset-1 : offset+1])
			offset += 1

			if offset+lenOutputCoin > len(proofbytes) {
				return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
			}
			err1 := proof.outputCoins[i].SetBytes(proofbytes[offset : offset+lenOutputCoin])
			if err1 != nil {
				return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
			}
		}
		offset += lenOutputCoin
	}
	//ComOutputValue   []*privacy.Point
	if offset >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment output coins value"))
	}
	lenComOutputValueArray := int(proofbytes[offset])
	offset += 1
	proof.commitmentOutputValue = make([]*operation.Point, lenComOutputValueArray)
	for i := 0; i < lenComOutputValueArray; i++ {
		if offset >= len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment output coins value"))
		}
		lenComOutputValue := int(proofbytes[offset])
		offset += 1

		if offset+lenComOutputValue > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment output coins value"))
		}
		proof.commitmentOutputValue[i], err = new(operation.Point).FromBytesS(proofbytes[offset : offset+lenComOutputValue])
		if err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		}
		offset += lenComOutputValue
	}
	//ComOutputSND     []*operation.Point
	if offset >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment output coins snd"))
	}
	lenComOutputSNDArray := int(proofbytes[offset])
	offset += 1
	proof.commitmentOutputSND = make([]*operation.Point, lenComOutputSNDArray)
	for i := 0; i < lenComOutputSNDArray; i++ {
		if offset >= len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment output coins snd"))
		}
		lenComOutputSND := int(proofbytes[offset])
		offset += 1

		if offset+lenComOutputSND > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment output coins snd"))
		}
		proof.commitmentOutputSND[i], err = new(operation.Point).FromBytesS(proofbytes[offset : offset+lenComOutputSND])

		if err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		}
		offset += lenComOutputSND
	}

	// commitmentOutputShardID
	if offset >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment output coins shardid"))
	}
	lenComOutputShardIdArray := int(proofbytes[offset])
	offset += 1
	proof.commitmentOutputShardID = make([]*operation.Point, lenComOutputShardIdArray)
	for i := 0; i < lenComOutputShardIdArray; i++ {
		if offset >= len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment output coins shardid"))
		}
		lenComOutputShardId := int(proofbytes[offset])
		offset += 1

		if offset+lenComOutputShardId > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment output coins shardid"))
		}
		proof.commitmentOutputShardID[i], err = new(operation.Point).FromBytesS(proofbytes[offset : offset+lenComOutputShardId])

		if err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		}
		offset += lenComOutputShardId
	}

	//ComInputSK 				*operation.Point
	if offset >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment input coins private key"))
	}
	lenComInputSK := int(proofbytes[offset])
	offset += 1
	if lenComInputSK > 0 {
		if offset+lenComInputSK > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment input coins private key"))
		}
		proof.commitmentInputSecretKey, err = new(operation.Point).FromBytesS(proofbytes[offset : offset+lenComInputSK])

		if err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		}
		offset += lenComInputSK
	}
	//ComInputValue 		[]*operation.Point
	if offset >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment input coins value"))
	}
	lenComInputValueArr := int(proofbytes[offset])
	offset += 1
	proof.commitmentInputValue = make([]*operation.Point, lenComInputValueArr)
	for i := 0; i < lenComInputValueArr; i++ {
		if offset >= len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment input coins value"))
		}
		lenComInputValue := int(proofbytes[offset])
		offset += 1

		if offset+lenComInputValue > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment input coins value"))
		}
		proof.commitmentInputValue[i], err = new(operation.Point).FromBytesS(proofbytes[offset : offset+lenComInputValue])

		if err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		}
		offset += lenComInputValue
	}
	//ComInputSND 			[]*operation.Point
	if offset >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment input coins snd"))
	}
	lenComInputSNDArr := int(proofbytes[offset])
	offset += 1
	proof.commitmentInputSND = make([]*operation.Point, lenComInputSNDArr)
	for i := 0; i < lenComInputSNDArr; i++ {
		if offset >= len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment input coins snd"))
		}
		lenComInputSND := int(proofbytes[offset])
		offset += 1

		if offset+lenComInputSND > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment input coins snd"))
		}
		proof.commitmentInputSND[i], err = new(operation.Point).FromBytesS(proofbytes[offset : offset+lenComInputSND])

		if err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		}
		offset += lenComInputSND
	}
	//ComInputShardID 	*operation.Point
	if offset >= len(proofbytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment input coins shardid"))
	}
	lenComInputShardID := int(proofbytes[offset])
	offset += 1
	if lenComInputShardID > 0 {
		if offset+lenComInputShardID > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment input coins shardid"))
		}
		proof.commitmentInputShardID, err = new(operation.Point).FromBytesS(proofbytes[offset : offset+lenComInputShardID])

		if err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		}
		offset += lenComInputShardID
	}

	// get commitments list
	proof.commitmentIndices = make([]uint64, len(proof.oneOfManyProof)*privacy_util.CommitmentRingSize)
	for i := 0; i < len(proof.oneOfManyProof)*privacy_util.CommitmentRingSize; i++ {
		if offset+common.Uint64Size > len(proofbytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range commitment indices"))
		}
		proof.commitmentIndices[i] = new(big.Int).SetBytes(proofbytes[offset : offset+common.Uint64Size]).Uint64()
		offset = offset + common.Uint64Size
	}

	//fmt.Printf("SETBYTES ------------------ %v\n", proof.Bytes())

	return nil
}

func (proof PaymentProof) verifyNoPrivacy(pubKey key.PublicKey, fee uint64, shardID byte, tokenID *common.Hash) (bool, error) {
	var sumInputValue, sumOutputValue uint64
	sumInputValue = 0
	sumOutputValue = 0

	pubKeyLastByteSender := pubKey[len(pubKey)-1]
	senderShardID := common.GetShardIDFromLastByte(pubKeyLastByteSender)
	cmShardIDSender := new(operation.Point)
	cmShardIDSender.ScalarMult(operation.PedCom.G[operation.PedersenShardIDIndex], new(operation.Scalar).FromBytes([operation.Ed25519KeySize]byte{senderShardID}))

	for i := 0; i < len(proof.inputCoins); i++ {
		// Check input coins' Serial number is created from input coins' input and sender's spending key
		valid, err := proof.serialNumberNoPrivacyProof[i].Verify(nil)
		if !valid {
			Logger.Log.Errorf("Verify serial number no privacy proof failed")
			return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberNoPrivacyProofFailedErr, err)
		}

		// Check input coins' cm is calculated correctly
		cmSK := proof.inputCoins[i].GetPublicKey()
		cmValue := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenValueIndex], new(operation.Scalar).FromUint64(proof.inputCoins[i].GetValue()))
		cmSND := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenSndIndex], proof.inputCoins[i].GetSNDerivator())
		cmRandomness := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenRandomnessIndex], proof.inputCoins[i].GetRandomness())
		cmTmp := new(operation.Point).Add(cmSK, cmValue)
		cmTmp.Add(cmTmp, cmSND)
		cmTmp.Add(cmTmp, cmShardIDSender)
		cmTmp.Add(cmTmp, cmRandomness)

		if !operation.IsPointEqual(cmTmp, proof.inputCoins[i].GetCommitment()) {
			Logger.Log.Errorf("Input coins %v commitment wrong!\n", i)
			return false, errhandler.NewPrivacyErr(errhandler.VerifyCoinCommitmentInputFailedErr, nil)
		}

		// Calculate sum of input values
		sumInputValue += proof.inputCoins[i].GetValue()
	}

	for i := 0; i < len(proof.outputCoins); i++ {
		// Check output coins' cm is calculated correctly
		shardID, err := proof.outputCoins[i].GetShardID()
		if err != nil {
			Logger.Log.Errorf("Cannot parse shardID of outputcoin error: %v", err)
			return false, err
		}
		cmSK := proof.outputCoins[i].CoinDetails.GetPublicKey()
		cmValue := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenValueIndex], new(operation.Scalar).FromUint64(proof.outputCoins[i].CoinDetails.GetValue()))
		cmSND := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenSndIndex], proof.outputCoins[i].CoinDetails.GetSNDerivator())
		cmShardID := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenShardIDIndex], new(operation.Scalar).FromBytes([operation.Ed25519KeySize]byte{shardID}))
		cmRandomness := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenRandomnessIndex], proof.outputCoins[i].CoinDetails.GetRandomness())

		cmTmp := new(operation.Point).Add(cmSK, cmValue)
		cmTmp.Add(cmTmp, cmSND)
		cmTmp.Add(cmTmp, cmShardID)
		cmTmp.Add(cmTmp, cmRandomness)

		if !operation.IsPointEqual(cmTmp, proof.outputCoins[i].GetCommitment()) {
			Logger.Log.Errorf("Output coins %v commitment wrong!\n", i)
			return false, errhandler.NewPrivacyErr(errhandler.VerifyCoinCommitmentOutputFailedErr, nil)
		}
	}

	//Calculate sum of output values and check overflow output's value
	if len(proof.outputCoins) > 0 {
		sumOutputValue = proof.outputCoins[0].CoinDetails.GetValue()

		for i := 1; i < len(proof.outputCoins); i++ {
			outValue := proof.outputCoins[i].CoinDetails.GetValue()
			sumTmp := sumOutputValue + outValue
			if sumTmp < sumOutputValue || sumTmp < outValue {
				return false, errhandler.NewPrivacyErr(errhandler.UnexpectedErr, fmt.Errorf("Overflow output value %v\n", outValue))
			}

			sumOutputValue += outValue
		}
	}

	// check overflow fee value
	tmp := sumOutputValue + fee
	if tmp < sumOutputValue || tmp < fee {
		return false, errhandler.NewPrivacyErr(errhandler.UnexpectedErr, fmt.Errorf("Overflow fee value %v\n", fee))
	}

	// check if sum of input values equal sum of output values
	if sumInputValue != sumOutputValue+fee {
		Logger.Log.Debugf("sumInputValue: %v\n", sumInputValue)
		Logger.Log.Debugf("sumOutputValue: %v\n", sumOutputValue)
		Logger.Log.Debugf("fee: %v\n", fee)
		Logger.Log.Errorf("Sum of inputs is not equal sum of output!\n")
		return false, errhandler.NewPrivacyErr(errhandler.VerifyAmountNoPrivacyFailedErr, nil)
	}
	return true, nil
}

func (proof PaymentProof) verifyHasPrivacy(pubKey key.PublicKey, fee uint64, shardID byte, tokenID *common.Hash, isBatch bool, additionalData interface{}) (bool, error) {
	// verify for input coins
	commitmentsPtr := additionalData.(*[][privacy_util.CommitmentRingSize]*operation.Point)
	commitments := *commitmentsPtr

	for i := 0; i < len(proof.oneOfManyProof); i++ {
		Logger.Log.Debugf("[TEST] input coins %v\n ShardID %v fee %v", i, shardID, fee)
		Logger.Log.Debugf("[TEST] commitments indices %v\n", proof.commitmentIndices[i*privacy_util.CommitmentRingSize:i*privacy_util.CommitmentRingSize+8])
		// Verify for the proof one-out-of-N commitments is a commitment to the coins being spent
		// Calculate cm input sum

		proof.oneOfManyProof[i].Statement.Commitments = commitments[i][:]

		valid, err := proof.oneOfManyProof[i].Verify()
		if !valid {
			Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: One out of many failed")
			return false, errhandler.NewPrivacyErr(errhandler.VerifyOneOutOfManyProofFailedErr, err)
		}
		// Verify for the Proof that input coins' serial number is derived from the committed derivator
		valid, err = proof.serialNumberProof[i].Verify(nil)
		if !valid {
			Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Serial number privacy failed")
			return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberPrivacyProofFailedErr, err)
		}
	}

	// Check output coins' cm is calculated correctly
	for i := 0; i < len(proof.outputCoins); i++ {
		cmTmp := new(operation.Point).Add(proof.outputCoins[i].CoinDetails.GetPublicKey(), proof.commitmentOutputValue[i])
		cmTmp.Add(cmTmp, proof.commitmentOutputSND[i])
		cmTmp.Add(cmTmp, proof.commitmentOutputShardID[i])

		if !operation.IsPointEqual(cmTmp, proof.outputCoins[i].GetCommitment()) {
			Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Commitment for output coins are not computed correctly")
			return false, errhandler.NewPrivacyErr(errhandler.VerifyCoinCommitmentOutputFailedErr, nil)
		}
	}

	// Verify the proof that output values and sum of them do not exceed v_max
	if isBatch == false {
		valid, err := proof.aggregatedRangeProof.Verify()
		if !valid {
			Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Multi-range failed")
			return false, errhandler.NewPrivacyErr(errhandler.VerifyAggregatedProofFailedErr, err)
		}
	}

	// Verify the proof that sum of all input values is equal to sum of all output values
	comInputValueSum := new(operation.Point).Identity()
	for i := 0; i < len(proof.commitmentInputValue); i++ {
		comInputValueSum.Add(comInputValueSum, proof.commitmentInputValue[i])
	}

	comOutputValueSum := new(operation.Point).Identity()
	for i := 0; i < len(proof.commitmentOutputValue); i++ {
		comOutputValueSum.Add(comOutputValueSum, proof.commitmentOutputValue[i])
	}

	if fee > 0 {
		comOutputValueSum.Add(comOutputValueSum, new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenValueIndex], new(operation.Scalar).FromUint64(uint64(fee))))
	}

	// Logger.Log.Infof("comInputValueSum: %v\n", comInputValueSum.ToBytesS())
	// Logger.Log.Infof("comOutputValueSum: %v\n", comOutputValueSum.ToBytesS())

	if !operation.IsPointEqual(comInputValueSum, comOutputValueSum) {
		Logger.Log.Debugf("comInputValueSum: ", comInputValueSum)
		Logger.Log.Debugf("comOutputValueSum: ", comOutputValueSum)
		Logger.Log.Error("VERIFICATION PAYMENT PROOF: Sum of input coins' value is not equal to sum of output coins' value")
		return false, errhandler.NewPrivacyErr(errhandler.VerifyAmountPrivacyFailedErr, nil)
	}

	return true, nil
}

func (proof PaymentProof) Verify(hasPrivacy bool, pubKey key.PublicKey, fee uint64, shardID byte, tokenID *common.Hash, isBatch bool, additionalData interface{}) (bool, error) {
	// has no privacy
	if !hasPrivacy {
		return proof.verifyNoPrivacy(pubKey, fee, shardID, tokenID)
	}

	return proof.verifyHasPrivacy(pubKey, fee, shardID, tokenID, isBatch, additionalData)
}

func (proof *PaymentProof) IsPrivacy() bool {
	if proof == nil || len(proof.GetOneOfManyProof()) == 0 {
		return false
	}
	return true
}

func (proof PaymentProof) ValidateSanity() (bool, error) {
	if len(proof.inputCoins) > 255 {
		return false, errors.New("Input coins in tx are very large:" + strconv.Itoa(len(proof.inputCoins)))
	}

	if len(proof.outputCoins) > 255 {
		return false, errors.New("Output coins in tx are very large:" + strconv.Itoa(len(proof.outputCoins)))
	}

	isPrivacy := proof.IsPrivacy()

	if isPrivacy {

		if !proof.aggregatedRangeProof.ValidateSanity() {
			return false, errors.New("validate sanity Aggregated range proof failed")
		}

		for i := 0; i < len(proof.GetOneOfManyProof()); i++ {
			if !proof.GetOneOfManyProof()[i].ValidateSanity() {
				return false, errors.New("validate sanity One out of many proof failed")
			}
		}
		for i := 0; i < len(proof.GetSerialNumberProof()); i++ {
			if !proof.GetSerialNumberProof()[i].ValidateSanity() {
				return false, errors.New("validate sanity Serial number proof failed")
			}
		}

		// check input coins with privacy
		for i := 0; i < len(proof.inputCoins); i++ {
			if !proof.inputCoins[i].GetKeyImage().PointValid() {
				return false, errors.New("validate sanity Serial number of input coin failed")
			}
		}
		// check output coins with privacy
		for i := 0; i < len(proof.outputCoins); i++ {
			if !proof.outputCoins[i].CoinDetails.GetPublicKey().PointValid() {
				return false, errors.New("validate sanity Public key of output coin failed")
			}
			if !proof.outputCoins[i].CoinDetails.GetCommitment().PointValid() {
				return false, errors.New("validate sanity Coin commitment of output coin failed")
			}
			if !proof.outputCoins[i].CoinDetails.GetSNDerivator().ScalarValid() {
				return false, errors.New("validate sanity SNDerivator of output coin failed")
			}
		}
		// check ComInputSK
		if !proof.GetCommitmentInputSecretKey().PointValid() {
			return false, errors.New("validate sanity ComInputSK of proof failed")
		}
		// check ComInputValue
		for i := 0; i < len(proof.GetCommitmentInputValue()); i++ {
			if !proof.GetCommitmentInputValue()[i].PointValid() {
				return false, errors.New("validate sanity ComInputValue of proof failed")
			}
		}
		//check ComInputSND
		for i := 0; i < len(proof.GetCommitmentInputSND()); i++ {
			if !proof.GetCommitmentInputSND()[i].PointValid() {
				return false, errors.New("validate sanity ComInputSND of proof failed")
			}
		}
		//check ComInputShardID
		if !proof.GetCommitmentInputShardID().PointValid() {
			return false, errors.New("validate sanity ComInputShardID of proof failed")
		}

		// check ComOutputShardID
		for i := 0; i < len(proof.GetCommitmentOutputShardID()); i++ {
			if !proof.GetCommitmentOutputShardID()[i].PointValid() {
				return false, errors.New("validate sanity ComOutputShardID of proof failed")
			}
		}
		//check ComOutputSND
		for i := 0; i < len(proof.GetCommitmentOutputShardID()); i++ {
			if !proof.GetCommitmentOutputShardID()[i].PointValid() {
				return false, errors.New("validate sanity ComOutputSND of proof failed")
			}
		}
		//check ComOutputValue
		for i := 0; i < len(proof.GetCommitmentOutputValue()); i++ {
			if !proof.GetCommitmentOutputValue()[i].PointValid() {
				return false, errors.New("validate sanity ComOutputValue of proof failed")
			}
		}
		if len(proof.GetCommitmentIndices()) != len(proof.inputCoins)*privacy_util.CommitmentRingSize {
			return false, errors.New("validate sanity CommitmentIndices of proof failed")

		}
	}

	if !isPrivacy {
		for i := 0; i < len(proof.GetSerialNumberNoPrivacyProof()); i++ {
			if !proof.GetSerialNumberNoPrivacyProof()[i].ValidateSanity() {
				return false, errors.New("validate sanity Serial number no privacy proof failed")
			}
		}
		// check input coins without privacy
		for i := 0; i < len(proof.inputCoins); i++ {
			if !proof.inputCoins[i].GetCommitment().PointValid() {
				return false, errors.New("validate sanity CoinCommitment of input coin failed")
			}
			if !proof.inputCoins[i].GetPublicKey().PointValid() {
				return false, errors.New("validate sanity PublicKey of input coin failed")
			}
			if !proof.inputCoins[i].GetKeyImage().PointValid() {
				return false, errors.New("validate sanity Serial number of input coin failed")
			}
			if !proof.inputCoins[i].GetRandomness().ScalarValid() {
				return false, errors.New("validate sanity Randomness of input coin failed")
			}
			if !proof.inputCoins[i].GetSNDerivator().ScalarValid() {
				return false, errors.New("validate sanity SNDerivator of input coin failed")
			}
		}

		// check output coins without privacy
		for i := 0; i < len(proof.outputCoins); i++ {
			if !proof.outputCoins[i].CoinDetails.GetCommitment().PointValid() {
				return false, errors.New("validate sanity CoinCommitment of output coin failed")
			}
			if !proof.outputCoins[i].CoinDetails.GetPublicKey().PointValid() {
				return false, errors.New("validate sanity PublicKey of output coin failed")
			}
			if !proof.outputCoins[i].CoinDetails.GetRandomness().ScalarValid() {
				return false, errors.New("validate sanity Randomness of output coin failed")
			}
			if !proof.outputCoins[i].CoinDetails.GetSNDerivator().ScalarValid() {
				return false, errors.New("validate sanity SNDerivator of output coin failed")
			}
		}
	}
	return true, nil
}
