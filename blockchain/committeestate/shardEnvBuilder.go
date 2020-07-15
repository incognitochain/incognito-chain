package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

//ShardEnvBuilder : Interface for building shard environment
type ShardEnvBuilder interface {
	BuildRecentCommitteesStr(recentCommitteesStr []string) ShardEnvBuilder
	BuildRecentSubtitutesStr(recentSubtitutesStr []string) ShardEnvBuilder
	BuildShardHeight(height uint64) ShardEnvBuilder
	BuildShardBlockHash(blockHash common.Hash) ShardEnvBuilder
	BuildTxs(txs []metadata.Transaction) ShardEnvBuilder
	BuildInstructions(instructions [][]string) ShardEnvBuilder
	BuildBeaconHeight(height uint64) ShardEnvBuilder
	BuildChainParamEpoch(chainParamEpoch uint64) ShardEnvBuilder
	BuildEpochBreakPointSwapNewKey(epochBreakPointSwapNewKey []uint64) ShardEnvBuilder
	BuildShardID(id byte) ShardEnvBuilder
	BuildMaxShardCommitteeSize(maxShardCommitteeSize int) ShardEnvBuilder
	BuildMinShardCommitteeSize(minShardCommitteeSize int) ShardEnvBuilder
	BuildOffset(offset int) ShardEnvBuilder
	BuildSwapOffset(swapOffset int) ShardEnvBuilder
	BuildProducersBlackList(producersBlackList map[string]uint8) ShardEnvBuilder
	BuildStakingTx(stakingTx map[string]string) ShardEnvBuilder
	BuildIsProcessShardBlockInstructionForKeyListV2(isProcessShardBlockInstructionForKeyListV2 bool) ShardEnvBuilder
	Build() ShardCommitteeStateEnvironment
}

//NewShardEnvBuilder :
func NewShardEnvBuilder() ShardEnvBuilder {
	return &shardCommitteeStateEnvironment{}
}

// ShardCommitteeStateEnvironment :
type ShardCommitteeStateEnvironment interface {
	RecentCommitteesStr() []string
	RecentSubtitutesStr() []string
	ShardHeight() uint64
	ShardBlockHash() common.Hash
	Txs() []metadata.Transaction
	Instructions() [][]string
	BeaconHeight() uint64
	ChainParamEpoch() uint64
	EpochBreakPointSwapNewKey() []uint64
	ShardID() byte
	MaxShardCommitteeSize() int
	MinShardCommitteeSize() int
	Offset() int
	SwapOffset() int
	ProducersBlackList() map[string]uint8
	StakingTx() map[string]string
	IsProcessShardBlockInstructionForKeyListV2() bool
}

//shardCommitteeStateEnvironment :
type shardCommitteeStateEnvironment struct {
	recentCommitteesStr                        []string
	recentSubtitutesStr                        []string
	shardHeight                                uint64
	shardBlockHash                             common.Hash
	txs                                        []metadata.Transaction
	instructions                               [][]string
	beaconHeight                               uint64
	chainParamEpoch                            uint64
	epochBreakPointSwapNewKey                  []uint64
	shardID                                    byte
	maxShardCommitteeSize                      int
	minShardCommitteeSize                      int
	offset                                     int
	swapOffset                                 int
	producersBlackList                         map[string]uint8
	stakingTx                                  map[string]string
	isProcessShardBlockInstructionForKeyListV2 bool
}

//BuildRecentCommitteesStr :
func (env *shardCommitteeStateEnvironment) BuildRecentCommitteesStr(recentCommitteesStr []string) ShardEnvBuilder {
	env.recentCommitteesStr = recentCommitteesStr
	return env
}

//BuildRecentSubtitutesStr :
func (env *shardCommitteeStateEnvironment) BuildRecentSubtitutesStr(recentSubtitutesStr []string) ShardEnvBuilder {
	env.recentSubtitutesStr = recentSubtitutesStr
	return env
}

//BuildShardHeight :
func (env *shardCommitteeStateEnvironment) BuildShardHeight(height uint64) ShardEnvBuilder {
	env.shardHeight = height
	return env
}

//BuildShardBlockHash :
func (env *shardCommitteeStateEnvironment) BuildShardBlockHash(blockHash common.Hash) ShardEnvBuilder {
	env.shardBlockHash = blockHash
	return env
}

//BuildTxs :
func (env *shardCommitteeStateEnvironment) BuildTxs(txs []metadata.Transaction) ShardEnvBuilder {
	env.txs = txs
	return env
}

//BuildInstructions :
func (env *shardCommitteeStateEnvironment) BuildInstructions(instructions [][]string) ShardEnvBuilder {
	env.instructions = instructions
	return env
}

//BuildBeaconHeight :
func (env *shardCommitteeStateEnvironment) BuildBeaconHeight(height uint64) ShardEnvBuilder {
	env.beaconHeight = height
	return env
}

//BuildChainParamEpoch :
func (env *shardCommitteeStateEnvironment) BuildChainParamEpoch(chainParamEpoch uint64) ShardEnvBuilder {
	env.chainParamEpoch = chainParamEpoch
	return env
}

