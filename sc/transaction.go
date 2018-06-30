package sc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"
)

// TxOut transaction output
type TxOut struct {
	Value        int
	ScriptPubKey string
}

// TxIn transaction input
type TxIn struct {
	Txid      Hash
	Vout      int
	ScriptSig []byte
}

// Transaction preresent a transaction
type Transaction struct {
	ID   Hash
	Vin  []TxIn
	Vout []TxOut
}

// IsCoinBase return true if the transaction is coinbase transaction
func (txIn *TxIn) IsCoinBase() bool {
	return txIn.Vout == -1
}

// CanUnlock check if the transaction input can unlock the given output
func (txIn *TxIn) CanUnlock(txOut TxOut) bool {
	return verifyOwnership(txIn.ScriptSig, txOut.ScriptPubKey)
}

// CalHash return hash of the transaction
func (tx Transaction) CalHash() Hash {
	return hash256(toBytes(tx))
}

// SetID set id of the transaction
func (tx *Transaction) SetID() {
	tx.ID = hash256(bytes.Join([][]byte{toBytes(tx), []byte(time.Now().String())}, []byte{}))
}

// NewCoinbase return a coinbase transaction
func NewCoinbase(to string) *Transaction {
	txIn := TxIn{
		Txid:      []byte{},
		Vout:      -1,
		ScriptSig: []byte(""),
	}
	txOut := TxOut{
		Value:        reward,
		ScriptPubKey: to,
	}

	tx := &Transaction{
		ID:   nil,
		Vin:  []TxIn{txIn},
		Vout: []TxOut{txOut},
	}
	tx.SetID()
	return tx
}

// Print print details of transaction to console
func (tx *Transaction) Print() {
	fmt.Println("Tx: ", hex.EncodeToString(tx.ID))
	for _, vin := range tx.Vin {
		fmt.Printf("\tVIn: \n\t\tTxId: %v\n\t\tVout: %v\n\t\tScriptSig: %v\n\n", hex.EncodeToString(vin.Txid), vin.Vout, vin.ScriptSig)
	}
	for _, vout := range tx.Vout {
		fmt.Printf("\tVOut: \n\t\tValue: %v\n\t\tScriptPubKey: %v\n\n", vout.Value, vout.ScriptPubKey)
	}
	fmt.Println()
}
