package tools

import (
	"context"
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/toolrunner"
	"github.com/stinkyfingers/go-email-agent/gmail"
)

type CheckEmailInput struct{}

type CheckEmailOutput struct {
	Count  int                  `json:"count" jsonschema:"the number of unread emails found"`
	Emails []gmail.EmailSummary `json:"emails" jsonschema:"the unread emails, each with id, from, and subject"`
}

func (t *ToolHandler) CheckEmailTool() (anthropic.BetaTool, error) {
	return toolrunner.NewBetaToolFromJSONSchema(
		"check_email",
		"list unread emails in the inbox",
		func(ctx context.Context, input CheckEmailInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			emails, err := gmail.CheckGmail(ctx, t.User.TokenStorage, t.User.Email)
			if err != nil {
				return anthropic.BetaToolResultBlockParamContentUnion{}, err
			}
			b, err := json.Marshal(CheckEmailOutput{Count: len(emails), Emails: emails})
			if err != nil {
				return anthropic.BetaToolResultBlockParamContentUnion{}, err
			}
			return anthropic.BetaToolResultBlockParamContentUnion{
				OfText: &anthropic.BetaTextBlockParam{
					Text: string(b),
				},
			}, nil
		},
	)
}
