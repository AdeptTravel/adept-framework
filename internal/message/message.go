// internal/message/message.go
//
// Adept – Messaging stub.
//
// Context
//   The forms subsystem (and other parts of Adept) enqueue outbound messages
//   such as emails and webhooks.  Until the real queue/worker pool is
//   finished, this stub logs the payload and returns nil so callers proceed
//   without blocking.
//
//   Replace the body of each Enqueue* function with code that publishes to
//   your queue of choice (e.g., Redis, NATS, SQS) when ready.
//
// Style
//   Two-space sentence spacing, Oxford comma, concise inline notes.
//
//------------------------------------------------------------------------------

package message

import (
	"context"
	"log"
	"net/http"
)

// Email represents a basic outbound email job.
type Email struct {
	To      []string
	Subject string
	Text    string
	HTML    string // optional – not used by stub
}

// EnqueueEmail logs the email payload.  Swap with real queue publisher later.
func EnqueueEmail(ctx context.Context, msg Email) error {
	log.Printf("[Adept] QUEUE Email → to=%v subject=%q len(text)=%d\n",
		msg.To, msg.Subject, len(msg.Text))
	return nil
}

// EnqueueWebhook logs the HTTP request details.  Swap with real queue later.
//
// Caller constructs the *http.Request with full context (headers, JSON body).
func EnqueueWebhook(ctx context.Context, req *http.Request) error {
	log.Printf("[Adept] QUEUE Webhook → %s %s (headers=%d)\n",
		req.Method, req.URL.String(), len(req.Header))
	return nil
}
