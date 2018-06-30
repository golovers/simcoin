package sc

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

var difficulty = 2
var reward = 5
var sigLen = 64
var scriptPubKey = "OP_DUP OP_HASH160 %s OP_EQUALVERIFY OP_CHECKSIG"
var errorNotHisMoney = errors.New("error: this guy is trying to spend money of someone else")
var errorNotEnoughMoney = errors.New("error: this guy is trying spend more that what he has")
var shatoshiNakamotoAddress = Address("1NHXs8UxcgHzDNxWNTcYjKv8MGY72rnbbE")

var poWer PoWer = NewSimPow()

// Blockchain the main chain
type Blockchain struct {
	db    Database
	miner *Account
}

// NewBlockchain return a new blockchain with genesis block inside
func NewBlockchain(miner *Account, db Database) *Blockchain {
	bc := &Blockchain{
		db:    db,
		miner: miner,
	}
	bc.addGenesisBlock()
	return bc
}

func (bc *Blockchain) addGenesisBlock() {
	if lb, _ := bc.db.Get(lastBlockKey); len(lb) == 0 {
		b := newBlock([]*Transaction{NewCoinbase(bc.ScriptPubKey(shatoshiNakamotoAddress))}, Hash{})
		b.Nonce = poWer.Work(b)
		bc.addBlock(b)
	}
}

// SetPoWer set the PoWer
func SetPoWer(power PoWer) {
	poWer = power
}

func (bc *Blockchain) addBlock(b *Block) {
	if !bc.isValidBlock(b) {
		fmt.Println("error: fake block....")
		return
	}
	bc.db.Put(b.CalHash(), toBytes(b))
	bc.db.Put(lastBlockKey, toBytes(b))
}

// isValidBlock validate the block is valid
func (bc *Blockchain) isValidBlock(block *Block) bool {
	// validate proof of work
	prefix := strings.Repeat("0", block.Difficulty)
	hashValue := hex.EncodeToString(block.CalHash())
	if !strings.HasPrefix(hashValue, prefix) {
		fmt.Println("error: invalid proof of work")
		return false
	}

	// ignore other validations if it is the genesis
	if block.IsGenesis() {
		return true
	}
	// verify its parent
	prevBlock := bc.getBlock(block.PrevHash)
	if bytes.Compare(block.PrevHash, prevBlock.CalHash()) != 0 {
		return false
	}
	// verify the transactions are valid; don't need to validate the coinbase
	for _, tx := range block.Transactions[1:] {
		if err := bc.validateTransaction(tx); err != nil {
			fmt.Println("error: invalid transaction")
			tx.Print()
			return false
		}
	}
	return true
}

func (bc *Blockchain) getBlock(key []byte) *Block {
	var b *Block
	data, _ := bc.db.Get(key)
	if len(data) == 0 {
		return nil
	}
	toObject(data, &b)
	return b
}

// MineNewBlock add a new block into the blockchain
func (bc *Blockchain) MineNewBlock(transactions []*Transaction) {
	var prevBlock *Block
	v, _ := bc.db.Get(lastBlockKey)
	toObject(v, &prevBlock)
	validTxs := bc.validTransactions(transactions)
	// add reward for mining a block
	txs := []*Transaction{NewCoinbase(bc.ScriptPubKey(bc.miner.GetAddress()))}
	txs = append(txs, validTxs...)
	b := newBlock(txs, prevBlock.CalHash())
	b.Nonce = poWer.Work(b)
	bc.addBlock(b)
}

// Spendable return total amount up to the given amount and prepare list transaction input for spending
func (bc *Blockchain) Spendable(acc *Account, amount int) (total int, spendable []TxIn) {
	spent := make(map[string][]bool)
	txs := make(map[string]*Transaction)
	it := NewBlockIterator(bc.db)
	for b := it.Next(); b != nil; b = it.Next() {
		for _, tx := range b.Transactions {
			txs[tx.ID.String()] = tx
			spent[tx.ID.String()] = make([]bool, len(tx.Vout))
		}
	}
	// mark all spent transaction output...
	it = NewBlockIterator(bc.db)
	for b := it.Next(); b != nil; b = it.Next() {
		for _, tx := range b.Transactions {
			for _, vin := range tx.Vin {
				if !vin.IsCoinBase() {
					spent[vin.Txid.String()][vin.Vout] = true
				}
			}
		}
	}
	utxos := make([]TxOut, 0)
	spendableTxIns := make([]TxIn, 0)
	for txid, spt := range spent {
		for idx, alreadySpent := range spt {
			if !alreadySpent {
				utxos = append(utxos, txs[txid].Vout[idx])
				spendableTxIns = append(spendableTxIns, TxIn{
					Txid: StringToHash(txid),
					Vout: idx,
				})
			}
		}
	}
	spendable = make([]TxIn, 0)
	total = 0
	for idx, txout := range utxos {
		if verifyOwnership(bc.ScriptSig(acc), txout.ScriptPubKey) {
			txInV := spendableTxIns[idx]
			txInV.ScriptSig = bc.ScriptSig(acc)
			spendable = append(spendable, txInV)
			// if amount == -1 {
			// 	total += txout.Value
			// } else {
			// 	if total >= amount {
			// 		break
			// 	}
			// 	total += txout.Value
			// }
			total += txout.Value
		}
	}
	return
}

