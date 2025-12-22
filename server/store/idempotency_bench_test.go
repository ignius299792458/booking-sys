package store

import (
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/ignius299792458/techkraft-ch-svr/model"
)

// BenchmarkHandleIdempotency benchmarks HandleIdempotency function
func BenchmarkHandleIdempotency(b *testing.B) {
	bucket := NewIdempotencyBucket().(*IDEMPOTENCY_BUCKET)

	bookingOrder := model.BookingOrder{
		UserID:         "bench-user",
		Tier:           model.TierVIP,
		SeatNo:         1,
		Status:         model.BookingStatusPending,
		IdempotencyKey: "bench-key",
		PaymentStatus:  model.PaymentStatusPending,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		order := bookingOrder
		order.IdempotencyKey = "bench-key-" + string(rune(i))
		_ = bucket.HandleIdempotency(order)
	}
}

// BenchmarkHandleIdempotency_Parallel benchmarks HandleIdempotency with parallel requests
func BenchmarkHandleIdempotency_Parallel(b *testing.B) {
	bucket := NewIdempotencyBucket().(*IDEMPOTENCY_BUCKET)

	bookingOrder := model.BookingOrder{
		UserID:         "bench-user",
		Tier:           model.TierVIP,
		SeatNo:         1,
		Status:         model.BookingStatusPending,
		IdempotencyKey: "bench-key",
		PaymentStatus:  model.PaymentStatusPending,
	}

	var mu sync.Mutex
	keyCounter := 0

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			keyCounter++
			keyID := keyCounter
			mu.Unlock()

			order := bookingOrder
			order.IdempotencyKey = "bench-key-" + string(rune(keyID))
			_ = bucket.HandleIdempotency(order)
		}
	})
}

// BenchmarkHandleIdempotency_P95 measures p95 latency for HandleIdempotency
func BenchmarkHandleIdempotency_P95(b *testing.B) {
	bucket := NewIdempotencyBucket().(*IDEMPOTENCY_BUCKET)

	bookingOrder := model.BookingOrder{
		UserID:         "p95-user",
		Tier:           model.TierVIP,
		SeatNo:         1,
		Status:         model.BookingStatusPending,
		IdempotencyKey: "p95-key",
		PaymentStatus:  model.PaymentStatusPending,
	}

	const targetP95Ms = 500 // 500ms target
	const iterations = 1000
	latencies := make([]time.Duration, 0, iterations)

	b.ResetTimer()

	for i := 0; i < iterations; i++ {
		order := bookingOrder
		order.IdempotencyKey = "p95-key-" + string(rune(i))

		start := time.Now()
		_ = bucket.HandleIdempotency(order)
		duration := time.Since(start)

		latencies = append(latencies, duration)
	}

	// Calculate p95
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})
	p95Index := int(float64(len(latencies)) * 0.95)
	if p95Index >= len(latencies) {
		p95Index = len(latencies) - 1
	}
	p95Latency := latencies[p95Index]

	b.ReportMetric(float64(p95Latency.Nanoseconds())/1e6, "p95_ms")
	b.ReportMetric(float64(len(latencies)), "iterations")

	if p95Latency.Milliseconds() > targetP95Ms {
		b.Errorf("P95 latency %dms exceeds target %dms", p95Latency.Milliseconds(), targetP95Ms)
	}
}

// BenchmarkHandleIdempotency_ExistingKey benchmarks with existing idempotency keys
func BenchmarkHandleIdempotency_ExistingKey(b *testing.B) {
	bucket := NewIdempotencyBucket().(*IDEMPOTENCY_BUCKET)

	// Pre-populate with idempotency keys
	for i := 0; i < 100; i++ {
		bucket.HandleIdempotency(model.BookingOrder{
			UserID:         "user-bench",
			Tier:           model.TierVIP,
			SeatNo:         uint32(i + 1),
			Status:         model.BookingStatusPending,
			IdempotencyKey: "existing-key-" + string(rune(i)),
			PaymentStatus:  model.PaymentStatusPending,
		})
	}

	bookingOrder := model.BookingOrder{
		UserID:         "user-bench",
		Tier:           model.TierVIP,
		SeatNo:         1,
		Status:         model.BookingStatusPending,
		IdempotencyKey: "existing-key-50", // Use existing key
		PaymentStatus:  model.PaymentStatusPending,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = bucket.HandleIdempotency(bookingOrder)
	}
}

// BenchmarkHandleIdempotency_ConcurrentSameKey benchmarks concurrent requests with same idempotency key
func BenchmarkHandleIdempotency_ConcurrentSameKey(b *testing.B) {
	bucket := NewIdempotencyBucket().(*IDEMPOTENCY_BUCKET)

	bookingOrder := model.BookingOrder{
		UserID:         "concurrent-user",
		Tier:           model.TierVIP,
		SeatNo:         1,
		Status:         model.BookingStatusPending,
		IdempotencyKey: "same-key",
		PaymentStatus:  model.PaymentStatusPending,
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = bucket.HandleIdempotency(bookingOrder)
		}
	})
}
