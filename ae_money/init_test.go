package ae_money

import (
	"fmt"
	"testing"

	"appengine/user"
)

var dummyLoginHandler = loginWrapper(func(p *requestParams) {
	w, u := p.w, p.u
	fmt.Fprint(w, u.String())
})

func TestLoginWrapper_LoggedOut(t *testing.T) {
	w, r, c := initTestRequestParams(t, nil)
	defer c.Close()

	dummyLoginHandler(&requestParams{w: w, r: r, c: c})

	expectCode(t, 401, w)
	expectBody(t, "", w)
}

func TestLoginWrapper_LoggedIn(t *testing.T) {
	w, r, c := initTestRequestParams(t, &user.User{Email: "test@example.com"})
	defer c.Close()

	dummyLoginHandler(&requestParams{w: w, r: r, c: c})

	expectCode(t, 200, w)
	expectBody(t, "test@example.com", w)
}