// Send sending money from an address to another address
func (bc *Blockchain) Send(from *Account, to *Account, amount int) {
	total, spendableTxIns := bc.Spendable(from, amount)
	if total < amount {
		panic(errorNotEnoughMoney)
	}
	vins := make([]TxIn, 0)
	for _, vin := range spendableTxIns {
		newvin := vin
		newvin.ScriptSig = bc.ScriptSig(from)
		vins = append(vins, newvin)
	}
	vouts := []TxOut{
		TxOut{
			Value:        amount,
			ScriptPubKey: bc.ScriptPubKey(to.GetAddress()),
		},
	}
	// sending change to the owner
	if total > amount {
		vouts = append(vouts, TxOut{
			Value:        total - amount,
			ScriptPubKey: bc.ScriptPubKey(from.GetAddress()),
		})
	}
	tx := &Transaction{
		Vin:  vins,
		Vout: vouts,
	}
	tx.SetID()
	bc.MineNewBlock([]*Transaction{tx})
}

// validateTransaction go over the block chain and check if the input can unlock the output it refers to
//TODO should cache the unspent tractions output somewhere...don't need to scan entire the blockchain for this...
func (bc *Blockchain) validateTransaction(tx *Transaction) error {
	// check if vin can be unlocked
	inAmount := 0
	it := NewBlockIterator(bc.db)
	for b := it.Next(); b != nil; b = it.Next() {
		for _, tran := range b.Transactions {
			for _, vin := range tx.Vin {
				if bytes.Compare(tran.ID, vin.Txid) == 0 {
					vout := tran.Vout[vin.Vout]
					if !vin.CanUnlock(vout) {
						return errorNotHisMoney
					}
					inAmount += vout.Value
				}
			}
		}
	}
	// check if the total amount in >= out amount...
	outAmount := 0
	for _, vout := range tx.Vout {
		outAmount += vout.Value
		if outAmount > inAmount {
			return errorNotEnoughMoney
		}
	}
	return nil
}

// validTransactions return valid transaction and ignore others
func (bc *Blockchain) validTransactions(trans []*Transaction) []*Transaction {
	txs := make([]*Transaction, 0)
	for _, tx := range trans {
		if err := bc.validateTransaction(tx); err != nil {
			fmt.Println("error: invalid transaction", err)
		} else {
			txs = append(txs, tx)
		}
	}
	return txs
}

// ScriptSig return scriptSig for unlocking coin
func (bc *Blockchain) ScriptSig(acc *Account) []byte {
	r, s, _ := ecdsa.Sign(rand.Reader, &acc.PriKey, hash160(acc.PubKey))
	sig := append(r.Bytes(), s.Bytes()...)
	scriptSig := append(sig, acc.PubKey...)
	return scriptSig[:]
}

// ScriptPubKey return P2PKH script for sending coin
func (bc *Blockchain) ScriptPubKey(address Address) string {
	return fmt.Sprintf(scriptPubKey, address)
}

// verifyOwnership execute P2PKH script: OP_DUP OP_HASH160 <pub key hash> OP_EQUALVERIFY OP_CHECKSIG
func verifyOwnership(scriptSig []byte, scriptPubKey string) bool {
	sig := scriptSig[:sigLen]
	pubKey := scriptSig[sigLen:]
	stack := &Stack{Values: make([][]byte, 0)}
	stack.Push(sig)
	stack.Push(pubKey)
	ops := strings.Fields(scriptPubKey)
	for _, op := range ops {
		if op == "OP_DUP" {
			stack.Push(stack.Peak())
		} else if op == "OP_HASH160" {
			v := stack.Pop()
			stack.Push(hash160(v))
		} else if op == "OP_EQUALVERIFY" {
			v1 := stack.Pop()
			v2 := stack.Pop()
			if bytes.Compare(v1, v2) != 0 {
				return false
			}
		} else if op == "OP_CHECKSIG" {
			var r, s big.Int
			r.SetBytes(sig[:len(sig)/2])
			s.SetBytes(sig[len(sig)/2:])
			var x big.Int
			x.SetBytes(pubKey[:len(pubKey)/2])
			var y big.Int
			y.SetBytes(pubKey[len(pubKey)/2:])
			pubkey := ecdsa.PublicKey{Curve: elliptic.P256(), X: &x, Y: &y}
			ok := ecdsa.Verify(&pubkey, hash160(pubKey), &r, &s)
			if !ok {
				return false
			}
		} else { // the address
			address := DecodeBase58(op)
			stack.Push(address[1 : len(address)-addressChecksumLen])
		}
	}
	return true
}

// Mine start mining blocks to get reward...
func (bc *Blockchain) Mine(n int) {
	for i := 0; i < n; i++ {
		bc.MineNewBlock([]*Transaction{})
	}
}

// PrintTransactions print all tractions that happen in the past
func (bc *Blockchain) PrintTransactions() {
	it := NewBlockIterator(bc.db)
	for b := it.Next(); b != nil; b = it.Next() {
		for _, tx := range b.Transactions {
			tx.Print()
		}
	}
}

// Print print entire the blockchain
func (bc *Blockchain) Print() {
	it := NewBlockIterator(bc.db)
	for b := it.Next(); b != nil; b = it.Next() {
		b.Print()
	}
}

//Validate validate if the blockchain is valid
func (bc *Blockchain) Validate() error {
	it := NewBlockIterator(bc.db)
	for b := it.Next(); b != nil; b = it.Next() {
		if !bc.isValidBlock(b) {
			return errors.New("error: invalid blockchain")
		}
	}
	return nil

}
