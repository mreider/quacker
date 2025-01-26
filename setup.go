// setup.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Config struct {
	MailgunAPIKey    string
	MailgunHost      string
	Hostname         string
	GitHubClientID   string
	GitHubClientSecret string
	GitHubRedirectURI string
}

func setup() {
	// Prompt user for test mode
	fmt.Print("Use test certificates (locahost)? (y/n): ")
	var useTest string
	fmt.Scanln(&useTest)
	useTest = strings.ToLower(useTest)

	var config Config

	if useTest == "y" {
		// Default to localhost for testing
		fmt.Println("Test mode enabled. Using 'localhost' as the hostname.")
		config.Hostname = "localhost"

		// Provide ngrok instructions for GitHub OAuth testing
		fmt.Println("\nTIP: To test GitHub OAuth on localhost, consider using a tunneling tool like ngrok:")
		fmt.Println("1. Install ngrok (https://ngrok.com/download).")
		fmt.Println("2. Run: ngrok http 8080")
		fmt.Println("3. Use the provided ngrok URL (e.g., https://abc123.ngrok.io) as your GitHub Redirect URI.")
		fmt.Println("4. Update the Redirect URI in your GitHub App settings accordingly.")

		// Generate self-signed certificate for localhost
		err := generateSelfSignedCert()
		if err != nil {
			fmt.Println("Failed to generate self-signed certificates:", err)
			os.Exit(1)
		}
	} else {
		fmt.Print("Hostname for Quacker (example.com): ")
		fmt.Scanln(&config.Hostname)

		// Validate the inputs for correctness
		if !validateHostname(config.Hostname) {
			fmt.Println("Invalid hostname")
			os.Exit(1)
		}

		// Generate certificates using Certbot
		certbotArgs := []string{"certonly", "--standalone", "--non-interactive", "--agree-tos", "-d", config.Hostname, "--email", "admin@" + config.Hostname}
		certbotCmd := exec.Command("certbot", certbotArgs...)
		certbotCmd.Stdout = os.Stdout
		certbotCmd.Stderr = os.Stderr
		if err := certbotCmd.Run(); err != nil {
			fmt.Println("Failed to generate certificates with Certbot")
			os.Exit(1)
		}

		// Check if certificates were successfully created
		if _, err := os.Stat("/etc/letsencrypt/live/" + config.Hostname + "/fullchain.pem"); os.IsNotExist(err) {
			fmt.Println("Certificate files not found. Certbot may have failed.")
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

	fmt.Println("Setup complete. Certificates are stored in /etc/letsencrypt/live/" + config.Hostname)
}

func generateSelfSignedCert() error {
	// Create self-signed certificates for localhost
	cmd := exec.Command("openssl", "req", "-x509", "-newkey", "rsa:4096", "-sha256", "-days", "365",
		"-nodes", "-keyout", "key.pem", "-out", "cert.pem",
		"-subj", "/CN=localhost")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func validateHostname(hostname string) bool {
	// Ensure the hostname is a valid domain name
	validHostname := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	return validHostname.MatchString(hostname)
}

func validateGitHubOAuth(clientID, clientSecret, redirectURI string) bool {
	url := "https://github.com/login/oauth/access_token"
	payload := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"redirect_uri":  redirectURI,
		"code":          "mock-code", // Mock code to test validation
	}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return false
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	return true
}
