// router.go

package main

import (
	_ "embed"
	"net/http"
	"github.com/gorilla/mux"
)

// Embed assets into the binary

//go:embed assets/css/bootstrap.min.css
var bootstrapCSS string

//go:embed assets/js/bootstrap.bundle.min.js
var bootstrapJS string

//go:embed assets/css/bootstrap.min.css.map
var bootstrapMap string

//go:embed assets/favicon.ico
var favicon []byte

func setupRouter() *mux.Router {
	r := mux.NewRouter()

	// Update the home page to point to the about page
	r.HandleFunc("/", aboutPage).Methods("GET")
	
	// Update site list and invitation handlers
	r.HandleFunc("/sites", siteListPage).Methods("GET")
	r.HandleFunc("/validate", withFloodControl(validateInvitationCodeHandler)).Methods("POST")

	// Other routes
	r.HandleFunc("/addsite", withFloodControl(addSite)).Methods("POST")
	r.HandleFunc("/js/{domain}", withFloodControl(serveJS)).Methods("GET")
	r.HandleFunc("/subscribe", withFloodControl(subscribe)).Methods("POST")
	r.HandleFunc("/unsubscribe", withFloodControl(unsubscribe)).Methods("GET")

	// Serve embedded assets
	r.HandleFunc("/assets/css/bootstrap.min.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write([]byte(bootstrapCSS))
	})

	r.HandleFunc("/assets/js/bootstrap.bundle.min.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write([]byte(bootstrapJS))
	})

	r.HandleFunc("/assets/css/bootstrap.min.css.map", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bootstrapMap))
	})

	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(favicon)
	})

	// Handle HTTP errors
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderErrorPage(w, "404 - Page Not Found")
	})

	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderErrorPage(w, "405 - Method Not Allowed")
	})

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					renderErrorPage(w, "500 - Internal Server Error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	})

	return r
}
