package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"math"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
)

// TxCustomTokenPrivacy is class tx which is inherited from P tx(supporting privacy) for fee
// and contain data(with supporting privacy format) to support issuing and transfer a custom token(token from end-user, look like erc-20)
// Dev or end-user can use this class tx to create an token type which use personal purpose
// TxCustomTokenPrivacy is an advance format of TxNormalToken
// so that user need to spend a lot fee to create this class tx
type TxCustomTokenPrivacy struct {
	Tx                                    // inherit from normal tx of P(supporting privacy) with a high fee to ensure that tx could contain a big data of privacy for token
	TxPrivacyTokenData TxPrivacyTokenData `json:"TxTokenPrivacyData"` // supporting privacy format
	// private field, not use for json parser, only use as temp variable
	cachedHash *common.Hash // cached hash data of tx
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) UnmarshalJSON(data []byte) error {
	tx := Tx{}
	err := json.Unmarshal(data, &tx)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(PrivacyTokenPRVJsonError, err)
	}
	temp := &struct {
		TxTokenPrivacyData interface{}
	}{}
	err = json.Unmarshal(data, &temp)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(PrivacyTokenJsonError, err)
	}
	TxTokenPrivacyDataJson, err := json.MarshalIndent(temp.TxTokenPrivacyData, "", "\t")
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(UnexpectedError, err)
	}
	err = json.Unmarshal(TxTokenPrivacyDataJson, &txCustomTokenPrivacy.TxPrivacyTokenData)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(PrivacyTokenJsonError, err)
	}
	txCustomTokenPrivacy.Tx = tx
	return nil
}

func (txCustomTokenPrivacy TxCustomTokenPrivacy) String() string {
	// get hash of tx
	record := txCustomTokenPrivacy.Tx.Hash().String()
	// add more hash of tx custom token data privacy
	tokenPrivacyDataHash, _ := txCustomTokenPrivacy.TxPrivacyTokenData.Hash()
	record += tokenPrivacyDataHash.String()
	if txCustomTokenPrivacy.Metadata != nil {
		record += string(txCustomTokenPrivacy.Metadata.Hash()[:])
	}
	return record
}

func (txCustomTokenPrivacy TxCustomTokenPrivacy) JSONString() string {
	data, err := json.MarshalIndent(txCustomTokenPrivacy, "", "\t")
	if err != nil {
		Logger.Log.Error(err)
		return ""
	}
	return string(data)
}

// Hash returns the hash of all fields of the transaction
func (txCustomTokenPrivacy *TxCustomTokenPrivacy) Hash() *common.Hash {
	if txCustomTokenPrivacy.cachedHash != nil {
		return txCustomTokenPrivacy.cachedHash
	}
	// final hash
	hash := common.HashH([]byte(txCustomTokenPrivacy.String()))
	return &hash
}

// GetTxActualSize computes the virtual size of a given transaction
// size of this tx = (normal TxNormal size) + (custom token data size)
func (txCustomTokenPrivacy TxCustomTokenPrivacy) GetTxActualSize() uint64 {
	normalTxSize := txCustomTokenPrivacy.Tx.GetTxActualSize()
	tokenDataSize := uint64(0)
	tokenDataSize += txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.GetTxActualSize()
	tokenDataSize += uint64(len(txCustomTokenPrivacy.TxPrivacyTokenData.PropertyName))
	tokenDataSize += uint64(len(txCustomTokenPrivacy.TxPrivacyTokenData.PropertySymbol))
	tokenDataSize += uint64(len(txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID))
	tokenDataSize += 4 // for TxPrivacyTokenData.Type
	tokenDataSize += 8 // for TxPrivacyTokenData.Amount
	meta := txCustomTokenPrivacy.Metadata
	if meta != nil {
		tokenDataSize += meta.CalculateSize()
	}

	return normalTxSize + uint64(math.Ceil(float64(tokenDataSize)/1024))
}

func (tx TxCustomTokenPrivacy) GetTxPrivacyTokenActualSize() uint64 {
	tokenDataSize := uint64(0)
	tokenDataSize += tx.TxPrivacyTokenData.TxNormal.GetTxActualSize()
	tokenDataSize += uint64(len(tx.TxPrivacyTokenData.PropertyName))
	tokenDataSize += uint64(len(tx.TxPrivacyTokenData.PropertySymbol))
	tokenDataSize += uint64(len(tx.TxPrivacyTokenData.PropertyID))
	tokenDataSize += 4 // for TxPrivacyTokenData.Type
	tokenDataSize += 8 // for TxPrivacyTokenData.Amount

	meta := tx.TxPrivacyTokenData.TxNormal.Metadata
	if meta != nil {
		tokenDataSize += meta.CalculateSize()
	}

	return uint64(math.Ceil(float64(tokenDataSize) / 1024))
}

