package model

import (
	"time"

	"github.com/google/uuid"
)

// ---- Domain constants ----

type Tier string

const (
	TierVIP      Tier = "VIP"
	TierFrontRow Tier = "FRONT_ROW"
	TierGA       Tier = "GA"
)

func (t Tier) IsValid() bool {
	switch t {
	case TierVIP, TierFrontRow, TierGA:
		return true
	default:
		return false
	}
}

// Store money as integer cents to avoid float precision issues.
const (
	PriceVIPCents      int64 = 10000 // $100.00
	PriceFrontRowCents int64 = 5000  // $50.00
	PriceGACents       int64 = 1000  // $10.00
)

// ---- Booking ----

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "PENDING"   // created, awaiting payment simulation
	BookingStatusConfirmed BookingStatus = "CONFIRMED" // inventory decremented and payment success
	BookingStatusFailed    BookingStatus = "FAILED"    // payment failed (or sold out)
	BookingStatusCanceled  BookingStatus = "CANCELED"  // optional future state
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusConfirmed PaymentStatus = "CONFIRMED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusCanceled  PaymentStatus = "CANCELED"
)

type Booking struct {
	ID       uuid.UUID     `json:"id"`
	UserID   string        `json:"userId"` // mocked user id
	Tier     Tier          `json:"tier"`
	Quantity uint16        `json:"quantity"`
	Status   BookingStatus `json:"status"`

	// Idempotency: retries of the same "Book" click should reuse this key.
	IdempotencyKey string `json:"idempotencyKey,omitempty"`

	// seat
	SeatNo uint32 `json:"seatNo"`

	// country
	Country  string `json:"country"`
	ZipCode  string `json:"zipCode"`
	Currency string `json:"currency"`

	// Payment
	TotalAmtInUSCent uint64 `json:"totalAmtInUSCent"`
	PaymentID        string `json:"paymentID"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type BookingOrder struct {
	UserID   string        `json:"userId"` // mocked user id
	Tier     Tier          `json:"tier"`
	Quantity uint16        `json:"quantity"`
	Status   BookingStatus `json:"status"`

	IdempotencyKey string `json:"idempotencyKey,omitempty"`

	// country
	Country  string `json:"country"`
	ZipCode  string `json:"zipCode"`
	Currency string `json:"currency"`

	// seat
	SeatNo uint32 `json:"seatNo"`

	// Payment
	TotalAmtInUSCent uint64 `json:"totalAmtInUSCent"`
	PaymentID        string `json:"paymentID"`
}

type BookingResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message,omitempty"`
	Booking *Booking `json:"booking,omitempty"`
}
