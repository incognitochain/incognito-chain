package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type RedeemRequest struct {
	UniqueRedeemID        string
	TokenID               string
	RedeemerAddress       string
	RedeemerRemoteAddress string
	RedeemAmount          uint64
	Custodians            []*MatchingRedeemCustodianDetail
	RedeemFee             uint64
	BeaconHeight          uint64
	TxReqID               common.Hash

	RedeemerExternalAddress string
	ShardHeight             uint64
	shardID                 byte
}

type MatchingRedeemCustodianDetail struct {
	IncAddress    string
	RemoteAddress string
	Amount        uint64
}

func (rq RedeemRequest) GetUniqueRedeemID() string {
	return rq.UniqueRedeemID
}

func (rq *RedeemRequest) SetUniqueRedeemID(uniqueRedeemID string) {
	rq.UniqueRedeemID = uniqueRedeemID
}

func (rq RedeemRequest) GetTokenID() string {
	return rq.TokenID
}

func (rq *RedeemRequest) SetTokenID(tokenID string) {
	rq.TokenID = tokenID
}

func (rq RedeemRequest) GetRedeemerAddress() string {
	return rq.RedeemerAddress
}

func (rq *RedeemRequest) SetRedeemerAddress(redeemerAddress string) {
	rq.RedeemerAddress = redeemerAddress
}

func (rq RedeemRequest) GetRedeemerRemoteAddress() string {
	return rq.RedeemerRemoteAddress
}

func (rq *RedeemRequest) SetRedeemerRemoteAddress(redeemerRemoteAddress string) {
	rq.RedeemerRemoteAddress = redeemerRemoteAddress
}

func (rq RedeemRequest) GetRedeemAmount() uint64 {
	return rq.RedeemAmount
}

func (rq *RedeemRequest) SetRedeemAmount(redeemAmount uint64) {
	rq.RedeemAmount = redeemAmount
}

func (rq RedeemRequest) GetCustodians() []*MatchingRedeemCustodianDetail {
	return rq.Custodians
}

func (rq *RedeemRequest) SetCustodians(custodians []*MatchingRedeemCustodianDetail) {
	rq.Custodians = custodians
}

func (rq RedeemRequest) GetRedeemFee() uint64 {
	return rq.RedeemFee
}

func (rq *RedeemRequest) SetRedeemFee(redeemFee uint64) {
	rq.RedeemFee = redeemFee
}

func (rq RedeemRequest) GetBeaconHeight() uint64 {
	return rq.BeaconHeight
}

func (rq *RedeemRequest) SetBeaconHeight(beaconHeight uint64) {
	rq.BeaconHeight = beaconHeight
}

func (rq RedeemRequest) GetTxReqID() common.Hash {
	return rq.TxReqID
}

func (rq *RedeemRequest) SetTxReqID(txReqID common.Hash) {
	rq.TxReqID = txReqID
}

func (rq RedeemRequest) GetShardHeight() uint64 {
	return rq.ShardHeight
}

func (rq *RedeemRequest) SetShardHeight(shardHeight uint64) {
	rq.ShardHeight = shardHeight
}

func (rq RedeemRequest) ShardID() byte {
	return rq.shardID
}

func (rq *RedeemRequest) SetShardID(shardID byte) {
	rq.shardID = shardID
}

func (rq *RedeemRequest) GetRedeemerExternalAddress() string {
	return rq.RedeemerExternalAddress
}

func (rq *RedeemRequest) SetRedeemerExternalAddress(redeemerAddress string) {
	rq.RedeemerExternalAddress = redeemerAddress
}

