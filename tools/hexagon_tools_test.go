package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
)

func TestHexagonTools(t *testing.T) {
	const wantAuth = "Bearer test-token"
	th := &ToolHandler{
		HexagonToken: "test-token",
	}

	tests := []struct {
		endpoint   string
		method     func() (anthropic.BetaTool, error)
		orderInput OrderInput
	}{
		{
			endpoint:   "/orders",
			method:     th.GetOrdersTool,
			orderInput: OrderInput{},
		},
		{
			endpoint:   "/orders/123",
			method:     th.GetOrderTool,
			orderInput: OrderInput{ExternalOrderID: "123"},
		},
		{
			endpoint:   "/orders/failed",
			method:     th.GetFailedOrdersTool,
			orderInput: OrderInput{},
		},
		{
			endpoint:   "/orders/123/barcodes",
			method:     th.GetOrderBarcodesTool,
			orderInput: OrderInput{ExternalOrderID: "123"},
		},
		{
			endpoint:   "/orders/123/event",
			method:     th.GetOrderEventTool,
			orderInput: OrderInput{ExternalOrderID: "123"},
		},
		{
			endpoint:   "/orders/123/double_sale_check",
			method:     th.GetOrderDoubleSaleCheckTool,
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
			th.HexagonURL = server.URL // set url

			tool, err := tt.method()
			if err != nil {
				t.Fatalf("%s: building tool failed: %v", tt.endpoint, err)
			}

			input, err := json.Marshal(tt.orderInput)
			if err != nil {
				t.Fatal(err)
			}

			results, err := tool.Execute(context.Background(), input)
			if err != nil {
				t.Fatalf("%s returned error: %v", tt.endpoint, err)
			}
			if len(results) != 1 || results[0].OfText == nil {
				t.Fatalf("%s: expected exactly one text result, got %+v", tt.endpoint, results)
			}

			// HexagonApiGet returns the raw response body as a plain string,
			// and each *Tool wraps it with json.Marshal before setting Text —
			// so a raw "test response" body round-trips through Marshal into
			// a quoted JSON string literal, not the bare bytes.
			want := `"test response"`
			if results[0].OfText.Text != want {
				t.Fatalf("got %s, expected %s", results[0].OfText.Text, want)
			}
		})
	}
}