type TxPrivacyTokenInitParams struct {
	senderKey          *privacy.PrivateKey
	paymentInfo        []*privacy.PaymentInfo
	inputCoin          []coin.PlainCoin
	feeNativeCoin      uint64
	tokenParams        *CustomTokenPrivacyParamTx
	transactionStateDB *statedb.StateDB
	bridgeStateDB      *statedb.StateDB
	metaData           metadata.Metadata
	hasPrivacyCoin     bool
	hasPrivacyToken    bool
	shardID            byte
	info               []byte
}

func NewTxPrivacyTokenInitParams(senderKey *privacy.PrivateKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoin []coin.PlainCoin,
	feeNativeCoin uint64,
	tokenParams *CustomTokenPrivacyParamTx,
	transactionStateDB *statedb.StateDB,
	metaData metadata.Metadata,
	hasPrivacyCoin bool,
	hasPrivacyToken bool,
	shardID byte,
	info []byte,
	bridgeStateDB *statedb.StateDB) *TxPrivacyTokenInitParams {
	params := &TxPrivacyTokenInitParams{
		shardID:            shardID,
		paymentInfo:        paymentInfo,
		metaData:           metaData,
		transactionStateDB: transactionStateDB,
		bridgeStateDB:      bridgeStateDB,
		feeNativeCoin:      feeNativeCoin,
		hasPrivacyCoin:     hasPrivacyCoin,
		hasPrivacyToken:    hasPrivacyToken,
		inputCoin:          inputCoin,
		senderKey:          senderKey,
		tokenParams:        tokenParams,
		info:               info,
	}
	return params
}

