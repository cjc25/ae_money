package transaction

import (
	"errors"
	"strings"

	"appengine/datastore"
)

type Account struct {
	splits       []Split
	total        AmountType
	Name         string         `json:"name"`
	DatastoreKey *datastore.Key `json:"key" datastore"-"`
}

type ByName []*Account

// Implementation of sort.Interface
func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func (a *Account) Validate() error {
	a.Name = strings.TrimSpace(a.Name)

	if a.Name == "" {
		return errors.New("Empty account name.")
	}
	return nil
}
