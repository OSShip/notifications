package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
		log.Printf("[dev] email to=%s subject=%s", to, subject)
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
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend error: %s", string(b))
	}
	return nil
}
