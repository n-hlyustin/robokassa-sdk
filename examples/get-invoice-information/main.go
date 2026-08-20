package main

import (
	"context"
	"fmt"
	"log"

	"github.com/n-hlyustin/robokassa-sdk"
	"github.com/n-hlyustin/robokassa-sdk/examples/internal/shared"
)

func main() {
	client := shared.NewClient()

	resp, err := client.Status().GetInvoiceInformationList(context.Background(), robokassa.InvoiceInformationListRequest{
		CurrentPage:     1,
		PageSize:        10,
		InvoiceStatuses: []string{"paid", "expired"},
		DateFrom:        "2024-01-01T00:00:00",
		DateTo:          "2026-12-31T23:59:59",
		InvoiceTypes:    []string{"onetime"},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(shared.Pretty(resp))
}
