package main

import (
	"context"
	"fmt"
	"log"

	"github.com/n-hlyustin/robokassa-sdk/examples/internal/shared"
)

func main() {
	client := shared.NewClient()

	resp, err := client.WebService().OpState(context.Background(), 1001)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.RawXML)
	fmt.Println(shared.Pretty(resp.State))
	fmt.Println(shared.Pretty(resp.RawData))
}
