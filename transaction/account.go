package transaction

import (
	"errors"
	"strings"
)

// An Account can receive Splits in a Transaction.
//
// Note that Accounts are more general than something like a real-life account
// at a bank. They can represent any category of income or expense, like
// "Salary" or "Rent."
type Account struct {
	total AmountType
	Name  string `json:"name"`
}

// Make sure an Account has valid fields. Useful if it was created with
// user-provided data.
func (a *Account) Validate() error {
	a.Name = strings.TrimSpace(a.Name)

	if a.Name == "" {
		return errors.New("Empty account name.")
	}
	return nil
}
