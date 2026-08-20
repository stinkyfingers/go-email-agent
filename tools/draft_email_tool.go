package tools

import (
	"context"
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/toolrunner"
	"github.com/stinkyfingers/go-email-agent/gmail"
)

type DraftEmailInput struct {
	MessageID string `json:"message_id" jsonschema:"the Email message ID of the unread email to reply to, as returned by check_email"`
	Body      string `json:"body" jsonschema:"the plain-text body of the reply to draft"`
}

type DraftEmailOutput struct {
	DraftID string `json:"draft_id" jsonschema:"the ID of the created Email draft"`
}

func (t *ToolHandler) DraftEmailTool() (anthropic.BetaTool, error) {
	return toolrunner.NewBetaToolFromJSONSchema(
		"draft_email",
		"create a Gmail draft replying to a specific unread email",
		func(ctx context.Context, input DraftEmailInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			draftID, err := gmail.DraftEmail(ctx, input.MessageID, input.Body, t.User.TokenStorage, t.User.Email)
			if err != nil {
				return anthropic.BetaToolResultBlockParamContentUnion{}, err
			}
			b, err := json.Marshal(DraftEmailOutput{DraftID: draftID})
			if err != nil {
				return anthropic.BetaToolResultBlockParamContentUnion{}, err
			}
			t.EmailDraftsCreated.Add(1)
			return anthropic.BetaToolResultBlockParamContentUnion{
				OfText: &anthropic.BetaTextBlockParam{
					Text: string(b),
				},
			}, nil
		},
	)
}
