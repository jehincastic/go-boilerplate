package validator

import "testing"

func TestValidatorAddErrorKeepsFirstMessage(t *testing.T) {
	v := New()
	v.AddError("email", "must be provided")
	v.AddError("email", "must be a valid email address")

	if v.Errors["email"] != "must be provided" {
		t.Fatalf("email error = %q", v.Errors["email"])
	}
	if v.Valid() {
		t.Fatal("expected validator to be invalid")
	}
}

func TestUnique(t *testing.T) {
	if !Unique([]string{"action", "comedy"}) {
		t.Fatal("expected unique values")
	}
	if Unique([]string{"action", "action"}) {
		t.Fatal("expected duplicate values to fail")
	}
}

func TestPermittedValue(t *testing.T) {
	if !PermittedValue("id", "id", "title", "-id") {
		t.Fatal("expected id to be permitted")
	}
	if PermittedValue("drop", "id", "title") {
		t.Fatal("expected drop to be rejected")
	}
}

func TestEmailRX(t *testing.T) {
	if !Matches("ada@example.com", EmailRX) {
		t.Fatal("expected a valid email")
	}
	if Matches("not-an-email", EmailRX) {
		t.Fatal("expected an invalid email")
	}
}
