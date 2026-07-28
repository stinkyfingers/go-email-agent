package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stinkyfingers/go-email-agent/gmail"
)

type CheckEmailInput struct{}

type CheckEmailOutput struct {
	Count  int                  `json:"count" jsonschema:"the number of unread emails found"`
	Emails []gmail.EmailSummary `json:"emails" jsonschema:"the unread emails, each with id, from, and subject"`
}

func (m *McpServer) CheckEmail(ctx context.Context, req *mcp.CallToolRequest, input CheckEmailInput) (
	*mcp.CallToolResult,
	CheckEmailOutput,
	error,
) {
	emails, err := gmail.CheckGmail(ctx, m.User.TokenStorage, m.User.Email)
	if err != nil {
		return nil, CheckEmailOutput{}, err
	}

	return nil, CheckEmailOutput{Count: len(emails), Emails: emails}, nil
}
