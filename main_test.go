package main

import (
	"context"
	"testing"

	"github.com/azure/azure-functions-golang-worker/sdk"
	"github.com/azure/azure-functions-golang-worker/sdk/bindings"
)

func TestProcessOrder(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		wantError  bool
	}{
		{"order", `{"orderId":"order-001","item":"coffee","quantity":2}`, false},
		{"missing id", `{"item":"coffee","quantity":2}`, true},
		{"missing item", `{"orderId":"order-001","quantity":2}`, true},
		{"zero quantity", `{"orderId":"order-001","item":"coffee","quantity":0}`, true},
		{"negative quantity", `{"orderId":"order-001","item":"coffee","quantity":-1}`, true},
		{"invalid JSON", `{`, true},
		{"null", `null`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := processOrder(context.Background(), bindings.ServiceBusMessage{Body: []byte(tc.body)})
			if (err != nil) != tc.wantError {
				t.Fatalf("error = %v, wantError = %v", err, tc.wantError)
			}
		})
	}
}

func TestServiceBusRegistration(t *testing.T) {
	count := 0
	newFunctionApp().GetRegisteredFunctions().Range(func(_, value any) bool {
		count++
		fn := value.(*sdk.RegisteredFunction)
		binding := fn.TriggerBinding()
		if fn.FuncName != "processOrder" || binding == nil || binding.ServiceBusBinding == nil {
			t.Fatal("missing Service Bus trigger")
		}
		b := binding.ServiceBusBinding
		if b.QueueName != "orders" || b.Connection != "ServiceBusConnection" || b.IsSessionsEnabled {
			t.Fatalf("unexpected queue configuration: %+v", b)
		}
		return true
	})
	if count != 1 {
		t.Fatalf("registered functions = %d, want 1", count)
	}
}
