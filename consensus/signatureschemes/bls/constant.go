package bls

const (
	// CCmprPnSz Compress point size
	CCompressSize = 32
	// CBigIntSz Big Int Byte array size
	CBigIntSize = 32
	// CMaskByte 0b10000000
	CMaskByte = 0x80
	// CNotMaskB 0b01111111
	CNotMaskByte = 0x7F

	CCommiteeSize = 256
)

const (
	// CErr Error prefix
	CDetailErr = "Details error: "
	// CErrInps Error input length
	CInputError = "Input params error"
	// CErrCmpr Error when work with (de)compress
	CCompressError = "(De)Compress error"
)