func (rq *RedeemRequest) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UniqueRedeemID          string
		TokenID                 string
		RedeemerAddress         string
		RedeemerRemoteAddress   string
		RedeemAmount            uint64
		Custodians              []*MatchingRedeemCustodianDetail
		RedeemFee               uint64
		BeaconHeight            uint64
		TxReqID                 common.Hash
		ShardID                 byte
		ShardHeight             uint64
		RedeemerExternalAddress string
	}{
		UniqueRedeemID:          rq.UniqueRedeemID,
		TokenID:                 rq.TokenID,
		RedeemerAddress:         rq.RedeemerAddress,
		RedeemerRemoteAddress:   rq.RedeemerRemoteAddress,
		RedeemAmount:            rq.RedeemAmount,
		Custodians:              rq.Custodians,
		RedeemFee:               rq.RedeemFee,
		BeaconHeight:            rq.BeaconHeight,
		TxReqID:                 rq.TxReqID,
		ShardID:                 rq.shardID,
		ShardHeight:             rq.ShardHeight,
		RedeemerExternalAddress: rq.RedeemerExternalAddress,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (rq *RedeemRequest) UnmarshalJSON(data []byte) error {
	temp := struct {
		UniqueRedeemID          string
		TokenID                 string
		RedeemerAddress         string
		RedeemerRemoteAddress   string
		RedeemAmount            uint64
		Custodians              []*MatchingRedeemCustodianDetail
		RedeemFee               uint64
		BeaconHeight            uint64
		TxReqID                 common.Hash
		ShardHeight             uint64
		ShardID                 byte
		RedeemerExternalAddress string
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	rq.UniqueRedeemID = temp.UniqueRedeemID
	rq.TokenID = temp.TokenID
	rq.RedeemerAddress = temp.RedeemerAddress
	rq.RedeemerRemoteAddress = temp.RedeemerRemoteAddress
	rq.RedeemAmount = temp.RedeemAmount
	rq.Custodians = temp.Custodians
	rq.RedeemFee = temp.RedeemFee
	rq.BeaconHeight = temp.BeaconHeight
	rq.TxReqID = temp.TxReqID
	rq.ShardHeight = temp.ShardHeight
	rq.shardID = temp.ShardID
	rq.RedeemerExternalAddress = temp.RedeemerExternalAddress
	return nil
}

func (mc MatchingRedeemCustodianDetail) GetIncognitoAddress() string {
	return mc.IncAddress
}

func (mc *MatchingRedeemCustodianDetail) SetIncognitoAddress(incognitoAddress string) {
	mc.IncAddress = incognitoAddress
}

func (mc MatchingRedeemCustodianDetail) GetRemoteAddress() string {
	return mc.RemoteAddress
}

func (mc *MatchingRedeemCustodianDetail) SetRemoteAddress(remoteAddress string) {
	mc.RemoteAddress = remoteAddress
}

func (mc MatchingRedeemCustodianDetail) GetAmount() uint64 {
	return mc.Amount
}

func (mc *MatchingRedeemCustodianDetail) SetAmount(amount uint64) {
	mc.Amount = amount
}

func (mc MatchingRedeemCustodianDetail) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		IncAddress    string
		RemoteAddress string
		Amount        uint64
	}{
		IncAddress:    mc.IncAddress,
		RemoteAddress: mc.RemoteAddress,
		Amount:        mc.Amount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (mc *MatchingRedeemCustodianDetail) UnmarshalJSON(data []byte) error {
	temp := struct {
		IncAddress    string
		RemoteAddress string
		Amount        uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	mc.IncAddress = temp.IncAddress
	mc.RemoteAddress = temp.RemoteAddress
	mc.Amount = temp.Amount
	return nil
}

func NewMatchingRedeemCustodianDetailWithValue(
	incAddress string,
	remoteAddress string,
	amount uint64) *MatchingRedeemCustodianDetail {

	return &MatchingRedeemCustodianDetail{
		IncAddress:    incAddress,
		RemoteAddress: remoteAddress,
		Amount:        amount,
	}
}

func NewRedeemRequest() *RedeemRequest {
	return &RedeemRequest{}
}

func NewRedeemRequestWithValue(
	uniqueRedeemID string,
	tokenID string,
	redeemerAddress string,
	redeemerRemoteAddress string,
	redeemAmount uint64,
	custodians []*MatchingRedeemCustodianDetail,
	redeemFee uint64,
	beaconHeight uint64,
	txReqID common.Hash,
	sharID byte,
	shardHeight uint64,
	redeemerExternalAddress string) *RedeemRequest {

	return &RedeemRequest{
		UniqueRedeemID:          uniqueRedeemID,
		TokenID:                 tokenID,
		RedeemerAddress:         redeemerAddress,
		RedeemerRemoteAddress:   redeemerRemoteAddress,
		RedeemAmount:            redeemAmount,
		Custodians:              custodians,
		RedeemFee:               redeemFee,
		BeaconHeight:            beaconHeight,
		TxReqID:                 txReqID,
		shardID:                 sharID,
		ShardHeight:             shardHeight,
		RedeemerExternalAddress: redeemerExternalAddress,
	}
}

type RedeemRequestObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                  int
	waitingRedeemRequestHash common.Hash
	waitingRedeemRequest     *RedeemRequest
	objectType               int
	deleted                  bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newRedeemRequestObject(db *StateDB, hash common.Hash) *RedeemRequestObject {
	return &RedeemRequestObject{
		version:                  defaultVersion,
		db:                       db,
		waitingRedeemRequestHash: hash,
		waitingRedeemRequest:     NewRedeemRequest(),
		objectType:               WaitingRedeemRequestObjectType,
		deleted:                  false,
	}
}

func newRedeemRequestObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*RedeemRequestObject, error) {
	var redeemRequest = NewRedeemRequest()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, redeemRequest)
		if err != nil {
			return nil, err
		}
	} else {
		redeemRequest, ok = data.(*RedeemRequest)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPortalWaitingRedeemRequestType, reflect.TypeOf(data))
		}
	}
	return &RedeemRequestObject{
		version:                  defaultVersion,
		waitingRedeemRequestHash: key,
		waitingRedeemRequest:     redeemRequest,
		db:                       db,
		objectType:               WaitingRedeemRequestObjectType,
		deleted:                  false,
	}, nil
}

func GenerateWaitingRedeemRequestObjectKey(redeemID string) common.Hash {
	prefixHash := GetWaitingRedeemRequestPrefix()
	valueHash := common.HashH([]byte(redeemID))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func GenerateMatchedRedeemRequestObjectKey(redeemID string) common.Hash {
	prefixHash := GetMatchedRedeemRequestPrefix()
	valueHash := common.HashH([]byte(redeemID))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t RedeemRequestObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *RedeemRequestObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t RedeemRequestObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *RedeemRequestObject) SetValue(data interface{}) error {
	redeemRequest, ok := data.(*RedeemRequest)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPortalWaitingRedeemRequestType, reflect.TypeOf(data))
	}
	t.waitingRedeemRequest = redeemRequest
	return nil
}

func (t RedeemRequestObject) GetValue() interface{} {
	return t.waitingRedeemRequest
}

func (t RedeemRequestObject) GetValueBytes() []byte {
	redeemRequest, ok := t.GetValue().(*RedeemRequest)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(redeemRequest)
	if err != nil {
		panic("failed to marshal redeem request")
	}
	return value
}

func (t RedeemRequestObject) GetHash() common.Hash {
	return t.waitingRedeemRequestHash
}

func (t RedeemRequestObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *RedeemRequestObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *RedeemRequestObject) Reset() bool {
	t.waitingRedeemRequest = NewRedeemRequest()
	return true
}

func (t RedeemRequestObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t RedeemRequestObject) IsEmpty() bool {
	temp := NewRedeemRequest()
	return reflect.DeepEqual(temp, t.waitingRedeemRequest) || t.waitingRedeemRequest == nil
}
