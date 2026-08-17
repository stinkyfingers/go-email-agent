package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// NotifySlack posts text to the Slack incoming webhook configured via the
// SLACK_WEBHOOK environment variable. If that's unset, it's a no-op — so
// local runs without Slack configured don't need special-casing.
func NotifySlack(ctx context.Context, text string) error {
	webhookURL := os.Getenv("SLACK_WEBHOOK")
	if webhookURL == "" {
		log.Print("no slack webhook configured")
		return nil
	}

	body, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return fmt.Errorf("marshaling slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("building slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("posting to slack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}
	return nil
}
