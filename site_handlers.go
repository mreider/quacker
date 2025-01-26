// site_handlers.go

package main

import (
	"net/http"
	"regexp"
	"strings"
)

var emailReg = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func addSitePage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<!DOCTYPE html>
	<html>
	<head>
		<link rel="stylesheet" href="/assets/css/bootstrap.min.css">
		<script src="/assets/js/bootstrap.bundle.min.js"></script>
	</head>
	<body>
		<div class="container mt-5">
			` + renderMenu(r) + `
			<div class="card">
				<div class="card-body">
					<h1 class="mb-4">Add a New Site</h1>
					<form action="/addsite" method="POST">
						<div class="mb-3">
							<label for="rss" class="form-label">RSS Feed URL</label>
							<input type="url" class="form-control" id="rss" name="rss" placeholder="Enter RSS feed URL" required>
						</div>
						<div class="mb-3">
							<label for="replyto" class="form-label">Reply-To Email</label>
							<input type="email" class="form-control" id="replyto" name="replyto" placeholder="Enter reply-to email" required>
						</div>
						<button type="submit" class="btn btn-primary">Save Site</button>
					</form>
				</div>
			</div>
			<footer class="mt-4 text-center">
				<small>&copy; 2025 Matthew Reider - <a href="https://github.com/mreider">GitHub</a></small>
			</footer>
		</div>
	</body>
	</html>`))
}

func addSite(w http.ResponseWriter, r *http.Request) {
	rss := r.FormValue("rss")
	replyTo := r.FormValue("replyto")

	if !emailReg.MatchString(replyTo) {
		http.Redirect(w, r, "/sites?error=Invalid+email+address", http.StatusFound)
		return
	}

	resp, err := http.Get(rss)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Redirect(w, r, "/sites?error=Invalid+RSS+URL", http.StatusFound)
		return
	}
	defer resp.Body.Close()
	domain := strings.Split(rss, "/")[2]
	cookie, err := r.Cookie("quacker_user")
	if err != nil {
		renderErrorPage(w, r, "401 - Unauthorized: Please log in.")
		return
	}
	loggedInUser := cookie.Value
	rdb.Set(ctx, "user_sites:"+loggedInUser+":"+domain, replyTo, 0)
	http.Redirect(w, r, "/sites?success=Site+saved", http.StatusFound)
}
