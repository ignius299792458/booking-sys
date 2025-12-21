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
	totalAmt := utils.CalculateAmount(req.Tier)

	// Create booking order
	bookingOrder := model.BookingOrder{
		UserID:           req.UserID,
		Tier:             req.Tier,
		Status:           model.BookingStatusPending,
		IdempotencyKey:   req.IdempotencyKey,
		Country:          req.Country,
		ZipCode:          req.ZipCode,
		Currency:         req.Currency,
		SeatNo:           req.SeatNo,
		TotalAmtInUSCent: totalAmt,
		PaymentID:        req.PaymentID,
		PaymentStatus:    req.PaymentStatus,
	}

	// Handle idempotency - check if this request was already processed
	idempotentOrder := idempotencyStore.HandleIdempotency(bookingOrder)

	if idempotentOrder.Status == model.BookingStatusConfirmed {
		// Booking already confirmed
		oldConfirmedBooking, err := bookingStore.GetBooking(idempotentOrder.SeatNo)
		if err != nil {
			utils.RespondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		utils.RespondSuccess(w, "booking already confirmed", &oldConfirmedBooking)
		return
	}

	// Register the booking
	newBooking, err := bookingStore.RegisterBooking(idempotentOrder)
	if err != nil {
		utils.RespondError(w, err.Error(), http.StatusConflict)
		return
	}

	// Update booking order with the created booking ID
	bookingOrder.Status = newBooking.Status

	// Update idempotency store with complete booking info
	if newBooking.Status != model.BookingStatusPending {
		idempotencyStore.HandleIdempotency(bookingOrder)
	}

	duration := time.Since(start).Milliseconds()
	slog.Info("Booking processed duration : ", "duration_ms", duration, "sec: ", duration/1000)

	slog.Info("Booking created",
		"booking_id", newBooking.ID,
		"user_id", newBooking.UserID,
		"seat", newBooking.SeatNo,
		"tier", newBooking.Tier)

	utils.RespondSuccess(w, "new booking successful", &newBooking)
}

func HandleAvailability(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reservedSeats := bookingStore.GetReservedSeats()

	// Calculate available counts based on seat ranges:
	// VIP: seats 1-30 (30 seats total)
	// FRONT_ROW: seats 31-60 (30 seats total)
	// GA: seats 61-100 (40 seats total)
	const (
		VIPTotal      uint32 = 30 // seats 1-30
		FrontRowTotal uint32 = 30 // seats 31-60
		GATotal       uint32 = 40 // seats 61-100
	)

	vipReserved := uint32(len(reservedSeats["VIP"]))
	frontRowReserved := uint32(len(reservedSeats["FRONT_ROW"]))
	gaReserved := uint32(len(reservedSeats["GA"]))

	// Helper function to calculate available seats
	calculateAvailableSeats := func(minSeat, maxSeat uint32, reserved []uint32) []uint32 {
		reservedSet := make(map[uint32]bool)
		for _, seat := range reserved {
			reservedSet[seat] = true
		}
		var available []uint32
		for i := minSeat; i <= maxSeat; i++ {
			if !reservedSet[i] {
				available = append(available, i)
			}
		}
		return available
	}

	tiers := []model.TierInfo{
		{
			Tier:          model.TierVIP,
			Price:         uint64(model.PriceVIPCents),
			TotalSeats:    VIPTotal,
			ReservedCount: vipReserved, // number of seats reserved for this tier
			AvailableList: calculateAvailableSeats(1, 30, reservedSeats["VIP"]),
		},
		{
			Tier:          model.TierFrontRow,
			Price:         uint64(model.PriceFrontRowCents),
			TotalSeats:    FrontRowTotal,
			ReservedCount: frontRowReserved, // number of seats reserved for this tier
			AvailableList: calculateAvailableSeats(31, 60, reservedSeats["FRONT_ROW"]),
		},
		{
			Tier:          model.TierGA,
			Price:         uint64(model.PriceGACents),
			TotalSeats:    GATotal,
			ReservedCount: gaReserved, // number of seats reserved for this tier
			AvailableList: calculateAvailableSeats(61, 100, reservedSeats["GA"]),
		},
	}

	response := model.AvailabilityResponse{
		Success: true,
		Message: "reserved seats retrieved successfully",
		Tiers:   tiers,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
