package sc

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"

	"golang.org/x/crypto/ripemd160"
)

func hash256(data []byte) Hash {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

func hash160(pub []byte) Hash {
	ripemder := ripemd160.New()
	_, err := ripemder.Write(hash256(pub))
	if err != nil {
		panic(err)
	}
	return ripemder.Sum(nil)
}

func toBytes(v interface{}) []byte {
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(v)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func toObject(data []byte, v interface{}) error {
	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(v)
}

func copyBytes(b []byte) []byte {
	if len(b) == 0 {
		return []byte{}
	}
	newb := make([]byte, len(b))
	copy(newb, b)
	return newb
}
