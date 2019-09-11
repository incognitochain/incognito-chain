package zkp

import (
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/aggregaterange"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/oneoutofmany"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/serialnumbernoprivacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/serialnumberprivacy"
)

// PaymentWitness contains all of witness for proving when spending coins
type PaymentWitness struct {
	privateKey          *big.Int
	inputCoins          []*privacy.InputCoin
	outputCoins         []*privacy.OutputCoin
	commitmentIndices   []uint64
	myCommitmentIndices []uint64

	oneOfManyWitness             []*oneoutofmany.OneOutOfManyWitness
	serialNumberWitness          []*serialnumberprivacy.SNPrivacyWitness
	serialNumberNoPrivacyWitness []*serialnumbernoprivacy.SNNoPrivacyWitness

	aggregatedRangeWitness *aggregaterange.AggregatedRangeWitness

	comOutputValue                 []*privacy.EllipticPoint
	comOutputSerialNumberDerivator []*privacy.EllipticPoint
	comOutputShardID               []*privacy.EllipticPoint

	comInputSecretKey             *privacy.EllipticPoint
	comInputValue                 []*privacy.EllipticPoint
	comInputSerialNumberDerivator []*privacy.EllipticPoint
	comInputShardID               *privacy.EllipticPoint

	randSecretKey *big.Int
}

func (paymentWitness PaymentWitness) GetRandSecretKey() *big.Int {
	return paymentWitness.randSecretKey
}

type PaymentWitnessParam struct {
	HasPrivacy              bool
	PrivateKey              *big.Int
	InputCoins              []*privacy.InputCoin
	OutputCoins             []*privacy.OutputCoin
	PublicKeyLastByteSender byte
	Commitments             []*privacy.EllipticPoint
	CommitmentIndices       []uint64
	MyCommitmentIndices     []uint64
	Fee                     uint64
}

