package robokassa

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/n-hlyustin/robokassa-sdk/internal/signature"
)

func TestVerifyResultURL(t *testing.T) {
	signerSvc := signature.New("md5")
	expected, err := signerSvc.CreatePaymentSignature(map[string]string{
		"OutSum":   "99.90",
		"InvId":    "1001",
		"Shp_item": "digital",
	}, "", "secret2", "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	service := &NotificationService{
		signer:    signerSvc,
		password1: "secret1",
		password2: "secret2",
		hashType:  "md5",
	}
	values := url.Values{}
	values.Set("OutSum", "99.90")
	values.Set("InvId", "1001")
	values.Set("SignatureValue", expected)
	values.Set("Shp_item", "digital")

	notification := service.ParseResultURL(values)
	if err := service.VerifyResultURL(notification); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseResultURLIncludesMetadata(t *testing.T) {
	values := url.Values{}
	values.Set("OutSum", "99.90")
	values.Set("InvId", "1001")
	values.Set("Fee", "3.50")
	values.Set("EMail", "buyer@example.com")
	values.Set("PaymentMethod", "BankCard")
	values.Set("IncCurrLabel", "BankCard")
	values.Set("SignatureValue", "abc")

	service := &NotificationService{}
	notification := service.ParseResultURL(values)

	if notification.Fee != "3.50" {
		t.Fatalf("unexpected Fee: %q", notification.Fee)
	}
	if notification.Email != "buyer@example.com" {
		t.Fatalf("unexpected Email: %q", notification.Email)
	}
	if notification.PaymentMethod != "BankCard" {
		t.Fatalf("unexpected PaymentMethod: %q", notification.PaymentMethod)
	}
	if notification.IncCurrLabel != "BankCard" {
		t.Fatalf("unexpected IncCurrLabel: %q", notification.IncCurrLabel)
	}
}

func TestResultURLHandlerReturnsOK(t *testing.T) {
	signerSvc := signature.New("md5")
	expected, err := signerSvc.CreatePaymentSignature(map[string]string{
		"OutSum":   "99.90",
		"InvId":    "1001",
		"Shp_item": "digital",
	}, "", "secret2", "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	service := &NotificationService{
		signer:    signerSvc,
		password2: "secret2",
		hashType:  "md5",
	}
	called := false
	handler := service.ResultURLHandler(func(w http.ResponseWriter, r *http.Request, notification ResultNotification) error {
		called = true
		if notification.InvoiceID != "1001" {
			t.Fatalf("unexpected invoice id: %s", notification.InvoiceID)
		}
		if notification.ShpFields["Shp_item"] != "digital" {
			t.Fatalf("unexpected ShpFields: %#v", notification.ShpFields)
		}
		return nil
	})

	body := "OutSum=99.90&InvId=1001&SignatureValue=" + expected + "&Shp_item=digital"
	req := httptest.NewRequest(http.MethodPost, "/result", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "OK1001" {
		t.Fatalf("unexpected response body: %q", rec.Body.String())
	}
	if !called {
		t.Fatal("expected callback to be called")
	}
}

func TestResultURLHandlerUsesPostFormForPostCallbacks(t *testing.T) {
	signerSvc := signature.New("md5")
	expected, err := signerSvc.CreatePaymentSignature(map[string]string{
		"OutSum": "99.90",
		"InvId":  "1001",
	}, "", "secret2", "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	service := &NotificationService{
		signer:    signerSvc,
		password2: "secret2",
		hashType:  "md5",
	}
	handler := service.ResultURLHandler(nil)
	body := "OutSum=99.90&InvId=1001&SignatureValue=" + expected
	req := httptest.NewRequest(http.MethodPost, "/result?OutSum=1&InvId=999&SignatureValue=bad", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "OK1001" {
		t.Fatalf("unexpected response body: %q", rec.Body.String())
	}
}

func TestResultURLHandlerRejectsInvalidSignature(t *testing.T) {
	service := &NotificationService{
		signer:    signature.New("md5"),
		password2: "secret2",
		hashType:  "md5",
	}
	handler := service.ResultURLHandler(nil)
	body := "OutSum=99.90&InvId=1001&SignatureValue=00000000000000000000000000000000"
	req := httptest.NewRequest(http.MethodPost, "/result", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestResultURLHandlerRejectsUnsupportedMethod(t *testing.T) {
	service := &NotificationService{
		signer:    signature.New("md5"),
		password2: "secret2",
		hashType:  "md5",
	}
	handler := service.ResultURLHandler(nil)
	req := httptest.NewRequest(http.MethodPut, "/result", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Allow") != "GET, POST" {
		t.Fatalf("unexpected Allow header: %q", rec.Header().Get("Allow"))
	}
}

func TestVerifyResultURLWithMultipleShpFields(t *testing.T) {
	signerSvc := signature.New("md5")
	expected, err := signerSvc.CreatePaymentSignature(map[string]string{
		"OutSum":    "99.90",
		"InvId":     "1001",
		"Shp_b":     "second",
		"Shp_a":     "first",
		"Unrelated": "ignored",
	}, "", "secret2", "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	service := &NotificationService{
		signer:    signerSvc,
		password2: "secret2",
		hashType:  "md5",
	}
	err = service.VerifyResultURL(ResultNotification{
		OutSum:         "99.90",
		InvoiceID:      "1001",
		SignatureValue: expected,
		ShpFields: map[string]string{
			"Shp_b": "second",
			"Shp_a": "first",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifySuccessURL(t *testing.T) {
	signerSvc := signature.New("md5")
	expected, err := signerSvc.CreatePaymentSignature(map[string]string{
		"OutSum":   "99.90",
		"InvId":    "1001",
		"Shp_item": "digital",
	}, "", "secret1", "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	service := &NotificationService{
		signer:    signerSvc,
		password1: "secret1",
		password2: "secret2",
		hashType:  "md5",
	}
	if err := service.VerifySuccessURL(SuccessNotification{
		OutSum:         "99.90",
		InvoiceID:      "1001",
		SignatureValue: expected,
		ShpFields: map[string]string{
			"Shp_item": "digital",
		},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifySuccessURLRejectsInvalidSignature(t *testing.T) {
	service := &NotificationService{
		signer:    signature.New("md5"),
		password1: "secret1",
		hashType:  "md5",
	}

	err := service.VerifySuccessURL(SuccessNotification{
		OutSum:         "99.90",
		InvoiceID:      "1001",
		SignatureValue: "00000000000000000000000000000000",
	})
	if err == nil {
		t.Fatal("expected invalid signature error")
	}
}

func TestParseSuccessURL(t *testing.T) {
	values := url.Values{}
	values.Set("OutSum", "99.90")
	values.Set("InvId", "1001")
	values.Set("SignatureValue", "abc")
	values.Set("Culture", "ru")
	values.Set("Shp_item", "digital")

	service := &NotificationService{}
	notification := service.ParseSuccessURL(values)
	if notification.OutSum != "99.90" {
		t.Fatalf("unexpected OutSum: %q", notification.OutSum)
	}
	if notification.InvoiceID != "1001" {
		t.Fatalf("unexpected InvoiceID: %q", notification.InvoiceID)
	}
	if notification.Culture != "ru" {
		t.Fatalf("unexpected Culture: %q", notification.Culture)
	}
	if notification.ShpFields["Shp_item"] != "digital" {
		t.Fatalf("unexpected Shp_item: %#v", notification.ShpFields)
	}
}

func TestVerifyResultURLRequiresFields(t *testing.T) {
	service := &NotificationService{
		signer:    signature.New("md5"),
		password2: "secret2",
		hashType:  "md5",
	}

	err := service.VerifyResultURL(ResultNotification{
		OutSum: "99.90",
	})
	if err == nil {
		t.Fatal("expected required fields error")
	}
}

func TestVerifyResultURLRejectsInvalidSignature(t *testing.T) {
	service := &NotificationService{
		signer:    signature.New("md5"),
		password1: "secret1",
		password2: "secret2",
		hashType:  "md5",
	}

	err := service.VerifyResultURL(ResultNotification{
		OutSum:         "99.90",
		InvoiceID:      "1001",
		SignatureValue: "00000000000000000000000000000000",
	})
	if err == nil {
		t.Fatal("expected invalid signature error")
	}
}
