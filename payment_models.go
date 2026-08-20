package robokassa

type CreateInvoiceRequest struct {
	InvoiceType      string      `json:"InvoiceType,omitempty"`
	Culture          string      `json:"Culture,omitempty"`
	InvID            int64       `json:"InvId"`
	OutSum           float64     `json:"OutSum"`
	Description      string      `json:"Description,omitempty"`
	MerchantComments string      `json:"MerchantComments,omitempty"`
	InvoiceItems     interface{} `json:"InvoiceItems,omitempty"`
	UserFields       interface{} `json:"UserFields,omitempty"`
	SuccessURL2Data  interface{} `json:"SuccessUrl2Data,omitempty"`
	FailURL2Data     interface{} `json:"FailUrl2Data,omitempty"`
}

type InvoiceItem struct {
	Name          string  `json:"Name"`
	Quantity      float64 `json:"Quantity"`
	Cost          float64 `json:"Cost"`
	Tax           string  `json:"Tax"`
	PaymentMethod string  `json:"PaymentMethod,omitempty"`
	PaymentObject string  `json:"PaymentObject,omitempty"`
}

type CreatePaymentRequest struct {
	OutSum      string
	InvID       string
	InvoiceID   string // Deprecated: use InvID. Kept for backwards compatibility.
	Description string

	Culture        string
	Encoding       string
	Email          string
	IncCurrLabel   string
	ExpirationDate string
	PaymentMethods []string

	Receipt           interface{}
	StepByStep        string
	ResultURL2        string
	SuccessURL2       string
	SuccessURL2Method string
	FailURL2          string
	FailURL2Method    string
	Token             string
	Recurring         string

	ShpFields map[string]string
	Extra     map[string]string
}

type CreatePaymentResponse struct {
	InvoiceID  string `json:"invoiceID"`
	InvoiceURL string
}
