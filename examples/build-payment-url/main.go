package main

import (
	"fmt"
	"log"

	"github.com/n-hlyustin/robokassa-sdk"
	"github.com/n-hlyustin/robokassa-sdk/examples/internal/shared"
)

func main() {
	client := shared.NewClient()

	url, err := client.Payment().BuildPaymentURL(robokassa.CreatePaymentRequest{
		OutSum:      "149.00",
		InvID:       "1003",
		Description: "Тестовый платеж",
		Culture:     "ru",
		Email:       "test@example.com",
		Receipt: robokassa.Receipt{
			Items: []robokassa.ReceiptItem{
				{
					Name:          "Товар",
					Quantity:      1,
					Sum:           149.00,
					Tax:           "none",
					PaymentMethod: "full_payment",
					PaymentObject: "service",
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(url)
}
