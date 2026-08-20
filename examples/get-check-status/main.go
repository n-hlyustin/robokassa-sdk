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

	resp, err := client.Receipt().GetCheckStatus(context.Background(), robokassa.CheckStatusRequest{
		MerchantID: "your_merchant_id",
		ID:         "1001",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(shared.Pretty(resp))
}