func TestHexagonLive(t *testing.T) {
	// t.Skip("un-skip to use for testing against local hexagon API during development")

	th := &ToolHandler{
		HexagonURL:   "http://localhost:3001/api/agent/v1",
		HexagonToken: "abcdef",
	}
	tests := []struct {
		description    string
		method         func() (anthropic.BetaTool, error)
		orderInput     OrderInput
		expectedOutput string
	}{
		{
			description:    "failed orders",
			method:         th.GetFailedOrdersTool,
			orderInput:     OrderInput{},
			expectedOutput: `{"orders":[{"id":55501,"external_order_id":"678-90123456-11111","market":"stubhub","status":"finalizing_stalled","quantity":2,"total_price_cents":18000,"customer_email":"buyer@example.com","team_name":"Los Angeles Dodgers","venue_name":"Dodger Stadium","created_at":"2026-08-10T18:00:00Z"},{"id":55502,"external_order_id":"901-23456789-22222","market":"seatgeek","status":"error","quantity":4,"total_price_cents":41200,"customer_email":"someone-else@example.com","team_name":"San Diego Padres","venue_name":"Petco Park","created_at":"2026-08-11T09:12:00Z"}],"count":2}`,
		},
		{
			description:    "get orders",
			method:         th.GetOrdersTool,
			orderInput:     OrderInput{},
			expectedOutput: `{"orders":[{"id":12345,"external_order_id":"678-90123456-45678","market":"stubhub","status":"fulfilled","quantity":2,"total_price_cents":24500,"created_at":"2026-08-10T18:00:00Z","fulfilled_at":"2026-08-10T19:15:00Z"},{"id":12346,"external_order_id":"901-23456789-01234","market":"seatgeek","status":"fulfilled","quantity":4,"total_price_cents":51000,"created_at":"2026-08-11T11:04:00Z","fulfilled_at":"2026-08-11T11:41:00Z"},{"id":12347,"external_order_id":"678-90123456-77777","market":"stubhub","status":"confirmed","quantity":1,"total_price_cents":9800,"created_at":"2026-08-12T15:22:00Z","fulfilled_at":null}],"count":3}`,
		},
		{
			description: "get order",
			method:      th.GetOrderTool,
			orderInput: OrderInput{
				ExternalOrderID: "12346",
			},
			expectedOutput: `{"id":12345,"external_order_id":"12346","market":"stubhub","status":"fulfilled","quantity":2,"total_price_cents":24500,"subtotal_price_cents":22000,"remittance_amount_cents":21120,"customer_email":"buyer@example.com","customer_first_name":"Alex","customer_last_name":"Rivera","confirmed_at":"2026-08-10T18:30:00Z","fulfilled_at":"2026-08-10T19:15:00Z","rejected_at":null,"refunded_at":null,"refund_amount_cents":null,"partial_refund":false,"delay_fill":false,"created_at":"2026-08-10T18:00:00Z","updated_at":"2026-08-10T19:15:00Z","raw_response":{"id":"678-90123456-45678","eventId":"154820391","eventDescription":"Padres at Dodgers","eventDateLocal":"2026-08-15T12:10:00","eventDateUTC":"2026-08-15T19:10:00Z","venueName":"Kitz Stadium","venueCity":"Rockford","venueState":"IL","performerName":"Los Angeles Dodgers","quantity":2,"deliveryOption":"Barcode","sellerPayout":{"amount":211.2,"currency":"USD"},"seats":[{"section":"108","row":"A","seat":"5"},{"section":"108","row":"A","seat":"6"}]}}`,
		},
		{
			description: "get order barcodes",
			method:      th.GetOrderBarcodesTool,
			orderInput: OrderInput{
				ExternalOrderID: "12346",
			},
			expectedOutput: `{"external_order_id":"12346","market":"stubhub","status":"fulfilled","tickets":[{"barcode":"1234567890123","enhanced_barcode":"1234567890123-ENH","delivery_type":"api_transfer","active":true,"section_code":"108","row":"A","seat_number":"5"},{"barcode":"1234567890124","enhanced_barcode":"1234567890124-ENH","delivery_type":"api_transfer","active":true,"section_code":"108","row":"A","seat_number":"6"}]}`,
		},
		{
			description: "get order event",
			method:      th.GetOrderEventTool,
			orderInput: OrderInput{
				ExternalOrderID: "12346",
			},
			expectedOutput: `{"external_order_id":"12346","market":"stubhub","event":{"name":"Padres at Dodgers","date_time":"2026-08-15T19:10:00Z","team_name":"Los Angeles Dodgers","venue_name":"Dodger Stadium"},"raw_response":{"id":"678-90123456-45678","eventId":"154820391","eventDescription":"Padres at Dodgers","eventDateLocal":"2026-08-15T12:10:00","eventDateUTC":"2026-08-15T19:10:00Z","venueName":"Kitz Stadium","venueCity":"Rockford","venueState":"IL","performerName":"Los Angeles Dodgers","quantity":2,"deliveryOption":"Barcode","sellerPayout":{"amount":211.2,"currency":"USD"},"seats":[{"section":"108","row":"A","seat":"5"},{"section":"108","row":"A","seat":"6"}]}}`,
		},
		{
			description: "get order double sale check",
			method:      th.GetOrderDoubleSaleCheckTool,
			orderInput: OrderInput{
				ExternalOrderID: "12346",
			},
			expectedOutput: `{"external_order_id":"12346","collisions":[],"checked_at":"2026-08-14T12:00:00Z"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			tool, err := tt.method()
			if err != nil {
				t.Fatalf("tool returned error: %v", err)
			}
			input, err := json.Marshal(tt.orderInput)
			if err != nil {
				t.Fatal(err)
			}

			results, err := tool.Execute(context.Background(), input)
			if err != nil {
				t.Fatalf("%s returned error: %v", tt.description, err)
			}
			if len(results) != 1 || results[0].OfText == nil {
				t.Fatalf("%s: expected exactly one text result, got %+v", tt.description, results)
			}
			if results[0].OfText.Text != tt.expectedOutput {
				t.Fatalf("got %s, expected %s", results[0].OfText.Text, tt.expectedOutput)
			}
		})
	}
}
