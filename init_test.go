package ae_money

import (
	"fmt"
	"net/http"
	"testing"

	"appengine"
	"appengine/user"
)

var dummyLoginHandler = loginWrapper(func(w http.ResponseWriter, r *http.Request, c appengine.Context, u *user.User) {
	fmt.Fprint(w, u.String())
})

func TestLoginWrapper_LoggedOut(t *testing.T) {
	w, r, c := initTestRequestParams(t, nil)
	defer c.Close()

	dummyLoginHandler(w, r, c)

	expectCode(t, 401, w)
	expectBody(t, "", w)
}

func TestLoginWrapper_LoggedIn(t *testing.T) {
	w, r, c := initTestRequestParams(t, &user.User{Email: "test@example.com"})
	defer c.Close()

	dummyLoginHandler(w, r, c)

	expectCode(t, 200, w)
	expectBody(t, "test@example.com", w)
}
