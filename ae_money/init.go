// Package ae_money is an extremely basic appengine application for managing
// personal finances.
package ae_money

import (
	"net/http"

	"github.com/gorilla/mux"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

// userKey provides the datastore key for a user.
func userKey(c appengine.Context, u *user.User) *datastore.Key {
	return datastore.NewKey(c, "User", u.String(), 0, nil)
}

// requestParams is used to pass gorilla/mux and appengine components to a
// request handler, allowing simpler wrapping of functions to provide common
// functionality.
type requestParams struct {
	w http.ResponseWriter
	r *http.Request
	c appengine.Context
	u *user.User
	v map[string]string
}

// baseWrapper is used to convert a normal golang mux http handler function
// into a wrappable one whose argument is a requestParams. It also extracts the
// appengine context and gorilla/mux variables.
func baseWrapper(f func(*requestParams)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f(&requestParams{w: w, r: r, c: appengine.NewContext(r), v: mux.Vars(r)})
	}
}

// loginWrapper is a requestParams function wrapper which enforces that an
// appengine request is for a logged in user, and extracts that user into the
// requestParams.
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

// Register handlers and get ready to serve.
func init() {
	r := mux.NewRouter()

	api := r.PathPrefix("/api/v{version:[0-9]+}").Subrouter()

	api.HandleFunc("/accounts/new", baseWrapper(loginWrapper(NewAccount))).
		Methods("POST")
	api.HandleFunc("/accounts/{key:[0-9]+}", baseWrapper(loginWrapper(ShowAccount))).
		Methods("GET")
	api.HandleFunc("/accounts/{key:[0-9]+}", baseWrapper(loginWrapper(DeleteAccount))).
		Methods("DELETE")
	api.HandleFunc("/accounts", baseWrapper(loginWrapper(ListAccounts))).
		Methods("GET")

	api.HandleFunc("/transactions/new", baseWrapper(loginWrapper(NewTransaction))).
		Methods("POST")

	http.Handle("/", r)
}
