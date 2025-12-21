package store

import (
	"sync"
	"testing"

	"github.com/ignius299792458/techkraft-ch-svr/model"
)

func TestRegisterBooking(t *testing.T) {
	tests := []struct {
		name          string
		bookingOrder  model.BookingOrder
		expectedError string
		setupFunc     func(*BOOKING_STORE_BUCKET)
	}{
		{
			name: "successful VIP booking",
			bookingOrder: model.BookingOrder{
				UserID:           "user-1",
				Tier:             model.TierVIP,
				SeatNo:           1,
				Status:           model.BookingStatusPending,
				IdempotencyKey:   "key-1",
				Country:          "USA",
				ZipCode:          "10001",
				Currency:         "USD",
				TotalAmtInUSCent: 10000,
				PaymentID:        "pay-1",
				PaymentStatus:    model.PaymentStatusConfirmed,
			},
			expectedError: "",
		},
		{
			name: "successful FRONT_ROW booking",
			bookingOrder: model.BookingOrder{
				UserID:           "user-2",
				Tier:             model.TierFrontRow,
				SeatNo:           31,
				Status:           model.BookingStatusPending,
				IdempotencyKey:   "key-2",
				Country:          "UK",
				ZipCode:          "SW1A",
				Currency:         "GBP",
				TotalAmtInUSCent: 5000,
				PaymentID:        "pay-2",
				PaymentStatus:    model.PaymentStatusConfirmed,
			},
			expectedError: "",
		},
		{
			name: "successful GA booking",
			bookingOrder: model.BookingOrder{
				UserID:           "user-3",
				Tier:             model.TierGA,
				SeatNo:           61,
				Status:           model.BookingStatusPending,
				IdempotencyKey:   "key-3",
				Country:          "Canada",
				ZipCode:          "M5H",
				Currency:         "CAD",
				TotalAmtInUSCent: 1000,
				PaymentID:        "pay-3",
				PaymentStatus:    model.PaymentStatusConfirmed,
			},
			expectedError: "",
		},
		{
			name: "invalid seat number - zero",
			bookingOrder: model.BookingOrder{
				UserID:         "user-4",
				Tier:           model.TierVIP,
				SeatNo:         0,
				IdempotencyKey: "key-4",
				PaymentStatus:  model.PaymentStatusPending,
			},
			expectedError: "invalid seat number",
		},
		{
			name: "invalid seat number - exceeds total",
			bookingOrder: model.BookingOrder{
				UserID:         "user-5",
				Tier:           model.TierGA,
				SeatNo:         101, // exceeds TOTAL_SEAT = 100
				IdempotencyKey: "key-5",
				PaymentStatus:  model.PaymentStatusPending,
			},
			expectedError: "invalid seat number",
		},
		{
			name: "double booking - seat already booked",
			bookingOrder: model.BookingOrder{
				UserID:         "user-6",
				Tier:           model.TierVIP,
				SeatNo:         5,
				IdempotencyKey: "key-6",
				PaymentStatus:  model.PaymentStatusPending,
			},
			expectedError: "seat already booked",
			setupFunc: func(b *BOOKING_STORE_BUCKET) {
				// Pre-book seat 5
				b.RegisterBooking(model.BookingOrder{
					UserID:         "user-existing",
					Tier:           model.TierVIP,
					SeatNo:         5,
					IdempotencyKey: "key-existing",
					PaymentStatus:  model.PaymentStatusConfirmed,
					PaymentID:      "pay-existing",
				})
			},
		},
		{
			name: "booking with confirmed payment status",
			bookingOrder: model.BookingOrder{
				UserID:         "user-7",
				Tier:           model.TierVIP,
				SeatNo:         6,
				Status:         model.BookingStatusPending,
				IdempotencyKey: "key-7",
				PaymentID:      "pay-7",
				PaymentStatus:  model.PaymentStatusConfirmed,
			},
			expectedError: "",
		},
		{
			name: "booking with failed payment status",
			bookingOrder: model.BookingOrder{
				UserID:         "user-8",
				Tier:           model.TierVIP,
				SeatNo:         7,
				Status:         model.BookingStatusPending,
				IdempotencyKey: "key-8",
				PaymentID:      "pay-8",
				PaymentStatus:  model.PaymentStatusFailed,
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBookingStoreBucket().(*BOOKING_STORE_BUCKET)

			if tt.setupFunc != nil {
				tt.setupFunc(b)
			}

			booking, err := b.RegisterBooking(tt.bookingOrder)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error '%s', got nil", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got '%s'", err.Error())
				}
				if booking.ID.String() == "" {
					t.Error("Expected booking ID to be set")
				}
				if booking.UserID != tt.bookingOrder.UserID {
					t.Errorf("Expected UserID %s, got %s", tt.bookingOrder.UserID, booking.UserID)
				}
				if booking.Tier != tt.bookingOrder.Tier {
					t.Errorf("Expected Tier %s, got %s", tt.bookingOrder.Tier, booking.Tier)
				}
				if booking.SeatNo != tt.bookingOrder.SeatNo {
					t.Errorf("Expected SeatNo %d, got %d", tt.bookingOrder.SeatNo, booking.SeatNo)
				}

				// Verify status based on payment status
				if tt.bookingOrder.PaymentStatus == model.PaymentStatusConfirmed && len(tt.bookingOrder.PaymentID) > 0 {
					if booking.Status != model.BookingStatusConfirmed {
						t.Errorf("Expected status CONFIRMED, got %s", booking.Status)
					}
				} else if tt.bookingOrder.PaymentStatus == model.PaymentStatusFailed || tt.bookingOrder.PaymentStatus == model.PaymentStatusCanceled {
					if booking.Status != model.BookingStatusCanceled {
						t.Errorf("Expected status CANCELED, got %s", booking.Status)
					}
				}
			}
		})
	}
}

