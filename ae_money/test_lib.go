package ae_money

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"appengine/aetest"
	"appengine/user"
)

// initTestRequestParams does common setup for request, response, and appengine
// contexts.
func initTestRequestParams(t *testing.T, u *user.User) (w *httptest.ResponseRecorder, r *http.Request, c aetest.Context) {
	w = httptest.NewRecorder()

	r, err := http.NewRequest("GET", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	c, err = aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}

	if u != nil {
		c.Login(u)
	}

	return
}

// Expectation function for HTTP response codes.
func expectCode(t *testing.T, expected int, w *httptest.ResponseRecorder) {
	if expected != w.Code {
		t.Errorf("Expected code %v, got %v", expected, w.Code)
	}
}

// Expectation function for HTTP response bodies.
func expectBody(t *testing.T, expected string, w *httptest.ResponseRecorder) {
	got := strings.TrimSpace(w.Body.String())
	if expected != got {
		t.Errorf("Expected body \"%v\", got \"%v\"", expected, got)
	}
}