// Init -  build normal tx component and privacy custom token data
func (txCustomTokenPrivacy *TxCustomTokenPrivacy) Init(params *TxPrivacyTokenInitParams) error {
	var err error
	// init data for tx PRV for fee
	normalTx := Tx{}
	fmt.Println("Fee =", params.feeNativeCoin)
	err = normalTx.Init(NewTxPrivacyInitParams(
		params.senderKey,
		params.paymentInfo,
		params.inputCoin,
		params.feeNativeCoin,
		params.hasPrivacyCoin,
		params.transactionStateDB,
		nil,
		params.metaData,
		params.info))
	if err != nil {
		return NewTransactionErr(PrivacyTokenInitPRVError, err)
	}
	// override TxCustomTokenPrivacyType type
	normalTx.Type = common.TxCustomTokenPrivacyType
	txCustomTokenPrivacy.Tx = normalTx

	// check tx size
	limitFee := uint64(0)
	estimateTxSizeParam := NewEstimateTxSizeParam(len(params.inputCoin), len(params.paymentInfo),
		params.hasPrivacyCoin, nil, params.tokenParams, limitFee)
	if txSize := EstimateTxSize(estimateTxSizeParam); txSize > common.MaxTxSize {
		return NewTransactionErr(ExceedSizeTx, nil, strconv.Itoa(int(txSize)))
	}

	// check action type and create privacy custom toke data
	var handled = false
	// Add token data component
	switch params.tokenParams.TokenTxType {
	case CustomTokenInit:
		// case init a new privacy custom token
		{
			handled = true
			txCustomTokenPrivacy.TxPrivacyTokenData = TxPrivacyTokenData{
				Type:           params.tokenParams.TokenTxType,
				PropertyName:   params.tokenParams.PropertyName,
				PropertySymbol: params.tokenParams.PropertySymbol,
				Amount:         params.tokenParams.Amount,
			}

			temp := Tx{}
			temp.Version = txVersion2
			temp.Type = common.TxNormalType

			// Amount, Randomness, SharedRandom is transparency until we call concealData
			message := []byte{}
			// Set Info
			if len(params.tokenParams.Receiver[0].Message) > 0 {
				if len(params.tokenParams.Receiver[0].Message) > coin.MaxSizeInfoCoin {
					return NewTransactionErr(ExceedSizeInfoOutCoinError, nil)
				}
				message = params.tokenParams.Receiver[0].Message
			}
			tempPaymentInfo := key.InitPaymentInfo(params.tokenParams.Receiver[0].PaymentAddress, params.tokenParams.Amount, message)
			c, errCoin := coin.NewCoinFromPaymentInfo(tempPaymentInfo)
			if errCoin != nil {
				Logger.Log.Errorf("Cannot create new coin based on payment info err %v", errCoin)
				return errCoin
			}
			tempOutputCoin := make([]coin.Coin, 1)
			tempOutputCoin[0] = c
			proof := new(privacy.ProofV2)
			proof.Init()
			if err = proof.SetOutputCoins(tempOutputCoin); err != nil {
				Logger.Log.Errorf("Init customPrivacyToken cannot set outputCoins")
				return err
			}

			temp.Proof = proof
			temp.PubKeyLastByteSender = params.tokenParams.Receiver[0].PaymentAddress.Pk[len(params.tokenParams.Receiver[0].PaymentAddress.Pk)-1]

			// sign Tx
			temp.SigPubKey = params.tokenParams.Receiver[0].PaymentAddress.Pk
			temp.sigPrivKey = *params.senderKey
			err = signTx(&temp)
			if err != nil {
				Logger.Log.Error(errors.New("can't sign this tx"))
				return NewTransactionErr(SignTxError, err)
			}

			txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal = temp
			hashInitToken, err := txCustomTokenPrivacy.TxPrivacyTokenData.Hash()
			if err != nil {
				Logger.Log.Error(errors.New("can't hash this token data"))
				return NewTransactionErr(UnexpectedError, err)
			}

			if params.tokenParams.Mintable {
				propertyID, err := common.Hash{}.NewHashFromStr(params.tokenParams.PropertyID)
				if err != nil {
					return NewTransactionErr(TokenIDInvalidError, err, propertyID.String())
				}
				txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID = *propertyID
				txCustomTokenPrivacy.TxPrivacyTokenData.Mintable = true
			} else {
				//NOTICE: @merman update PropertyID calculated from hash of tokendata and shardID
				newHashInitToken := common.HashH(append(hashInitToken.GetBytes(), params.shardID))
				Logger.Log.Debug("New Privacy Token %+v ", newHashInitToken)
				existed := statedb.PrivacyTokenIDExisted(params.transactionStateDB, newHashInitToken)
				if existed {
					Logger.Log.Error("INIT Tx Custom Token Privacy is Existed", newHashInitToken)
					return NewTransactionErr(TokenIDExistedError, errors.New("this token is existed in network"))
				}
				txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID = newHashInitToken
				Logger.Log.Debugf("A new token privacy wil be issued with ID: %+v", txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID.String())
			}
		}
	case CustomTokenTransfer:
		{
			handled = true
			// make a transfering for privacy custom token
			// fee always 0 and reuse function of normal tx for custom token ID
			temp := Tx{}

			propertyID, _ := common.Hash{}.NewHashFromStr(params.tokenParams.PropertyID)
			existed := statedb.PrivacyTokenIDExisted(params.transactionStateDB, *propertyID)
			if !existed {
				isBridgeToken := false
				allBridgeTokensBytes, err := statedb.GetAllBridgeTokens(params.bridgeStateDB)
				if err != nil {
					return NewTransactionErr(TokenIDExistedError, err)
				}
				if len(allBridgeTokensBytes) > 0 {
					var allBridgeTokens []*rawdbv2.BridgeTokenInfo
					err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)
					if err != nil {
						return NewTransactionErr(TokenIDExistedError, err)
					}
					for _, bridgeTokens := range allBridgeTokens {
						if propertyID.IsEqual(bridgeTokens.TokenID) {
							isBridgeToken = true
							break
						}
					}
				}
				if !isBridgeToken {
					return NewTransactionErr(TokenIDExistedError, errors.New("invalid Token ID"))
				}
			}
			Logger.Log.Debugf("Token %+v wil be transfered with", propertyID)
			txCustomTokenPrivacy.TxPrivacyTokenData = TxPrivacyTokenData{
				Type:           params.tokenParams.TokenTxType,
				PropertyName:   params.tokenParams.PropertyName,
				PropertySymbol: params.tokenParams.PropertySymbol,
				PropertyID:     *propertyID,
				Mintable:       params.tokenParams.Mintable,
			}
			//fmt.Println("Checking input coins")
			//fmt.Println(len(params.tokenParams.TokenInput))
			for i := 0; i < len(params.tokenParams.TokenInput); i += 1 {
				//fmt.Println("Version =", params.tokenParams.TokenInput[i].GetVersion())
				//fmt.Println("Commitment =", params.tokenParams.TokenInput[i].GetCommitment())
				//fmt.Println("Value =", params.tokenParams.TokenInput[i].GetValue())

				KI := params.tokenParams.TokenInput[i].GetKeyImage()
				fmt.Println("KeyImage =", KI)

				//ok, err := statedb.HasSerialNumber(params.transactionStateDB, *propertyID, KI.ToBytesS(), 0)
			}

			err := temp.Init(NewTxPrivacyInitParams(params.senderKey,
				params.tokenParams.Receiver,
				params.tokenParams.TokenInput,
				params.tokenParams.Fee,
				params.hasPrivacyToken,
				params.transactionStateDB,
				propertyID,
				nil,
				nil))
			if err != nil {
				return NewTransactionErr(PrivacyTokenInitTokenDataError, err)
			}
			txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal = temp
		}
	}
	if !handled {
		return NewTransactionErr(PrivacyTokenTxTypeNotHandleError, errors.New("can't handle this TokenTxType"))
	}
	return nil
}

// ValidateType - check type of tx
func (txCustomTokenPrivacy TxCustomTokenPrivacy) ValidateType() bool {
	return txCustomTokenPrivacy.Type == common.TxCustomTokenPrivacyType
}

