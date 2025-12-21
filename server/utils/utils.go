package utils

import (
	"encoding/json"
	"net/http"

	"github.com/ignius299792458/techkraft-ch-svr/model"
)

func ValidateBookingRequest(req *model.BookingOrder) error {
	if req.UserID == "" {
		return NewValidationError("user_id is required")
	}
	if req.IdempotencyKey == "" {
		return NewValidationError("idempotency_key is required")
	}
	if req.SeatNo == 0 {
		return NewValidationError("seat_no must be greater than 0")
	}
	if req.Quantity == 0 {
		return NewValidationError("quantity must be greater than 0")
	}
	if !req.Tier.IsValid() {
		return NewValidationError("invalid tier")
	}
	return nil
}

func CalculateAmount(tier model.Tier, quantity uint16) uint64 {
	var pricePerSeat uint16
	switch tier {
	case model.TierVIP:
		pricePerSeat = 50000 // $500 in cents
	case model.TierFrontRow:
		pricePerSeat = 20000 // $200 in cents
	case model.TierGA:
		pricePerSeat = 10000 // $100 in cents
	default:
		pricePerSeat = 10000
	}
	return uint64(pricePerSeat * quantity)
}

func RespondSuccess(w http.ResponseWriter, message string, booking *model.Booking) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.BookingResponse{
		Success: true,
		Message: message,
		Booking: booking,
	})
}

func RespondError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(model.BookingResponse{
		Success: false,
		Message: message,
	})
}
