package statedb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
)

type RebuildInfo struct {
	mode            string
	rebuildRootHash common.Hash
	pivotRootHash   common.Hash
	rebuildFFIndex  int64
	pivotFFIndex    int64
}

func NewRebuildInfo(mode string, rebuildRoot, pivotRoot common.Hash, rebuildFFIndex, pivotFFIndex int64) *RebuildInfo {

	return &RebuildInfo{
		mode,
		rebuildRoot,
		pivotRoot,
		rebuildFFIndex,
		pivotFFIndex,
	}
}

func NewEmptyRebuildInfo(mode string) *RebuildInfo {
	return &RebuildInfo{
		"",
		common.EmptyRoot,
		common.EmptyRoot,
		0,
		0,
	}
}

func (r RebuildInfo) String() string {
	return fmt.Sprintf("%v %v %v %v %v", r.mode, r.rebuildRootHash.String(), r.pivotRootHash.String(), r.rebuildFFIndex, r.pivotFFIndex)
}

func (r RebuildInfo) IsEmpty() bool {
	return r.rebuildRootHash.IsEqual(&common.EmptyRoot)
}

func (r RebuildInfo) Copy() *RebuildInfo {
	return &RebuildInfo{
		r.mode,
		r.rebuildRootHash,
		r.pivotRootHash,
		r.rebuildFFIndex,
		r.pivotFFIndex,
	}
}

func (r RebuildInfo) ToBytes() []byte {
	res := make([]byte, 1+32+32+8+8)
	switch r.mode {
	case common.STATEDB_ARCHIVE_MODE:
		res[0] = 0
	case common.STATEDB_BATCH_COMMIT_MODE:
		res[0] = 1
	case common.STATEDB_LITE_MODE:
		res[0] = 2
	}

	copy(res[1:33], r.rebuildRootHash.Bytes())
	copy(res[33:65], r.pivotRootHash.Bytes())
	binary.LittleEndian.PutUint64(res[65:], uint64(r.rebuildFFIndex))
	binary.LittleEndian.PutUint64(res[73:], uint64(r.pivotFFIndex))
	return res
}

func (r *RebuildInfo) FromBytes(data []byte) (err error) {
	if len(data) == 32 || len(data) == 0 { //legacy shard root hash (only root hash of archive mode)
		r.mode = common.STATEDB_ARCHIVE_MODE
		err = r.rebuildRootHash.SetBytes(data)
		if err != nil {
			return err
		}
		return nil
	}

	switch data[0] {
	case 0:
		r.mode = common.STATEDB_ARCHIVE_MODE
	case 1:
		r.mode = common.STATEDB_BATCH_COMMIT_MODE
	case 2:
		r.mode = common.STATEDB_LITE_MODE
	}

	err = r.rebuildRootHash.SetBytes(data[1:33])
	if err != nil {
		return err
	}
	err = r.pivotRootHash.SetBytes(data[33:65])
	if err != nil {
		return err
	}
	err = binary.Read(bytes.NewBuffer(data[65:]), binary.LittleEndian, &r.rebuildFFIndex)
	if err != nil {
		return err
	}

	err = binary.Read(bytes.NewBuffer(data[73:]), binary.LittleEndian, &r.pivotFFIndex)
	if err != nil {
		return err
	}
	return nil
}
