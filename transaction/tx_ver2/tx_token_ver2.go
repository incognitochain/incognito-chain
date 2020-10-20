package tx_ver2

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"sort"
	"bytes"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/utils"

)

type TxTokenDataVersion2 struct {
	PropertyID     	common.Hash
	PropertyName   	string
	PropertySymbol 	string
	SigPubKey      	[]byte `json:"SigPubKey,omitempty"` // 33 bytes
	Sig            	[]byte `json:"Sig,omitempty"`       //
	Proof          	privacy.Proof

	Type     		int
	Mintable 		bool
}

func (td *TxTokenDataVersion2) Hash() (*common.Hash, error){
	tempSig := td.Sig
	tempPk := td.SigPubKey
	td.Sig = []byte{}
	td.SigPubKey = []byte{}
	inBytes, err := json.Marshal(td)
	if err!=nil{
		return nil, err
	}
	hash := common.HashH(inBytes)
	td.Sig = tempSig
	td.SigPubKey = tempPk
	return &hash, nil
}

func makeTxToken(txPRV *Tx, pubkey, sig []byte, proof privacy.Proof) Tx{
	result := Tx{
		TxBase: tx_generic.TxBase{
			Version: 	txPRV.Version,
			Type: 		txPRV.Type,
			LockTime: 	txPRV.LockTime,
			Fee: 		0,
			Info: 		[]byte{},
			PubKeyLastByteSender: txPRV.PubKeyLastByteSender,
			Metadata: 	nil,
			SigPubKey: 	pubkey,
			Sig:		sig,
			Proof: 		proof,
		},
	}
	// if there is a signing key we cached previously, use it
	result.SetPrivateKey(txPRV.GetPrivateKey())
	return result
}

type TxToken struct {
	tx 				Tx 					`json:"Tx"`
	tokenData 		TxTokenDataVersion2 `json:"TokenData"`
	cachedTxNormal	*Tx
}

func (tx *TxToken) Hash() *common.Hash{
	firstHash := tx.tx.Hash()
	secondHash, err := tx.tokenData.Hash()
	if err!=nil{
		return nil
	}
	result := common.HashH(append(firstHash[:], secondHash[:]...))
	return &result
}
func (td TxTokenDataVersion2) ToCompatTokenData(ttx metadata.Transaction) tx_generic.TxTokenData{
	return tx_generic.TxTokenData{
		TxNormal: 		ttx,
		PropertyID: 	td.PropertyID,
		PropertyName: 	td.PropertyName,
		PropertySymbol: td.PropertySymbol,
		Type: 			td.Type,
		Mintable: 		td.Mintable,
		Amount: 		0,
	}
}
func decomposeTokenData(td tx_generic.TxTokenData) (*TxTokenDataVersion2, *Tx, error){
	result := TxTokenDataVersion2{
		PropertyID: 	td.PropertyID,
		PropertyName: 	td.PropertyName,
		PropertySymbol: td.PropertySymbol,
		Type: 			td.Type,
		Mintable: 		td.Mintable,
	}
	tx, ok := td.TxNormal.(*Tx)
	if !ok{
		return nil, nil, errors.New("Error while casting a transaction to v2")
	}
	return &result, tx, nil
}
func (tx *TxToken) GetTxBase() metadata.Transaction{
	return &tx.tx
}
func (tx *TxToken) SetTxBase(inTx metadata.Transaction) error{
	temp, ok := inTx.(*Tx)
	if !ok{
		return errors.New("Cannot set TxBase : wrong type")
	}
	tx.tx = *temp
	return nil
}
func (tx *TxToken) GetTxNormal() metadata.Transaction{
	if tx.cachedTxNormal!=nil{
		return tx.cachedTxNormal
	}
	result := makeTxToken(&tx.tx, tx.tokenData.SigPubKey, tx.tokenData.Sig, tx.tokenData.Proof)
	tx.cachedTxNormal = &result
	return &result
}
func (tx *TxToken) SetTxNormal(inTx metadata.Transaction) error{
	temp, ok := inTx.(*Tx)
	if !ok{
		return errors.New("Cannot set TxNormal : wrong type")
	}
	tx.tokenData.SigPubKey = temp.SigPubKey
	tx.tokenData.Sig = temp.Sig
	tx.tokenData.Proof = temp.Proof
	tx.cachedTxNormal = nil
	return nil
}

func checkIsBridgeTokenID(bridgeStateDB *statedb.StateDB, tokenID *common.Hash) error {
	isBridgeToken := false
	allBridgeTokensBytes, err := statedb.GetAllBridgeTokens(bridgeStateDB)
	if err != nil {
		return utils.NewTransactionErr(utils.TokenIDExistedError, err)
	}
	if len(allBridgeTokensBytes) > 0 {
		var allBridgeTokens []*rawdbv2.BridgeTokenInfo
		err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)
		if err != nil {
			return utils.NewTransactionErr(utils.TokenIDExistedError, err)
		}
		for _, bridgeTokens := range allBridgeTokens {
			if tokenID.IsEqual(bridgeTokens.TokenID) {
				isBridgeToken = true
				break
			}
		}
	}
	if !isBridgeToken {
		return utils.NewTransactionErr(utils.TokenIDExistedError, errors.New("invalid Token ID"))
	}
	return nil
}

