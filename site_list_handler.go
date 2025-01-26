// site_list_handler.go

package main

import (
	"fmt"
	"net/http"
	"strings"
)

func siteListPage(w http.ResponseWriter, r *http.Request) {
	// Get logged-in user from the session cookie
	cookie, err := r.Cookie("quacker_user")
	if err != nil {
		renderErrorPage(w, r, "401 - Unauthorized: Please log in.")
		return
	}
	loggedInUser := cookie.Value

	// Fetch sites for the logged-in user only
	sites, _ := rdb.Keys(ctx, "user_sites:"+loggedInUser+":*").Result()

	errorMessage := r.URL.Query().Get("error")
	successMessage := r.URL.Query().Get("success")

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
					<h1 class="mb-4">List of Sites</h1>
					<a href="/addsite" class="btn btn-primary mb-4">Add New Site</a>`))

	if errorMessage != "" {
		w.Write([]byte(`<div class="alert alert-danger" role="alert">` + errorMessage + `</div>`))
	}

	if successMessage != "" {
		w.Write([]byte(`<div class="alert alert-success" role="alert">` + successMessage + `</div>`))
	}

	w.Write([]byte(`
					<table class="table table-striped">
						<thead>
							<tr>
								<th>Domain</th>
								<th>Actions</th>
							</tr>
						</thead>
						<tbody>`))

	if len(sites) == 0 {
		w.Write([]byte(`<tr><td colspan="2">No sites available</td></tr>`))
	} else {
		for _, siteKey := range sites {
			domain := strings.TrimPrefix(siteKey, "user_sites:"+loggedInUser+":")
			w.Write([]byte(fmt.Sprintf(`<tr>
				<td>%s</td>
				<td><a href="/js/%s" class="btn btn-info btn-sm">JS</a></td>
			</tr>`, domain, domain)))
		}
	}

	w.Write([]byte(`
						</tbody>
					</table>
				</div>
			</div>
			<footer class="mt-4 text-center">
				<small>&copy; 2025 Matthew Reider - <a href="https://github.com/mreider">GitHub</a></small>
			</footer>
		</div>
	</body>
	</html>`))
}
