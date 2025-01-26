// setup.go

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Config struct {
	MailgunAPIKey string
	MailgunHost   string
	Hostname      string
}

func setup() {
	// Prompt the user for the Mailgun API key and hostname
	fmt.Print("Mailgun API Key: ")
	var config Config
	fmt.Scanln(&config.MailgunAPIKey)
	fmt.Print("Mailgun Host (sandbox.mailgun.org): ")
	fmt.Scanln(&config.MailgunHost)

	// Prompt user for test mode
	fmt.Print("Use test certificates? (y/n): ")
	var useTest string
	fmt.Scanln(&useTest)
	useTest = strings.ToLower(useTest)

	if useTest == "y" {
		// Default to localhost for testing
		fmt.Println("Test mode enabled. Using 'localhost' as the hostname.")
		config.Hostname = "localhost"

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
		if !validateHostname(config.Hostname) || !validateHostname(config.MailgunHost) {
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
