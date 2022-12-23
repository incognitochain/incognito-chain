package committeestate

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/pkg/errors"
)

type BeaconDelegatorInfo struct {
	StakingAmount            uint64
	CurrentDelegators        int
	CurrentDelegatorsDetails map[string]interface{}
	// WaitingDelegators        int
	// WaitingDelegatorsDetails map[string]interface{}
	locker *sync.RWMutex
}

func NewBeaconDelegatorInfo() *BeaconDelegatorInfo {
	return &BeaconDelegatorInfo{
		StakingAmount:            0,
		CurrentDelegators:        0,
		CurrentDelegatorsDetails: map[string]interface{}{},
		locker:                   &sync.RWMutex{},
	}
}

func (b *BeaconDelegatorInfo) AddStakingAmount(amount uint64) {
	b.StakingAmount += amount
}

func (b *BeaconDelegatorInfo) Add(newDelegator string) error {
	if _, exist := b.CurrentDelegatorsDetails[newDelegator]; exist {
		return nil
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

func (b *BeaconDelegatorInfo) GetStakingAmount() uint64 {
	b.locker.RLock()
	defer b.locker.RUnlock()
	return b.StakingAmount
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

func (b *BeaconDelegateState) AddStakingAmount(newCandidate string, stakingAmount uint64) {
	if _, ok := b.DelegateInfo[newCandidate]; !ok {
		Logger.log.Infof("Added staking amount to unexist candidate %v %v", newCandidate, stakingAmount)
		b.DelegateInfo[newCandidate] = NewBeaconDelegatorInfo()
		b.DelegateInfo[newCandidate].AddStakingAmount(stakingAmount)
	}
	Logger.log.Infof("Added staking amount to candidate %v %v", newCandidate[len(newCandidate)-5:], stakingAmount)
	b.DelegateInfo[newCandidate].AddStakingAmount(stakingAmount)
}

func (b *BeaconDelegateState) AddBeaconCandidate(newCandidate string, stakingAmount uint64) {
	b.DelegateInfo[newCandidate] = NewBeaconDelegatorInfo()
	b.DelegateInfo[newCandidate].StakingAmount = stakingAmount
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

func (b *BeaconDelegateState) GetBeaconCandidatePower(bPK string) uint64 {
	res := uint64(0)
	if info, ok := b.DelegateInfo[bPK]; ok {
		res = info.StakingAmount / config.Param().StakingAmountShard
		res += uint64(info.CurrentDelegators)
	}
	return res
}

func (b *BeaconDelegateState) Backup() []byte {
	keys := []string{}
	var values []struct {
		O string
		N string
	}
	for k := range b.NextEpochDelegate {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	values = make([]struct {
		O string
		N string
	}, len(keys))
	for idx, k := range keys {
		values[idx].O = b.NextEpochDelegate[k].Old
		values[idx].N = b.NextEpochDelegate[k].New
	}
	dBK := struct {
		K []string
		V []struct {
			O string
			N string
		}
	}{
		K: keys,
		V: values,
	}
	nextEpochInfo, err := json.Marshal(dBK)
	if err != nil {
		panic(err)
	}
	return nextEpochInfo
}

func (b *BeaconDelegateState) Restore(bcState BeaconCommitteeState, stateDB *statedb.StateDB, data []byte) error {
	dBK := struct {
		K []string
		V []struct {
			O string
			N string
		}
	}{}
	err := json.Unmarshal(data, &dBK)
	if err != nil {
		panic(err)
	}
	if b.NextEpochDelegate == nil {
		b.NextEpochDelegate = map[string]struct {
			Old string
			New string
		}{}
		for idx, k := range dBK.K {
			b.NextEpochDelegate[k] = struct {
				Old string
				New string
			}{
				Old: dBK.V[idx].O,
				New: dBK.V[idx].N,
			}
		}
	}
	bc := bcState.GetBeaconCommittee()
	if bs := bcState.GetBeaconSubstitute(); len(bs) != 0 {
		bc = append(bc, bs...)
	}
	if bw := bcState.GetBeaconWaiting(); len(bw) != 0 {
		bc = append(bc, bw...)
	}
	for _, k := range bc {
		pkString, _ := k.ToBase58()
		if bStakerInfo, exist, err := statedb.GetBeaconStakerInfo(stateDB, pkString); (exist) && (err == nil) {
			b.AddBeaconCandidate(pkString, bStakerInfo.StakingAmount())
		} else {
			panic(err)
		}
	}
	staker := bcState.GetAllCandidateSubstituteCommittee()
	for _, stakerPKStr := range staker {
		if stakerInfo, exist, err := statedb.GetShardStakerInfo(stateDB, stakerPKStr); (exist) && (err == nil) {
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
