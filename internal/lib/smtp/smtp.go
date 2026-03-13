package smtp

import (
	"cinema/internal/lib/config"
	"cinema/internal/lib/env"
	"cinema/internal/lib/sl"
	"context"
	"log/slog"

	"github.com/wneessen/go-mail"
)

type Client struct {
	log    *slog.Logger
	client *mail.Client
	from   string
}

func New(log *slog.Logger, cfg config.SMTPConfig, e env.Env) (*Client, error) {
	const op = "lib.smtp.new"

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
		return nil, sl.WrapErr(op, err)
	}

	return &Client{log: log, client: c, from: cfg.From}, nil
}

func (c *Client) Connect() {
	const op = "lib.smtp.connect"

	log := c.log.With(slog.String("op", op))

	if err := c.client.DialWithContext(context.Background()); err != nil {
		log.Error("smtp connection failed: " + err.Error())
	}
	_ = c.client.Close()

	log.Info("smtp connection successful")
}

func (c *Client) Send(to, subject, body string) error {
	const op = "lib.smtp.send"

	m := mail.NewMsg()

	if err := m.From(c.from); err != nil {
		return sl.WrapErr(op, err)
	}
	if err := m.To(to); err != nil {
		return sl.WrapErr(op, err)
	}

	m.Subject(subject)
	m.SetBodyString(mail.TypeTextHTML, body)

	if err := c.client.DialAndSend(m); err != nil {
		return sl.WrapErr(op, err)
	}

	return nil
}
