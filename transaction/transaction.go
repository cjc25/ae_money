package transaction

import (
	"errors"
	"fmt"
)

type AmountType int

type Split struct {
	Amount  AmountType
	Account *Account
}

type Transaction struct {
	splits []Split
	total  AmountType
}

func NewTransaction() *Transaction {
	return &Transaction{}
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
		if split.Account == nil {
			return errors.New("Split with nil Account.")
		}
		if seen[split.Account] {
			return errors.New("Multiple Splits for same Account.")
		}
		seen[split.Account] = true
	}

	return nil
}

func (x *Transaction) Commit() error {
	if err := x.Valid(); err != nil {
		return err
	}

	for _, split := range x.splits {
		split.Account.splits = append(split.Account.splits, split)
		split.Account.total += split.Amount
	}

	x.clear()
	return nil
}

func (x *Transaction) clear() {
	x.total = 0
	x.splits = nil
}