// this signs only on the hash of the data in it
func (tx *Tx) proveToken(params *tx_generic.TxPrivacyInitParams) (bool, error) {
	utils.Logger.Log.Debugf("CREATING sub-TX (token)")
	if err := tx_generic.ValidateTxParams(params); err != nil {
		return false, err
	}

	// Init tx and params (tx and params will be changed)
	if err := tx.InitializeTxAndParams(params); err != nil {
		return false, err
	}
	isBurning, err := tx.proveCA(params)
	if err != nil {
		return false, err
	}
	return isBurning, nil
}

func (txToken *TxToken) initToken(params *tx_generic.TxTokenParams) error {
	txToken.tokenData.Type = params.TokenParams.TokenTxType
	txToken.tokenData.PropertyName = params.TokenParams.PropertyName
	txToken.tokenData.PropertySymbol = params.TokenParams.PropertySymbol
	txToken.tokenData.Mintable = params.TokenParams.Mintable

	switch params.TokenParams.TokenTxType {
	case utils.CustomTokenInit:
		{
			temp := new(Tx)
			temp.SetVersion(utils.TxVersion2Number)
			temp.SetType(common.TxNormalType)
			temp.Proof = new(privacy.ProofV2)
			temp.Proof.Init()
			
			// set output coins; hash everything but commitment; save the hash to compute the new token ID later
			message := []byte{}
			if len(params.TokenParams.Receiver[0].Message) > 0 {
				if len(params.TokenParams.Receiver[0].Message) > privacy.MaxSizeInfoCoin {
					return utils.NewTransactionErr(utils.ExceedSizeInfoOutCoinError, nil)
				}
				message = params.TokenParams.Receiver[0].Message
			}
			tempPaymentInfo := &privacy.PaymentInfo{PaymentAddress: params.TokenParams.Receiver[0].PaymentAddress, Amount: params.TokenParams.Amount, Message: message}
			createdTokenCoin, errCoin := privacy.NewCoinFromPaymentInfo(tempPaymentInfo)
			if errCoin != nil {
				utils.Logger.Log.Errorf("Cannot create new coin based on payment info err %v", errCoin)
				return errCoin
			}
			if err := temp.Proof.SetOutputCoins([]privacy.Coin{createdTokenCoin}); err != nil {
				utils.Logger.Log.Errorf("Init customPrivacyToken cannot set outputCoins")
				return err
			}
			createdTokenCoin.SetCommitment(new(privacy.Point).Identity())
			hashInitToken, err := txToken.tokenData.Hash()
			if err != nil {
				utils.Logger.Log.Error(errors.New("can't hash this token data"))
				return utils.NewTransactionErr(utils.UnexpectedError, err)
			}
			
			temp.Sig, _, err = tx_generic.SignNoPrivacy(params.SenderKey, temp.Hash()[:])
			if err != nil {
				utils.Logger.Log.Error(errors.New("can't signOnMessage this tx"))
				return utils.NewTransactionErr(utils.SignTxError, err)
			}
			temp.SigPubKey = params.TokenParams.Receiver[0].PaymentAddress.Pk
			txToken.SetTxNormal(temp)
			
			var plainTokenID *common.Hash
			if params.TokenParams.Mintable {
				propertyID, err := common.Hash{}.NewHashFromStr(params.TokenParams.PropertyID)
				if err != nil {
					return utils.NewTransactionErr(utils.TokenIDInvalidError, err, propertyID.String())
				}
				plainTokenID = propertyID
			} else {
				//NOTICE: @merman update PropertyID calculated from hash of tokendata and shardID
				newHashInitToken := common.HashH(append(hashInitToken.GetBytes(), params.ShardID))
				utils.Logger.Log.Debug("New Privacy Token %+v ", newHashInitToken)
				existed := statedb.PrivacyTokenIDExisted(params.TransactionStateDB, newHashInitToken)
				if existed {
					utils.Logger.Log.Error("INIT Tx Custom Token Privacy is Existed", newHashInitToken)
					return utils.NewTransactionErr(utils.TokenIDExistedError, errors.New("this token is existed in network"))
				}
				plainTokenID = &newHashInitToken
				utils.Logger.Log.Debugf("A new token privacy wil be issued with ID: %+v", txToken.tokenData.PropertyID.String())
			}

			// fmt.Printf("While init token, its ID is %s\n", plainTokenID.String())
			// set the unblinded asset tag
			err = createdTokenCoin.SetPlainTokenID(plainTokenID)
			if err!=nil{
				return utils.NewTransactionErr(utils.UnexpectedError, err)
			}
			txToken.tokenData.PropertyID = *plainTokenID
			return nil
		}
	case utils.CustomTokenTransfer:
		{
			propertyID, _ := common.TokenStringToHash(params.TokenParams.PropertyID)
			dbFacingTokenID := common.ConfidentialAssetID
			utils.Logger.Log.Debugf("Token %+v wil be transfered with", propertyID)

			txParams := tx_generic.NewTxPrivacyInitParams(
				params.SenderKey,
				params.TokenParams.Receiver,
				params.TokenParams.TokenInput,
				params.TokenParams.Fee,
				params.HasPrivacyToken,
				params.TransactionStateDB,
				propertyID,
				nil,
				nil,
			)
			txNormal := new(Tx)
			isBurning, err := txNormal.proveToken(txParams)
			if err != nil {
				return utils.NewTransactionErr(utils.PrivacyTokenInitTokenDataError, err)
			}
			txToken.SetTxNormal(txNormal)
			if isBurning{
				// show plain tokenID if this is a burning TX
				txToken.tokenData.PropertyID = *propertyID
			}else{
				// tokenID is already hidden in asset tags in coin, here we use the umbrella ID
				txToken.tokenData.PropertyID = dbFacingTokenID
			}
			return nil
		}
	default:
		return utils.NewTransactionErr(utils.PrivacyTokenTxTypeNotHandleError, errors.New("can't handle this TokenTxType"))
	}
}

