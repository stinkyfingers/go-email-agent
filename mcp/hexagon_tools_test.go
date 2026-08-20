package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHexagonTools(t *testing.T) {
	const wantAuth = "Bearer test-token"
	m := &McpServer{
		HexagonToken: "test-token",
	}

	tests := []struct {
		endpoint string
		method   func(
			ctx context.Context,
			req *mcp.CallToolRequest,
			input OrderInput,
		) (
			*mcp.CallToolResult,
			HexagonOutput,
			error,
		)
		orderInput OrderInput
	}{
		{
			endpoint:   "/orders",
			method:     m.GetOrders,
			orderInput: OrderInput{},
		},
		{
			endpoint:   "/orders/123",
			method:     m.GetOrder,
			orderInput: OrderInput{ExternalOrderID: "123"},
		},
		{
			endpoint:   "/orders/failed",
			method:     m.GetFailedOrders,
			orderInput: OrderInput{},
		},
		{
			endpoint:   "/orders/123/barcodes",
			method:     m.GetOrderBarcodes,
			orderInput: OrderInput{ExternalOrderID: "123"},
		},
		{
			endpoint:   "/orders/123/event",
			method:     m.GetOrderEvent,
			orderInput: OrderInput{ExternalOrderID: "123"},
		},
		{
			endpoint:   "/orders/123/double_sale_check",
			method:     m.GetOrderDoubleSaleCheck,
			orderInput: OrderInput{ExternalOrderID: "123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.endpoint, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.endpoint {
					t.Errorf("unexpected request path: %s, expected %s", r.URL.Path, tt.endpoint)
				}
				if got := r.Header.Get("Authorization"); got != wantAuth {
					t.Errorf("Authorization header = %q, want %q", got, wantAuth)
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("test response")) // live response will be JSON, but can be any data type.
			}))
			defer server.Close()
			m.HexagonURL = server.URL // set url

			_, out, err := tt.method(context.Background(), nil, tt.orderInput)
			if err != nil {
				t.Fatalf("%s returned error: %v", tt.endpoint, err)
			}
			if out.Data != "test response" {
				t.Fatalf("expected 'test response', got %s", out.Data)
			}
		})
	}
}

