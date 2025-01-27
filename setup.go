// setup.go

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Config struct {
	MailgunAPIKey      string
	MailgunHost        string
	Hostname           string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURI  string
}

func setup() {
	// Prompt user for test mode
	fmt.Print("Use test hostname (localhost)? (y/n): ")
	var useTest string
	fmt.Scanln(&useTest)
	useTest = strings.ToLower(useTest)

	var config Config

	if useTest == "y" {
		// Default to localhost for testing
		fmt.Println("Test mode enabled. Using 'localhost' as the hostname.")
		config.Hostname = "localhost"
	} else {
		fmt.Print("Hostname for Quacker (example.com): ")
		fmt.Scanln(&config.Hostname)

		// Validate the inputs for correctness
		if !validateHostname(config.Hostname) {
			fmt.Println("Invalid hostname")
			os.Exit(1)
		}
	}

	// Prompt the user for the Mailgun API key and hostname
	fmt.Print("Mailgun API Key: ")
	fmt.Scanln(&config.MailgunAPIKey)
	fmt.Print("Mailgun Host (sandbox.mailgun.org): ")
	fmt.Scanln(&config.MailgunHost)

	// Prompt for GitHub OAuth credentials
	fmt.Print("GitHub Client ID: ")
	fmt.Scanln(&config.GitHubClientID)
	fmt.Print("GitHub Client Secret: ")
	fmt.Scanln(&config.GitHubClientSecret)
	fmt.Print("GitHub Redirect URI (e.g., http://yourdomain.com/login/callback): ")
	fmt.Scanln(&config.GitHubRedirectURI)

	// Validate the GitHub OAuth credentials
	if !validateGitHubOAuth(config.GitHubClientID, config.GitHubClientSecret, config.GitHubRedirectURI) {
		fmt.Println("Invalid GitHub OAuth credentials or Redirect URI.")
		os.Exit(1)
	}

	// Serialize the configuration and save it to Redis
	configJSON, err := json.Marshal(config)
	if err != nil {
		fmt.Println("Failed to serialize config")
		os.Exit(1)
	}

	if err := rdb.Set(ctx, "config", configJSON, 0).Err(); err != nil {
		fmt.Println("Failed to save config")
		os.Exit(1)
	}

	fmt.Println("Setup complete. The server will run on port 8085 using HTTP.")
}

func validateHostname(hostname string) bool {
	// Ensure the hostname is a valid domain name
	validHostname := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	return validHostname.MatchString(hostname)
}

func validateGitHubOAuth(clientID, clientSecret, redirectURI string) bool {
	// Mock validation of GitHub OAuth credentials
	return clientID != "" && clientSecret != "" && strings.HasPrefix(redirectURI, "http")
}
