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

	payment, err := client.Payment().SendForm(context.Background(), robokassa.CreatePaymentRequest{
		OutSum:      "99.90",
		InvID:       "1002",
		Description: "Тестовый платеж",
		Culture:     "ru",
		Email:       "test@example.com",
		Receipt: robokassa.Receipt{
			Items: []robokassa.ReceiptItem{
				{
					Name:          "Товар",
					Quantity:      1,
					Sum:           99.90,
					Tax:           "none",
					PaymentMethod: "full_payment",
					PaymentObject: "service",
				},
			},
		},
		ShpFields: map[string]string{
			"Shp_item": "digital",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(payment.InvoiceURL)
}
