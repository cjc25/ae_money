package ae_money

import (
	"encoding/json"
	"net/http"

	"github.com/cjc25/ae_money/transaction"

	"appengine/datastore"
)

type DatastoreAccount struct {
	Account      *transaction.Account `json:"account"`
	DatastoreKey *datastore.Key       `json:"key"`
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
	for i := 0; i < len(keys); i++ {
		result[i].Account = &accounts[i]
		result[i].DatastoreKey = keys[i]
	}

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

	persisted := DatastoreAccount{&a, k}
	e := json.NewEncoder(w)
	if err = e.Encode(&persisted); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
