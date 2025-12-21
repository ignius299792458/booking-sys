package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ignius299792458/techkraft-ch-svr/model"
	"github.com/ignius299792458/techkraft-ch-svr/store"
)

// setupTestHandlers creates fresh instances for testing
func setupTestHandlers() {
	bookingStore = store.NewBookingStoreBucket()
	idempotencyStore = store.NewIdempotencyBucket()
}

func TestHandleBooking(t *testing.T) {
	setupTestHandlers()

	tests := []struct {
		name           string
		requestBody    model.BookingOrder
		expectedStatus int
		expectedError  string
		setupFunc      func() // setup function to prepare test state
	}{
		{
			name: "successful VIP booking",
			requestBody: model.BookingOrder{
				UserID:         "user-123",
				Tier:           model.TierVIP,
				SeatNo:         1,
				Country:        "USA",
				ZipCode:        "10001",
				Currency:       "USD",
				IdempotencyKey: "key-1",
				PaymentID:      "pay-1",
				PaymentStatus:  model.PaymentStatusConfirmed,
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name: "successful FRONT_ROW booking",
			requestBody: model.BookingOrder{
				UserID:         "user-456",
				Tier:           model.TierFrontRow,
				SeatNo:         31,
				Country:        "UK",
				ZipCode:        "SW1A",
				Currency:       "GBP",
				IdempotencyKey: "key-2",
				PaymentID:      "pay-2",
				PaymentStatus:  model.PaymentStatusConfirmed,
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name: "successful GA booking",
			requestBody: model.BookingOrder{
				UserID:         "user-789",
				Tier:           model.TierGA,
				SeatNo:         61,
				Country:        "Canada",
				ZipCode:        "M5H",
				Currency:       "CAD",
				IdempotencyKey: "key-3",
				PaymentID:      "pay-3",
				PaymentStatus:  model.PaymentStatusConfirmed,
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name: "missing user_id",
			requestBody: model.BookingOrder{
				UserID:         "",
				Tier:           model.TierVIP,
				SeatNo:         1,
				IdempotencyKey: "key-4",
				PaymentStatus:  model.PaymentStatusPending,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "user_id is required",
		},
		{
			name: "missing idempotency_key",
			requestBody: model.BookingOrder{
				UserID:        "user-123",
				Tier:          model.TierVIP,
				SeatNo:        1,
				PaymentStatus: model.PaymentStatusPending,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "idempotency_key is required",
		},
		{
			name: "invalid tier",
			requestBody: model.BookingOrder{
				UserID:         "user-123",
				Tier:           "INVALID_TIER",
				SeatNo:         1,
				IdempotencyKey: "key-5",
				PaymentStatus:  model.PaymentStatusPending,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid tier",
		},
		{
			name: "invalid seat number (zero)",
			requestBody: model.BookingOrder{
				UserID:         "user-123",
				Tier:           model.TierVIP,
				SeatNo:         0,
				IdempotencyKey: "key-6",
				PaymentStatus:  model.PaymentStatusPending,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "seat_no must be greater than 0",
		},
		{
			name: "double booking same seat",
			requestBody: model.BookingOrder{
				UserID:         "user-123",
				Tier:           model.TierVIP,
				SeatNo:         2,
				Country:        "USA",
				ZipCode:        "10001",
				Currency:       "USD",
				IdempotencyKey: "key-7",
				PaymentID:      "pay-7",
				PaymentStatus:  model.PaymentStatusConfirmed,
			},
			setupFunc: func() {
				// Pre-book seat 2
				bookingStore.RegisterBooking(model.BookingOrder{
					UserID:         "user-existing",
					Tier:           model.TierVIP,
					SeatNo:         2,
					IdempotencyKey: "key-existing",
					PaymentID:      "pay-existing",
					PaymentStatus:  model.PaymentStatusConfirmed,
				})
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "seat already booked",
		},
		{
			name: "idempotent request - returns existing booking",
			requestBody: model.BookingOrder{
				UserID:         "user-123",
				Tier:           model.TierVIP,
				SeatNo:         3,
				Country:        "USA",
				ZipCode:        "10001",
				Currency:       "USD",
				IdempotencyKey: "key-idempotent",
				PaymentID:      "pay-idempotent",
				PaymentStatus:  model.PaymentStatusConfirmed,
			},
			expectedStatus: http.StatusOK,
			setupFunc: func() {
				// First booking with this idempotency key
				bookingStore.RegisterBooking(model.BookingOrder{
					UserID:         "user-123",
					Tier:           model.TierVIP,
					SeatNo:         3,
					IdempotencyKey: "key-idempotent",
					PaymentID:      "pay-idempotent",
					PaymentStatus:  model.PaymentStatusConfirmed,
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestHandlers()
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/booking/ticket", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			HandleBooking(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response model.BookingResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if tt.expectedError != "" {
				if response.Message != tt.expectedError {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedError, response.Message)
				}
				if response.Success {
					t.Error("Expected success to be false for error case")
				}
			} else {
				if !response.Success {
					t.Errorf("Expected success to be true, got false. Message: %s", response.Message)
				}
				if response.Booking == nil {
					t.Error("Expected booking in response")
				} else {
					// Verify booking details
					if response.Booking.UserID != tt.requestBody.UserID {
						t.Errorf("Expected UserID %s, got %s", tt.requestBody.UserID, response.Booking.UserID)
					}
					if response.Booking.Tier != tt.requestBody.Tier {
						t.Errorf("Expected Tier %s, got %s", tt.requestBody.Tier, response.Booking.Tier)
					}
					if response.Booking.SeatNo != tt.requestBody.SeatNo {
						t.Errorf("Expected SeatNo %d, got %d", tt.requestBody.SeatNo, response.Booking.SeatNo)
					}
				}
			}
		})
	}
}

func TestHandleAvailability(t *testing.T) {
	setupTestHandlers()

	tests := []struct {
		name          string
		setupFunc     func() // setup function to create bookings
		expectedTiers int
		validateFunc  func(*testing.T, model.AvailabilityResponse)
	}{
		{
			name:          "empty availability - all seats available",
			expectedTiers: 3,
			validateFunc: func(t *testing.T, resp model.AvailabilityResponse) {
				if !resp.Success {
					t.Error("Expected success to be true")
				}
				if len(resp.Tiers) != 3 {
					t.Errorf("Expected 3 tiers, got %d", len(resp.Tiers))
				}

				// Check VIP tier
				vipTier := resp.Tiers[0]
				if vipTier.Tier != model.TierVIP {
					t.Errorf("Expected first tier to be VIP, got %s", vipTier.Tier)
				}
				if vipTier.TotalSeats != 30 {
					t.Errorf("Expected VIP TotalSeats 30, got %d", vipTier.TotalSeats)
				}
				if vipTier.ReservedCount != 0 {
					t.Errorf("Expected VIP ReservedCount 0, got %d", vipTier.ReservedCount)
				}
				if len(vipTier.AvailableList) != 30 {
					t.Errorf("Expected 30 available VIP seats, got %d", len(vipTier.AvailableList))
				}

				// Check FRONT_ROW tier
				frontRowTier := resp.Tiers[1]
				if frontRowTier.Tier != model.TierFrontRow {
					t.Errorf("Expected second tier to be FRONT_ROW, got %s", frontRowTier.Tier)
				}
				if frontRowTier.TotalSeats != 30 {
					t.Errorf("Expected FRONT_ROW TotalSeats 30, got %d", frontRowTier.TotalSeats)
				}

				// Check GA tier
				gaTier := resp.Tiers[2]
				if gaTier.Tier != model.TierGA {
					t.Errorf("Expected third tier to be GA, got %s", gaTier.Tier)
				}
				if gaTier.TotalSeats != 40 {
					t.Errorf("Expected GA TotalSeats 40, got %d", gaTier.TotalSeats)
				}
			},
		},
		{
			name: "availability with reserved seats",
			setupFunc: func() {
				// Book some seats
				bookingStore.RegisterBooking(model.BookingOrder{
					UserID:         "user-1",
					Tier:           model.TierVIP,
					SeatNo:         1,
					IdempotencyKey: "key-1",
					PaymentID:      "pay-1",
					PaymentStatus:  model.PaymentStatusConfirmed,
				})
				bookingStore.RegisterBooking(model.BookingOrder{
					UserID:         "user-2",
					Tier:           model.TierVIP,
					SeatNo:         2,
					IdempotencyKey: "key-2",
					PaymentID:      "pay-2",
					PaymentStatus:  model.PaymentStatusConfirmed,
				})
				bookingStore.RegisterBooking(model.BookingOrder{
					UserID:         "user-3",
					Tier:           model.TierFrontRow,
					SeatNo:         31,
					IdempotencyKey: "key-3",
					PaymentID:      "pay-3",
					PaymentStatus:  model.PaymentStatusConfirmed,
				})
			},
			expectedTiers: 3,
			validateFunc: func(t *testing.T, resp model.AvailabilityResponse) {
				if !resp.Success {
					t.Error("Expected success to be true")
				}

				// Find VIP tier
				var vipTier *model.TierInfo
				for i := range resp.Tiers {
					if resp.Tiers[i].Tier == model.TierVIP {
						vipTier = &resp.Tiers[i]
						break
					}
				}
				if vipTier == nil {
					t.Fatal("VIP tier not found in response")
				}
				if vipTier.ReservedCount != 2 {
					t.Errorf("Expected VIP ReservedCount 2, got %d", vipTier.ReservedCount)
				}
				if len(vipTier.AvailableList) != 28 {
					t.Errorf("Expected 28 available VIP seats, got %d", len(vipTier.AvailableList))
				}
				// Verify seat 1 and 2 are not in available list
				for _, seat := range vipTier.AvailableList {
					if seat == 1 || seat == 2 {
						t.Errorf("Seat %d should not be in available list", seat)
					}
				}

				// Find FRONT_ROW tier
				var frontRowTier *model.TierInfo
				for i := range resp.Tiers {
					if resp.Tiers[i].Tier == model.TierFrontRow {
						frontRowTier = &resp.Tiers[i]
						break
					}
				}
				if frontRowTier == nil {
					t.Fatal("FRONT_ROW tier not found in response")
				}
				if frontRowTier.ReservedCount != 1 {
					t.Errorf("Expected FRONT_ROW ReservedCount 1, got %d", frontRowTier.ReservedCount)
				}
				if len(frontRowTier.AvailableList) != 29 {
					t.Errorf("Expected 29 available FRONT_ROW seats, got %d", len(frontRowTier.AvailableList))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestHandlers()
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			req := httptest.NewRequest(http.MethodGet, "/booking/availability", nil)
			w := httptest.NewRecorder()

			HandleAvailability(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			var response model.AvailabilityResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, response)
			}
		})
	}
}
