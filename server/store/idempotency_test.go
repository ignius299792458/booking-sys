package store

import (
	"sync"
	"testing"

	"github.com/ignius299792458/techkraft-ch-svr/model"
)

func TestHandleIdempotency(t *testing.T) {
	tests := []struct {
		name           string
		bookingOrder   model.BookingOrder
		expectedReturn model.BookingOrder
		setupFunc      func(*IDEMPOTENCY_BUCKET)
		validateFunc   func(*testing.T, *IDEMPOTENCY_BUCKET, model.BookingOrder)
	}{
		{
			name: "new idempotency key - stores and returns order",
			bookingOrder: model.BookingOrder{
				UserID:         "user-1",
				Tier:           model.TierVIP,
				SeatNo:         1,
				Status:         model.BookingStatusPending,
				IdempotencyKey: "key-new",
				PaymentStatus:  model.PaymentStatusPending,
			},
			expectedReturn: model.BookingOrder{
				UserID:         "user-1",
				Tier:           model.TierVIP,
				SeatNo:         1,
				Status:         model.BookingStatusPending,
				IdempotencyKey: "key-new",
				PaymentStatus:  model.PaymentStatusPending,
			},
			validateFunc: func(t *testing.T, ib *IDEMPOTENCY_BUCKET, result model.BookingOrder) {
				stored, exists := ib.IDEMPOTENCY_STORE["key-new"]
				if !exists {
					t.Error("Expected idempotency key to be stored")
				}
				if stored.UserID != "user-1" {
					t.Errorf("Expected stored UserID 'user-1', got '%s'", stored.UserID)
				}
			},
		},
		{
			name: "existing idempotency key - returns stored order",
			bookingOrder: model.BookingOrder{
				UserID:         "user-2",
				Tier:           model.TierVIP,
				SeatNo:         2,
				Status:         model.BookingStatusPending,
				IdempotencyKey: "key-existing",
				PaymentStatus:  model.PaymentStatusPending,
			},
			expectedReturn: model.BookingOrder{
				UserID:         "user-original",
				Tier:           model.TierFrontRow,
				SeatNo:         31,
				Status:         model.BookingStatusPending,
				IdempotencyKey: "key-existing",
				PaymentStatus:  model.PaymentStatusPending,
			},
			setupFunc: func(ib *IDEMPOTENCY_BUCKET) {
				ib.IDEMPOTENCY_STORE["key-existing"] = model.BookingOrder{
					UserID:         "user-original",
					Tier:           model.TierFrontRow,
					SeatNo:         31,
					Status:         model.BookingStatusPending,
					IdempotencyKey: "key-existing",
					PaymentStatus:  model.PaymentStatusPending,
				}
			},
			validateFunc: func(t *testing.T, ib *IDEMPOTENCY_BUCKET, result model.BookingOrder) {
				// Should return original order, not the new one
				if result.UserID != "user-original" {
					t.Errorf("Expected UserID 'user-original', got '%s'", result.UserID)
				}
				if result.Tier != model.TierFrontRow {
					t.Errorf("Expected Tier FRONT_ROW, got %s", result.Tier)
				}
				if result.SeatNo != 31 {
					t.Errorf("Expected SeatNo 31, got %d", result.SeatNo)
				}
			},
		},
		{
			name: "update status for existing key - non-pending status",
			bookingOrder: model.BookingOrder{
				UserID:         "user-3",
				Tier:           model.TierVIP,
				SeatNo:         3,
				Status:         model.BookingStatusConfirmed,
				IdempotencyKey: "key-update",
				PaymentStatus:  model.PaymentStatusConfirmed,
			},
			setupFunc: func(ib *IDEMPOTENCY_BUCKET) {
				ib.IDEMPOTENCY_STORE["key-update"] = model.BookingOrder{
					UserID:         "user-3",
					Tier:           model.TierVIP,
					SeatNo:         3,
					Status:         model.BookingStatusPending,
					IdempotencyKey: "key-update",
					PaymentStatus:  model.PaymentStatusPending,
				}
			},
			validateFunc: func(t *testing.T, ib *IDEMPOTENCY_BUCKET, result model.BookingOrder) {
				// Should update status in store
				stored := ib.IDEMPOTENCY_STORE["key-update"]
				if stored.Status != model.BookingStatusConfirmed {
					t.Errorf("Expected stored status CONFIRMED, got %s", stored.Status)
				}
				if result.Status != model.BookingStatusConfirmed {
					t.Errorf("Expected returned status CONFIRMED, got %s", result.Status)
				}
			},
		},
		{
			name: "update status - pending status does not update",
			bookingOrder: model.BookingOrder{
				UserID:         "user-4",
				Tier:           model.TierVIP,
				SeatNo:         4,
				Status:         model.BookingStatusPending,
				IdempotencyKey: "key-pending",
				PaymentStatus:  model.PaymentStatusPending,
			},
			setupFunc: func(ib *IDEMPOTENCY_BUCKET) {
				ib.IDEMPOTENCY_STORE["key-pending"] = model.BookingOrder{
					UserID:         "user-4",
					Tier:           model.TierVIP,
					SeatNo:         4,
					Status:         model.BookingStatusPending,
					IdempotencyKey: "key-pending",
					PaymentStatus:  model.PaymentStatusPending,
				}
			},
			validateFunc: func(t *testing.T, ib *IDEMPOTENCY_BUCKET, result model.BookingOrder) {
				// Should return existing order without updating
				stored := ib.IDEMPOTENCY_STORE["key-pending"]
				if stored.Status != model.BookingStatusPending {
					t.Errorf("Expected stored status to remain PENDING, got %s", stored.Status)
				}
			},
		},
		{
			name: "multiple idempotency keys",
			bookingOrder: model.BookingOrder{
				UserID:         "user-5",
				Tier:           model.TierGA,
				SeatNo:         61,
				Status:         model.BookingStatusPending,
				IdempotencyKey: "key-multi-1",
				PaymentStatus:  model.PaymentStatusPending,
			},
			setupFunc: func(ib *IDEMPOTENCY_BUCKET) {
				// Store multiple keys
				ib.IDEMPOTENCY_STORE["key-multi-1"] = model.BookingOrder{
					UserID:         "user-5",
					Tier:           model.TierGA,
					SeatNo:         61,
					IdempotencyKey: "key-multi-1",
				}
				ib.IDEMPOTENCY_STORE["key-multi-2"] = model.BookingOrder{
					UserID:         "user-6",
					Tier:           model.TierVIP,
					SeatNo:         10,
					IdempotencyKey: "key-multi-2",
				}
			},
			validateFunc: func(t *testing.T, ib *IDEMPOTENCY_BUCKET, result model.BookingOrder) {
				// Verify both keys still exist
				if _, exists := ib.IDEMPOTENCY_STORE["key-multi-1"]; !exists {
					t.Error("Expected key-multi-1 to exist")
				}
				if _, exists := ib.IDEMPOTENCY_STORE["key-multi-2"]; !exists {
					t.Error("Expected key-multi-2 to exist")
				}
				if len(ib.IDEMPOTENCY_STORE) != 2 {
					t.Errorf("Expected 2 stored keys, got %d", len(ib.IDEMPOTENCY_STORE))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ib := NewIdempotencyBucket().(*IDEMPOTENCY_BUCKET)

			if tt.setupFunc != nil {
				tt.setupFunc(ib)
			}

			result := ib.HandleIdempotency(tt.bookingOrder)

			// Verify returned order matches expected
			if result.IdempotencyKey != tt.expectedReturn.IdempotencyKey {
				t.Errorf("Expected IdempotencyKey %s, got %s", tt.expectedReturn.IdempotencyKey, result.IdempotencyKey)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, ib, result)
			}
		})
	}
}

