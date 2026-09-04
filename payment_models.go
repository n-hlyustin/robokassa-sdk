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

// RecurringPaymentRequest представляет запрос на повторяющийся платеж
type RecurringPaymentRequest struct {
	// PreviousInvoiceID - номер счета первого платежа в серии (обязательный)
	PreviousInvoiceID string `json:"PreviousInvoiceID"`
	// InvoiceID - номер нового счета, сгенерированный магазином (обязательный)
	InvoiceID string `json:"InvoiceID"`
	// OutSum - сумма платежа (обязательный)
	OutSum string `json:"OutSum"`
	// Description - описание платежа (обязательный)
	Description string `json:"Description"`
	// Culture - языковая локализация (ru/en)
	Culture string `json:"Culture,omitempty"`
	// Encoding - кодировка (utf-8/windows-1251)
	Encoding string `json:"Encoding,omitempty"`
	// Email - email покупателя
	Email string `json:"Email,omitempty"`
	// Receipt - данные для фискализации
	Receipt interface{} `json:"Receipt,omitempty"`
	// StepByStep - пошаговый режим
	StepByStep string `json:"StepByStep,omitempty"`
	// ResultURL2 - URL для уведомлений
	ResultURL2 string `json:"ResultUrl2,omitempty"`
	// SuccessURL2 - URL для успешной оплаты
	SuccessURL2 string `json:"SuccessUrl2,omitempty"`
	// SuccessURL2Method - метод для SuccessURL2 (GET/POST)
	SuccessURL2Method string `json:"SuccessUrl2Method,omitempty"`
	// FailURL2 - URL для неуспешной оплаты
	FailURL2 string `json:"FailUrl2,omitempty"`
	// FailURL2Method - метод для FailURL2 (GET/POST)
	FailURL2Method string `json:"FailUrl2Method,omitempty"`
	// Token - токен для привязки карты
	Token string `json:"Token,omitempty"`
	// ShpFields - пользовательские поля с префиксом Shp_
	ShpFields map[string]string `json:"-"`
	// Extra - дополнительные параметры
	Extra map[string]string `json:"-"`
}

// RecurringPaymentResponse представляет ответ на повторяющийся платеж
type RecurringPaymentResponse struct {
	InvoiceID string `json:"InvoiceID"`
	Raw       []byte
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
