package main

import (
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

var (
	ctx = context.Background()
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
)

func ensureSetup() {
	// Check if the configuration exists in Redis
	_, err := rdb.Get(ctx, "config").Result()
	if err != nil {
		fmt.Println("Please configure Quacker first using ./quacker --setup")
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--setup":
			setup()
		case "--run":
			runServer()
		case "--job":
			job()
		case "--allow":
			if len(os.Args) < 3 {
				fmt.Println("Usage: quacker --allow <GitHubUsername>")
				os.Exit(1)
			}
			allowGitHubUser(os.Args[2])
		case "--remove":
			if len(os.Args) < 3 {
				fmt.Println("Usage: quacker --remove <GitHubUsername>")
				os.Exit(1)
			}
			removeGitHubUser(os.Args[2])
		default:
			fmt.Println("Invalid argument")
		}
	} else {
		runServer()
	}
}
