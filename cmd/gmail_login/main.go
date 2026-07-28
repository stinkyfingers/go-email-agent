package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/stinkyfingers/go-email-agent/gmail"
	"github.com/stinkyfingers/go-email-agent/tokenstorage"
)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// TODO make storage configurable
	localTokenStore := tokenstorage.NewLocalTokenStore()

	gmailClient := gmail.NewGmail(localTokenStore)

	err = gmailClient.RunOAuthServer(ctx)
	if err != nil {
		log.Fatal(err)
	}

}
