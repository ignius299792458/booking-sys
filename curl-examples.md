# API cURL Examples

## Get Availability

Get ticket availability and reserved seats for all tiers.

```bash
curl -X GET http://localhost:8080/booking/availability \
  -H "Content-Type: application/json"
```

### Expected Response:

```json
{
  "success": true,
  "message": "reserved seats retrieved successfully",
  "reservedSeats": {
    "VIP": [1, 5, 10],
    "FRONT_ROW": [31, 35, 40],
    "GA": [61, 65, 70]
  },
  "tiers": [
    {
      "tier": "VIP",
      "price": 10000,
      "available": 27,
      "reserved": [1, 5, 10]
    },
    {
      "tier": "FRONT_ROW",
      "price": 5000,
      "available": 27,
      "reserved": [31, 35, 40]
    },
    {
      "tier": "GA",
      "price": 1000,
      "available": 37,
      "reserved": [61, 65, 70]
    }
  ]
}
```

## Book Ticket

Create a new ticket booking.

```bash
curl -X POST http://localhost:8080/booking/ticket \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-123",
    "tier": "VIP",
    "seatNo": 2,
    "country": "USA",
    "zipCode": "10001",
    "currency": "USD",
    "idempotencyKey": "unique-key-123",
    "totalAmtInUSCent": 10000,
    "paymentID": "pay_123456",
    "paymentStatus": "CONFIRMED",
    "status": "PENDING"
  }'
```

### Expected Response:

```json
{
  "success": true,
  "message": "new booking successful",
  "booking": {
    "id": "uuid-here",
    "userId": "user-123",
    "tier": "VIP",
    "status": "CONFIRMED",
    "seatNo": 2,
    "country": "USA",
    "zipCode": "10001",
    "currency": "USD",
    "totalAmtInUSCent": 10000,
    "paymentID": "pay_123456",
    "paymentStatus": "CONFIRMED",
    "createdAt": "2024-01-01T12:00:00Z",
    "updatedAt": "2024-01-01T12:00:00Z"
  }
}
```

## Pretty Print JSON Response

To format the JSON response nicely, pipe to `jq`:

```bash
curl -X GET http://localhost:8080/booking/availability \
  -H "Content-Type: application/json" | jq
```

Or use Python:

```bash
curl -X GET http://localhost:8080/booking/availability \
  -H "Content-Type: application/json" | python -m json.tool
```

