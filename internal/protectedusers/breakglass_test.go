package protectedusers

import (
	"bytes"
	"strings"
	"testing"
)

func TestValidatePasswordPolicy(t *testing.T) {
	cases := []struct {
		pw string
		ok bool
	}{
		{"ValidPass123A", true},
		{"short1A", false},        // <12 chars
		{"alllowercase123", false}, // no upper
		{"ALLUPPERCASE123", false}, // no lower
		{"NoDigitsHereXYZ", false}, // no digit
		{"ABcd123456789", true},
	}
	for _, tc := range cases {
		err := validatePasswordPolicy(tc.pw)
		if (err == nil) != tc.ok {
			t.Errorf("validatePasswordPolicy(%q): err=%v, want ok=%v", tc.pw, err, tc.ok)
		}
	}
}

func TestReadPasswordTwice_Match(t *testing.T) {
	in := bytes.NewBufferString("ValidPass123A\nValidPass123A\n")
	out := &bytes.Buffer{}
	got, err := readPasswordTwice(in, out)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got != "ValidPass123A" {
		t.Errorf("got %q, want %q", got, "ValidPass123A")
	}
	if !strings.Contains(out.String(), "New password:") {
		t.Errorf("expected prompt; got %q", out.String())
	}
}

func TestReadPasswordTwice_Mismatch(t *testing.T) {
	in := bytes.NewBufferString("Foo123Bar456!\nFOO123Bar456!\n")
	_, err := readPasswordTwice(in, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "match") {
		t.Errorf("expected mismatch error, got %v", err)
	}
}

func TestParseBool(t *testing.T) {
	trueCases := []string{"true", "True", "TRUE", "1", "yes", "Yes", "y", "Y"}
	falseCases := []string{"false", "False", "0", "no", "No", "n", "N"}
	for _, in := range trueCases {
		got, err := parseBool(in)
		if err != nil {
			t.Errorf("parseBool(%q): %v", in, err)
		}
		if !got {
			t.Errorf("parseBool(%q) = false, want true", in)
		}
	}
	for _, in := range falseCases {
		got, err := parseBool(in)
		if err != nil {
			t.Errorf("parseBool(%q): %v", in, err)
		}
		if got {
			t.Errorf("parseBool(%q) = true, want false", in)
		}
	}
	if _, err := parseBool("maybe"); err == nil {
		t.Error("parseBool(\"maybe\") should error")
	}
}

func TestRequireIdentityConfirmation_ExplicitMatch(t *testing.T) {
	err := requireIdentityConfirmation(BreakGlassOptions{IdentityConfirmation: "BREAK-GLASS"})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestRequireIdentityConfirmation_ExplicitMismatch(t *testing.T) {
	err := requireIdentityConfirmation(BreakGlassOptions{IdentityConfirmation: "yes"})
	if err == nil {
		t.Error("expected mismatch error")
	}
}

func TestRequireIdentityConfirmation_StdinMatch(t *testing.T) {
	out := &bytes.Buffer{}
	err := requireIdentityConfirmation(BreakGlassOptions{
		Stdin:  bytes.NewBufferString("BREAK-GLASS\n"),
		Stdout: out,
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !strings.Contains(out.String(), "BREAK-GLASS") {
		t.Errorf("expected prompt to be written; got %q", out.String())
	}
}

func TestRequireIdentityConfirmation_StdinMismatch(t *testing.T) {
	err := requireIdentityConfirmation(BreakGlassOptions{
		Stdin:  bytes.NewBufferString("yes\n"),
		Stdout: &bytes.Buffer{},
	})
	if err == nil {
		t.Error("expected mismatch error")
	}
}

func TestAdminForceEditProtected_RequiresConfirm(t *testing.T) {
	err := AdminForceEditProtected(nil, "sqlite3", BreakGlassOptions{
		Email: "x@y.z",
		Field: "password",
	})
	if err == nil || !strings.Contains(err.Error(), "--confirm") {
		t.Errorf("expected --confirm error; got %v", err)
	}
}

func TestAdminForceEditProtected_RejectsNonProtectedEmail(t *testing.T) {
	err := AdminForceEditProtected(nil, "sqlite3", BreakGlassOptions{
		Email:   "random@example.com",
		Field:   "password",
		Confirm: true,
	})
	if err == nil || !strings.Contains(err.Error(), "not on the protected list") {
		t.Errorf("expected protected-list rejection; got %v", err)
	}
}

func TestRebindToPostgres(t *testing.T) {
	cases := map[string]string{
		"":                              "",
		"SELECT 1":                      "SELECT 1",
		"UPDATE x SET a = ? WHERE b = ?": "UPDATE x SET a = $1 WHERE b = $2",
	}
	for in, want := range cases {
		got := rebindToPostgres(in)
		if got != want {
			t.Errorf("rebindToPostgres(%q) = %q, want %q", in, got, want)
		}
	}
}
