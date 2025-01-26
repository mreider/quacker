// invitation_handler.go

package main

import (
	"net/http"
)

func renderInvitationForm(errorMessage, successMessage string) string {
	formHTML := `
	<div class="card mb-4">
		<div class="card-body">
			<h2 class="mb-4">Enter Invitation Code</h2>`

	if errorMessage != "" {
		formHTML += `<div class="alert alert-danger" role="alert">` + errorMessage + `</div>`
	}

	if successMessage != "" {
		formHTML += `<div class="alert alert-success" role="alert">` + successMessage + `</div>`
	}

	formHTML += `
			<form method="post" action="/validate">
				<div class="form-group mb-3">
					<label for="code">Invitation Code:</label>
					<input type="text" class="form-control" id="code" name="code" placeholder="Enter your invitation code" required>
				</div>
				<div class="d-grid">
					<button type="submit" class="btn btn-primary">Validate</button>
				</div>
			</form>
		</div>
	</div>`

	return formHTML
}

func validateInvitationCodeHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if rdb.Get(ctx, "code:"+code).Err() != nil {
		http.Redirect(w, r, "/sites?error=Invalid+invitation+code", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/sites?success=Valid+invitation+code", http.StatusFound)
}
