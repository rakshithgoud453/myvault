// Package memguard provides helpers for handling sensitive data in memory.
//
// Go's GC does not guarantee when (or whether) memory is reclaimed and zeroed.
// These helpers zero sensitive byte slices as soon as they are no longer needed,
// reducing the window in which secrets are visible in a core dump, swap file,
// or memory inspection tool.
//
// Note: strings in Go are immutable and cannot be zeroed after creation.
// Where possible, pass passphrases as []byte and zero them with ZeroBytes.
package memguard

// ZeroBytes overwrites b with zeros.
// Call this as soon as the sensitive data is no longer needed:
//
//	data, _ := crypto.Decrypt(...)
//	defer memguard.ZeroBytes(data)
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
