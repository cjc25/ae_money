package ae_money

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

func contextWrapper(f func(http.ResponseWriter, *http.Request, appengine.Context)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		f(w, r, c)
	}
}

func loginWrapper(f func(http.ResponseWriter, *http.Request, appengine.Context, *user.User)) func(http.ResponseWriter, *http.Request, appengine.Context) {
	return func(w http.ResponseWriter, r *http.Request, c appengine.Context) {
		u := user.Current(c)
		if u == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		f(w, r, c, u)
	}
}

func userKey(c appengine.Context, u *user.User) *datastore.Key {
	return datastore.NewKey(c, "User", u.String(), 0, nil)
}

func init() {
	r := mux.NewRouter()

	api := r.PathPrefix("/api/v{version:[0-9]+}").Subrouter()
	api.HandleFunc("/accounts", contextWrapper(loginWrapper(ListAccounts))).
		Methods("GET")
	api.HandleFunc("/accounts/new", contextWrapper(loginWrapper(NewAccount))).
		Methods("POST")

	r.HandleFunc("/", root)
	http.Handle("/", r)
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ae_money")
}
