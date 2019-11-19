package wire

const (
	// Message version
	Version = 1

	// Total in bytes of header message
	MessageHeaderSize = 24

	// size of cmd type in header message
	MessageCmdTypeSize = 12

	MaxBlockPayload = 2048000 // 2 Mb

	MaxGetAddrPayload = 1000 // 1 1Kb
)
