package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

func main() {
	if err := send(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func send() error {
	namespace := os.Getenv("SERVICEBUS_NAMESPACE")
	if namespace == "" {
		return fmt.Errorf("set SERVICEBUS_NAMESPACE to the fully qualified namespace")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}
	client, err := azservicebus.NewClient(namespace, credential, nil)
	if err != nil {
		return err
	}
	defer client.Close(context.Background())
	sender, err := client.NewSender("orders", nil)
	if err != nil {
		return err
	}
	defer sender.Close(context.Background())
	contentType := "application/json"
	messageID := fmt.Sprintf("sample-%d", time.Now().UnixNano())
	if err := sender.SendMessage(ctx, &azservicebus.Message{
		Body:        []byte(`{"orderId":"order-001","item":"coffee","quantity":2}`),
		ContentType: &contentType,
		MessageID:   &messageID,
	}, nil); err != nil {
		return err
	}
	fmt.Printf("Sent order-001 to orders (message %s).\n", messageID)
	return nil
}
