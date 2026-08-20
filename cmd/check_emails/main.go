package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/stinkyfingers/go-email-agent/agentfilestorage"
	"github.com/stinkyfingers/go-email-agent/llm"
	"github.com/stinkyfingers/go-email-agent/tokenstorage"
	"github.com/stinkyfingers/go-email-agent/tools"
	"github.com/stinkyfingers/go-email-agent/user"
)

var initialMessage = "Check my unread inbox. For every unread email that meets the criteria in your " +
	"instructions, draft a reply per those instructions. Skip anything that doesn't " +
	"meet the criteria — don't draft a reply for it. When you're done, summarize what " +
	"you drafted and what you skipped, and why."

func main() {
	ctx := context.Background()

	// envvars
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("error loading .env file: %v", err)
	}
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		log.Fatal("AWS_REGION is empty")
	}
	hexagonUrl := os.Getenv("HEXAGON_URL")
	if hexagonUrl == "" {
		log.Fatal("HEXAGON_URL is empty")
	}
	hexagonToken := os.Getenv("HEXAGON_BEARER_TOKEN")
	if hexagonToken == "" {
		log.Fatal("HEXAGON_BEARER_TOKEN is empty")
	}
	anthropicModelID := os.Getenv("ANTHROPIC_MODEL_ID")
	if anthropicModelID == "" {
		log.Fatal("ANTHROPIC_MODEL_ID is empty")
	}
	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		log.Fatal("S3_BUCKET is empty")
	}
	ssmPrefix := os.Getenv("SSM_PREFIX")
	if ssmPrefix == "" {
		log.Fatal("SSM_PREFIX is empty")
	}

	// establish storage
	var tokenStore tokenstorage.Storage
	var agentFileStore agentfilestorage.Storage
	tokenStore, err := tokenstorage.NewSSMTokenStore(ctx, ssmPrefix)
	if err != nil {
		log.Fatal(err)
	}
	agentFileStore, err = agentfilestorage.NewS3AgentFileStorage(ctx, s3Bucket)
	if err != nil {
		log.Fatal(err)
	}

	if os.Getenv("STORAGE") == "local" {
		tokenStore = tokenstorage.NewLocalTokenStore()
		agentFileStore = agentfilestorage.NewLocalAgentFileStorage()
	}

	// list users to iterate over
	users, err := user.ListUsers(tokenStore, agentFileStore)
	if err != nil {
		log.Fatal(err)
	}

	// run email agent for each user. Tally errors and log.
	var userErrors []string
	for _, user := range users {
		fmt.Println("running email agent for user", user.Email)
		toolHandler := tools.NewToolHandler(user, hexagonUrl, hexagonToken)
		anthropic := llm.NewAnthropic(user, anthropicModelID, awsRegion, toolHandler)
		if err := anthropic.Run(ctx, initialMessage); err != nil {
			userErrors = append(userErrors, fmt.Sprintf("anthropic error for user %s: %v", user.Email, err))
		}
	}
	// error output
	if len(userErrors) > 0 {
		log.Println("email agent errors: ", strings.Join(userErrors, "\n"))
	}
	fmt.Println("email agent run complete")
}
