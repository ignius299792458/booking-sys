package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/ignius299792458/techkraft-ch-svr/model"
)

// BenchmarkHandleBooking benchmarks the HandleBooking function
func BenchmarkHandleBooking(b *testing.B) {
	setupTestHandlers()

	bookingOrder := model.BookingOrder{
		UserID:         "bench-user",
		Tier:           model.TierVIP,
		SeatNo:         1,
		Country:        "USA",
		ZipCode:        "10001",
		Currency:       "USD",
		IdempotencyKey: "bench-key",
		PaymentID:      "pay-bench",
		PaymentStatus:  model.PaymentStatusConfirmed,
	}

	_, _ = json.Marshal(bookingOrder)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		seatNo := uint32(1)
		for pb.Next() {
			// Use different seat numbers to avoid conflicts
			order := bookingOrder
			order.SeatNo = seatNo
			order.IdempotencyKey = "bench-key-" + string(rune(seatNo))
			seatNo++
			if seatNo > 100 {
				seatNo = 1
			}

			body, _ := json.Marshal(order)
			req := httptest.NewRequest(http.MethodPost, "/booking/ticket", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			HandleBooking(w, req)

			// Reset stores for next iteration
			if w.Code == http.StatusConflict {
				setupTestHandlers()
			}
		}
	})
}

// BenchmarkHandleAvailability benchmarks the HandleAvailability function
func BenchmarkHandleAvailability(b *testing.B) {
	setupTestHandlers()

	// Pre-populate with some bookings
	for i := uint32(1); i <= 10; i++ {
		bookingStore.RegisterBooking(model.BookingOrder{
			UserID:         "user-bench",
			Tier:           model.TierVIP,
			SeatNo:         i,
			IdempotencyKey: "key-bench",
			PaymentID:      "pay-bench",
			PaymentStatus:  model.PaymentStatusConfirmed,
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/booking/availability", nil)

	for b.Loop() {
		w := httptest.NewRecorder()
		HandleAvailability(w, req)
	}
}

// BenchmarkHandleAvailability_Parallel benchmarks HandleAvailability with parallel requests
func BenchmarkHandleAvailability_Parallel(b *testing.B) {
	setupTestHandlers()

	// Pre-populate with bookings
	for i := uint32(1); i <= 20; i++ {
		bookingStore.RegisterBooking(model.BookingOrder{
			UserID:         "user-bench",
			Tier:           model.TierVIP,
			SeatNo:         i,
			IdempotencyKey: "key-bench",
			PaymentID:      "pay-bench",
			PaymentStatus:  model.PaymentStatusConfirmed,
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/booking/availability", nil)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			HandleAvailability(w, req)
		}
	})
}

// BenchmarkHandleBooking_P95 measures p95 latency for HandleBooking
func BenchmarkHandleBooking_P95(b *testing.B) {
	setupTestHandlers()

	bookingOrder := model.BookingOrder{
		UserID:         "p95-user",
		Tier:           model.TierVIP,
		SeatNo:         1,
		Country:        "USA",
		ZipCode:        "10001",
		Currency:       "USD",
		IdempotencyKey: "p95-key",
		PaymentID:      "pay-p95",
		PaymentStatus:  model.PaymentStatusConfirmed,
	}

	const targetP95Ms = 500 // 500ms target
	const iterations = 1000
	latencies := make([]time.Duration, 0, iterations)

	b.ResetTimer()

	for i := 0; i < iterations; i++ {
		order := bookingOrder
		order.SeatNo = uint32(i%100 + 1)
		order.IdempotencyKey = "p95-key-" + string(rune(i))

		body, _ := json.Marshal(order)
		req := httptest.NewRequest(http.MethodPost, "/booking/ticket", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		start := time.Now()
		HandleBooking(w, req)
		duration := time.Since(start)

		latencies = append(latencies, duration)

		// Reset stores periodically to avoid conflicts
		if i%100 == 0 && i > 0 {
			setupTestHandlers()
		}
	}

	// Calculate p95
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})
	p95Index := int(float64(len(latencies)) * 0.95)
	p95Latency := latencies[p95Index]

	b.ReportMetric(float64(p95Latency.Nanoseconds())/1e6, "p95_ms")
	b.ReportMetric(float64(len(latencies)), "iterations")

	if p95Latency.Milliseconds() > targetP95Ms {
		b.Errorf("P95 latency %dms exceeds target %dms", p95Latency.Milliseconds(), targetP95Ms)
	}
}

// BenchmarkHandleAvailability_P95 measures p95 latency for HandleAvailability
func BenchmarkHandleAvailability_P95(b *testing.B) {
	setupTestHandlers()

	// Pre-populate with bookings
	for i := uint32(1); i <= 50; i++ {
		bookingStore.RegisterBooking(model.BookingOrder{
			UserID:         "user-p95",
			Tier:           model.TierVIP,
			SeatNo:         i,
			IdempotencyKey: "key-p95",
			PaymentID:      "pay-p95",
			PaymentStatus:  model.PaymentStatusConfirmed,
		})
	}

	const targetP95Ms = 500 // 500ms target
	const iterations = 1000
	latencies := make([]time.Duration, 0, iterations)

	req := httptest.NewRequest(http.MethodGet, "/booking/availability", nil)
	b.ResetTimer()

	for i := 0; i < iterations; i++ {
		w := httptest.NewRecorder()

		start := time.Now()
		HandleAvailability(w, req)
		duration := time.Since(start)

		latencies = append(latencies, duration)
	}

	// Calculate p95
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})
	p95Index := int(float64(len(latencies)) * 0.95)
	p95Latency := latencies[p95Index]

	b.ReportMetric(float64(p95Latency.Nanoseconds())/1e6, "p95_ms")
	b.ReportMetric(float64(len(latencies)), "iterations")

	if p95Latency.Milliseconds() > targetP95Ms {
		b.Errorf("P95 latency %dms exceeds target %dms", p95Latency.Milliseconds(), targetP95Ms)
	}
}

// BenchmarkHandleBooking_Concurrent measures concurrent booking performance
func BenchmarkHandleBooking_Concurrent(b *testing.B) {
	setupTestHandlers()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		userID := 0
		for pb.Next() {
			userID++
			seatNo := uint32(userID%100 + 1)

			order := model.BookingOrder{
				UserID:         "concurrent-user-" + string(rune(userID)),
				Tier:           model.TierVIP,
				SeatNo:         seatNo,
				Country:        "USA",
				ZipCode:        "10001",
				Currency:       "USD",
				IdempotencyKey: "concurrent-key-" + string(rune(userID)),
				PaymentID:      "pay-concurrent",
				PaymentStatus:  model.PaymentStatusConfirmed,
			}

			body, _ := json.Marshal(order)
			req := httptest.NewRequest(http.MethodPost, "/booking/ticket", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			HandleBooking(w, req)
		}
	})
}
