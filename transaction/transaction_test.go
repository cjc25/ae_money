package transaction

import (
	"fmt"
	"testing"
)

func TestValidTransaction(t *testing.T) {
	x := &Transaction{}

	x.AddSplits([]*Split{
		NewSplit(4, &Account{}), NewSplit(2, &Account{}), NewSplit(-6, &Account{}),
	})

	if err := x.Valid(); err != nil {
		t.Errorf("Expected valid transaction, was not: %v", err)
	}
}

func TestInvalidTransaction_Empty(t *testing.T) {
	x := &Transaction{}
	if err := x.Valid(); err == nil {
		t.Error("Empty transaction was valid")
	}
}

func TestInvalidTransaction_NonZero(t *testing.T) {
	x := &Transaction{}
	x.AddSplits([]*Split{NewSplit(4, &Account{}), NewSplit(-3, &Account{})})
	if err := x.Valid(); err == nil {
		t.Errorf("Expected invalid transaction, was: %v", x)
	}
}

func TestInvalidTransaction_NilAccount(t *testing.T) {
	x := &Transaction{}
	x.AddSplits([]*Split{NewSplit(4, nil), NewSplit(-4, &Account{})})
	if err := x.Valid(); err == nil {
		t.Errorf("Expected invalid transaction, was: %v", x)
	}
}

func TestInvalidTransaction_DuplicateAccount(t *testing.T) {
	x := &Transaction{}
	acct := Account{}
	x.AddSplits([]*Split{NewSplit(4, &acct), NewSplit(-4, &acct)})
	if err := x.Valid(); err == nil {
		t.Errorf("Expected invalid transaction, was: %v", x)
	}
}

func TestCommit_TransactionID(t *testing.T) {
	a1, a2 := Account{}, Account{}
	x := &Transaction{}
	x.AddSplits([]*Split{NewSplit(-100, &a1), NewSplit(100, &a2)})
	if err := x.Commit(); err != nil {
		t.Fatalf("Expected initial commit to be valid: %v", x)
	}

	if a1.splits[0].transactionID != a2.splits[0].transactionID {
		t.Errorf("Expected equal transaction IDs. Instead, a1: '%v', a2: '%v",
			a1.splits[0].transactionID, a2.splits[0].transactionID)
	}
}

func TestDoubleCommit(t *testing.T) {
	a1, a2 := Account{}, Account{}
	x := &Transaction{}
	x.AddSplits([]*Split{NewSplit(-100, &a1), NewSplit(100, &a2)})
	if err := x.Commit(); err != nil {
		t.Fatalf("Expected initial commit to be valid: %v", x)
	}

	if err := x.Commit(); err == nil {
		t.Errorf("Expected second commit to be invalid: error: %v, x: %v", err, x)
	}
	if len(a1.splits) != 1 || len(a2.splits) != 1 {
		t.Errorf("Expected accounts to have a single split: a1: %v, a2: %v", a1, a2)
	}
}

func ExampleTransaction() {
	salary, checking, savings := Account{}, Account{}, Account{}
	fmt.Printf("Initial amounts: salary: %v, checking: %v, savings: %v\n",
		salary.total, checking.total, savings.total)

	x := &Transaction{}
	x.AddSplits([]*Split{
		NewSplit(-1000, &salary),
		NewSplit(800, &checking),
	})
	fmt.Printf("Transaction error: %v\n", x.Commit())

	x.AddSplit(NewSplit(200, &savings))
	fmt.Println("New split added")
	fmt.Printf("x.Commit() successful?: %v\n", x.Commit() == nil)

	fmt.Printf("Final amounts: salary: %v, checking: %v, savings: %v\n",
		salary.total, checking.total, savings.total)
	fmt.Printf("Splits per account: salary: %v, checking: %v, savings: %v\n",
		len(salary.splits), len(checking.splits), len(savings.splits))

	// Output:
	// Initial amounts: salary: 0, checking: 0, savings: 0
	// Transaction error: Nonzero total: -200
	// New split added
	// x.Commit() successful?: true
	// Final amounts: salary: -1000, checking: 800, savings: 200
	// Splits per account: salary: 1, checking: 1, savings: 1
}
