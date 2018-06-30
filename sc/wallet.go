package sc

import (
	"fmt"
)

// Wallet represent a place to store priv/pub keys and allow to send money
type Wallet interface {
	Info(accName string)
	Send(from, to string, amount int)
	Print()
	Add(name string, acc *Account)
}

// MemWallet represent a memory wallet
type MemWallet struct {
	accounts map[string]*Account
	bc       *Blockchain
}

// NewMemWallet return a new memory wallet
func NewMemWallet(bc *Blockchain) *MemWallet {
	return &MemWallet{
		accounts: make(map[string]*Account),
		bc:       bc,
	}
}

// Info print information of the account by the given name
func (w *MemWallet) Info(accName string) {
	balance, _ := w.bc.Spendable(w.accounts[accName], -1)
	fmt.Printf(`
	----------------------------------
		Name: %s
		Balance: %d $C
	----------------------------------
		`, accName, balance)
}

// Send sending money from an account to another account
func (w *MemWallet) Send(from, to string, amount int) {
	if _, ok := w.accounts[from]; !ok {
		fmt.Printf("error: <%s> account not found\n", from)
		return
	}
	if _, ok := w.accounts[to]; !ok {
		fmt.Printf("error: <%s> account not found\n", to)
		return
	}
	w.bc.Send(w.accounts[from], w.accounts[to], amount)
}

// Add a new account
func (w *MemWallet) Add(name string, acc *Account) {
	w.accounts[name] = acc
}

// Print info of all available accounts
func (w *MemWallet) Print() {
	for accName := range w.accounts {
		w.Info(accName)
	}
}

// Balance return balance of the given account
func (w *MemWallet) Balance(accName string) int {
	balance, _ := w.bc.Spendable(w.accounts[accName], -1)
	return balance
}
