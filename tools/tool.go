package tools

import (
	"sync/atomic"

	"github.com/stinkyfingers/go-email-agent/user"
)

type ToolHandler struct {
	User               *user.User
	HexagonURL         string
	HexagonToken       string
	EmailDraftsCreated atomic.Int32
}

func NewToolHandler(user *user.User, hexagonURL, hexagonToken string) *ToolHandler {
	return &ToolHandler{
		User:               user,
		HexagonURL:         hexagonURL,
		HexagonToken:       hexagonToken,
		EmailDraftsCreated: atomic.Int32{}, // explicitly init to zero
	}
}
