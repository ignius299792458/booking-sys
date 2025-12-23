package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/ignius299792458/techkraft-ch-svr/model"
	"github.com/ignius299792458/techkraft-ch-svr/store"
)

// BenchmarkEndToEndBooking benchmarks the complete booking flow
func BenchmarkEndToEndBooking(b *testing.B) {
	// Setup fresh stores
	bookingStore := store.NewBookingStoreBucket()
	idempotencyStore := store.NewIdempotencyBucket()

	// Create a test server
	mux := http.NewServeMux()
	mux.HandleFunc("POST /booking/ticket", func(w http.ResponseWriter, r *http.Request) {
		// Mock handler for benchmarking
		var req model.BookingOrder
		json.NewDecoder(r.Body).Decode(&req)

		idempotentOrder := idempotencyStore.HandleIdempotency(req)
		_, err := bookingStore.RegisterBooking(idempotentOrder)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	bookingOrder := model.BookingOrder{
		UserID:         "e2e-user",
		Tier:           model.TierVIP,
		SeatNo:         1,
		Country:        "USA",
		ZipCode:        "10001",
		Currency:       "USD",
		IdempotencyKey: "e2e-key",
		PaymentID:      "pay-e2e",
		PaymentStatus:  model.PaymentStatusConfirmed,
	}

	for i := 0; b.Loop(); i++ {
		order := bookingOrder
		order.SeatNo = uint32(i%100 + 1)
		order.IdempotencyKey = "e2e-key-" + string(rune(i))

		body, _ := json.Marshal(order)
		resp, err := http.Post(server.URL+"/booking/ticket", "application/json", bytes.NewBuffer(body))
		if err == nil && resp != nil {
			resp.Body.Close()
		}
	}
}

// BenchmarkP95Latency measures p95 latency for booking endpoint
func BenchmarkP95Latency(b *testing.B) {
	const targetP95Ms = 500
	const iterations = 1000
	const concurrentUsers = 50

	bookingStore := store.NewBookingStoreBucket()
	idempotencyStore := store.NewIdempotencyBucket()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /booking/ticket", func(w http.ResponseWriter, r *http.Request) {
		var req model.BookingOrder
		json.NewDecoder(r.Body).Decode(&req)

		idempotentOrder := idempotencyStore.HandleIdempotency(req)
		_, err := bookingStore.RegisterBooking(idempotentOrder)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	latencies := make([]time.Duration, 0, iterations*concurrentUsers)
	var mu sync.Mutex

	b.ResetTimer()

	var wg sync.WaitGroup
	for user := 0; user < concurrentUsers; user++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			for i := 0; i < iterations/concurrentUsers; i++ {
				order := model.BookingOrder{
					UserID:         fmt.Sprintf("user-%d", userID),
					Tier:           model.TierVIP,
					SeatNo:         uint32((userID*iterations/concurrentUsers+i)%100 + 1),
					Country:        "USA",
					ZipCode:        "10001",
					Currency:       "USD",
					IdempotencyKey: fmt.Sprintf("key-%d-%d", userID, i),
					PaymentID:      "pay-bench",
					PaymentStatus:  model.PaymentStatusConfirmed,
				}

				body, _ := json.Marshal(order)
				start := time.Now()
				resp, err := http.Post(server.URL+"/booking/ticket", "application/json", bytes.NewBuffer(body))
				duration := time.Since(start)

				if err == nil && resp != nil {
					resp.Body.Close()
				}

				mu.Lock()
				latencies = append(latencies, duration)
				mu.Unlock()
			}
		}(user)
	}
	wg.Wait()

	// Calculate percentiles
	slices.Sort(latencies)

	p50Index := int(float64(len(latencies)) * 0.50)
	p95Index := int(float64(len(latencies)) * 0.95)
	p99Index := int(float64(len(latencies)) * 0.99)

	if p50Index >= len(latencies) {
		p50Index = len(latencies) - 1
	}
	if p95Index >= len(latencies) {
		p95Index = len(latencies) - 1
	}
	if p99Index >= len(latencies) {
		p99Index = len(latencies) - 1
	}

	p50Latency := latencies[p50Index]
	p95Latency := latencies[p95Index]
	p99Latency := latencies[p99Index]

	b.ReportMetric(float64(p50Latency.Nanoseconds())/1e6, "p50_ms")
	b.ReportMetric(float64(p95Latency.Nanoseconds())/1e6, "p95_ms")
	b.ReportMetric(float64(p99Latency.Nanoseconds())/1e6, "p99_ms")
	b.ReportMetric(float64(len(latencies)), "total_requests")

	if p95Latency.Milliseconds() > targetP95Ms {
		b.Errorf("P95 latency %dms exceeds target %dms", p95Latency.Milliseconds(), targetP95Ms)
	}

	fmt.Printf("\n=== Performance Summary ===\n")
	fmt.Printf("Total Requests: %d\n", len(latencies))
	fmt.Printf("P50 Latency: %dms\n", p50Latency.Milliseconds())
	fmt.Printf("P95 Latency: %dms (target: %dms)\n", p95Latency.Milliseconds(), targetP95Ms)
	fmt.Printf("P99 Latency: %dms\n", p99Latency.Milliseconds())
	fmt.Printf("========================\n")
}
