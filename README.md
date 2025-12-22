# Concert Ticket Booking System

A full-stack ticket booking application built with Go backend and Next.js frontend, designed to prevent double-booking and support a global user base.

## Architecture Overview

### Backend (Go)

- **Language**: Go 1.25.1
- **Server**: Standard library `net/http`
- **Storage**: In-memory map-based store (can be replaced with PostgreSQL)
- **Concurrency Control**: Seat-level mutex locks using `sync.Map`
- **Idempotency**: Client-side idempotency keys with server-side deduplication

### Frontend (Next.js)

- **Framework**: Next.js 16.1.0 with App Router
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: React hooks

## Project Structure

```
techkraft-ch/
├── server/                 # Go backend
│   ├── cmd/               # Application entry point
│   ├── handlers/          # HTTP handlers
│   ├── model/             # Domain models and types
│   ├── router/            # Route definitions
│   ├── store/             # Data storage layer
│   └── utils/             # Utility functions
└── client/                # Next.js frontend
    ├── app/               # Next.js app router pages
    ├── components/        # React components
    ├── lib/               # API client and utilities
    └── types/             # TypeScript type definitions
```

## Setup Instructions

### Prerequisites

- Go 1.25.1 or later
- Node.js 18+ and npm
- (Optional) PostgreSQL for production

### Backend Setup

1. Navigate to the server directory:

```bash
cd server
```

2. Install dependencies:

```bash
go mod download
```

3. Run the server:

```bash
go run cmd/main.go
```

The server will start on `http://localhost:8080`

### Frontend Setup

1. Navigate to the client directory:

```bash
cd client
```

2. Install dependencies:

```bash
npm install
```

3. Create `.env.local` file in the client directory:

```bash
echo "NEXT_PUBLIC_API_URL=http://localhost:8080" > .env.local
```

Or manually create `.env.local` with:

```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

4. Run the development server:

```bash
npm run dev
```

The frontend will be available at `http://localhost:3000`

## API Endpoints

### GET `/booking/availability`

Returns available ticket counts per tier.

**Response:**

```json
{
  "success": true,
  "message": "availability retrieved successfully",
  "availability": {
    "VIP": 95,
    "FRONT_ROW": 4850,
    "GA": 44950
  },
  "tiers": [
    {
      "tier": "VIP",
      "price": 10000,
      "available": 95
    },
    ...
  ]
}
```

### POST `/booking/ticket`

Creates a new ticket booking.

**Request Body:**

```json
{
  "userId": "user123",
  "tier": "VIP",
  "seatNo": 42,
  "country": "USA",
  "zipCode": "10001",
  "currency": "USD",
  "idempotencyKey": "unique-key-123",
  "totalAmtInUSCent": 10000,
  "paymentID": "pay_123",
  "paymentStatus": "CONFIRMED",
  "status": "PENDING"
}
```

**Response:**

```json
{
  "success": true,
  "message": "new booking successful",
  "booking": {
    "id": "uuid",
    "userId": "user123",
    "tier": "VIP",
    "status": "CONFIRMED",
    ...
  }
}
```

## Ticket Tiers & Pricing

- **VIP**: $100.00 (seats 1-100)
- **Front Row**: $50.00 (seats 101-5000)
- **General Admission (GA)**: $10.00 (seats 5001-50000)

Total capacity: 50,000 seats

## Design Decisions & Trade-offs

### 1. Concurrency & Double-Booking Prevention

**Implementation:**

- **Seat-level locking**: Each seat has its own mutex lock (`sync.Map` of mutexes)
- **Critical section**: The check-and-reserve operation is atomic within the seat's lock
- **Idempotency**: Client-generated idempotency keys prevent duplicate bookings from retries

**Code Location:**

- `server/store/persist.go`: `RegisterBooking()` method uses seat-level locks
- `server/store/idempotency.go`: Handles idempotency key deduplication

**Trade-offs:**

- ✅ Fine-grained locking allows high concurrency (different seats can be booked simultaneously)
- ✅ Prevents race conditions at the seat level
- ⚠️ In-memory storage means data is lost on restart (acceptable for demo, needs persistence for production)

**How it prevents double-booking:**

```go
// Acquire seat-specific lock
seatBookingLock := b.getSeatLock(bookingOrderData.SeatNo)
seatBookingLock.Lock()
defer seatBookingLock.Unlock()

// Check if seat is already booked (atomic within lock)
if _, exists := b.BOOKING_STORE[bookingOrderData.SeatNo]; exists {
    return model.Booking{}, errors.New("seat already booked")
}

// Reserve the seat
b.BOOKING_STORE[newBooking.SeatNo] = newBooking
```

### 2. Storage Layer

**Current Implementation:**

- In-memory map-based storage (`map[uint32]model.Booking`)
- Fast lookups O(1) for seat availability checks
- No persistence (data lost on restart)

**Production Migration Path:**

- Replace `BOOKING_STORE` with PostgreSQL table
- Use database transactions with row-level locking (`SELECT FOR UPDATE`)
- Maintain the same interface (`BookingStore`) for easy migration

**Trade-offs:**

- ✅ Simple and fast for development/demo
- ✅ No database setup required
- ⚠️ Not production-ready (needs PostgreSQL for persistence)

### 3. Idempotency Handling

**Implementation:**