// this signs on the hash of both sub TXs
func (tx *Tx) provePRV(params *tx_generic.TxPrivacyInitParams, hashedTokenMessage []byte) error {
	outputCoins, err := utils.NewCoinV2ArrayFromPaymentInfoArray(params.PaymentInfo, params.TokenID, params.StateDB)
	if err != nil {
		utils.Logger.Log.Errorf("Cannot parse outputCoinV2 to outputCoins, error %v ", err)
		return err
	}

	// inputCoins is plainCoin because it may have coinV1 with coinV2
	inputCoins := params.InputCoins

	tx.Proof, err = privacy.ProveV2(inputCoins, outputCoins, nil, false, params.PaymentInfo)
	if err != nil {
		utils.Logger.Log.Errorf("Error in privacy_v2.Prove, error %v ", err)
		return err
	}

	if tx.ShouldSignMetaData() {
		if err := tx.signMetadata(params.SenderSK); err != nil {
			utils.Logger.Log.Error("Cannot signOnMessage txMetadata in shouldSignMetadata")
			return err
		}
	}

	// Get Hash of the whole txToken then sign on it
	message := common.HashH(append(tx.Hash()[:], hashedTokenMessage...))
	err = tx.signOnMessage(inputCoins, outputCoins, params, message[:])
	return err
}

func (txToken *TxToken) initPRV(feeTx * Tx, params *tx_generic.TxPrivacyInitParams) error {
	txTokenDataHash, err := txToken.tokenData.Hash()
	if err != nil {
		utils.Logger.Log.Errorf("Cannot calculate txPrivacyTokenData Hash, err %v", err)
		return err
	}
	if err := feeTx.provePRV(params, txTokenDataHash[:]); err != nil {
		return utils.NewTransactionErr(utils.PrivacyTokenInitPRVError, err)
	}
	// override TxCustomTokenPrivacyType type
	feeTx.SetType(common.TxCustomTokenPrivacyType)
	txToken.tx = *feeTx

	return nil
}

func (txToken *TxToken) Init(paramsInterface interface{}) error {
	params, ok := paramsInterface.(*tx_generic.TxTokenParams)
	if !ok {
		return errors.New("Cannot init TxCustomTokenPrivacy because params is not correct")
	}

	// Check validate params first, before creating tx token
	// Because there are some validation must be made first
	// Please dont change their order when you dont really understand
	txPrivacyParams := tx_generic.NewTxPrivacyInitParams(
		params.SenderKey,
		params.PaymentInfo,
		params.InputCoin,
		params.FeeNativeCoin,
		params.HasPrivacyCoin,
		params.TransactionStateDB,
		nil,
		params.MetaData,
		params.Info,
	)
	if err := tx_generic.ValidateTxParams(txPrivacyParams); err != nil {
		return err
	}
	// Init tx and params (tx and params will be changed)
	tx := new(Tx)
	if err := tx.InitializeTxAndParams(txPrivacyParams); err != nil {
		return err
	}

	// Check if this tx is nonPrivacyNonInput
	// Case 1: tx ptoken transfer with ptoken fee
	// Case 2: tx Reward
	// If it is non privacy non input then return
	if check, err := tx.IsNonPrivacyNonInput(txPrivacyParams); check {
		return err
	}

	// check tx size
	limitFee := uint64(0)
	estimateTxSizeParam := tx_generic.NewEstimateTxSizeParam(len(params.InputCoin), len(params.PaymentInfo),
		params.HasPrivacyCoin, nil, params.TokenParams, limitFee)
	if txSize := tx_generic.EstimateTxSize(estimateTxSizeParam); txSize > common.MaxTxSize {
		return utils.NewTransactionErr(utils.ExceedSizeTx, nil, strconv.Itoa(int(txSize)))
	}

	// Init Token first
	if err := txToken.initToken(params); err != nil {
		utils.Logger.Log.Errorf("Cannot init token ver2: err %v", err)
		return err
	}

	// Init PRV Fee on the whole transaction
	if err := txToken.initPRV(tx, txPrivacyParams); err != nil {
		utils.Logger.Log.Errorf("Cannot init token ver2: err %v", err)
		return err
	}

	return nil
}

