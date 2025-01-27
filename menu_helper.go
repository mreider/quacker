package main

import "net/http"

func renderMenu(r *http.Request) string {
	cookie, err := r.Cookie("quacker_user")
	userLoggedIn := (err == nil)
	loggedInUser := ""
	if userLoggedIn {
		loggedInUser = cookie.Value
	}

	menu := `
	<nav class="navbar navbar-expand-lg navbar-light bg-light mb-4">
		<a class="navbar-brand" href="/">🦆 Quacker</a>
		<button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav" aria-controls="navbarNav" aria-expanded="false" aria-label="Toggle navigation">
			<span class="navbar-toggler-icon"></span>
		</button>
		<div class="collapse navbar-collapse" id="navbarNav">
			<ul class="navbar-nav mr-auto">
	`

	if userLoggedIn {
		menu += `
				<li class="nav-item">
					<a class="nav-link" href="/sites">Sites</a>
				</li>
			`
	}

	menu += `
			</ul>
			<div class="navbar-text ms-auto">
	`

	if userLoggedIn {
		menu += `
				<a href="/logout" class="nav-link">Logout ` + loggedInUser + `</a>
			`
	}

	menu += `
			</div>
		</div>
	</nav>`

	return menu
}
