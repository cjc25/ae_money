package ae_money

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.google.com/p/go-uuid/uuid"

	"github.com/cjc25/ae_money/transaction"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

// Setup method which adds required Accounts owned by User u to the test
// datastore.
func insertAccountsOrDie(t *testing.T, c appengine.Context, a []transaction.Account, u *user.User) []*datastore.Key {
	accountKeys := make([]*datastore.Key, len(a))
	for i := range a {
		accountKeys[i] = datastore.NewIncompleteKey(c, "Account", userKey(c, u))
	}

	k, err := datastore.PutMulti(c, accountKeys, a)
	if err != nil {
		t.Fatal(err)
	}

	return k
}

// Expectation function for a list of Accounts and the expected Keys.
func expectListAccountsResponse(t *testing.T, w *httptest.ResponseRecorder, k []*datastore.Key, a []transaction.Account) {
	if len(a) != len(k) {
		t.Fatalf("BAD TEST: Expected keys %v and accounts %v not same length", k, a)
	}

	var got []DatastoreAccount
	d := json.NewDecoder(w.Body)
	err := d.Decode(&got)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != len(a) {
		t.Fatalf("Expected %v accounts in response, got %v", len(a), len(got))
	}
	for i := range got {
		if got[i].IntID != k[i].IntID() {
			t.Errorf("Expected index %v id %v, got %v", i, k[i].IntID(), got[i].IntID)
		}
		if got[i].Account.Name != a[i].Name {
			t.Errorf("Expected index %v name %v, got %v", i, a[i].Name, got[i].Account.Name)
		}
	}
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

	a := []transaction.Account{{Name: "a1"}, {Name: "a2"}}
	k := insertAccountsOrDie(t, c, a, u)

	ListAccounts(&requestParams{w: w, r: r, c: c, u: u})

	expectListAccountsResponse(t, w, k, a)
}

func TestListAccounts_MultipleUsers(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	a := []transaction.Account{{Name: "a1"}}
	k := insertAccountsOrDie(t, c, a, u)
	insertAccountsOrDie(t, c, []transaction.Account{{Name: "a2"}}, &user.User{Email: "other@example.com"})

	ListAccounts(&requestParams{w: w, r: r, c: c, u: u})

	expectListAccountsResponse(t, w, k, a)
}

// Convenience function to add expected Splits to a specific Account.
func insertSplitsOrDie(t *testing.T, c appengine.Context, s []*transaction.Split, accountKey *datastore.Key) {
	splitKeys := make([]*datastore.Key, len(s))
	for i := range s {
		splitKeys[i] = datastore.NewKey(c, "Split", uuid.NewRandom().String(), 0, accountKey)
	}
	_, err := datastore.PutMulti(c, splitKeys, s)
	if err != nil {
		t.Fatal(err)
	}
}

func TestShowAccount_Success(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, _, c := initTestRequestParams(t, u)
	defer c.Close()

	a := []transaction.Account{{Name: "a1"}}
	k := insertAccountsOrDie(t, c, a, u)
	insertSplitsOrDie(t, c, []*transaction.Split{&transaction.Split{Amount: 123}}, k[0])
	v := map[string]string{"key": fmt.Sprint(k[0].IntID())}

	ShowAccount(&requestParams{w: w, c: c, u: u, v: v})

	expectCode(t, http.StatusOK, w)

	result := DatastoreAccountAndSplits{}
	d := json.NewDecoder(w.Body)
	if err := d.Decode(&result); err != nil {
		t.Fatal(err)
	}

	if result.IntID != k[0].IntID() {
		t.Errorf("Expected result id to be %v, got %v", k[0].IntID(), result.IntID)
	}
	if result.Account.Name != a[0].Name {
		t.Errorf("Expected result name to be %v, got %v", a[0].Name, result.Account.Name)
	}
	if len(result.Splits) != 1 {
		t.Fatalf("Expected 1 split in result. Got %v", len(result.Splits))
	}
	if result.Splits[0].Amount != 123 {
		t.Errorf("Expected split amount 123, got %v", result.Splits[0].Amount)
	}
}

func TestShowAccount_FailureNoSuchAccount(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, _, c := initTestRequestParams(t, u)
	defer c.Close()

	v := map[string]string{"key": "123456"}

	ShowAccount(&requestParams{w: w, c: c, u: u, v: v})

	expectCode(t, http.StatusNotFound, w)
	expectBody(t, "", w)
}

