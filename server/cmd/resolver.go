package main

import (
	"net/http"

	"github.com/ignius299792458/dbi-svr/router"
)

func resolver(mux *http.ServeMux) {

	// booking module
	bookingMux := http.NewServeMux()
	router.BookingRouter(bookingMux)
	mux.Handle("/booking/", http.StripPrefix("/booking", bookingMux))
}
