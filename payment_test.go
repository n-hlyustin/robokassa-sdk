package robokassa

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/robokassa/sdk-go-main/internal/signature"
)

func TestBuildPaymentURL(t *testing.T) {
	service := &PaymentService{
		signer:        signature.New("md5"),
		merchantLogin: "demo",
		password1:     "secret1",
		paymentURL:    "https://auth.robokassa.ru/Merchant/Index/",
		hashType:      "md5",
	}

	got, err := service.BuildPaymentURL(CreatePaymentRequest{
		OutSum:      "99.90",
		InvID:       "1001",
		Description: "test order",
		ShpFields: map[string]string{
			"Shp_item": "digital",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	checks := []string{
		"MerchantLogin=demo",
		"OutSum=99.90",
		"InvId=1001",
		"Description=test+order",
		"SignatureValue=",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("payment url does not contain %q: %s", check, got)
		}
	}

	expectedOrder := []string{
		"MerchantLogin=",
		"OutSum=",
		"InvId=",
		"Description=",
		"SignatureValue=",
		"Shp_item=",
	}
	assertQueryOrder(t, got, expectedOrder)
}

func TestBuildPaymentURLUsesLegacyInvoiceIDAsInvID(t *testing.T) {
	service := &PaymentService{
		signer:        signature.New("md5"),
		merchantLogin: "demo",
		password1:     "secret1",
		paymentURL:    "https://auth.robokassa.ru/Merchant/Index.aspx",
		hashType:      "md5",
	}

	got, err := service.BuildPaymentURL(CreatePaymentRequest{
		OutSum:      "99.90",
		InvoiceID:   "1001",
		Description: "test order",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("unexpected URL error: %v", err)
	}
	if parsed.Query().Get("InvId") != "1001" {
		t.Fatalf("expected legacy InvoiceID to be sent as InvId, got URL %s", got)
	}
	if parsed.Query().Get("InvoiceID") != "" {
		t.Fatalf("did not expect InvoiceID query parameter, got URL %s", got)
	}
}

func TestBuildPaymentURLIncludesReceiptAndSignatureModifiers(t *testing.T) {
	service := &PaymentService{
		signer:        signature.New("md5"),
		merchantLogin: "demo",
		password1:     "secret1",
		paymentURL:    "https://auth.robokassa.ru/Merchant/Index.aspx",
		hashType:      "md5",
	}

	got, err := service.BuildPaymentURL(CreatePaymentRequest{
		OutSum:            "8.96",
		InvID:             "12345",
		Description:       "test order",
		SuccessURL2:       "https://example.com/success",
		SuccessURL2Method: "POST",
		FailURL2:          "https://example.com/fail",
		FailURL2Method:    "GET",
		Receipt: map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"name":     "product",
					"quantity": 1,
					"sum":      8.96,
					"tax":      "none",
				},
			},
		},
		ShpFields: map[string]string{
			"Shp_item": "digital",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("unexpected URL error: %v", err)
	}
	query := parsed.Query()
	if query.Get("Receipt") == "" {
		t.Fatalf("expected Receipt in payment URL: %s", got)
	}
	if !strings.Contains(parsed.RawQuery, "Receipt=%25") {
		t.Fatalf("expected Receipt to be pre-escaped before query encoding: %s", got)
	}
	if !strings.HasPrefix(query.Get("Receipt"), "%7B") {
		t.Fatalf("expected decoded Receipt query value to remain URL-encoded JSON, got %s", query.Get("Receipt"))
	}
	if query.Get("SuccessUrl2") != "https://example.com/success" {
		t.Fatalf("expected SuccessUrl2 in payment URL: %s", got)
	}
	if query.Get("SignatureValue") == "" {
		t.Fatalf("expected SignatureValue in payment URL: %s", got)
	}
	assertQueryOrder(t, got, []string{
		"MerchantLogin=",
		"OutSum=",
		"InvId=",
		"Description=",
		"Receipt=",
		"SignatureValue=",
		"SuccessUrl2=",
		"SuccessUrl2Method=",
		"FailUrl2=",
		"FailUrl2Method=",
		"Shp_item=",
	})
}

func TestBuildPaymentURLReceiptWithCyrillicCanBeDecoded(t *testing.T) {
	service := &PaymentService{
		signer:        signature.New("md5"),
		merchantLogin: "demo",
		password1:     "secret1",
		paymentURL:    "https://auth.robokassa.ru/Merchant/Index.aspx",
		hashType:      "md5",
	}

	got, err := service.BuildPaymentURL(CreatePaymentRequest{
		OutSum:      "149.00",
		InvID:       "1003",
		Description: "Платеж через redirect",
		Receipt: Receipt{
			Items: []ReceiptItem{
				{Name: "Платеж через redirect", Quantity: 1, Sum: 149, Tax: "none"},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("unexpected URL error: %v", err)
	}
	receipt := parsed.Query().Get("Receipt")
	if !strings.HasPrefix(receipt, "%7B") {
		t.Fatalf("expected once-decoded Receipt to be URL-encoded JSON, got %s", receipt)
	}
	decoded, err := url.QueryUnescape(receipt)
	if err != nil {
		t.Fatalf("unexpected Receipt decode error: %v", err)
	}
	if !strings.Contains(decoded, `"name":"Платеж через redirect"`) {
		t.Fatalf("unexpected decoded Receipt JSON: %s", decoded)
	}
}

func TestBuildPaymentURLRejectsReservedExtra(t *testing.T) {
	service := &PaymentService{
		signer:        signature.New("md5"),
		merchantLogin: "demo",
		password1:     "secret1",
		paymentURL:    "https://auth.robokassa.ru/Merchant/Index.aspx",
		hashType:      "md5",
	}

	_, err := service.BuildPaymentURL(CreatePaymentRequest{
		OutSum:      "99.90",
		InvID:       "1001",
		Description: "test order",
		Extra: map[string]string{
			"OutSum": "1.00",
		},
	})
	if err == nil {
		t.Fatal("expected reserved Extra parameter error")
	}
}

func TestSendFormReturnsInvoiceURL(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		assertEncodedBodyOrder(t, string(body), []string{
			"MerchantLogin=",
			"OutSum=",
			"InvId=",
			"Description=",
			"Receipt=",
			"SignatureValue=",
		})
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"invoiceID":"f87e382e-d2d7-c384-c212-fd584a44985c"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	service := &PaymentService{
		transport:     &transport{httpClient: httpClient},
		signer:        signature.New("md5"),
		merchantLogin: "demo",
		password1:     "secret1",
		invoiceURL:    "https://auth.robokassa.ru/Merchant/Index/",
		paymentCurl:   "https://auth.robokassa.ru/Merchant/Indexjson.aspx",
		hashType:      "md5",
	}

	got, err := service.SendForm(context.Background(), CreatePaymentRequest{
		OutSum:      "99.90",
		InvID:       "1002",
		Description: "test order",
		Receipt: Receipt{
			Items: []ReceiptItem{
				{Name: "test order", Quantity: 1, Sum: 99.90, Tax: "none"},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://auth.robokassa.ru/Merchant/Index/f87e382e-d2d7-c384-c212-fd584a44985c"
	if got != want {
		t.Fatalf("unexpected invoice URL:\nwant %s\n got %s", want, got)
	}
}

func TestSendJWTSuccess(t *testing.T) {
	var requested bool
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requested = true
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %s", r.Header.Get("Content-Type"))
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(string(body), `"`) || !strings.Contains(string(body), ".") {
			t.Fatalf("expected JSON-encoded JWT token body, got %s", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"isSuccess":true,"url":"https://auth.robokassa.ru/merchant/Invoice/token"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	service := &PaymentService{
		transport:     &transport{httpClient: httpClient},
		signer:        signature.New("md5"),
		merchantLogin: "demo",
		password1:     "secret1",
		jwtAPIURL:     "https://services.robokassa.ru/InvoiceServiceWebApi/api/CreateInvoice",
		hashType:      "md5",
	}

	got, err := service.SendJWT(context.Background(), CreateInvoiceRequest{
		InvID:       1001,
		OutSum:      99.90,
		Description: "test order",
		InvoiceItems: []InvoiceItem{
			{Name: "test order", Quantity: 1, Cost: 99.90, Tax: "none"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !requested {
		t.Fatal("expected HTTP request")
	}
	if got != "https://auth.robokassa.ru/merchant/Invoice/token" {
		t.Fatalf("unexpected JWT payment URL: %s", got)
	}
}

func TestSendJWTReturnsRobokassaMessage(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"isSuccess":false,"message":"Заказ с таким Id уже существует"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	service := &PaymentService{
		transport:     &transport{httpClient: httpClient},
		signer:        signature.New("md5"),
		merchantLogin: "demo",
		password1:     "secret1",
		jwtAPIURL:     "https://services.robokassa.ru/InvoiceServiceWebApi/api/CreateInvoice",
		hashType:      "md5",
	}

	_, err := service.SendJWT(context.Background(), CreateInvoiceRequest{
		InvID:       1001,
		OutSum:      99.90,
		Description: "test order",
	})
	if err == nil {
		t.Fatal("expected Robokassa message error")
	}
	if !strings.Contains(err.Error(), "Заказ с таким Id уже существует") {
		t.Fatalf("expected Robokassa message in error, got %v", err)
	}
}

func TestSendJWTRejectsNon2xxStatus(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString(`{"message":"server error"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	service := &PaymentService{
		transport:     &transport{httpClient: httpClient},
		signer:        signature.New("md5"),
		merchantLogin: "demo",
		password1:     "secret1",
		jwtAPIURL:     "https://services.robokassa.ru/InvoiceServiceWebApi/api/CreateInvoice",
		hashType:      "md5",
	}

	_, err := service.SendJWT(context.Background(), CreateInvoiceRequest{
		InvID:       1001,
		OutSum:      99.90,
		Description: "test order",
	})
	if err == nil {
		t.Fatal("expected HTTP status error")
	}
	var sdkErr *SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %T", err)
	}
	if sdkErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", sdkErr.StatusCode)
	}
}

func TestSendJWTTransportError(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})}
	service := &PaymentService{
		transport:     &transport{httpClient: httpClient},
		signer:        signature.New("md5"),
		merchantLogin: "demo",
		password1:     "secret1",
		jwtAPIURL:     "https://services.robokassa.ru/InvoiceServiceWebApi/api/CreateInvoice",
		hashType:      "md5",
	}

	_, err := service.SendJWT(context.Background(), CreateInvoiceRequest{
		InvID:       1001,
		OutSum:      99.90,
		Description: "test order",
	})
	if err == nil {
		t.Fatal("expected transport error")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

func assertQueryOrder(t *testing.T, rawURL string, parts []string) {
	t.Helper()
	assertEncodedBodyOrder(t, rawURL, parts)
}

func assertEncodedBodyOrder(t *testing.T, value string, parts []string) {
	t.Helper()
	previous := -1
	for _, part := range parts {
		current := strings.Index(value, part)
		if current == -1 {
			t.Fatalf("encoded value does not contain %q: %s", part, value)
		}
		if current < previous {
			t.Fatalf("parameter %q is out of order in %s", part, value)
		}
		previous = current
	}
}
