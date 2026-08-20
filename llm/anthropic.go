package llm

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/bedrock"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stinkyfingers/go-email-agent/notification"
	"github.com/stinkyfingers/go-email-agent/tools"
	"github.com/stinkyfingers/go-email-agent/user"
)

type Anthropic struct {
	session     *mcpsdk.ClientSession
	user        *user.User
	modelID     string
	region      string
	toolHandler *tools.ToolHandler
}

func NewAnthropic(user *user.User, modelID, region string, toolHandler *tools.ToolHandler) *Anthropic {
	return &Anthropic{
		user:        user,
		modelID:     modelID,
		region:      region,
		toolHandler: toolHandler,
	}
}

func (a *Anthropic) Run(ctx context.Context, initialMessage string) error {
	if a.user.AgentFileStorage == nil {
		return fmt.Errorf("user has no agentfilestorage configured")
	}

	client := anthropic.NewClient(
		bedrock.WithLoadDefaultConfig(ctx, awsconfig.WithRegion(a.region)),
	)

	// load policy file
	// policyFile (your triage/drafting instructions) is read from
	// source and specified as a cached system block, so the same content isn't
	// re-priced on every turn of the tool-use loop below.
	policyFile, err := a.user.AgentFileStorage.RetrieveFile(a.user.Email)
	if err != nil {
		return err
	}
	system := []anthropic.BetaTextBlockParam{{
		Text: string(policyFile),
		CacheControl: anthropic.BetaCacheControlEphemeralParam{
			TTL: "5m",
		},
	}}

	// load tools
	tools, err := a.tools()
	if err != nil {
		return fmt.Errorf("listing mcp tools: %w", err)
	}

	runner := client.Beta.Messages.NewToolRunner(tools, anthropic.BetaToolRunnerParams{
		BetaMessageNewParams: anthropic.BetaMessageNewParams{
			Model:     a.modelID,
			MaxTokens: 1000,
			Messages: []anthropic.BetaMessageParam{
				anthropic.NewBetaUserMessage(
					anthropic.NewBetaTextBlock(initialMessage),
				),
			},
			System: system,
		},

		MaxIterations: 5,
	})
	finalMessage, err := runner.RunToCompletion(ctx)
	if err != nil {
		return err
	}
	// stdout print details
	for _, msg := range finalMessage.Content {
		if msg.Type == "text" {
			fmt.Println(msg.Text)
		}
	}
	// slack notification if drafts were created
	if n := a.toolHandler.EmailDraftsCreated.Load(); n > 0 {
		msg := fmt.Sprintf("📬 go-email-agent created %d draft(s). Check Gmail Drafts to review.", n)
		if err := notification.NotifySlack(ctx, msg); err != nil {
			log.Printf("slack notification failed: %v", err)
		}
	}
	return nil
}

func (a *Anthropic) tools() ([]anthropic.BetaTool, error) {
	sayHiTool, err := a.toolHandler.SayHiTool()
	if err != nil {
		return nil, err
	}
	checkEmailTool, err := a.toolHandler.CheckEmailTool()
	if err != nil {
		return nil, err
	}
	draftEmailTool, err := a.toolHandler.DraftEmailTool()
	if err != nil {
		return nil, err
	}
	ordersTool, err := a.toolHandler.GetOrdersTool()
	if err != nil {
		return nil, err
	}
	orderTool, err := a.toolHandler.GetOrderTool()
	if err != nil {
		return nil, err
	}
	failedOrdersTool, err := a.toolHandler.GetFailedOrdersTool()
	if err != nil {
		return nil, err
	}
	orderBarcodesTool, err := a.toolHandler.GetOrderBarcodesTool()
	if err != nil {
		return nil, err
	}
	orderEventTool, err := a.toolHandler.GetOrderEventTool()
	if err != nil {
		return nil, err
	}
	orderDoubleSaleCheckTool, err := a.toolHandler.GetOrderDoubleSaleCheckTool()
	if err != nil {
		return nil, err
	}

	tools := []anthropic.BetaTool{
		sayHiTool,
		checkEmailTool,
		draftEmailTool,
		ordersTool,
		orderTool,
		failedOrdersTool,
		orderBarcodesTool,
		orderEventTool,
		orderDoubleSaleCheckTool,
	}
	return tools, nil
}
