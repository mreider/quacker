// serve_js_handler.go

package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
)

func serveJS(w http.ResponseWriter, r *http.Request) {
	domain := mux.Vars(r)["domain"]
	configJSON, err := rdb.Get(ctx, "config").Result()
	if err != nil {
		http.Error(w, "Configuration not found", http.StatusInternalServerError)
		return
	}
	var config Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		http.Error(w, "Invalid configuration format", http.StatusInternalServerError)
		return
	}

	if rdb.Get(ctx, "user:"+domain).Err() != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	jsCode := fmt.Sprintf(`<script>
	function subscribe(email) {
		fetch('%s/subscribe', {
			method: 'POST',
			headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
			body: 'email=' + encodeURIComponent(email)
		}).then(response => response.text()).then(alert);
	}
	</script>
	<form onsubmit="event.preventDefault(); subscribe(this.email.value);">
		<input type="email" name="email" required placeholder="Your email" class="form-control">
		<button type="submit" class="btn btn-primary mt-2">Subscribe</button>
	</form>`, config.Hostname)

	w.Write([]byte(fmt.Sprintf(`<!DOCTYPE html>
	<html>
	<head>
		<link rel="stylesheet" href="/assets/css/bootstrap.min.css">
		<script src="/assets/js/bootstrap.bundle.min.js"></script>
	</head>
	<body>
		<div class="container mt-5">
		` + renderMenu() + `
			<div class="card">
				<div class="card-body">
					<h2>Copy and Paste This JS to Your Static Site</h2>
					<p>Add this code to the header of your static site. For instructions, visit <a href="https://mreider.com/quacker/">mreider.com/quacker</a>.</p>
					<div class="mb-3">
						<textarea id="js-code" class="form-control" rows="10" readonly>%s</textarea>
					</div>
					<button class="btn btn-primary" onclick="navigator.clipboard.writeText(document.getElementById('js-code').value)">Copy to Clipboard</button>
				</div>
			</div>
		</div>
	</body>
	</html>`, jsCode)))
}
