package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
)

/*Why is it better to pass the logger through a struct rather than declaring var logger =
slog.New(...) at the top of the file?*/

/*Answer: Passing the logger through a struct rather than declaring allows Dependency Injection:
Using a struct allows for easier dependency injection, making it simpler to test your code.
You can create instances of the struct with different logger configurations for testing purposes without affecting the global state.*/

type application struct {
	logger *slog.Logger
}

/* If a handler calls w.Write([]byte("ok"))
without calling w.WriteHeader() first, what status code
does the client receive? What will rw.statusCode hold if
you pre-set it to http.StatusOK? */

/*Answer: If a handler calls w.Write() without an explicit call to w.WriteHeader(),
the client receives a status code of 200 OK. If the pre-set rw.statusCode is set to
http.StatusOK, it will hold the value 200.*/

func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "status: available\n")
	app.logger.Info("healthcheck handler called")
}

func (app *application) listBooks(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "list of books (coming soon)\n")
	app.logger.Info("listBooks handler called")
}

func (app *application) getBook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "get book with id: %s\n", id)
	app.logger.Info("getBook handler called", "id", id)
}

func (app *application) createBook(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "book created (coming soon)\n")
	app.logger.Info("createBook handler called")
}

func (app *application) deleteBook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	w.WriteHeader(http.StatusNoContent)
	app.logger.Info("deleteBook handler called", "id", id)

	/*Why is app a pointer (&application{}) rather than a value?*/

	/*Answer: app is a pointer to an application struct rather than a value because it allows
	for more efficient memory usage and easier modification of the struct's fields.
	Using a pointer means that when we pass app around in our code, we are passing a reference
	to the same underlying data rather than creating copies of the struct.*/

	/*How does each handler method access the logger — what does the receiver give you?*/

	/*Answer: Each handler method can access the logger through the application struct's receiver.
	The receiver allows the handler methods to access the fields of the application struct, including the logger.*/

	/*If you later added a second dependency (e.g. a database), where exactly would you add it?*/

	/*Answer: If I later added a second dependency, such as a database, I would add it as a field in the application struct.
	This way, all dependencies are centralized within the application struct, making it easier to manage and pass around as needed.*/

}

/*Why does deleteBook not write a response body? What does HTTP 204 communicate to the
client?*/

/*Answer: The deleteBook handler does not write a response body because HTTP 204 No Content
indicates that the server successfully processed the request, but there is no content to send in the response.*/

type responseWriter struct {
	http.ResponseWriter     // embed the real writer
	statusCode          int // captured status code
}

// WriteHeader intercepts the status code before forwarding it.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
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

/*Answer: Each request will pass through the middleware once. The middleware is applied to the entire router,
so regardless of which handler is invoked for a given request, it will go through the middleware exactly one
time before reaching the specific handler. The middleware acts as a single layer that processes all incoming
requests before they are routed to their respective handlers.*/

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

func main() {
	// Create a structured logger writing to stdout
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// Inject the logger into the application struct
	app := &application{
		logger: logger,
	}
	// Register routes with method-qualified patterns (Go 1.22+)
	mux := http.NewServeMux()

	/*what does the method-qualified pattern "GET /v1/healthcheck" do?*/

	/*Answer: The method-qualified pattern "GET /v1/healthcheck" specifies that the handler should only be invoked for
	GET requests to the /v1/healthcheck endpoint. This allows for more precise routing, ensuring that the handler is only
	called when the correct HTTP method is used*/

	mux.HandleFunc("GET /v1/healthcheck", app.healthcheck)

	/*what does the method-qualified pattern "GET /v1/books/" do?*/
	/*Answer: The method-qualified pattern "GET /v1/books/" specifies that the handler should only be invoked for
	GET requests to the /v1/books/ endpoint. This allows the handler to respond specifically to requests that are
	intended to list books, ensuring that it is not called for other HTTP methods or endpoints.*/
	mux.HandleFunc("GET /v1/books", app.listBooks)

	/*what does the method-qualified pattern "GET /v1/books/{id}" do?*/
	/*Answer: The method-qualified pattern "GET /v1/books/{id}" specifies that the handler should only be invoked for
	GET requests to the /v1/books/{id} endpoint, where {id} is a placeholder for a variable part of the URL.
	This allows the handler to extract the id value from the URL and use it to retrieve the specific book information.*/

	mux.HandleFunc("GET /v1/books/{id}", app.getBook)

	/*what does the method-qualified pattern "POST /v1/books" do?*/
	/*Answer: The method-qualified pattern "POST /v1/books" specifies that the handler should only be invoked for
	POST requests to the /v1/books endpoint. This allows the handler to respond specifically to requests that are
	intended to create a new book, ensuring that it is not called for other HTTP methods or endpoints.*/
	mux.HandleFunc("POST /v1/books", app.createBook)

	/*what does the method-qualified pattern "DELETE /v1/books/{id}" do?*/
	/*Answer: The method-qualified pattern "DELETE /v1/books/{id}" specifies that the handler should only be invoked for
	DELETE requests to the /v1/books/{id} endpoint, where {id} is a placeholder for a variable part of the URL.
	This allows the handler to extract the id value from the URL and use it to delete the specific book information.*/
	mux.HandleFunc("DELETE /v1/books/{id}", app.deleteBook)
	// Log a message before the server starts
	logger.Info("starting server", "addr", ":4000")
	// Wrap the entire router with logging middleware
	err := http.ListenAndServe(":4000", loggingMiddleware(mux))
	log.Fatal(err)

	/*For Test 6 you did not write a 404 handler. Where did the 404 come from? What status code
	does the middleware log, and what does that confirm about your custom response writer?*/

	/*Answer: The 404 status code comes from the default behavior of the http.ServeMux when
	a request does not match any registered route. The middleware logs the 404 status code,
	which confirms that our custom response writer is correctly capturing the status code set by
	the http.ServeMux when it returns a 404 for unmatched routes. This shows that our response writer
	is functioning as intended, allowing us to log the actual status code sent to the client even when
	it is generated by the default routing behavior.*/
}
