package sc

import "testing"

func TestGeneratingAddress(t *testing.T) {
	acc := NewAccount()
	address := acc.GetAddress()
	if !ValidateAddress(string(address)) {
		t.Errorf("address is not valid")
	}
}