func TestRegisterBooking_Concurrency(t *testing.T) {
	b := NewBookingStoreBucket().(*BOOKING_STORE_BUCKET)
	seatNo := uint32(10)
	numGoroutines := 10

	var wg sync.WaitGroup
	errors := make([]error, numGoroutines)
	bookings := make([]model.Booking, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			booking, err := b.RegisterBooking(model.BookingOrder{
				UserID:         "user-concurrent",
				Tier:           model.TierVIP,
				SeatNo:         seatNo,
				IdempotencyKey: "key-concurrent",
				PaymentStatus:  model.PaymentStatusConfirmed,
				PaymentID:      "pay-concurrent",
			})
			errors[idx] = err
			bookings[idx] = booking
		}(i)
	}
	wg.Wait()

	// Only one booking should succeed
	successCount := 0
	for _, err := range errors {
		if err == nil {
			successCount++
		}
	}

	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful booking, got %d", successCount)
	}

	// Verify only one booking exists in store
	reservedSeats := b.GetReservedSeats()
	vipReserved := reservedSeats["VIP"]
	found := false
	for _, seat := range vipReserved {
		if seat == seatNo {
			if found {
				t.Error("Seat should only be booked once")
			}
			found = true
		}
	}
	if !found {
		t.Error("Seat should be booked")
	}
}

