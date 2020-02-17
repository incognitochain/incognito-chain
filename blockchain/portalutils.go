package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/pkg/errors"
	"strings"
)

type CurrentPortalState struct {
	CustodianPoolState map[string]*lvdb.CustodianState // key : beaconHeight || custodian_address
	PortingRequests    map[string]*lvdb.PortingRequest // key : beaconHeight || UniquePortingID
	RedeemRequests     map[string]*lvdb.RedeemRequest  // key : beaconHeight || UniqueRedeemID
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
		IncognitoAddress: incognitoAddress,
		TotalCollateral:  totalColl,
		FreeCollateral:   freeColl,
		HoldingPubTokens: holdingPubTokens,
		LockedAmountCollateral: lockedAmountCollateral,
		RemoteAddresses:  remoteAddresses,
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

//todo: need to be updated, get all porting/redeem requests from DB
func InitCurrentPortalStateFromDB(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (*CurrentPortalState, error) {
	custodianPoolState, err := getCustodianPoolState(db, beaconHeight)
	if err != nil {
		return nil, err
	}
	portingRequestsState, err := getPortingRequestsState(db, beaconHeight)
	if err != nil {
		return nil, err
	}
	redeemRequestsState, err := getRedeemRequestsState(db, beaconHeight)
	if err != nil {
		return nil, err
	}

	return &CurrentPortalState{
		CustodianPoolState: custodianPoolState,
		PortingRequests:    portingRequestsState,
		RedeemRequests:     redeemRequestsState,
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
	err = storePortingRequestsState(db, beaconHeight, currentPortalState.PortingRequests)
	if err != nil {
		return err
	}
	err = storeRedeemRequestsState(db, beaconHeight, currentPortalState.RedeemRequests)
	if err != nil {
		return err
	}
	return nil
}

func storePortingRequestsState(db database.DatabaseInterface,
	beaconHeight uint64,
	portingRequestState map[string]*lvdb.PortingRequest) error  {
	for contribKey, contribution := range portingRequestState {
		newKey := replaceKeyByBeaconHeight(contribKey, beaconHeight)
		contributionBytes, err := json.Marshal(contribution)
		if err != nil {
			return err
		}
		err = db.Put([]byte(newKey), contributionBytes)
		if err != nil {
			return database.NewDatabaseError(database.StorePortingRequestStateError, errors.Wrap(err, "db.lvdb.put"))
		}
	}
	return nil
}

func storeRedeemRequestsState(db database.DatabaseInterface,
	beaconHeight uint64,
	redeemRequestState map[string]*lvdb.RedeemRequest) error  {
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
	custodianState map[string]*lvdb.CustodianState) error  {
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

func getPortingRequestsState(
	db database.DatabaseInterface,
	beaconHeight uint64,
) (map[string]*lvdb.PortingRequest, error) {
	portingRequestState := make(map[string]*lvdb.PortingRequest)
	portingRequestStateKeysBytes, portingRequestStateValuesBytes, err := db.GetAllRecordsPortalByPrefix(beaconHeight, lvdb.PortalPortingRequestsPrefix)
	if err != nil {
		return nil, err
	}
	for idx, portingRequestStateKeyBytes := range portingRequestStateKeysBytes {
		var portingRequest lvdb.PortingRequest
		err = json.Unmarshal(portingRequestStateValuesBytes[idx], &portingRequest)
		if err != nil {
			return nil, err
		}

		portingRequestState[string(portingRequestStateKeyBytes)] = &portingRequest
	}
	return portingRequestState, nil
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

func getAmountAdaptable(amount uint64, exchangeRate uint64) (uint64, error)  {
	convertPubTokenToPRVFloat64 := (float64(amount) * 1.5) * float64(exchangeRate)
	convertPubTokenToPRVInt64 := uint64(convertPubTokenToPRVFloat64) // 2.2 -> 2

	return convertPubTokenToPRVInt64, nil
}

func getPubTokenByTotalCollateral(total uint64, exchangeRate uint64) (uint64, error)  {
	pubToken := float64(total) / float64(exchangeRate) / 1.5
	pubTokenByCollateral := uint64(pubToken) // 2.2 -> 2

	return pubTokenByCollateral, nil
}

