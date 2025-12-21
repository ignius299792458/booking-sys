package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/ignius299792458/techkraft-ch-svr/model"
	"github.com/ignius299792458/techkraft-ch-svr/store"
	"github.com/ignius299792458/techkraft-ch-svr/utils"
)

var (
	bookingStore     store.BookingStore
	idempotencyStore store.Idempotency
)

func init() {
	bookingStore = store.NewBookingStoreBucket()
	idempotencyStore = store.NewIdempotencyBucket()
}

func HandleBooking(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Set headers early
	w.Header().Set("Content-Type", "application/json")

	// Parse request
	var req model.BookingOrder
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := utils.ValidateBookingRequest(&req); err != nil {
		utils.RespondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Calculate amount based on tier
	totalAmt := utils.CalculateAmount(req.Tier, req.Quantity)

	// Create booking order
	bookingOrder := model.BookingOrder{
		UserID:           req.UserID,
		Tier:             req.Tier,
		Quantity:         req.Quantity,
		Status:           model.BookingStatusPending,
		IdempotencyKey:   req.IdempotencyKey,
		Country:          req.Country,
		ZipCode:          req.ZipCode,
		Currency:         req.Currency,
		SeatNo:           req.SeatNo,
		TotalAmtInUSCent: totalAmt,
		PaymentID:        req.PaymentID,
	}

	// Handle idempotency - check if this request was already processed
	idempotentOrder := idempotencyStore.HandleIdempotency(bookingOrder)

	// Register the booking
	newBooking, err := bookingStore.RegisterBooking(idempotentOrder)
	if err != nil {
		utils.RespondError(w, err.Error(), http.StatusConflict)
		return
	}

	// Update booking order with the created booking ID
	bookingOrder.Status = newBooking.Status

	// Update idempotency store with complete booking info
	idempotencyStore.HandleIdempotency(bookingOrder)

	duration := time.Since(start).Milliseconds()
	slog.Info("Booking created",
		"booking_id", newBooking.ID,
		"user_id", newBooking.UserID,
		"seat", newBooking.SeatNo,
		"tier", newBooking.Tier,
		"duration_ms", duration)

	utils.RespondSuccess(w, "new booking successful", &newBooking)
}
