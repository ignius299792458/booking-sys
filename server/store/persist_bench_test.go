package store

import (
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/ignius299792458/techkraft-ch-svr/model"
)

// BenchmarkRegisterBooking benchmarks RegisterBooking function
func BenchmarkRegisterBooking(b *testing.B) {
	bucket := NewBookingStoreBucket().(*BOOKING_STORE_BUCKET)

	bookingOrder := model.BookingOrder{
		UserID:           "bench-user",
		Tier:             model.TierVIP,
		SeatNo:           1,
		Status:           model.BookingStatusPending,
		IdempotencyKey:   "bench-key",
		Country:          "USA",
		ZipCode:          "10001",
		Currency:         "USD",
		TotalAmtInUSCent: 10000,
		PaymentID:        "pay-bench",
		PaymentStatus:    model.PaymentStatusConfirmed,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		order := bookingOrder
		order.SeatNo = uint32(i%100 + 1)
		order.IdempotencyKey = "bench-key-" + string(rune(i))

		_, err := bucket.RegisterBooking(order)
		if err != nil && err.Error() != "seat already booked" {
			b.Fatalf("Unexpected error: %v", err)
		}

		// Reset bucket periodically to avoid all seats being booked
		if i%50 == 0 && i > 0 {
			bucket = NewBookingStoreBucket().(*BOOKING_STORE_BUCKET)
		}
	}
}

// BenchmarkRegisterBooking_Parallel benchmarks RegisterBooking with parallel requests
func BenchmarkRegisterBooking_Parallel(b *testing.B) {
	bucket := NewBookingStoreBucket().(*BOOKING_STORE_BUCKET)

	bookingOrder := model.BookingOrder{
		UserID:           "bench-user",
		Tier:             model.TierVIP,
		SeatNo:           1,
		Status:           model.BookingStatusPending,
		IdempotencyKey:   "bench-key",
		Country:          "USA",
		ZipCode:          "10001",
		Currency:         "USD",
		TotalAmtInUSCent: 10000,
		PaymentID:        "pay-bench",
		PaymentStatus:    model.PaymentStatusConfirmed,
	}

	var mu sync.Mutex
	seatCounter := uint32(1)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			seatNo := seatCounter
			seatCounter++
			if seatCounter > 100 {
				seatCounter = 1
			}
			mu.Unlock()

			order := bookingOrder
			order.SeatNo = seatNo
			order.IdempotencyKey = "bench-key-" + string(rune(seatNo))

			_, err := bucket.RegisterBooking(order)
			if err != nil && err.Error() != "seat already booked" {
				// Ignore double booking errors in parallel benchmark
			}
		}
	})
}

// BenchmarkRegisterBooking_P95 measures p95 latency for RegisterBooking
func BenchmarkRegisterBooking_P95(b *testing.B) {
	bucket := NewBookingStoreBucket().(*BOOKING_STORE_BUCKET)

	bookingOrder := model.BookingOrder{
		UserID:           "p95-user",
		Tier:             model.TierVIP,
		SeatNo:           1,
		Status:           model.BookingStatusPending,
		IdempotencyKey:   "p95-key",
		Country:          "USA",
		ZipCode:          "10001",
		Currency:         "USD",
		TotalAmtInUSCent: 10000,
		PaymentID:        "pay-p95",
		PaymentStatus:    model.PaymentStatusConfirmed,
	}

	const targetP95Ms = 500 // 500ms target
	const iterations = 1000
	latencies := make([]time.Duration, 0, iterations)

	b.ResetTimer()

	for i := 0; i < iterations; i++ {
		order := bookingOrder
		order.SeatNo = uint32(i%100 + 1)
		order.IdempotencyKey = "p95-key-" + string(rune(i))

		start := time.Now()
		_, err := bucket.RegisterBooking(order)
		duration := time.Since(start)

		if err != nil && err.Error() != "seat already booked" {
			// Skip errors for latency calculation
			continue
		}

		latencies = append(latencies, duration)

		// Reset bucket periodically
		if i%100 == 0 && i > 0 {
			bucket = NewBookingStoreBucket().(*BOOKING_STORE_BUCKET)
		}
	}

	// Calculate p95
	if len(latencies) > 0 {
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
}

// BenchmarkGetReservedSeats benchmarks GetReservedSeats function
func BenchmarkGetReservedSeats(b *testing.B) {
	bucket := NewBookingStoreBucket().(*BOOKING_STORE_BUCKET)

	// Pre-populate with bookings
	for i := uint32(1); i <= 50; i++ {
		bucket.RegisterBooking(model.BookingOrder{
			UserID:         "user-bench",
			Tier:           model.TierVIP,
			SeatNo:         i,
			IdempotencyKey: "key-bench",
			PaymentID:      "pay-bench",
			PaymentStatus:  model.PaymentStatusConfirmed,
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = bucket.GetReservedSeats()
	}
}

// BenchmarkGetReservedSeats_Parallel benchmarks GetReservedSeats with parallel requests
func BenchmarkGetReservedSeats_Parallel(b *testing.B) {
	bucket := NewBookingStoreBucket().(*BOOKING_STORE_BUCKET)

	// Pre-populate with bookings
	for i := uint32(1); i <= 50; i++ {
		bucket.RegisterBooking(model.BookingOrder{
			UserID:         "user-bench",
			Tier:           model.TierVIP,
			SeatNo:         i,
			IdempotencyKey: "key-bench",
			PaymentID:      "pay-bench",
			PaymentStatus:  model.PaymentStatusConfirmed,
		})
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = bucket.GetReservedSeats()
		}
	})
}

// BenchmarkGetReservedSeats_P95 measures p95 latency for GetReservedSeats
func BenchmarkGetReservedSeats_P95(b *testing.B) {
	bucket := NewBookingStoreBucket().(*BOOKING_STORE_BUCKET)

	// Pre-populate with varying number of bookings
	for i := uint32(1); i <= 50; i++ {
		bucket.RegisterBooking(model.BookingOrder{
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

	b.ResetTimer()

	for i := 0; i < iterations; i++ {
		start := time.Now()
		_ = bucket.GetReservedSeats()
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

// BenchmarkGetBooking benchmarks GetBooking function
func BenchmarkGetBooking(b *testing.B) {
	bucket := NewBookingStoreBucket().(*BOOKING_STORE_BUCKET)

	// Pre-populate with a booking
	bucket.RegisterBooking(model.BookingOrder{
		UserID:         "user-bench",
		Tier:           model.TierVIP,
		SeatNo:         1,
		IdempotencyKey: "key-bench",
		PaymentID:      "pay-bench",
		PaymentStatus:  model.PaymentStatusConfirmed,
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = bucket.GetBooking(1)
	}
}

// BenchmarkGetBooking_Parallel benchmarks GetBooking with parallel requests
func BenchmarkGetBooking_Parallel(b *testing.B) {
	bucket := NewBookingStoreBucket().(*BOOKING_STORE_BUCKET)

	// Pre-populate with bookings
	for i := uint32(1); i <= 10; i++ {
		bucket.RegisterBooking(model.BookingOrder{
			UserID:         "user-bench",
			Tier:           model.TierVIP,
			SeatNo:         i,
			IdempotencyKey: "key-bench",
			PaymentID:      "pay-bench",
			PaymentStatus:  model.PaymentStatusConfirmed,
		})
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		seatNo := uint32(1)
		for pb.Next() {
			_, _ = bucket.GetBooking(seatNo)
			seatNo++
			if seatNo > 10 {
				seatNo = 1
			}
		}
	})
}
