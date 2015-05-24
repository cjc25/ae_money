package ae_money

import (
	"fmt"
	"net/http"
	"testing"

	"appengine/user"
)

// dummyLoginHandler is wrapped by loginWrapper, and prints the logged in user
// name to the ResponseWriter.
var dummyLoginHandler = loginWrapper(func(p *requestParams) {
	w, u := p.w, p.u
	fmt.Fprint(w, u.String())
})

// Verify that a logged out user's request returns an unauthorized error.
func TestLoginWrapper_LoggedOut(t *testing.T) {
	w, r, c := initTestRequestParams(t, nil)
	defer c.Close()

	dummyLoginHandler(&requestParams{w: w, r: r, c: c})

	expectCode(t, http.StatusUnauthorized, w)
	expectBody(t, "", w)
}

// Verify that a logged in user's request succeeds.
func TestLoginWrapper_LoggedIn(t *testing.T) {
	w, r, c := initTestRequestParams(t, &user.User{Email: "test@example.com"})
	defer c.Close()

	dummyLoginHandler(&requestParams{w: w, r: r, c: c})

	expectCode(t, http.StatusOK, w)
	expectBody(t, "test@example.com", w)
}
