package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/n-hlyustin/robokassa-sdk"
	"github.com/n-hlyustin/robokassa-sdk/examples/internal/shared"
)

func main() {
	client := shared.NewClient()
	invID := time.Now().Unix()

	url, err := client.Payment().SendJWT(context.Background(), robokassa.CreateInvoiceRequest{
		InvID:       invID,
		OutSum:      99.90,
		Description: "Тестовый платеж",
		Culture:     "ru",
		InvoiceItems: []robokassa.InvoiceItem{
			{
				Name:          "Товар",
				Quantity:      1,
				Cost:          99.90,
				Tax:           "none",
				PaymentMethod: "full_payment",
				PaymentObject: "service",
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(url)
}
