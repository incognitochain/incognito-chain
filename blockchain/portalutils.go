package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/pkg/errors"
	"sort"
	"strings"
)

type CurrentPortalState struct {
	CustodianPoolState     map[string]*lvdb.CustodianState       // key : beaconHeight || custodian_address
	ExchangeRatesRequests  map[string]*lvdb.ExchangeRatesRequest // key : beaconHeight | TxID
	WaitingPortingRequests map[string]*lvdb.PortingRequest       // key : beaconHeight || UniquePortingID
	WaitingRedeemRequests  map[string]*lvdb.RedeemRequest        // key : beaconHeight || UniqueRedeemID
}


type CustodianStateSlice struct {
	Key   string
	Value *lvdb.CustodianState
}

func NewCustodianState(
	incognitoAddress string,
	totalColl uint64,
	freeColl uint64,
	holdingPubTokens map[string]uint64,
	lockedAmountCollateral map[string]uint64,
	remoteAddresses map[string]string,
) (*lvdb.CustodianState, error) {
	return &lvdb.CustodianState{
		IncognitoAddress:       incognitoAddress,
		TotalCollateral:        totalColl,
		FreeCollateral:         freeColl,
		HoldingPubTokens:       holdingPubTokens,
		LockedAmountCollateral: lockedAmountCollateral,
		RemoteAddresses:        remoteAddresses,
	}, nil
}

func NewPortingRequestState(
	uniquePortingID string,
	txReqID common.Hash,
	tokenID string,
	porterAddress string,
	amount uint64,
	custodians map[string]lvdb.MatchingPortingCustodianDetail,
	portingFee uint64,
) (*lvdb.PortingRequest, error) {
	return &lvdb.PortingRequest{
		UniquePortingID: uniquePortingID,
		TxReqID:         txReqID,
		TokenID:         tokenID,
		PorterAddress:   porterAddress,
		Amount:          amount,
		Custodians:      custodians,
		PortingFee:      portingFee,
	}, nil
}

func NewExchangeRatesState(
	senderAddress string,
	rates map[string]lvdb.ExchangeRatesDetail,
) (*lvdb.ExchangeRatesRequest, error) {
	return &lvdb.ExchangeRatesRequest{
		SenderAddress: senderAddress,
		Rates:         rates,
	}, nil
}

