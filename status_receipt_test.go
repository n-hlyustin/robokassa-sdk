package robokassa

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/robokassa/sdk-go-main/internal/signature"
)

func TestGetInvoiceInformationListRejectsNon2xxStatus(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       io.NopCloser(bytes.NewBufferString(`bad gateway`)),
			Header:     make(http.Header),
		}, nil
	})}

	service := &StatusService{
		transport:      &transport{httpClient: httpClient},
		signer:         signature.New("md5"),
		merchantLogin:  "demo",
		password1:      "secret1",
		invoiceListURL: "https://services.robokassa.ru/InvoiceServiceWebApi/api/GetInvoiceInformationList",
	}

	_, err := service.GetInvoiceInformationList(context.Background(), InvoiceInformationListRequest{
		CurrentPage:     1,
		PageSize:        10,
		InvoiceStatuses: []string{"Paid"},
		DateFrom:        "2026-01-01T00:00:00",
		DateTo:          "2026-01-31T23:59:59",
		InvoiceTypes:    []string{"OneTime"},
	})
	assertSDKStatus(t, err, http.StatusBadGateway)
}

func TestSendSecondCheckRejectsNon2xxStatus(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString(`server error`)),
			Header:     make(http.Header),
		}, nil
	})}

	service := &ReceiptService{
		transport:      &transport{httpClient: httpClient},
		signer:         signature.New("md5"),
		password1:      "secret1",
		hashType:       "md5",
		secondCheckURL: "https://ws.roboxchange.com/RoboFiscal/Receipt/Attach",
	}

	_, err := service.SendSecondCheck(context.Background(), SecondCheckRequest{
		"merchantId": "demo",
		"id":         "receipt-id",
	})
	assertSDKStatus(t, err, http.StatusInternalServerError)
}

func TestGetCheckStatusRejectsNon2xxStatus(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(bytes.NewBufferString(`unavailable`)),
			Header:     make(http.Header),
		}, nil
	})}

	service := &ReceiptService{
		transport:      &transport{httpClient: httpClient},
		signer:         signature.New("md5"),
		password1:      "secret1",
		hashType:       "md5",
		checkStatusURL: "https://ws.roboxchange.com/RoboFiscal/Receipt/Status",
	}

	_, err := service.GetCheckStatus(context.Background(), CheckStatusRequest{
		MerchantID: "demo",
		ID:         "receipt-id",
	})
	assertSDKStatus(t, err, http.StatusServiceUnavailable)
}

func assertSDKStatus(t *testing.T, err error, status int) {
	t.Helper()
	if err == nil {
		t.Fatal("expected HTTP status error")
	}
	var sdkErr *SDKError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected SDKError, got %T", err)
	}
	if sdkErr.StatusCode != status {
		t.Fatalf("unexpected status: want %d got %d", status, sdkErr.StatusCode)
	}
}
