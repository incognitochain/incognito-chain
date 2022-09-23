package statedb

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
)

type RebuildInfo struct {
	lastRootHash  common.Hash
	pivotRootHash common.Hash
	lastFFIndex   int64
	pivotFFIndex  int64
}

func NewRebuildInfo(rebuildRoot, pivotRoot common.Hash, rebuildFFIndex, pivotFFIndex int64) *RebuildInfo {

	return &RebuildInfo{
		rebuildRoot,
		pivotRoot,
		rebuildFFIndex,
		pivotFFIndex,
	}
}

func NewEmptyRebuildInfo() *RebuildInfo {
	return &RebuildInfo{
		common.EmptyRoot,
		common.EmptyRoot,
		0,
		-1,
	}
}

func (r RebuildInfo) GetLastRootHash() common.Hash {
	return r.lastRootHash
}

func (r RebuildInfo) String() string {
	return fmt.Sprintf("rebuild:%v-%v pivot:%v-%v", r.lastRootHash.String(), r.lastFFIndex, r.pivotRootHash.String(), r.pivotFFIndex)
}

func (r RebuildInfo) IsEmpty() bool {
	return r.lastRootHash.IsEqual(&common.EmptyRoot)
}

func (r RebuildInfo) Copy() *RebuildInfo {
	return &RebuildInfo{
		r.lastRootHash,
		r.pivotRootHash,
		r.lastFFIndex,
		r.pivotFFIndex,
	}
}

func (r *RebuildInfo) MarshalJSON() ([]byte, error) {
	res := make([]byte, 32+32+8+8)

	copy(res[:32], r.lastRootHash.Bytes())
	copy(res[32:64], r.pivotRootHash.Bytes())
	binary.LittleEndian.PutUint64(res[64:], uint64(r.lastFFIndex))
	binary.LittleEndian.PutUint64(res[72:], uint64(r.pivotFFIndex))
	sEnc := base64.StdEncoding.EncodeToString(res)
	return []byte("\"" + sEnc + "\""), nil
}

func (r *RebuildInfo) UnmarshalJSON(byteArr []byte) (err error) {

	data, err := base64.StdEncoding.DecodeString(string(byteArr))
	if err != nil {
		return err
	}

	err = r.lastRootHash.SetBytes(data[:32])
	if err != nil {
		return err
	}
	err = r.pivotRootHash.SetBytes(data[32:64])
	if err != nil {
		return err
	}
	err = binary.Read(bytes.NewBuffer(data[64:]), binary.LittleEndian, &r.lastFFIndex)
	if err != nil {
		return err
	}

	err = binary.Read(bytes.NewBuffer(data[72:]), binary.LittleEndian, &r.pivotFFIndex)
	if err != nil {
		return err
	}
	return nil
}
