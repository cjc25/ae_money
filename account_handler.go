package ae_money

import (
	"encoding/json"
	"net/http"

	"github.com/cjc25/ae_money/transaction"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

func ListAccounts(w http.ResponseWriter, r *http.Request, c appengine.Context, u *user.User) {
	q := datastore.NewQuery("Account").Ancestor(userKey(c, u)).Order("Name")
	// We make an empty slice so we can return [] if there are no accounts.
	accounts := make([]transaction.Account, 0)
	keys, err := q.GetAll(c, &accounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i := 0; i < len(keys); i++ {
		accounts[i].DatastoreKey = keys[i]
	}

	e := json.NewEncoder(w)
	err = e.Encode(accounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func NewAccount(w http.ResponseWriter, r *http.Request, c appengine.Context, u *user.User) {
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

	a.DatastoreKey = k
	e := json.NewEncoder(w)
	if err = e.Encode(&a); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
