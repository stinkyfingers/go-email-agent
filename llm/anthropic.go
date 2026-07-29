package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/bedrock"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stinkyfingers/go-email-agent/user"
)

type Anthropic struct {
	bedrockClient  *Bedrock
	toolSessionMap map[string]*mcpsdk.ClientSession // tool name: session
	user           *user.User
	modelID        string
}

func NewAnthropic(bedrockClient *Bedrock, user *user.User, modelID string) *Anthropic {
	return &Anthropic{
		bedrockClient:  bedrockClient,
		toolSessionMap: make(map[string]*mcpsdk.ClientSession),
		user:           user,
		modelID:        modelID,
	}
}

// Run sends a message to Claude on Bedrock and drives the tool-use loop,
// dispatching any tool calls to the connected MCP server.
// If an email draft was generated, Slack is notified.
func (a *Anthropic) Run(ctx context.Context, messages []anthropic.MessageParam) error {
	if a.user.AgentFileStorage == nil {
		return fmt.Errorf("user has no agentfilestorage configured")
	}

	client := anthropic.NewClient(
		bedrock.WithLoadDefaultConfig(ctx, awsconfig.WithRegion(a.bedrockClient.Region)),
	)
	// load tools
	tools, err := a.mcpTools(ctx)
	if err != nil {
		return fmt.Errorf("listing mcp tools: %w", err)
	}

	// load policy file
	// policyFile (your triage/drafting instructions) is read from
	// source and specified as a cached system block, so the same content isn't
	// re-priced on every turn of the tool-use loop below.
	policyFile, err := a.user.AgentFileStorage.RetrieveFile(a.user.Email)
	if err != nil {
		return err
	}
	system := []anthropic.TextBlockParam{{
		Text:         string(policyFile),
		CacheControl: anthropic.NewCacheControlEphemeralParam(),
	}}

	var draftsCreated int

	for {
		resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     a.modelID,
			MaxTokens: 1024,
			System:    system,
			Messages:  messages,
			Tools:     tools,
		})
		if err != nil {
			return fmt.Errorf("calling bedrock: %w", err)
		}
		messages = append(messages, resp.ToParam())

		var toolResults []anthropic.ContentBlockParamUnion
		for _, block := range resp.Content {
			switch variant := block.AsAny().(type) {
			case anthropic.TextBlock:
				log.Println(variant.Text)
			case anthropic.ToolUseBlock:
				result, isError := a.callMCPTool(ctx, variant)
				if variant.Name == "draft_email" && !isError {
					draftsCreated++
				}
				toolResults = append(toolResults, anthropic.NewToolResultBlock(variant.ID, result, isError))
			}
		}

		if resp.StopReason != anthropic.StopReasonToolUse {
			if draftsCreated > 0 {
				msg := fmt.Sprintf("📬 go-email-agent created %d draft(s). Check Gmail Drafts to review.", draftsCreated)
				if err := notifySlack(ctx, msg); err != nil {
					log.Printf("slack notification failed: %v", err)
				}
			}
			return nil
		}
		messages = append(messages, anthropic.NewUserMessage(toolResults...))
	}
}

// mcpTools discovers the tools exposed by the connected MCP server and
// converts each into an Anthropic tool definition.
func (a *Anthropic) mcpTools(ctx context.Context) ([]anthropic.ToolUnionParam, error) {
	var tools []anthropic.ToolUnionParam

	for _, mcpSession := range a.bedrockClient.MCPSessions {
		list, err := mcpSession.ListTools(ctx, nil)
		if err != nil {
			return nil, err
		}
		for _, t := range list.Tools {
			fmt.Println("enabling tool", t.Name, t.Description)
			schemaBytes, err := json.Marshal(t.InputSchema)
			if err != nil {
				return nil, fmt.Errorf("marshaling schema for tool %q: %w", t.Name, err)
			}
			var inputSchema anthropic.ToolInputSchemaParam
			if err := json.Unmarshal(schemaBytes, &inputSchema); err != nil {
				return nil, fmt.Errorf("parsing schema for tool %q: %w", t.Name, err)
			}

			tools = append(tools, anthropic.ToolUnionParam{
				OfTool: &anthropic.ToolParam{
					Name:        t.Name,
					Description: anthropic.String(t.Description),
					InputSchema: inputSchema,
				},
			})
			a.toolSessionMap[t.Name] = mcpSession
		}
	}
	return tools, nil
}

// callMCPTool executes a Claude tool_use request against the connected MCP
// server and returns the result text (and whether it was an error) for use
// in a tool_result block.
func (a *Anthropic) callMCPTool(ctx context.Context, toolUse anthropic.ToolUseBlock) (result string, isError bool) {
	var args map[string]any
	if err := json.Unmarshal(toolUse.Input, &args); err != nil {
		return fmt.Sprintf("invalid tool input: %v", err), true
	}

	mcpSession, ok := a.toolSessionMap[toolUse.Name]
	if !ok {
		return fmt.Sprintf("tool %s not found", toolUse.Name), true
	}

	res, err := mcpSession.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      toolUse.Name,
		Arguments: args,
	})
	if err != nil {
		return fmt.Sprintf("mcp tool call failed: %v", err), true
	}

	var text strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcpsdk.TextContent); ok {
			text.WriteString(tc.Text)
		}
	}
	return text.String(), res.IsError
}
