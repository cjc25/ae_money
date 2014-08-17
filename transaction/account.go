package transaction

import (
	"errors"
	"strings"
)

type Account struct {
	splits []Split
	total  AmountType
	Name   string `json:"name"`
}

func (a *Account) Validate() error {
	a.Name = strings.TrimSpace(a.Name)

	if a.Name == "" {
		return errors.New("Empty account name.")
	}
	return nil
}
