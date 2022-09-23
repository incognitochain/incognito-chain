package statedb

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
)

type RebuildInfo struct {
	rebuildRootHash common.Hash
	pivotRootHash   common.Hash
	rebuildFFIndex  int64
	pivotFFIndex    int64
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

func (r RebuildInfo) GetRootHash() common.Hash {
	return r.rebuildRootHash
}

func (r RebuildInfo) String() string {
	return fmt.Sprintf("rebuild:%v-%v pivot:%v-%v", r.rebuildRootHash.String(), r.rebuildFFIndex, r.pivotRootHash.String(), r.pivotFFIndex)
}

func (r RebuildInfo) IsEmpty() bool {
	return r.rebuildRootHash.IsEqual(&common.EmptyRoot)
}

func (r RebuildInfo) Copy() *RebuildInfo {
	return &RebuildInfo{
		r.rebuildRootHash,
		r.pivotRootHash,
		r.rebuildFFIndex,
		r.pivotFFIndex,
	}
}

func (r *RebuildInfo) MarshalJSON() ([]byte, error) {
	res := make([]byte, 32+32+8+8)

	copy(res[:32], r.rebuildRootHash.Bytes())
	copy(res[32:64], r.pivotRootHash.Bytes())
	binary.LittleEndian.PutUint64(res[64:], uint64(r.rebuildFFIndex))
	binary.LittleEndian.PutUint64(res[72:], uint64(r.pivotFFIndex))
	sEnc := base64.StdEncoding.EncodeToString(res)
	return []byte("\"" + sEnc + "\""), nil
}

func (r *RebuildInfo) UnmarshalJSON(byteArr []byte) (err error) {

	data, err := base64.StdEncoding.DecodeString(string(byteArr))
	if err != nil {
		return err
	}

	err = r.rebuildRootHash.SetBytes(data[:32])
	if err != nil {
		return err
	}
	err = r.pivotRootHash.SetBytes(data[32:64])
	if err != nil {
		return err
	}
	err = binary.Read(bytes.NewBuffer(data[64:]), binary.LittleEndian, &r.rebuildFFIndex)
	if err != nil {
		return err
	}

	err = binary.Read(bytes.NewBuffer(data[72:]), binary.LittleEndian, &r.pivotFFIndex)
	if err != nil {
		return err
	}
	return nil
}
