package ae_money

import (
	"encoding/json"
	"net/http"

	"code.google.com/p/go-uuid/uuid"

	"github.com/cjc25/ae_money/transaction"

	"appengine"
	"appengine/datastore"
)

type TransactionRequest struct {
	Amounts  []transaction.AmountType `json:"amounts"`
	Accounts []int64                  `json:"accounts"`
	Memo     string                   `json:"memo"`
}

func NewTransaction(p *requestParams) {
	w, r, c, u := p.w, p.r, p.c, p.u

	d := json.NewDecoder(r.Body)
	var request TransactionRequest
	if err := d.Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(request.Amounts) != len(request.Accounts) {
		http.Error(w, "Amounts and accounts of different lengths", http.StatusBadRequest)
		return
	}

	userKey := userKey(c, u)
	transactionId := uuid.NewRandom().String()
	accountKeys := make([]*datastore.Key, len(request.Accounts))
	splitKeys := make([]*datastore.Key, len(request.Accounts))
	splits := make([]*transaction.Split, len(request.Accounts))

	for i := range request.Accounts {
		accountKeys[i] = datastore.NewKey(c, "Account", "", request.Accounts[i], userKey)
		splitKeys[i] = datastore.NewKey(c, "Split", transactionId, 0, accountKeys[i])
		splits[i] = &transaction.Split{request.Amounts[i], request.Accounts[i], request.Memo}
	}

	x := transaction.NewTransaction()
	x.AddSplits(splits)

	if err := x.ValidateAmount(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		accounts := make([]transaction.Account, len(accountKeys))
		if err := datastore.GetMulti(c, accountKeys, accounts); err != nil {
			return err
		}
		for i, account := range accounts {
			x.AddAccount(&account, accountKeys[i].IntID())
		}

		if err := x.Commit(); err != nil {
			return err
		}

		putStatus := make(chan error)

		go func() {
			_, err := datastore.PutMulti(c, accountKeys, accounts)
			putStatus <- err
		}()
		go func() {
			_, err := datastore.PutMulti(c, splitKeys, splits)
			putStatus <- err
		}()

		err := <-putStatus
		if err != nil {
			return err
		}
		return <-putStatus
	}, nil)
	if err != nil {
		// TODO(cjc25): This might not be a 400: if e.g. datastore failed it should
		// be a 500. Interpret err and return the right thing.
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
