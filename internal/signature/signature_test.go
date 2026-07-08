package signature

import "testing"

func TestBase64URL(t *testing.T) {
	svc := New("md5")
	got := svc.Base64URL([]byte("hello"))
	if got != "aGVsbG8" {
		t.Fatalf("unexpected base64url: %s", got)
	}
}

func TestJWTSignMD5(t *testing.T) {
	svc := New("md5")
	got := svc.JWTSignMD5("header.payload", "merchant", "password1")
	if got == "" {
		t.Fatal("expected non-empty signature")
	}
}

func TestCreatePaymentSignature(t *testing.T) {
	svc := New("md5")
	got, err := svc.CreatePaymentSignature(map[string]string{
		"OutSum":   "99.90",
		"InvId":    "1001",
		"Shp_item": "digital",
	}, "demo", "secret1", "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "6f3ff72da89c99859c09fac8b1b61cb5" {
		t.Fatalf("unexpected signature: %s", got)
	}
}

func TestCreatePaymentSignatureAlgorithms(t *testing.T) {
	svc := New("md5")
	tests := []struct {
		algo string
		size int
	}{
		{"md5", 32},
		{"sha256", 64},
		{"sha512", 128},
	}

	for _, tt := range tests {
		t.Run(tt.algo, func(t *testing.T) {
			got, err := svc.CreatePaymentSignature(map[string]string{
				"OutSum": "99.90",
				"InvId":  "1001",
			}, "demo", "secret1", tt.algo)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.size {
				t.Fatalf("unexpected signature size for %s: %d", tt.algo, len(got))
			}
		})
	}
}

func TestCreatePaymentSignatureFallsBackToMD5ForUnknownAlgo(t *testing.T) {
	svc := New("md5")
	params := map[string]string{
		"OutSum": "99.90",
		"InvId":  "1001",
	}
	got, err := svc.CreatePaymentSignature(params, "demo", "secret1", "unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want, err := svc.CreatePaymentSignature(params, "demo", "secret1", "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("expected unknown algorithm to fall back to md5")
	}
}

func TestCreatePaymentSignatureWithModifiers(t *testing.T) {
	svc := New("md5")
	got, err := svc.CreatePaymentSignatureFromParams(PaymentSignatureParams{
		Login:             "demo",
		OutSum:            "8.96",
		InvID:             "12345",
		Receipt:           "%7B%22items%22%3A%5B%7B%22name%22%3A%22product%22%2C%22quantity%22%3A1%2C%22sum%22%3A8.96%2C%22tax%22%3A%22none%22%7D%5D%7D",
		SuccessURL2:       "https://example.com/success",
		SuccessURL2Method: "POST",
		FailURL2:          "https://example.com/fail",
		FailURL2Method:    "GET",
		Password:          "secret1",
		ShpFields: map[string]string{
			"Shp_item": "digital",
		},
	}, "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "4ec77aca3fd8a369fdaff8c2975daa94" {
		t.Fatalf("unexpected signature: %s", got)
	}
}

func TestSignOpState(t *testing.T) {
	svc := New("md5")
	got, err := svc.SignOpState("demo", "1001", "secret2", "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "1d0bb6fc90e81942e833f9fb4bdf0751" {
		t.Fatalf("unexpected signature: %s", got)
	}
}
