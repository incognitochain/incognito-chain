package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type InitTokenResponse struct {
	MetadataBase
	RequestedTxID common.Hash
}

func NewInitTokenResponse(requestedTxID common.Hash, metaType int) *InitTokenResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &InitTokenResponse{
		RequestedTxID: requestedTxID,
		MetadataBase:  metadataBase,
	}
}

func (iRes InitTokenResponse) CheckTransactionFee(tr Transaction, minFeePerKb uint64, minFeePerTx uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes InitTokenResponse) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID) in current block
	return true, nil
}

// ValidateSanityData performs the following verification:
//  1. Check transaction type
func (iRes InitTokenResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	//Step 1
	if tx.GetType() != common.TxCustomTokenPrivacyType {
		return false, false, NewMetadataTxError(InitTokenResponseValidateSanityDataError, fmt.Errorf("tx InitTokenResponse must have type `%v`", common.TxCustomTokenPrivacyType))
	}

	return false, true, nil
}

func (iRes InitTokenResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == InitTokenResponseMeta
}

func (iRes InitTokenResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *InitTokenResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

// VerifyMinerCreatedTxBeforeGettingInBlock validates if the response is a reply to an instruction from the beacon chain.
// The response is valid for a specific instruction if
//  1. the instruction has a valid metadata type
//  2. the requested txIDs match
//  3. the tokenID has not been accumulated (i.e, not seen in the current block)
//  4. the minted public key and the one in the instruction match
//  5. the minted tx random and the one in the instruction match
//  6. the minted amount and the requested amount match
//  7. the minted and requested tokens match
//
// It returns false if no instruction from the beacon satisfies the above conditions.
//
// TODO: reviewers should double-check if the above conditions are sufficient
func (iRes InitTokenResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *MintData,
	shardID byte,
	tx Transaction,
	chainRetriever ChainRetriever,
	ac *AccumulatedValues,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
) (bool, error) {
	idx := -1
	Logger.log.Infof("Number of instructions: %v\n", len(mintData.Insts))
	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not InitTokenRequest instruction
			continue
		}

		Logger.log.Infof("Currently processing instruction: %v\n", inst)

		//Step 1
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 ||
			instMetaType != strconv.Itoa(InitTokenRequestMeta) {
			continue
		}

		contentBytes, err := base64.StdEncoding.DecodeString(inst[3])
		if err != nil {
			Logger.log.Errorf("an error occurred while parsing instruction content: %v\n", err)
			continue
		}
		var acceptedInst InitTokenAcceptedInst
		err = json.Unmarshal(contentBytes, &acceptedInst)
		if err != nil {
			Logger.log.Errorf("an error occurred while parsing instruction content: %v\n", err)
			continue
		}

		//Step 2
		if iRes.RequestedTxID.String() != acceptedInst.RequestedTxID.String() {
			Logger.log.Infof("txHash mismatch: %v != %v\n", iRes.RequestedTxID.String(), acceptedInst.RequestedTxID.String())
			continue
		}

		//Step 3
		if !ac.CanProcessTokenInit(acceptedInst.TokenID) {
			Logger.log.Infof("tokenID %v has been added to accumulate values\n", acceptedInst.TokenID.String())
			continue
		}

		//Step 4 + 5 + 6 + 7
		recvPubKey, txRandom, err := coin.ParseOTAInfoFromString(acceptedInst.OTAStr, acceptedInst.TxRandomStr)
		if err != nil {
			return false, fmt.Errorf("cannot parse OTA params (%v, %v): %v", acceptedInst.OTAStr, acceptedInst.TxRandomStr, err)
		}

		_, mintedCoin, mintedTokenID, err := tx.GetTxMintData()
		if err != nil {
			return false, fmt.Errorf("cannot get minted data of txResp %v: %v", tx.Hash().String(), err)
		}

		if !bytes.Equal(mintedCoin.GetPublicKey().ToBytesS(), recvPubKey.ToBytesS()) {
			Logger.log.Infof("public keys mismatch: %v != %v\n", mintedCoin.GetPublicKey().ToBytesS(), recvPubKey.ToBytesS())
			continue
		}

		if !bytes.Equal(mintedCoin.GetTxRandom().Bytes(), txRandom.Bytes()) {
			Logger.log.Infof("txRandoms mismatch: %v != %v\n", mintedCoin.GetTxRandom().Bytes(), txRandom.Bytes())
			continue
		}

		if mintedCoin.GetValue() != acceptedInst.Amount {
			Logger.log.Infof("amounts mismatch: %v != %v\n", mintedCoin.GetValue(), acceptedInst.Amount)
			continue
		}

		if mintedTokenID.String() != acceptedInst.TokenID.String() {
			Logger.log.Infof("tokenID mismatch: %v != %v\n", mintedTokenID.String(), acceptedInst.TokenID.String())
			continue
		}

		idx = i
		ac.InitTokens = append(ac.InitTokens, &acceptedInst.TokenID)
		break
	}

	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf("no InitTokenRequest tx found for tx %s", tx.Hash().String())
	}
	mintData.InstsUsed[idx] = 1

	return true, nil
}
