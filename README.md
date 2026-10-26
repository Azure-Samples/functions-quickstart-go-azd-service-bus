<!--
---
name: Azure Functions Go Service Bus Trigger using Azure Developer CLI
description: Process orders from Service Bus in Go on Azure Functions Flex Consumption using managed identity.
page_type: sample
languages:
- azdeveloper
- go
products:
- azure
- azure-functions
- azure-service-bus
urlFragment: functions-quickstart-go-azd-service-bus
---
-->

# Azure Functions Go Service Bus Trigger

This sample follows the [Go timer quickstart](https://github.com/Azure-Samples/functions-quickstart-go-azd-timer)
and deploys to Linux Flex Consumption with `azd`. Go worker packages are preview
packages. The `processOrder` function consumes JSON orders from the `orders` queue:

```json
{"orderId":"order-001","item":"coffee","quantity":2}
```

The handler validates the order and logs its fields and message metadata.
Returning `nil` allows the Functions host to complete the message automatically.
Returning an error abandons it for redelivery; the queue's maximum delivery count
is five, after which Service Bus moves it to the dead-letter queue. Expired
messages are also dead-lettered (the default TTL is one day).
Delivery is at least once: production handlers should make side effects
idempotent, for example by recording the order ID transactionally.

## Prerequisites

- [Go 1.25 or later](https://go.dev/dl/)
- [Azure Functions Core Tools 4.12 or later](https://learn.microsoft.com/azure/azure-functions/functions-run-local)
- [Azure CLI](https://learn.microsoft.com/cli/azure/install-azure-cli)
- [Azure Developer CLI](https://learn.microsoft.com/azure/developer/azure-developer-cli/install-azd)
- Azure subscription with permissions to provision resources and assign roles
- [Azurite](https://learn.microsoft.com/azure/storage/common/storage-use-azurite) for local host storage
- Optional: VS Code with the Azure Functions and Go extensions

## Initialize and deploy

```powershell
git clone https://github.com/Azure-Samples/functions-quickstart-go-azd-service-bus.git
cd functions-quickstart-go-azd-service-bus
az login
azd auth login
azd env new servicebus-dev
azd env set AZURE_PRINCIPAL_ID (az ad signed-in-user show --query id -o tsv)
azd up
```

Infrastructure includes a Standard Service Bus namespace and queue, managed
identity, host storage, Application Insights, and a VNet with private endpoints
for **host storage** by default. The Service Bus endpoint is public and secured
by Entra ID; SAS authentication is disabled. Standard Service Bus has ongoing
charges.

The app identity has **Azure Service Bus Data Receiver** scoped to the queue;
the optional test user has **Azure Service Bus Data Sender**. The trigger uses an
identity-based connection, not a connection string.

To disable the host storage VNet for development:

```shell
azd env set VNET_ENABLED false
azd up
```

## Send an order and verify

The Go publisher uses `DefaultAzureCredential`, including your Azure CLI login:

```powershell
$deployment = azd env get-values --output json | ConvertFrom-Json
$env:SERVICEBUS_NAMESPACE = $deployment.SERVICEBUS_NAMESPACE
go run ./cmd/send
```

Allow a few minutes for initial RBAC propagation. In the function app's
Application Insights **Logs**, run:

```kusto
traces
| where timestamp > ago(15m)
| where message contains "order processed"
| project timestamp, message, customDimensions
```

You should see `order_id=order-001`, `item=coffee`, and `quantity=2`.
Use Service Bus Explorer in the Azure portal to inspect or resubmit dead-lettered
messages; dead-letter queues do not automatically retry messages.

## Run locally

Use a real Azure Service Bus development namespace and an identity with Data
Receiver on `orders`. The sender role alone does not authorize a local listener;
assign Data Receiver to your development identity separately.

```powershell
Copy-Item local.settings.example.json local.settings.json
```

Replace `ServiceBusConnection__fullyQualifiedNamespace` with your namespace
hostname. Start Azurite in another terminal, sign in with `az login`, and run
`func start`. The host builds the Go app automatically; no `function.json` is
needed. Azurite emulates host storage, not Service Bus.
Local and deployed listeners on the same queue compete for messages: stop one
listener or use a separate development queue.

## Tests, redeployment, and cleanup

```shell
go test ./...
go vet ./...
azd up
azd down
```

Tests cover valid orders, malformed payloads, missing fields, non-positive
quantities, and trigger configuration.
