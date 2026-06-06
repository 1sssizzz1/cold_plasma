package service

import "testing"

func TestNormalizeRussianPhone(t *testing.T) {
	cases := map[string]string{
		"+7 900 123-45-67":  "+79001234567",
		"8 (900) 123-45-67": "+79001234567",
		"9001234567":        "+79001234567",
	}
	for in, want := range cases {
		got, err := NormalizeRussianPhone(in)
		if err != nil {
			t.Fatalf("NormalizeRussianPhone(%q) returned error: %v", in, err)
		}
		if got != want {
			t.Fatalf("NormalizeRussianPhone(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeRussianPhoneRejectsNonRussianNumbers(t *testing.T) {
	for _, in := range []string{"+1 202 555-0199", "+375291234567", "+7 812 123-45-67", "123"} {
		if got, err := NormalizeRussianPhone(in); err == nil {
			t.Fatalf("NormalizeRussianPhone(%q) = %q, want error", in, got)
		}
	}
}
