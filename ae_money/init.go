package ae_money

import (
	"fmt"
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
		p := requestParams{w: w, r: r}
		p.c = appengine.NewContext(r)
		p.v = mux.Vars(r)

		f(&p)
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
	api.HandleFunc("/accounts", baseWrapper(loginWrapper(ListAccounts))).
		Methods("GET")
	api.HandleFunc("/accounts/new", baseWrapper(loginWrapper(NewAccount))).
		Methods("POST")

	r.HandleFunc("/", root)
	http.Handle("/", r)
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ae_money")
}
