package transaction

import (
	"errors"
	"fmt"
	"math/big"

	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"

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
	conversion, err = privacy_v2.Prove(inputCoins, outputCoins, params.hasPrivacy, &params.paymentInfo)
	if err != nil {
		return err
	}
	tx.Proof = &conversion

	err = signTxVer2(tx, params)
	return err
}

// verifySigTx - verify signature on tx
func verifySigTxVer2(tx *Tx) (bool, error) {
	// check input transaction
	if tx.Sig == nil || tx.SigPubKey == nil {
		return false, NewTransactionErr(UnexpectedError, errors.New("input transaction must be an signed one"))
	}
	var err error

	ring, err := new(mlsag.Ring).FromBytes(tx.SigPubKey)
	if err != nil {
		return false, err
	}

	txSig, err := new(mlsag.MlsagSig).FromBytes(tx.Sig)
	if err != nil {
		return false, err
	}

	message := (*tx.Proof).Bytes()
	return mlsag.Verify(txSig, ring, message)
}

// TODO privacy
func (*TxVersion2) Verify(tx *Tx, hasPrivacy bool, transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	var valid bool
	var err error

	if valid, err := verifySigTxVer2(tx); !valid {
		if err != nil {
			Logger.Log.Errorf("Error verifying signature ver2 with tx hash %s: %+v \n", tx.Hash().String(), err)
			return false, NewTransactionErr(VerifyTxSigFailError, err)
		}
		Logger.Log.Errorf("FAILED VERIFICATION SIGNATURE ver2 with tx hash %s", tx.Hash().String())
		return false, NewTransactionErr(VerifyTxSigFailError, fmt.Errorf("FAILED VERIFICATION SIGNATURE ver2 with tx hash %s", tx.Hash().String()))
	}

	if tx.Proof == nil {
		return true, nil
	}

	tokenID, err = parseTokenID(tokenID)
	if err != nil {
		return false, err
	}
	txProof := *tx.Proof
	inputCoins := txProof.GetInputCoins()
	outputCoins := txProof.GetOutputCoins()
	if err := validateSndFromOutputCoin(outputCoins); err != nil {
		return false, err
	}

	if isNewTransaction {
		for i := 0; i < len(outputCoins); i++ {
			// Check output coins' SND is not exists in SND list (Database)
			if ok, err := CheckSNDerivatorExistence(tokenID, outputCoins[i].CoinDetails.GetSNDerivator(), transactionStateDB); ok || err != nil {
				if err != nil {
					Logger.Log.Error(err)
				}
				Logger.Log.Errorf("snd existed: %d\n", i)
				return false, NewTransactionErr(SndExistedError, err, fmt.Sprintf("snd existed: %d\n", i))
			}
		}
	}

	if !hasPrivacy {
		// Check input coins' commitment is exists in cm list (Database)
		for i := 0; i < len(inputCoins); i++ {
			ok, err := tx.CheckCMExistence(inputCoins[i].CoinDetails.GetCoinCommitment().ToBytesS(), transactionStateDB, shardID, tokenID)
			if !ok || err != nil {
				if err != nil {
					Logger.Log.Error(err)
				}
				return false, NewTransactionErr(InputCommitmentIsNotExistedError, err)
			}
		}
	}
	// Verify the payment proof
	var p interface{} = txProof
	var txProofV1 privacy.ProofV1 = p.(privacy.ProofV1)
	commitments, err := getCommitmentsInDatabase(&txProofV1, hasPrivacy, tx.SigPubKey, tx.Fee, transactionStateDB, shardID, tokenID, isBatch)
	if err != nil {
		return false, err
	}

	valid, err = (*tx.Proof).Verify(hasPrivacy, tx.SigPubKey, tx.Fee, shardID, tokenID, isBatch, commitments)

	if !valid {
		if err != nil {
			Logger.Log.Error(err)
		}
		Logger.Log.Error("FAILED VERIFICATION PAYMENT PROOF")
		err1, ok := err.(*privacy.PrivacyError)
		if ok {
			// parse error detail
			if err1.Code == privacy.ErrCodeMessage[errhandler.VerifyOneOutOfManyProofFailedErr].Code {
				if isNewTransaction {
					return false, NewTransactionErr(VerifyOneOutOfManyProofFailedErr, err1, tx.Hash().String())
				} else {
					// for old txs which be get from sync block or validate new block
					if tx.LockTime <= ValidateTimeForOneoutOfManyProof {
						// only verify by sign on block because of issue #504(that mean we should pass old tx, which happen before this issue)
						return true, nil
					} else {
						return false, NewTransactionErr(VerifyOneOutOfManyProofFailedErr, err1, tx.Hash().String())
					}
				}
			}
		}
		return false, NewTransactionErr(TxProofVerifyFailError, err, tx.Hash().String())
	}
	Logger.Log.Debugf("SUCCESSED VERIFICATION PAYMENT PROOF ")
	return true, nil
}