//BuildEpochBreakPointSwapNewKey :
func (env *shardCommitteeStateEnvironment) BuildEpochBreakPointSwapNewKey(
	epochBreakPointSwapNewKey []uint64) ShardEnvBuilder {
	env.epochBreakPointSwapNewKey = epochBreakPointSwapNewKey
	return env
}

//BuildShardID :
func (env *shardCommitteeStateEnvironment) BuildShardID(id byte) ShardEnvBuilder {
	env.shardID = id
	return env
}

//BuildMaxShardCommitteeSize :
func (env *shardCommitteeStateEnvironment) BuildMaxShardCommitteeSize(maxShardCommitteeSize int) ShardEnvBuilder {
	env.maxShardCommitteeSize = maxShardCommitteeSize
	return env
}

//BuildMinShardCommitteeSize :
func (env *shardCommitteeStateEnvironment) BuildMinShardCommitteeSize(minShardCommitteeSize int) ShardEnvBuilder {
	env.minShardCommitteeSize = minShardCommitteeSize
	return env
}

//BuildOffset :
func (env *shardCommitteeStateEnvironment) BuildOffset(offset int) ShardEnvBuilder {
	env.offset = offset
	return env
}

//BuildSwapOffset :
func (env *shardCommitteeStateEnvironment) BuildSwapOffset(swapOffset int) ShardEnvBuilder {
	env.swapOffset = swapOffset
	return env
}

//BuildProducersBlackList :
func (env *shardCommitteeStateEnvironment) BuildProducersBlackList(producersBlackList map[string]uint8) ShardEnvBuilder {
	env.producersBlackList = producersBlackList
	return env
}

//BuildStakingTx :
func (env *shardCommitteeStateEnvironment) BuildStakingTx(stakingTx map[string]string) ShardEnvBuilder {
	env.stakingTx = stakingTx
	return env
}

//BuildIsProcessShardBlockInstructionForKeyListV2 :
func (env *shardCommitteeStateEnvironment) BuildIsProcessShardBlockInstructionForKeyListV2(
	isProcessShardBlockInstructionForKeyListV2 bool) ShardEnvBuilder {
	env.isProcessShardBlockInstructionForKeyListV2 = isProcessShardBlockInstructionForKeyListV2
	return env
}

////

//RecentCommitteesStr :
func (env *shardCommitteeStateEnvironment) RecentCommitteesStr() []string {
	return env.recentCommitteesStr
}

//RecentSubtitutesStr :
func (env *shardCommitteeStateEnvironment) RecentSubtitutesStr() []string {
	return env.recentSubtitutesStr
}

//ShardHeight :
func (env *shardCommitteeStateEnvironment) ShardHeight() uint64 {
	return env.shardHeight
}

//ShardBlockHash :
func (env *shardCommitteeStateEnvironment) ShardBlockHash() common.Hash {
	return env.shardBlockHash
}

//Txs :
func (env *shardCommitteeStateEnvironment) Txs() []metadata.Transaction {
	return env.txs
}

//Instructions :
func (env *shardCommitteeStateEnvironment) Instructions() [][]string {
	return env.instructions
}

//BeaconHeight :
func (env *shardCommitteeStateEnvironment) BeaconHeight() uint64 {
	return env.beaconHeight
}

//ChainParamEpoch :
func (env *shardCommitteeStateEnvironment) ChainParamEpoch() uint64 {
	return env.chainParamEpoch
}

//EpochBreakPointSwapNewKey :
func (env *shardCommitteeStateEnvironment) EpochBreakPointSwapNewKey() []uint64 {
	return env.epochBreakPointSwapNewKey
}

//ShardID :
func (env *shardCommitteeStateEnvironment) ShardID() byte {
	return env.shardID
}

//MaxShardCommitteeSize :
func (env *shardCommitteeStateEnvironment) MaxShardCommitteeSize() int {
	return env.maxShardCommitteeSize
}

//MinShardCommitteeSize :
func (env *shardCommitteeStateEnvironment) MinShardCommitteeSize() int {
	return env.minShardCommitteeSize
}

//Offset :
func (env *shardCommitteeStateEnvironment) Offset() int {
	return env.offset
}

//SwapOffset :
func (env *shardCommitteeStateEnvironment) SwapOffset() int {
	return env.swapOffset
}

//ProducersBlackList :
func (env *shardCommitteeStateEnvironment) ProducersBlackList() map[string]uint8 {
	return env.producersBlackList
}

//StakingTx :
func (env *shardCommitteeStateEnvironment) StakingTx() map[string]string {
	return env.stakingTx
}

//IsProcessShardBlockInstructionForKeyListV2 :
func (env *shardCommitteeStateEnvironment) IsProcessShardBlockInstructionForKeyListV2() bool {
	return env.isProcessShardBlockInstructionForKeyListV2
}

//Build :
func (env *shardCommitteeStateEnvironment) Build() ShardCommitteeStateEnvironment {
	return env
}
