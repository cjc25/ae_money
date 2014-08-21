package transaction

import "testing"

func TestNewAccount_ExistingID(t *testing.T) {
	a, k := NewAccount("a1", 12345)

	if k != 12345 {
		t.Errorf("Expected key to be 12345, got %v", k)
	}

	if accountMap[k] != a {
		t.Errorf("Expected account to be %v, got %v", a, accountMap[k])
	}
}

func TestNewAccount_NewID(t *testing.T) {
	a, k := NewAccount("a1", 0)

	if k == 0 {
		t.Errorf("Expected nonzero key, got %v", k)
	}

	if accountMap[k] != a {
		t.Errorf("Expected account to be %v, got %v", a, accountMap[k])
	}
}

func TestNewAccount_UniqueIDs(t *testing.T) {
	a1, k1 := NewAccount("a1", 0)
	a2, k2 := NewAccount("a2", 0)

	if k1 == k2 {
		t.Errorf("Expected unique account keys, but both were %v", k1)
	}

	if accountMap[k1] != a1 {
		t.Errorf("Expected k1 to be %v, got %v", a1, accountMap[k1])
	}
	if accountMap[k2] != a2 {
		t.Errorf("Expected k2 to be %v, got %v", a2, accountMap[k2])
	}
}

func TestValidate_ValidNoChange(t *testing.T) {
	a := Account{Name: "valid"}

	if err := a.Validate(); err != nil {
		t.Errorf("Expected valid, got %v", err)
	}
	if a.Name != "valid" {
		t.Errorf("Expected name 'valid', got '%v'", a.Name)
	}
}

func TestValidate_ValidTrimWhitespace(t *testing.T) {
	a := Account{Name: "\ttrim whitespace   "}

	if err := a.Validate(); err != nil {
		t.Errorf("Expected valid, got %v", err)
	}
	if a.Name != "trim whitespace" {
		t.Errorf("Expected name 'trim whitespace', got '%v'", a.Name)
	}
}

func TestValidate_InvalidNoName(t *testing.T) {
	a := Account{Name: ""}

	err := a.Validate()
	if err == nil {
		t.Errorf("Expected invalid, got valid.")
	}
}
