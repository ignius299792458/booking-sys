package store

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ignius299792458/techkraft-ch-svr/model"
)

/*
* VIP, FRONT_ROW, GA
  - seats can be set manual arr, but for now range is used

* VIP 	:- first 30 (seats 1-30)
* FRONT_ROW 	:- 31 to 60 (seats 31-60)
* GA 		:- 61 to 100 (seats 61-100)
* TOTAL_SEAT = 100
*/
type BOOKING_STORE_BUCKET struct {
	BOOKING_STORE map[uint32]model.Booking // SeatNo -> Booking
	TOTAL_SEAT    uint32

	// seat-level locks (seat number as key)
	seatLocks sync.Map // map[uint32]*sync.Mutex
}

type BookingStore interface {
	RegisterBooking(bookingOrderData model.BookingOrder) (model.Booking, error)
	GetBooking(seatNo uint32) (model.Booking, error)
	getSeatLock(seatNo uint32) *sync.Mutex
	GetReservedSeats() map[string][]uint32
}

func NewBookingStoreBucket() BookingStore {
	return &BOOKING_STORE_BUCKET{
		BOOKING_STORE: make(map[uint32]model.Booking),
		TOTAL_SEAT:    100,
	}
}

// getSeatLock returns a mutex dedicated to a single seat.
func (b *BOOKING_STORE_BUCKET) getSeatLock(seatNo uint32) *sync.Mutex {
	lock, _ := b.seatLocks.LoadOrStore(seatNo, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// RegisterBooking attempts to register a new booking.
func (b *BOOKING_STORE_BUCKET) RegisterBooking(
	bookingOrderData model.BookingOrder,
) (model.Booking, error) {

	// basic validation (cheap checks first)
	if bookingOrderData.SeatNo == 0 || bookingOrderData.SeatNo > b.TOTAL_SEAT {
		return model.Booking{}, errors.New("invalid seat number")
	}

	// acquire seat-level lock
	seatBookingLock := b.getSeatLock(bookingOrderData.SeatNo)
	seatBookingLock.Lock()
	defer seatBookingLock.Unlock()

	// ---- CRITICAL SECTION (seat-scoped) ----

	// prevent double booking
	if _, exists := b.BOOKING_STORE[bookingOrderData.SeatNo]; exists {
		return model.Booking{}, errors.New("seat already booked")
	}

	newBooking := model.Booking{
		ID:     uuid.New(),
		UserID: bookingOrderData.UserID,
		Tier:   bookingOrderData.Tier,
		Status: bookingOrderData.Status,

		IdempotencyKey: bookingOrderData.IdempotencyKey,

		SeatNo: bookingOrderData.SeatNo,

		Country:  bookingOrderData.Country,
		ZipCode:  bookingOrderData.ZipCode,
		Currency: bookingOrderData.Currency,

		TotalAmtInUSCent: bookingOrderData.TotalAmtInUSCent,
		PaymentID:        bookingOrderData.PaymentID,
		PaymentStatus:    bookingOrderData.PaymentStatus,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// if paymentId isn't nil or empty and paymentStatus is confirmed, set booking status to confirmed
	if len(newBooking.PaymentID) > 0 && newBooking.PaymentStatus == model.PaymentStatusConfirmed {
		newBooking.Status = model.BookingStatusConfirmed
	}

	// if paymentStatus is failed or canceled, set booking status to canceled
	if newBooking.PaymentStatus == model.PaymentStatusFailed || newBooking.PaymentStatus == model.PaymentStatusCanceled {
		newBooking.Status = model.BookingStatusCanceled
	}

	b.BOOKING_STORE[newBooking.SeatNo] = newBooking

	return newBooking, nil
}

// GetReservedSeats returns a map of reserved seat numbers categorized by tier.
func (b *BOOKING_STORE_BUCKET) GetReservedSeats() map[string][]uint32 {
	var VIPReservedSeats []uint32
	var FRONTROWReservedSeats []uint32
	var GAReservedSeats []uint32

	for seatNo, booking := range b.BOOKING_STORE {
		switch booking.Tier {
		case model.TierVIP:
			VIPReservedSeats = append(VIPReservedSeats, seatNo)
		case model.TierFrontRow:
			FRONTROWReservedSeats = append(FRONTROWReservedSeats, seatNo)
		case model.TierGA:
			GAReservedSeats = append(GAReservedSeats, seatNo)
		}
	}

	return map[string][]uint32{
		"VIP":       VIPReservedSeats,
		"FRONT_ROW": FRONTROWReservedSeats,
		"GA":        GAReservedSeats,
	}
}

func (b *BOOKING_STORE_BUCKET) GetBooking(seatNo uint32) (model.Booking, error) {
	booking, exists := b.BOOKING_STORE[seatNo]
	if !exists {
		return model.Booking{}, errors.New("booking not found")
	}
	return booking, nil
}
