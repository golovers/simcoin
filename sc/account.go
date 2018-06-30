package sc

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

const version = byte(0x00)
const addressChecksumLen = 4

// Account account
type Account struct {
	PubKey PubKey
	PriKey ecdsa.PrivateKey
}

// NewAccount return a new wallet
func NewAccount() *Account {
	acc := &Account{}
	priv, pub := newKeyPair()
	acc.PriKey = priv
	acc.PubKey = pub
	return acc
}

func newKeyPair() (ecdsa.PrivateKey, PubKey) {
	curve := elliptic.P256()
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err)
	}
	pub := append(priv.PublicKey.X.Bytes(), priv.PublicKey.Y.Bytes()...)
	return *priv, pub[:]

}

// GetAddress get address
func (acc *Account) GetAddress() Address {
	payload := hash160(acc.PubKey)
	payload = append([]byte{version}, payload...)
	payload = append(payload, checksum(payload)...)

	return Address(EncodeBase58([]byte(payload)))
}

func checksum(payload []byte) []byte {
	return hash256(hash256(payload))[:addressChecksumLen]
}

// ValidateAddress check if address is valid
func ValidateAddress(address string) bool {
	pubHash := DecodeBase58(address)
	actualChecksum := pubHash[len(pubHash)-addressChecksumLen:]
	version := pubHash[0]
	pubKeyHash := pubHash[1 : len(pubHash)-addressChecksumLen]
	newChecksum := checksum(append([]byte{version}, pubKeyHash...))
	return bytes.Compare(actualChecksum, newChecksum) == 0
}
