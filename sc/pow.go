package sc

import (
	"bytes"
	"encoding/hex"
	"strings"
)

// PoWer is an interface of proof of work
type PoWer interface {
	Work(b *Block) int
}

// SimPow a simple proof of  work
type SimPow struct {
}

// NewSimPow return a new proof of work object
func NewSimPow() *SimPow {
	return &SimPow{}
}

// data prepare data for producing proof of work
func (pow *SimPow) data(block *Block, nonce int) []byte {
	return bytes.Join([][]byte{
		toBytes(block.Timestamp),
		block.PrevHash,
		toBytes(block.Difficulty),
		toBytes(nonce),
		block.MerkleRoot(),
	}, []byte{})
}

// Work perform the proof of work. The job is considered as done if the hash of data
// contains same # of leading zero as the target difficulty
func (pow *SimPow) Work(block *Block) int {
	nonce := 0
	prefix := strings.Repeat("0", difficulty)
	for {
		v := hex.EncodeToString(hash256(pow.data(block, nonce)))
		if strings.HasPrefix(v, prefix) {
			return nonce
		}
		nonce++
	}
}
