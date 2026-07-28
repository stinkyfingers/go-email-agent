package main

import (
	"context"
	"fmt"
	"log"
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
	// anthropic.NewUserMessage(anthropic.NewTextBlock(
	// 	"Check my unread inbox. For every unread email that meets the criteria in your " +
	// 		"instructions, draft a reply per those instructions. Skip anything that doesn't " +
	// 		"meet the criteria — don't draft a reply for it. When you're done, summarize what " +
	// 		"you drafted and what you skipped, and why.",
	// )),
	anthropic.NewUserMessage(anthropic.NewTextBlock(
		"Say hello to anyone needing a greeting. List first 3 tables. Respond to unread emails that meet criteria", // TODO
	)),
}

func main() {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// TODO: make user and storage configurable
	localTokenStore := tokenstorage.NewLocalTokenStore()
	localAgentFileStorage := agentfilestorage.NewLocalAgentFileStorage()

	users, err := user.ListUsers(localTokenStore, localAgentFileStorage)
	if err != nil {
		log.Fatal(err)
	}

	// run email agent for each user. Tally errors and log.
	var userErrors []string
	for _, user := range users {
		mcpServer := mcp.NewMcpServer(user)
		mcpSession, err := mcpServer.Connect(ctx)
		if err != nil {
			userErrors = append(userErrors, fmt.Sprintf("connect error for user %s: %v", user.Email, err))
		}
		hexagonMCPSession, err := mcp.HexagonMCPSession(ctx)
		if err != nil {
			userErrors = append(userErrors, fmt.Sprintf("hexagon mcp session error for user %s: %v", user.Email, err))
		}

		bedrock := llm.NewBedrock("us-west-1", mcpSession, hexagonMCPSession)
		anthropic := llm.NewAnthropic(bedrock, user)
		if err := anthropic.Run(ctx, messages); err != nil {
			userErrors = append(userErrors, fmt.Sprintf("anthropic error for user %s: %v", user.Email, err))
		}
	}
	if len(userErrors) > 0 {
		log.Println("email agent errors: ", strings.Join(userErrors, "\n"))
	}
}