// Build prepares witnesses for all protocol need to be proved when create tx
// if hashPrivacy = false, witness includes spending key, input coins, output coins
// otherwise, witness includes all attributes in PaymentWitness struct
func (wit *PaymentWitness) Init(PaymentWitnessParam PaymentWitnessParam) *privacy.PrivacyError {

	hasPrivacy := PaymentWitnessParam.HasPrivacy
	privateKey := PaymentWitnessParam.PrivateKey
	inputCoins := PaymentWitnessParam.InputCoins
	outputCoins := PaymentWitnessParam.OutputCoins
	publicKeyLastByteSender := PaymentWitnessParam.PublicKeyLastByteSender
	commitments := PaymentWitnessParam.Commitments
	commitmentIndices := PaymentWitnessParam.CommitmentIndices
	myCommitmentIndices := PaymentWitnessParam.MyCommitmentIndices
	_ = PaymentWitnessParam.Fee

	var r = rand.Reader

	if !hasPrivacy {
		for _, outCoin := range outputCoins {
			outCoin.CoinDetails.SetRandomness(privacy.RandScalar(r))
			err := outCoin.CoinDetails.CommitAll()
			if err != nil {
				return privacy.NewPrivacyErr(privacy.CommitNewOutputCoinNoPrivacyErr, nil)
			}
		}

		wit.privateKey = privateKey
		wit.inputCoins = inputCoins
		wit.outputCoins = outputCoins

		publicKey := inputCoins[0].CoinDetails.GetPublicKey()

		wit.serialNumberNoPrivacyWitness = make([]*serialnumbernoprivacy.SNNoPrivacyWitness, len(inputCoins))
		for i := 0; i < len(inputCoins); i++ {
			/***** Build witness for proving that serial number is derived from the committed derivator *****/
			if wit.serialNumberNoPrivacyWitness[i] == nil {
				wit.serialNumberNoPrivacyWitness[i] = new(serialnumbernoprivacy.SNNoPrivacyWitness)
			}
			wit.serialNumberNoPrivacyWitness[i].Set(inputCoins[i].CoinDetails.GetSerialNumber(), publicKey, inputCoins[i].CoinDetails.GetSNDerivator(), wit.privateKey)
		}
		return nil
	}

	wit.privateKey = privateKey
	wit.inputCoins = inputCoins
	wit.outputCoins = outputCoins
	wit.commitmentIndices = commitmentIndices
	wit.myCommitmentIndices = myCommitmentIndices

	numInputCoin := len(wit.inputCoins)

	randInputSK := privacy.RandScalar(r)
	// set rand sk for Schnorr signature
	wit.randSecretKey = new(big.Int).Set(randInputSK)

	cmInputSK := privacy.PedCom.CommitAtIndex(wit.privateKey, randInputSK, privacy.PedersenPrivateKeyIndex)
	wit.comInputSecretKey = new(privacy.EllipticPoint)
	wit.comInputSecretKey.Set(cmInputSK.GetX(), cmInputSK.GetY())

	randInputShardID := privacy.RandScalar(r)
	senderShardID := common.GetShardIDFromLastByte(publicKeyLastByteSender)
	wit.comInputShardID = privacy.PedCom.CommitAtIndex(big.NewInt(int64(senderShardID)), randInputShardID, privacy.PedersenShardIDIndex)

	wit.comInputValue = make([]*privacy.EllipticPoint, numInputCoin)
	wit.comInputSerialNumberDerivator = make([]*privacy.EllipticPoint, numInputCoin)
	// It is used for proving 2 commitments commit to the same value (input)
	//cmInputSNDIndexSK := make([]*privacy.EllipticPoint, numInputCoin)

	randInputValue := make([]*big.Int, numInputCoin)
	randInputSND := make([]*big.Int, numInputCoin)
	//randInputSNDIndexSK := make([]*big.Int, numInputCoin)

	// cmInputValueAll is sum of all input coins' value commitments
	cmInputValueAll := new(privacy.EllipticPoint)
	cmInputValueAll.Zero()
	randInputValueAll := big.NewInt(0)

	// Summing all commitments of each input coin into one commitment and proving the knowledge of its Openings
	cmInputSum := make([]*privacy.EllipticPoint, numInputCoin)
	randInputSum := make([]*big.Int, numInputCoin)
	// randInputSumAll is sum of all randomess of coin commitments
	randInputSumAll := big.NewInt(0)

	wit.oneOfManyWitness = make([]*oneoutofmany.OneOutOfManyWitness, numInputCoin)
	wit.serialNumberWitness = make([]*serialnumberprivacy.SNPrivacyWitness, numInputCoin)

	commitmentTemps := make([][]*privacy.EllipticPoint, numInputCoin)
	randInputIsZero := make([]*big.Int, numInputCoin)

	preIndex := 0

	for i, inputCoin := range wit.inputCoins {
		// commit each component of coin commitment
		randInputValue[i] = privacy.RandScalar(r)
		randInputSND[i] = privacy.RandScalar(r)

		wit.comInputValue[i] = privacy.PedCom.CommitAtIndex(new(big.Int).SetUint64(inputCoin.CoinDetails.GetValue()), randInputValue[i], privacy.PedersenValueIndex)
		wit.comInputSerialNumberDerivator[i] = privacy.PedCom.CommitAtIndex(inputCoin.CoinDetails.GetSNDerivator(), randInputSND[i], privacy.PedersenSndIndex)

		cmInputValueAll = cmInputValueAll.Add(wit.comInputValue[i])

		randInputValueAll.Add(randInputValueAll, randInputValue[i])
		randInputValueAll.Mod(randInputValueAll, privacy.Curve.Params().N)

		/***** Build witness for proving one-out-of-N commitments is a commitment to the coins being spent *****/
		cmInputSum[i] = cmInputSK.Add(wit.comInputValue[i])
		cmInputSum[i] = cmInputSum[i].Add(wit.comInputSerialNumberDerivator[i])
		cmInputSum[i] = cmInputSum[i].Add(wit.comInputShardID)

		randInputSum[i] = new(big.Int).Set(randInputSK)
		randInputSum[i].Add(randInputSum[i], randInputValue[i])
		randInputSum[i].Add(randInputSum[i], randInputSND[i])
		randInputSum[i].Add(randInputSum[i], randInputShardID)
		randInputSum[i].Mod(randInputSum[i], privacy.Curve.Params().N)

		randInputSumAll.Add(randInputSumAll, randInputSum[i])
		randInputSumAll.Mod(randInputSumAll, privacy.Curve.Params().N)

		// commitmentTemps is a list of commitments for protocol one-out-of-N
		commitmentTemps[i] = make([]*privacy.EllipticPoint, privacy.CommitmentRingSize)

		randInputIsZero[i] = big.NewInt(0)
		randInputIsZero[i].Sub(inputCoin.CoinDetails.GetRandomness(), randInputSum[i])
		randInputIsZero[i].Mod(randInputIsZero[i], privacy.Curve.Params().N)
		var err error
		for j := 0; j < privacy.CommitmentRingSize; j++ {
			commitmentTemps[i][j], err = commitments[preIndex+j].Sub(cmInputSum[i])
			if err != nil {
				return privacy.NewPrivacyErr(privacy.UnexpectedErr, err)

			}
		}

		if wit.oneOfManyWitness[i] == nil {
			wit.oneOfManyWitness[i] = new(oneoutofmany.OneOutOfManyWitness)
		}
		indexIsZero := myCommitmentIndices[i] % privacy.CommitmentRingSize

		wit.oneOfManyWitness[i].Set(commitmentTemps[i], randInputIsZero[i], indexIsZero)
		preIndex = privacy.CommitmentRingSize * (i + 1)
		// ---------------------------------------------------

		/***** Build witness for proving that serial number is derived from the committed derivator *****/
		if wit.serialNumberWitness[i] == nil {
			wit.serialNumberWitness[i] = new(serialnumberprivacy.SNPrivacyWitness)
		}
		stmt := new(serialnumberprivacy.SerialNumberPrivacyStatement)
		stmt.Set(inputCoin.CoinDetails.GetSerialNumber(), cmInputSK, wit.comInputSerialNumberDerivator[i])
		wit.serialNumberWitness[i].Set(stmt, privateKey, randInputSK, inputCoin.CoinDetails.GetSNDerivator(), randInputSND[i])
		// ---------------------------------------------------
	}

	numOutputCoin := len(wit.outputCoins)

	randOutputValue := make([]*big.Int, numOutputCoin)
	randOutputSND := make([]*big.Int, numOutputCoin)
	cmOutputValue := make([]*privacy.EllipticPoint, numOutputCoin)
	cmOutputSND := make([]*privacy.EllipticPoint, numOutputCoin)

	cmOutputSum := make([]*privacy.EllipticPoint, numOutputCoin)
	randOutputSum := make([]*big.Int, numOutputCoin)

	cmOutputSumAll := new(privacy.EllipticPoint)
	cmOutputSumAll.Zero()

	// cmOutputValueAll is sum of all value coin commitments
	cmOutputValueAll := new(privacy.EllipticPoint)
	cmOutputValueAll.Zero()
	randOutputValueAll := big.NewInt(0)

	randOutputShardID := make([]*big.Int, numOutputCoin)
	cmOutputShardID := make([]*privacy.EllipticPoint, numOutputCoin)

	for i, outputCoin := range wit.outputCoins {
		if i == len(outputCoins)-1 {
			randOutputValue[i] = new(big.Int).Sub(randInputValueAll, randOutputValueAll)
			randOutputValue[i].Mod(randOutputValue[i], privacy.Curve.Params().N)
		} else {
			randOutputValue[i] = privacy.RandScalar(r)
		}

		randOutputSND[i] = privacy.RandScalar(r)
		randOutputShardID[i] = privacy.RandScalar(r)

		cmOutputValue[i] = privacy.PedCom.CommitAtIndex(new(big.Int).SetUint64(outputCoin.CoinDetails.GetValue()), randOutputValue[i], privacy.PedersenValueIndex)
		cmOutputSND[i] = privacy.PedCom.CommitAtIndex(outputCoin.CoinDetails.GetSNDerivator(), randOutputSND[i], privacy.PedersenSndIndex)

		receiverShardID := common.GetShardIDFromLastByte(outputCoins[i].CoinDetails.GetPubKeyLastByte())
		cmOutputShardID[i] = privacy.PedCom.CommitAtIndex(big.NewInt(int64(receiverShardID)), randOutputShardID[i], privacy.PedersenShardIDIndex)

		randOutputSum[i] = big.NewInt(0)
		randOutputSum[i].Add(randOutputValue[i], randOutputSND[i])
		randOutputSum[i].Add(randOutputSum[i], randOutputShardID[i])
		randOutputSum[i].Mod(randOutputSum[i], privacy.Curve.Params().N)

		cmOutputSum[i] = new(privacy.EllipticPoint)
		cmOutputSum[i].Zero()
		cmOutputSum[i] = cmOutputValue[i].Add(cmOutputSND[i])
		cmOutputSum[i] = cmOutputSum[i].Add(outputCoins[i].CoinDetails.GetPublicKey())
		cmOutputSum[i] = cmOutputSum[i].Add(cmOutputShardID[i])

		cmOutputValueAll = cmOutputValueAll.Add(cmOutputValue[i])
		randOutputValueAll.Add(randOutputValueAll, randOutputValue[i])
		randOutputValueAll.Mod(randOutputValueAll, privacy.Curve.Params().N)

		// calculate final commitment for output coins
		outputCoins[i].CoinDetails.SetCoinCommitment(cmOutputSum[i])
		outputCoins[i].CoinDetails.SetRandomness(randOutputSum[i])

		cmOutputSumAll = cmOutputSumAll.Add(cmOutputSum[i])
	}

	// For Multi Range Protocol
	// proving each output value is less than vmax
	// proving sum of output values is less than vmax
	outputValue := make([]*big.Int, numOutputCoin)
	for i := 0; i < numOutputCoin; i++ {
		if outputCoins[i].CoinDetails.GetValue() > 0 {
			outputValue[i] = big.NewInt(int64(outputCoins[i].CoinDetails.GetValue()))
		} else {
			return privacy.NewPrivacyErr(privacy.UnexpectedErr, errors.New("output coin's value is less than 0"))
		}
	}
	if wit.aggregatedRangeWitness == nil {
		wit.aggregatedRangeWitness = new(aggregaterange.AggregatedRangeWitness)
	}
	wit.aggregatedRangeWitness.Set(outputValue, randOutputValue)
	// ---------------------------------------------------

	// save partial commitments (value, input, shardID)
	wit.comOutputValue = cmOutputValue
	wit.comOutputSerialNumberDerivator = cmOutputSND
	wit.comOutputShardID = cmOutputShardID
	return nil
}

