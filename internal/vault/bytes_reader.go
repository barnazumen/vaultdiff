package vault

import "bytes"

// bytesReader wraps a byte slice in an io.Reader, provided as a small helper
// shared across vault package HTTP helpers.
func bytesReader(b []byte) *bytes.Reader {
	return bytes.NewReader(b)
}
