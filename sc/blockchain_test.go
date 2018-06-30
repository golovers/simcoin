package sc

import (
	"testing"
)

func TestVerifyOwnership(t *testing.T) {
	miner := NewAccount()
	db, _ := NewMemDatabase()
	blockchain := NewBlockchain(miner, db)

	if err := blockchain.Validate(); err != nil {
		t.Errorf("invalid blockchain\n")
	}

	sig := blockchain.ScriptSig(miner)
	scriptPubKey := blockchain.ScriptPubKey(miner.GetAddress())

	if !verifyOwnership(sig, scriptPubKey) {
		t.Errorf("failed to verify ownership")
	}
}

func TestTransactions(t *testing.T) {
	miner := NewAccount()
	db, _ := NewMemDatabase()
	blockchain := NewBlockchain(miner, db)

	if err := blockchain.Validate(); err != nil {
		t.Errorf("invalid blockchain\n")
	}
	w := NewMemWallet(blockchain)
	w.Add("miner", miner)
	w.Add("alice", NewAccount())
	w.Add("bob", NewAccount())

	blockchain.Mine(5)
	assertEquals(t, "miner", 25, w.Balance("miner"))
	assertEquals(t, "bob", 0, w.Balance("bob"))
	assertEquals(t, "alice", 0, w.Balance("alice"))

	w.Send("miner", "alice", 2)
	w.Send("miner", "bob", 2)
	w.Send("alice", "bob", 1)

	assertEquals(t, "miner", 36, w.Balance("miner"))
	assertEquals(t, "bob", 3, w.Balance("bob"))
	assertEquals(t, "alice", 1, w.Balance("alice"))
}

func assertEquals(t *testing.T, acc1 string, v1, v2 int) {
	if v1 != v2 {
		t.Errorf("balance of %s should be %d but got %d\n", acc1, v1, v2)
	}
}
