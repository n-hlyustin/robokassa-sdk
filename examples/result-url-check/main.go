package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	robokassa "github.com/robokassa/sdk-go-main"
	"github.com/robokassa/sdk-go-main/examples/internal/shared"
)

func main() {
	client := shared.NewClient()
	addr := os.Getenv("ROBOKASSA_RESULT_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	mux := http.NewServeMux()
	mux.Handle("/result", client.Notification().ResultURLHandler(func(w http.ResponseWriter, r *http.Request, notification robokassa.ResultNotification) error {
		// В production-коде найдите заказ по notification.InvoiceID, сверьте
		// notification.OutSum и ShpFields, затем отметьте заказ как оплаченный.
		log.Printf("valid ResultURL: inv_id=%s out_sum=%s shp=%v", notification.InvoiceID, notification.OutSum, notification.ShpFields)
		return nil
	}))

	log.Printf("listening on http://%s/result", displayAddr(addr))
	log.Fatal(http.ListenAndServe(addr, mux))
}

func displayAddr(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "localhost" + addr
	}
	return addr
}
