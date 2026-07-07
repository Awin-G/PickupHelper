package service

import (
	"context"
	"crypto/rand"
	"log/slog"
	"math/big"
)

// SMSProvider abstracts the SMS verification-code channel.
// In v1 a stub implementation is used for all environments; v2 will swap
// in an Aliyun SMS implementation without changing call sites.
type SMSProvider interface {
	// Send delivers the code to the given phone. Returns nil on success.
	Send(ctx context.Context, phone, code string) error
	// GenerateCode returns a 6-digit code — fixed "123456" in dev/test,
	// random in production.
	GenerateCode() string
}

// stubSMS logs the code instead of sending a real SMS. dev/test always
// produce "123456"; production produces a random 6-digit string.
type stubSMS struct {
	env string
	log *slog.Logger
}

// NewSMSProvider returns a stub SMSProvider for the given environment.
// env values "dev" and "test" yield the fixed code "123456"; any other
// value (e.g. "prod") yields random 6-digit codes.
func NewSMSProvider(env string, log *slog.Logger) SMSProvider {
	if log == nil {
		log = slog.Default()
	}
	return &stubSMS{env: env, log: log}
}

func (s *stubSMS) Send(ctx context.Context, phone, code string) error {
	s.log.InfoContext(ctx, "sms stub: send code",
		slog.String("phone", phone),
		slog.String("code", code),
		slog.String("env", s.env))
	return nil
}

func (s *stubSMS) GenerateCode() string {
	if s.env == "dev" || s.env == "test" {
		return "123456"
	}
	return randomCode(6)
}

// randomCode returns an n-digit numeric string with a uniform distribution
// and no leading-zero bias. n must be > 0.
func randomCode(n int) string {
	if n <= 0 {
		return ""
	}
	digits := make([]byte, n)
	for i := 0; i < n; i++ {
		b, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			// Should never happen with crypto/rand; fall back to 0.
			digits[i] = '0'
			continue
		}
		digits[i] = '0' + byte(b.Int64())
	}
	return string(digits)
}
