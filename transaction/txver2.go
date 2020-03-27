package transaction

import (
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/common/base58"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
)

type TxVersion2 struct{}

func generateMlsagRing(params *TxPrivacyInitParams, pi int, shardID byte) (*mlsag.Ring, error) {
	outputCoinsPtr, err := parseOutputCoins(params)
	if err != nil {
		return nil, err
	}
	outputCoins := *outputCoinsPtr
	inputCoins := params.inputCoins

	// loop to create list usable commitments from usableInputCoins
	listUsableCommitments := make(map[common.Hash][]byte)
	listUsableCommitmentsIndices := make([]common.Hash, len(inputCoins))
	// tick index of each usable commitment with full db commitments
	mapIndexCommitmentsInUsableTx := make(map[string]*big.Int)
	for i, in := range inputCoins {
		usableCommitment := in.CoinDetails.GetCoinCommitment().ToBytesS()
		commitmentInHash := common.HashH(usableCommitment)
		listUsableCommitments[commitmentInHash] = usableCommitment
		listUsableCommitmentsIndices[i] = commitmentInHash
		index, err := statedb.GetCommitmentIndex(params.stateDB, *params.tokenID, usableCommitment, shardID)
		if err != nil {
			Logger.Log.Error(err)
			return nil, err
		}
		commitmentInBase58Check := base58.Base58Check{}.Encode(usableCommitment, common.ZeroByte)
		mapIndexCommitmentsInUsableTx[commitmentInBase58Check] = index
	}
	lenCommitment, err := statedb.GetCommitmentLength(params.stateDB, *params.tokenID, shardID)
	if err != nil {
		Logger.Log.Error(err)
		return nil, err
	}
	if lenCommitment == nil {
		Logger.Log.Error(errors.New("Commitments is empty"))
		return nil, errors.New("Commitments is empty")
	}

	outputCommitments := new(operation.Point).Identity()
	for i := 0; i < len(outputCoins); i += 1 {
		outputCommitments.Add(outputCommitments, outputCoins[i].CoinDetails.GetCoinCommitment())
	}

	ring := make([][]*operation.Point, privacy.RingSize)
	key := params.senderSK
	for i := 0; i < privacy.RingSize; i += 1 {
		sumInputs := new(operation.Point).Identity()
		row := make([]*operation.Point, len(inputCoins))
		if i == pi {
			for j := 0; j < len(inputCoins); j += 1 {
				privKey := new(operation.Scalar).FromBytesS(*key)
				row[j] = new(operation.Point).ScalarMultBase(privKey)
				sumInputs.Add(sumInputs, inputCoins[j].CoinDetails.GetCoinCommitment())
			}
		} else {
			for j := 0; j < len(inputCoins); j += 1 {
				index, _ := common.RandBigIntMaxRange(lenCommitment)
				ok, err := statedb.HasCommitmentIndex(params.stateDB, *params.tokenID, index.Uint64(), shardID)
				if ok && err == nil {
					commitment, publicKey, _ := statedb.GetCommitmentAndPublicKeyByIndex(params.stateDB, *params.tokenID, index.Uint64(), shardID)
					if _, found := listUsableCommitments[common.HashH(commitment)]; !found {
						row[j], err = new(operation.Point).FromBytesS(publicKey)
						if err != nil {
							return nil, err
						}

						temp, err := new(operation.Point).FromBytesS(commitment)
						if err != nil {
							return nil, err
						}

						sumInputs.Add(sumInputs, temp)
					}
				}
			}
		}
		row = append(row, sumInputs.Sub(sumInputs, outputCommitments))
		ring = append(ring, row)
	}
	mlsagring := mlsag.NewRing(ring)
	return mlsagring, nil
}

func createPrivKeyMlsag(params *TxPrivacyInitParams) *[]*operation.Scalar {
	outputCoinsPtr, _ := parseOutputCoins(params)
	outputCoins := *outputCoinsPtr
	inputCoins := params.inputCoins

	sumRand := new(operation.Scalar).FromUint64(0)
	for _, in := range inputCoins {
		sumRand.Add(sumRand, in.CoinDetails.GetRandomness())
	}
	for _, out := range outputCoins {
		sumRand.Add(sumRand, out.CoinDetails.GetRandomness())
	}

	sk := new(operation.Scalar).FromBytesS(*params.senderSK)
	privKeyMlsag := make([]*operation.Scalar, len(inputCoins)+1)
	for i := 0; i < len(inputCoins); i += 1 {
		privKeyMlsag[i] = sk
	}
	privKeyMlsag[len(inputCoins)] = sumRand
	return &privKeyMlsag
}

// signTx - signs tx
func signTxVer2(tx *Tx, params *TxPrivacyInitParams) error {
	if tx.Sig != nil {
		return NewTransactionErr(UnexpectedError, errors.New("input transaction must be an unsigned one"))
	}

	var pi int = common.RandIntInterval(0, privacy.RingSize-1)
	shardID := common.GetShardIDFromLastByte(tx.PubKeyLastByteSender)
	ring, err := generateMlsagRing(params, pi, shardID)
	if err != nil {
		return err
	}
	privKeysMlsag := *createPrivKeyMlsag(params)

	sag := mlsag.NewMlsag(privKeysMlsag, ring, pi)

	tx.sigPrivKey, err = privacy.ArrayScalarToBytes(&privKeysMlsag)
	if err != nil {
		return err
	}

	tx.SigPubKey, err = ring.ToBytes()
	if err != nil {
		return err
	}

	message := (*tx.Proof).Bytes()
	mlsagSignature, err := sag.Sign(message)
	if err != nil {
		return err
	}

	tx.Sig, err = mlsagSignature.ToBytes()
	return err
}

func (*TxVersion2) Prove(tx *Tx, params *TxPrivacyInitParams) error {
	outputCoins, err := parseOutputCoins(params)
	if err != nil {
		return err
	}
	inputCoins := &params.inputCoins

	var conversion privacy.Proof
	conversion, err = privacy_v2.Prove(inputCoins, outputCoins, params.hasPrivacy)
	if err != nil {
		return err
	}
	tx.Proof = &conversion

	err = signTxVer2(tx, params)
	return err
}

func (*TxVersion2) Verify(tx *Tx, hasPrivacy bool, transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	return true, nil
}