// ValidateTxWithCurrentMempool - validate for serrial number use in tx is double with other tx in mempool
func (txCustomTokenPrivacy TxCustomTokenPrivacy) ValidateTxWithCurrentMempool(mr metadata.MempoolRetriever) error {
	poolSerialNumbersHashH := mr.GetSerialNumbersHashH()
	err := txCustomTokenPrivacy.validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(DoubleSpendError, err)
	}
	// TODO: will move this to mempool process
	if txCustomTokenPrivacy.TxPrivacyTokenData.Type == CustomTokenInit && txCustomTokenPrivacy.GetMetadata() == nil {
		initTokenID := txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID
		txsInMem := mr.GetTxsInMem()
		for _, tx := range txsInMem {
			// try parse to TxCustomTokenPrivacy
			var privacyTokenTx, ok = tx.Tx.(*TxCustomTokenPrivacy)
			if ok && privacyTokenTx.TxPrivacyTokenData.Type == CustomTokenInit && privacyTokenTx.GetMetadata() == nil {
				// check > 1 tx init token by the same token ID
				if privacyTokenTx.TxPrivacyTokenData.PropertyID.IsEqual(&initTokenID) {
					return NewTransactionErr(TokenIDInvalidError, fmt.Errorf("had already tx for initing token ID %s in pool", privacyTokenTx.TxPrivacyTokenData.PropertyID.String()), privacyTokenTx.TxPrivacyTokenData.PropertyID.String())
				}
			}
		}
	}

	return nil
}

