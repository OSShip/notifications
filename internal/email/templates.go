package email

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
)

//go:embed templates/*.html
var templateFS embed.FS

var eventTemplates = map[string]string{
	"enrollment.confirmed":  "enrollment.confirmed.html",
	"payout.recorded":       "payout.recorded.html",
	"session.scheduled":     "session.scheduled.html",
	"session.reminder_due":  "session.reminder_due.html",
	"mentor.approved":       "mentor.approved.html",
	"listing.created":       "listing.created.html",
}

var eventSubjects = map[string]string{
	"enrollment.confirmed": "Enrollment Confirmed",
	"payout.recorded":      "Payout Recorded",
	"session.scheduled":    "Session Scheduled",
	"session.reminder_due": "Session Reminder",
	"mentor.approved":      "Mentor Application Approved",
	"listing.created":      "Listing Created",
}

type templateData struct {
	EventType string
	Payload   map[string]string
}

func Render(eventType string, payload json.RawMessage) (subject, html string, ok bool) {
	tplFile, known := eventTemplates[eventType]
	if !known {
		return "", "", false
	}
	subject = eventSubjects[eventType]

	var fields map[string]string
	_ = json.Unmarshal(payload, &fields)
	if fields == nil {
		fields = map[string]string{}
	}

	tpl, err := template.ParseFS(templateFS, "templates/"+tplFile)
	if err != nil {
		return subject, fmt.Sprintf("<p>%s</p>", subject), true
	}
	var buf bytes.Buffer
	data := templateData{EventType: eventType, Payload: fields}
	if err := tpl.Execute(&buf, data); err != nil {
		return subject, fmt.Sprintf("<p>%s</p>", subject), true
	}
	return subject, buf.String(), true
}

func ResolveRecipient(eventType string, payload json.RawMessage) string {
	var fields map[string]string
	_ = json.Unmarshal(payload, &fields)
	if fields == nil {
		return ""
	}
	switch eventType {
	case "enrollment.confirmed", "session.scheduled", "session.reminder_due":
		if e := strings.TrimSpace(fields["student_email"]); e != "" {
			return e
		}
	case "listing.created", "mentor.approved", "payout.recorded":
		if e := strings.TrimSpace(fields["mentor_email"]); e != "" {
			return e
		}
	}
	if e := strings.TrimSpace(fields["student_email"]); e != "" {
		return e
	}
	return strings.TrimSpace(fields["mentor_email"])
}
