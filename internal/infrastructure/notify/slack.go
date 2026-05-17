package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	requestsuc "github.com/morimasaki-web/ponsu/internal/usecase/requests"
)

type SlackNotifier struct {
	WebhookURL string
	Client     *http.Client
}

func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{WebhookURL: webhookURL}
}

func (n *SlackNotifier) Notify(ctx context.Context, in requestsuc.Notification) error {
	if n == nil {
		return nil
	}
	if strings.TrimSpace(n.WebhookURL) == "" {
		return nil
	}

	client := n.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	text := formatSlackText(in)
	payload := map[string]string{"text": text}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.WebhookURL, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}
	return nil
}

func formatSlackText(n requestsuc.Notification) string {
	kind := "Request"
	switch n.Kind {
	case requestsuc.NotificationKindSubmitted:
		kind = "Request submitted"
	case requestsuc.NotificationKindApproved:
		kind = "Request approved"
	case requestsuc.NotificationKindRejected:
		kind = "Request rejected"
	}

	title := strings.TrimSpace(n.RequestTitle)
	if title == "" {
		title = n.RequestID.String()
	}

	actor := strings.TrimSpace(n.Actor)
	if actor == "" {
		actor = "(unknown)"
	}

	url := strings.TrimSpace(n.URL)
	if url == "" {
		url = "(no url)"
	}

	// Slack mrkdwn link: <url|label>
	link := url
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		link = "<" + url + "|Open request>"
	}

	return kind + "\n" +
		"Title: " + title + "\n" +
		"Actor: " + actor + "\n" +
		"URL: " + link
}

var _ requestsuc.Notifier = (*SlackNotifier)(nil)

var ErrSlackNotConfigured = errors.New("slack webhook url is empty")
