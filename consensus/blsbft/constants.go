package blsbft

import "time"

const (
	PROPOSE  = "PROPOSE"
	LISTEN   = "LISTEN"
	AGREE    = "AGREE"
	NEWROUND = "NEWROUND"
)

//
const (
	TIMEOUT                 = 60 * time.Second
	HIGHEST_BLOCK_CONFIDENT = 3
)
