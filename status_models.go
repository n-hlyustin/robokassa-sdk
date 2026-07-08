package robokassa

type InvoiceInformationListRequest struct {
	CurrentPage     int      `json:"CurrentPage"`
	PageSize        int      `json:"PageSize"`
	InvoiceStatuses []string `json:"InvoiceStatuses"`
	DateFrom        string   `json:"DateFrom"`
	DateTo          string   `json:"DateTo"`
	InvoiceTypes    []string `json:"InvoiceTypes"`
}

type InvoiceInformationListResponse map[string]interface{}
