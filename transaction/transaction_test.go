package transaction

import (
	"fmt"
	"testing"
)

func TestAddAccount_ExistingID(t *testing.T) {
	x := NewTransaction()
	a := &Account{Name: "a1"}
	k := x.AddAccount(a, 12345)

	if k != 12345 {
		t.Errorf("Expected key to be 12345, got %v", k)
	}

	if x.accountMap[k] != a {
		t.Errorf("Expected account to be %v, got %v", a, x.accountMap[k])
	}
}

func TestAddAccount_NewID(t *testing.T) {
	x := NewTransaction()
	a := &Account{Name: "a1"}
	k := x.AddAccount(a, 0)

	if k == 0 {
		t.Errorf("Expected nonzero key, got %v", k)
	}

	if x.accountMap[k] != a {
		t.Errorf("Expected account to be %v, got %v", a, x.accountMap[k])
	}
}

func TestAddAccount_UniqueIDs(t *testing.T) {
	x := NewTransaction()
	a1, a2 := &Account{Name: "a1"}, &Account{Name: "a2"}
	k1, k2 := x.AddAccount(a1, 0), x.AddAccount(a2, 0)

	if k1 == k2 {
		t.Errorf("Expected unique account keys, but both were %v", k1)
	}

	if x.accountMap[k1] != a1 {
		t.Errorf("Expected k1 to be %v, got %v", a1, x.accountMap[k1])
	}
	if x.accountMap[k2] != a2 {
		t.Errorf("Expected k2 to be %v, got %v", a2, x.accountMap[k2])
	}
}

func TestValidTransaction(t *testing.T) {
	x := NewTransaction()
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
	x := NewTransaction()
	if err := x.ValidateAmount(); err == nil {
		t.Errorf("Empty transaction %v had valid amount", x)
	}
	if err := x.ValidateAccounts(); err == nil {
		t.Errorf("Empty transaction %v had valid accounts", x)
	}
}

func TestInvalidTransaction_NonZero(t *testing.T) {
	x := NewTransaction()
	k1 := x.AddAccount(&Account{Name: "a1"}, 0)
	k2 := x.AddAccount(&Account{Name: "a2"}, 0)
	x.AddSplits([]*Split{&Split{Amount: 4, Account: k1}, &Split{Amount: -3, Account: k2}})

	if err := x.ValidateAmount(); err == nil {
		t.Errorf("Nonzero transaction %v had valid amount", x)
	}
	if err := x.ValidateAccounts(); err != nil {
		t.Errorf("Transaction %v had invalid accounts: %v", x, err)
	}
}

func TestInvalidTransaction_AccountNotPresent(t *testing.T) {
	x := NewTransaction()
	k1 := x.AddAccount(&Account{Name: "a1"}, 0)
	x.AddSplits([]*Split{&Split{Amount: 4, Account: k1}, &Split{Amount: -4, Account: k1 + 1}})

	if err := x.ValidateAmount(); err != nil {
		t.Errorf("Transaction %v had invalid amount: %v", x, err)
	}
	if err := x.ValidateAccounts(); err == nil {
		t.Errorf("Transaction %v had valid accounts", x)
	}
}

func TestInvalidTransaction_DuplicateAccount(t *testing.T) {
	x := NewTransaction()
	k1 := x.AddAccount(&Account{Name: "a1"}, 0)
	x.AddSplits([]*Split{&Split{Amount: 4, Account: k1}, &Split{Amount: -4, Account: k1}})

	if err := x.ValidateAmount(); err != nil {
		t.Errorf("Transaction %v had invalid amount: %v", x, err)
	}
	if err := x.ValidateAccounts(); err == nil {
		t.Errorf("Transaction %v had valid accounts", x)
	}
}

func TestCommit(t *testing.T) {
	x := NewTransaction()
	a1, a2 := &Account{Name: "a1"}, &Account{Name: "a2"}
	k1, k2 := x.AddAccount(a1, 0), x.AddAccount(a2, 0)
	x.AddSplits([]*Split{&Split{Amount: -100, Account: k1}, &Split{Amount: 100, Account: k2}})

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
	x := NewTransaction()
	salary, checking, savings := &Account{Name: "Salary"}, &Account{Name: "Checking"}, &Account{Name: "Savings"}
	salaryKey, checkingKey, savingsKey := x.AddAccount(salary, 0), x.AddAccount(checking, 0), x.AddAccount(savings, 0)
	fmt.Printf("Initial amounts: salary: %v, checking: %v, savings: %v\n",
		salary.total, checking.total, savings.total)

	x.AddSplits([]*Split{
		&Split{Amount: -1000, Account: salaryKey},
		&Split{Amount: 800, Account: checkingKey},
	})
	fmt.Printf("Transaction error: %v\n", x.Commit())

	x.AddSplit(&Split{Amount: 200, Account: savingsKey})
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
