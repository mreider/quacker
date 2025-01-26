// error_handler.go

package main

import (
	"fmt"
	"net/http"
)

func renderErrorPage(w http.ResponseWriter, message string) {
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
					<h1 class="mb-4">Error</h1>
					<p class="text-danger">%s</p>
				</div>
			</div>
			<footer class="mt-4 text-center">
				<small>&copy; 2025 Matthew Reider - <a href="https://github.com/mreider" target="_blank">GitHub</a></small>
			</footer>
		</div>
	</body>
	</html>`, renderMenu(), message)))
}
