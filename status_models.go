package robokassa

type InvoiceInformationListRequest struct {
	CurrentPage     int      `json:"CurrentPage"`
	PageSize        int      `json:"PageSize"`
	InvoiceStatuses []string `json:"InvoiceStatuses"`
	DateFrom        string   `json:"DateFrom"`
	DateTo          string   `json:"DateTo"`
	InvoiceTypes    []string `json:"InvoiceTypes"`
}

type InvoiceInformationListInvoiceItem struct {
	Name          string  `json:"name"`
	Quantity      float64 `json:"quantity"`
	Cost          float64 `json:"cost"`
	Tax           string  `json:"tax"`
	PaymentMethod string  `json:"paymentMethod,omitempty"`
	PaymentObject string  `json:"paymentObject,omitempty"`
}

type InvoiceInformationList struct {
	ID                              string                              `json:"id"`
	InvID                           int64                               `json:"invId"`
	InvoiceType                     string                              `json:"invoiceType"`
	Created                         string                              `json:"created"`
	Modified                        string                              `json:"modified"`
	OutSum                          float64                             `json:"outSum"`
	Description                     string                              `json:"description"`
	InvoiceStatus                   string                              `json:"invoiceStatus"`
	InvoicePaymentURL               string                              `json:"invoicePaymentUrl"`
	IsCustomerNotificationEmailSent bool                                `json:"isCustomerNotificationEmailSent"`
	Aliases                         []string                            `json:"aliases"`
	InvoiceItems                    []InvoiceInformationListInvoiceItem `json:"invoiceItems"`
	UserFields                      interface{}                         `json:"userFields"`
	Payments                        []interface{}                       `json:"payments"`
	IsWithoutFreeSale               bool                                `json:"isWithoutFreeSale"`
}

type InvoiceInformationListResponse struct {
	InvoiceInformationList []InvoiceInformationList `json:"invoiceInformationList"`
	Total                  int                      `json:"total"`
	IsSuccess              bool                     `json:"isSuccess"`
	Message                string                   `json:"message,omitempty"`
}
