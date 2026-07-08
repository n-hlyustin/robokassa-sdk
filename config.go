package robokassa

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	defaultPaymentURL        = "https://auth.robokassa.ru/Merchant/Index.aspx"
	defaultPaymentInvoiceURL = "https://auth.robokassa.ru/Merchant/Index/"
	defaultPaymentCurlURL    = "https://auth.robokassa.ru/Merchant/Indexjson.aspx"
	defaultJWTAPIURL         = "https://services.robokassa.ru/InvoiceServiceWebApi/api/CreateInvoice"
	defaultInvoiceListURL    = "https://services.robokassa.ru/InvoiceServiceWebApi/api/GetInvoiceInformationList"
	defaultWebServiceURL     = "https://auth.robokassa.ru/Merchant/WebService/Service.asmx"
	defaultCurrenciesURL     = "https://auth.robokassa.ru/Merchant/WebService/Service.asmx"
	defaultSecondCheckURL    = "https://ws.roboxchange.com/RoboFiscal/Receipt/Attach"
	defaultCheckStatusURL    = "https://ws.roboxchange.com/RoboFiscal/Receipt/Status"
	defaultHashType          = "md5"
	defaultHTTPTimeout       = 15 * time.Second
)

type Config struct {
	Login         string
	Password1     string
	Password2     string
	TestPassword1 string
	TestPassword2 string
	HashType      string
	IsTest        bool
	HTTPTimeout   time.Duration

	PaymentURL        string
	PaymentInvoiceURL string
	PaymentCurlURL    string
	JWTAPIURL         string
	WebServiceURL     string
	CurrenciesURL     string
	InvoiceListURL    string
	SecondCheckURL    string
	CheckStatusURL    string
}

func (c *Config) normalize() error {
	if strings.TrimSpace(c.Login) == "" {
		return fmt.Errorf("login is required")
	}
	if strings.TrimSpace(c.Password1) == "" {
		return fmt.Errorf("password1 is required")
	}
	if strings.TrimSpace(c.Password2) == "" {
		return fmt.Errorf("password2 is required")
	}
	if c.IsTest {
		if strings.TrimSpace(c.TestPassword1) == "" {
			return fmt.Errorf("test_password1 is required in test mode")
		}
		if strings.TrimSpace(c.TestPassword2) == "" {
			return fmt.Errorf("test_password2 is required in test mode")
		}
		c.Password1 = c.TestPassword1
		c.Password2 = c.TestPassword2
	}
	if c.HashType == "" {
		c.HashType = defaultHashType
	}
	c.HashType = strings.ToLower(c.HashType)
	switch c.HashType {
	case "md5", "sha256", "sha512":
	default:
		return fmt.Errorf("unsupported hash type %q", c.HashType)
	}
	if c.HTTPTimeout == 0 {
		c.HTTPTimeout = defaultHTTPTimeout
	}
	if c.PaymentURL == "" {
		c.PaymentURL = defaultPaymentURL
	}
	if c.PaymentInvoiceURL == "" {
		c.PaymentInvoiceURL = defaultPaymentInvoiceURL
	}
	if c.PaymentCurlURL == "" {
		c.PaymentCurlURL = defaultPaymentCurlURL
	}
	if c.JWTAPIURL == "" {
		c.JWTAPIURL = defaultJWTAPIURL
	}
	if c.WebServiceURL == "" {
		c.WebServiceURL = defaultWebServiceURL
	}
	if c.CurrenciesURL == "" {
		c.CurrenciesURL = defaultCurrenciesURL
	}
	if c.InvoiceListURL == "" {
		c.InvoiceListURL = defaultInvoiceListURL
	}
	if c.SecondCheckURL == "" {
		c.SecondCheckURL = defaultSecondCheckURL
	}
	if c.CheckStatusURL == "" {
		c.CheckStatusURL = defaultCheckStatusURL
	}
	return nil
}

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}
