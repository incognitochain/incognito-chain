package env

type ValidationEnviroment interface {
	IsPrivacy() bool
	IsConfimed() bool
	TxType() string
	TxAction() int
	ShardID() int
	ShardHeight() uint64
	BeaconHeight() uint64
	ConfirmedTime() int64
	Version() int
}
