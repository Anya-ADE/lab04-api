package main

import (
	"log"
	"net/http"
	"time"
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

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Wrap the real writer with our custom one, pre-set to 200
		rw := &responseWriter{w, http.StatusOK}
		// Pass the wrapped writer to the next handler
		next.ServeHTTP(rw, r)
		// The handler has now finished. rw.statusCode holds the real status.
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, rw.statusCode, time.Since(start))
	})
}

/* What do time.Now() and time.Since(start) give you together? */

/*Answer: time.Now() captures the current time at the moment it is called,
and time.Since(start) calculates the duration that has elapsed since the captured start time.
Together, they provide a way to measure how long a particular operation or request took to complete. */

/*Why do we pass rw (not w) into next.ServeHTTP?*/

/*Answer: We pass rw (the wrapped response writer) into next.ServeHTTP instead of w
(the original response writer) because rw is designed to capture the status code that the handler sets.
By using rw, we can intercept calls to WriteHeader and store the status code in rw.statusCode, allowing us to log it later.
If we passed w directly, we would not be able to capture the status code, and our logging
would not reflect the actual response status sent to the client. */

/*What would rw.statusCode be if the handler sent a 404?*/

/*Answer: If the handler sent a 404 status code, rw.statusCode would be set to 404.
This is because the WriteHeader method of our custom responseWriter captures the status code whenever it is called,
so when the handler calls w.WriteHeader(http.StatusNotFound), rw.statusCode will be updated to reflect that value. */

/*The middleware is applied once in main, wrapping the entire router. If you have five handlers,
how many times does each request pass through the middleware?*/

/*
Answer: Each request will pass through the middleware once. The middleware is applied to the entire router,
so regardless of which handler is invoked for a given request, it will go through the middleware exactly one
time before reaching the specific handler. The middleware acts as a single layer that processes all incoming
requests before they are routed to their respective handlers.
*/
func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		rw.Write([]byte("ok"))
		println("Captured Status Code:", rw.statusCode)
	})
	logMiddleware := loggingMiddleware(mux)

	err := http.ListenAndServe(":4000", logMiddleware)
	log.Printf("Server started on http://localhost:4000")

	if err != nil {
		panic(err)
	}

}
