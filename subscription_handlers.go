// subscription_handlers.go

package main

import (
	"fmt"
	"net/http"
)

func subscribe(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	if !emailReg.MatchString(email) {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}
	ref := r.Header.Get("Referer")
	if ref == "" || rdb.Get(ctx, "user:"+ref).Err() != nil {
		http.Error(w, "Unsupported domain", http.StatusForbidden)
		return
	}
	rdb.SAdd(ctx, "subs:"+ref, email)
	w.Write([]byte("Subscribed"))
}

func unsubscribe(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	domain := r.URL.Query().Get("domain")
	if email == "" || domain == "" {
		renderErrorPage(w, r, "Missing email or domain")
		return
	}
	if rdb.SRem(ctx, "subs:"+domain, email).Val() > 0 {
		w.Write([]byte(fmt.Sprintf(`<!DOCTYPE html>
		<html>
		<head>
			<link rel="stylesheet" href="/assets/css/bootstrap.min.css">
			<script src="/assets/js/bootstrap.bundle.min.js"></script>
		</head>
		<body>
			<div class="container mt-5">
				%s
				<div class="card">
					<div class="card-body">
						<h1 class="mb-4">Unsubscribed</h1>
						<p><strong>%s</strong> has been unsubscribed from <strong>%s</strong>.</p>
						<p>If you believe this was a mistake, you can <a href="%s" class="btn btn-link">re-subscribe</a>.</p>
					</div>
				</div>
				<footer class="mt-4 text-center">
					<small>&copy; 2025 Matthew Reider - <a href="https://github.com/mreider" target="_blank">GitHub</a></small>
				</footer>
			</div>
		</body>
		</html>`, renderMenu(r), email, domain, domain)))
	} else {
		renderErrorPage(w, r, "Subscription not found")
	}
}
