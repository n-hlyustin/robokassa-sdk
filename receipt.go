package robokassa

import (
	"context"
	"encoding/json"
	"fmt"
)

type ReceiptService struct {
	transport      *transport
	signer         signer
	password1      string
	hashType       string
	secondCheckURL string
	checkStatusURL string
}

func (s *ReceiptService) BuildSecondCheckToken(req SecondCheckRequest) (string, error) {
	raw, err := json.Marshal(req)
	if err != nil {
		return "", &SDKError{Op: "receipt.build_second_check_token", Message: "failed to encode JSON", Err: err}
	}
	payload := s.signer.Base64URL(raw)
	signatureValue, err := s.signer.SignFiscal(payload, s.password1, s.hashType)
	if err != nil {
		return "", err
	}
	return payload + "." + signatureValue, nil
}

func (s *ReceiptService) SendSecondCheck(ctx context.Context, req SecondCheckRequest) (string, error) {
	body, err := s.BuildSecondCheckToken(req)
	if err != nil {
		return "", err
	}
	resp, err := s.transport.post(ctx, s.secondCheckURL, []byte(body), map[string]string{
		"Content-Type": "application/json",
	})
	if err != nil {
		return "", err
	}
	if resp.Status < 200 || resp.Status >= 300 {
		return "", &SDKError{Op: "receipt.send_second_check", StatusCode: resp.Status, Message: "unexpected HTTP status"}
	}
	return string(resp.Body), nil
}

func (s *ReceiptService) GetCheckStatus(ctx context.Context, req CheckStatusRequest) (map[string]interface{}, error) {
	if req.MerchantID == "" || req.ID == "" {
		return nil, fmt.Errorf("required fields: MerchantID, ID")
	}
	payload := SecondCheckRequest{
		"merchantId": req.MerchantID,
		"id":         req.ID,
	}
	if req.OriginID != "" {
		payload["originId"] = req.OriginID
	}
	body, err := s.BuildSecondCheckToken(payload)
	if err != nil {
		return nil, err
	}
	resp, err := s.transport.post(ctx, s.checkStatusURL, []byte(body), map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	})
	if err != nil {
		return nil, err
	}
	if resp.Status < 200 || resp.Status >= 300 {
		return nil, &SDKError{Op: "receipt.get_check_status", StatusCode: resp.Status, Message: "unexpected HTTP status"}
	}
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return nil, &SDKError{Op: "receipt.get_check_status", StatusCode: resp.Status, Message: "bad JSON in response", Err: err}
	}
	return data, nil
}
