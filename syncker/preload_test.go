package syncker

import (
	"testing"
)

func Test_preloadDatabase(t *testing.T) {
	preloadDatabase(1, 101, "http://127.0.0.1:20004", "", "", nil)
}
