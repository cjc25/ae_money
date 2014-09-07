// +build appengine

package transaction

import (
	"testing"

	"appengine/datastore"
)

func TestAccountSaveAndLoad(t *testing.T) {
	saved := &Account{Name: "myname", total: 12345}

	propChan := make(chan datastore.Property)
	go func() {
		err := saved.Save(propChan)
		if err != nil {
			t.Errorf("Failed to save %v: %v", saved, err)
		}
	}()

	loaded := &Account{}
	err := loaded.Load(propChan)
	if err != nil {
		t.Errorf("Failed to load into %v: %v", loaded, err)
	}

	if *loaded != *saved {
		t.Errorf("Loaded value %v was not the same as saved value %v", loaded, saved)
	}
}
