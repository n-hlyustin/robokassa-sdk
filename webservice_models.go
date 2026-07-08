package robokassa

type PaymentMethodsResponse struct {
	RawXML  string
	Result  XMLResult
	Groups  []PaymentMethodGroup
	RawData map[string]interface{}
}

type OpStateResponse struct {
	RawXML  string
	Result  XMLResult
	State   *InvoiceState
	RawData map[string]interface{}
}

type CurrenciesResponse struct {
	RawXML  string
	Result  XMLResult
	Groups  []CurrencyGroup
	RawData map[string]interface{}
}

type XMLResult struct {
	Code        string
	Description string
}

type PaymentMethodGroup struct {
	Code        string
	Description string
	Items       []PaymentMethod
}

type PaymentMethod struct {
	Code        string
	Description string
}

type CurrencyGroup struct {
	Code        string
	Description string
	Items       []Currency
}

type Currency struct {
	Label string
	Name  string
}

type InvoiceState struct {
	InvoiceID     string
	StateCode     string
	StateName     string
	OutSum        string
	IncSum        string
	PaymentMethod string
	OpKey         string
}
