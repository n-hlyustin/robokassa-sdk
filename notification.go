package robokassa

import (
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type NotificationService struct {
	signer    signer
	password1 string
	password2 string
	hashType  string
}

type ResultNotification struct {
	OutSum         string
	InvoiceID      string
	Fee            string
	Email          string
	PaymentMethod  string
	IncCurrLabel   string
	SignatureValue string
	ShpFields      map[string]string
}

type SuccessNotification struct {
	OutSum         string
	InvoiceID      string
	Culture        string
	SignatureValue string
	ShpFields      map[string]string
}

type ResultHandlerFunc func(http.ResponseWriter, *http.Request, ResultNotification) error

func (s *NotificationService) ParseResultURL(values url.Values) ResultNotification {
	notification := ResultNotification{
		OutSum:         values.Get("OutSum"),
		InvoiceID:      firstNonEmpty(values.Get("InvId"), values.Get("InvoiceID")),
		Fee:            values.Get("Fee"),
		Email:          firstNonEmpty(values.Get("EMail"), values.Get("Email")),
		PaymentMethod:  values.Get("PaymentMethod"),
		IncCurrLabel:   values.Get("IncCurrLabel"),
		SignatureValue: firstNonEmpty(values.Get("SignatureValue"), values.Get("Signature")),
		ShpFields:      map[string]string{},
	}
	for key, items := range values {
		if strings.HasPrefix(strings.ToLower(key), "shp_") && len(items) > 0 {
			notification.ShpFields[key] = items[0]
		}
	}
	return notification
}

func (s *NotificationService) ResultURLHandler(next ResultHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		values, err := parseCallbackValues(r)
		if err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}

		notification := s.ParseResultURL(values)
		if err := s.VerifyResultURL(notification); err != nil {
			http.Error(w, "invalid signature", http.StatusBadRequest)
			return
		}
		if next != nil {
			if err := next(w, r, notification); err != nil {
				http.Error(w, "callback failed", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK" + notification.InvoiceID))
	})
}

func (s *NotificationService) ParseSuccessURL(values url.Values) SuccessNotification {
	notification := SuccessNotification{
		OutSum:         values.Get("OutSum"),
		InvoiceID:      firstNonEmpty(values.Get("InvId"), values.Get("InvoiceID")),
		Culture:        values.Get("Culture"),
		SignatureValue: firstNonEmpty(values.Get("SignatureValue"), values.Get("Signature")),
		ShpFields:      map[string]string{},
	}
	for key, items := range values {
		if strings.HasPrefix(strings.ToLower(key), "shp_") && len(items) > 0 {
			notification.ShpFields[key] = items[0]
		}
	}
	return notification
}

func parseCallbackValues(r *http.Request) (url.Values, error) {
	switch r.Method {
	case http.MethodGet:
		return r.URL.Query(), nil
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		return r.PostForm, nil
	default:
		return nil, fmt.Errorf("unsupported method %s", r.Method)
	}
}

func (s *NotificationService) VerifyResultURL(notification ResultNotification) error {
	if notification.OutSum == "" || notification.InvoiceID == "" || notification.SignatureValue == "" {
		return fmt.Errorf("required fields: OutSum, InvoiceID, SignatureValue")
	}
	payload := map[string]string{
		"OutSum": notification.OutSum,
		"InvId":  notification.InvoiceID,
	}
	for key, value := range notification.ShpFields {
		payload[key] = value
	}
	expected, err := s.signer.CreatePaymentSignature(payload, "", s.password2, s.hashType)
	if err != nil {
		return err
	}
	if !secureSignatureEqual(expected, notification.SignatureValue) {
		return fmt.Errorf("invalid ResultURL signature")
	}
	return nil
}

func (s *NotificationService) VerifySuccessURL(notification SuccessNotification) error {
	if notification.OutSum == "" || notification.InvoiceID == "" || notification.SignatureValue == "" {
		return fmt.Errorf("required fields: OutSum, InvoiceID, SignatureValue")
	}
	payload := map[string]string{
		"OutSum": notification.OutSum,
		"InvId":  notification.InvoiceID,
	}
	for key, value := range notification.ShpFields {
		payload[key] = value
	}
	expected, err := s.signer.CreatePaymentSignature(payload, "", s.password1, s.hashType)
	if err != nil {
		return err
	}
	if !secureSignatureEqual(expected, notification.SignatureValue) {
		return fmt.Errorf("invalid SuccessURL signature")
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func secureSignatureEqual(expected, actual string) bool {
	expectedBytes, err := hex.DecodeString(strings.TrimSpace(expected))
	if err != nil {
		return false
	}
	actualBytes, err := hex.DecodeString(strings.TrimSpace(actual))
	if err != nil {
		return false
	}
	if len(expectedBytes) != len(actualBytes) {
		return false
	}
	return subtle.ConstantTimeCompare(expectedBytes, actualBytes) == 1
}
