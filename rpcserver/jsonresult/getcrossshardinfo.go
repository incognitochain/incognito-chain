package jsonresult

type CrossOutputCoinResult struct {
	SenderShardID byte `json:"SenderShardID"`
	ReceiverShardID byte `json:"ReceiverShardID"`
	BlockHeight uint64 `json:"BlockHeight"`
	BlockHash string `json:"BlockHash"`
	PaymentAddress string `json:"PaymentAddress"`
	Value uint64 `json:"Value"`
}
type CrossCustomTokenPrivacyResult struct {
	SenderShardID byte `json:"SenderShardID"`
	ReceiverShardID byte `json:"ReceiverShardID"`
	BlockHeight uint64 `json:"BlockHeight"`
	BlockHash string `json:"BlockHash"`
	PaymentAddress string `json:"PaymentAddress"`
	TokenID string `json:"TokenID"`
	Value uint64 `json:"Value"`
}
type CrossCustomTokenResult struct {
	SenderShardID byte `json:"SenderShardID"`
	ReceiverShardID byte `json:"ReceiverShardID"`
	BlockHeight uint64 `json:"BlockHeight"`
	BlockHash string `json:"BlockHash"`
	PaymentAddress string `json:"PaymentAddress"`
	TokenID string `json:"TokenID"`
	Value uint64 `json:"Value"`
}
