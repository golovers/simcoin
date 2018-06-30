package sc

import (
	"bytes"
	"testing"
)

func TestBase58(t *testing.T) {
	v := "test value"
	eV := EncodeBase58([]byte(v))
	dV := DecodeBase58(eV)
	if bytes.Compare([]byte(v), dV) != 0 {
		t.Errorf("base58 encoding works incorrectly")
	}
}