func (txToken *TxToken) InitTxTokenSalary(otaCoin *privacy.CoinV2, privKey *privacy.PrivateKey, stateDB *statedb.StateDB, metaData metadata.Metadata, coinID *common.Hash, coinName string) error {
	var err error
	// Check validate params
	txPrivacyParams := tx_generic.NewTxPrivacyInitParams(
		privKey, []*privacy.PaymentInfo{}, nil, 0, false, stateDB, nil, metaData, nil,
	)
	if err := tx_generic.ValidateTxParams(txPrivacyParams); err != nil {
		return err
	}
	// check tx size
	publicKeyBytes := otaCoin.GetPublicKey().ToBytesS()
	if txSize := tx_generic.EstimateTxSizeOfInitTokenSalary(publicKeyBytes, otaCoin.GetValue(), coinName, coinID); txSize > common.MaxTxSize {
		return utils.NewTransactionErr(utils.ExceedSizeTx, nil, strconv.Itoa(int(txSize)))
	}

	// Create TxToken
	var propertyID [common.HashSize]byte
	copy(propertyID[:], coinID[:])
	txToken.tokenData.PropertyID = propertyID
	txToken.tokenData.Type = utils.CustomTokenInit
	txToken.tokenData.PropertyName = coinName
	txToken.tokenData.PropertySymbol = coinName
	txToken.tokenData.Mintable= true

	tempOutputCoin := []privacy.Coin{otaCoin}
	proof := new(privacy.ProofV2)
	proof.Init()
	if err = proof.SetOutputCoins(tempOutputCoin); err != nil {
		utils.Logger.Log.Errorf("Init customPrivacyToken cannot set outputCoins")
		return err
	}
	temp := new(Tx)
	// temp.Version = utils.TxVersion2Number
	// temp.Type = common.TxNormalType
	temp.Proof = proof
	// temp.PubKeyLastByteSender = publicKeyBytes[len(publicKeyBytes)-1]
	// signOnMessage Tx
	//temp.sigPrivKey = *privKey
	// if temp.Sig, temp.SigPubKey, err = tx_generic.SignNoPrivacy(privKey, temp.Hash()[:]); err != nil {
	// 	utils.Logger.Log.Error(errors.New("can't signOnMessage this tx"))
	// 	return utils.NewTransactionErr(utils.SignTxError, err)
	// }
	temp.Sig = []byte{}
	temp.SigPubKey = otaCoin.GetPublicKey().ToBytesS()
	txToken.SetTxNormal(temp)

	// Init tx fee params
	tx := new(Tx)
	if err := tx.InitializeTxAndParams(txPrivacyParams); err != nil {
		return err
	}
	tx.SetType(common.TxCustomTokenPrivacyType)
	tx.SetPrivateKey(*txPrivacyParams.SenderSK)

	hashedTokenMessage, err := txToken.tokenData.Hash()
	if err!=nil{
		return utils.NewTransactionErr(utils.SignTxError, err)
	}

	message := common.HashH(append(tx.Hash()[:], hashedTokenMessage[:]...))
	if tx.Sig, tx.SigPubKey, err = tx_generic.SignNoPrivacy(privKey, message[:]); err != nil {
		utils.Logger.Log.Error(errors.New(fmt.Sprintf("Cannot signOnMessage tx %v\n", err)))
		return utils.NewTransactionErr(utils.SignTxError, err)
	}

	txToken.SetTxBase(tx)
	return nil
}

func (tx *TxToken) ValidateTxSalary(db *statedb.StateDB) (bool, error) {
	// verify signature
	if valid, err := tx_generic.VerifySigNoPrivacy(tx.tx.Sig, tx.tx.SigPubKey, tx.Hash()[:]); !valid {
		if err != nil {
			utils.Logger.Log.Debugf("Error verifying signature of tx: %+v", err)
			return false, utils.NewTransactionErr(utils.VerifyTxSigFailError, err)
		}
		return false, nil
	}
	// check whether output coin's input exists in input list or not
	tokenID := tx.GetTokenID()

	// Check commitment
	outputCoins := tx.GetTxNormal().GetProof().GetOutputCoins()
	if len(outputCoins) != 1 {
		return false, utils.NewTransactionErr(utils.UnexpectedError, errors.New("length outputCoins of proof is not 1"))
	}
	outputCoin := outputCoins[0].(*privacy.CoinV2)
	cmpCommitment, err := outputCoin.ComputeCommitmentCA()
	if err!=nil || !privacy.IsPointEqual(cmpCommitment, outputCoin.GetCommitment()) {
		return false, utils.NewTransactionErr(utils.UnexpectedError, errors.New("check output coin's coin commitment isn't calculated correctly"))
	}

	// Check shardID
	coinShardID, errShard := outputCoin.GetShardID()
	if errShard != nil {
		errStr := fmt.Sprintf("error when getting coin shardID, err: %v", errShard)
		return false, utils.NewTransactionErr(utils.UnexpectedError, errors.New(errStr))
	}
	if coinShardID != common.GetShardIDFromLastByte(tx.tx.PubKeyLastByteSender) {
		return false, utils.NewTransactionErr(utils.UnexpectedError, errors.New("output coin's shardID is different from tx pubkey last byte"))
	}

	// Check database for ota
	found, err := statedb.HasOnetimeAddress(db, *tokenID, outputCoin.GetPublicKey().ToBytesS())
	if err != nil {
		utils.Logger.Log.Errorf("Cannot check public key existence in DB, err %v", err)
		return false, err
	}
	if found {
		utils.Logger.Log.Error("ValidateTxSalary got error: found onetimeaddress in database")
		return false, errors.New("found onetimeaddress in database")
	}
	return true, nil
}

