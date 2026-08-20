package robokassa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

var paymentParamOrder = []string{
	"MerchantLogin",
	"OutSum",
	"InvId",
	"Description",
	"Receipt",
	"SignatureValue",
	"IsTest",
	"Culture",
	"Encoding",
	"Email",
	"IncCurrLabel",
	"ExpirationDate",
	"PaymentMethods",
	"StepByStep",
	"ResultUrl2",
	"SuccessUrl2",
	"SuccessUrl2Method",
	"FailUrl2",
	"FailUrl2Method",
	"Token",
	"Recurring",
}

type PaymentService struct {
	transport     *transport
	signer        signer
	merchantLogin string
	password1     string
	isTest        bool
	paymentURL    string
	invoiceURL    string
	paymentCurl   string
	jwtAPIURL     string
	hashType      string
}

type signer interface {
	Base64URL([]byte) string
	SignFiscal(string, string, string) (string, error)
	JWTSignMD5(string, string, string) string
	EncodeJWTParts(interface{}, interface{}) (string, string, string, error)
	CreatePaymentSignature(map[string]string, string, string, string) (string, error)
	SignOpState(string, string, string, string) (string, error)
}

func (s *PaymentService) BuildPaymentURL(req CreatePaymentRequest) (string, error) {
	params, err := s.prepareFormParams(req)
	if err != nil {
		return "", err
	}
	return s.paymentURL + "?" + encodePaymentParams(params), nil
}

func (s *PaymentService) SendForm(ctx context.Context, req CreatePaymentRequest) (*CreatePaymentResponse, error) {
	params, sigParams, err := s.prepareCurlParams(req)
	if err != nil {
		return nil, err
	}
	signatureValue, err := s.signer.CreatePaymentSignature(sigParams, s.merchantLogin, s.password1, s.hashType)
	if err != nil {
		return nil, err
	}
	params.Set("SignatureValue", signatureValue)

	resp, err := s.transport.post(ctx, s.paymentCurl, []byte(encodePaymentParams(params)), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, &SDKError{Op: "payment.send_form", StatusCode: resp.Status, Message: "unexpected HTTP status"}
	}

	data := CreatePaymentResponse{}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return nil, &SDKError{Op: "payment.send_form", Message: "failed to parse JSON", Err: err}
	}
	if data.InvoiceID == "" {
		return nil, &SDKError{Op: "payment.send_form", Message: "invoiceID not found in response"}
	}
	data.InvoiceURL = s.invoiceURL + fmt.Sprint(data.InvoiceID)
	return &data, nil
}

func (s *PaymentService) SendJWT(ctx context.Context, req CreateInvoiceRequest) (string, error) {
	payload, err := s.buildJWTPayload(req)
	if err != nil {
		return "", err
	}
	_, _, toSign, err := s.signer.EncodeJWTParts(map[string]string{"alg": "MD5", "typ": "JWT"}, payload)
	if err != nil {
		return "", &SDKError{Op: "payment.send_jwt", Message: "failed to encode JWT", Err: err}
	}
	token := toSign + "." + s.signer.JWTSignMD5(toSign, s.merchantLogin, s.password1)
	body, err := json.Marshal(token)
	if err != nil {
		return "", &SDKError{Op: "payment.send_jwt", Message: "failed to encode request body", Err: err}
	}

	resp, err := s.transport.post(ctx, s.jwtAPIURL, body, map[string]string{
		"Content-Type": "application/json",
	})
	if err != nil {
		return "", err
	}
	if resp.Status < 200 || resp.Status >= 300 {
		return "", &SDKError{Op: "payment.send_jwt", StatusCode: resp.Status, Message: "unexpected HTTP status"}
	}

	var data struct {
		URL       string `json:"url"`
		IsSuccess bool   `json:"isSuccess"`
		Message   string `json:"message"`
	}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return "", &SDKError{Op: "payment.send_jwt", StatusCode: resp.Status, Message: "bad JSON in response", Err: err}
	}
	if data.URL == "" {
		message := data.Message
		if message == "" {
			message = fmt.Sprintf("url not found in response: %s", string(resp.Body))
		}
		return "", &SDKError{Op: "payment.send_jwt", StatusCode: resp.Status, Message: message}
	}
	return data.URL, nil
}

func (s *PaymentService) prepareFormParams(req CreatePaymentRequest) (url.Values, error) {
	params, sigParams, err := s.prepareCurlParams(req)
	if err != nil {
		return nil, err
	}
	signatureValue, err := s.signer.CreatePaymentSignature(sigParams, s.merchantLogin, s.password1, s.hashType)
	if err != nil {
		return nil, err
	}
	params.Set("SignatureValue", signatureValue)
	return params, nil
}