// Prove creates big proof
func (wit *PaymentWitness) Prove(hasPrivacy bool) (*PaymentProof, *privacy.PrivacyError) {
	proof := new(PaymentProof)
	proof.Init()

	proof.inputCoins = wit.inputCoins
	proof.outputCoins = wit.outputCoins
	proof.commitmentOutputValue = wit.comOutputValue
	proof.commitmentOutputSND = wit.comOutputSerialNumberDerivator
	proof.commitmentOutputShardID = wit.comOutputShardID

	proof.commitmentInputSecretKey = wit.comInputSecretKey
	proof.commitmentInputValue = wit.comInputValue
	proof.commitmentInputSND = wit.comInputSerialNumberDerivator
	proof.commitmentInputShardID = wit.comInputShardID
	proof.commitmentIndices = wit.commitmentIndices

	// if hasPrivacy == false, don't need to create the zero knowledge proof
	// proving user has spending key corresponding with public key in input coins
	// is proved by signing with spending key
	if !hasPrivacy {
		// Proving that serial number is derived from the committed derivator
		for i := 0; i < len(wit.inputCoins); i++ {
			snNoPrivacyProof, err := wit.serialNumberNoPrivacyWitness[i].Prove(nil)
			if err != nil {
				return nil, privacy.NewPrivacyErr(privacy.ProveSerialNumberNoPrivacyErr, err)
			}
			proof.serialNumberNoPrivacyProof = append(proof.serialNumberNoPrivacyProof, snNoPrivacyProof)
		}
		return proof, nil
	}

	// if hasPrivacy == true
	numInputCoins := len(wit.oneOfManyWitness)

	for i := 0; i < numInputCoins; i++ {
		// Proving one-out-of-N commitments is a commitment to the coins being spent
		oneOfManyProof, err := wit.oneOfManyWitness[i].Prove()
		if err != nil {
			return nil, privacy.NewPrivacyErr(privacy.ProveOneOutOfManyErr, err)
		}
		proof.oneOfManyProof = append(proof.oneOfManyProof, oneOfManyProof)

		// Proving that serial number is derived from the committed derivator
		serialNumberProof, err := wit.serialNumberWitness[i].Prove(nil)
		if err != nil {
			return nil, privacy.NewPrivacyErr(privacy.ProveSerialNumberPrivacyErr, err)
		}
		proof.serialNumberProof = append(proof.serialNumberProof, serialNumberProof)
	}
	var err error

	// Proving that each output values and sum of them does not exceed v_max
	proof.aggregatedRangeProof, err = wit.aggregatedRangeWitness.Prove()
	if err != nil {
		return nil, privacy.NewPrivacyErr(privacy.ProveAggregatedRangeErr, err)
	}

	privacy.Logger.Log.Debug("Privacy log: PROVING DONE!!!")
	return proof, nil
}
