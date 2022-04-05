package statedb

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
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
		mode,
		common.EmptyRoot,
		common.EmptyRoot,
		0,
		-1,
	}
}

func (r RebuildInfo) GetRootHash() common.Hash {
	return r.rebuildRootHash
}

func (r RebuildInfo) String() string {
	return fmt.Sprintf("mode:%v rebuild:%v-%v pivot:%v-%v", r.mode, r.rebuildRootHash.String(), r.rebuildFFIndex, r.pivotRootHash.String(), r.pivotFFIndex)
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

func (r *RebuildInfo) MarshalJSON() ([]byte, error) {
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
	sEnc := base64.StdEncoding.EncodeToString(res)
	return []byte("\"" + sEnc + "\""), nil
}

func (r *RebuildInfo) UnmarshalJSON(byteArr []byte) (err error) {

	if len(byteArr) == 66 { //legacy shard root hash (only root hash of archive mode)
		r.mode = common.STATEDB_ARCHIVE_MODE
		hash, err := common.Hash{}.NewHashFromStr(string(byteArr[1:65]))
		r.rebuildRootHash = *hash
		return err
	}

	byteArr = byteArr[1 : len(byteArr)-1]
	data, err := base64.StdEncoding.DecodeString(string(byteArr))
	if err != nil {
		return err
	}

	switch data[0] {
	case 0:
		r.mode = common.STATEDB_ARCHIVE_MODE
	case 1:
		r.mode = common.STATEDB_BATCH_COMMIT_MODE
	case 2:
		r.mode = common.STATEDB_LITE_MODE
	default:
		return errors.New("Cannot parse mode")
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