func TestShowAccount_FailureOtherUsersAccount(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, _, c := initTestRequestParams(t, u)
	defer c.Close()

	k := insertAccountsOrDie(t, c, []transaction.Account{{Name: "a1"}}, &user.User{Email: "other@example.com"})
	insertSplitsOrDie(t, c, []*transaction.Split{&transaction.Split{Amount: 123}}, k[0])
	v := map[string]string{"key": fmt.Sprint(k[0].IntID())}

	ShowAccount(&requestParams{w: w, c: c, u: u, v: v})

	expectCode(t, http.StatusNotFound, w)
	expectBody(t, "", w)
}

func TestNewAccount_Success(t *testing.T) {
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

	if k[0].IntID() != result.IntID {
		t.Errorf("Expected returned id '%v' to match queried id '%v'",
			result.IntID, k[0].IntID())
	}
}

// Expectation function for responses to failed NewAccount requests.
func expectBadNewAccountResponse(t *testing.T, c appengine.Context, u *user.User, w *httptest.ResponseRecorder) {
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

func TestNewAccount_FailureNoBody(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	r.Body = ioutil.NopCloser(bytes.NewBufferString(""))
	NewAccount(&requestParams{w: w, r: r, c: c, u: u})

	expectBadNewAccountResponse(t, c, u, w)
}

func TestNewAccount_FailureNoAccountName(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	r.Body = ioutil.NopCloser(bytes.NewBufferString(`{"name":"  "}`))
	NewAccount(&requestParams{w: w, r: r, c: c, u: u})

	expectBadNewAccountResponse(t, c, u, w)
}

// Expectation function for the number of Accounts a User has.
func expectNumAccounts(t *testing.T, c appengine.Context, u *user.User, expected_count int) {
	q := datastore.NewQuery("Account").Ancestor(userKey(c, u))
	actual_count, err := q.Count(c)
	if err != nil {
		t.Fatal(err)
	}
	if actual_count != expected_count {
		t.Errorf("Expected %v account(s), got %v", expected_count, actual_count)
	}
}

// The only failure that expects a bad request response, since it's perfectly
// fine to delete an account that doesn't exist: nothing happens.
func TestDeleteAccount_FailureSplitsInAccount(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, _, c := initTestRequestParams(t, u)
	defer c.Close()

	a := []transaction.Account{{Name: "a1"}}
	k := insertAccountsOrDie(t, c, a, u)[0]
	insertSplitsOrDie(t, c, []*transaction.Split{&transaction.Split{Amount: 123}}, k)

	v := map[string]string{"key": fmt.Sprint(k.IntID())}

	DeleteAccount(&requestParams{w: w, c: c, u: u, v: v})
	expectCode(t, http.StatusBadRequest, w)
	expectNumAccounts(t, c, u, 1)
}

func TestDeleteAccount_FailureNoSuchAccount(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, _, c := initTestRequestParams(t, u)
	defer c.Close()

	a := []transaction.Account{{Name: "a1"}}
	k := insertAccountsOrDie(t, c, a, u)[0]

	v := map[string]string{"key": fmt.Sprint(k.IntID() + 1)}

	DeleteAccount(&requestParams{w: w, c: c, u: u, v: v})
	expectCode(t, http.StatusOK, w)
	expectNumAccounts(t, c, u, 1)
}

// When deleting another User's account, report StatusOK as the "error code" to
// avoid exposing info.
func TestDeleteAccount_FailureOtherUser(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, _, c := initTestRequestParams(t, u)
	defer c.Close()

	a := []transaction.Account{{Name: "a1"}}
	k := insertAccountsOrDie(t, c, a, &user.User{Email: "other@example.com"})[0]

	v := map[string]string{"key": fmt.Sprint(k.IntID())}

	DeleteAccount(&requestParams{w: w, c: c, u: u, v: v})
	expectCode(t, http.StatusOK, w)
	expectNumAccounts(t, c, u, 0)
}

func TestDeleteAccount_Success(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, _, c := initTestRequestParams(t, u)
	defer c.Close()

	a := []transaction.Account{{Name: "a1"}}
	k := insertAccountsOrDie(t, c, a, u)[0]

	v := map[string]string{"key": fmt.Sprint(k.IntID())}

	DeleteAccount(&requestParams{w: w, c: c, u: u, v: v})
	expectCode(t, http.StatusOK, w)
	expectNumAccounts(t, c, u, 0)
}
