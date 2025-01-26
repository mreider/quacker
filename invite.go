// invite.go

package main

import (
	"fmt"
	"math/rand"
	"time"
	"os"
)

func generateInvite() {
	// Ensure setup is complete before generating invite codes
	ensureSetup()

	// Generate a random invitation code
	code := generateInviteCode()

	// Save the invite code to Redis with a 48-hour expiration
	err := rdb.Set(ctx, "code:"+code, "valid", 48*time.Hour).Err()
	if err != nil {
		fmt.Println("Failed to store invite code in Redis")
		os.Exit(1)
	}

	fmt.Printf("Generated invite code: %s\n", code)
}

func generateInviteCode() string {
	// Generate a random 12-character string consisting of letters and numbers
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	code := make([]rune, 12)
	rand.Seed(time.Now().UnixNano())
	for i := range code {
		code[i] = letters[rand.Intn(len(letters))]
	}
	return string(code)
}