func TestHexagonLive(t *testing.T) {
	t.Skip("un-skip to use for testing against local hexagon API during development")
	const wantAuth = "Bearer test-token"

	m := &McpServer{
		HexagonURL:   "http://localhost:3001/api/agent/v1",
		HexagonToken: "abcdef",
	}
	tests := []struct {
		description string
		method      func(
			ctx context.Context,
			req *mcp.CallToolRequest,
			input OrderInput,
		) (
			*mcp.CallToolResult,
			HexagonOutput,
			error,
		)
		orderInput     OrderInput
		expectedOutput string
	}{
		{
			description:    "failed orders",
			method:         m.GetFailedOrders,
			orderInput:     OrderInput{},
			expectedOutput: `{"orders":[{"id":55501,"external_order_id":"678-90123456-11111","market":"stubhub","status":"finalizing_stalled","quantity":2,"total_price_cents":18000,"customer_email":"buyer@example.com","team_name":"Los Angeles Dodgers","venue_name":"Dodger Stadium","created_at":"2026-08-10T18:00:00Z"},{"id":55502,"external_order_id":"901-23456789-22222","market":"seatgeek","status":"error","quantity":4,"total_price_cents":41200,"customer_email":"someone-else@example.com","team_name":"San Diego Padres","venue_name":"Petco Park","created_at":"2026-08-11T09:12:00Z"}],"count":2}`,
		},
		{
			description:    "get orders",
			method:         m.GetOrders,
			orderInput:     OrderInput{},
			expectedOutput: `{"orders":[{"id":12345,"external_order_id":"678-90123456-45678","market":"stubhub","status":"fulfilled","quantity":2,"total_price_cents":24500,"created_at":"2026-08-10T18:00:00Z","fulfilled_at":"2026-08-10T19:15:00Z"},{"id":12346,"external_order_id":"901-23456789-01234","market":"seatgeek","status":"fulfilled","quantity":4,"total_price_cents":51000,"created_at":"2026-08-11T11:04:00Z","fulfilled_at":"2026-08-11T11:41:00Z"},{"id":12347,"external_order_id":"678-90123456-77777","market":"stubhub","status":"confirmed","quantity":1,"total_price_cents":9800,"created_at":"2026-08-12T15:22:00Z","fulfilled_at":null}],"count":3}`,
		},
		{
			description: "get order",
			method:      m.GetOrder,
			orderInput: OrderInput{
				ExternalOrderID: "12346",
			},
			expectedOutput: `{"id":12345,"external_order_id":"12346","market":"stubhub","status":"fulfilled","quantity":2,"total_price_cents":24500,"subtotal_price_cents":22000,"remittance_amount_cents":21120,"customer_email":"buyer@example.com","customer_first_name":"Alex","customer_last_name":"Rivera","confirmed_at":"2026-08-10T18:30:00Z","fulfilled_at":"2026-08-10T19:15:00Z","rejected_at":null,"refunded_at":null,"refund_amount_cents":null,"partial_refund":false,"delay_fill":false,"created_at":"2026-08-10T18:00:00Z","updated_at":"2026-08-10T19:15:00Z","raw_response":{"id":"678-90123456-45678","eventId":"154820391","eventDescription":"Padres at Dodgers","eventDateLocal":"2026-08-15T12:10:00","eventDateUTC":"2026-08-15T19:10:00Z","venueName":"Kitz Stadium","venueCity":"Rockford","venueState":"IL","performerName":"Los Angeles Dodgers","quantity":2,"deliveryOption":"Barcode","sellerPayout":{"amount":211.2,"currency":"USD"},"seats":[{"section":"108","row":"A","seat":"5"},{"section":"108","row":"A","seat":"6"}]}}`,
		},
		{
			description: "get order barcodes",
			method:      m.GetOrderBarcodes,
			orderInput: OrderInput{
				ExternalOrderID: "12346",
			},
			expectedOutput: `{"external_order_id":"12346","market":"stubhub","status":"fulfilled","tickets":[{"barcode":"1234567890123","enhanced_barcode":"1234567890123-ENH","delivery_type":"api_transfer","active":true,"section_code":"108","row":"A","seat_number":"5"},{"barcode":"1234567890124","enhanced_barcode":"1234567890124-ENH","delivery_type":"api_transfer","active":true,"section_code":"108","row":"A","seat_number":"6"}]}`,
		},
		{
			description: "get order event",
			method:      m.GetOrderEvent,
			orderInput: OrderInput{
				ExternalOrderID: "12346",
			},
			expectedOutput: `{"external_order_id":"12346","market":"stubhub","event":{"name":"Padres at Dodgers","date_time":"2026-08-15T19:10:00Z","team_name":"Los Angeles Dodgers","venue_name":"Dodger Stadium"},"raw_response":{"id":"678-90123456-45678","eventId":"154820391","eventDescription":"Padres at Dodgers","eventDateLocal":"2026-08-15T12:10:00","eventDateUTC":"2026-08-15T19:10:00Z","venueName":"Kitz Stadium","venueCity":"Rockford","venueState":"IL","performerName":"Los Angeles Dodgers","quantity":2,"deliveryOption":"Barcode","sellerPayout":{"amount":211.2,"currency":"USD"},"seats":[{"section":"108","row":"A","seat":"5"},{"section":"108","row":"A","seat":"6"}]}}`,
		},
		{
			description: "get order double sale check",
			method:      m.GetOrderDoubleSaleCheck,
			orderInput: OrderInput{
				ExternalOrderID: "12346",
			},
			expectedOutput: `{"external_order_id":"12346","collisions":[],"checked_at":"2026-08-14T12:00:00Z"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			_, out, err := tt.method(context.Background(), nil, tt.orderInput)
			if err != nil {
				t.Fatalf("GetFailedOrders returned error: %v", err)
			}

			if out.Data != tt.expectedOutput {
				t.Fatalf("got %s, expected %s", out.Data, tt.expectedOutput)
			}
		})
	}
}
