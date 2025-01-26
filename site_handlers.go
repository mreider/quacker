// site_handlers.go

package main

import (
	"net/http"
	"strings"
	"regexp"
)

var emailReg = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func addSite(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if rdb.Get(ctx, "code:"+code).Err() != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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
	rdb.Set(ctx, "user:"+domain, replyTo, 0)
	http.Redirect(w, r, "/sites?success=Site+saved", http.StatusFound)
}
