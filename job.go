// job.go

package main

import (
	"fmt"
	"net"
	"time"
	"io/ioutil"
	"net/http"
	"strings"
	"encoding/json"
)

func job() {
	// Ensure setup is complete before running the job
	ensureSetup()

	// Retrieve the configuration from Redis
	configJSON, err := rdb.Get(ctx, "config").Result()
	if err != nil {
		fmt.Println("Failed to retrieve configuration")
		return
	}
	var config Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		fmt.Println("Failed to parse configuration")
		return
	}

	// Get all user sites from Redis
	users, _ := rdb.Keys(ctx, "user:*").Result()
	for _, u := range users {
		domain := strings.TrimPrefix(u, "user:")

		// Check if the DNS TXT record is valid
		if !checkDNSRecord(domain) {
			rdb.Del(ctx, "user:"+domain)
			rdb.Del(ctx, "subs:"+domain)
			continue
		}

		// Get all subscribers for the site
		subs, _ := rdb.SMembers(ctx, "subs:"+domain).Result()

		// Fetch RSS feed posts
		posts := fetchRSS(domain)
		for _, post := range posts {
			// Send emails to subscribers if the post hasn't been sent before
			for _, subscriber := range subs {
				if !rdb.SIsMember(ctx, "sent:"+domain, post).Val() {
					unsubscribeLink := fmt.Sprintf("https://%s/unsubscribe?email=%s&domain=%s", config.Hostname, subscriber, domain)
					emailHTML := generateEmailHTML(domain, post, unsubscribeLink)
					sendEmail(subscriber, domain, post, emailHTML)
					rdb.SAdd(ctx, "sent:"+domain, post)
				}
			}
		}
	}

	// Clean up old sent records
	cleanSent()
}

func checkDNSRecord(domain string) bool {
	// Verify if the DNS TXT record contains "quack-quack"
	records, err := net.LookupTXT(domain)
	if err != nil {
		return false
	}
	for _, record := range records {
		if record == "quack-quack" {
			return true
		}
	}
	return false
}

func fetchRSS(domain string) []string {
	// Fetch the RSS feed for the domain
	resp, err := http.Get("https://" + domain + "/rss")
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	// Placeholder: Extract posts from the RSS feed body
	return extractLinksFromRSS(body)
}

func extractLinksFromRSS(body []byte) []string {
	// Placeholder logic for extracting links from RSS feed
	return []string{"post1", "post2", "post3"}
}

func cleanSent() {
	// Remove sent records older than 4 days
	expiration := time.Now().Add(-96 * time.Hour).Unix()
	keys, _ := rdb.Keys(ctx, "sent:*").Result()
	for _, key := range keys {
		timestamp, _ := rdb.Get(ctx, key).Int64()
		if timestamp < expiration {
			rdb.Del(ctx, key)
		}
	}
}
