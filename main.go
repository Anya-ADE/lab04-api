package main

import (
	"net/http"
)

/* If a handler calls w.Write([]byte("ok"))
without calling w.WriteHeader() first, what status code
does the client receive? What will rw.statusCode hold if
you pre-set it to http.StatusOK? */

/*Answer: If a handler calls w.Write() without an explicit call to w.WriteHeader(),
the client receives a status code of 200 OK. If the pre-set rw.statusCode is set to
http.StatusOK, it will hold the value 200.*/

type responseWriter struct {
	http.ResponseWriter     // embed the real writer
	statusCode          int // captured status code
}

// WriteHeader intercepts the status code before forwarding it.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		rw.Write([]byte("ok"))
		println("Captured Status Code:", rw.statusCode)
	})

	err := http.ListenAndServe(":4000", mux)
	if err != nil {
		panic(err)
	}

}
