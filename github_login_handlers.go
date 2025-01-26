// github_login_handlers.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	// Redirect to GitHub OAuth page
	clientID := rdb.Get(ctx, "config:github_client_id").Val()
	redirectURI := rdb.Get(ctx, "config:github_redirect_uri").Val()
	if clientID == "" || redirectURI == "" {
		renderErrorPage(w, r, "GitHub OAuth is not configured.")
		return
	}

	oauthURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=read:user",
		clientID,
		redirectURI,
	)
	http.Redirect(w, r, oauthURL, http.StatusFound)
}

func handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		renderErrorPage(w, r, "GitHub login failed. Please try again.")
		return
	}

	// Exchange code for access token
	accessToken, err := exchangeGitHubCodeForToken(code)
	if err != nil {
		renderErrorPage(w, r, "GitHub login failed during token exchange.")
		return
	}

	// Fetch user info from GitHub API
	username, err := fetchGitHubUsername(accessToken)
	if err != nil {
		renderErrorPage(w, r, "Failed to retrieve GitHub user information.")
		return
	}

	// Check if the user is allowed in Redis
	if rdb.Get(ctx, "github_user:"+username).Err() != nil {
		renderErrorPage(w, r, "Access denied. This Quacker instance is restricted to pre-approved GitHub users.")
		return
	}

	// Successful login, store user in session or indicate login success
	setSession(w, username)
	http.Redirect(w, r, "/", http.StatusFound)
}

func exchangeGitHubCodeForToken(code string) (string, error) {
	clientID := rdb.Get(ctx, "config:github_client_id").Val()
	clientSecret := rdb.Get(ctx, "config:github_client_secret").Val()
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("GitHub OAuth credentials are not configured")
	}

	url := "https://github.com/login/oauth/access_token"
	payload := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
	}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return "", err
	}

	accessToken, ok := responseData["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("Failed to retrieve access token")
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

	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return "", err
	}

	username, ok := responseData["login"].(string)
	if !ok {
		return "", fmt.Errorf("Failed to retrieve GitHub username")
	}

	return username, nil
}

func setSession(w http.ResponseWriter, username string) {
	// Implement session handling logic
	http.SetCookie(w, &http.Cookie{
		Name:  "quacker_user",
		Value: username,
		Path:  "/",
	})
}