// validateDoubleSpendTxWithCurrentMempool - from proof of PRV and pToken,
// check serrial numbers is valid,
// not double spend with any tx in mempool
// this a private func -> call by ValidateTxWithCurrentMempool
func (txCustomTokenPrivacy TxCustomTokenPrivacy) validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH map[common.Hash][]common.Hash) error {
	// check proof of PRV and pToken
	if txCustomTokenPrivacy.Proof == nil && txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof == nil {
		return errors.New("empty tx")
	}

	// collect serial number for PRV
	temp := make(map[common.Hash]interface{})
	if txCustomTokenPrivacy.Proof != nil {
		for _, desc := range txCustomTokenPrivacy.Proof.GetInputCoins() {
			hash := common.HashH(desc.GetKeyImage().ToBytesS())
			temp[hash] = nil
		}
	}
	// collect serial number for pToken
	if txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof != nil {
		for _, desc := range txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof.GetInputCoins() {
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

func (txCustomTokenPrivacy TxCustomTokenPrivacy) ValidateTxWithBlockChain(
	bcr metadata.BlockchainRetriever,
	shardID byte,
	stateDB *statedb.StateDB,
) error {
	err := txCustomTokenPrivacy.ValidateDoubleSpendWithBlockchain(bcr, shardID, stateDB, nil)
	if err != nil {
		return NewTransactionErr(InvalidDoubleSpendPRVError, err)
	}
	err = txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.ValidateDoubleSpendWithBlockchain(bcr, shardID, stateDB, txCustomTokenPrivacy.GetTokenID())
	if err != nil {
		return NewTransactionErr(InvalidDoubleSpendPrivacyTokenError, err)
	}
	return nil
}

// ValidateSanityData - validate sanity data of PRV and pToken
func (txCustomTokenPrivacy TxCustomTokenPrivacy) ValidateSanityData(bcr metadata.BlockchainRetriever, beaconHeight uint64) (bool, error) {
	meta := txCustomTokenPrivacy.Tx.Metadata
	if meta != nil {
		isContinued, ok, err := meta.ValidateSanityData(bcr, &txCustomTokenPrivacy, beaconHeight)
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}

	// validate sanity data for PRV
	//result, err := txCustomTokenPrivacy.Tx.validateNormalTxSanityData()
	result, err := txCustomTokenPrivacy.Tx.ValidateSanityData(bcr, beaconHeight)
	if err != nil {
		return result, NewTransactionErr(InvalidSanityDataPRVError, err)
	}
	// validate sanity for pToken

	//result, err = txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.validateNormalTxSanityData()
	result, err = txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.ValidateSanityData(bcr, beaconHeight)
	if err != nil {
		return result, NewTransactionErr(InvalidSanityDataPrivacyTokenError, err)
	}
	return result, nil
}

// ValidateTxByItself - validate tx by itself, check signature, proof,... and metadata
func (txCustomTokenPrivacy TxCustomTokenPrivacy) ValidateTxByItself(
	hasPrivacyCoin bool,
	transactionStateDB *statedb.StateDB,
	bridgeStateDB *statedb.StateDB,
	bcr metadata.BlockchainRetriever,
	shardID byte,
	isNewTransaction bool,
) (bool, error) {
	// no need to check for tx init token
	if txCustomTokenPrivacy.TxPrivacyTokenData.Type == CustomTokenInit {
		return txCustomTokenPrivacy.Tx.ValidateTransaction(hasPrivacyCoin, transactionStateDB, bridgeStateDB, shardID, nil, false, isNewTransaction)
	}
	// check for proof, signature ...
	if ok, err := txCustomTokenPrivacy.ValidateTransaction(hasPrivacyCoin, transactionStateDB, bridgeStateDB, shardID, nil, false, isNewTransaction); !ok {
		return false, err
	}
	// check for metadata
	if txCustomTokenPrivacy.Metadata != nil {
		validateMetadata := txCustomTokenPrivacy.Metadata.ValidateMetadataByItself()
		if !validateMetadata {
			return validateMetadata, NewTransactionErr(UnexpectedError, errors.New("Metadata is invalid"))
		}
		return validateMetadata, nil
	}
	return true, nil
}

// ValidateTransaction - verify proof, signature, ... of PRV and pToken
func (txCustomTokenPrivacy *TxCustomTokenPrivacy) ValidateTransaction(hasPrivacyCoin bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	// validate for PRV
	ok, err := txCustomTokenPrivacy.Tx.ValidateTransaction(hasPrivacyCoin, transactionStateDB, bridgeStateDB, shardID, nil, isBatch, isNewTransaction)
	if ok {
		// validate for pToken
		tokenID := txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID
		if txCustomTokenPrivacy.TxPrivacyTokenData.Type == CustomTokenInit {
			if txCustomTokenPrivacy.Type == common.TxRewardType && txCustomTokenPrivacy.TxPrivacyTokenData.Mintable {
				isBridgeCentralizedToken, _ := statedb.IsBridgeTokenExistedByType(bridgeStateDB, tokenID, true)
				isBridgeDecentralizedToken, _ := statedb.IsBridgeTokenExistedByType(bridgeStateDB, tokenID, false)
				if isBridgeCentralizedToken || isBridgeDecentralizedToken {
					return true, nil
				}
				return false, nil
			} else {
				// check exist token
				if statedb.PrivacyTokenIDExisted(transactionStateDB, tokenID) {
					return false, nil
				}
				return true, nil
			}
		} else {
			return txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.ValidateTransaction(txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.IsPrivacy(), transactionStateDB, bridgeStateDB, shardID, &tokenID, isBatch, isNewTransaction)
		}
	}
	return false, err
}

// GetProof - return proof PRV of tx
func (txCustomTokenPrivacy TxCustomTokenPrivacy) GetProof() privacy.Proof {
	return txCustomTokenPrivacy.Proof
}

// VerifyMinerCreatedTxBeforeGettingInBlock
func (txCustomTokenPrivacy TxCustomTokenPrivacy) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []metadata.Transaction,
	txsUsed []int,
	insts [][]string,
	instsUsed []int,
	shardID byte,
	bcr metadata.BlockchainRetriever,
	accumulatedValues *metadata.AccumulatedValues,
) (bool, error) {
	if !txCustomTokenPrivacy.TxPrivacyTokenData.Mintable {
		return true, nil
	}
	meta := txCustomTokenPrivacy.Metadata
	if meta == nil {
		Logger.Log.Error("Mintable custom token must contain metadata")
		return false, nil
	}
	if !meta.IsMinerCreatedMetaType() {
		return false, nil
	}
	return meta.VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock, txsUsed, insts, instsUsed, shardID, &txCustomTokenPrivacy, bcr, accumulatedValues)
}

// GetTokenReceivers - return receivers in tx, who receive token
func (txCustomTokenPrivacy TxCustomTokenPrivacy) GetTokenReceivers() ([][]byte, []uint64) {
	pubkeys := [][]byte{}
	amounts := []uint64{}
	// get proof pToken
	if txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof == nil {
		return pubkeys, amounts
	}
	// fetch payment info
	for _, coin := range txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof.GetOutputCoins() {
		coinPubKey := coin.GetPublicKey().ToBytesS()
		added := false
		// coinPubKey := vout.PaymentAddress.Pk
		for i, key := range pubkeys {
			if bytes.Equal(coinPubKey, key) {
				added = true
				amounts[i] += coin.GetValue()
				break
			}
		}
		if !added {
			pubkeys = append(pubkeys, coinPubKey)
			amounts = append(amounts, coin.GetValue())
		}
	}
	return pubkeys, amounts
}

// GetTokenUniqueReceiver
func (txCustomTokenPrivacy TxCustomTokenPrivacy) GetTokenUniqueReceiver() (bool, []byte, uint64) {
	sender := []byte{}
	if txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof == nil {
		return false, []byte{}, 0
	}
	proof := txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof
	if len(proof.GetInputCoins()) > 0 && proof.GetInputCoins()[0] != nil {
		sender = proof.GetInputCoins()[0].GetPublicKey().ToBytesS()
	}
	pubkeys, amounts := txCustomTokenPrivacy.GetTokenReceivers()
	pubkey := []byte{}
	amount := uint64(0)
	count := 0
	for i, pk := range pubkeys {
		if !bytes.Equal(pk, sender) {
			pubkey = pk
			amount = amounts[i]
			count += 1
		}
	}
	return count == 1, pubkey, amount
}

// GetTransferData
func (txCustomTokenPrivacy TxCustomTokenPrivacy) GetTransferData() (bool, []byte, uint64, *common.Hash) {
	pubkeys, amounts := txCustomTokenPrivacy.GetReceivers()
	if len(pubkeys) == 0 {
		Logger.Log.Error("GetTransferData receive 0 output, it should has exactly 1 output")
		return false, nil, 0, &common.PRVCoinID
	}
	if len(pubkeys) > 1 {
		Logger.Log.Error("GetTransferData receiver: More than 1 receiver")
		return false, nil, 0, &txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID
	}
	return true, pubkeys[0], amounts[0], &txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID
}

// IsCoinsBurning - checking this is a burning pToken
func (txCustomTokenPrivacy TxCustomTokenPrivacy) IsCoinsBurning(bcr metadata.BlockchainRetriever, beaconHeight uint64) bool {
	// get proof of pToken
	proof := txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof
	if proof == nil || len(proof.GetOutputCoins()) == 0 {
		return false
	}
	//  validate receiver with burning address
	senderPKBytes := []byte{}
	if len(proof.GetInputCoins()) > 0 {
		senderPKBytes = proof.GetInputCoins()[0].GetPublicKey().ToBytesS()
	}

	//get burning address
	burningAddress := bcr.GetBurningAddress(beaconHeight)
	keyWalletBurningAccount, err := wallet.Base58CheckDeserialize(burningAddress)
	if err != nil {
		Logger.Log.Errorf("Can not deserialize burn address: %v\n", burningAddress)
		return false
	}

	keysetBurningAccount := keyWalletBurningAccount.KeySet
	paymentAddressBurningAccount := keysetBurningAccount.PaymentAddress
	for _, outCoin := range proof.GetOutputCoins() {
		outPKBytes := outCoin.GetPublicKey().ToBytesS()
		if !bytes.Equal(senderPKBytes, outPKBytes) && !bytes.Equal(outPKBytes, paymentAddressBurningAccount.Pk[:]) {
			return false
		}
	}
	return true
}

// CalculateBurnAmount - get tx value for pToken
func (txCustomTokenPrivacy TxCustomTokenPrivacy) CalculateTxValue() uint64 {
	if txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof == nil {
		return 0
	}
	proof := txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof
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

	if txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.IsPrivacy() {
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

func (txCustomTokenPrivacy TxCustomTokenPrivacy) ListSerialNumbersHashH() []common.Hash {
	tx := txCustomTokenPrivacy.Tx
	result := []common.Hash{}
	if tx.Proof != nil {
		for _, d := range tx.Proof.GetInputCoins() {
			hash := common.HashH(d.GetKeyImage().ToBytesS())
			result = append(result, hash)
		}
	}
	customTokenPrivacy := txCustomTokenPrivacy.TxPrivacyTokenData
	if customTokenPrivacy.TxNormal.Proof != nil {
		for _, d := range customTokenPrivacy.TxNormal.Proof.GetInputCoins() {
			hash := common.HashH(d.GetKeyImage().ToBytesS())
			result = append(result, hash)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})
	return result
}

// GetSigPubKey - return sig pubkey for pToken
func (txCustomTokenPrivacy TxCustomTokenPrivacy) GetSigPubKey() []byte {
	return txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.SigPubKey
}

// GetTxFeeToken - return Token Fee use to pay for privacy token Tx
func (tx TxCustomTokenPrivacy) GetTxFeeToken() uint64 {
	return tx.TxPrivacyTokenData.TxNormal.Fee
}

func (txCustomTokenPrivacy TxCustomTokenPrivacy) GetTokenID() *common.Hash {
	return &txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID
}

// GetTxFee - return fee PRV of Tx which contain privacy token Tx
func (tx TxCustomTokenPrivacy) GetTxFee() uint64 {
	return tx.Tx.GetTxFee()
}

type TxPrivacyTokenInitParamsForASM struct {
	//senderKey       *privacy.PrivateKey
	//paymentInfo     []*privacy.PaymentInfo
	//inputCoin       []*coin.PlainCoinV1
	//feeNativeCoin   uint64
	//tokenParams     *CustomTokenPrivacyParamTx
	//transactionStateDB              database.DatabaseInterface
	//metaData        metadata.Metadata
	//hasPrivacyCoin  bool
	//hasPrivacyToken bool
	//shardID         byte
	//info            []byte

	txParam                           TxPrivacyTokenInitParams
	commitmentIndicesForNativeToken   []uint64
	commitmentBytesForNativeToken     [][]byte
	myCommitmentIndicesForNativeToken []uint64
	sndOutputsForNativeToken          []*privacy.Scalar

	commitmentIndicesForPToken   []uint64
	commitmentBytesForPToken     [][]byte
	myCommitmentIndicesForPToken []uint64
	sndOutputsForPToken          []*privacy.Scalar
}

func (param *TxPrivacyTokenInitParamsForASM) SetMetaData(meta metadata.Metadata) {
	param.txParam.metaData = meta
}

func NewTxPrivacyTokenInitParamsForASM(
	senderKey *privacy.PrivateKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoin []coin.PlainCoin,
	feeNativeCoin uint64,
	tokenParams *CustomTokenPrivacyParamTx,
	metaData metadata.Metadata,
	hasPrivacyCoin bool,
	hasPrivacyToken bool,
	shardID byte,
	info []byte,
	commitmentIndicesForNativeToken []uint64,
	commitmentBytesForNativeToken [][]byte,
	myCommitmentIndicesForNativeToken []uint64,
	sndOutputsForNativeToken []*privacy.Scalar,

	commitmentIndicesForPToken []uint64,
	commitmentBytesForPToken [][]byte,
	myCommitmentIndicesForPToken []uint64,
	sndOutputsForPToken []*privacy.Scalar) *TxPrivacyTokenInitParamsForASM {

	txParam := NewTxPrivacyTokenInitParams(senderKey, paymentInfo, inputCoin, feeNativeCoin, tokenParams, nil, metaData, hasPrivacyCoin, hasPrivacyToken, shardID, info, nil)
	params := &TxPrivacyTokenInitParamsForASM{
		txParam:                           *txParam,
		commitmentIndicesForNativeToken:   commitmentIndicesForNativeToken,
		commitmentBytesForNativeToken:     commitmentBytesForNativeToken,
		myCommitmentIndicesForNativeToken: myCommitmentIndicesForNativeToken,
		sndOutputsForNativeToken:          sndOutputsForNativeToken,

		commitmentIndicesForPToken:   commitmentIndicesForPToken,
		commitmentBytesForPToken:     commitmentBytesForPToken,
		myCommitmentIndicesForPToken: myCommitmentIndicesForPToken,
		sndOutputsForPToken:          sndOutputsForPToken,
	}
	return params
}

// TODO Privacy, please make it interface coin and plaincoin
// Init -  build normal tx component and privacy custom token data
func (txCustomTokenPrivacy *TxCustomTokenPrivacy) InitForASM(params *TxPrivacyTokenInitParamsForASM, serverTime int64) error {
	var err error
	// init data for tx PRV for fee
	normalTx := Tx{}
	//err = normalTx.InitForASM(NewTxPrivacyInitParamsForASM(
	//	params.txParam.senderKey,
	//	params.txParam.paymentInfo,
	//	params.txParam.inputCoin,
	//	params.txParam.feeNativeCoin,
	//	params.txParam.hasPrivacyCoin,
	//	nil,
	//	params.txParam.metaData,
	//	params.txParam.info,
	//	params.commitmentIndicesForNativeToken,
	//	params.commitmentBytesForNativeToken,
	//	params.myCommitmentIndicesForNativeToken,
	//	params.sndOutputsForNativeToken,
	//), serverTime)
	err = normalTx.InitForASM(NewTxPrivacyInitParamsForASM(
		params.txParam.senderKey,
		params.txParam.paymentInfo,
		params.txParam.inputCoin,
		params.txParam.feeNativeCoin,
		params.txParam.hasPrivacyCoin,
		nil,
		params.txParam.metaData,
		params.txParam.info,
		params.commitmentIndicesForNativeToken,
		params.commitmentBytesForNativeToken,
		params.myCommitmentIndicesForNativeToken,
		params.sndOutputsForNativeToken,
	))
	if err != nil {
		return NewTransactionErr(PrivacyTokenInitPRVError, err)
	}

	// override TxCustomTokenPrivacyType type
	normalTx.Type = common.TxCustomTokenPrivacyType
	txCustomTokenPrivacy.Tx = normalTx

	// check action type and create privacy custom toke data
	var handled = false
	// Add token data component
	switch params.txParam.tokenParams.TokenTxType {
	case CustomTokenInit:
		// case init a new privacy custom token
		{
			handled = true
			txCustomTokenPrivacy.TxPrivacyTokenData = TxPrivacyTokenData{
				Type:           params.txParam.tokenParams.TokenTxType,
				PropertyName:   params.txParam.tokenParams.PropertyName,
				PropertySymbol: params.txParam.tokenParams.PropertySymbol,
				Amount:         params.txParam.tokenParams.Amount,
			}

			// issue token with data of privacy
			temp := Tx{}
			temp.Type = common.TxNormalType
			temp.Proof = privacy.NewProofWithVersion(txCustomTokenPrivacy.Version)
			temp.Proof.Init()
			tempOutputCoin := make([]coin.Coin, 1)
			c := new(coin.CoinV1)
			c.CoinDetails.SetValue(params.txParam.tokenParams.Amount)
			PK, err := new(privacy.Point).FromBytesS(params.txParam.tokenParams.Receiver[0].PaymentAddress.Pk)
			if err != nil {
				return NewTransactionErr(DecompressPaymentAddressError, err)
			}
			c.CoinDetails.SetPublicKey(PK)
			c.CoinDetails.SetRandomness(privacy.RandomScalar())

			// set info coin for output coin
			if len(params.txParam.tokenParams.Receiver[0].Message) > 0 {
				if len(params.txParam.tokenParams.Receiver[0].Message) > privacy.MaxSizeInfoCoin {
					return NewTransactionErr(ExceedSizeInfoOutCoinError, nil)
				}
				c.CoinDetails.SetInfo(params.txParam.tokenParams.Receiver[0].Message)
			}
			sndOut := privacy.RandomScalar()
			c.CoinDetails.SetSNDerivator(sndOut)
			err = c.CoinDetails.CommitAll()
			if err != nil {
				return NewTransactionErr(CommitOutputCoinError, err)
			}

			tempOutputCoin[0] = c
			if err = temp.Proof.SetOutputCoins(tempOutputCoin); err != nil {
				Logger.Log.Errorf("TxPrivacyToken InitASM cannot set output coins: err %v", err)
				return err
			}

			// get last byte
			temp.PubKeyLastByteSender = params.txParam.tokenParams.Receiver[0].PaymentAddress.Pk[len(params.txParam.tokenParams.Receiver[0].PaymentAddress.Pk)-1]

			// sign Tx
			temp.SigPubKey = params.txParam.tokenParams.Receiver[0].PaymentAddress.Pk
			temp.sigPrivKey = *params.txParam.senderKey
			err = signTx(&temp)
			if err != nil {
				Logger.Log.Error(errors.New("can't sign this tx"))
				return NewTransactionErr(SignTxError, err)
			}

			txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal = temp
			hashInitToken, err := txCustomTokenPrivacy.TxPrivacyTokenData.Hash()
			if err != nil {
				Logger.Log.Error(errors.New("can't hash this token data"))
				return NewTransactionErr(UnexpectedError, err)
			}

			if params.txParam.tokenParams.Mintable {
				propertyID, err := common.Hash{}.NewHashFromStr(params.txParam.tokenParams.PropertyID)
				if err != nil {
					return NewTransactionErr(TokenIDInvalidError, err, propertyID.String())
				}
				txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID = *propertyID
				txCustomTokenPrivacy.TxPrivacyTokenData.Mintable = true
			} else {
				//NOTICE: @merman update PropertyID calculated from hash of tokendata and shardID
				newHashInitToken := common.HashH(append(hashInitToken.GetBytes(), params.txParam.shardID))
				txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID = newHashInitToken
			}
		}
	case CustomTokenTransfer:
		{
			handled = true
			// make a transfering for privacy custom token
			// fee always 0 and reuse function of normal tx for custom token ID
			temp := Tx{}
			propertyID, _ := common.Hash{}.NewHashFromStr(params.txParam.tokenParams.PropertyID)
			txCustomTokenPrivacy.TxPrivacyTokenData = TxPrivacyTokenData{
				Type:           params.txParam.tokenParams.TokenTxType,
				PropertyName:   params.txParam.tokenParams.PropertyName,
				PropertySymbol: params.txParam.tokenParams.PropertySymbol,
				PropertyID:     *propertyID,
				Mintable:       params.txParam.tokenParams.Mintable,
			}
			//err := temp.InitForASM(NewTxPrivacyInitParamsForASM(
			//	params.txParam.senderKey,
			//	params.txParam.tokenParams.Receiver,
			//	params.txParam.tokenParams.TokenInput,
			//	params.txParam.tokenParams.Fee,
			//	params.txParam.hasPrivacyToken,
			//	propertyID,
			//	nil,
			//	params.txParam.info,
			//	params.commitmentIndicesForPToken,
			//	params.commitmentBytesForPToken,
			//	params.myCommitmentIndicesForPToken,
			//	params.sndOutputsForPToken,
			//), serverTime)
			err := temp.InitForASM(NewTxPrivacyInitParamsForASM(
				params.txParam.senderKey,
				params.txParam.tokenParams.Receiver,
				params.txParam.tokenParams.TokenInput,
				params.txParam.tokenParams.Fee,
				params.txParam.hasPrivacyToken,
				propertyID,
				nil,
				params.txParam.info,
				params.commitmentIndicesForPToken,
				params.commitmentBytesForPToken,
				params.myCommitmentIndicesForPToken,
				params.sndOutputsForPToken,
			))
			if err != nil {
				return NewTransactionErr(PrivacyTokenInitTokenDataError, err)
			}
			txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal = temp
		}
	}

	if !handled {
		return NewTransactionErr(PrivacyTokenTxTypeNotHandleError, errors.New("can't handle this TokenTxType"))
	}
	return nil
}
