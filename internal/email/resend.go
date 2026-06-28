package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type Sender struct {
	APIKey    string
	FromEmail string
}

func NewSender(apiKey, fromEmail string) *Sender {
	return &Sender{APIKey: apiKey, FromEmail: fromEmail}
}

func (s *Sender) Send(to, subject, html string) error {
	if s.APIKey == "" {
		slog.Info("[dev] email logged only", "to", to, "subject", subject)
		return nil
	}
	payload := map[string]interface{}{
		"from":    s.FromEmail,
		"to":      []string{to},
		"subject": subject,
		"html":    html,
	}
	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(data))
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("resend request failed", "to", to, "err", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		slog.Error("resend API error", "to", to, "status", resp.StatusCode, "body", string(b))
		return fmt.Errorf("resend error: %s", string(b))
	}
	slog.Debug("resend email sent", "to", to, "subject", subject)
	return nil
}
