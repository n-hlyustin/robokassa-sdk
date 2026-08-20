package robokassa

import (
	"net/http"
	"time"

	"github.com/n-hlyustin/robokassa-sdk/internal/signature"
)

type Client struct {
	config     Config
	httpClient *http.Client
	transport  *transport
	signer     *signature.Service

	payment    *PaymentService
	receipt    *ReceiptService
	webService *WebService
	status     *StatusService
	notify     *NotificationService
}

func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if err := cfg.normalize(); err != nil {
		return nil, err
	}

	client := &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
	}
	for _, opt := range opts {
		opt(client)
	}
	if client.httpClient.Timeout == 0 {
		client.httpClient.Timeout = 15 * time.Second
	}

	client.transport = &transport{httpClient: client.httpClient}
	client.signer = signature.New(cfg.HashType)
	client.payment = &PaymentService{
		transport:     client.transport,
		signer:        client.signer,
		merchantLogin: cfg.Login,
		password1:     cfg.Password1,
		isTest:        cfg.IsTest,
		paymentURL:    cfg.PaymentURL,
		invoiceURL:    cfg.PaymentInvoiceURL,
		paymentCurl:   cfg.PaymentCurlURL,
		jwtAPIURL:     cfg.JWTAPIURL,
		hashType:      cfg.HashType,
	}
	client.receipt = &ReceiptService{
		transport:      client.transport,
		signer:         client.signer,
		password1:      cfg.Password1,
		hashType:       cfg.HashType,
		secondCheckURL: cfg.SecondCheckURL,
		checkStatusURL: cfg.CheckStatusURL,
	}
	client.webService = &WebService{
		transport:     client.transport,
		signer:        client.signer,
		merchantLogin: cfg.Login,
		password2:     cfg.Password2,
		hashType:      cfg.HashType,
		webServiceURL: cfg.WebServiceURL,
		currenciesURL: cfg.CurrenciesURL,
	}
	client.status = &StatusService{
		transport:      client.transport,
		signer:         signature.New("md5"),
		merchantLogin:  cfg.Login,
		password1:      cfg.Password1,
		invoiceListURL: cfg.InvoiceListURL,
	}
	client.notify = &NotificationService{
		signer:    client.signer,
		password1: cfg.Password1,
		password2: cfg.Password2,
		hashType:  cfg.HashType,
	}

	return client, nil
}

func (c *Client) Payment() *PaymentService { return c.payment }
func (c *Client) Receipt() *ReceiptService { return c.receipt }
func (c *Client) WebService() *WebService  { return c.webService }
func (c *Client) Status() *StatusService   { return c.status }
func (c *Client) Notification() *NotificationService {
	return c.notify
}
