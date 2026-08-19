package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
	"github.com/stinkyfingers/go-email-agent/agentfilestorage"
	"github.com/stinkyfingers/go-email-agent/llm"
	"github.com/stinkyfingers/go-email-agent/mcp"
	"github.com/stinkyfingers/go-email-agent/tokenstorage"
	"github.com/stinkyfingers/go-email-agent/user"
)

var messages = []anthropic.MessageParam{
	anthropic.NewUserMessage(anthropic.NewTextBlock(
		"Check my unread inbox. For every unread email that meets the criteria in your " +
			"instructions, draft a reply per those instructions. Skip anything that doesn't " +
			"meet the criteria — don't draft a reply for it. When you're done, summarize what " +
			"you drafted and what you skipped, and why.",
	)),
}

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


	// hex mcp server
	hexagonMCPSession, err := mcp.HexagonMCPSession(ctx, hexagonUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer hexagonMCPSession.Close()

	// run email agent for each user. Tally errors and log.
	var userErrors []string
	for _, user := range users {
		fmt.Println("running email agent for user", user.Email)
		// MCP servers
		mcpServer := mcp.NewMcpServer(user)
		mcpSession, err := mcpServer.Connect(ctx)
		if err != nil {
			userErrors = append(userErrors, fmt.Sprintf("connect error for user %s: %v", user.Email, err))
		}

		// Run LLM
		bedrock := llm.NewBedrock(awsRegion, mcpSession, hexagonMCPSession)
		anthropic := llm.NewAnthropic(bedrock, user, anthropicModelID)
		if err := anthropic.Run(ctx, messages); err != nil {
			userErrors = append(userErrors, fmt.Sprintf("anthropic error for user %s: %v", user.Email, err))
		}
	}
	// error output
	if len(userErrors) > 0 {
		log.Println("email agent errors: ", strings.Join(userErrors, "\n"))
	}
	fmt.Println("email agent run complete")
}
