# MEWS-Plugin
Integration for the Connecter API from MEWS

## For testing purposes curl this

Simulates Mews Guest Check-In function. 
https://mews-systems.gitbook.io/connector-api/use-cases/kiosk

```
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "EnterpriseId": "hotel-test-123",
    "IntegrationId": "mews-int-456",
    "Events": [
      {
        "Discriminator": "ServiceOrderUpdated",
        "Value": {
          "Id": "reservation-TEST-2025-11-20"
        }
      }
    ]
  }' && echo -e "\n✓ Test webhook sent"
  ```