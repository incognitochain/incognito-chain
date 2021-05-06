package privacy

type ValidationEnviroment interface {
	IsPrivacy() bool
	IsConfimed() bool
	TxType() string
	TxAction() int
	ShardID() int
	ShardHeight() uint64
	BeaconHeight() uint64
	ConfimedTime() int64
	Version() int
}