func (txToken *TxToken) verifySig(transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash) (bool, error) {
	// check input transaction
	txFee := &txToken.tx
	if txFee.GetSig() == nil || txFee.GetSigPubKey() == nil {
		return false, utils.NewTransactionErr(utils.UnexpectedError, errors.New("input transaction must be a signed one"))
	}
	var err error

	// Verify TxToken Salary: NonPrivacyNonInput
	if txFee.GetProof() == nil {
		hashedTokenMessage, err := txToken.tokenData.Hash()
		if err!=nil{
			return false, err
		}
		message := common.HashH(append(txFee.Hash()[:], hashedTokenMessage[:]...))
		if valid, err := tx_generic.VerifySigNoPrivacy(txFee.GetSig(), txFee.GetSigPubKey(), message[:]); !valid {
			if err != nil {
				utils.Logger.Log.Debugf("Error verifying signature of tx: %+v", err)
				return false, utils.NewTransactionErr(utils.VerifyTxSigFailError, err)
			}
			return false, nil
		}
		return true, nil
	}

	// Reform Ring
	sumOutputCoinsWithFee := tx_generic.CalculateSumOutputsWithFee(txFee.GetProof().GetOutputCoins(), txFee.GetTxFee())
	ring, err := getRingFromSigPubKeyAndLastColumnCommitment(
		txFee.GetSigPubKey(), sumOutputCoinsWithFee,
		transactionStateDB, shardID, tokenID,
	)
	if err != nil {
		utils.Logger.Log.Errorf("Error when querying database to construct mlsag ring: %v ", err)
		return false, err
	}

	// Reform MLSAG Signature
	inputCoins := txFee.GetProof().GetInputCoins()
	keyImages := make([]*privacy.Point, len(inputCoins)+1)
	for i := 0; i < len(inputCoins); i += 1 {
		if inputCoins[i].GetKeyImage()==nil {
			utils.Logger.Log.Errorf("Error when reconstructing mlsagSignature: missing keyImage")
			return false, err
		}
		keyImages[i] = inputCoins[i].GetKeyImage()
	}
	// The last column is gone, so just fill in any value
	keyImages[len(inputCoins)] = privacy.RandomPoint()
	mlsagSignature, err := getMLSAGSigFromTxSigAndKeyImages(txFee.GetSig(), keyImages)
	if err != nil {
		utils.Logger.Log.Errorf("Error when reconstructing mlsagSignature: %v ", err)
		return false, err
	}

	txTokenDataHash, err := txToken.tokenData.Hash()
	if err != nil {
		utils.Logger.Log.Errorf("Error when getting txTokenData Hash: %v ", err)
		return false, err

	}
	message := common.HashH(append(txFee.Hash()[:], txTokenDataHash[:]...))
	return mlsag.Verify(mlsagSignature, ring, message[:])
}

func (txToken TxToken) ValidateTxByItself(hasPrivacyCoin bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, chainRetriever metadata.ChainRetriever, shardID byte, isNewTransaction bool, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever) (bool, error) {
	// check for proof, signature ...
	valid, err := txToken.ValidateTransaction(hasPrivacyCoin, transactionStateDB, bridgeStateDB, shardID, nil, false, isNewTransaction)
	if !valid {
		return false, err
	}
	valid, err = tx_generic.MdValidate(&txToken, hasPrivacyCoin, transactionStateDB, bridgeStateDB, shardID, isNewTransaction)
	if !valid {
		return false, err
	}
	return true, nil
}

func (txToken TxToken) ValidateTransaction(hasPrivacyCoin bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	var err error
	if tokenID, err = tx_generic.ParseTokenID(tokenID); err != nil {
		return false, err
	}
	ok, err := txToken.verifySig(transactionStateDB, shardID, tokenID)
	if !ok {
		utils.Logger.Log.Errorf("FAILED VERIFICATION SIGNATURE ver2 (token) with tx hash %s: %+v \n", txToken.Hash().String(), err)
		return false, utils.NewTransactionErr(utils.VerifyTxSigFailError, err)
	}else {
		// validate for pToken
		tokenID := txToken.tokenData.PropertyID
		switch txToken.tokenData.Type {
		case utils.CustomTokenInit:
			if txToken.tokenData.Mintable {
				return true, nil
			} else {
				// check exist token
				if statedb.PrivacyTokenIDExisted(transactionStateDB, tokenID) {
					return false, errors.New("Cannot validate Tx Init Token. It is tx mint from User")
				}
				return true, nil
			}
		case utils.CustomTokenTransfer:
			if txToken.GetType() == common.TxTokenConversionType {
				return validateConversionVer1ToVer2(txToken.GetTxNormal(), transactionStateDB, shardID, &tokenID)
			} else {
				resTxTokenData, err :=  txToken.GetTxNormal().ValidateTransaction(
					true,
					transactionStateDB, bridgeStateDB, shardID, &tokenID, isBatch, isNewTransaction)
				if err!= nil{
					return resTxTokenData, err
				}
				txFeeProof := txToken.tx.GetProof()
				if txFeeProof == nil {
					return resTxTokenData, nil
				}
				resTxFee, err := txFeeProof.Verify(false, txToken.tx.GetSigPubKey(), 0, shardID, &common.PRVCoinID, isBatch, nil)
				return resTxFee && resTxTokenData, err

			}
		default:
			return false, errors.New("Cannot validate Tx Token. Unavailable type")
		}
	}
}

