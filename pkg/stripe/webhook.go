package stripe

import (
	"errors"
	"io"
	"net/http"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

func ConstructEvent(r *http.Request, signingSecret string) (*stripe.Event, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	sig := r.Header.Get("Stripe-Signature")
	e, err := webhook.ConstructEvent(body, sig, signingSecret)
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func IsWebhookClientError(err error) bool {
	switch {
	case errors.Is(err, webhook.ErrInvalidHeader):
		return true
	case errors.Is(err, webhook.ErrNoValidSignature):
		return true
	case errors.Is(err, webhook.ErrNotSigned):
		return true
	case errors.Is(err, webhook.ErrTooOld):
		return true
	default:
		return false
	}
}
