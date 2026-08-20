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

	payload := robokassa.SecondCheckRequest{
		"merchantId": "your_merchant_id",
		"id":         "1001",
		"originId":   "1001",
		"operation":  "sell",
		"sno":        "osn",
		"items": []map[string]interface{}{
			{
				"name":     "Товар",
				"quantity": 1,
				"sum":      99.90,
				"tax":      "none",
			},
		},
	}

	resp, err := client.Receipt().SendSecondCheck(context.Background(), payload)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp)
}
