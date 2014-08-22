package ae_money

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cjc25/ae_money/transaction"

	"appengine/datastore"
)

type DatastoreAccount struct {
	Account *transaction.Account `json:"account"`
	IntID   int64                `json:"key"`
}

type DatastoreAccountAndSplits struct {
	DatastoreAccount
	Splits []transaction.Split `json:"splits"`
}

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
	accountKey := datastore.NewKey(c, "Account", "", accountIntID, userKey(c, u))
	var a transaction.Account
	err = datastore.Get(c, accountKey, &a)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	q := datastore.NewQuery("Split").Ancestor(accountKey)
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
