package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/toolrunner"
)

type OrderInput struct {
	ExternalOrderID string `json:"external_order_id"`
}

func (t *ToolHandler) GetOrdersTool() (anthropic.BetaTool, error) {
	return toolrunner.NewBetaToolFromJSONSchema(
		"get_orders",
		"list orders",
		func(ctx context.Context, input OrderInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			data, err := t.HexagonApiGet("orders")
			if err != nil {
				return anthropic.BetaToolResultBlockParamContentUnion{}, err
			}

			return anthropic.BetaToolResultBlockParamContentUnion{
				OfText: &anthropic.BetaTextBlockParam{
					Text: data,
				},
			}, nil
		},
	)
}

func (t *ToolHandler) GetOrderTool() (anthropic.BetaTool, error) {
	return toolrunner.NewBetaToolFromJSONSchema(
		"get_order",
		"get a specific order by external_order_id",
		func(ctx context.Context, input OrderInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			data, err := t.HexagonApiGet(fmt.Sprintf("orders/%s", input.ExternalOrderID))
			if err != nil {
				return anthropic.BetaToolResultBlockParamContentUnion{}, err
			}

			return anthropic.BetaToolResultBlockParamContentUnion{
				OfText: &anthropic.BetaTextBlockParam{
					Text: data,
				},
			}, nil
		},
	)
}

func (t *ToolHandler) GetFailedOrdersTool() (anthropic.BetaTool, error) {
	return toolrunner.NewBetaToolFromJSONSchema(
		"get_failed_orders",
		"list failed orders",
		func(ctx context.Context, input OrderInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			data, err := t.HexagonApiGet("orders/failed")
			if err != nil {
				return anthropic.BetaToolResultBlockParamContentUnion{}, err
			}

			return anthropic.BetaToolResultBlockParamContentUnion{
				OfText: &anthropic.BetaTextBlockParam{
					Text: data,
				},
			}, nil
		},
	)
}

func (t *ToolHandler) GetOrderBarcodesTool() (anthropic.BetaTool, error) {
	return toolrunner.NewBetaToolFromJSONSchema(
		"get_order_barcodes",
		"get order barcodes by external_order_id",
		func(ctx context.Context, input OrderInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			data, err := t.HexagonApiGet(fmt.Sprintf("orders/%s/barcodes", input.ExternalOrderID))
			if err != nil {
				return anthropic.BetaToolResultBlockParamContentUnion{}, err
			}

			return anthropic.BetaToolResultBlockParamContentUnion{
				OfText: &anthropic.BetaTextBlockParam{
					Text: data,
				},
			}, nil
		},
	)
}

func (t *ToolHandler) GetOrderEventTool() (anthropic.BetaTool, error) {
	return toolrunner.NewBetaToolFromJSONSchema(
		"get_order_event",
		"get order event by external_order_id",
		func(ctx context.Context, input OrderInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			data, err := t.HexagonApiGet(fmt.Sprintf("orders/%s/event", input.ExternalOrderID))
			if err != nil {
				return anthropic.BetaToolResultBlockParamContentUnion{}, err
			}
			return anthropic.BetaToolResultBlockParamContentUnion{
				OfText: &anthropic.BetaTextBlockParam{
					Text: data,
				},
			}, nil
		},
	)
}

func (t *ToolHandler) GetOrderDoubleSaleCheckTool() (anthropic.BetaTool, error) {
	return toolrunner.NewBetaToolFromJSONSchema(
		"get_order_double_sale_check",
		"check for double sales of an order by external_order_id",
		func(ctx context.Context, input OrderInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			data, err := t.HexagonApiGet(fmt.Sprintf("orders/%s/double_sale_check", input.ExternalOrderID))
			if err != nil {
				return anthropic.BetaToolResultBlockParamContentUnion{}, err
			}

			return anthropic.BetaToolResultBlockParamContentUnion{
				OfText: &anthropic.BetaTextBlockParam{
					Text: data,
				},
			}, nil
		},
	)
}

/* utils */

// HexagonApiGet makes a GET call to the Hexagon API
func (t *ToolHandler) HexagonApiGet(endpoint string) (string, error) {
	requestURL := fmt.Sprintf("%s/%s", t.HexagonURL, endpoint)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t.HexagonToken))
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error from hexagon API: %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
