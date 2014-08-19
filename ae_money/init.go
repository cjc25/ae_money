package ae_money

import (
	"net/http"

	"github.com/gorilla/mux"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

type requestParams struct {
	w http.ResponseWriter
	r *http.Request
	c appengine.Context
	u *user.User
	v map[string]string
}

func baseWrapper(f func(*requestParams)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f(&requestParams{w: w, r: r, c: appengine.NewContext(r), v: mux.Vars(r)})
	}
}

func loginWrapper(f func(*requestParams)) func(*requestParams) {
	return func(p *requestParams) {
		p.u = user.Current(p.c)
		if p.u == nil {
			p.w.WriteHeader(http.StatusUnauthorized)
			return
		}

		f(p)
	}
}

func userKey(c appengine.Context, u *user.User) *datastore.Key {
	return datastore.NewKey(c, "User", u.String(), 0, nil)
}

func init() {
	r := mux.NewRouter()

	api := r.PathPrefix("/api/v{version:[0-9]+}").Subrouter()
	api.HandleFunc("/accounts/new", baseWrapper(loginWrapper(NewAccount))).
		Methods("POST")
	api.HandleFunc("/accounts", baseWrapper(loginWrapper(ShowAccount))).
		Queries("key", "{key:[0-9]+}").
		Methods("GET")
	api.HandleFunc("/accounts", baseWrapper(loginWrapper(ListAccounts))).
		Methods("GET")

	http.Handle("/", r)
}