func TestHandleIdempotency_Concurrency(t *testing.T) {
	ib := NewIdempotencyBucket().(*IDEMPOTENCY_BUCKET)
	idempotencyKey := "key-concurrent"
	numGoroutines := 20

	var wg sync.WaitGroup
	results := make([]model.BookingOrder, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			result := ib.HandleIdempotency(model.BookingOrder{
				UserID:         "user-concurrent",
				Tier:           model.TierVIP,
				SeatNo:         uint32(idx + 1),
				Status:         model.BookingStatusPending,
				IdempotencyKey: idempotencyKey,
				PaymentStatus:  model.PaymentStatusPending,
			})
			results[idx] = result
		}(i)
	}
	wg.Wait()

	// All results should have the same idempotency key
	for i, result := range results {
		if result.IdempotencyKey != idempotencyKey {
			t.Errorf("Result %d: Expected IdempotencyKey %s, got %s", i, idempotencyKey, result.IdempotencyKey)
		}
	}

	// Only one order should be stored (first one wins)
	if len(ib.IDEMPOTENCY_STORE) != 1 {
		t.Errorf("Expected 1 stored order, got %d", len(ib.IDEMPOTENCY_STORE))
	}

	// Verify stored order
	stored, exists := ib.IDEMPOTENCY_STORE[idempotencyKey]
	if !exists {
		t.Fatal("Expected idempotency key to be stored")
	}
	if stored.IdempotencyKey != idempotencyKey {
		t.Errorf("Expected stored IdempotencyKey %s, got %s", idempotencyKey, stored.IdempotencyKey)
	}
}

func TestHandleIdempotency_StatusUpdate(t *testing.T) {
	ib := NewIdempotencyBucket().(*IDEMPOTENCY_BUCKET)
	idempotencyKey := "key-status-update"

	// First call - pending status
	order1 := ib.HandleIdempotency(model.BookingOrder{
		UserID:         "user-status",
		Tier:           model.TierVIP,
		SeatNo:         20,
		Status:         model.BookingStatusPending,
		IdempotencyKey: idempotencyKey,
		PaymentStatus:  model.PaymentStatusPending,
	})

	if order1.Status != model.BookingStatusPending {
		t.Errorf("Expected status PENDING, got %s", order1.Status)
	}

	// Second call - confirmed status (should update)
	order2 := ib.HandleIdempotency(model.BookingOrder{
		UserID:         "user-status",
		Tier:           model.TierVIP,
		SeatNo:         20,
		Status:         model.BookingStatusConfirmed,
		IdempotencyKey: idempotencyKey,
		PaymentStatus:  model.PaymentStatusConfirmed,
	})

	if order2.Status != model.BookingStatusConfirmed {
		t.Errorf("Expected status CONFIRMED, got %s", order2.Status)
	}

	// Verify stored order is updated
	stored := ib.IDEMPOTENCY_STORE[idempotencyKey]
	if stored.Status != model.BookingStatusConfirmed {
		t.Errorf("Expected stored status CONFIRMED, got %s", stored.Status)
	}

	// Third call - pending status again (should not update)
	order3 := ib.HandleIdempotency(model.BookingOrder{
		UserID:         "user-status",
		Tier:           model.TierVIP,
		SeatNo:         20,
		Status:         model.BookingStatusPending,
		IdempotencyKey: idempotencyKey,
		PaymentStatus:  model.PaymentStatusPending,
	})

	// Should return confirmed status (not update back to pending)
	if order3.Status != model.BookingStatusConfirmed {
		t.Errorf("Expected status to remain CONFIRMED, got %s", order3.Status)
	}
}
