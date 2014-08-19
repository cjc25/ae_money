// Package transaction implements basic double-entry accounting.
package transaction

import (
	"errors"
	"fmt"

	"code.google.com/p/go-uuid/uuid"
)

// The type used for a Split's value.
type AmountType int64

// The type of unique identifier of each Transaction.
//
// This is a UUID in string form, for ease of use as an appengine datastore.Key
// string ID (which must be a UTF-8 string instead of an arbitrary []byte).
type TransactionID string

// A Split is the addition or subtraction of an amount from a single Account,
// as part of a Transaction.
type Split struct {
	Amount        AmountType
	account       *Account
	transactionID TransactionID
}

// Create a new Split that targets a given Account.
func NewSplit(amount AmountType, account *Account) *Split {
	return &Split{amount, account, ""}
}

func (s *Split) Account() *Account {
	return s.account
}

func (s *Split) TransactionID() TransactionID {
	return s.transactionID
}

// A Transaction is a series of Splits that conform to double-entry Accounting
// rules, in order to transfer value between accounts.
//
// Under double-entry accounting, each credit must have a corresponding debit
// in a different account, likewise, a Transaction can't be Committed unless
// all of its Splits are for different Accounts, and add to 0.
//
// This means that some Accounts will be abstract, for example "Salary," which
// is debited to account for the credit to your checking account and will grow
// more negative over time, and "Expenses," which is credited to account for
// spending from your account.
type Transaction struct {
	splits []Split
	total  AmountType
}

// Convenience function to add multiple Splits to a Transaction.
func (x *Transaction) AddSplits(splits []*Split) {
	for _, split := range splits {
		x.AddSplit(split)
	}
}

func (x *Transaction) AddSplit(split *Split) {
	x.splits = append(x.splits, *split)
	x.total += split.Amount
}

// Make sure a Transaction can be successfully Committed.
//
// Verifies that the Transaction has Splits, they add to 0, and they all are for
// different Accounts. If these tests fail, it returns non-nil.
func (x *Transaction) Valid() error {
	if len(x.splits) == 0 {
		return errors.New("No Splits in Transaction.")
	}

	if x.total != 0 {
		return fmt.Errorf("Nonzero total: %v", x.total)
	}

	seen := make(map[*Account]bool)
	for _, split := range x.splits {
		if split.account == nil {
			return errors.New("Split with nil Account.")
		}
		if seen[split.account] {
			return errors.New("Multiple Splits for same Account.")
		}
		seen[split.account] = true
	}

	return nil
}

// Commit the Splits in x to their respective Accounts, if x is Valid.
func (x *Transaction) Commit() error {
	if err := x.Valid(); err != nil {
		return err
	}

	xid := TransactionID(uuid.NewRandom().String())

	for _, split := range x.splits {
		split.transactionID = xid
		split.account.splits = append(split.account.splits, split)
		split.account.total += split.Amount
	}

	x.clear()
	return nil
}

func (x *Transaction) clear() {
	x.total = 0
	x.splits = nil
}
