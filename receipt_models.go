package robokassa

type Receipt struct {
	Sno   string        `json:"sno,omitempty"`
	Items []ReceiptItem `json:"items"`
}

type ReceiptItem struct {
	Name          string  `json:"name"`
	Quantity      float64 `json:"quantity"`
	Sum           float64 `json:"sum"`
	Tax           string  `json:"tax"`
	PaymentMethod string  `json:"payment_method,omitempty"`
	PaymentObject string  `json:"payment_object,omitempty"`
	Nomenclature  string  `json:"nomenclature_code,omitempty"`
}

type CheckStatusRequest struct {
	MerchantID string `json:"merchantId"`
	ID         string `json:"id"`
	OriginID   string `json:"originId,omitempty"`
}

type SecondCheckRequest map[string]interface{}
