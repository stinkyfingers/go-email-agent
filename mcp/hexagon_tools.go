package mcp

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type OrderInput struct {
	ExternalOrderID string `json:"external_order_id"`
}

type HexagonOutput struct {
	Data any
}

// GetOrders calls hex endpoint: /api/agent/v1/orders
func (m *McpServer) GetOrders(ctx context.Context, req *mcp.CallToolRequest, input OrderInput) (
	*mcp.CallToolResult,
	HexagonOutput,
	error,
) {
	data, err := m.HexagonApiGet("orders")
	if err != nil {
		return nil, HexagonOutput{}, err
	}
	return nil, HexagonOutput{Data: data}, nil
}

// GetOrder calls hex endpoint: /api/agent/v1/orders/:external_order_id
func (m *McpServer) GetOrder(ctx context.Context, req *mcp.CallToolRequest, input OrderInput) (
	*mcp.CallToolResult,
	HexagonOutput,
	error,
) {
	data, err := m.HexagonApiGet(fmt.Sprintf("orders/%s", input.ExternalOrderID))
	if err != nil {
		return nil, HexagonOutput{}, err
	}
	return nil, HexagonOutput{Data: data}, nil
}

// GetFailedOrders calls hex endpoint: /api/agent/v1/orders/failed
func (m *McpServer) GetFailedOrders(ctx context.Context, req *mcp.CallToolRequest, input OrderInput) (
	*mcp.CallToolResult,
	HexagonOutput,
	error,
) {
	// var orders HexagonOutput
	data, err := m.HexagonApiGet("orders/failed")
	if err != nil {
		return nil, HexagonOutput{}, err
	}
	return nil, HexagonOutput{Data: data}, nil
}

// GetOrderBarcodes calls hex endpoint: /api/agent/v1/orders/:external_order_id/barcodes
func (m *McpServer) GetOrderBarcodes(ctx context.Context, req *mcp.CallToolRequest, input OrderInput) (
	*mcp.CallToolResult,
	HexagonOutput,
	error,
) {
	data, err := m.HexagonApiGet(fmt.Sprintf("orders/%s/barcodes", input.ExternalOrderID))
	if err != nil {
		return nil, HexagonOutput{}, err
	}
	return nil, HexagonOutput{Data: data}, nil
}

// GetOrderEvent calls hex endpoint: /api/agent/v1/orders/:external_order_id/event
func (m *McpServer) GetOrderEvent(ctx context.Context, req *mcp.CallToolRequest, input OrderInput) (
	*mcp.CallToolResult,
	HexagonOutput,
	error,
) {
	data, err := m.HexagonApiGet(fmt.Sprintf("orders/%s/event", input.ExternalOrderID))
	if err != nil {
		return nil, HexagonOutput{}, err
	}
	return nil, HexagonOutput{Data: data}, nil
}

// GetOrderDoubleSaleCheck calls hex endpoint: /api/agent/v1/orders/:external_order_id/double_check_sale
func (m *McpServer) GetOrderDoubleSaleCheck(ctx context.Context, req *mcp.CallToolRequest, input OrderInput) (
	*mcp.CallToolResult,
	HexagonOutput,
	error,
) {
	data, err := m.HexagonApiGet(fmt.Sprintf("orders/%s/double_sale_check", input.ExternalOrderID))
	if err != nil {
		return nil, HexagonOutput{}, err
	}
	return nil, HexagonOutput{Data: data}, nil
}

/* utils */

// HexagonApiGet makes a GET call to the Hexagon API
func (m *McpServer) HexagonApiGet(endpoint string) (string, error) {
	requestURL := fmt.Sprintf("%s/%s", m.HexagonURL, endpoint)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", m.HexagonToken))
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