//todo: need to be updated, get all porting/redeem requests from DB
func InitCurrentPortalStateFromDB(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (*CurrentPortalState, error) {
	custodianPoolState, err := getCustodianPoolState(db, beaconHeight)
	if err != nil {
		return nil, err
	}
	waitingPortingReqs, err := getWaitingPortingRequests(db, beaconHeight)
	if err != nil {
		return nil, err
	}
	waitingRedeemReqs, err := getWaitingRedeemRequests(db, beaconHeight)
	if err != nil {
		return nil, err
	}

	return &CurrentPortalState{
		CustodianPoolState:     custodianPoolState,
		WaitingPortingRequests: waitingPortingReqs,
		WaitingRedeemRequests:  waitingRedeemReqs,
	}, nil
}

func storePortalStateToDB(
	db database.DatabaseInterface,
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
) error {
	err := storeCustodianState(db, beaconHeight, currentPortalState.CustodianPoolState)
	if err != nil {
		return err
	}
	err = storeWaitingPortingRequests(db, beaconHeight, currentPortalState.WaitingPortingRequests)
	if err != nil {
		return err
	}
	err = storeWaitingRedeemRequests(db, beaconHeight, currentPortalState.WaitingRedeemRequests)
	if err != nil {
		return err
	}

	return nil
}

func storeRedeemRequestsState(db database.DatabaseInterface,
	beaconHeight uint64,
	redeemRequestState map[string]*lvdb.RedeemRequest) error {
	for contribKey, contribution := range redeemRequestState {
		newKey := replaceKeyByBeaconHeight(contribKey, beaconHeight)
		contributionBytes, err := json.Marshal(contribution)
		if err != nil {
			return err
		}
		err = db.Put([]byte(newKey), contributionBytes)
		if err != nil {
			return database.NewDatabaseError(database.StoreRedeemRequestStateError, errors.Wrap(err, "db.lvdb.put"))
		}
	}
	return nil
}

func storeCustodianState(db database.DatabaseInterface,
	beaconHeight uint64,
	custodianState map[string]*lvdb.CustodianState) error {
	for contribKey, contribution := range custodianState {
		newKey := replaceKeyByBeaconHeight(contribKey, beaconHeight)
		contributionBytes, err := json.Marshal(contribution)
		if err != nil {
			return err
		}
		err = db.Put([]byte(newKey), contributionBytes)
		if err != nil {
			return database.NewDatabaseError(database.StoreCustodianDepositStateError, errors.Wrap(err, "db.lvdb.put"))
		}
	}
	return nil
}

func storeWaitingPortingRequests(db database.DatabaseInterface,
	beaconHeight uint64,
	waitingPortingReqs map[string]*lvdb.PortingRequest) error {
	//todo:
	//for contribKey, contribution := range waitingPortingReqs {
	//	newKey := replaceKeyByBeaconHeight(contribKey, beaconHeight)
	//	contributionBytes, err := json.Marshal(contribution)
	//	if err != nil {
	//		return err
	//	}
	//	err = db.Put([]byte(newKey), contributionBytes)
	//	if err != nil {
	//		return database.NewDatabaseError(database.StoreCustodianDepositStateError, errors.Wrap(err, "db.lvdb.put"))
	//	}
	//}
	return nil
}

func storeWaitingRedeemRequests(db database.DatabaseInterface,
	beaconHeight uint64,
	waitingRedeemReqs map[string]*lvdb.RedeemRequest) error {
	//todo:
	//for contribKey, contribution := range waitingRedeemReqs {
	//	newKey := replaceKeyByBeaconHeight(contribKey, beaconHeight)
	//	contributionBytes, err := json.Marshal(contribution)
	//	if err != nil {
	//		return err
	//	}
	//	err = db.Put([]byte(newKey), contributionBytes)
	//	if err != nil {
	//		return database.NewDatabaseError(database.StoreCustodianDepositStateError, errors.Wrap(err, "db.lvdb.put"))
	//	}
	//}
	return nil
}

func replaceKeyByBeaconHeight(key string, newBeaconHeight uint64) string {
	parts := strings.Split(key, "-")
	if len(parts) <= 1 {
		return key
	}
	parts[1] = fmt.Sprintf("%d", newBeaconHeight)
	newKey := ""
	for idx, part := range parts {
		if idx == len(parts)-1 {
			newKey += part
			continue
		}
		newKey += (part + "-")
	}
	return newKey
}

func getCustodianPoolState(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (map[string]*lvdb.CustodianState, error) {
	custodianPoolState := make(map[string]*lvdb.CustodianState)
	custodianPoolStateKeysBytes, custodianPoolStateValuesBytes, err := db.GetAllRecordsPortalByPrefix(beaconHeight, lvdb.PortalCustodianStatePrefix)
	if err != nil {
		return nil, err
	}
	for idx, custodianPoolStateKeyBytes := range custodianPoolStateKeysBytes {
		var custodianState lvdb.CustodianState
		err = json.Unmarshal(custodianPoolStateValuesBytes[idx], &custodianState)
		if err != nil {
			return nil, err
		}
		custodianPoolState[string(custodianPoolStateKeyBytes)] = &custodianState
	}
	return custodianPoolState, nil
}

func getRedeemRequestsState(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (map[string]*lvdb.RedeemRequest, error) {
	redeemRequestState := make(map[string]*lvdb.RedeemRequest)
	redeemRequestStateKeysBytes, redeemRequestStateValuesBytes, err := db.GetAllRecordsPortalByPrefix(beaconHeight, lvdb.PortalRedeemRequestsPrefix)
	if err != nil {
		return nil, err
	}
	for idx, portingRequestStateKeyBytes := range redeemRequestStateKeysBytes {
		var redeemRequest lvdb.RedeemRequest
		err = json.Unmarshal(redeemRequestStateValuesBytes[idx], &redeemRequest)
		if err != nil {
			return nil, err
		}

		redeemRequestState[string(portingRequestStateKeyBytes)] = &redeemRequest
	}
	return redeemRequestState, nil
}

func GetFinalExchangeRatesByKey(
	db database.DatabaseInterface,
	key []byte,
) (*lvdb.FinalExchangeRates, error) {
	finalExchangeRatesItem, err := db.GetItemPortalByPrefix(key)

	if err != nil {
		return nil, err
	}

	var finalExchangeRatesState lvdb.FinalExchangeRates

	//get value via idx
	var slice []byte
	slice = append(slice, finalExchangeRatesItem)

	err = json.Unmarshal(slice, &finalExchangeRatesState)
	if err != nil {
		return nil, err
	}

	return &finalExchangeRatesState, nil
}


func getAmountAdaptable(amount uint64, exchangeRate uint64) (uint64, error) {
	convertPubTokenToPRVFloat64 := (float64(amount) * 1.5) * float64(exchangeRate)
	convertPubTokenToPRVInt64 := uint64(convertPubTokenToPRVFloat64) // 2.2 -> 2

	return convertPubTokenToPRVInt64, nil
}

func getWaitingPortingRequests(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (map[string]*lvdb.PortingRequest, error) {
	//todo:
	//custodianPoolState := make(map[string]*lvdb.CustodianState)
	//custodianPoolStateKeysBytes, custodianPoolStateValuesBytes, err := db.GetAllRecordsPortalByPrefix(beaconHeight, lvdb.PortalCustodianStatePrefix)
	//if err != nil {
	//	return nil, err
	//}
	//for idx, custodianPoolStateKeyBytes := range custodianPoolStateKeysBytes {
	//	var custodianState lvdb.CustodianState
	//	err = json.Unmarshal(custodianPoolStateValuesBytes[idx], &custodianState)
	//	if err != nil {
	//		return nil, err
	//	}
	//	custodianPoolState[string(custodianPoolStateKeyBytes)] = &custodianState
	//}
	//return custodianPoolState, nil
	return nil, nil
}

func getWaitingRedeemRequests(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (map[string]*lvdb.RedeemRequest, error) {
	//todo:
	//custodianPoolState := make(map[string]*lvdb.CustodianState)
	//custodianPoolStateKeysBytes, custodianPoolStateValuesBytes, err := db.GetAllRecordsPortalByPrefix(beaconHeight, lvdb.PortalCustodianStatePrefix)
	//if err != nil {
	//	return nil, err
	//}
	//for idx, custodianPoolStateKeyBytes := range custodianPoolStateKeysBytes {
	//	var custodianState lvdb.CustodianState
	//	err = json.Unmarshal(custodianPoolStateValuesBytes[idx], &custodianState)
	//	if err != nil {
	//		return nil, err
	//	}
	//	custodianPoolState[string(custodianPoolStateKeyBytes)] = &custodianState
	//}
	//return custodianPoolState, nil
	return nil, nil
}


func removeWaitingPortingReqByKey (key string, state *CurrentPortalState) bool {
	if state.WaitingPortingRequests[key] != nil {
		delete(state.WaitingPortingRequests, key)
		return true
	}

	return false
}

func sortCustodianByAmountAscent(metadata metadata.PortalUserRegister, custodianState map[string]*lvdb.CustodianState, custodianStateSlice []CustodianStateSlice) {
	//convert to slice
	for k, v := range custodianState {
		//check pTokenId, select only ptokenid
		_, tokenIdExist := v.RemoteAddresses[metadata.PTokenId]
		if !tokenIdExist {
			continue
		}

		custodianStateSlice = append(custodianStateSlice, struct{
			Key string
			Value *lvdb.CustodianState
		}{
			Key: k,
			Value: v,
		})
	}

	sort.Slice(custodianStateSlice, func(i, j int) bool {
		return custodianStateSlice[i].Value.FreeCollateral <= custodianStateSlice[j].Value.FreeCollateral
	})
}

func pickSingleCustodian(metadata metadata.PortalUserRegister, exchangeRate *lvdb.FinalExchangeRates, custodianStateSlice []CustodianStateSlice) (map[string]lvdb.MatchingPortingCustodianDetail, error) {
	//pToken to PRV
	totalPTokenAfterUp150Percent := float64(metadata.RegisterAmount) * 1.5
	totalPTokenAfterUp150PercentUnit64 := uint64(totalPTokenAfterUp150Percent)

	totalPRV := exchangeRate.ExchangePToken2PRVByTokenId(metadata.PTokenId, totalPTokenAfterUp150PercentUnit64)

	for _, kv := range custodianStateSlice {
		if kv.Value.FreeCollateral >= totalPRV {
			result := make(map[string]lvdb.MatchingPortingCustodianDetail)
			result[kv.Key] = lvdb.MatchingPortingCustodianDetail{
				RemoteAddress: metadata.PTokenAddress,
				Amount: metadata.RegisterAmount,
				LockedAmountCollateral: totalPRV,
			}

			return result, nil
		}
	}

	return  nil, nil
}

func pickMultipleCustodian (metadata metadata.PortalUserRegister, exchangeRate *lvdb.FinalExchangeRates, custodianStateSlice []CustodianStateSlice) (map[string]lvdb.MatchingPortingCustodianDetail, error){
	//get multiple custodian
	var totalPubTokenAfterPick uint64

	multipleCustodian := make(map[string]lvdb.MatchingPortingCustodianDetail)
	for i := len(custodianStateSlice) - 1; i >= 0; i-- {
		custodianItem := custodianStateSlice[i]
		if totalPubTokenAfterPick >= metadata.RegisterAmount {
			break
		}

		totalPTokenExchange := exchangeRate.ExchangePRV2PTokenByTokenId(metadata.PTokenId, custodianItem.Value.FreeCollateral)

		//temp
		totalPTokenAdaptable := float64(totalPTokenExchange) / 1.5
		totalPTokenCustodianCouldBeKept := uint64(totalPTokenAdaptable) //final 2.2 -> 2

		totalPTokenAfterUp150Percent := float64(totalPTokenCustodianCouldBeKept) * 1.5
		totalPTokenAfterUp150PercentUnit64 := uint64(totalPTokenAfterUp150Percent)

		totalPRV := exchangeRate.ExchangePToken2PRVByTokenId(metadata.PTokenId, totalPTokenAfterUp150PercentUnit64) //final

		//verify collateral
		if custodianItem.Value.FreeCollateral >= totalPRV {
			multipleCustodian[custodianItem.Key] = lvdb.MatchingPortingCustodianDetail{
				RemoteAddress: metadata.PTokenAddress,
				Amount: totalPTokenCustodianCouldBeKept,
				LockedAmountCollateral: totalPRV,
			}

			totalPubTokenAfterPick = totalPubTokenAfterPick + totalPTokenCustodianCouldBeKept

			continue
		}

		Logger.log.Errorf("current portal state is nil")
		return nil, errors.New("Pick multiple custodian is fail")
	}

	//verify total amount
	var verifyTotalPubTokenAfterPick uint64
	for _, eachCustodian := range multipleCustodian {
		verifyTotalPubTokenAfterPick = verifyTotalPubTokenAfterPick + eachCustodian.Amount
	}

	if verifyTotalPubTokenAfterPick != metadata.RegisterAmount {
		return nil, errors.New("Total public token do not match")
	}

	return multipleCustodian, nil
}

func calculatePortingFees(totalPToken uint64) uint64  {
	result := 0.01 * float64(totalPToken) / 100
	return uint64(result)
}
