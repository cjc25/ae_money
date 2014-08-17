package transaction

import (
	"errors"
	"fmt"

	"code.google.com/p/go-uuid/uuid"
)

type AmountType int

// TransactionID is a string because datastore can't handle arbitrary byte
// strings.
type TransactionID string

type Split struct {
	Amount        AmountType
	account       *Account
	transactionID TransactionID
}

func NewSplit(amount AmountType, account *Account) *Split {
	return &Split{amount, account, ""}
}

func (s *Split) Account() *Account {
	return s.account
}

func (s *Split) TransactionID() TransactionID {
	return s.transactionID
}

type Transaction struct {
	splits []Split
	total  AmountType
}

func (x *Transaction) AddSplits(splits []*Split) {
	for _, split := range splits {
		x.AddSplit(split)
	}
}

func (x *Transaction) AddSplit(split *Split) {
	x.splits = append(x.splits, *split)
	x.total += split.Amount
}

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