func (s *PaymentService) prepareCurlParams(req CreatePaymentRequest) (url.Values, map[string]string, error) {
	if req.OutSum == "" || req.Description == "" {
		return nil, nil, fmt.Errorf("required fields: OutSum, Description")
	}
	params := url.Values{}
	invID := req.paymentInvID()
	params.Set("MerchantLogin", s.merchantLogin)
	params.Set("OutSum", req.OutSum)
	params.Set("Description", req.Description)
	if invID != "" {
		params.Set("InvId", invID)
	}
	if req.Culture != "" {
		params.Set("Culture", req.Culture)
	}
	if req.Encoding != "" {
		params.Set("Encoding", req.Encoding)
	}
	if req.Email != "" {
		params.Set("Email", req.Email)
	}
	if req.IncCurrLabel != "" {
		params.Set("IncCurrLabel", req.IncCurrLabel)
	}
	if req.ExpirationDate != "" {
		params.Set("ExpirationDate", req.ExpirationDate)
	}
	for _, method := range req.PaymentMethods {
		params.Add("PaymentMethods", method)
	}
	receipt := ""
	if req.Receipt != nil {
		raw, err := json.Marshal(req.Receipt)
		if err != nil {
			return nil, nil, &SDKError{Op: "payment.prepare_form", Message: "failed to encode receipt", Err: err}
		}
		receipt = url.QueryEscape(string(raw))
		params.Set("Receipt", receipt)
	}
	for _, item := range []struct {
		key   string
		value string
	}{
		{"StepByStep", req.StepByStep},
		{"ResultUrl2", req.ResultURL2},
		{"SuccessUrl2", req.SuccessURL2},
		{"SuccessUrl2Method", req.SuccessURL2Method},
		{"FailUrl2", req.FailURL2},
		{"FailUrl2Method", req.FailURL2Method},
		{"Token", req.Token},
		{"Recurring", req.Recurring},
	} {
		if item.value != "" {
			params.Set(item.key, item.value)
		}
	}
	if s.isTest {
		params.Set("IsTest", "1")
	}
	keys := make([]string, 0, len(req.ShpFields))
	for key := range req.ShpFields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		params.Set(key, req.ShpFields[key])
	}
	for key, value := range req.Extra {
		if isReservedPaymentParam(key) || strings.HasPrefix(strings.ToLower(key), "shp_") {
			return nil, nil, fmt.Errorf("reserved payment parameter %q must be set via typed fields", key)
		}
		params.Set(key, value)
	}

	sigParams := map[string]string{
		"OutSum": req.OutSum,
		"InvId":  invID,
	}
	for _, item := range []struct {
		key   string
		value string
	}{
		{"Receipt", receipt},
		{"StepByStep", req.StepByStep},
		{"ResultUrl2", req.ResultURL2},
		{"SuccessUrl2", req.SuccessURL2},
		{"SuccessUrl2Method", req.SuccessURL2Method},
		{"FailUrl2", req.FailURL2},
		{"FailUrl2Method", req.FailURL2Method},
		{"Token", req.Token},
	} {
		if item.value != "" {
			sigParams[item.key] = item.value
		}
	}
	for _, key := range keys {
		sigParams[key] = req.ShpFields[key]
	}
	return params, sigParams, nil
}

func (s *PaymentService) buildJWTPayload(req CreateInvoiceRequest) (map[string]interface{}, error) {
	if req.OutSum == 0 || req.InvID == 0 {
		return nil, fmt.Errorf("required fields: OutSum, InvID")
	}
	payload := map[string]interface{}{
		"MerchantLogin": s.merchantLogin,
		"InvoiceType":   "OneTime",
		"Culture":       "ru",
		"InvId":         req.InvID,
		"OutSum":        req.OutSum,
	}
	if req.InvoiceType != "" {
		payload["InvoiceType"] = req.InvoiceType
	}
	if req.Culture != "" {
		payload["Culture"] = req.Culture
	}
	if req.Description != "" {
		payload["Description"] = req.Description
	}
	if req.MerchantComments != "" {
		payload["MerchantComments"] = req.MerchantComments
	}
	if req.InvoiceItems != nil {
		payload["InvoiceItems"] = req.InvoiceItems
	}
	if req.UserFields != nil {
		payload["UserFields"] = req.UserFields
	}
	if req.SuccessURL2Data != nil {
		payload["SuccessUrl2Data"] = req.SuccessURL2Data
	}
	if req.FailURL2Data != nil {
		payload["FailUrl2Data"] = req.FailURL2Data
	}
	return payload, nil
}

func (req CreatePaymentRequest) paymentInvID() string {
	if req.InvID != "" {
		return req.InvID
	}
	return req.InvoiceID
}

func isReservedPaymentParam(key string) bool {
	switch strings.ToLower(key) {
	case "merchantlogin",
		"outsum",
		"invid",
		"invoiceid",
		"description",
		"signaturevalue",
		"culture",
		"encoding",
		"email",
		"inccurrlabel",
		"expirationdate",
		"paymentmethods",
		"receipt",
		"stepbystep",
		"resulturl2",
		"successurl2",
		"successurl2method",
		"failurl2",
		"failurl2method",
		"token",
		"recurring",
		"istest":
		return true
	default:
		return false
	}
}

func encodePaymentParams(params url.Values) string {
	encoded := make([]string, 0, len(params))
	seen := make(map[string]bool, len(params))

	for _, key := range paymentParamOrder {
		appendEncodedParam(&encoded, params, key)
		seen[key] = true
	}

	shpKeys := make([]string, 0)
	extraKeys := make([]string, 0)
	for key := range params {
		if seen[key] {
			continue
		}
		if strings.HasPrefix(strings.ToLower(key), "shp_") {
			shpKeys = append(shpKeys, key)
			continue
		}
		extraKeys = append(extraKeys, key)
	}
	sort.Strings(shpKeys)
	sort.Strings(extraKeys)

	for _, key := range shpKeys {
		appendEncodedParam(&encoded, params, key)
	}
	for _, key := range extraKeys {
		appendEncodedParam(&encoded, params, key)
	}

	return strings.Join(encoded, "&")
}

func appendEncodedParam(encoded *[]string, params url.Values, key string) {
	values, ok := params[key]
	if !ok {
		return
	}
	escapedKey := url.QueryEscape(key)
	for _, value := range values {
		*encoded = append(*encoded, escapedKey+"="+url.QueryEscape(value))
	}
}