func (txToken TxToken) ValidateSanityData(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) (bool, error) {
	if txToken.GetTxBase().GetProof() == nil && txToken.GetTxNormal().GetProof() == nil {
		return false, errors.New("Tx Privacy Ver 2 must have a proof")
	}
	if txToken.GetTokenID().String() == common.PRVCoinID.String(){
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, errors.New("cannot transfer PRV via txtoken"))
	}
	// validate metadata
	check, err := tx_generic.MdValidateSanity(&txToken.tx, chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	if !check || err != nil {
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, err)
	}
	// validate sanity for tx pToken + metadata
	check, err = tx_generic.ValidateSanity(txToken.GetTxNormal(), chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	if !check || err != nil {
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, err)
	}
	// validate sanity for tx pToken + without metadata
	check1, err1 := tx_generic.ValidateSanity(&txToken.tx, chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	if !check1 || err1 != nil {
		return false, utils.NewTransactionErr(utils.InvalidSanityDataPrivacyTokenError, err1)
	}
	return true, nil
}

// TODO : update this
func (txToken TxToken) GetTxActualSize() uint64 {
	sizeTx := tx_generic.GetTxActualSizeInBytes(&txToken.tx)

	sizeTx += tx_generic.GetTxActualSizeInBytes(txToken.GetTxNormal())
	sizeTx += uint64(len(txToken.tokenData.PropertyName))
	sizeTx += uint64(len(txToken.tokenData.PropertySymbol))
	sizeTx += uint64(len(txToken.tokenData.PropertyID))
	sizeTx += 4 // Type
	sizeTx += 1 // Mintable
	sizeTx += 8 // Amount

	meta := txToken.tx.GetMetadata()
	if meta != nil {
		sizeTx += meta.CalculateSize()
	}

	result := uint64(math.Ceil(float64(sizeTx) / 1024))
	return result
}

//-- OVERRIDE--
func (tx TxToken) GetVersion() int8 { return tx.tx.Version }

func (tx *TxToken) SetVersion(version int8) { tx.tx.Version = version }

func (tx TxToken) GetMetadataType() int {
	if tx.tx.Metadata != nil {
		return tx.tx.Metadata.GetType()
	}
	return metadata.InvalidMeta
}

func (tx TxToken) GetType() string { return tx.tx.Type }

func (tx *TxToken) SetType(t string) { tx.tx.Type = t }

func (tx TxToken) GetLockTime() int64 { return tx.tx.LockTime }

func (tx *TxToken) SetLockTime(locktime int64) { tx.tx.LockTime = locktime }

func (tx TxToken) GetSenderAddrLastByte() byte { return tx.tx.PubKeyLastByteSender }

func (tx *TxToken) SetGetSenderAddrLastByte(b byte) { tx.tx.PubKeyLastByteSender = b }

func (tx TxToken) GetTxFee() uint64 { return tx.tx.Fee }

func (tx *TxToken) SetTxFee(fee uint64) { tx.tx.Fee = fee }

func (tx TxToken) GetTxFeeToken() uint64 { return uint64(0) }

func (tx TxToken) GetInfo() []byte { return tx.tx.Info }

func (tx *TxToken) SetInfo(info []byte) { tx.tx.Info = info }

// not supported
func (tx TxToken) GetSigPubKey() []byte { return []byte{} }
func (tx *TxToken) SetSigPubKey(sigPubkey []byte) {  }
func (tx TxToken) GetSig() []byte { return []byte{} }
func (tx *TxToken) SetSig(sig []byte) {}
func (tx TxToken) GetProof() privacy.Proof { return nil }
func (tx *TxToken) SetProof(proof privacy.Proof) {}
func (tx TxToken) GetCachedActualSize() *uint64{
	return nil
}
func (tx *TxToken) SetCachedActualSize(sz *uint64){}

func (tx TxToken) GetCachedHash() *common.Hash{
	return nil
}
func (tx *TxToken) SetCachedHash(h *common.Hash){}
func (tx *TxToken) Verify(hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	return false, nil
}

func (tx TxToken) GetTokenID() *common.Hash { return &tx.tokenData.PropertyID }

func (tx TxToken) GetMetadata() metadata.Metadata { return tx.tx.Metadata }

func (tx *TxToken) SetMetadata(meta metadata.Metadata) { tx.tx.Metadata = meta }
func (tx TxToken) GetPrivateKey() []byte{
	return tx.tx.GetPrivateKey()
}
func (tx *TxToken) SetPrivateKey(sk []byte){
	tx.tx.SetPrivateKey(sk)
}

func (tx TxToken) GetReceivers() ([][]byte, []uint64) {
	return nil, nil
}

func (tx TxToken) ListSerialNumbersHashH() []common.Hash {
	result := []common.Hash{}
	if tx.tx.GetProof() != nil {
		for _, d := range tx.tx.GetProof().GetInputCoins() {
			hash := common.HashH(d.GetKeyImage().ToBytesS())
			result = append(result, hash)
		}
	}
	if tx.GetTxNormal().GetProof() != nil {
		for _, d := range tx.GetTxNormal().GetProof().GetInputCoins() {
			hash := common.HashH(d.GetKeyImage().ToBytesS())
			result = append(result, hash)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})
	return result
}