func TestGetBooking(t *testing.T) {
	tests := []struct {
		name          string
		seatNo        uint32
		setupFunc     func(*BOOKING_STORE_BUCKET)
		expectedError string
		validateFunc  func(*testing.T, model.Booking)
	}{
		{
			name:   "successful retrieval",
			seatNo: 15,
			setupFunc: func(b *BOOKING_STORE_BUCKET) {
				b.RegisterBooking(model.BookingOrder{
					UserID:         "user-get",
					Tier:           model.TierVIP,
					SeatNo:         15,
					IdempotencyKey: "key-get",
					PaymentStatus:  model.PaymentStatusConfirmed,
					PaymentID:      "pay-get",
				})
			},
			expectedError: "",
			validateFunc: func(t *testing.T, booking model.Booking) {
				if booking.UserID != "user-get" {
					t.Errorf("Expected UserID 'user-get', got '%s'", booking.UserID)
				}
				if booking.SeatNo != 15 {
					t.Errorf("Expected SeatNo 15, got %d", booking.SeatNo)
				}
				if booking.Tier != model.TierVIP {
					t.Errorf("Expected Tier VIP, got %s", booking.Tier)
				}
			},
		},
		{
			name:          "booking not found",
			seatNo:        99,
			expectedError: "booking not found",
		},
		{
			name:   "retrieve different tier booking",
			seatNo: 50,
			setupFunc: func(b *BOOKING_STORE_BUCKET) {
				b.RegisterBooking(model.BookingOrder{
					UserID:         "user-front",
					Tier:           model.TierFrontRow,
					SeatNo:         50,
					IdempotencyKey: "key-front",
					PaymentStatus:  model.PaymentStatusConfirmed,
					PaymentID:      "pay-front",
				})
			},
			expectedError: "",
			validateFunc: func(t *testing.T, booking model.Booking) {
				if booking.Tier != model.TierFrontRow {
					t.Errorf("Expected Tier FRONT_ROW, got %s", booking.Tier)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket := NewBookingStoreBucket().(*BOOKING_STORE_BUCKET)

			if tt.setupFunc != nil {
				tt.setupFunc(bucket)
			}

			booking, err := bucket.GetBooking(tt.seatNo)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error '%s', got nil", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got '%s'", err.Error())
				}
				if tt.validateFunc != nil {
					tt.validateFunc(t, booking)
				}
			}
		})
	}
}

