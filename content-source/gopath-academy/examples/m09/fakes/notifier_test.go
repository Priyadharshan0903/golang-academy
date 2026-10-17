package fakes

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// [start fakes]

// fakeMailer records what would have been sent, and can fail on demand.
// It is a fake: a working, simplified implementation, not a scripted mock.
type fakeMailer struct {
	sent   []string
	failTo string
}

func (f *fakeMailer) Send(_ context.Context, to, subject, _ string) error {
	if to == f.failTo {
		return errors.New("smtp: connection refused")
	}
	f.sent = append(f.sent, to+": "+subject)
	return nil
}

// fixedClock always returns the same time.
type fixedClock time.Time

func (c fixedClock) Now() time.Time { return time.Time(c) }

// [end fakes]

// [start test]

func TestSendOverdue(t *testing.T) {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	mail := &fakeMailer{failTo: "broken@example.com"}
	r := &Reminder{Mail: mail, Clock: fixedClock(now)}

	invoices := []Invoice{
		{ID: "A1", Email: "ana@example.com", Due: now.AddDate(0, 0, -10)},
		{ID: "B2", Email: "bo@example.com", Due: now.AddDate(0, 0, 5)},                // not due yet
		{ID: "C3", Email: "cara@example.com", Due: now.AddDate(0, 0, -1), Paid: true}, // paid
		{ID: "D4", Email: "broken@example.com", Due: now.AddDate(0, 0, -3)},           // send fails
	}

	sent, err := r.SendOverdue(context.Background(), invoices)

	if sent != 1 {
		t.Errorf("sent = %d, want 1", sent)
	}
	if want := "ana@example.com: Invoice A1 is 10 days overdue"; len(mail.sent) != 1 || mail.sent[0] != want {
		t.Errorf("emails = %q, want [%q]", mail.sent, want)
	}
	if err == nil || !strings.Contains(err.Error(), "invoice D4") {
		t.Errorf("err = %v, want a failure mentioning D4", err)
	}
}

// [end test]
