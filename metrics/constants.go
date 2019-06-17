package metrics

const (
	Measurement      = "Measurement"
	Tag              = "Tag"
	TagValue         = "TagValue"
	MeasurementValue = "MeasurementValue"
	GrafanaURL       = "http://128.199.96.206:8086/write?db=mydb"
)

// Measurement
const (
	TxPoolValidated                   = "TxPoolValidated"
	TxPoolValidationDetails           = "TxPoolValidationDetails"
	TxPoolValidatedWithType           = "TxPoolValidatedWithType"
	TxPoolEntered                     = "TxPoolEntered"
	TxPoolEnteredWithType             = "TxPoolEnteredWithType"
	TxPoolAddedAfterValidation        = "TxPoolAddedAfterValidation"
	TxPoolRemoveAfterInBlock          = "TxPoolRemoveAfterInBlock"
	TxPoolRemoveAfterInBlockWithType  = "TxPoolRemoveAfterInBlockWithType"
	TxPoolRemoveAfterLifeTime         = "TxPoolRemoveAfterLifeTime"
	TxAddedIntoPoolType               = "TxAddedIntoPoolType"
	TxPoolPrivacyOrNot                = "TxPoolPrivacyOrNot"
	PoolSize                          = "PoolSize"
	TxInOneBlock                      = "TxInOneBlock"
	TxPoolDuplicateTxs                = "DuplicateTxs"
	NumOfBlockInsertToChain           = "NumOfBlockInsertToChain"
	TxPoolRemovedNumber               = "TxPoolRemovedNumber"
	TxPoolRemovedTime                 = "TxPoolRemovedTime"
	TxPoolRemovedTimeDetails          = "TxPoolRemovedTimeDetails"
	TxPoolTxBeginEnter                = "TxPoolTxBeginEnter"
)

// tag
const (
	BlockHeightTag       = "blockheight"
	TxSizeTag            = "txsize"
	TxSizeWithTypeTag    = "txsizewithtype"
	PoolSizeMetric       = "poolsize"
	TxTypeTag            = "txtype"
	ValidateConditionTag = "validatecond"
	TxPrivacyOrNotTag    = "txprivacyornot"
	ShardIDTag           = "shardid"
)

//Tag value
const (
	Beacon                                = "beacon"
	Shard                                 = "shard"
	TxPrivacy                             = "privacy"
	TxNormalPrivacy                       = "normaltxprivacy"
	TxNoPrivacy                           = "noprivacy"
	TxNormalNoPrivacy                     = "normaltxnoprivacy"
	Condition1                            = "condition1"
	Condition2                            = "condition2"
	Condition3                            = "condition3"
	Condition4                            = "condition4"
	Condition5                            = "condition5"
	Condition6                            = "condition6"
	Condition7                            = "condition7"
	Condition8                            = "condition8"
	Condition9                            = "condition9"
	Condition10                           = "condition10"
	Condition11                           = "condition11"
	VTBITxTypeMetic                       = "vtbitxtype"
)
