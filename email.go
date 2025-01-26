// email.go

package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"encoding/json"
)

func generateEmailHTML(domain, post, unsubscribeLink string) string {
	// Generate the HTML content for the email
	return fmt.Sprintf(
		`<html>
			<body>
				<h1>New Post from %s</h1>
				<p><a href="https://%s/%s">%s</a></p>
				<p>Click the link to read the full post.</p>
				<p><a href="%s">Unsubscribe</a></p>
			</body>
		</html>`,
		domain, domain, post, post, unsubscribeLink,
	)
}

func sendEmail(email, domain, post, emailHTML string) {
	// Fetch Mailgun API key and host from Redis
	configJSON, err := rdb.Get(ctx, "config").Result()
	if err != nil {
		fmt.Println("Failed to retrieve Mailgun configuration")
		return
	}

	var config Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		fmt.Println("Failed to parse Mailgun configuration")
		return
	}

	// Fetch the reply-to email for the blog owner
	replyTo, err := rdb.Get(ctx, "reply-to:"+domain).Result()
	if err != nil {
		fmt.Println("Failed to retrieve reply-to email for domain", domain)
		replyTo = "no-reply@" + config.MailgunHost // Default fallback
	}

	// Prepare Mailgun API request
	mailgunAPI := "https://api.mailgun.net/v3/" + config.MailgunHost + "/messages"
	form := url.Values{}
	form.Add("from", "Quacker <no-reply@"+config.MailgunHost+">")
	form.Add("to", email)
	form.Add("subject", "New post from "+domain)
	form.Add("html", emailHTML)
	form.Add("h:Reply-To", replyTo)

	req, err := http.NewRequest("POST", mailgunAPI, strings.NewReader(form.Encode()))
	if err != nil {
		fmt.Println("Failed to create Mailgun request")
		return
	}
	req.SetBasicAuth("api", config.MailgunAPIKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send email: ", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Mailgun API error: %s\n", resp.Status)
		return
	}

	fmt.Printf("Email successfully sent to %s\n", email)
}
