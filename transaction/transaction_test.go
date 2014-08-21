package transaction

import (
	"fmt"
	"testing"
)

func TestValidTransaction(t *testing.T) {
	x := &Transaction{}
	x.AddSplits([]*Split{&Split{Amount: 4}, &Split{Amount: 2}, &Split{Amount: -6}})

	if err := x.ValidateAmount(); err != nil {
		t.Errorf("Expected transaction %v to have valid amount but it did not: %v",
			x, err)
	}
	if err := x.ValidateAccounts(); err == nil {
		t.Errorf("Got valid accounts for transaction %v", x)
	}
}

func TestInvalidTransaction_Empty(t *testing.T) {
	x := &Transaction{}
	if err := x.ValidateAmount(); err == nil {
		t.Errorf("Empty transaction %v had valid amount", x)
	}
	if err := x.ValidateAccounts(); err == nil {
		t.Errorf("Empty transaction %v had valid accounts", x)
	}
}

func TestInvalidTransaction_NonZero(t *testing.T) {
	x := &Transaction{}
	_, k1 := NewAccount("a1", 0)
	_, k2 := NewAccount("a2", 0)
	x.AddSplits([]*Split{&Split{Amount: 4, account: k1}, &Split{Amount: -3, account: k2}})

	if err := x.ValidateAmount(); err == nil {
		t.Errorf("Nonzero transaction %v had valid amount", x)
	}
	if err := x.ValidateAccounts(); err != nil {
		t.Errorf("Transaction %v had invalid accounts: %v", x, err)
	}
}

func TestInvalidTransaction_AccountNotPresent(t *testing.T) {
	x := &Transaction{}
	_, k1 := NewAccount("a1", 0)
	x.AddSplits([]*Split{&Split{Amount: 4, account: k1}, &Split{Amount: -4, account: k1 + 1}})

	if err := x.ValidateAmount(); err != nil {
		t.Errorf("Transaction %v had invalid amount: %v", x, err)
	}
	if err := x.ValidateAccounts(); err == nil {
		t.Errorf("Transaction %v had valid accounts", x)
	}
}

func TestInvalidTransaction_DuplicateAccount(t *testing.T) {
	x := &Transaction{}
	_, k1 := NewAccount("a1", 0)
	x.AddSplits([]*Split{&Split{Amount: 4, account: k1}, &Split{Amount: -4, account: k1}})

	if err := x.ValidateAmount(); err != nil {
		t.Errorf("Transaction %v had invalid amount: %v", x, err)
	}
	if err := x.ValidateAccounts(); err == nil {
		t.Errorf("Transaction %v had valid accounts", x)
	}
}

func TestCommit(t *testing.T) {
	x := &Transaction{}
	a1, k1 := NewAccount("a1", 0)
	a2, k2 := NewAccount("a2", 0)
	x.AddSplits([]*Split{&Split{Amount: -100, account: k1}, &Split{Amount: 100, account: k2}})

	if err := x.Commit(); err != nil {
		t.Fatalf("Expected transaction %v to commit, but got: %v", x, err)
	}

	if a1.total != -100 {
		t.Errorf("Expected a1 total to be -100, but got %v", a1.total)
	}
	if a2.total != 100 {
		t.Errorf("Expected a2 total to be 100, but got %v", a2.total)
	}
}

func ExampleTransaction() {
	salary, salaryKey := NewAccount("Salary", 0)
	checking, checkingKey := NewAccount("Checking", 0)
	savings, savingsKey := NewAccount("Savings", 0)
	fmt.Printf("Initial amounts: salary: %v, checking: %v, savings: %v\n",
		salary.total, checking.total, savings.total)

	x := &Transaction{}
	x.AddSplits([]*Split{
		&Split{Amount: -1000, account: salaryKey},
		&Split{Amount: 800, account: checkingKey},
	})
	fmt.Printf("Transaction error: %v\n", x.Commit())

	x.AddSplit(&Split{Amount: 200, account: savingsKey})
	fmt.Println("New split added")
	fmt.Printf("x.Commit() successful?: %v\n", x.Commit() == nil)

	fmt.Printf("Final amounts: salary: %v, checking: %v, savings: %v\n",
		salary.total, checking.total, savings.total)

	// Output:
	// Initial amounts: salary: 0, checking: 0, savings: 0
	// Transaction error: Nonzero total: -200
	// New split added
	// x.Commit() successful?: true
	// Final amounts: salary: -1000, checking: 800, savings: 200
}
