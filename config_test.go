package robokassa

import "testing"

func TestConfigNormalizeUsesTestPasswords(t *testing.T) {
	cfg := Config{
		Login:         "demo",
		Password1:     "prod1",
		Password2:     "prod2",
		TestPassword1: "test1",
		TestPassword2: "test2",
		IsTest:        true,
	}
	if err := cfg.normalize(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Password1 != "test1" || cfg.Password2 != "test2" {
		t.Fatalf("test passwords were not applied: %#v", cfg)
	}
}

func TestConfigNormalizeRejectsBadHash(t *testing.T) {
	cfg := Config{
		Login:     "demo",
		Password1: "prod1",
		Password2: "prod2",
		HashType:  "sha1",
	}
	if err := cfg.normalize(); err == nil {
		t.Fatal("expected error for unsupported hash")
	}
}
