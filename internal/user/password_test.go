package user

import "testing"

func TestArgon2PasswordRoundTrip(t *testing.T) {
	user := &User{}
	if err := user.SetPassword("pa55word"); err != nil {
		t.Fatal(err)
	}
	if user.PasswordHash == "" || user.PasswordHash[:10] != "$argon2id$" {
		t.Fatalf("hash = %q", user.PasswordHash)
	}

	match, err := user.PasswordMatches("pa55word")
	if err != nil {
		t.Fatal(err)
	}
	if !match {
		t.Fatal("expected the password to match")
	}

	match, err = user.PasswordMatches("wrong-password")
	if err != nil {
		t.Fatal(err)
	}
	if match {
		t.Fatal("expected a mismatched password to fail")
	}
}

func TestPasswordMatchesRejectsMalformedHash(t *testing.T) {
	user := &User{PasswordHash: "not-a-hash"}
	_, err := user.PasswordMatches("pa55word")
	if err == nil {
		t.Fatal("expected a malformed hash to fail")
	}
}
