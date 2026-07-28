package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stinkyfingers/go-email-agent/gmail"
)

type DraftEmailInput struct {
	MessageID string `json:"message_id" jsonschema:"the Email message ID of the unread email to reply to, as returned by check_email"`
	Body      string `json:"body" jsonschema:"the plain-text body of the reply to draft"`
}

type DraftEmailOutput struct {
	DraftID string `json:"draft_id" jsonschema:"the ID of the created Email draft"`
}

func (m *McpServer) DraftEmail(ctx context.Context, req *mcp.CallToolRequest, input DraftEmailInput) (
	*mcp.CallToolResult,
	DraftEmailOutput,
	error,
) {
	draftID, err := gmail.DraftEmail(ctx, input.MessageID, input.Body, m.User.TokenStorage, m.User.Email)
	if err != nil {
		return nil, DraftEmailOutput{}, err
	}
	return nil, DraftEmailOutput{DraftID: draftID}, nil
}
