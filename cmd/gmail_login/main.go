package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/stinkyfingers/go-email-agent/gmail"
	"github.com/stinkyfingers/go-email-agent/tokenstorage"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("error loading .env file: %v", err)
	}
	// envvars
	ssmPrefix := os.Getenv("SSM_PREFIX")
	if ssmPrefix == "" {
		log.Fatal("SSM_PREFIX is empty")
	}
	// storage type
	var tokenStore tokenstorage.Storage
	tokenStore, err := tokenstorage.NewSSMTokenStore(ctx, ssmPrefix)
	if err != nil {
		log.Fatal(err)
	}
	if os.Getenv("STORAGE") == "local" {
		tokenStore = tokenstorage.NewLocalTokenStore()
	}
	// gmail client
	gmailClient, err := gmail.NewGmail(tokenStore)
	if err != nil {
		log.Fatal(err)
	}
	// run http server
	if err := gmailClient.RunOAuthServer(ctx); err != nil {
		log.Fatal(err)
	}

}
