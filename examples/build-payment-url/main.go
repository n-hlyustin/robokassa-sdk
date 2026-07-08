package main

import (
	"fmt"
	"log"

	robokassa "github.com/robokassa/sdk-go-main"
	"github.com/robokassa/sdk-go-main/examples/internal/shared"
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
