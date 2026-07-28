package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const hexagonUrl = "http://localhost:3001" // TODO - make configurable

// HexagonMCPSession returns a session. It's a good idea to call session.Close() when complete.
func HexagonMCPSession(ctx context.Context) (*mcp.ClientSession, error) {
	// The real MCP server is the Ruby/Rails app in ticketco-web, which mounts
	// MCP::Server::Transports::StreamableHTTPTransport at "/mcp" (see
	// ticketco-web/config/routes.rb).
	client := mcp.NewClient(&mcp.Implementation{Name: "mcp-client", Version: "v1.0.0"}, nil)
	transport := &mcp.StreamableClientTransport{Endpoint: fmt.Sprintf("%s/mcp", hexagonUrl)}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("connecting to MCP server: %w", err)
	}
	return session, nil
}