func (tx TxToken) String() string {
	jsb, err := json.Marshal(tx)
	if err!=nil{
		return ""
	}
	return string(jsb)
	// record := strconv.Itoa(int(tx.tx.Version))
	// record += strconv.FormatInt(tx.tx.LockTime, 10)
	// record += strconv.FormatUint(tx.tx.Fee, 10)
	// if tx.Proof != nil {
	// 	record += base64.StdEncoding.EncodeToString(tx.tx.Proof.Bytes())
	// }
	// if tx.Metadata != nil {
	// 	metadataHash := tx.Metadata.Hash()
	// 	record += metadataHash.String()
	// }
	// return record
}
func (tx *TxToken) CalculateTxValue() uint64 {
	proof := tx.GetTxNormal().GetProof()
	if proof == nil {
		return 0
	}
	if proof.GetOutputCoins() == nil || len(proof.GetOutputCoins()) == 0 {
		return 0
	}
	if proof.GetInputCoins() == nil || len(proof.GetInputCoins()) == 0 { // coinbase tx
		txValue := uint64(0)
		for _, outCoin := range proof.GetOutputCoins() {
			txValue += outCoin.GetValue()
		}
		return txValue
	}

	if tx.GetTxNormal().IsPrivacy() {
		return 0
	}

	senderPKBytes := proof.GetInputCoins()[0].GetPublicKey().ToBytesS()
	txValue := uint64(0)
	for _, outCoin := range proof.GetOutputCoins() {
		outPKBytes := outCoin.GetPublicKey().ToBytesS()
		if bytes.Equal(senderPKBytes, outPKBytes) {
			continue
		}
		txValue += outCoin.GetValue()
	}
	return txValue
}

func (tx TxToken) CheckTxVersion(maxTxVersion int8) bool {
	return !(tx.tx.Version > maxTxVersion)
}
func (tx TxToken) ShouldSignMetaData() bool {
	if tx.tx.GetMetadata() == nil {
		return false
	}
	return tx.tx.GetMetadata().ShouldSignMetaData()
}

func (tx TxToken) IsSalaryTx() bool {
	if tx.tx.GetType() != common.TxRewardType {
		return false
	}
	if len(tx.tokenData.Proof.GetInputCoins()) > 0 {
		return false
	}
	return true
}

func (tx TxToken) IsPrivacy() bool {
	// In the case of NonPrivacyNonInput, we do not have proof
	if tx.tx.Proof == nil {
		return false
	}
	return tx.tx.Proof.IsPrivacy()
}

func (txToken *TxToken) IsCoinsBurning(bcr metadata.ChainRetriever, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) bool {
	// get proof of pToken
	proof := txToken.GetTxNormal().GetProof()
	if proof == nil || len(proof.GetOutputCoins()) == 0 {
		return false
	}
	return txToken.GetTxNormal().IsCoinsBurning(bcr, retriever, viewRetriever, beaconHeight)
}

func (txToken *TxToken) CheckAuthorizedSender([]byte) (bool, error) {
	return false, errors.New("TxToken does not have CheckAuthorizedSender")
}

func (tx *TxToken) GetReceiverData() ([]privacy.Coin, error) {
	if tx.tx.Proof != nil && len(tx.tx.Proof.GetOutputCoins()) > 0 {
		return tx.tx.Proof.GetOutputCoins(), nil
	}
	return nil, nil
}

func (txToken *TxToken) GetTransferData() (bool, []byte, uint64, *common.Hash) {
	pubkeys, amounts := txToken.GetTxNormal().GetReceivers()
	if len(pubkeys) == 0 {
		utils.Logger.Log.Error("GetTransferData receive 0 output, it should has exactly 1 output")
		return false, nil, 0, &txToken.tokenData.PropertyID
	}
	if len(pubkeys) > 1 {
		utils.Logger.Log.Error("GetTransferData receiver: More than 1 receiver")
		return false, nil, 0, &txToken.tokenData.PropertyID
	}
	return true, pubkeys[0], amounts[0], &txToken.tokenData.PropertyID
}

func (txToken TxToken) ValidateType() bool {
	return txToken.tx.GetType() == common.TxCustomTokenPrivacyType
}

func (txToken *TxToken) ValidateTxWithCurrentMempool(mr metadata.MempoolRetriever) error {
	poolSerialNumbersHashH := mr.GetSerialNumbersHashH()
	err := txToken.validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH)
	if err != nil {
		utils.Logger.Log.Error(err)
		return utils.NewTransactionErr(utils.DoubleSpendError, err)
	}
	// TODO: will move this to mempool process
	if txToken.tokenData.Type == utils.CustomTokenInit && txToken.tx.GetMetadata() == nil {
		initTokenID := txToken.tokenData.PropertyID
		txsInMem := mr.GetTxsInMem()
		for _, tx := range txsInMem {
			// try parse to TxTokenBase
			var tokenTx, ok = tx.Tx.(tx_generic.TransactionToken)
			if ok {
				txTokenData := tokenTx.GetTxTokenData()
				if txTokenData.Type == utils.CustomTokenInit && tokenTx.GetMetadata() == nil {
					// check > 1 tx init token by the same token ID
					if txTokenData.PropertyID.IsEqual(&initTokenID) {
						return utils.NewTransactionErr(utils.TokenIDInvalidError, fmt.Errorf("had already tx for initing token ID %s in pool", txTokenData.PropertyID.String()), txTokenData.PropertyID.String())
					}
				}
			}
		}
	}
	return nil
}

