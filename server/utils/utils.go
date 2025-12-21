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
	if !req.Tier.IsValidTier() {
		return NewValidationError("invalid tier")
	}
	if !req.PaymentStatus.IsValidPaymentStatus() {
		return NewValidationError("invalid payment status")
	}
	return nil
}

func CalculateAmount(tier model.Tier) uint64 {
	switch tier {
	case model.TierVIP:
		return 100 * 100 // $100 in cents
	case model.TierFrontRow:
		return 50 * 100 // $50 in cents
	case model.TierGA:
		return 10 * 100 // $10 in cents
	default:
		return 10 * 100 // default to $10 in cents
	}
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
