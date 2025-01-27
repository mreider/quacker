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

	// Middleware to check if user is logged in
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if needsAuthentication(r.URL.Path) {
				if _, err := r.Cookie("quacker_user"); err != nil {
					renderErrorPage(w, r, "401 - Unauthorized: Please log in.")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	})

	// Update the home page to point to the about page
	r.HandleFunc("/", aboutPage).Methods("GET")

	// Add GitHub OAuth routes
	r.HandleFunc("/login/github", handleGitHubLogin).Methods("GET")
	r.HandleFunc("/login/callback", handleGitHubCallback).Methods("GET")

	// Update site list handlers
	r.HandleFunc("/sites", siteListPage).Methods("GET")

	// Other routes
	r.HandleFunc("/addsite", addSite).Methods("POST")
	r.HandleFunc("/js/{domain}", serveJS).Methods("GET")
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
		renderErrorPage(w, r, "404 - Page Not Found")
	})

	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderErrorPage(w, r, "405 - Method Not Allowed")
	})

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					renderErrorPage(w, r, "500 - Internal Server Error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	})

	return r
}

func needsAuthentication(path string) bool {
	// Define routes that require authentication
	protectedRoutes := []string{"/sites", "/addsite", "/js/"}
	for _, route := range protectedRoutes {
		if len(path) >= len(route) && path[:len(route)] == route {
			return true
		}
	}
	return false
}