func TestGetReservedSeats(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(*BOOKING_STORE_BUCKET)
		validateFunc func(*testing.T, map[string][]uint32)
	}{
		{
			name: "empty reserved seats",
			validateFunc: func(t *testing.T, reserved map[string][]uint32) {
				if len(reserved["VIP"]) != 0 {
					t.Errorf("Expected 0 VIP reserved seats, got %d", len(reserved["VIP"]))
				}
				if len(reserved["FRONT_ROW"]) != 0 {
					t.Errorf("Expected 0 FRONT_ROW reserved seats, got %d", len(reserved["FRONT_ROW"]))
				}
				if len(reserved["GA"]) != 0 {
					t.Errorf("Expected 0 GA reserved seats, got %d", len(reserved["GA"]))
				}
			},
		},
		{
			name: "reserved seats by tier",
			setupFunc: func(b *BOOKING_STORE_BUCKET) {
				// Book VIP seats
				b.RegisterBooking(model.BookingOrder{
					UserID:         "user-vip-1",
					Tier:           model.TierVIP,
					SeatNo:         1,
					IdempotencyKey: "key-vip-1",
					PaymentStatus:  model.PaymentStatusConfirmed,
					PaymentID:      "pay-vip-1",
				})
				b.RegisterBooking(model.BookingOrder{
					UserID:         "user-vip-2",
					Tier:           model.TierVIP,
					SeatNo:         2,
					IdempotencyKey: "key-vip-2",
					PaymentStatus:  model.PaymentStatusConfirmed,
					PaymentID:      "pay-vip-2",
				})
				// Book FRONT_ROW seats
				b.RegisterBooking(model.BookingOrder{
					UserID:         "user-front-1",
					Tier:           model.TierFrontRow,
					SeatNo:         31,
					IdempotencyKey: "key-front-1",
					PaymentStatus:  model.PaymentStatusConfirmed,
					PaymentID:      "pay-front-1",
				})
				b.RegisterBooking(model.BookingOrder{
					UserID:         "user-front-2",
					Tier:           model.TierFrontRow,
					SeatNo:         32,
					IdempotencyKey: "key-front-2",
					PaymentStatus:  model.PaymentStatusConfirmed,
					PaymentID:      "pay-front-2",
				})
				b.RegisterBooking(model.BookingOrder{
					UserID:         "user-front-3",
					Tier:           model.TierFrontRow,
					SeatNo:         33,
					IdempotencyKey: "key-front-3",
					PaymentStatus:  model.PaymentStatusConfirmed,
					PaymentID:      "pay-front-3",
				})
				// Book GA seats
				b.RegisterBooking(model.BookingOrder{
					UserID:         "user-ga-1",
					Tier:           model.TierGA,
					SeatNo:         61,
					IdempotencyKey: "key-ga-1",
					PaymentStatus:  model.PaymentStatusConfirmed,
					PaymentID:      "pay-ga-1",
				})
			},
			validateFunc: func(t *testing.T, reserved map[string][]uint32) {
				// Check VIP
				if len(reserved["VIP"]) != 2 {
					t.Errorf("Expected 2 VIP reserved seats, got %d", len(reserved["VIP"]))
				}
				vipSeats := reserved["VIP"]
				if !contains(vipSeats, 1) || !contains(vipSeats, 2) {
					t.Error("VIP reserved seats should include 1 and 2")
				}

				// Check FRONT_ROW
				if len(reserved["FRONT_ROW"]) != 3 {
					t.Errorf("Expected 3 FRONT_ROW reserved seats, got %d", len(reserved["FRONT_ROW"]))
				}
				frontRowSeats := reserved["FRONT_ROW"]
				if !contains(frontRowSeats, 31) || !contains(frontRowSeats, 32) || !contains(frontRowSeats, 33) {
					t.Error("FRONT_ROW reserved seats should include 31, 32, and 33")
				}

				// Check GA
				if len(reserved["GA"]) != 1 {
					t.Errorf("Expected 1 GA reserved seat, got %d", len(reserved["GA"]))
				}
				gaSeats := reserved["GA"]
				if !contains(gaSeats, 61) {
					t.Error("GA reserved seats should include 61")
				}
			},
		},
		{
			name: "reserved seats across all tiers",
			setupFunc: func(b *BOOKING_STORE_BUCKET) {
				// Book seats from each tier
				for i := uint32(1); i <= 5; i++ {
					b.RegisterBooking(model.BookingOrder{
						UserID:         "user-vip",
						Tier:           model.TierVIP,
						SeatNo:         i,
						IdempotencyKey: "key-vip",
						PaymentStatus:  model.PaymentStatusConfirmed,
						PaymentID:      "pay-vip",
					})
				}
				for i := uint32(31); i <= 35; i++ {
					b.RegisterBooking(model.BookingOrder{
						UserID:         "user-front",
						Tier:           model.TierFrontRow,
						SeatNo:         i,
						IdempotencyKey: "key-front",
						PaymentStatus:  model.PaymentStatusConfirmed,
						PaymentID:      "pay-front",
					})
				}
				for i := uint32(61); i <= 65; i++ {
					b.RegisterBooking(model.BookingOrder{
						UserID:         "user-ga",
						Tier:           model.TierGA,
						SeatNo:         i,
						IdempotencyKey: "key-ga",
						PaymentStatus:  model.PaymentStatusConfirmed,
						PaymentID:      "pay-ga",
					})
				}
			},
			validateFunc: func(t *testing.T, reserved map[string][]uint32) {
				if len(reserved["VIP"]) != 5 {
					t.Errorf("Expected 5 VIP reserved seats, got %d", len(reserved["VIP"]))
				}
				if len(reserved["FRONT_ROW"]) != 5 {
					t.Errorf("Expected 5 FRONT_ROW reserved seats, got %d", len(reserved["FRONT_ROW"]))
				}
				if len(reserved["GA"]) != 5 {
					t.Errorf("Expected 5 GA reserved seats, got %d", len(reserved["GA"]))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBookingStoreBucket().(*BOOKING_STORE_BUCKET)

			if tt.setupFunc != nil {
				tt.setupFunc(b)
			}

			reserved := b.GetReservedSeats()

			// Verify structure
			if reserved == nil {
				t.Fatal("Expected reserved seats map, got nil")
			}
			if reserved["VIP"] == nil {
				t.Error("Expected VIP key in reserved seats map")
			}
			if reserved["FRONT_ROW"] == nil {
				t.Error("Expected FRONT_ROW key in reserved seats map")
			}
			if reserved["GA"] == nil {
				t.Error("Expected GA key in reserved seats map")
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, reserved)
			}
		})
	}
}

// Helper function to check if slice contains value
func contains(slice []uint32, val uint32) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
