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

func (t Tier) IsValidTier() bool {
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

func (t BookingStatus) IsValidBookingStatus() bool {
	switch t {
	case BookingStatusPending, BookingStatusConfirmed, BookingStatusFailed, BookingStatusCanceled:
		return true
	default:
		return false
	}
}

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusConfirmed PaymentStatus = "CONFIRMED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusCanceled  PaymentStatus = "CANCELED"
)

func (t PaymentStatus) IsValidPaymentStatus() bool {
	switch t {
	case PaymentStatusPending, PaymentStatusConfirmed, PaymentStatusFailed, PaymentStatusCanceled:
		return true
	default:
		return false
	}
}

type Booking struct {
	ID     uuid.UUID     `json:"id"`
	UserID string        `json:"userId"` // mocked user id
	Tier   Tier          `json:"tier"`
	Status BookingStatus `json:"status"`

	// Idempotency: retries of the same "Book" click should reuse this key.
	IdempotencyKey string `json:"idempotencyKey,omitempty"`

	// seat
	SeatNo uint32 `json:"seatNo"`

	// country
	Country  string `json:"country"`
	ZipCode  string `json:"zipCode"`
	Currency string `json:"currency"`

	// Payment
	TotalAmtInUSCent uint64        `json:"totalAmtInUSCent"`
	PaymentID        string        `json:"paymentID"`
	PaymentStatus    PaymentStatus `json:"paymentStatus"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type BookingOrder struct {
	UserID string        `json:"userId"` // mocked user id
	Tier   Tier          `json:"tier"`
	Status BookingStatus `json:"status"`

	IdempotencyKey string `json:"idempotencyKey,omitempty"`

	// country
	Country  string `json:"country"`
	ZipCode  string `json:"zipCode"`
	Currency string `json:"currency"`

	// seat
	SeatNo uint32 `json:"seatNo"`

	// Payment
	TotalAmtInUSCent uint64        `json:"totalAmtInUSCent"`
	PaymentID        string        `json:"paymentID"`
	PaymentStatus    PaymentStatus `json:"paymentStatus"`
}

type BookingResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message,omitempty"`
	Booking *Booking `json:"booking,omitempty"`
}

type AvailabilityResponse struct {
	Success        bool                `json:"success"`
	Message        string              `json:"message,omitempty"`
	AvaliableSeats map[string][]uint32 `json:"avaliableSeats,omitempty"`
	Tiers          []TierInfo          `json:"tiers,omitempty"`
}

type TierInfo struct {
	Tier          Tier     `json:"tier"`
	Price         uint64   `json:"price"`         // in US cents
	TotalSeats    uint32   `json:"totalSeats"`    // total seats for this tier
	ReservedCount uint32   `json:"reservedCount"` // number of seats reserved for this tier
	AvailableList []uint32 `json:"availableList,omitempty"`
}
