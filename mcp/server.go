package mcp

import (
	"context"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stinkyfingers/go-email-agent/user"
)

/*
	Runs an "internal" MCP server to offer tools
*/

type McpServer struct {
	User         *user.User
	HexagonURL   string
	HexagonToken string
}

func NewMcpServer(user *user.User, hexagonURL, hexagonToken string) *McpServer {
	return &McpServer{
		User:         user,
		HexagonURL:   hexagonURL,
		HexagonToken: hexagonToken,
	}
}

// Run starts the MCP server on stdio, for running it as a standalone process.
func (m *McpServer) Run(ctx context.Context) error {
	return m.newServer().Run(ctx, &mcp.StdioTransport{})
}

// Connect starts the MCP server in-process and returns a client session wired
// to it over an in-memory transport, for callers that live in the same binary
// as the server (e.g. the Bedrock tool-use loop).
func (m *McpServer) Connect(ctx context.Context) (*mcp.ClientSession, error) {
	// nil check
	if m.User.TokenStorage == nil {
		return nil, fmt.Errorf("user has no tokenstorage configured")
	}
	if m.User.AgentFileStorage == nil {
		return nil, fmt.Errorf("user has no agentfilestorage configured")
	}

	server := m.newServer()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	// Servers must be connected before clients.
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		return nil, err
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "go-email-agent", Version: "v1.0.0"}, nil)
	return client.Connect(ctx, clientTransport, nil)
}

func (m *McpServer) newServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)

	// Define and register the tool.
	mcp.AddTool(server, &mcp.Tool{Name: "greet", Description: "say hi"}, m.SayHi)
	mcp.AddTool(server, &mcp.Tool{Name: "check_email", Description: "list unread emails in the inbox"}, m.CheckEmail)
	mcp.AddTool(server, &mcp.Tool{Name: "draft_email", Description: "create a Gmail draft replying to a specific unread email"}, m.DraftEmail)
	mcp.AddTool(server, &mcp.Tool{Name: "get_orders", Description: "list orders"}, m.GetOrders)
	mcp.AddTool(server, &mcp.Tool{Name: "get_order", Description: "get a specific order by external_order_id"}, m.GetOrder)
	mcp.AddTool(server, &mcp.Tool{Name: "get_failed_orders", Description: "list failed orders"}, m.GetFailedOrders)
	mcp.AddTool(server, &mcp.Tool{Name: "get_order_barcodes", Description: "get order barcodes by external_order_id"}, m.GetOrderBarcodes)
	mcp.AddTool(server, &mcp.Tool{Name: "get_order_event", Description: "get order event by external_order_id"}, m.GetOrderEvent)
	mcp.AddTool(server, &mcp.Tool{Name: "get_order_double_sale_check", Description: "check for double sales of an order by external_order_id"}, m.GetOrderDoubleSaleCheck)

	return server
}
