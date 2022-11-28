package committeestate

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/pkg/errors"
)

type BeaconDelegatorInfo struct {
	CurrentDelegators        int
	CurrentDelegatorsDetails map[string]interface{}
	// WaitingDelegators        int
	// WaitingDelegatorsDetails map[string]interface{}
	locker *sync.RWMutex
}

func NewBeaconDelegatorInfo() *BeaconDelegatorInfo {
	return &BeaconDelegatorInfo{
		CurrentDelegators:        0,
		CurrentDelegatorsDetails: map[string]interface{}{},
		locker:                   &sync.RWMutex{},
	}
}

func (b *BeaconDelegatorInfo) Add(newDelegator string) error {
	if _, exist := b.CurrentDelegatorsDetails[newDelegator]; exist {
		return errors.Errorf("This delegator %v already added", newDelegator)
	}
	b.CurrentDelegators++
	b.CurrentDelegatorsDetails[newDelegator] = nil
	return nil
}

func (b *BeaconDelegatorInfo) Remove(delegator string) {
	if _, exist := b.CurrentDelegatorsDetails[delegator]; !exist {
		return
	}
	b.CurrentDelegators--
	delete(b.CurrentDelegatorsDetails, delegator)
}

func (b *BeaconDelegatorInfo) GetCurrentDelegators() int {
	b.locker.RLock()
	defer b.locker.RUnlock()
	return b.CurrentDelegators
}

func (b *BeaconDelegatorInfo) GetCurrentDelegatorsDetails() map[string]interface{} {
	res := map[string]interface{}{}
	b.locker.RLock()
	defer b.locker.RUnlock()
	for k, v := range b.CurrentDelegatorsDetails {
		res[k] = v
	}
	return res
}

func (b *BeaconDelegatorInfo) GetCurrentDelegatorsList() []string {
	keys := []string{}
	for k := range b.CurrentDelegatorsDetails {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

type BeaconDelegateState struct {
	DelegateInfo      map[string]*BeaconDelegatorInfo
	NextEpochDelegate map[string]struct {
		Old string
		New string
	}
	locker *sync.RWMutex
}

func InitBeaconDelegateState(bCState BeaconCommitteeState) (*BeaconDelegateState, error) {
	res := &BeaconDelegateState{
		DelegateInfo: map[string]*BeaconDelegatorInfo{},
		NextEpochDelegate: map[string]struct {
			Old string
			New string
		}{},
		locker: &sync.RWMutex{},
	}
	bc := bCState.GetBeaconCommittee()
	if bs := bCState.GetBeaconSubstitute(); len(bs) != 0 {
		bc = append(bc, bs...)
	}
	bcStrs, err := incognitokey.CommitteeKeyListToString(bc)
	if err != nil {
		return nil, err
	}
	for _, bcStr := range bcStrs {
		res.DelegateInfo[bcStr] = NewBeaconDelegatorInfo()
	}
	return res, nil
}

func (b *BeaconDelegateState) Clone() *BeaconDelegateState {
	res := &BeaconDelegateState{
		DelegateInfo: map[string]*BeaconDelegatorInfo{},
		NextEpochDelegate: map[string]struct {
			Old string
			New string
		}{},
		locker: &sync.RWMutex{},
	}
	for k, v := range b.DelegateInfo {
		res.DelegateInfo[k] = v
	}
	for k, v := range b.NextEpochDelegate {
		res.NextEpochDelegate[k] = v
	}
	return res
}

func (b *BeaconDelegateState) AcceptNextEpochChange() error {
	for delegator, delegateChange := range b.NextEpochDelegate {
		if bcInfo, ok := b.DelegateInfo[delegateChange.New]; ok {
			err := bcInfo.Add(delegator)
			if err != nil {
				return err
			}
		}
		if bcInfo, ok := b.DelegateInfo[delegateChange.Old]; ok {
			bcInfo.Remove(delegator)
		}
	}
	b.NextEpochDelegate = map[string]struct {
		Old string
		New string
	}{}
	return nil
}

func (b *BeaconDelegateState) AddReDelegate(delegator, oldDelegate, newDelegate string) {
	b.NextEpochDelegate[delegator] = struct {
		Old string
		New string
	}{
		Old: oldDelegate,
		New: newDelegate,
	}
}

func (b *BeaconDelegateState) GetDelegateInfo(beaconPK string) (BeaconDelegatorInfo, error) {
	if dInfo, ok := b.DelegateInfo[beaconPK]; ok {
		return *dInfo, nil
	}
	return BeaconDelegatorInfo{}, errors.Errorf("Can not found PK %v in delegate state", beaconPK)
}

func (b *BeaconDelegateState) GetDelegateState() map[string]BeaconDelegatorInfo {

	res := map[string]BeaconDelegatorInfo{}
	for k, v := range b.DelegateInfo {
		res[k] = *v
	}
	return res
}

func (b *BeaconDelegateState) Hash() common.Hash {

	res := ""
	keys := []string{}
	for k, _ := range b.DelegateInfo {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		res += k
		res += "-"
		dInfo := b.DelegateInfo[k]
		res += fmt.Sprintf("%v-", dInfo.CurrentDelegators)
		for _, delegator := range dInfo.GetCurrentDelegatorsList() {
			res += delegator
			res += "-"
		}
	}
	keys = []string{}
	for k, _ := range b.NextEpochDelegate {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		res += k
		res += "-"
		res += b.NextEpochDelegate[k].Old
		res += "-"
		res += b.NextEpochDelegate[k].New
		res += "-"
	}
	return common.HashH([]byte(res))
}

func (b *BeaconDelegateState) GetBeaconCandidatePower(bPK string) uint {
	res := uint(2) //35000/1750
	if info, ok := b.DelegateInfo[bPK]; ok {
		res += uint(info.CurrentDelegators)
	}
	return res
}

func (b *BeaconDelegateState) Backup(db incdb.Database) error {
	nextEpochInfo, err := json.Marshal(b.NextEpochDelegate)
	if err != nil {
		return err
	}
	return rawdbv2.StoreBeaconNextDelegate(db, nextEpochInfo)
}

func (b *BeaconDelegateState) Restore(bcState BeaconCommitteeState, stateDB *statedb.StateDB, db incdb.Database) error {
	data, err := rawdbv2.GetBeaconNextDelegte(db)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, b.NextEpochDelegate)
	if err != nil {
		return err
	}
	if b.NextEpochDelegate == nil {
		b.NextEpochDelegate = map[string]struct {
			Old string
			New string
		}{}
	}
	staker := bcState.GetAllCandidateSubstituteCommittee()
	for _, stakerPKStr := range staker {
		if stakerInfo, exist, err := statedb.GetStakerInfo(stateDB, stakerPKStr); (exist) && (err == nil) {
			if stakerInfo.HasCredit() {
				curD := stakerInfo.Delegate()
				if info, ok := b.NextEpochDelegate[stakerPKStr]; ok {
					curD = info.Old
				}
				if len(curD) > 0 {
					b.DelegateInfo[curD].Add(stakerPKStr)
				}
			}
		}
	}
	return nil
}
