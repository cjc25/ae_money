package ae_money

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cjc25/ae_money/transaction"

	"appengine"
	"appengine/datastore"
)

// DatastoreAccount wraps transaction.Account for JSON responses that include a
// datastore key.
type DatastoreAccount struct {
	Account *transaction.Account `json:"account"`
	IntID   int64                `json:"key"`
}

// DatastoreAccountAndSplits wraps a DatastoreAccount and a slice of
// transaction.Split for JSON responses.
type DatastoreAccountAndSplits struct {
	DatastoreAccount
	Splits []transaction.Split `json:"splits"`
}

// ListAccounts gets the logged in user's accounts from datastore.
func ListAccounts(p *requestParams) {
	// Unwrap requestParams for easy access.
	w, c, u := p.w, p.c, p.u

	q := datastore.NewQuery("Account").Ancestor(userKey(c, u)).Order("Name")
	// We make an empty slice so we can return [] if there are no accounts.
	accounts := make([]transaction.Account, 0)
	keys, err := q.GetAll(c, &accounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]DatastoreAccount, len(accounts))
	for i := range keys {
		result[i].Account = &accounts[i]
		result[i].IntID = keys[i].IntID()
	}

	e := json.NewEncoder(w)
	err = e.Encode(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ShowAccount prints a specific Account's details, including Splits. The
// Account to print is extracted from the gorilla/mux vars.
func ShowAccount(p *requestParams) {
	w, c, u, v := p.w, p.c, p.u, p.v

	var accountIntID int64
	_, err := fmt.Sscan(v["key"], &accountIntID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// We get the account, then query splits separately: Accounts can't be
	// updated yet, but even if they could they're just names. Consistency is
	// unimportant.
	// TODO(cjc25): Is the above still true when Accounts include totals?
	accountKey := datastore.NewKey(c, "Account", "", accountIntID, userKey(c, u))
	var a transaction.Account
	err = datastore.Get(c, accountKey, &a)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	q := datastore.NewQuery("Split").Ancestor(accountKey).Order("Date").Order("-Amount")
	// We make an empty slice so we can return [] if there are no splits.
	splits := make([]transaction.Split, 0)
	_, err = q.GetAll(c, &splits)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := &DatastoreAccountAndSplits{DatastoreAccount{&a, accountKey.IntID()}, splits}
	e := json.NewEncoder(w)
	err = e.Encode(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// NewAccount creates a new Account. The Account is read as JSON from the
// request body.
func NewAccount(p *requestParams) {
	// Unwrap requestParams for easy access.
	w, r, c, u := p.w, p.r, p.c, p.u

	// We specifically want to decode into a transaction.Account so we don't pick
	// up a key.
	var a transaction.Account
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&a); err != nil {
		// TODO(cjc25): Could this reveal too much?
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := a.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accountKey := datastore.NewIncompleteKey(c, "Account", userKey(c, u))
	k, err := datastore.Put(c, accountKey, &a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	persisted := DatastoreAccount{&a, k.IntID()}
	e := json.NewEncoder(w)
	if err = e.Encode(&persisted); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// DeleteAccount deletes an Account owned by the logged in user. The Account to
// delete is extracted from the gorilla/mux vars.
func DeleteAccount(p *requestParams) {
	w, c, u, v := p.w, p.c, p.u, p.v

	var accountIntID int64
	_, err := fmt.Sscan(v["key"], &accountIntID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	accountKey := datastore.NewKey(c, "Account", "", accountIntID, userKey(c, u))
	splitsQuery := datastore.NewQuery("Split").Ancestor(accountKey)

	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		count, err := splitsQuery.Count(c)
		if err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("Can't delete an account which still has %v splits", count)
		}

		return datastore.Delete(c, accountKey)
	}, nil)
	if err != nil {
		// TODO(cjc25): This might not be a 400: if e.g. datastore failed it should
		// be a 500. Interpret err and return the right thing.
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
