package robokassa

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type StatusService struct {
	transport      *transport
	signer         signer
	merchantLogin  string
	password1      string
	invoiceListURL string
}

func (s *StatusService) GetInvoiceInformationList(ctx context.Context, req InvoiceInformationListRequest) (*InvoiceInformationListResponse, error) {
	if req.CurrentPage == 0 {
		return nil, fmt.Errorf("missing required field: CurrentPage")
	}
	if req.PageSize == 0 {
		return nil, fmt.Errorf("missing required field: PageSize")
	}
	if len(req.InvoiceStatuses) == 0 {
		return nil, fmt.Errorf("missing required field: InvoiceStatuses")
	}
	if req.DateFrom == "" {
		return nil, fmt.Errorf("missing required field: DateFrom")
	}
	if req.DateTo == "" {
		return nil, fmt.Errorf("missing required field: DateTo")
	}
	if len(req.InvoiceTypes) == 0 {
		return nil, fmt.Errorf("missing required field: InvoiceTypes")
	}

	payload := map[string]interface{}{
		"MerchantLogin":   s.merchantLogin,
		"CurrentPage":     req.CurrentPage,
		"PageSize":        req.PageSize,
		"InvoiceStatuses": normalizeInvoiceList(req.InvoiceStatuses),
		"DateFrom":        req.DateFrom,
		"DateTo":          req.DateTo,
		"InvoiceTypes":    normalizeInvoiceList(req.InvoiceTypes),
	}

	_, _, toSign, err := s.signer.EncodeJWTParts(map[string]string{"alg": "MD5", "typ": "JWT"}, payload)
	if err != nil {
		return nil, &SDKError{Op: "status.get_invoice_information_list", Message: "failed to encode JWT", Err: err}
	}
	token := toSign + "." + s.signer.JWTSignMD5(toSign, s.merchantLogin, s.password1)
	body, err := json.Marshal(token)
	if err != nil {
		return nil, &SDKError{Op: "status.get_invoice_information_list", Message: "failed to encode request body", Err: err}
	}

	resp, err := s.transport.post(ctx, s.invoiceListURL, body, map[string]string{
		"Content-Type": "application/json",
	})
	if err != nil {
		return nil, err
	}
	if resp.Status < 200 || resp.Status >= 300 {
		return nil, &SDKError{Op: "status.get_invoice_information_list", StatusCode: resp.Status, Message: "unexpected HTTP status"}
	}

	var data InvoiceInformationListResponse
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return nil, &SDKError{Op: "status.get_invoice_information_list", StatusCode: resp.Status, Message: "bad JSON in response", Err: err}
	}

	if !data.IsSuccess {
		return nil, &SDKError{Op: "status.get_invoice_information_list", StatusCode: resp.Status, Message: "body status is not successed", Err: fmt.Errorf("%s", data.Message)}
	}
	return &data, nil
}

func normalizeInvoiceList(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		switch strings.ToLower(item) {
		case "paid":
			out = append(out, "Paid")
		case "expired":
			out = append(out, "Expired")
		case "notpaid":
			out = append(out, "Notpaid")
		case "onetime":
			out = append(out, "OneTime")
		case "reusable":
			out = append(out, "Reusable")
		default:
			out = append(out, item)
		}
	}
	return out
}
