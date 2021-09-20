//nolint:forcetypeassert // This file uses uncheck type assertions
package tx_generic //nolint:revive

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

type ValidationEnv struct {
	isPrivacy    bool
	isConfimed   bool
	hasCA        bool
	txType       string
	txAction     int
	shardID      int
	shardHeight  uint64
	beaconHeight uint64
	confimedTime int64
	version      int
	sigPubKey    []byte
	dataFromDB   [][]byte
	tokenID      common.Hash
}

func NewValEnv(
	isPrivacyTx bool,
	txType string,
	sHeight uint64,
	bHeight uint64,
	sID int,
	confimedTime uint64,
	ver int,
	act int,
) *ValidationEnv {
	return &ValidationEnv{}
}

func DefaultValEnv() *ValidationEnv {
	return &ValidationEnv{
		isConfimed:   false,
		shardHeight:  0,
		beaconHeight: 0,
		confimedTime: 0,
		txAction:     common.TxActTranfer,
		version:      0,
	}
}

func WithPrivacy(vE metadata.ValidationEnviroment) *ValidationEnv {
	vEnv := vE.(*ValidationEnv)
	vEnv.isPrivacy = true
	return vEnv
}

func WithCA(vE metadata.ValidationEnviroment, hasCA bool) *ValidationEnv {
	vEnv := vE.(*ValidationEnv)
	vEnv.hasCA = hasCA
	return vEnv
}

func WithNoPrivacy(vE metadata.ValidationEnviroment) *ValidationEnv {
	vEnv := vE.(*ValidationEnv)
	vEnv.isPrivacy = false
	return vEnv
}

func WithShardID(vE metadata.ValidationEnviroment, sID int) *ValidationEnv {
	vEnv := vE.(*ValidationEnv)
	vEnv.shardID = sID
	return vEnv
}

func WithShardHeight(vE metadata.ValidationEnviroment, sHeight uint64) *ValidationEnv {
	vEnv := vE.(*ValidationEnv)
	vEnv.shardHeight = sHeight
	return vEnv
}

func WithBeaconHeight(vE metadata.ValidationEnviroment, bcHeight uint64) *ValidationEnv {
	vEnv := vE.(*ValidationEnv)
	vEnv.beaconHeight = bcHeight
	return vEnv
}

func WithConfirmedTime(vE metadata.ValidationEnviroment, confirmedTime int64) *ValidationEnv {
	vEnv := vE.(*ValidationEnv)
	vEnv.confimedTime = confirmedTime
	return vEnv
}

func WithType(vE metadata.ValidationEnviroment, t string) *ValidationEnv {
	vEnv := vE.(*ValidationEnv)
	vEnv.txType = t
	return vEnv
}

func WithVersion(vE metadata.ValidationEnviroment, ver int8) *ValidationEnv {
	vEnv := vE.(*ValidationEnv)
	vEnv.version = int(ver)
	return vEnv
}

func WithAct(vE metadata.ValidationEnviroment, act int) *ValidationEnv {
	vEnv := vE.(*ValidationEnv)
	vEnv.txAction = act
	return vEnv
}

func WithDBData(vE metadata.ValidationEnviroment, data [][]byte) *ValidationEnv {
	vEnv := vE.(*ValidationEnv)
	vEnv.dataFromDB = data
	return vEnv
}

func WithTokenID(vE metadata.ValidationEnviroment, tokenID common.Hash) *ValidationEnv {
	vEnv := vE.(*ValidationEnv)
	vEnv.tokenID = tokenID
	return vEnv
}

func WithSigPubkey(vE metadata.ValidationEnviroment, sigPubkey []byte) *ValidationEnv {
	vEnv := vE.(*ValidationEnv)
	vEnv.sigPubKey = sigPubkey
	return vEnv
}

func (vE *ValidationEnv) IsPrivacy() bool {
	return vE.isPrivacy
}
func (vE *ValidationEnv) HasCA() bool {
	return vE.hasCA
}
func (vE *ValidationEnv) IsConfimed() bool {
	return vE.isConfimed
}
func (vE *ValidationEnv) TxType() string {
	return vE.txType
}
func (vE *ValidationEnv) ShardID() int {
	return vE.shardID
}
func (vE *ValidationEnv) ShardHeight() uint64 {
	return vE.shardHeight
}
func (vE *ValidationEnv) BeaconHeight() uint64 {
	return vE.beaconHeight
}
func (vE *ValidationEnv) ConfirmedTime() int64 {
	return vE.confimedTime
}
func (vE *ValidationEnv) Version() int {
	return vE.version
}

func (vE *ValidationEnv) TxAction() int {
	return vE.txAction
}

func (vE *ValidationEnv) SigPubKey() []byte {
	return vE.sigPubKey
}

func (vE *ValidationEnv) TokenID() common.Hash {
	return vE.tokenID
}

func (vE *ValidationEnv) DBData() [][]byte {
	return vE.dataFromDB
}

func (vE *ValidationEnv) Clone() *ValidationEnv {
	return &ValidationEnv{
		isConfimed:   vE.IsConfimed(),
		shardHeight:  vE.ShardHeight(),
		beaconHeight: vE.BeaconHeight(),
		confimedTime: vE.ConfirmedTime(),
		shardID:      vE.ShardID(),
		version:      vE.Version(),
		isPrivacy:    vE.IsPrivacy(),
		txAction:     vE.TxAction(),
		txType:       vE.TxType(),
		sigPubKey:    vE.SigPubKey(),
		hasCA:        vE.HasCA(),
		tokenID:      vE.TokenID(),
		dataFromDB:   vE.DBData(),
	}
}
