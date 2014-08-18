package ae_money

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/cjc25/ae_money/transaction"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

func insertOrDie(t *testing.T, c appengine.Context, a []DatastoreAccount) {
	putKeys := make([]*datastore.Key, len(a))
	putAccounts := make([]*transaction.Account, len(a))
	for i, value := range a {
		putKeys[i] = value.DatastoreKey
		putAccounts[i] = value.Account
	}

	keys, err := datastore.PutMulti(c, putKeys, putAccounts)
	if err != nil {
		t.Fatal(err)
	}

	// Update the DatastoreAccount keys, in case any were incomplete.
	for i := 0; i < len(keys); i++ {
		a[i].DatastoreKey = keys[i]
	}
}

func decodeListResponse(t *testing.T, w *httptest.ResponseRecorder) []DatastoreAccount {
	var got []DatastoreAccount
	d := json.NewDecoder(w.Body)
	err := d.Decode(&got)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestListAccounts_Empty(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	ListAccounts(&requestParams{w: w, r: r, c: c, u: u})
	expectBody(t, "[]", w)
}

func TestListAccounts_OneUser(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	userKey := datastore.NewKey(c, "User", u.String(), 0, nil)

	a := []DatastoreAccount{
		{&transaction.Account{Name: "a1"}, datastore.NewIncompleteKey(c, "Account", userKey)},
		{&transaction.Account{Name: "a2"}, datastore.NewIncompleteKey(c, "Account", userKey)},
	}
	insertOrDie(t, c, a)

	ListAccounts(&requestParams{w: w, r: r, c: c, u: u})

	got := decodeListResponse(t, w)
	if !reflect.DeepEqual(a, got) {
		t.Errorf("Expected %v, got %v", a, got)
	}
}

func TestListAccounts_MultipleUsers(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	firstUserKey := userKey(c, u)
	otherUserKey := userKey(c, &user.User{Email: "other@example.com"})

	a := []DatastoreAccount{
		{&transaction.Account{Name: "a1"}, datastore.NewIncompleteKey(c, "Account", firstUserKey)},
		{&transaction.Account{Name: "a2"}, datastore.NewIncompleteKey(c, "Account", otherUserKey)},
	}
	insertOrDie(t, c, a)

	ListAccounts(&requestParams{w: w, r: r, c: c, u: u})
	got := decodeListResponse(t, w)

	if !reflect.DeepEqual(a[:1], got) {
		t.Errorf("Expected %v, got %v", a[:1], got)
	}
}

func TestAddAccount_Success(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	r.Body = ioutil.NopCloser(bytes.NewBufferString(`{"name":"a1"}`))

	NewAccount(&requestParams{w: w, r: r, c: c, u: u})

	q := datastore.NewQuery("Account").Ancestor(userKey(c, u))
	var accounts []transaction.Account
	k, err := q.GetAll(c, &accounts)
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatalf("Expected 1 account, got %v", len(accounts))
	}
	if accounts[0].Name != "a1" {
		t.Errorf("Expected account name 'a1', got '%v'", accounts[0].Name)
	}

	// Check that the keys are the same by checking the strings are the same
	result := DatastoreAccount{}
	d := json.NewDecoder(w.Body)
	err = d.Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if k[0].Encode() != result.DatastoreKey.Encode() {
		t.Errorf("Expected returned key '%v' to match queried key '%v'",
			result.DatastoreKey.Encode(), k[0].Encode())
	}
}

func expectBadAddAccountResponse(t *testing.T, c appengine.Context, u *user.User, w *httptest.ResponseRecorder) {
	q := datastore.NewQuery("Account").Ancestor(userKey(c, u)).KeysOnly()
	keys, err := q.GetAll(c, make([]transaction.Account, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 0 {
		t.Errorf("Expected no accounts to be inserted, but saw %v", len(keys))
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 status. Got %v", w.Code)
	}
}

func TestAddAccount_FailureNoBody(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	r.Body = ioutil.NopCloser(bytes.NewBufferString(""))
	NewAccount(&requestParams{w: w, r: r, c: c, u: u})

	expectBadAddAccountResponse(t, c, u, w)
}

func TestAddAccount_FailureNoAccountName(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	r.Body = ioutil.NopCloser(bytes.NewBufferString(`{"name":"  "}`))
	NewAccount(&requestParams{w: w, r: r, c: c, u: u})

	expectBadAddAccountResponse(t, c, u, w)
}
