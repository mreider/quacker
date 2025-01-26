// allow_helper.go

package main

import (
	"fmt"
	"os"
	"strings"
)

// Allow a GitHub user by storing their username in Redis
func allowGitHubUser(username string) {
	ensureSetup()

	err := rdb.Set(ctx, "github_user:"+username, "allowed", 0).Err()
	if err != nil {
		fmt.Println("Failed to store GitHub username in Redis")
		os.Exit(1)
	}

	fmt.Printf("Allowed GitHub user: %s\n", username)
}

// Remove a GitHub user and all associated data
func removeGitHubUser(username string) {
	ensureSetup()

	// Remove GitHub user from Redis
	err := rdb.Del(ctx, "github_user:"+username).Err()
	if err != nil {
		fmt.Println("Failed to remove GitHub username from Redis")
		os.Exit(1)
	}

	// Retrieve and delete all sites associated with the user
	sites, _ := rdb.Keys(ctx, "user_sites:"+username+":*").Result()
	for _, siteKey := range sites {
		// Remove subscribers for the site
		subscribersKey := "subs:" + strings.TrimPrefix(siteKey, "user_sites:")
		if err := rdb.Del(ctx, subscribersKey).Err(); err != nil {
			fmt.Printf("Failed to remove subscribers for site: %s\n", siteKey)
		}
		if err := rdb.Del(ctx, siteKey).Err(); err != nil {
			fmt.Printf("Failed to remove site: %s\n", siteKey)
		}
	}

	fmt.Printf("Removed GitHub user and associated data: %s\n", username)
}
