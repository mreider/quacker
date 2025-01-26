// menu_helper.go

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
		<div class="collapse navbar-collapse">
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
			<div class="navbar-text ml-auto">
	`

	if userLoggedIn {
		menu += `
				Logged in: ` + loggedInUser + ` |
				<a href="/logout" class="nav-link d-inline">Logout</a>
			`
	}

	menu += `
			</div>
		</div>
	</nav>`

	return menu
}
