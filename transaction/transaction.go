// Package transaction implements basic double-entry transactions.
package transaction

import (
	"errors"
	"fmt"
)

// The type used for a Split's value.
type AmountType int64

// A Split is the addition or subtraction of an amount from a single account,
// as part of a transaction.
type Split struct {
	Amount  AmountType
	Account int64
}

// A Transaction is a series of splits that conform to double-entry accounting
// rules, in order to transfer value between accounts.
//
// Under double-entry accounting, each credit must have a corresponding debit
// in a different account. Likewise, a transaction can't be committed unless
// all of its splits add to 0 and are for different accounts.
//
// This means that some accounts will be abstract, for example "Salary," which
// is debited to account for the credit to your checking account.
type Transaction struct {
	splits []*Split
	total  AmountType

	accountMap map[int64]*Account
	nextId     int64
}

// Create a new Transaction, which tracks accounts and splits.
func NewTransaction() *Transaction {
	return &Transaction{accountMap: make(map[int64]*Account), nextId: 1}
}

// Add an account to a transaction.
//
// If id is zero, a unique non-zero id will be created for it. The returned id
// should be set in the Account field of splits assigned to this account.
func (x *Transaction) AddAccount(a *Account, id int64) int64 {
	if id == 0 {
		// If we didn't provide an id, set a brand new one.
		for {
			_, ok := x.accountMap[x.nextId]
			if !ok {
				id = x.nextId
				break
			}
			x.nextId++
		}
	}

	x.accountMap[id] = a
	return id
}

// Convenience function to add multiple splits to a transaction.
func (x *Transaction) AddSplits(splits []*Split) {
	for _, split := range splits {
		x.AddSplit(split)
	}
}

// Add a single split to the transaction
func (x *Transaction) AddSplit(split *Split) {
	x.splits = append(x.splits, split)
	x.total += split.Amount
}

// Check that the transaction's splits have valid amounts.
//
// The splits are valid if there is at least one, none of them are for a 0
// amount, and their amounts all add to 0.
func (x *Transaction) ValidateAmount() error {
	if len(x.splits) == 0 {
		return errors.New("No splits in transaction.")
	}

	if x.total != 0 {
		return fmt.Errorf("Nonzero total: %v", x.total)
	}

	for _, split := range x.splits {
		if split.Amount == 0 {
			// TODO(cjc25): Consider doing this in AddSplit and failing.
			return errors.New("Contains split with 0 amount")
		}
	}

	return nil
}

// Check that the transaction's splits are for valid accounts.
//
// The splits are valid if they are all for different accounts, and the
// accounts have all been created with NewAccount.
func (x *Transaction) ValidateAccounts() error {
	if len(x.splits) == 0 {
		return errors.New("No splits in transaction.")
	}

	seen := make(map[int64]bool)
	for _, split := range x.splits {
		if _, ok := x.accountMap[split.Account]; !ok {
			return fmt.Errorf("Nonexistant account %v", split.Account)
		}
		if seen[split.Account] {
			return errors.New("Multiple Splits for same Account.")
		}
		seen[split.Account] = true
	}

	return nil
}

// Commit the Splits in x to their respective Accounts, if x is Valid.
func (x *Transaction) Commit() error {
	if err := x.ValidateAmount(); err != nil {
		return err
	}
	if err := x.ValidateAccounts(); err != nil {
		return err
	}

	for _, split := range x.splits {
		x.accountMap[split.Account].total += split.Amount
	}

	return nil
}
