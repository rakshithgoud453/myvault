package memguard_test

import (
	"testing"

	"github.com/rakshithgoud453/myvault/internal/memguard"
)

func TestZeroBytes(t *testing.T) {
	buf := []byte("secretpassphrase123")
	memguard.ZeroBytes(buf)

	for i, b := range buf {
		if b != 0 {
			t.Errorf("byte at index %d was not zeroed: %d", i, b)
		}
	}
}

func TestZeroBytesNilOrEmpty(t *testing.T) {
	// Should not panic
	memguard.ZeroBytes(nil)
	memguard.ZeroBytes([]byte{})
}
