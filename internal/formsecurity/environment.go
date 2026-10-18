package formsecurity

import (
	"context"
	"errors"
	"os"
	"strconv"

	"github.com/drumandbytes/eraser/internal/config"
	"github.com/drumandbytes/eraser/internal/email"
)

func FromEnvironment() *Service {
	cfg := Config{SiteKey: os.Getenv("CQ_TURNSTILE_SITE_KEY")}
	if secret := os.Getenv("CQ_TURNSTILE_SECRET"); secret != "" {
		cfg.CheckHuman = Turnstile(secret)
	}
	port, _ := strconv.Atoi(os.Getenv("CQ_VERIFICATION_SMTP_PORT"))
	from, err := canonicalEmail(os.Getenv("CQ_VERIFICATION_FROM"))
	host, user, password := os.Getenv("CQ_VERIFICATION_SMTP_HOST"), os.Getenv("CQ_VERIFICATION_SMTP_USER"), os.Getenv("CQ_VERIFICATION_SMTP_PASSWORD")
	if os.Getenv("CQ_VERIFICATION_EMAIL_ENABLED") == "true" && err == nil && port > 0 && port < 65536 && host != "" && user != "" && password != "" {
		sender := email.NewSMTPSender(config.SMTPConfig{Host: host, Port: port, Username: user, Password: password}, from)
		cfg.SendCode = func(ctx context.Context, to, code string) error {
			result := sender.Send(ctx, email.Message{From: from, To: to, Subject: "Confirm your Cease & Quiet email address", Body: "Your Cease & Quiet verification code is: " + code + "\n\nIt expires in 10 minutes. If you did not request this, ignore this email. No broker request has been sent."})
			if !result.Success {
				return errors.New("verification delivery unavailable")
			}
			return nil
		}
	}
	return New(cfg)
}
