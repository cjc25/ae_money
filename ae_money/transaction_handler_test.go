package ae_money

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/cjc25/ae_money/transaction"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

// Convenience function to wrap the TransactionRequest in a format that an HTTP
// Handler expects.
func buildTestTransactionRequest(t *testing.T, amounts []transaction.AmountType, accounts []int64, memo, date string) io.ReadCloser {
	request := &TransactionRequest{Amounts: amounts, Accounts: accounts, Memo: memo, Date: date}
	b := bytes.Buffer{}
	e := json.NewEncoder(&b)
	if err := e.Encode(request); err != nil {
		t.Fatal(err)
	}
	return ioutil.NopCloser(&b)
}

func expectSplits(t *testing.T, c appengine.Context, u *user.User, accountKeys []*datastore.Key, expected []transaction.AmountType, memo string) {
	if len(accountKeys) != len(expected) {
		t.Fatalf("Can't check splits: %v expected account keys != %v expected splits.",
			len(accountKeys), len(expected))
	}

	q := datastore.NewQuery("Split").Ancestor(userKey(c, u)).Order("Amount")
	got := make([]transaction.Split, 0)
	xKeys, err := q.GetAll(c, &got)
	if err != nil {
		t.Fatal(err)
	}

	if len(xKeys) != len(accountKeys) {
		t.Fatalf("Expected to get %v splits, but got %v", len(accountKeys), len(xKeys))
	}

	for i := range accountKeys {
		if got[i].Amount != expected[i] {
			t.Errorf("Expected split %v to have amount %v but got %v",
				i, expected[i], got[i].Amount)
		}
		if got[i].Memo != memo {
			t.Errorf("Expected split %v to have memo %v but got %v",
				i, memo, got[i].Memo)
		}

		if xKeys[i].Parent().Encode() != accountKeys[i].Encode() {
			t.Errorf("Expected split %v to have account %v but got %v",
				i, accountKeys[i].Encode(), xKeys[i].Parent().Encode())
		}
		if xKeys[i].StringID() != xKeys[0].StringID() {
			t.Errorf("Expected all splits to have the same ID, but split 0 had %v and split %v had %v",
				xKeys[0].StringID(), i, xKeys[i].StringID())
		}
	}
}

func TestTransactionSuccess(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	accountKeys := insertAccountsOrDie(t, c,
		[]transaction.Account{{Name: "a1"}, {Name: "a2"}}, u)
	r.Body = buildTestTransactionRequest(t,
		[]transaction.AmountType{-123, 123},
		[]int64{accountKeys[0].IntID(), accountKeys[1].IntID()},
		"Test transaction",
		"2014-11-01",
	)

	NewTransaction(&requestParams{w: w, r: r, c: c, u: u})

	expectCode(t, http.StatusOK, w)
	expectBody(t, "", w)
	expectSplits(t, c, u, accountKeys, []transaction.AmountType{-123, 123}, "Test transaction")
}

func TestTransactionNoDate(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	accountKeys := insertAccountsOrDie(t, c,
		[]transaction.Account{{Name: "a1"}, {Name: "a2"}}, u)
	r.Body = buildTestTransactionRequest(t,
		[]transaction.AmountType{-123, 123},
		[]int64{accountKeys[0].IntID(), accountKeys[1].IntID()},
		"Test transaction",
		"",
	)

	NewTransaction(&requestParams{w: w, r: r, c: c, u: u})

	expectCode(t, http.StatusBadRequest, w)
	expectSplits(t, c, u, nil, nil, "")
}

func TestTransactionBadDate(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	accountKeys := insertAccountsOrDie(t, c,
		[]transaction.Account{{Name: "a1"}, {Name: "a2"}}, u)
	r.Body = buildTestTransactionRequest(t,
		[]transaction.AmountType{-123, 123},
		[]int64{accountKeys[0].IntID(), accountKeys[1].IntID()},
		"Test transaction",
		"Not a real date",
	)

	NewTransaction(&requestParams{w: w, r: r, c: c, u: u})

	expectCode(t, http.StatusBadRequest, w)
	expectSplits(t, c, u, nil, nil, "")
}

func TestTransactionDifferentLengths(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	accountKeys := insertAccountsOrDie(t, c,
		[]transaction.Account{{Name: "a1"}, {Name: "a2"}}, u)
	r.Body = buildTestTransactionRequest(t,
		[]transaction.AmountType{123, -100, -23},
		[]int64{accountKeys[0].IntID(), accountKeys[1].IntID()},
		"Bad transaction",
		"2014-11-01",
	)

	NewTransaction(&requestParams{w: w, r: r, c: c, u: u})

	expectCode(t, http.StatusBadRequest, w)
	expectSplits(t, c, u, nil, nil, "")
}

func TestTransactionNoSplits(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	r.Body = ioutil.NopCloser(bytes.NewBufferString(`{}`))

	NewTransaction(&requestParams{w: w, r: r, c: c, u: u})

	expectCode(t, http.StatusBadRequest, w)
	expectSplits(t, c, u, nil, nil, "")
}

func TestTransactionNonZero(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	accountKeys := insertAccountsOrDie(t, c,
		[]transaction.Account{{Name: "a1"}, {Name: "a2"}}, u)
	r.Body = buildTestTransactionRequest(t,
		[]transaction.AmountType{123, 456},
		[]int64{accountKeys[0].IntID(), accountKeys[1].IntID()},
		"Bad transaction",
		"2014-11-01",
	)

	NewTransaction(&requestParams{w: w, r: r, c: c, u: u})

	expectCode(t, http.StatusBadRequest, w)
	expectSplits(t, c, u, nil, nil, "")
}

func TestTransactionDuplicateAccount(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	accountKeys := insertAccountsOrDie(t, c,
		[]transaction.Account{{Name: "a1"}}, u)
	r.Body = buildTestTransactionRequest(t,
		[]transaction.AmountType{-123, 123},
		[]int64{accountKeys[0].IntID(), accountKeys[0].IntID()},
		"Bad transaction",
		"2014-11-01",
	)

	NewTransaction(&requestParams{w: w, r: r, c: c, u: u})

	expectCode(t, http.StatusBadRequest, w)
	expectSplits(t, c, u, nil, nil, "")
}

func TestTransactionOtherUserAccount(t *testing.T) {
	u := &user.User{Email: "test@example.com"}
	w, r, c := initTestRequestParams(t, u)
	defer c.Close()

	accountKeys := insertAccountsOrDie(t, c, []transaction.Account{{Name: "a1"}}, u)
	accountKeys = append(accountKeys, insertAccountsOrDie(
		t, c, []transaction.Account{{Name: "a2"}}, &user.User{Email: "other@example.com"})...)
	r.Body = buildTestTransactionRequest(t,
		[]transaction.AmountType{-123, 123},
		[]int64{accountKeys[0].IntID(), accountKeys[1].IntID()},
		"Bad transaction",
		"2014-11-01",
	)

	NewTransaction(&requestParams{w: w, r: r, c: c, u: u})

	expectCode(t, http.StatusBadRequest, w)
	expectSplits(t, c, u, nil, nil, "")
}