func (txToken *TxToken) validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH map[common.Hash][]common.Hash) error {
	// check proof of PRV and pToken
	if txToken.tx.GetProof() == nil && txToken.GetTxNormal().GetProof() == nil {
		return errors.New("empty tx")
	}

	// collect serial number for PRV
	temp := make(map[common.Hash]interface{})
	if txToken.tx.GetProof() != nil {
		for _, desc := range txToken.tx.GetProof().GetInputCoins() {
			hash := common.HashH(desc.GetKeyImage().ToBytesS())
			temp[hash] = nil
		}
	}
	// collect serial number for pToken
	txNormalProof := txToken.GetTxNormal().GetProof()
	if txNormalProof != nil {
		for _, desc := range txNormalProof.GetInputCoins() {
			hash := common.HashH(desc.GetKeyImage().ToBytesS())
			temp[hash] = nil
		}
	}

	// check with pool serial number in mempool
	for _, listSerialNumbers := range poolSerialNumbersHashH {
		for _, serialNumberHash := range listSerialNumbers {
			if _, ok := temp[serialNumberHash]; ok {
				return errors.New("double spend")
			}
		}
	}
	return nil
}

func (txToken *TxToken) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *metadata.MintData, shardID byte, bcr metadata.ChainRetriever, accumulatedValues *metadata.AccumulatedValues, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever) (bool, error) {
	if !txToken.tokenData.Mintable {
		return true, nil
	}
	meta := txToken.tx.GetMetadata()
	if meta == nil {
		utils.Logger.Log.Error("Mintable custom token must contain metadata")
		return false, nil
	}
	if !meta.IsMinerCreatedMetaType() {
		return false, nil
	}
	return meta.VerifyMinerCreatedTxBeforeGettingInBlock(mintData, shardID, txToken, bcr, accumulatedValues, retriever, viewRetriever)
}

func (txToken *TxToken) GetTxTokenData() tx_generic.TxTokenData { return txToken.tokenData.ToCompatTokenData(txToken.GetTxNormal()) }
func (txToken *TxToken) SetTxTokenData(data tx_generic.TxTokenData) error{ 
	td, txN, err := decomposeTokenData(data)
	if err == nil{
		txToken.tokenData = *td
		return txToken.SetTxNormal(txN)
	}
	return err
}

func (txToken *TxToken) GetTxMintData() (bool, privacy.Coin, *common.Hash, error) {
	tokenID := txToken.tokenData.PropertyID
	return tx_generic.GetTxMintData(txToken.GetTxNormal(), &tokenID)
}

func (txToken *TxToken) GetTxBurnData() (bool, privacy.Coin, *common.Hash, error) {
	tokenID := txToken.tokenData.PropertyID
	isBurn, burnCoin, _, err := txToken.GetTxNormal().GetTxBurnData()
	return isBurn, burnCoin, &tokenID, err
}

func (tx *TxToken) ValidateDoubleSpendWithBlockchain(shardID byte, stateDB *statedb.StateDB, tokenID *common.Hash) error {
	prvCoinID := &common.Hash{}
	err := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return err
	}
	if tokenID != nil {
		err := prvCoinID.SetBytes(tokenID.GetBytes())
		if err != nil {
			return err
		}
	}
	if tx.tx.Proof == nil {
		return nil
	}
	err = tx.tx.ValidateDoubleSpendWithBlockchain(shardID, stateDB, nil)
	if err!=nil{
		return err
	}
	if tx.GetTxNormal().GetProof() == nil {
		return nil
	}
	err = tx.GetTxNormal().ValidateDoubleSpendWithBlockchain(shardID, stateDB, prvCoinID)
	return err
}

func (txToken *TxToken) ValidateTxWithBlockChain(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, shardID byte, stateDB *statedb.StateDB) error {
	err := txToken.ValidateDoubleSpendWithBlockchain(shardID, stateDB, txToken.GetTokenID())
	return err
}

func (tx *TxToken) ValidateTxReturnStaking(stateDB *statedb.StateDB) bool { return true }



// func (txToken *TxToken) UnmarshalJSON(data []byte) error {
// 	var err error
// 	txToken.Tx = &Tx{}
// 	if err = json.Unmarshal(data, txToken.Tx); err != nil {
// 		return err
// 	}

// 	temp := &struct {
// 		TxTokenData tx_generic.TxTokenData `json:"TxTokenPrivacyData"`
// 	}{}
// 	temp.tokenData.TxNormal = &Tx{}
// 	err = json.Unmarshal(data, &temp)
// 	if err != nil {
// 		utils.Logger.Log.Error(err)
// 		return utils.NewTransactionErr(utils.PrivacyTokenJsonError, err)
// 	}
// 	txToken.tokenData = temp.tokenData
// 	return nil
// }