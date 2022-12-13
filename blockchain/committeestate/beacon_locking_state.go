package committeestate

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/instruction"
)

type BeaconLockingState struct {
	isChange bool
	Data     map[string]LockingInfo
}

func NewBeaconLockingState() *BeaconLockingState {
	return &BeaconLockingState{
		isChange: true,
		Data:     map[string]LockingInfo{},
	}
}

func (b *BeaconLockingState) Backup() []byte {
	if !b.isChange {
		return nil
	}
	keys := []string{}
	var values []struct {
		U uint64
		R byte
	}
	for k := range b.Data {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	values = make([]struct {
		U uint64
		R byte
	}, len(keys))
	for idx, k := range keys {
		values[idx].R = b.Data[k].Reason
		values[idx].U = b.Data[k].UnlockAtEpoch
	}
	dBK := struct {
		K []string
		V []struct {
			U uint64
			R byte
		}
	}{
		K: keys,
		V: values,
	}

	nextEpochInfo, err := json.Marshal(dBK)
	if err != nil {
		panic(err)
		// return nil
	}
	return nextEpochInfo
}

func (b *BeaconLockingState) Restore(data []byte) error {
	dBK := struct {
		K []string
		V []struct {
			U uint64
			R byte
		}
	}{}
	err := json.Unmarshal(data, &dBK)
	if err != nil {
		panic(err)
	}
	b = NewBeaconLockingState()
	b.isChange = false
	for idx, k := range dBK.K {
		b.Data[k] = LockingInfo{
			Reason:        dBK.V[idx].R,
			UnlockAtEpoch: dBK.V[idx].U,
		}
	}
	return nil
}

func (b *BeaconLockingState) LockNewCandidate(pk string, unlockEpoch uint64, reason byte) {
	b.Data[pk] = LockingInfo{UnlockAtEpoch: unlockEpoch, Reason: reason}
	b.isChange = true
}

func (b *BeaconLockingState) UnlockCandidate(pk string) {
	delete(b.Data, pk)
	b.isChange = true
}

func (b *BeaconLockingState) GetReturnPK(epoch uint64) map[string]LockingInfo {
	res := map[string]LockingInfo{}
	for candidate, lockedCandidate := range b.Data {
		fmt.Printf("Locking info at epoch %v of candidate %v: locked until %v\n", epoch, candidate[len(candidate)-5:], lockedCandidate.UnlockAtEpoch)
		if lockedCandidate.UnlockAtEpoch <= epoch {
			res[candidate] = lockedCandidate
		}
	}
	if len(res) == 0 {
		return nil
	}
	return res
}
func (b *BeaconLockingState) GetAllLockingPK() []string {
	res := []string{}
	for candidate, _ := range b.Data {
		res = append(res, candidate)
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

type LockingInfo struct {
	// PublicKey     string
	UnlockAtEpoch uint64
	Reason        byte
}

func NewBeaconLockingInfo() *LockingInfo {
	return &LockingInfo{}
}

func (b *BeaconCommitteeStateV4) LockNewCandidate(pk string, unlockEpoch uint64, reason byte) {
	b.bLockingState.LockNewCandidate(pk, unlockEpoch, reason)
}

func (b *BeaconCommitteeStateV4) GetReturnStakingInstruction(bcStateDB *statedb.StateDB, epoch uint64) *instruction.ReturnBeaconStakeInstruction {
	returnPKs := []string{}
	returnAmounts := []uint64{}
	returnReason := []byte{}
	unlockInfo := b.bLockingState.GetReturnPK(epoch)
	if unlockInfo == nil {
		return nil
	}
	for unlockPK, unlockInfo := range unlockInfo {
		if info, has, err := statedb.GetBeaconStakerInfo(bcStateDB, unlockPK); (err == nil) && (has) {
			returnPKs = append(returnPKs, unlockPK)
			returnAmounts = append(returnAmounts, info.StakingAmount())
			returnReason = append(returnReason, unlockInfo.Reason)
			b.bLockingState.UnlockCandidate(unlockPK)
		}
	}
	b.committeeChange.AddRemovedBeaconStakers(returnPKs)
	return instruction.NewReturnBeaconStakeInsWithValue(returnPKs, returnReason, returnAmounts)
}
