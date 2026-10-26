param name string
param location string = resourceGroup().location
param tags object = {}
param principalId string
param userPrincipalId string = ''

resource namespace 'Microsoft.ServiceBus/namespaces@2024-01-01' = {
  name: name
  location: location
  tags: tags
  sku: {
    name: 'Standard'
    tier: 'Standard'
  }
  properties: {
    minimumTlsVersion: '1.2'
    disableLocalAuth: true
    publicNetworkAccess: 'Enabled'
  }
}

resource queue 'Microsoft.ServiceBus/namespaces/queues@2024-01-01' = {
  parent: namespace
  name: 'orders'
  properties: {
    lockDuration: 'PT1M'
    defaultMessageTimeToLive: 'P1D'
    deadLetteringOnMessageExpiration: true
    maxDeliveryCount: 5
    requiresSession: false
  }
}

var receiverRole = '4f6d3b9b-027b-4f4c-9142-0e5a2a2247e0'
var senderRole = '69a216fc-b8fb-44d8-bc22-1f3c2cd27a39'

resource receiver 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(queue.id, principalId, receiverRole)
  scope: queue
  properties: {
    principalId: principalId
    principalType: 'ServicePrincipal'
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', receiverRole)
  }
}

resource sender 'Microsoft.Authorization/roleAssignments@2022-04-01' = if (!empty(userPrincipalId)) {
  name: guid(queue.id, userPrincipalId, senderRole)
  scope: queue
  properties: {
    principalId: userPrincipalId
    principalType: 'User'
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', senderRole)
  }
}

output fullyQualifiedNamespace string = '${namespace.name}.servicebus.windows.net'