- Client generates idempotency keys using timestamp + random string
- Keys stored in browser `sessionStorage` to survive page refreshes
- Server maintains idempotency store to prevent duplicate processing

**Trade-offs:**

- ✅ Prevents duplicate bookings from network retries
- ✅ Client-side key generation reduces server load
- ⚠️ SessionStorage-based keys don't survive browser restarts (acceptable for this use case)

### 4. API Design

**RESTful endpoints:**

- `GET /booking/availability` - Read-only, cacheable
- `POST /booking/ticket` - Idempotent (via idempotency key)

**Trade-offs:**

- ✅ Simple and intuitive
- ✅ Follows REST conventions
- ⚠️ Could add PATCH for status updates, but current design is sufficient

### 5. Frontend Architecture

**Component Structure:**

- `TicketCatalog`: Displays available tickets
- `BookingForm`: Handles booking submission
- Separation of concerns: API calls in `lib/api.ts`, types in `types/index.ts`

**Trade-offs:**

- ✅ Simple component hierarchy
- ✅ Easy to test and maintain
- ⚠️ Could add state management (Redux/Zustand) for complex state, but hooks are sufficient

## Scaling to Production Requirements

### Availability Target: 99.99% (Four Nines)

**Design Approach:**

1. **Multi-Region Deployment**

   - Deploy backend instances in multiple regions (US-East, EU-West, Asia-Pacific)
   - Use a global load balancer (AWS Global Accelerator, Cloudflare)
   - Database replication: Primary in one region, read replicas in others

2. **High Availability Components**

   - **Load Balancer**: Health checks, automatic failover
   - **Application Servers**: Multiple instances behind load balancer
   - **Database**: PostgreSQL with streaming replication, automatic failover
   - **Monitoring**: Real-time health checks, alerting on failures

3. **Disaster Recovery**
   - Automated backups (hourly snapshots)
   - Cross-region database replication
   - RTO (Recovery Time Objective): < 5 minutes
   - RPO (Recovery Point Objective): < 1 minute

**Estimated Downtime:** ~52 minutes/year (99.99% uptime)

### Scale: ~1M DAU, ~50K Concurrent Users

**Architecture Changes:**

1. **Database**

   - PostgreSQL with connection pooling (PgBouncer)
   - Read replicas for availability queries (can scale horizontally)
   - Partitioning by seat ranges or booking dates
   - Indexes on `seatNo`, `userId`, `createdAt`

2. **Caching Layer**

   - Redis for availability counts (updated on each booking)
   - Cache invalidation: Write-through on booking creation
   - TTL: 5 seconds (balance between freshness and load)

3. **Application Servers**

   - Horizontal scaling: 10-20 instances behind load balancer
   - Stateless design (all state in database)
   - Auto-scaling based on CPU/memory/request rate

4. **API Rate Limiting**

   - Per-user rate limits (e.g., 10 bookings/minute)
   - Per-IP rate limits to prevent abuse
   - Implemented at load balancer or API gateway level

5. **Queue System** (Optional)
   - For peak traffic: Queue booking requests
   - Process asynchronously to handle bursts
   - Trade-off: Slight delay for better throughput

**Estimated Infrastructure:**

- Database: PostgreSQL (r6g.2xlarge equivalent) with 3 read replicas
- Application: 10-20 instances (t3.medium equivalent)
- Cache: Redis cluster (cache.r6g.large equivalent)
- Load Balancer: Application Load Balancer

### Performance: p95 < 500ms

**Optimization Strategies:**

1. **Database Optimization**

   - Indexes: `CREATE INDEX idx_seat_no ON bookings(seat_no)`
   - Connection pooling: 50-100 connections per instance
   - Prepared statements for common queries

2. **Caching**

   - Availability counts cached in Redis (5s TTL)
   - Reduces database load for GET requests
   - Write-through cache updates on booking

3. **Code Optimizations**

   - Seat-level locks minimize lock contention
   - Fast path: Check availability before acquiring lock
   - Batch operations where possible

4. **Monitoring & Profiling**
   - APM tools (Datadog, New Relic) to identify bottlenecks
   - Database query profiling
   - Distributed tracing for request flow

**Expected Performance:**

- GET `/booking/availability`: < 50ms (cached)
- POST `/booking/ticket`: < 200ms (with database)
- p95 target: < 500ms ✅

## Testing Strategy

### Unit Tests

- Store layer: Test seat locking and double-booking prevention
- Handlers: Test request validation and error handling
- Utils: Test price calculation and validation

### Integration Tests

- End-to-end booking flow
- Concurrent booking attempts (race condition testing)
- Idempotency key handling

### Load Tests

- Simulate 50K concurrent users
- Measure p95 latency
- Verify no double-bookings under load

## Future Enhancements

1. **Payment Integration**

   - Integrate with payment providers (Stripe, PayPal)
   - Handle payment webhooks
   - Update booking status based on payment result

2. **User Authentication**

   - JWT-based authentication
   - User profiles and booking history

3. **Real-time Updates**

   - WebSocket for live availability updates
   - Notify users when tickets become available

4. **Analytics**

   - Track booking patterns
   - Popular seat ranges
   - Revenue analytics

5. **Email Notifications**
   - Booking confirmations
   - Reminders before event

## License

This project is part of a take-home assignment for Techkraft Inc.
