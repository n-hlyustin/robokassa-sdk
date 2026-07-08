package signature

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"sort"
	"strings"
)

type Service struct {
	defaultAlgo string
}

type PaymentSignatureParams struct {
	Login             string
	OutSum            string
	InvID             string
	Receipt           string
	StepByStep        string
	ResultURL2        string
	SuccessURL2       string
	SuccessURL2Method string
	FailURL2          string
	FailURL2Method    string
	Token             string
	Password          string
	ShpFields         map[string]string
}

func New(defaultAlgo string) *Service {
	if defaultAlgo == "" {
		defaultAlgo = "md5"
	}
	return &Service{defaultAlgo: strings.ToLower(defaultAlgo)}
}

func (s *Service) Base64URL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func (s *Service) SignFiscal(base64Payload, secret, algo string) (string, error) {
	sum, err := hashBytes(normalizeAlgo(algo, s.defaultAlgo), []byte(base64Payload+secret))
	if err != nil {
		return "", err
	}
	return s.Base64URL([]byte(hex.EncodeToString(sum))), nil
}

func (s *Service) JWTSignMD5(dataToSign, merchantLogin, password1 string) string {
	key := []byte(merchantLogin + ":" + password1)
	mac := hmac.New(md5.New, key)
	_, _ = mac.Write([]byte(dataToSign))
	return s.Base64URL(mac.Sum(nil))
}

func (s *Service) EncodeJWTParts(header, payload interface{}) (string, string, string, error) {
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", "", "", err
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", "", "", err
	}
	encodedHeader := s.Base64URL(headerJSON)
	encodedPayload := s.Base64URL(payloadJSON)
	return encodedHeader, encodedPayload, encodedHeader + "." + encodedPayload, nil
}

func (s *Service) CreatePaymentSignature(params map[string]string, login, password1, algo string) (string, error) {
	shpFields := make(map[string]string)
	for key, value := range params {
		if strings.HasPrefix(strings.ToLower(key), "shp_") {
			shpFields[key] = value
		}
	}

	return s.CreatePaymentSignatureFromParams(PaymentSignatureParams{
		Login:             login,
		OutSum:            params["OutSum"],
		InvID:             firstNonEmpty(params["InvId"], params["InvID"], params["InvoiceID"]),
		Receipt:           params["Receipt"],
		StepByStep:        params["StepByStep"],
		ResultURL2:        params["ResultUrl2"],
		SuccessURL2:       params["SuccessUrl2"],
		SuccessURL2Method: params["SuccessUrl2Method"],
		FailURL2:          params["FailUrl2"],
		FailURL2Method:    params["FailUrl2Method"],
		Token:             params["Token"],
		Password:          password1,
		ShpFields:         shpFields,
	}, algo)
}

func (s *Service) CreatePaymentSignatureFromParams(params PaymentSignatureParams, algo string) (string, error) {
	parts := make([]string, 0, 12+len(params.ShpFields))
	if params.Login != "" {
		parts = append(parts, params.Login)
	}
	parts = append(parts, params.OutSum, params.InvID)
	for _, value := range []string{
		params.Receipt,
		params.StepByStep,
		params.ResultURL2,
		params.SuccessURL2,
		params.SuccessURL2Method,
		params.FailURL2,
		params.FailURL2Method,
		params.Token,
	} {
		if value != "" {
			parts = append(parts, value)
		}
	}
	parts = append(parts, params.Password)
	parts = append(parts, sortedShpPairs(params.ShpFields)...)

	sum, err := hashBytes(normalizeAlgo(algo, s.defaultAlgo), []byte(strings.Join(parts, ":")))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sum), nil
}

func sortedShpPairs(fields map[string]string) []string {
	pairs := make([]string, 0, len(fields))
	for key, value := range fields {
		if strings.HasPrefix(strings.ToLower(key), "shp_") {
			pairs = append(pairs, key+"="+value)
		}
	}
	sort.Strings(pairs)
	return pairs
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (s *Service) SignOpState(login, invoiceID, password2, algo string) (string, error) {
	sum, err := hashBytes(normalizeAlgo(algo, s.defaultAlgo), []byte(login+":"+invoiceID+":"+password2))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sum), nil
}

func normalizeAlgo(algo, fallback string) string {
	value := strings.ToLower(strings.TrimSpace(algo))
	if value == "" {
		value = strings.ToLower(strings.TrimSpace(fallback))
	}
	switch value {
	case "md5", "sha256", "sha512":
		return value
	default:
		return "md5"
	}
}

func hashBytes(algo string, data []byte) ([]byte, error) {
	var h hash.Hash
	switch algo {
	case "md5":
		h = md5.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		return nil, fmt.Errorf("unsupported hash algorithm %q", algo)
	}
	_, _ = h.Write(data)
	return h.Sum(nil), nil
}
