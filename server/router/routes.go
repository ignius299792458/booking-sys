package router

import (
	"net/http"

	"github.com/ignius299792458/techkraft-ch-svr/handlers"
)

func BookingRouter(bookingMux *http.ServeMux) {

	// bookingMux.HandleFunc("GET /seats", getSeats)

	bookingMux.HandleFunc("POST /book", handlers.HandleBooking)
}
