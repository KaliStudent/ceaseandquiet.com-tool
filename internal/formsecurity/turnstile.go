package formsecurity

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func Turnstile(secret string) func(context.Context, string) error {
	client := &http.Client{Timeout: 4 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return func(ctx context.Context, token string) error {
		body := url.Values{"secret": {secret}, "response": {token}}
		r, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://challenges.cloudflare.com/turnstile/v0/siteverify", strings.NewReader(body.Encode()))
		if err != nil {
			return errors.New("bot check unavailable")
		}
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response, err := client.Do(r)
		if err != nil {
			return errors.New("bot check unavailable")
		}
		defer func() { _ = response.Body.Close() }()
		var result struct {
			Success  bool   `json:"success"`
			Hostname string `json:"hostname"`
			Action   string `json:"action"`
		}
		if response.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(response.Body, 16384)).Decode(&result) != nil || !result.Success || result.Hostname != "abcw.xyz" || result.Action != "profile_form" {
			return errors.New("bot check failed")
		}
		return nil
	}
}
