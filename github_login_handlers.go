package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type OAuthConfig struct {
	MailgunAPIKey      string `json:"MailgunAPIKey"`
	MailgunHost        string `json:"MailgunHost"`
	Hostname           string `json:"Hostname"`
	GitHubClientID     string `json:"GitHubClientID"`
	GitHubClientSecret string `json:"GitHubClientSecret"`
	GitHubRedirectURI  string `json:"GitHubRedirectURI"`
}

func getConfig() (*OAuthConfig, error) {
	configJSON, err := rdb.Get(ctx, "config").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve config from Redis: %w", err)
	}

	var config OAuthConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
}

func handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	config, err := getConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error retrieving config:", err)
		renderErrorPage(w, r, "GitHub OAuth is not configured.")
		return
	}

	oauthURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=read:user",
		config.GitHubClientID,
		config.GitHubRedirectURI,
	)
	http.Redirect(w, r, oauthURL, http.StatusFound)
}

func handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		fmt.Fprintln(os.Stderr, "GitHub callback missing code parameter")
		renderErrorPage(w, r, "GitHub login failed. Please try again.")
		return
	}

	fmt.Fprintln(os.Stderr, "Received code:", code) // Debug log

	accessToken, err := exchangeGitHubCodeForToken(code)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Token exchange error:", err) // Debug log
		renderErrorPage(w, r, "GitHub login failed during token exchange.")
		return
	}

	username, err := fetchGitHubUsername(accessToken)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Fetch username error:", err) // Debug log
		renderErrorPage(w, r, "Failed to retrieve GitHub user information.")
		return
	}

	if rdb.Get(ctx, "github_user:"+username).Err() != nil {
		fmt.Fprintln(os.Stderr, "Access denied for user:", username) // Debug log
		renderErrorPage(w, r, "Access denied. This Quacker instance is restricted to pre-approved GitHub users.")
		return
	}

	setSession(w, username)
	http.Redirect(w, r, "/", http.StatusFound)
}

func exchangeGitHubCodeForToken(code string) (string, error) {
	config, err := getConfig()
	if err != nil {
		return "", fmt.Errorf("GitHub OAuth credentials are not configured")
	}

	url := "https://github.com/login/oauth/access_token"
	payload := map[string]string{
		"client_id":     config.GitHubClientID,
		"client_secret": config.GitHubClientSecret,
		"code":          code,
		"redirect_uri":  config.GitHubRedirectURI,
	}
	jsonPayload, _ := json.Marshal(payload)

	fmt.Fprintln(os.Stderr, "Sending token exchange request with payload:", string(jsonPayload)) // Debug log

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error during token exchange: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token exchange response: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Token exchange response:", string(body)) // Debug log

	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return "", fmt.Errorf("failed to parse token exchange response: %w", err)
	}

	if errorMsg, exists := responseData["error"]; exists {
		fmt.Fprintln(os.Stderr, "GitHub token exchange error:", errorMsg) // Debug log
		return "", fmt.Errorf("GitHub token exchange error: %s", errorMsg)
	}

	accessToken, ok := responseData["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access token not found in response")
	}

	return accessToken, nil
}

func fetchGitHubUsername(token string) (string, error) {
	url := "https://api.github.com/user"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Fprintln(os.Stderr, "Fetch username response:", string(body)) // Debug log

	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return "", err
	}

	username, ok := responseData["login"].(string)
	if !ok {
		return "", fmt.Errorf("failed to retrieve GitHub username")
	}

	return username, nil
}

func setSession(w http.ResponseWriter, username string) {
	http.SetCookie(w, &http.Cookie{
		Name:  "quacker_user",
		Value: username,
		Path:  "/",
	})
}
