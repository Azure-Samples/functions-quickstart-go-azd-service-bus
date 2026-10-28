package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/azure/azure-functions-golang-worker/sdk"
	"github.com/azure/azure-functions-golang-worker/sdk/bindings"
	"github.com/azure/azure-functions-golang-worker/worker"
)

type Order struct {
	OrderID  string `json:"orderId"`
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
}

func main() {
	worker.Start(newFunctionApp())
}

func newFunctionApp() *sdk.App {
	app := sdk.FunctionApp()
	app.ServiceBusQueue("processOrder", processOrder,
		sdk.WithQueueName("orders"),
		sdk.WithConnection("ServiceBusConnection"),
	)
	return app
}

func processOrder(ctx context.Context, message bindings.ServiceBusMessage) error {
	var order Order
	if err := json.Unmarshal(message.Body, &order); err != nil {
		return fmt.Errorf("decode order: %w", err)
	}
	if order.OrderID == "" || order.Item == "" || order.Quantity <= 0 {
		return fmt.Errorf("orderId, item, and a positive quantity are required")
	}
	slog.InfoContext(ctx, "order processed",
		"order_id", order.OrderID,
		"item", order.Item,
		"quantity", order.Quantity,
		"message_id", message.MessageId,
		"delivery_count", message.DeliveryCount,
	)
	return nil
}
