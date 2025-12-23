package store

import (
	"sync"

	"github.com/ignius299792458/techkraft-ch-svr/model"
)

type IDEMPOTENCY_BUCKET struct {
	IDEMPOTENCY_STORE sync.Map // map[string]model.BookingOrder - IdempotencyKey -> BookingOrder

	// idempotency-level locks (idempotency key as key)
	idempotencyLocks sync.Map // map[string]*sync.Mutex
}

type Idempotency interface {
	HandleIdempotency(bookingOrderData model.BookingOrder) model.BookingOrder
	getIdempotencyKeyLock(idempotencyKey string) *sync.Mutex
}

func NewIdempotencyBucket() Idempotency {
	return &IDEMPOTENCY_BUCKET{
		IDEMPOTENCY_STORE: sync.Map{},
	}
}

// getIdempotencyKeyLock returns a mutex dedicated to a single idempotency key.
func (ib *IDEMPOTENCY_BUCKET) getIdempotencyKeyLock(idempotencyKey string) *sync.Mutex {
	lock, _ := ib.idempotencyLocks.LoadOrStore(idempotencyKey, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// HandleIdempotency attempts.
func (ib *IDEMPOTENCY_BUCKET) HandleIdempotency(
	bookingOrderData model.BookingOrder,
) model.BookingOrder {

	// acquire idempotency key-level lock
	idempotencyKeyLock := ib.getIdempotencyKeyLock(bookingOrderData.IdempotencyKey)
	idempotencyKeyLock.Lock()
	defer idempotencyKeyLock.Unlock()

	// ---- CRITICAL SECTION (idempotency key-scoped) ----

	// prevent non-idempotency booking
	if bookingOrderInterface, exists := ib.IDEMPOTENCY_STORE.Load(bookingOrderData.IdempotencyKey); exists {
		bookingOrder := bookingOrderInterface.(model.BookingOrder)
		if bookingOrderData.Status != model.BookingStatusPending {
			// update the stored booking status to confirmed
			bookingOrder.Status = bookingOrderData.Status
			ib.IDEMPOTENCY_STORE.Store(bookingOrderData.IdempotencyKey, bookingOrder)
		}
		return bookingOrder
	}

	// idempotency-key : bookingOrderData the booking
	ib.IDEMPOTENCY_STORE.Store(bookingOrderData.IdempotencyKey, bookingOrderData)

	return bookingOrderData
}
