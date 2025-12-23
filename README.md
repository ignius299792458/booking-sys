# Concert Ticket Booking System

An end-to-end ticket booking platform implemented with a Go backend and a Next.js frontend, intentionally designed for **low-latency performance, strong transactional guarantees, and global availability**. The architecture _prioritizes concurrency control and idempotency_ to ensure tickets are never **double-booked**, regardless of traffic spikes or retry behavior.

## Important Note on Requirements vs Implementation

**If some requirements don't match the implementation, please note that several assumptions were made during development.** The primary focus was on achieving **p99 latency < 300ms**, solving the **double-booking issue**, and other **mission-critical factors** like concurrency control and idempotency. Some code may be implemented in a simplified or non-optimal way due to this performance-first approach.

## Assumptions Made and Implemented in This Project

This section documents all assumptions made during development that may differ from the original requirements or best practices. These assumptions were made to prioritize performance, double-booking prevention, and mission-critical reliability.

### 1. Seat Capacity and Tier Ranges

**Assumption:** Reduced seat capacity 100 seats for development and testing purposes.

**Actual Implementation:**

- VIP: seats 1-30 (30 seats)
- Front Row: seats 31-60 (30 seats)
- General Admission (GA): seats 61-100 (40 seats)
- **Total: 100 seats**

**Location:**

- `server/store/persist.go`: `TOTAL_SEAT = 100`
- `server/handlers/handler.go`: Tier seat ranges
- `client/types/index.ts`: `getSeatRangeForTier()` function

### 2. Storage Layer - In-Memory Instead of Database

**Assumption:** Used in-memory map-based storage instead of PostgreSQL or other persistent database.

**Implementation:**

- `map[uint32]model.Booking` for booking storage
- `map[string]model.BookingOrder` for idempotency storage
- ⚠️ NOTE: Data is lost on server restart

**Reasons:**

- To reflect skills about concurrency and race-conditions handling
- To express effort made on how I graunteed zero double-booking.
- Eliminates database connection overhead for completing task
- Easy Delivery of Assignment
- Swift Examiner Running Environment setup

**⚠️ NOTE:** Not production-ready; requires database migration for persistence.

**Location:** `server/store/persist.go`, `server/store/idempotency.go`

### 3. User Authentication - Auto-Generated UUIDs

**Assumption:** No real user authentication system; userId is auto-generated UUID on frontend.

**Implementation:**

- Frontend generates UUID using `crypto.randomUUID()` or fallback algorithm
- No login/logout functionality
- No user session management
- No user profile or booking history

**Location:** `client/components/BookingForm.tsx`

### 4. Payment Processing - Simulated/Mocked

**Assumption:** Payment processing is simulated; payment status is passed directly in the request body.

**Implementation:**

- Client directly populates random `paymentID` and `paymentStatus` in booking request, based on booking status is handled on server side
- No actual payment gateway integration
- No payment webhooks or callbacks

**Location:** `server/handlers/handler.go`, `server/store/persist.go`

### 5. Idempotency Key Storage - SessionStorage

**Assumption:** Idempotency keys stored in browser `sessionStorage` instead of server-side generation or persistent storage.

**Implementation:**

- Client generates idempotency keys: `tier-seatNo-userId`
- Keys stored in `sessionStorage` (lost on browser close)
- Server maintains in-memory idempotency store and use as cache

  **Location:** `client/lib/api.ts` (idempotency key functions), `server/store/idempotency.go`

## Performance Target - p99 < 300ms (vs p95 < 500ms)

Focused on achieving p99 latency < 300ms, which is stricter than typical p95 < 500ms requirement.

**Implementation:**

- Seat-level locking for minimal contention
- In-memory storage for zero database latency
- Fast validation checks before acquiring locks
- Optimized critical sections

## Code Quality Note

**If Some code is implemented in a simplified or non-optimal way** because the primary focus was on:

1. **Performance (p99 < 300ms)**

2. **Double-Booking Prevention**

3. **Mission-Critical Factors**: Focus on reliability of core booking flow:
   - Concurrency control is solid
   - Race conditions are prevented
   - But non-critical features may be minimal or missing

**Examples of simplified implementations:**

- Error messages are basic (not user-friendly)
- No comprehensive input sanitization
- Limited logging and monitoring
- Basic validation (sufficient but not exhaustive)
- No rate limiting or abuse prevention
- Simplified response structures

---

A full-stack ticket booking application built with Go backend and Next.js frontend, designed to prevent double-booking and support a global user base.

## Architecture Overview

### Backend (Go)

- **Language**: Go 1.25.1
- **Server**: Standard library `net/http`
- **Storage**: In-memory map-based store
- **Concurrency Control**: Seat-level mutex locks using `sync.Map`
- **Idempotency**: Client-side idempotency keys and server-side deduplication

### Frontend (Next.js)

- **Framework**: Next.js 16.1.0 with App Router, TS
- **Styling**: Tailwind CSS
- **State Management**: no external libraries, simple react-hooks are used

## Project Structure

```shell
# tree techkraft-ch

techkraft-ch/
├── server/                # Go backend
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
# terminal
go run ./cmd

# can be run using debugger config from
./.vscode/launch.json

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

3. Run the development server:

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
  "tiers": [
    {
      "tier": "VIP",
      "price": 10000,
      "totalSeats": 30,
      "reservedCount": 1,
      "availableList": [1,3,.....30]
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
  "paymentID": "pay_123",
  "paymentStatus": "CONFIRMED"
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

## Concert Ticket Booking Design Decisions & Trade-offs

### 1. Concurrency & Double-Booking Prevention

**Implementation:**

- **Seat-level locking**: Each seat has its own mutex lock (`sync.Map` of mutexes)
- **Critical section**: The check-and-reserve operation is atomic within the seat's lock
- **Idempotency**: Client-generated idempotency keys prevent duplicate bookings from retries many cases

**Code Location:**

- `server/store/persist.go`: `RegisterBooking()` method uses seat-level locks
- `server/store/idempotency.go`: Handles idempotency key deduplication and minimal caching system too

**Trade-offs:**

- ✅ Fine-grained locking allows high concurrency (different seats can be booked simultaneously) rather than locking whole booking_collection_bucket
- ✅ Prevents race conditions at the seat level
- ⚠️ In-memory storage means data is lost on restart (acceptable for demo, needs persistence for production)

**How it prevents double-booking:**

```go
// File: server/store/persist.go Func: RegisterBooking(bookingOrderData)

/*
   * Acquire seat-specific lock
   * Check if seat is already booked (atomic within lock)
   * Reserve the seat
*/

```

### 2. Storage Layer

**Current Implementation:**

- In-memory map-based storage
- Fast lookups O(1) for seat availability checks
- In memory storage

**Trade-offs:**

- ✅ Simple and fast for development/demo
- ✅ No database setup required
- ⚠️ Not production-ready

### 3. Idempotency Handling

**Implementation:**

- Client generates idempotency keys using timestamp + random string
- Keys stored in browser `sessionStorage` to survive page refreshes
- Server maintains idempotency store to prevent duplicate processing

**Trade-offs:**

- ✅ Prevents duplicate bookings from network retries
- ✅ Client-side key generation reduces server load

---

---

# Non-Funtional Requirements

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

   - Indexing
   - Connection pooling: 50-100 connections per instance
   - Prepared statements for common queries

2. **Caching**

   - Availability counts cached in Redis (5s TTL) and reserved collection caching
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

# BENCHMARK

Assignment Reviewer is requested to find Benchmark Result at `BENCHMARK_RESULTS.md`
