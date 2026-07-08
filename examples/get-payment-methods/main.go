package main

import (
	"context"
	"fmt"
	"log"

	"github.com/robokassa/sdk-go-main/examples/internal/shared"
)

func main() {
	client := shared.NewClient()

	resp, err := client.WebService().GetPaymentMethods(context.Background(), "ru")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.RawXML)
	fmt.Println(shared.Pretty(resp.Groups))
	fmt.Println(shared.Pretty(resp.RawData))
}
