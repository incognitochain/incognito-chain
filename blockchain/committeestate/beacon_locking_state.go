package committeestate

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/instruction"
)

type LockingInfo struct {
	PublicKey     string
	UnlockAtEpoch uint64
	Reason        byte
}

func NewBeaconLockingInfo() *LockingInfo {
	return &LockingInfo{}
}

func (b *BeaconCommitteeStateV4) LockNewCandidate(pk string, unlockEpoch uint64, reason byte) {
	b.beaconLocking = append(b.beaconLocking, &LockingInfo{PublicKey: pk, UnlockAtEpoch: unlockEpoch, Reason: reason})
}

func (b *BeaconCommitteeStateV4) UnlockCandidateAtIndex(id int) {
	b.beaconLocking[id] = b.beaconLocking[len(b.beaconLocking)-1] // Copy last element to index i.
	b.beaconLocking = b.beaconLocking[:len(b.beaconLocking)-1]    // Truncate slice.
}

func (b *BeaconCommitteeStateV4) GetReturnStakingInstruction(bcStateDB *statedb.StateDB, epoch uint64) *instruction.ReturnBeaconStakeInstruction {
	returnPKs := []string{}
	returnTxIDs := [][]string{}
	returnReason := []byte{}
	for _, lockedCandidate := range b.beaconLocking {
		fmt.Printf("Locking info at epoch %v of candidate %v: locked until %v\n", epoch, lockedCandidate.PublicKey[len(lockedCandidate.PublicKey)-5:], lockedCandidate.UnlockAtEpoch)
		if lockedCandidate.UnlockAtEpoch == epoch {
			if info, has, err := statedb.GetBeaconStakerInfo(bcStateDB, lockedCandidate.PublicKey); (err == nil) && (has) {
				txIDs := info.TxStakingIDs()
				txIDsString := []string{}
				for _, txID := range txIDs {
					txIDsString = append(txIDsString, txID.String())
				}
				returnPKs = append(returnPKs, lockedCandidate.PublicKey)
				returnTxIDs = append(returnTxIDs, txIDsString)
				returnReason = append(returnReason, lockedCandidate.Reason)
			}
		}
	}
	if len(returnPKs) == 0 {
		return nil
	}
	return instruction.NewReturnBeaconStakeInsWithValue(returnPKs, returnTxIDs, returnReason)
}
