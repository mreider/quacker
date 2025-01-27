// about_handler.go

package main

import (
	"net/http"
)

func aboutPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<!DOCTYPE html>
	<html>
	<head>
		<link rel="stylesheet" href="/assets/css/bootstrap.min.css">
		<script src="/assets/js/bootstrap.bundle.min.js"></script>
	</head>
	<body>
		<div class="container mt-5">
			`  + renderMenu(r) + `
			<div class="card">
				<div class="card-body">
					<h1>Welcome to Quacker</h1>
					<p>Quacker is a service designed to add subscription capabilities to static websites generated by tools like Jekyll or Hugo.</p>
					<a href="/login/github" class="btn btn-primary mt-3">Login with GitHub</a>
				</div>
			</div>

			<footer class="mt-4 text-center">
				<small>&copy; 2025 Matthew Reider - <a href="https://github.com/mreider">GitHub</a></small>
			</footer>
		</div>
	</body>
	</html>`))
}
