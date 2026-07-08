package shared

import (
	"encoding/json"
	"log"
	"os"
	"time"

	robokassa "github.com/robokassa/sdk-go-main"
)

func NewClient() *robokassa.Client {
	client, err := robokassa.NewClient(robokassa.Config{
		Login:         os.Getenv("ROBOKASSA_LOGIN"),
		Password1:     os.Getenv("ROBOKASSA_PASSWORD1"),
		Password2:     os.Getenv("ROBOKASSA_PASSWORD2"),
		TestPassword1: os.Getenv("ROBOKASSA_TEST_PASSWORD1"),
		TestPassword2: os.Getenv("ROBOKASSA_TEST_PASSWORD2"),
		HashType:      getenv("ROBOKASSA_HASH_TYPE", "md5"),
		IsTest:        os.Getenv("ROBOKASSA_IS_TEST") == "1",
		HTTPTimeout:   15 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func Pretty(v interface{}) string {
	raw, _ := json.MarshalIndent(v, "", "  ")
	return string(raw)
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
