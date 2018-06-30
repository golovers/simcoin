package sc

import (
	"bytes"
	"fmt"
	"time"
)

// Block prepresent a block in the blockchain
type Block struct {
	Timestamp  time.Time
	PrevHash   Hash
	Difficulty int
	Nonce      int

	Transactions []*Transaction
}

// MerkleRoot return hash of the transactions
func (block *Block) MerkleRoot() Hash {
	if len(block.Transactions) == 0 {
		return hash256([]byte{})
	}
	if len(block.Transactions) == 1 {
		data := bytes.Join([][]byte{block.Transactions[0].CalHash(), block.Transactions[0].CalHash()}, []byte{})
		return hash256(data)
	}
	merkle := make([]Hash, 0)
	for _, tx := range block.Transactions {
		merkle = append(merkle, tx.CalHash())
	}
	for len(merkle) > 1 {
		tmpMerkle := make([]Hash, 0)
		if len(merkle)%2 != 0 {
			merkle = append(merkle, merkle[len(merkle)-1])
		}
		for i := 0; i < len(merkle); i++ {
			v := bytes.Join([][]byte{merkle[i], merkle[i+1]}, []byte{})
			tmpMerkle = append(tmpMerkle, hash256(v))
			i++
		}
		merkle = tmpMerkle
	}
	return merkle[0]
}

// CalHash return hash of the block
func (block *Block) CalHash() Hash {
	b := bytes.Join([][]byte{toBytes(block.Timestamp), block.PrevHash, toBytes(block.Difficulty),
		toBytes(block.Nonce), block.MerkleRoot()}, []byte{})
	return hash256(b)
}

// newBlock mine a new block with the given transactions
func newBlock(transactions []*Transaction, prevHash Hash) *Block {
	b := &Block{
		Transactions: transactions,
		PrevHash:     prevHash,
		Timestamp:    time.Now(),
		Difficulty:   difficulty,
	}
	return b
}

// IsGenesis return if this block is genesis block
func (block *Block) IsGenesis() bool {
	return len(block.PrevHash) == 0
}

// Print print block info to console
func (block *Block) Print() {
	fmt.Printf("Hash: %s\n", block.CalHash().String())
	fmt.Printf("PrevHash: %s\n", block.PrevHash.String())
	fmt.Println("Nonce:", block.Nonce)
	fmt.Println()
}
