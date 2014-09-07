package transaction

import "testing"

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

func TestMarshalJSON(t *testing.T) {
	a := Account{Name: "myname", total: 12345}

	json, err := a.MarshalJSON()
	if err != nil {
		t.Error(err)
	}

	expected := `{"name":"myname","total":12345}`
	got := string(json)

	if got != expected {
		t.Errorf("Expected JSON string %v but got %v", expected, got)
	}
}
