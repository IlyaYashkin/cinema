package smtp

import (
	"cinema/internal/lib/config"
	"cinema/internal/lib/env"
	"context"
	"fmt"
	"log/slog"

	"github.com/wneessen/go-mail"
)

type Client struct {
	log    *slog.Logger
	client *mail.Client
	from   string
}

func New(log *slog.Logger, cfg config.SMTPConfig, e env.Env) (*Client, error) {
	tlsPolicyOption := mail.WithTLSPolicy(mail.TLSOpportunistic)

	if e.Is(env.Prod) {
		tlsPolicyOption = mail.WithTLSPolicy(mail.TLSMandatory)
	}

	c, err := mail.NewClient(
		cfg.Host,
		mail.WithPort(cfg.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(cfg.Username),
		mail.WithPassword(cfg.Password),
		tlsPolicyOption,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create smtp client: %w", err)
	}

	return &Client{log: log, client: c, from: cfg.From}, nil
}

func (c *Client) Connect() {
	if err := c.client.DialWithContext(context.Background()); err != nil {
		c.log.Error("smtp connection failed: " + err.Error())
	}
	_ = c.client.Close()

	c.log.Info("smtp connection successful")
}

func (c *Client) Send(to, subject, body string) error {
	m := mail.NewMsg()

	if err := m.From(c.from); err != nil {
		return fmt.Errorf("failed to set from: %w", err)
	}
	if err := m.To(to); err != nil {
		return fmt.Errorf("failed to set to: %w", err)
	}

	m.Subject(subject)
	m.SetBodyString(mail.TypeTextHTML, body)

	if err := c.client.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
