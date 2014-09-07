// +build appengine

package transaction

import (
	"fmt"

	"appengine/datastore"
)

// Implement PropertyLoadSaver for transaction.Account to save the hidden field
// total.
func (a *Account) Load(c <-chan datastore.Property) error {
	err := error(nil)

	for p := range c {
		if p.Name == "Name" {
			a.Name = p.Value.(string)
		} else if p.Name == "Total" {
			a.total = AmountType(p.Value.(int64))
		} else {
			err = fmt.Errorf("Unexpected property type %v", p.Name)
		}
	}

	return err
}

// See Account.Load.
func (a *Account) Save(c chan<- datastore.Property) error {
	defer close(c)

	c <- datastore.Property{
		Name:  "Name",
		Value: a.Name,
	}
	c <- datastore.Property{
		Name:  "Total",
		Value: int64(a.total),
	}

	return nil
}
