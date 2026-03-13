package smtp

import (
	"cinema/internal/lib/sl"
	"cinema/internal/lib/smtp"
	"context"
	"fmt"
)

type EmailSender struct {
	client               *smtp.Client
	resetPasswordBaseUrl string
}

func NewEmailSender(client *smtp.Client, resetPasswordBaseUrl string) *EmailSender {
	return &EmailSender{
		client:               client,
		resetPasswordBaseUrl: resetPasswordBaseUrl,
	}
}

func (e *EmailSender) SendPasswordResetNotification(_ context.Context, email, resetToken string) error {
	const op = "sso.notification.send_password_reset"

	resetPasswordLink := fmt.Sprintf("%s?token=%s", e.resetPasswordBaseUrl, resetToken)

	subject := "Сброс пароля"
	body := fmt.Sprintf("Для сброса пароля перейдите по ссылке: <a href=\"%s\">ссылка</a>", resetPasswordLink)

	err := e.client.Send(email, subject, body)
	if err != nil {
		return sl.WrapErr(op, err)
	}

	return nil
}
