package fakes

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// [start code]

// Mailer sends email. The real implementation talks to an SMTP server
// or an API; tests use a fake.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

// Clock tells the time. Tests use a fixed clock.
type Clock interface {
	Now() time.Time
}

// Reminder sends invoice reminders.
type Reminder struct {
	Mail  Mailer
	Clock Clock
}

type Invoice struct {
	ID    string
	Email string
	Due   time.Time
	Paid  bool
}

// SendOverdue emails every unpaid invoice past its due date and returns
// how many reminders were sent. It keeps going when one email fails and
// reports all failures together.
func (r *Reminder) SendOverdue(ctx context.Context, invoices []Invoice) (int, error) {
	now := r.Clock.Now()
	sent := 0
	var errs []error
	for _, inv := range invoices {
		if inv.Paid || !now.After(inv.Due) {
			continue
		}
		days := int(now.Sub(inv.Due).Hours() / 24)
		subject := fmt.Sprintf("Invoice %s is %d days overdue", inv.ID, days)
		if err := r.Mail.Send(ctx, inv.Email, subject, "Please pay."); err != nil {
			errs = append(errs, fmt.Errorf("invoice %s: %w", inv.ID, err))
			continue
		}
		sent++
	}
	return sent, errors.Join(errs...)
}

// [end code]
