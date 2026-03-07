package smtp

import (
	"cinema/internal/lib/config"
	"fmt"

	"github.com/wneessen/go-mail"
)

type Client struct {
	client *mail.Client
	from   string
}

func New(cfg config.SMTPConfig) (*Client, error) {
	c, err := mail.NewClient(
		cfg.Host,
		mail.WithPort(cfg.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(cfg.Username),
		mail.WithPassword(cfg.Password),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create smtp client: %w", err)
	}

	return &Client{client: c, from: cfg.From}, nil
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
